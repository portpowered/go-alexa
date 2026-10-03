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

const sdkPrimitiveOutput = "pkg/alexaapimodels/wire_constants.gen.go"

type sdkPrimitive struct {
	Schema    string
	Component string
	Value     string
	Type      string
}

func readSDKPrimitiveSchemas() (map[string]sdkPrimitive, error) {
	primitives := make(map[string]sdkPrimitive)
	seen := make(map[string]bool)

	graphql, err := readGraphQLSchema()
	if err != nil {
		return nil, err
	}

	for _, set := range openAPIGeneratedSets() {
		if seen[set.Schema] {
			continue
		}

		seen[set.Schema] = true

		document, err := readSchemaDocument(set.Schema)
		if err != nil {
			return nil, err
		}

		err = checkEventFieldBindings(set.Schema, document.Components.Schemas, graphql)
		if err != nil {
			return nil, err
		}

		for component, schema := range document.Components.Schemas {
			err = checkOpenGraphQLBindings(component, schema, graphql)
			if err != nil {
				return nil, err
			}

			err = checkSDKGraphQLBindings(schema, graphql)
			if err != nil {
				return nil, err
			}

			err = collectSDKPrimitiveSchemas(set.Schema, component, schema, primitives)
			if err != nil {
				return nil, err
			}
		}
	}

	return primitives, nil
}

func checkSDKPrimitiveInventory() ([]inventoryRow, error) {
	primitives, err := readSDKPrimitiveSchemas()
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(token.NewFileSet(), sdkPrimitiveOutput, mustRead(sdkPrimitiveOutput), parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse SDK constant projections: %w", err)
	}

	if !isGeneratedFile(sdkPrimitiveOutput) {
		return nil, diagnosticf("SDK constant projections lack a generated marker")
	}

	err = verifySDKPrimitiveDeclarations(file, primitives)
	if err != nil {
		return nil, err
	}

	err = scanSDKPrimitiveSources("pkg/alexaapimodels")
	if err != nil {
		return nil, err
	}

	err = checkRawWirePrimitiveValues()
	if err != nil {
		return nil, err
	}

	usages, err := findPrimitiveCallSites(primitives, primitivePackage{ImportPath: sdkImportPath, Name: "alexaapimodels", Directory: "pkg/alexaapimodels"})
	if err != nil {
		return nil, err
	}

	typeRows, err := inventorySDKPrimitiveTypes(primitives)
	if err != nil {
		return nil, err
	}

	rows := make([]inventoryRow, 0, len(primitives))
	for name, primitive := range primitives {
		rows = append(rows, inventoryRow{
			Schema: primitive.Schema, Component: primitive.Component + "." + name,
			GoType:     "alexaapimodels." + name + " (" + primitive.Type + " constant)",
			Definition: sdkPrimitiveOutput, Generator: "go run ./tools/wireconstants; SDK semantic projection from shared schema values",
			CallSites: usages[name],
		})
	}

	return append(rows, typeRows...), nil
}

func checkRawWirePrimitiveValues() error {
	values, err := generatedWirePrimitiveValues()
	if err != nil {
		return err
	}

	generated := registeredGeneratedFiles()
	// Route constants are regenerated and drift-checked by tools/apiroutes.
	generated["pkg/internal/apiroutes/routes.gen.go"] = true

	return rejectRawWirePrimitiveValues("pkg", values, generated)
}

func generatedWirePrimitiveValues() (map[string]bool, error) {
	values := make(map[string]bool)
	files := registeredGeneratedFiles()

	files["pkg/dependencymodels/wire_constants.gen.go"] = true

	for path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, mustRead(path), 0)
		if err != nil {
			return nil, fmt.Errorf("parse generated wire values %s: %w", path, err)
		}

		collectGeneratedStringValues(file, values)
	}

	return values, nil
}

func collectGeneratedStringValues(file *ast.File, values map[string]bool) {
	for _, declaration := range file.Decls {
		group, isGroup := declaration.(*ast.GenDecl)
		if !isGroup || group.Tok != token.CONST {
			continue
		}

		ast.Inspect(group, func(node ast.Node) bool {
			literal, isLiteral := node.(*ast.BasicLit)
			if !isLiteral || literal.Kind != token.STRING {
				return true
			}

			value, err := strconv.Unquote(literal.Value)
			if err == nil && value != "" {
				values[value] = true
			}

			return true
		})
	}
}

