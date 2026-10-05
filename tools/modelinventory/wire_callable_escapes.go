package main

import (
	"go/ast"
	"strings"
)

const jsonEncodeMethod = "Encode"
const jsonUnmarshalMethod = "Unmarshal"

// Unknown callbacks cannot be inspected for fixed wire keys or values. Require
// a local body, generated constructor, or an import-resolved read-only codec.
func unverifiedWireCallable(
	file *ast.File, call *ast.CallExpr, aliases map[string]bool, path string,
	models map[string]generatedModel, assignments wireSourceAssignments,
) bool {
	if verifiedLocalWireCallable(call.Fun, assignments) ||
		isWireCopyFunction(file, call.Fun, assignments, make(map[wireSourceVariable]bool)) ||
		isGeneratedWireConstruction(call.Fun, aliases, path, models) ||
		verifiedWireCallable(file, call.Fun, assignments) || verifiedSDKDependencyCall(path, call.Fun, assignments) ||
		verifiedGeneratedGraphQLCall(file, path, call.Fun) {
		return false
	}

	for _, argument := range call.Args {
		if generatedWireReceiver(argument, aliases, path, models, make(map[wireSourceVariable]bool), assignments) {
			return true
		}
	}

	return false
}

// The native genqlient operation file is separately checked against exact SDL
// selections and generation drift. Its imported Client performs the inventoried
// request handoff through the SDK's injected GraphQL transport.
func verifiedGeneratedGraphQLCall(file *ast.File, path string, expression ast.Expr) bool {
	if path != "pkg/dependencies/graphql/generated.go" {
		return false
	}

	selector, selected := expression.(*ast.SelectorExpr)
	if !selected || selector.Sel.Name != "MakeRequest" {
		return false
	}

	receiver, named := selector.X.(*ast.Ident)
	if !named || receiver.Obj == nil {
		return false
	}

	parameter, declared := receiver.Obj.Decl.(*ast.Field)
	if !declared {
		return false
	}

	client, typed := parameter.Type.(*ast.SelectorExpr)
	if !typed || client.Sel.Name != "Client" {
		return false
	}

	qualifier, imported := client.X.(*ast.Ident)
	owner := primitivePackage{ImportPath: "github.com/Khan/genqlient/graphql", Name: "graphql", Directory: ""}

	return imported && qualifier.Obj == nil && primitiveImportAliases(file, owner)[qualifier.Name]
}

func verifiedWireCallable(file *ast.File, expression ast.Expr, assignments wireSourceAssignments) bool {
	if array, conversion := expression.(*ast.ArrayType); conversion {
		element, primitive := array.Elt.(*ast.Ident)

		return primitive && element.Name == "byte"
	}

	if identifier, plain := expression.(*ast.Ident); plain {
		if identifier.Obj != nil {
			_, conversion := identifier.Obj.Decl.(*ast.TypeSpec)

			return conversion
		}

		switch identifier.Name {
		case "len", "cap", "append", builtinDeleteName, "clear", schemaStringType, "bool", "int", "int32", "int64", "float32", "float64":
			return true
		}
	}

	selector, selected := expression.(*ast.SelectorExpr)
	if !selected {
		return false
	}

	if selector.Sel.Name == "After" || selector.Sel.Name == "Before" || selector.Sel.Name == "Equal" {
		return verifiedTimeValue(file, selector.X)
	}

	if selector.Sel.Name == "Decode" || selector.Sel.Name == jsonEncodeMethod {
		return verifiedJSONCodec(file, selector.X, assignments, make(map[wireSourceVariable]bool))
	}

	return verifiedImportedWireCallable(file, selector)
}
func verifiedImportedWireCallable(file *ast.File, selector *ast.SelectorExpr) bool {
	identifier, imported := selector.X.(*ast.Ident)
	if !imported || identifier.Obj != nil {
		return false
	}

	stringsOwner := primitivePackage{ImportPath: "strings", Name: "strings", Directory: ""}
	if primitiveImportAliases(file, stringsOwner)[identifier.Name] {
		switch selector.Sel.Name {
		case "TrimPrefix", "TrimSpace", "TrimRight", "ToLower", "NewReader", "Join", "HasPrefix", "Contains":
			return true
		}
	}

	htmlOwner := primitivePackage{ImportPath: "html", Name: "html", Directory: ""}
	if primitiveImportAliases(file, htmlOwner)[identifier.Name] {
		return selector.Sel.Name == "EscapeString"
	}

	timeOwner := primitivePackage{ImportPath: "time", Name: "time", Directory: ""}
	if primitiveImportAliases(file, timeOwner)[identifier.Name] {
		return selector.Sel.Name == "Parse" || selector.Sel.Name == "ParseInLocation"
	}

	formatOwner := primitivePackage{ImportPath: "fmt", Name: "fmt", Directory: ""}
	if primitiveImportAliases(file, formatOwner)[identifier.Name] {
		switch selector.Sel.Name {
		case "Sprint", "Sprintf", "Sprintln", "Errorf":
			return true
		}
	}

	owner := primitivePackage{ImportPath: "encoding/json", Name: "json", Directory: ""}
	if primitiveImportAliases(file, owner)[identifier.Name] {
		return selector.Sel.Name == "Marshal" || selector.Sel.Name == "MarshalIndent" || selector.Sel.Name == jsonUnmarshalMethod
	}

	return false
}

