package main

import "go/ast"

type wireSourceVariable struct {
	declaration ast.Node
	name        string
}

type wireSourceAssignments map[wireSourceVariable][]ast.Expr

func indexWireSourceAssignments(file *ast.File) wireSourceAssignments {
	assignments := make(wireSourceAssignments)

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
				assignments[key] = append(assignments[key], assignment.Rhs[index])
			}
		}

		return true
	})

	return assignments
}
