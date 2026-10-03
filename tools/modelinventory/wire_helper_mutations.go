package main

import "go/ast"

// Local helper parameters retain generated receiver provenance, so moving a
// mutation into a helper cannot turn a library-defined wire map into open input.
func indexWireHelperParameters(
	file *ast.File, aliases map[string]bool, path string, models map[string]generatedModel, assignments wireSourceAssignments,
) {
	for {
		before := len(assignments.generatedParameters)

		ast.Inspect(file, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall {
				return true
			}

			for _, function := range wireLocalHelpers(call.Fun, assignments, make(map[wireSourceVariable]bool)) {
				if function.Type.Params == nil {
					continue
				}

				index := 0

				for _, field := range function.Type.Params.List {
					for _, name := range field.Names {
						if index < len(call.Args) && generatedWireReceiver(
							call.Args[index], aliases, path, models, make(map[wireSourceVariable]bool), assignments,
						) {
							assignments.generatedParameters[wireSourceVariable{declaration: field, name: name.Name}] = true
						}

						index++
					}
				}
			}

			return true
		})

		if before == len(assignments.generatedParameters) {
			return
		}
	}
}

func wireLocalHelpers(expression ast.Expr, assignments wireSourceAssignments, visiting map[wireSourceVariable]bool) []*ast.FuncDecl {
	switch expression := expression.(type) {
	case *ast.IndexExpr:
		return wireLocalHelpers(expression.X, assignments, visiting)
	case *ast.IndexListExpr:
		return wireLocalHelpers(expression.X, assignments, visiting)
	case *ast.ParenExpr:
		return wireLocalHelpers(expression.X, assignments, visiting)
	}

	identifier, isIdentifier := expression.(*ast.Ident)
	if !isIdentifier || identifier.Obj == nil {
		return nil
	}

	if function, isFunction := identifier.Obj.Decl.(*ast.FuncDecl); isFunction {
		return []*ast.FuncDecl{function}
	}

	declaration, isNode := identifier.Obj.Decl.(ast.Node)

	key := wireSourceVariable{declaration: declaration, name: identifier.Name}

	if !isNode || visiting[key] {
		return nil
	}

	visiting[key] = true

	var functions []*ast.FuncDecl

	for _, value := range wireAliasExpressions(identifier, assignments) {
		functions = append(functions, wireLocalHelpers(value, assignments, visiting)...)
	}

	return functions
}

func inspectWireCopyCall(
	file *ast.File, call *ast.CallExpr, aliases map[string]bool, path string, models map[string]generatedModel,
	assignments wireSourceAssignments, report func(*ast.BasicLit),
) {
	if len(call.Args) != 2 || !isWireCopyFunction(file, call.Fun, assignments, make(map[wireSourceVariable]bool)) ||
		!generatedWireReceiver(call.Args[0], aliases, path, models, make(map[wireSourceVariable]bool), assignments) {
		return
	}

	inspectWireReceiverKeys(call.Args[0], make(map[wireSourceVariable]bool), report, assignments)
	inspectWireValueLiterals(call.Args[1], make(map[wireSourceVariable]bool), report, assignments)
}

func isWireCopyFunction(file *ast.File, expression ast.Expr, assignments wireSourceAssignments, visiting map[wireSourceVariable]bool) bool {
	switch expression := expression.(type) {
	case *ast.IndexExpr:
		return isWireCopyFunction(file, expression.X, assignments, visiting)
	case *ast.IndexListExpr:
		return isWireCopyFunction(file, expression.X, assignments, visiting)
	case *ast.ParenExpr:
		return isWireCopyFunction(file, expression.X, assignments, visiting)
	}

	if identifier, isIdentifier := expression.(*ast.Ident); isIdentifier {
		if identifier.Obj == nil {
			return identifier.Name == "copy"
		}

		declaration, isNode := identifier.Obj.Decl.(ast.Node)

		key := wireSourceVariable{declaration: declaration, name: identifier.Name}

		if !isNode || visiting[key] {
			return false
		}

		visiting[key] = true
		for _, value := range wireAliasExpressions(identifier, assignments) {
			if isWireCopyFunction(file, value, assignments, visiting) {
				return true
			}
		}

		return false
	}

	selector, isSelector := expression.(*ast.SelectorExpr)
	if !isSelector || selector.Sel.Name != "Copy" {
		return false
	}

	identifier, isIdentifier := selector.X.(*ast.Ident)
	if !isIdentifier || identifier.Obj != nil {
		return false
	}

	owner := primitivePackage{ImportPath: "maps", Name: "maps", Directory: ""}

	return primitiveImportAliases(file, owner)[identifier.Name]
}
