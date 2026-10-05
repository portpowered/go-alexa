package main

import "go/ast"

type wireSourceVariable struct {
	declaration ast.Node
	name        string
}

type wireSourceContext struct {
	file *ast.File
	path string
}

type wireSourceAssignments struct {
	values              map[wireSourceVariable][]ast.Expr
	callArguments       map[wireSourceVariable][]ast.Expr
	boundParameters     map[wireSourceVariable]bool
	generatedParameters map[wireSourceVariable]bool
	globalVariables     map[wireSourceVariable]bool
	sdkDependencies     map[string]bool
	sdkOwner            *ast.TypeSpec
	callableFields      map[string][]ast.Expr
	functionBodies      map[*ast.FuncType]*ast.BlockStmt
	functionDecls       map[*ast.FuncType]*ast.FuncDecl
	methods             map[string][]*ast.FuncDecl
	sourceContexts      map[ast.Node]wireSourceContext
	resultIndexes       map[wireSourceVariable]map[ast.Expr]int
	resultSelections    map[*ast.CallExpr]int
}

//nolint:cyclop,funlen,gocognit // One package-local AST pass indexes aliases, contexts, declarations, and tuple results together.
func indexWireSourceAssignments(file *ast.File, paths ...string) wireSourceAssignments {
	path := ""
	if len(paths) != 0 {
		path = paths[0]
	}

	assignments := wireSourceAssignments{
		values:              make(map[wireSourceVariable][]ast.Expr),
		callArguments:       make(map[wireSourceVariable][]ast.Expr),
		boundParameters:     make(map[wireSourceVariable]bool),
		generatedParameters: make(map[wireSourceVariable]bool),
		globalVariables:     make(map[wireSourceVariable]bool),
		sdkDependencies:     make(map[string]bool),
		sdkOwner:            nil,
		callableFields:      make(map[string][]ast.Expr),
		functionBodies:      make(map[*ast.FuncType]*ast.BlockStmt),
		functionDecls:       make(map[*ast.FuncType]*ast.FuncDecl),
		methods:             make(map[string][]*ast.FuncDecl),
		sourceContexts:      make(map[ast.Node]wireSourceContext),
		resultIndexes:       make(map[wireSourceVariable]map[ast.Expr]int),
		resultSelections:    make(map[*ast.CallExpr]int),
	}

	assignments.sdkDependencies, assignments.sdkOwner = indexSDKDependencyFields(file)

	for name, object := range file.Scope.Objects {
		if declaration, variable := object.Decl.(*ast.ValueSpec); variable && object.Kind == ast.Var {
			assignments.globalVariables[wireSourceVariable{declaration: declaration, name: name}] = true
		}
	}

	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv != nil {
			assignments.methods[function.Name.Name] = append(assignments.methods[function.Name.Name], function)
		}
	}

	ast.Inspect(file, func(node ast.Node) bool {
		if node != nil {
			assignments.sourceContexts[node] = wireSourceContext{file: file, path: path}
		}

		switch function := node.(type) {
		case *ast.FuncDecl:
			assignments.functionBodies[function.Type] = function.Body
			assignments.functionDecls[function.Type] = function
		case *ast.FuncLit:
			assignments.functionBodies[function.Type] = function.Body
		}

		if entry, field := node.(*ast.KeyValueExpr); field {
			if name, named := entry.Key.(*ast.Ident); named {
				assignments.callableFields[name.Name] = append(assignments.callableFields[name.Name], entry.Value)
			}
		}

		if declaration, valueSpec := node.(*ast.ValueSpec); valueSpec {
			if len(declaration.Values) == 1 {
				if _, call := declaration.Values[0].(*ast.CallExpr); call && len(declaration.Names) > 1 {
					for index, name := range declaration.Names {
						key := wireSourceVariable{declaration: declaration, name: name.Name}
						assignWireResultIndex(assignments.resultIndexes, key, declaration.Values[0], index)
					}
				}
			}
		}

		assignment, isAssignment := node.(*ast.AssignStmt)
		if !isAssignment {
			return true
		}

		for index, left := range assignment.Lhs {
			identifier, isIdentifier := left.(*ast.Ident)
			if !isIdentifier || identifier.Obj == nil {
				continue
			}

			if index >= len(assignment.Rhs) && (len(assignment.Rhs) != 1 || len(assignment.Lhs) == 1) {
				continue
			}

			declaration, isNode := identifier.Obj.Decl.(ast.Node)
			if isNode {
				key := wireSourceVariable{declaration: declaration, name: identifier.Name}

				rhsIndex := index
				if rhsIndex >= len(assignment.Rhs) {
					rhsIndex = 0
				}

				assignments.values[key] = append(assignments.values[key], assignment.Rhs[rhsIndex])

				if len(assignment.Rhs) == 1 && len(assignment.Lhs) > 1 {
					_, call := assignment.Rhs[0].(*ast.CallExpr)
					if call {
						assignWireResultIndex(assignments.resultIndexes, key, assignment.Rhs[0], index)
					}
				}
			}
		}

		return true
	})

	return assignments
}

func assignWireResultIndex(indexes map[wireSourceVariable]map[ast.Expr]int, key wireSourceVariable, expression ast.Expr, index int) {
	if indexes[key] == nil {
		indexes[key] = make(map[ast.Expr]int)
	}

	indexes[key][expression] = index
}

func wireAliasExpressions(identifier *ast.Ident, assignments wireSourceAssignments) []ast.Expr {
	if identifier.Obj == nil {
		return nil
	}

	declaration, isNode := identifier.Obj.Decl.(ast.Node)
	if !isNode {
		return nil
	}

	var values []ast.Expr

	if value := wireAliasValue(identifier.Name, declaration); value != nil {
		values = append(values, value)
	}

	key := wireSourceVariable{declaration: declaration, name: identifier.Name}
	for _, value := range assignments.values[key] {
		_, parameter := declaration.(*ast.Field)
		if parameter || assignments.globalVariables[key] || value.Pos() < identifier.Pos() {
			values = append(values, value)
		}
	}

	return values
}
