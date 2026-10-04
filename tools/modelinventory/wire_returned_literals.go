package main

import (
	"go/ast"
	"maps"
)

// Package function order does not constrain execution. Inspect each resolved
// helper's returns, with a body guard for recursive and mutually recursive calls.
func inspectWireReturnedLiterals(
	call *ast.CallExpr, visiting map[wireSourceVariable]bool, report func(*ast.BasicLit), assignments wireSourceAssignments,
) bool {
	resolved := false

	for _, function := range wireLocalHelpers(call.Fun, assignments, make(map[wireSourceVariable]bool)) {
		if function == nil {
			continue
		}

		body := assignments.functionBodies[function]
		resolved = resolved || body != nil

		key := wireSourceVariable{declaration: body, name: "returned literal"}

		if body == nil || visiting[key] {
			continue
		}

		visiting[key] = true
		bound := wireReturnParameterValues(function, call, assignments)

		ast.Inspect(body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}

			if returned, valid := node.(*ast.ReturnStmt); valid {
				if len(returned.Results) != 0 {
					inspectWireValueLiterals(returned.Results[0], visiting, report, bound)
				}

				return false
			}

			return true
		})
	}

	return resolved
}

func wireReturnParameterValues(function *ast.FuncType, call *ast.CallExpr, assignments wireSourceAssignments) wireSourceAssignments {
	assignments.values = maps.Clone(assignments.values)
	if function.Params == nil {
		return assignments
	}

	index := 0

	for _, field := range function.Params.List {
		for _, name := range field.Names {
			if index < len(call.Args) {
				key := wireSourceVariable{declaration: field, name: name.Name}
				assignments.values[key] = []ast.Expr{call.Args[index]}
			}

			index++
		}
	}

	return assignments
}
