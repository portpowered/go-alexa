// Command oapiallOfix preserves schema-defined Go embedding in compatibility
// models after oapi-codegen expands allOf into flattened fields. Unsupported
// composition and field collisions fail rather than silently changing Go APIs.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	schemaPath                       = "api/compat/controls.yaml"
	outputPath                       = "pkg/dependencymodels/controls_compat.gen.go"
	localReferencePrefix             = "#/components/schemas/"
	generatedFileMode    fs.FileMode = 0o600
)

var errComposition = errors.New("unsupported or inconsistent schema composition")

type schemaDocument struct {
	Components struct {
		Schemas map[string]schemaNode `yaml:"schemas"`
	} `yaml:"components"`
}

type schemaNode struct {
	Ref        string                `yaml:"$ref"`
	GoTypeName string                `yaml:"x-go-type-name"` //nolint:tagliatelle // Preserve oapi-codegen's schema extension spelling.
	Properties map[string]schemaNode `yaml:"properties"`
	AllOf      []schemaNode          `yaml:"allOf"`
}

func main() {
	err := rewrite()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rewrite() error {
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read composition schema: %w", err)
	}

	source, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("read generated models: %w", err)
	}

	output, err := transform(schemaData, source)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, output, generatedFileMode)
	if err != nil {
		return fmt.Errorf("write composed generated models: %w", err)
	}

	return nil
}

func transform(schemaData, source []byte) ([]byte, error) {
	var document schemaDocument

	err := yaml.Unmarshal(schemaData, &document)
	if err != nil {
		return nil, fmt.Errorf("parse composition schema: %w", err)
	}

	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, outputPath, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse generated models: %w", err)
	}

	models := collectModels(file)
	components := make([]string, 0, len(document.Components.Schemas))

	for name := range document.Components.Schemas {
		components = append(components, name)
	}

	sort.Strings(components)

	for _, name := range components {
		definition := document.Components.Schemas[name]
		if len(definition.AllOf) == 0 {
			continue
		}

		err = composeModel(name, definition, document.Components.Schemas, models)
		if err != nil {
			return nil, err
		}
	}

	var output bytes.Buffer

	err = format.Node(&output, fileSet, file)
	if err != nil {
		return nil, fmt.Errorf("format composed generated models: %w", err)
	}

	return output.Bytes(), nil
}

func collectModels(file *ast.File) map[string]*ast.StructType {
	models := make(map[string]*ast.StructType)

	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.TYPE {
			continue
		}

		for _, spec := range group.Specs {
			model, isType := spec.(*ast.TypeSpec)
			if !isType {
				continue
			}

			fields, ok := model.Type.(*ast.StructType)
			if ok {
				models[model.Name.Name] = fields
			}
		}
	}

	return models
}

func goName(name string, definition schemaNode) string {
	if definition.GoTypeName != "" {
		return definition.GoTypeName
	}

	return name
}

func composeModel(name string, definition schemaNode, schemas map[string]schemaNode, models map[string]*ast.StructType) error {
	model, exists := models[goName(name, definition)]
	if !exists {
		return fmt.Errorf("%w: component %s has no generated struct", errComposition, name)
	}

	parentName, parent, err := resolveParent(definition.AllOf, schemas, models)
	if err != nil {
		return fmt.Errorf("compose %s: %w", name, err)
	}

	inherited := make(map[string]*ast.Field, len(parent.Fields.List))

	for _, field := range parent.Fields.List {
		property := jsonFieldName(field)
		if _, collision := definition.Properties[property]; collision {
			return fmt.Errorf("%w: component %s overrides inherited property %s", errComposition, name, property)
		}

		inherited[property] = field
	}

	retained, err := retainOwnFields(model, parentName, inherited, definition.Properties)
	if err != nil {
		return fmt.Errorf("compose %s: %w", name, err)
	}

	for _, field := range model.Fields.List {
		if len(field.Names) == 0 {
			return nil
		}
	}

	embedded := &ast.Field{
		Doc: nil, Names: nil, Type: ast.NewIdent(parentName), Tag: nil, Comment: nil,
	}
	model.Fields.List = append([]*ast.Field{embedded}, retained...)

	return nil
}

