package main

import "go/ast"

type wireSourceVariable struct {
	declaration ast.Node
	name        string
}

type wireSourceAssignments struct {
	values              map[wireSourceVariable][]ast.Expr
	generatedParameters map[wireSourceVariable]bool
	sdkDependencies     map[string]bool
	sdkOwner            *ast.TypeSpec
	callableFields      map[string][]ast.Expr
	functionBodies      map[*ast.FuncType]*ast.BlockStmt
	methods             map[string][]*ast.FuncDecl
}

func indexWireSourceAssignments(file *ast.File) wireSourceAssignments {
	assignments := wireSourceAssignments{
		values:              make(map[wireSourceVariable][]ast.Expr),
		generatedParameters: make(map[wireSourceVariable]bool),
		sdkDependencies:     make(map[string]bool),
		sdkOwner:            nil,
		callableFields:      make(map[string][]ast.Expr),
		functionBodies:      make(map[*ast.FuncType]*ast.BlockStmt),
		methods:             make(map[string][]*ast.FuncDecl),
	}

	assignments.sdkDependencies, assignments.sdkOwner = indexSDKDependencyFields(file)

	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv != nil {
			assignments.methods[function.Name.Name] = append(assignments.methods[function.Name.Name], function)
		}
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch function := node.(type) {
		case *ast.FuncDecl:
			assignments.functionBodies[function.Type] = function.Body
		case *ast.FuncLit:
			assignments.functionBodies[function.Type] = function.Body
		}

		if entry, field := node.(*ast.KeyValueExpr); field {
			if name, named := entry.Key.(*ast.Ident); named {
				assignments.callableFields[name.Name] = append(assignments.callableFields[name.Name], entry.Value)
			}
		}

		assignment, isAssignment := node.(*ast.AssignStmt)
		if !isAssignment {
			return true
		}

		for index, left := range assignment.Lhs {
			identifier, isIdentifier := left.(*ast.Ident)
			if !isIdentifier || identifier.Obj == nil || index >= len(assignment.Rhs) {
				continue
			}

			declaration, isNode := identifier.Obj.Decl.(ast.Node)
			if isNode {
				key := wireSourceVariable{declaration: declaration, name: identifier.Name}
				assignments.values[key] = append(assignments.values[key], assignment.Rhs[index])
			}
		}

		return true
	})

	return assignments
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
		if parameter || value.Pos() < identifier.Pos() {
			values = append(values, value)
		}
	}

	return values
}
