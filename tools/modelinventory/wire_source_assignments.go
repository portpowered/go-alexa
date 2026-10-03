package main

import "go/ast"

type wireSourceVariable struct {
	declaration ast.Node
	name        string
}

type wireSourceAssignments struct {
	values              map[wireSourceVariable][]ast.Expr
	generatedParameters map[wireSourceVariable]bool
}

func indexWireSourceAssignments(file *ast.File) wireSourceAssignments {
	assignments := wireSourceAssignments{
		values:              make(map[wireSourceVariable][]ast.Expr),
		generatedParameters: make(map[wireSourceVariable]bool),
	}

	ast.Inspect(file, func(node ast.Node) bool {
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
		if value.Pos() < identifier.Pos() {
			values = append(values, value)
		}
	}

	return values
}
