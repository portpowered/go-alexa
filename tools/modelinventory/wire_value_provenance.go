package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

const maxWireValueResolutionDepth = 128

// inspectUnresolvedWireValueCalls follows local return values before deciding
// whether a call used by a generated wire value is caller-owned. Imported
// results and unresolved helper paths fail closed; a direct caller parameter
// or caller-provided function parameter remains open.
func inspectUnresolvedWireValueCalls(
	expression ast.Expr, file *ast.File, path string, aliases map[string]bool, models map[string]generatedModel,
	assignments wireSourceAssignments, report func(token.Pos, string),
) {
	inspectWireValueExpression(expression, wireSourceContext{file: file, path: path}, aliases, path, models,
		assignments, make(map[wireSourceVariable]bool), 0, report)
}

//nolint:cyclop // The expression switch is one exhaustive traversal over Go syntax forms.
func inspectWireValueExpression(
	expression ast.Expr, fallback wireSourceContext, aliases map[string]bool, path string,
	models map[string]generatedModel, assignments wireSourceAssignments,
	visiting map[wireSourceVariable]bool, depth int, report func(token.Pos, string),
) {
	if expression == nil {
		return
	}

	if depth > maxWireValueResolutionDepth {
		report(expression.Pos(), "wire value helper resolution exceeded the traversal limit")

		return
	}

	switch value := expression.(type) {
	case *ast.BasicLit, *ast.FuncLit:
		return
	case *ast.Ident:
		inspectWireValueAlias(value, fallback, aliases, path, models, assignments, visiting, depth, report)
	case *ast.SelectorExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.CallExpr:
		callContext := wireSourceForNode(value, fallback, assignments)
		for _, argument := range value.Args {
			inspectWireValueExpression(argument, wireSourceForNode(argument, callContext, assignments), aliases, path, models,
				assignments, visiting, depth+1, report)
		}

		inspectWireCallResult(value, callContext, aliases, path, models, assignments, visiting, depth, report)
		inspectWireValueExpression(value.Fun, wireSourceForNode(value.Fun, callContext, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.CompositeLit:
		for _, element := range value.Elts {
			inspectWireValueExpression(element, wireSourceForNode(element, fallback, assignments), aliases, path, models,
				assignments, visiting, depth+1, report)
		}
	case *ast.KeyValueExpr:
		inspectWireValueExpression(value.Key, wireSourceForNode(value.Key, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
		inspectWireValueExpression(value.Value, wireSourceForNode(value.Value, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.BinaryExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
		inspectWireValueExpression(value.Y, wireSourceForNode(value.Y, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.UnaryExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.StarExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.ParenExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.IndexExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
		inspectWireValueExpression(value.Index, wireSourceForNode(value.Index, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.IndexListExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)

		for _, index := range value.Indices {
			inspectWireValueExpression(index, wireSourceForNode(index, fallback, assignments), aliases, path, models,
				assignments, visiting, depth+1, report)
		}
	case *ast.SliceExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
		inspectWireValueExpression(value.Low, wireSourceForNode(value.Low, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
		inspectWireValueExpression(value.High, wireSourceForNode(value.High, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
		inspectWireValueExpression(value.Max, wireSourceForNode(value.Max, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.TypeAssertExpr:
		inspectWireValueExpression(value.X, wireSourceForNode(value.X, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	case *ast.Ellipsis:
		inspectWireValueExpression(value.Elt, wireSourceForNode(value.Elt, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	}
}

func inspectWireValueAlias(
	identifier *ast.Ident, fallback wireSourceContext, aliases map[string]bool, path string,
	models map[string]generatedModel, assignments wireSourceAssignments,
	visiting map[wireSourceVariable]bool, depth int, report func(token.Pos, string),
) {
	if identifier.Obj == nil {
		return
	}

	declaration, node := identifier.Obj.Decl.(ast.Node)

	key := wireSourceVariable{declaration: declaration, name: identifier.Name}

	if !node || visiting[key] {
		if node && wireSourceAliasHasCall(identifier, assignments) {
			report(identifier.Pos(), "wire value has recursive or unresolved provenance")
		}

		return
	}

	visiting[key] = true
	defer delete(visiting, key)

	sources := wireAliasExpressions(identifier, assignments)

	if field, parameter := declaration.(*ast.Field); parameter && wireFunctionParameter(field, assignments) &&
		!wireCallerOwnedParameter(field, assignments) && !assignments.boundParameters[key] {
		callArguments := assignments.callArguments[key]
		for _, argument := range callArguments {
			inspectWireValueLiterals(argument, make(map[wireSourceVariable]bool), func(literal *ast.BasicLit) {
				value, err := strconv.Unquote(literal.Value)
				if err == nil && value != "" {
					report(literal.Pos(), fmt.Sprintf("fixed wire value %q must be generated from schema", value))
				}
			}, assignments)
		}

		sources = append(sources, callArguments...)
		if len(callArguments) == 0 && len(sources) == 0 {
			report(identifier.Pos(), "wire value depends on an unresolved helper parameter")

			return
		}
	}

	sources = uniqueWireValueExpressions(sources)

	for _, source := range sources {
		if call, isCall := source.(*ast.CallExpr); isCall {
			if index, selected := assignments.resultIndexes[key][source]; selected {
				previous, existed := assignments.resultSelections[call]
				assignments.resultSelections[call] = index
				inspectWireValueExpression(source, wireSourceForNode(source, fallback, assignments), aliases, path, models,
					assignments, visiting, depth+1, report)

				if existed {
					assignments.resultSelections[call] = previous
				} else {
					delete(assignments.resultSelections, call)
				}

				continue
			}
		}

		inspectWireValueExpression(source, wireSourceForNode(source, fallback, assignments), aliases, path, models,
			assignments, visiting, depth+1, report)
	}
}

func uniqueWireValueExpressions(expressions []ast.Expr) []ast.Expr {
	unique := make([]ast.Expr, 0, len(expressions))

	for _, expression := range expressions {
		found := false

		for _, previous := range unique {
			if previous == expression {
				found = true

				break
			}
		}

		if !found {
			unique = append(unique, expression)
		}
	}

	return unique
}

func wireFunctionParameter(field *ast.Field, assignments wireSourceAssignments) bool {
	for function := range assignments.functionBodies {
		if function.Params == nil {
			continue
		}

		for _, candidate := range function.Params.List {
			if candidate == field {
				return true
			}
		}
	}

	return false
}

func wireCallerOwnedParameter(field *ast.Field, assignments wireSourceAssignments) bool {
	for function, declaration := range assignments.functionDecls {
		if function.Params == nil {
			continue
		}

		for _, candidate := range function.Params.List {
			if candidate == field {
				return declaration.Name.IsExported()
			}
		}
	}

	return false
}

func wireSourceAliasHasCall(identifier *ast.Ident, assignments wireSourceAssignments) bool {
	for _, expression := range wireAliasExpressions(identifier, assignments) {
		found := false

		ast.Inspect(expression, func(node ast.Node) bool {
			if _, call := node.(*ast.CallExpr); call {
				found = true

				return false
			}

			return !found
		})

		if found {
			return true
		}
	}

	return false
}

//nolint:cyclop,funlen,gocognit,nestif // Helper-return branches each enforce distinct fail-closed proof conditions.
func inspectWireCallResult(
	call *ast.CallExpr, context wireSourceContext, aliases map[string]bool, path string,
	models map[string]generatedModel, assignments wireSourceAssignments,
	visiting map[wireSourceVariable]bool, depth int, report func(token.Pos, string),
) {
	if len(call.Args) == 1 && (wireTypeConversion(call.Fun) || isGeneratedWireConstruction(call.Fun, aliases, context.path, models)) {
		return
	}

	if isGeneratedWireBuiltin(call.Fun) {
		return
	}

	if wireCallerProvidedFunction(call.Fun, assignments, make(map[wireSourceVariable]bool)) {
		return
	}

	if context.file != nil && verifiedWireCallable(context.file, call.Fun, assignments) ||
		context.path != "" && verifiedSDKDependencyCall(context.path, call.Fun, assignments) ||
		wireCallerOwnedGetterCall(call, assignments) ||
		wireHelperCopiesParameters(call, assignments) ||
		wireCallerOwnedHelperCall(call, assignments, visiting, depth) {
		return
	}

	functions := wireLocalHelpers(call.Fun, assignments, make(map[wireSourceVariable]bool))

	resolved := len(functions) != 0

	if resolved {
		for _, function := range functions {
			if function == nil {
				report(call.Pos(), "wire value has an unresolved library helper")

				continue
			}

			body := assignments.functionBodies[function]
			if body == nil {
				report(call.Pos(), "wire value has an unresolved library helper body")

				continue
			}

			key := wireSourceVariable{declaration: body, name: "wire value result"}
			if visiting[key] {
				name := "anonymous"
				if declaration := assignments.functionDecls[function]; declaration != nil {
					name = declaration.Name.Name
				}

				report(call.Pos(), fmt.Sprintf("wire value helper %s recursion prevents provenance proof", name))

				continue
			}

			if depth+1 > maxWireValueResolutionDepth {
				report(call.Pos(), "wire value helper resolution exceeded the traversal limit")

				continue
			}

			visiting[key] = true
			helperContext := wireSourceForNode(function, context, assignments)
			bound := wireReturnParameterValues(function, call, assignments)

			selectedResult, hasSelectedResult := assignments.resultSelections[call]

			ast.Inspect(body, func(node ast.Node) bool {
				if _, closure := node.(*ast.FuncLit); closure {
					return false
				}

				returned, isReturn := node.(*ast.ReturnStmt)
				if !isReturn {
					return true
				}

				results := wireReturnExpressions(function, returned)
				if hasSelectedResult {
					if selectedResult >= len(results) {
						report(call.Pos(), "wire value helper result selection could not be resolved")

						return false
					}

					results = results[selectedResult : selectedResult+1]
				}

				for _, result := range results {
					inspectWireValueExpression(result, wireSourceForNode(result, helperContext, assignments), aliases, path,
						models, bound, visiting, depth+1, report)
				}

				return false
			})

			delete(visiting, key)
		}

		return
	}

	report(call.Pos(), "unresolved call result in generated wire value is not caller-owned")
}

func wireHelperCopiesParameters(call *ast.CallExpr, assignments wireSourceAssignments) bool {
	functions := wireLocalHelpers(call.Fun, assignments, make(map[wireSourceVariable]bool))
	if len(functions) == 0 {
		return false
	}

	for _, function := range functions {
		if function == nil || assignments.functionBodies[function] == nil {
			return false
		}

		body := assignments.functionBodies[function]
		selectedResult, hasSelectedResult := assignments.resultSelections[call]
		returnedAny := false
		allCopied := true

		ast.Inspect(body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}

			returned, isReturn := node.(*ast.ReturnStmt)
			if !isReturn {
				return true
			}

			results := wireReturnExpressions(function, returned)
			if hasSelectedResult {
				if selectedResult >= len(results) {
					allCopied = false

					return false
				}

				results = results[selectedResult : selectedResult+1]
			}

			for _, result := range results {
				returnedAny = true

				if !wireCopiedParameterExpression(function, result, assignments, make(map[wireSourceVariable]bool)) {
					allCopied = false

					return false
				}
			}

			return false
		})

		if !returnedAny || !allCopied {
			return false
		}
	}

	return true
}

func wireCopiedParameterExpression(
	function *ast.FuncType, expression ast.Expr, assignments wireSourceAssignments,
	visiting map[wireSourceVariable]bool,
) bool {
	switch value := expression.(type) {
	case *ast.Ident:
		if value.Obj == nil {
			return value.Name == "nil"
		}

		field, parameter := value.Obj.Decl.(*ast.Field)
		if parameter && wireParameterBelongsToFunction(function, field) {
			return true
		}

		declaration, ok := value.Obj.Decl.(ast.Node)

		key := wireSourceVariable{declaration: declaration, name: value.Name}

		if !ok || visiting[key] {
			return false
		}

		visiting[key] = true
		defer delete(visiting, key)

		sources := wireAliasExpressions(value, assignments)
		if len(sources) == 0 {
			return false
		}

		for _, source := range uniqueWireValueExpressions(sources) {
			if !wireCopiedParameterExpression(function, source, assignments, visiting) {
				return false
			}
		}

		return true
	case *ast.ParenExpr:
		return wireCopiedParameterExpression(function, value.X, assignments, visiting)
	case *ast.StarExpr:
		return wireCopiedParameterExpression(function, value.X, assignments, visiting)
	case *ast.UnaryExpr:
		return wireCopiedParameterExpression(function, value.X, assignments, visiting)
	case *ast.SelectorExpr:
		return wireCopiedParameterExpression(function, value.X, assignments, visiting)
	case *ast.IndexExpr:
		return wireCopiedParameterExpression(function, value.X, assignments, visiting)
	case *ast.SliceExpr:
		return wireCopiedParameterExpression(function, value.X, assignments, visiting)
	}

	return false
}

func wireParameterBelongsToFunction(function *ast.FuncType, field *ast.Field) bool {
	if function.Params == nil {
		return false
	}

	for _, candidate := range function.Params.List {
		if candidate == field {
			return true
		}
	}

	return false
}

func wireCallerOwnedGetterCall(call *ast.CallExpr, assignments wireSourceAssignments) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)

	return ok && strings.HasPrefix(selector.Sel.Name, "Get") && wireCallerOwnedExpression(
		selector.X, assignments, make(map[wireSourceVariable]bool), 0,
	)
}

//nolint:gocognit // Resolving helper returns requires callsite bindings and recursion state.
func wireCallerOwnedHelperCall(
	call *ast.CallExpr, assignments wireSourceAssignments, visiting map[wireSourceVariable]bool, depth int,
) bool {
	functions := wireLocalHelpers(call.Fun, assignments, make(map[wireSourceVariable]bool))
	if len(functions) == 0 {
		return false
	}

	for _, function := range functions {
		if function == nil || assignments.functionBodies[function] == nil {
			return false
		}

		body := assignments.functionBodies[function]

		key := wireSourceVariable{declaration: body, name: "caller-owned helper result"}
		if visiting[key] || depth+1 > maxWireValueResolutionDepth {
			return false
		}

		visiting[key] = true
		bound := wireReturnParameterValues(function, call, assignments)
		returnedAny := false
		allCallerOwned := true
		selectedResult, hasSelectedResult := assignments.resultSelections[call]

		ast.Inspect(body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}

			returned, isReturn := node.(*ast.ReturnStmt)
			if !isReturn {
				return true
			}

			results := wireReturnExpressions(function, returned)
			if hasSelectedResult {
				if selectedResult >= len(results) {
					allCallerOwned = false

					return false
				}

				results = results[selectedResult : selectedResult+1]
			}

			for _, result := range results {
				returnedAny = true

				if !wireCallerOwnedExpression(result, bound, visiting, depth+1) {
					allCallerOwned = false

					return false
				}
			}

			return false
		})
		delete(visiting, key)

		if !returnedAny || !allCallerOwned {
			return false
		}
	}

	return true
}

//nolint:cyclop,gocognit // This expression walk accepts only recursively proven caller-owned values.
func wireCallerOwnedExpression(
	expression ast.Expr, assignments wireSourceAssignments, visiting map[wireSourceVariable]bool, depth int,
) bool {
	if expression == nil || depth > maxWireValueResolutionDepth {
		return false
	}

	switch value := expression.(type) {
	case *ast.Ident:
		if value.Obj == nil {
			return value.Name == "nil" || value.Name == "true" || value.Name == "false"
		}

		declaration, ok := value.Obj.Decl.(ast.Node)
		if !ok {
			return false
		}

		key := wireSourceVariable{declaration: declaration, name: value.Name}
		if visiting[key] {
			return false
		}

		if field, parameter := declaration.(*ast.Field); parameter && wireFunctionParameter(field, assignments) &&
			!assignments.boundParameters[key] {
			if wireCallerOwnedParameter(field, assignments) {
				return true
			}
		}

		visiting[key] = true
		defer delete(visiting, key)

		sources := wireAliasExpressions(value, assignments)
		if field, parameter := declaration.(*ast.Field); parameter && wireFunctionParameter(field, assignments) {
			sources = append(sources, assignments.callArguments[key]...)
		}

		if len(sources) == 0 {
			return false
		}

		for _, source := range uniqueWireValueExpressions(sources) {
			if !wireCallerOwnedExpression(source, assignments, visiting, depth+1) {
				return false
			}
		}

		return true
	case *ast.SelectorExpr:
		return wireCallerOwnedExpression(value.X, assignments, visiting, depth+1)
	case *ast.ParenExpr:
		return wireCallerOwnedExpression(value.X, assignments, visiting, depth+1)
	case *ast.StarExpr:
		return wireCallerOwnedExpression(value.X, assignments, visiting, depth+1)
	case *ast.IndexExpr:
		return wireCallerOwnedExpression(value.X, assignments, visiting, depth+1)
	case *ast.SliceExpr:
		return wireCallerOwnedExpression(value.X, assignments, visiting, depth+1)
	case *ast.TypeAssertExpr:
		return wireCallerOwnedExpression(value.X, assignments, visiting, depth+1)
	case *ast.CallExpr:
		if len(value.Args) == 1 && wireTypeConversion(value.Fun) {
			return wireCallerOwnedExpression(value.Args[0], assignments, visiting, depth+1)
		}

		return wireCallerOwnedHelperCall(value, assignments, visiting, depth+1)
	}

	return false
}

func wireSourceForNode(node ast.Node, fallback wireSourceContext, assignments wireSourceAssignments) wireSourceContext {
	if node == nil {
		return fallback
	}

	if source, exists := assignments.sourceContexts[node]; exists {
		return source
	}

	return fallback
}

func isGeneratedWireBuiltin(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok || identifier.Obj != nil {
		return false
	}

	switch identifier.Name {
	case "make", builtinNewName, "len", "cap", builtinAppendName, "copy", builtinClearName, "delete":
		return true
	default:
		return false
	}
}

func wireCallerProvidedFunction(expression ast.Expr, assignments wireSourceAssignments, visiting map[wireSourceVariable]bool) bool {
	switch value := expression.(type) {
	case *ast.ParenExpr:
		return wireCallerProvidedFunction(value.X, assignments, visiting)
	case *ast.IndexExpr:
		return wireCallerProvidedFunction(value.X, assignments, visiting)
	case *ast.IndexListExpr:
		return wireCallerProvidedFunction(value.X, assignments, visiting)
	case *ast.Ident:
		if value.Obj == nil {
			return false
		}

		field, parameter := value.Obj.Decl.(*ast.Field)
		if !parameter || !isFunctionField(field) {
			return false
		}

		key := wireSourceVariable{declaration: field, name: value.Name}
		if visiting[key] {
			return false
		}

		visiting[key] = true
		defer delete(visiting, key)

		values := wireAliasExpressions(value, assignments)
		if wireFunctionParameter(field, assignments) && !wireCallerOwnedParameter(field, assignments) &&
			!assignments.boundParameters[key] {
			values = append(values, assignments.callArguments[key]...)
		}

		if len(values) == 0 {
			return functionFieldIsParameter(field, assignments) && wireCallerOwnedParameter(field, assignments)
		}

		for _, source := range values {
			if wireCallerProvidedFunction(source, assignments, visiting) {
				return true
			}
		}
	}

	return false
}

func isFunctionField(field *ast.Field) bool {
	_, ok := field.Type.(*ast.FuncType)

	return ok
}

func functionFieldIsParameter(field *ast.Field, assignments wireSourceAssignments) bool {
	for function := range assignments.functionBodies {
		if function.Params == nil {
			continue
		}

		for _, candidate := range function.Params.List {
			if candidate == field {
				return true
			}
		}
	}

	return false
}