func rejectRawWirePrimitiveValues(root string, values map[string]bool, generated map[string]bool) error {
	var problems []string

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk wire primitive sources: %w", walkErr)
		}

		path = filepath.ToSlash(path)
		// The generic fixture recorder accepts arbitrary caller protocols and does not construct provider requests.
		if entry.IsDir() && path == "pkg/testing" {
			return filepath.SkipDir
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			generated[path] || path == "pkg/dependencymodels/wire_constants.gen.go" {
			return nil
		}

		set := token.NewFileSet()

		file, err := parser.ParseFile(set, path, mustRead(path), 0)
		if err != nil {
			return fmt.Errorf("parse wire primitive source %s: %w", path, err)
		}

		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}

			value, err := strconv.Unquote(literal.Value)
			if err == nil && values[value] {
				problems = append(problems, fmt.Sprintf("%s: raw schema-known wire value %q must use a generated constant", set.Position(literal.Pos()), value))
			}

			return true
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("scan wire primitive values: %w", err)
	}

	if len(problems) != 0 {
		return diagnosticf("%s", strings.Join(problems, "\n"))
	}

	return nil
}

func verifySDKPrimitiveDeclarations(file *ast.File, primitives map[string]sdkPrimitive) error {
	seen := make(map[string]bool)

	for _, declaration := range file.Decls {
		group, isGroup := declaration.(*ast.GenDecl)
		if !isGroup || group.Tok != token.CONST {
			return diagnosticf("SDK projection contains a non-constant declaration")
		}

		for _, specification := range group.Specs {
			entry, isValue := specification.(*ast.ValueSpec)
			if !isValue || len(entry.Names) != 1 || len(entry.Values) != 1 {
				return diagnosticf("SDK projection must declare each constant explicitly")
			}

			name := entry.Names[0].Name

			primitive, exists := primitives[name]

			if !exists || seen[name] {
				return diagnosticf("SDK projection contains unregistered or duplicate constant %s", name)
			}

			err := verifySDKPrimitiveValue(entry, primitive)
			if err != nil {
				return err
			}

			seen[name] = true
		}
	}

	if len(seen) != len(primitives) {
		return diagnosticf("SDK projection omits schema constants")
	}

	return nil
}

func verifySDKPrimitiveValue(entry *ast.ValueSpec, primitive sdkPrimitive) error {
	name := entry.Names[0].Name

	literal, isLiteral := entry.Values[0].(*ast.BasicLit)

	if !isLiteral || literal.Kind != token.STRING {
		return diagnosticf("SDK projection %s is not a schema string value", name)
	}

	decoded, err := strconv.Unquote(literal.Value)
	if err != nil {
		return fmt.Errorf("decode SDK constant %s: %w", name, err)
	}

	if decoded != primitive.Value {
		return diagnosticf("SDK projection %s differs from its schema value", name)
	}

	actualType := ""

	if entry.Type != nil {
		identifier, isIdentifier := entry.Type.(*ast.Ident)
		if !isIdentifier {
			return diagnosticf("SDK projection %s has an unsupported type", name)
		}

		actualType = identifier.Name
	}

	if actualType != primitive.Type {
		return diagnosticf("SDK projection %s differs from its schema Go type", name)
	}

	return nil
}

func scanSDKPrimitiveSources(root string) error {
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk SDK primitives: %w", walkErr)
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || filepath.ToSlash(path) == sdkPrimitiveOutput {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, mustRead(path), 0)
		if err != nil {
			return fmt.Errorf("parse SDK primitive source %s: %w", path, err)
		}

		return rejectHandwrittenSDKConstants(file, path)
	})
	if err != nil {
		return fmt.Errorf("check SDK primitive sources: %w", err)
	}

	return nil
}

func rejectHandwrittenSDKConstants(file *ast.File, path string) error {
	for _, declaration := range file.Decls {
		group, isGroup := declaration.(*ast.GenDecl)
		if !isGroup || group.Tok != token.CONST {
			continue
		}

		foundString := false

		ast.Inspect(group, func(node ast.Node) bool {
			literal, isLiteral := node.(*ast.BasicLit)
			if isLiteral && literal.Kind == token.STRING {
				foundString = true
			}

			return true
		})

		if foundString {
			return diagnosticf("%s: handwritten SDK string primitive must be generated from schema", path)
		}
	}

	return nil
}

func collectSDKPrimitiveSchemas(path, component string, schema schemaNode, primitives map[string]sdkPrimitive) error {
	for name, value := range schema.CompatibilityConstants {
		if !strings.HasPrefix(name, "SDK") {
			continue
		}

		publicName := strings.TrimPrefix(name, "SDK")
		if _, exists := primitives[publicName]; exists {
			return diagnosticf("SDK primitive %s has duplicate schema owners", publicName)
		}

		primitives[publicName] = sdkPrimitive{Schema: path, Component: component, Value: value, Type: schema.SDKType}
	}

	for name, property := range schema.Properties {
		err := collectSDKPrimitiveSchemas(path, component+"."+name, property, primitives)
		if err != nil {
			return err
		}
	}

	return nil
}
