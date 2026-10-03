package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

func checkGeneratedWireConstructions(models map[string]generatedModel) error {
	generated := registeredGeneratedFiles()
	generated[wirePrimitiveOutput] = true
	generated["pkg/internal/apiroutes/routes.gen.go"] = true

	err := checkGeneratedWireConstructionRoot("pkg", generated, models)
	if err == nil {
		err = checkGeneratedWireConstructionRoot("cmd", generated, models)
	}

	return err
}

func rejectRawGeneratedWireConstructionsWithAssignments(
	file *ast.File, set *token.FileSet, path string, models map[string]generatedModel, assignments wireSourceAssignments,
) error {
	owner := primitivePackage{ImportPath: wireImportPath, Name: "alexamodels", Directory: "pkg/dependencymodels"}
	aliases := primitiveImportAliases(file, owner)

	var problems []string

	ast.Inspect(file, func(node ast.Node) bool {
		if call, isCall := node.(*ast.CallExpr); isCall {
			inspectWireCopyCall(file, call, aliases, path, models, assignments, func(raw *ast.BasicLit) {
				decoded, err := strconv.Unquote(raw.Value)
				if err == nil && decoded != "" {
					problems = append(problems, fmt.Sprintf("%s: fixed wire copy %q must be generated from schema", set.Position(raw.Pos()), decoded))
				}
			})
		}

		if assignment, isAssignment := node.(*ast.AssignStmt); isAssignment {
			inspectWireMutationAssignment(assignment, aliases, path, models, assignments, set, &problems)
		}

		literal, isComposite := node.(*ast.CompositeLit)
		if !isComposite || !generatedWireDeclaredType(literal.Type, aliases, path, models) {
			return true
		}

		inspectWireValueLiterals(literal, make(map[wireSourceVariable]bool), func(raw *ast.BasicLit) {
			decoded, err := strconv.Unquote(raw.Value)
			if err == nil && decoded != "" {
				problems = append(problems, fmt.Sprintf("%s: fixed wire value %q must be generated from schema", set.Position(raw.Pos()), decoded))
			}
		}, assignments)

		return false
	})

	if len(problems) != 0 {
		return diagnosticf("%s", strings.Join(problems, "\n"))
	}

	return nil
}

func isGeneratedWireConstruction(expression ast.Expr, aliases map[string]bool, path string, models map[string]generatedModel) bool {
	name := ""

	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		identifier, isIdentifier := expression.X.(*ast.Ident)
		if isIdentifier && identifier.Obj == nil && aliases[identifier.Name] {
			name = expression.Sel.Name
		}
	case *ast.Ident:
		if filepath.ToSlash(filepath.Dir(path)) == "pkg/dependencymodels" {
			name = expression.Name
		}
	}

	model, exists := models[name]

	return exists && strings.HasPrefix(model.File, "pkg/dependencymodels/")
}

// Follow source-local aliases as well as direct literals. Parameters and
// generated constants have no local value declaration and remain caller inputs.
func inspectWireValueLiterals(expression ast.Expr, visiting map[wireSourceVariable]bool, report func(*ast.BasicLit), assignments wireSourceAssignments) {
	if expression == nil {
		return
	}

	ast.Inspect(expression, func(node ast.Node) bool {
		if literal, isComposite := node.(*ast.CompositeLit); isComposite {
			if _, isMap := literal.Type.(*ast.MapType); isMap {
				for _, element := range literal.Elts {
					if entry, isEntry := element.(*ast.KeyValueExpr); isEntry {
						inspectWireValueLiterals(entry.Key, visiting, report, assignments)
						inspectWireValueLiterals(entry.Value, visiting, report, assignments)
					}
				}

				return false
			}
		}

		if field, isField := node.(*ast.KeyValueExpr); isField {
			if key, isLiteral := field.Key.(*ast.BasicLit); isLiteral && key.Kind == token.STRING {
				report(key)
			}

			inspectWireValueLiterals(field.Value, visiting, report, assignments)

			return false
		}

		if literal, isLiteral := node.(*ast.BasicLit); isLiteral && literal.Kind == token.STRING {
			report(literal)
		}

		if identifier, isIdentifier := node.(*ast.Ident); isIdentifier {
			inspectWireAlias(identifier, visiting, report, assignments)
		}

		return true
	})
}

func inspectWireAlias(identifier *ast.Ident, visiting map[wireSourceVariable]bool, report func(*ast.BasicLit), assignments wireSourceAssignments) {
	if identifier.Obj == nil {
		return
	}

	declaration, isNode := identifier.Obj.Decl.(ast.Node)

	key := wireSourceVariable{declaration: declaration, name: identifier.Name}

	if !isNode || visiting[key] {
		return
	}

	visiting[key] = true

	if value := wireAliasValue(identifier.Name, declaration); value != nil {
		inspectWireValueLiterals(value, visiting, report, assignments)
	}

	for _, value := range assignments.values[key] {
		if value.Pos() < identifier.Pos() {
			inspectWireValueLiterals(value, visiting, report, assignments)
		}
	}
}