func resolveParent(references []schemaNode, schemas map[string]schemaNode, models map[string]*ast.StructType) (string, *ast.StructType, error) {
	if len(references) != 1 {
		return "", nil, fmt.Errorf("%w: expected one allOf parent reference", errComposition)
	}

	reference := references[0]
	if !strings.HasPrefix(reference.Ref, localReferencePrefix) || len(reference.Properties) != 0 || len(reference.AllOf) != 0 {
		return "", nil, fmt.Errorf("%w: allOf requires a local component reference", errComposition)
	}

	name := strings.TrimPrefix(reference.Ref, localReferencePrefix)

	definition, exists := schemas[name]
	if !exists || len(definition.AllOf) != 0 || len(definition.Properties) == 0 {
		return "", nil, fmt.Errorf("%w: parent %s must be a declared non-composed object", errComposition, name)
	}

	parentName := goName(name, definition)

	model, exists := models[parentName]
	if !exists {
		return "", nil, fmt.Errorf("%w: parent %s has no generated struct", errComposition, name)
	}

	err := checkParentFields(name, definition, model)
	if err != nil {
		return "", nil, err
	}

	return parentName, model, nil
}

func checkParentFields(name string, definition schemaNode, model *ast.StructType) error {
	fields := make(map[string]bool, len(model.Fields.List))

	for _, field := range model.Fields.List {
		property := jsonFieldName(field)
		if property == "" || fields[property] {
			return fmt.Errorf("%w: parent %s has untagged or duplicate fields", errComposition, name)
		}

		fields[property] = true
	}

	if len(fields) != len(definition.Properties) {
		return fmt.Errorf("%w: parent %s field count differs from schema", errComposition, name)
	}

	for property := range definition.Properties {
		if !fields[property] {
			return fmt.Errorf("%w: parent %s lacks schema property %s", errComposition, name, property)
		}
	}

	return nil
}

func retainOwnFields(model *ast.StructType, parentName string, inherited map[string]*ast.Field, own map[string]schemaNode) ([]*ast.Field, error) {
	retained := make([]*ast.Field, 0, len(own))
	seen := make(map[string]bool, len(inherited)+len(own))
	embedded := false

	for _, field := range model.Fields.List {
		if len(field.Names) == 0 {
			parent, ok := field.Type.(*ast.Ident)
			if !ok || parent.Name != parentName || embedded {
				return nil, fmt.Errorf("%w: unexpected embedded field", errComposition)
			}

			embedded = true

			continue
		}

		property := jsonFieldName(field)
		if property == "" || seen[property] {
			return nil, fmt.Errorf("%w: untagged or duplicate child field", errComposition)
		}

		seen[property] = true

		base, inheritedField := inherited[property]
		if inheritedField {
			if !sameField(field, base) {
				return nil, fmt.Errorf("%w: inherited property %s has a different Go declaration", errComposition, property)
			}

			continue
		}

		if _, declared := own[property]; !declared {
			return nil, fmt.Errorf("%w: child field %s is absent from schema", errComposition, property)
		}

		retained = append(retained, field)
	}

	if embedded {
		for property := range inherited {
			if seen[property] {
				return nil, fmt.Errorf("%w: child repeats an embedded field", errComposition)
			}

			seen[property] = true
		}
	}

	if len(seen) != len(inherited)+len(own) {
		return nil, fmt.Errorf("%w: generated child omits schema fields", errComposition)
	}

	return retained, nil
}

func sameField(field, base *ast.Field) bool {
	if len(field.Names) != 1 || len(base.Names) != 1 || field.Names[0].Name != base.Names[0].Name {
		return false
	}

	var actual, expected bytes.Buffer

	actualErr := format.Node(&actual, token.NewFileSet(), field.Type)
	expectedErr := format.Node(&expected, token.NewFileSet(), base.Type)

	return actualErr == nil && expectedErr == nil && bytes.Equal(actual.Bytes(), expected.Bytes()) && field.Tag.Value == base.Tag.Value
}

func jsonFieldName(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}

	tag, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}

	return strings.Split(reflect.StructTag(tag).Get("json"), ",")[0]
}
