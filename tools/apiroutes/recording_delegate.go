package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

const (
	wireRecordingDelegate = "cmd/go-alexa/internal/cli/wire_recording.go#RoundTrip"
	recordingSendMethod   = "RoundTrip"
)

// The recorder forwards an already schema-bound request through an injected
// transport. It may restore Body, but cannot change the route or add sends.
func checkWireRecordingDelegate(decl *ast.FuncDecl, fset *token.FileSet, imports packageImports, violations *[]string) {
	addViolation := func(node ast.Node) {
		*violations = append(*violations, fmt.Sprintf(
			"%s: recording transport must forward the original request through r.transport once", fset.Position(node.Pos()),
		))
	}
	if len(decl.Type.Params.List) != 1 || len(decl.Type.Params.List[0].Names) != 1 {
		addViolation(decl)

		return
	}

	parameter := decl.Type.Params.List[0].Names[0]
	sends := 0

	ast.Inspect(decl.Body, func(node ast.Node) bool {
		if recordingRequestMutation(node, parameter.Name) {
			addViolation(node)
		}

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		primitive := networkPrimitive(call, imports)
		if primitive != "" {
			sends++

			if !validRecordingSend(call, primitive, decl, parameter) {
				addViolation(call)
			}
		}

		if recordingRequestEscape(call, primitive, parameter.Name) {
			addViolation(call)
		}

		return true
	})

	if sends != 1 {
		addViolation(decl)
	}
}

func recordingBodyField(expression ast.Expr, parameter string) bool {
	field, ok := expression.(*ast.SelectorExpr)

	return ok && field.Sel.Name == "Body" && isIdentifier(field.X, parameter)
}

func recordingRequestMutation(node ast.Node, parameter string) bool {
	var values []ast.Expr

	switch declaration := node.(type) {
	case *ast.AssignStmt:
		for _, target := range declaration.Lhs {
			if rootIdentifier(target) == parameter && !recordingBodyField(target, parameter) {
				return true
			}
		}

		values = declaration.Rhs
	case *ast.ValueSpec:
		values = declaration.Values
	default:
		return false
	}

	for _, value := range values {
		if rootIdentifier(value) == parameter {
			return true
		}
	}

	return false
}

func validRecordingSend(call *ast.CallExpr, primitive string, decl *ast.FuncDecl, parameter *ast.Ident) bool {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || primitive != recordingSendMethod || len(call.Args) != 1 {
		return false
	}

	receiver, owner := injectedClientReceiver(selector.X)
	argument, isArgument := call.Args[0].(*ast.Ident)

	return receiver == "r.transport" && owner == receiverObject(decl) && isArgument && argument.Obj == parameter.Obj
}

func recordingRequestEscape(call *ast.CallExpr, primitive, parameter string) bool {
	for _, argument := range call.Args {
		if rootIdentifier(argument) == parameter && primitive != recordingSendMethod && !recordingBodyField(argument, parameter) {
			return true
		}
	}

	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector || rootIdentifier(selector.X) != parameter {
		return false
	}

	field, isField := selector.X.(*ast.SelectorExpr)
	if !isField || !isIdentifier(field.X, parameter) {
		return true
	}

	return (field.Sel.Name != "URL" || selector.Sel.Name != "String") && (field.Sel.Name != "Header" || selector.Sel.Name != "Clone")
}