func wireAliasValue(name string, declaration ast.Node) ast.Expr {
	switch declaration := declaration.(type) {
	case *ast.ValueSpec:
		for index, identifier := range declaration.Names {
			if identifier.Name == name && index < len(declaration.Values) {
				return declaration.Values[index]
			}
		}
	case *ast.AssignStmt:
		for index, left := range declaration.Lhs {
			identifier, isIdentifier := left.(*ast.Ident)
			if isIdentifier && identifier.Name == name && index < len(declaration.Rhs) {
				return declaration.Rhs[index]
			}
		}
	}

	return nil
}

func generatedWireReceiver(
	expression ast.Expr, aliases map[string]bool, path string, models map[string]generatedModel,
	visiting map[wireSourceVariable]bool, assignments wireSourceAssignments,
) bool {
	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.UnaryExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.StarExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.IndexExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.SliceExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.TypeAssertExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.ParenExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting, assignments)
	case *ast.CompositeLit:
		return generatedWireDeclaredType(expression.Type, aliases, path, models)
	case *ast.Ident:
		return generatedWireIdentifier(expression, aliases, path, models, visiting, assignments)
	}

	return false
}

func generatedWireIdentifier(
	identifier *ast.Ident, aliases map[string]bool, path string, models map[string]generatedModel,
	visiting map[wireSourceVariable]bool, assignments wireSourceAssignments,
) bool {
	if identifier.Obj == nil {
		return false
	}

	declaration, isNode := identifier.Obj.Decl.(ast.Node)

	key := wireSourceVariable{declaration: declaration, name: identifier.Name}

	if !isNode || visiting[key] {
		return false
	}

	visiting[key] = true

	if assignments.generatedParameters[key] {
		return true
	}

	switch declaration := declaration.(type) {
	case *ast.Field:
		return generatedWireDeclaredType(declaration.Type, aliases, path, models)
	case *ast.ValueSpec:
		if generatedWireDeclaredType(declaration.Type, aliases, path, models) {
			return true
		}
	}

	if value := wireAliasValue(identifier.Name, declaration); value != nil &&
		generatedWireReceiver(value, aliases, path, models, visiting, assignments) {
		return true
	}

	for _, value := range assignments.values[key] {
		if value.Pos() < identifier.Pos() && generatedWireReceiver(value, aliases, path, models, visiting, assignments) {
			return true
		}
	}

	return false
}

func generatedWireDeclaredType(expression ast.Expr, aliases map[string]bool, path string, models map[string]generatedModel) bool {
	switch expression := expression.(type) {
	case *ast.StarExpr:
		return generatedWireDeclaredType(expression.X, aliases, path, models)
	case *ast.ArrayType:
		return generatedWireDeclaredType(expression.Elt, aliases, path, models)
	case *ast.MapType:
		return generatedWireDeclaredType(expression.Value, aliases, path, models)
	case *ast.ParenExpr:
		return generatedWireDeclaredType(expression.X, aliases, path, models)
	}

	return isGeneratedWireConstruction(expression, aliases, path, models)
}

func wireAssignmentParts(expression ast.Expr) (ast.Expr, ast.Expr) {
	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		return expression.X, nil
	case *ast.IndexExpr:
		return expression.X, expression.Index
	}

	return nil, nil
}

func inspectWireMutationAssignment(
	assignment *ast.AssignStmt, aliases map[string]bool, path string, models map[string]generatedModel,
	assignments wireSourceAssignments, set *token.FileSet, problems *[]string,
) {
	for index, left := range assignment.Lhs {
		receiver, key := wireAssignmentParts(left)
		if receiver == nil || index >= len(assignment.Rhs) ||
			!generatedWireReceiver(receiver, aliases, path, models, make(map[wireSourceVariable]bool), assignments) {
			continue
		}

		report := func(raw *ast.BasicLit) {
			decoded, err := strconv.Unquote(raw.Value)
			if err == nil && decoded != "" {
				*problems = append(*problems, fmt.Sprintf("%s: fixed wire assignment %q must be generated from schema", set.Position(raw.Pos()), decoded))
			}
		}
		inspectWireReceiverKeys(receiver, make(map[wireSourceVariable]bool), report, assignments)
		inspectWireValueLiterals(key, make(map[wireSourceVariable]bool), report, assignments)
		inspectWireValueLiterals(assignment.Rhs[index], make(map[wireSourceVariable]bool), report, assignments)
	}
}