// The concrete SDK REST/GraphQL clients are checked separately at their typed
// generated-model boundaries. Callback fields and arbitrary imported mutators
// are not part of these package-private dependencies.
func verifiedSDKDependencyCall(path string, expression ast.Expr, assignments wireSourceAssignments) bool {
	if !strings.HasPrefix(path, "pkg/alexa/") {
		return false
	}

	call, selected := expression.(*ast.SelectorExpr)
	if !selected {
		return false
	}

	field, selected := call.X.(*ast.SelectorExpr)
	if !selected {
		return false
	}

	identifier, plain := field.X.(*ast.Ident)
	if !plain || identifier.Obj == nil {
		return false
	}

	receiver, parameter := identifier.Obj.Decl.(*ast.Field)
	if !parameter {
		return false
	}

	pointer, typed := receiver.Type.(*ast.StarExpr)
	if !typed {
		return false
	}

	name, typed := pointer.X.(*ast.Ident)

	return typed && name.Obj != nil && name.Obj.Decl == assignments.sdkOwner && assignments.sdkDependencies[field.Sel.Name]
}

func indexSDKDependencyFields(file *ast.File) (map[string]bool, *ast.TypeSpec) {
	var owner *ast.TypeSpec

	verified := make(map[string]bool)

	for _, declaration := range file.Decls {
		group, typed := declaration.(*ast.GenDecl)
		if !typed {
			continue
		}

		for _, spec := range group.Specs {
			model, typed := spec.(*ast.TypeSpec)
			if !typed || model.Name.Name != "Session" {
				continue
			}

			object, typed := model.Type.(*ast.StructType)
			if typed {
				indexConcreteSDKDependencies(file, object, verified)

				owner = model
			}
		}
	}

	return verified, owner
}
func indexConcreteSDKDependencies(file *ast.File, object *ast.StructType, verified map[string]bool) {
	for _, field := range object.Fields.List {
		pointer, typed := field.Type.(*ast.StarExpr)
		if !typed {
			continue
		}

		selector, typed := pointer.X.(*ast.SelectorExpr)
		if !typed || selector.Sel.Name != "Client" {
			continue
		}

		imported, typed := selector.X.(*ast.Ident)
		if !typed {
			continue
		}

		for _, dependency := range []string{"rest", "graphql"} {
			owner := primitivePackage{ImportPath: "github.com/portpowered/go-alexa/pkg/dependencies/" + dependency, Name: dependency, Directory: ""}
			if primitiveImportAliases(file, owner)[imported.Name] {
				for _, name := range field.Names {
					verified[name.Name] = true
				}
			}
		}
	}
}

func verifiedJSONCodec(file *ast.File, expression ast.Expr, assignments wireSourceAssignments, visiting map[wireSourceVariable]bool) bool {
	if identifier, named := expression.(*ast.Ident); named && identifier.Obj != nil {
		declaration, valid := identifier.Obj.Decl.(ast.Node)

		key := wireSourceVariable{declaration: declaration, name: identifier.Name}

		if !valid || visiting[key] {
			return false
		}

		visiting[key] = true
		for _, value := range wireAliasExpressions(identifier, assignments) {
			if verifiedJSONCodec(file, value, assignments, visiting) {
				return true
			}
		}
	}

	call, called := expression.(*ast.CallExpr)
	if !called {
		return false
	}

	selector, selected := call.Fun.(*ast.SelectorExpr)
	if !selected || selector.Sel.Name != "NewDecoder" && selector.Sel.Name != "NewEncoder" {
		return false
	}

	identifier, imported := selector.X.(*ast.Ident)
	if !imported || identifier.Obj != nil {
		return false
	}

	owner := primitivePackage{ImportPath: "encoding/json", Name: "json", Directory: ""}

	return primitiveImportAliases(file, owner)[identifier.Name]
}

func verifiedTimeValue(file *ast.File, expression ast.Expr) bool {
	call, called := expression.(*ast.CallExpr)
	if !called {
		return false
	}

	selector, selected := call.Fun.(*ast.SelectorExpr)
	if !selected {
		return false
	}

	if selector.Sel.Name == "Add" || selector.Sel.Name == "AddDate" {
		return verifiedTimeValue(file, selector.X)
	}

	identifier, imported := selector.X.(*ast.Ident)
	if !imported || identifier.Obj != nil || selector.Sel.Name != "Now" {
		return false
	}

	owner := primitivePackage{ImportPath: "time", Name: "time", Directory: ""}

	return primitiveImportAliases(file, owner)[identifier.Name]
}

func verifiedLocalWireCallable(expression ast.Expr, assignments wireSourceAssignments) bool {
	functions := wireLocalHelpers(expression, assignments, make(map[wireSourceVariable]bool))
	if len(functions) == 0 {
		return false
	}

	for _, function := range functions {
		if function == nil {
			return false
		}
	}

	return true
}
