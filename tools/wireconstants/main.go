// Command wireconstants generates untyped compatibility constants from schema enums.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	outputPath                         = "pkg/dependencymodels/wire_constants.gen.go"
	sdkOutputPath                      = "pkg/alexaapimodels/wire_constants.gen.go"
	outputPermissions    os.FileMode   = 0o600
	directoryPermissions os.FileMode   = 0o750
	errInvalidMetadata   metadataError = "invalid schema constant metadata"
)

type metadataError string

func (err metadataError) Error() string { return string(err) }

func schemaPaths() []string {
	return []string{
		"api/asyncapi.yaml",
		"api/behaviors.yaml",
		"api/compat/devices.yaml",
		"api/compat/auth.yaml",
		"api/compat/endpoints.yaml",
		"api/compat/events.yaml",
		"api/feature-events.yaml",
		"api/openapi/sources/endpoints.yaml",
		"api/openapi/sources/account-linking.yaml",
		"api/openapi/sources/first-party-devices.yaml",
		"api/openapi/sources/media.yaml",
	}
}

type schemaDocument struct {
	Components struct {
		Schemas    map[string]schemaComponent    `yaml:"schemas"`
		Parameters map[string]parameterComponent `yaml:"parameters"`
	} `yaml:"components"`
}

type schemaComponent struct {
	Enum                   []string                   `yaml:"enum"`
	Default                string                     `yaml:"default"`
	CompatibilityConstant  []string                   `yaml:"x-go-compat-constant-names"`  //nolint:tagliatelle // Preserve the checked-in OpenAPI extension name.
	CompatibilityName      string                     `yaml:"x-go-compat-constant-name"`   //nolint:tagliatelle // Preserve the checked-in OpenAPI extension name.
	CompatibilityConstants map[string]string          `yaml:"x-go-compat-constant-values"` //nolint:tagliatelle // Preserve the checked-in OpenAPI extension name.
	CompatibilityType      string                     `yaml:"x-go-compat-constant-type"`   //nolint:tagliatelle // Preserve the checked-in OpenAPI extension name.
	SDKType                string                     `yaml:"x-go-sdk-constant-type"`      //nolint:tagliatelle // Preserve the checked-in schema extension name.
	WireKeyName            string                     `yaml:"x-go-wire-key-constant-name"` //nolint:tagliatelle // Preserve the checked-in schema extension name.
	Properties             map[string]schemaComponent `yaml:"properties"`
}

type parameterComponent struct {
	Schema schemaComponent `yaml:"schema"`
}

type compatibilityConstant struct {
	Name  string
	Value string
	Type  string
}

func main() {
	contents, err := readSchemas(schemaPaths())
	if err != nil {
		fatalf("read schemas: %v", err)
	}

	generated, err := renderCompatibilityConstants(contents)
	if err != nil {
		fatalf("generate constants: %v", err)
	}

	err = writeGenerated(outputPath, generated)
	if err != nil {
		fatalf("write %s: %v", outputPath, err)
	}

	sdkGenerated, err := renderSDKConstants(contents)
	if err != nil {
		fatalf("generate SDK constant projections: %v", err)
	}

	err = writeGenerated(sdkOutputPath, sdkGenerated)
	if err != nil {
		fatalf("write %s: %v", sdkOutputPath, err)
	}
}

func readSchemas(paths []string) ([][]byte, error) {
	contents := make([][]byte, 0, len(paths))

	for _, schemaPath := range paths {
		data, err := os.ReadFile(schemaPath) //nolint:gosec // Paths are the fixed repository schema inputs returned by schemaPaths.
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", schemaPath, err)
		}

		contents = append(contents, data)
	}

	return contents, nil
}

