package main

import "go/ast"

const builtinDeleteName = "delete"

func wireTypeConversion(expression ast.Expr) bool {
	switch expression := expression.(type) {
	case *ast.ParenExpr:
		return wireTypeConversion(expression.X)
	case *ast.MapType, *ast.ArrayType, *ast.InterfaceType:
		return true
	case *ast.Ident:
		if expression.Obj != nil {
			_, conversion := expression.Obj.Decl.(*ast.TypeSpec)

			return conversion
		}

		return expression.Name == "any" || expression.Name == schemaStringType
	}

	return false
}

func inspectWireBuiltinMutation(
	call *ast.CallExpr, aliases map[string]bool, path string, models map[string]generatedModel,
	assignments wireSourceAssignments, report func(*ast.BasicLit),
) {
	name, builtin := call.Fun.(*ast.Ident)
	if !builtin || name.Obj != nil || (name.Name != "append" && name.Name != builtinDeleteName) || len(call.Args) < 2 ||
		!generatedWireReceiver(call.Args[0], aliases, path, models, make(map[wireSourceVariable]bool), assignments) {
		return
	}

	for _, argument := range call.Args[1:] {
		inspectWireValueLiterals(argument, make(map[wireSourceVariable]bool), report, assignments)
	}
}
