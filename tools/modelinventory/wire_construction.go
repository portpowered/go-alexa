package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
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

func checkGeneratedWireConstructionRoot(root string, generated map[string]bool, models map[string]generatedModel) error {
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk wire construction sources: %w", walkErr)
		}

		path = filepath.ToSlash(path)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || generated[path] {
			return nil
		}

		set := token.NewFileSet()

		file, parseErr := parser.ParseFile(set, path, mustRead(path), 0)
		if parseErr != nil {
			return fmt.Errorf("parse wire construction source %s: %w", path, parseErr)
		}

		return rejectRawGeneratedWireConstructions(file, set, path, models)
	})
	if err != nil {
		return fmt.Errorf("check generated wire construction: %w", err)
	}

	return nil
}

func rejectRawGeneratedWireConstructions(file *ast.File, set *token.FileSet, path string, models map[string]generatedModel) error {
	owner := primitivePackage{ImportPath: wireImportPath, Name: "alexamodels", Directory: "pkg/dependencymodels"}
	aliases := primitiveImportAliases(file, owner)
	assignments := indexWireSourceAssignments(file)

	var problems []string

	ast.Inspect(file, func(node ast.Node) bool {
		if assignment, isAssignment := node.(*ast.AssignStmt); isAssignment {
			for index, left := range assignment.Lhs {
				receiver, key := wireAssignmentParts(left)
				if receiver == nil || index >= len(assignment.Rhs) ||
					!generatedWireReceiver(receiver, aliases, path, models, make(map[ast.Node]bool)) {
					continue
				}

				report := func(raw *ast.BasicLit) {
					decoded, err := strconv.Unquote(raw.Value)
					if err == nil && decoded != "" {
						problems = append(problems, fmt.Sprintf("%s: fixed wire assignment %q must be generated from schema", set.Position(raw.Pos()), decoded))
					}
				}
				inspectWireValueLiterals(key, make(map[wireSourceVariable]bool), report, assignments)
				inspectWireValueLiterals(assignment.Rhs[index], make(map[wireSourceVariable]bool), report, assignments)
			}
		}

		literal, isComposite := node.(*ast.CompositeLit)
		if !isComposite || !isGeneratedWireConstruction(literal.Type, aliases, path, models) {
			return true
		}

		for _, element := range literal.Elts {
			value := element
			if field, isField := element.(*ast.KeyValueExpr); isField {
				value = field.Value
			}

			inspectWireValueLiterals(value, make(map[wireSourceVariable]bool), func(raw *ast.BasicLit) {
				decoded, err := strconv.Unquote(raw.Value)
				if err == nil && decoded != "" {
					problems = append(problems, fmt.Sprintf("%s: fixed wire value %q must be generated from schema", set.Position(raw.Pos()), decoded))
				}
			}, assignments)
		}

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

	for _, value := range assignments[key] {
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
	expression ast.Expr, aliases map[string]bool, path string, models map[string]generatedModel, visiting map[ast.Node]bool,
) bool {
	switch expression := expression.(type) {
	case *ast.SelectorExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting)
	case *ast.UnaryExpr:
		return generatedWireReceiver(expression.X, aliases, path, models, visiting)
	case *ast.CompositeLit:
		return isGeneratedWireConstruction(expression.Type, aliases, path, models)
	case *ast.Ident:
		return generatedWireIdentifier(expression, aliases, path, models, visiting)
	}

	return false
}

func generatedWireIdentifier(
	identifier *ast.Ident, aliases map[string]bool, path string, models map[string]generatedModel, visiting map[ast.Node]bool,
) bool {
	if identifier.Obj == nil {
		return false
	}

	declaration, isNode := identifier.Obj.Decl.(ast.Node)
	if !isNode || visiting[declaration] {
		return false
	}

	visiting[declaration] = true

	switch declaration := declaration.(type) {
	case *ast.Field:
		return generatedWireDeclaredType(declaration.Type, aliases, path, models)
	case *ast.ValueSpec:
		if generatedWireDeclaredType(declaration.Type, aliases, path, models) {
			return true
		}

		for index, name := range declaration.Names {
			if name.Name == identifier.Name && index < len(declaration.Values) {
				return generatedWireReceiver(declaration.Values[index], aliases, path, models, visiting)
			}
		}
	case *ast.AssignStmt:
		for index, left := range declaration.Lhs {
			name, isIdentifier := left.(*ast.Ident)
			if isIdentifier && name.Name == identifier.Name && index < len(declaration.Rhs) {
				return generatedWireReceiver(declaration.Rhs[index], aliases, path, models, visiting)
			}
		}
	}

	return false
}

func generatedWireDeclaredType(expression ast.Expr, aliases map[string]bool, path string, models map[string]generatedModel) bool {
	if pointer, isPointer := expression.(*ast.StarExpr); isPointer {
		expression = pointer.X
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