func renderCompatibilityConstants(schemaContents [][]byte) ([]byte, error) {
	constants := make([]compatibilityConstant, 0)
	names := make(map[string]struct{})

	for schemaIndex, contents := range schemaContents {
		var document schemaDocument

		err := yaml.Unmarshal(contents, &document)
		if err != nil {
			return nil, fmt.Errorf("decode schema %d: %w", schemaIndex, err)
		}

		schemaNames := make([]string, 0, len(document.Components.Schemas))
		for name := range document.Components.Schemas {
			schemaNames = append(schemaNames, name)
		}

		sort.Strings(schemaNames)

		for _, componentName := range schemaNames {
			err = collectSchemaConstants(componentName, document.Components.Schemas[componentName], &constants, names)
			if err != nil {
				return nil, err
			}
		}

		parameterNames := make([]string, 0, len(document.Components.Parameters))
		for name := range document.Components.Parameters {
			parameterNames = append(parameterNames, name)
		}

		sort.Strings(parameterNames)

		for _, parameterName := range parameterNames {
			err = collectSchemaConstants("parameter "+parameterName, document.Components.Parameters[parameterName].Schema, &constants, names)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(constants) == 0 {
		return nil, fmt.Errorf("%w: schemas define no compatibility constants", errInvalidMetadata)
	}

	return renderConstants(constants, "alexamodels")
}

func renderConstants(constants []compatibilityConstant, packageName string) ([]byte, error) {
	sort.Slice(constants, func(left, right int) bool {
		return constants[left].Name < constants[right].Name
	})

	var source bytes.Buffer

	source.WriteString("// Code generated by go run ./tools/wireconstants from OpenAPI schema enum metadata; DO NOT EDIT.\n")
	source.WriteString("package " + packageName + "\n\nconst (\n")

	for _, constant := range constants {
		declarationName := constant.Name
		if constant.Type != "" {
			declarationName += " " + constant.Type
		}

		_, err := fmt.Fprintf(&source, "\t%s = %s\n", declarationName, strconv.Quote(constant.Value))
		if err != nil {
			return nil, fmt.Errorf("write constant %s: %w", constant.Name, err)
		}
	}

	source.WriteString(")\n")

	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go source: %w", err)
	}

	return formatted, nil
}

func collectSchemaConstants(path string, schema schemaComponent, constants *[]compatibilityConstant, names map[string]struct{}) error {
	if len(schema.CompatibilityConstant) > 0 {
		if len(schema.Enum) == 0 || len(schema.Enum) != len(schema.CompatibilityConstant) {
			return fmt.Errorf("%w: schema component %s has %d enum values and %d compatibility constant names",
				errInvalidMetadata, path, len(schema.Enum), len(schema.CompatibilityConstant))
		}

		for index, constantName := range schema.CompatibilityConstant {
			err := addCompatibilityConstant(path, constantName, schema.Enum[index], schema.CompatibilityType, constants, names)
			if err != nil {
				return err
			}
		}
	}

	if schema.CompatibilityName != "" {
		if schema.Default == "" {
			return fmt.Errorf("%w: schema component %s names compatibility constant %s but has no default value", errInvalidMetadata, path, schema.CompatibilityName)
		}

		err := addCompatibilityConstant(path, schema.CompatibilityName, schema.Default, schema.CompatibilityType, constants, names)
		if err != nil {
			return err
		}
	}

	constantNames := make([]string, 0, len(schema.CompatibilityConstants))
	for name := range schema.CompatibilityConstants {
		constantNames = append(constantNames, name)
	}

	sort.Strings(constantNames)

	for _, constantName := range constantNames {
		err := addCompatibilityConstant(path, constantName, schema.CompatibilityConstants[constantName], schema.CompatibilityType, constants, names)
		if err != nil {
			return err
		}
	}

	propertyNames := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		propertyNames = append(propertyNames, name)
	}

	sort.Strings(propertyNames)

	for _, propertyName := range propertyNames {
		property := schema.Properties[propertyName]

		propertyPath := path + "." + propertyName
		if property.WireKeyName != "" {
			err := addCompatibilityConstant(propertyPath, property.WireKeyName, propertyName, "", constants, names)
			if err != nil {
				return err
			}
		}

		err := collectSchemaConstants(propertyPath, property, constants, names)
		if err != nil {
			return err
		}
	}

	return nil
}

func addCompatibilityConstant(path, constantName, value, constantType string, constants *[]compatibilityConstant, names map[string]struct{}) error {
	if !isGoIdentifier(constantName) {
		return fmt.Errorf("%w: schema component %s has invalid Go constant name %q", errInvalidMetadata, path, constantName)
	}

	if _, exists := names[constantName]; exists {
		return fmt.Errorf("%w: duplicate compatibility constant name %s", errInvalidMetadata, constantName)
	}

	if constantType != "" && !isGoIdentifier(constantType) {
		return fmt.Errorf("%w: schema component %s has invalid Go constant type %q", errInvalidMetadata, path, constantType)
	}

	names[constantName] = struct{}{}

	*constants = append(*constants, compatibilityConstant{Name: constantName, Value: value, Type: constantType})

	return nil
}

func isGoIdentifier(value string) bool {
	if value == "" || strings.ContainsRune(value, '-') {
		return false
	}

	for index, char := range value {
		if index == 0 {
			if char != '_' && (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') {
				return false
			}

			continue
		}

		if char != '_' && (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return false
		}
	}

	return true
}

func fatalf(formatString string, values ...any) {
	_, _ = fmt.Fprintf(os.Stderr, formatString+"\n", values...)
	os.Exit(1)
}

func writeGenerated(path string, contents []byte) error {
	directory := filepath.Dir(path)

	err := os.MkdirAll(directory, directoryPermissions)
	if err != nil {
		return fmt.Errorf("create %s: %w", directory, err)
	}

	err = os.WriteFile(path, contents, outputPermissions)
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}
