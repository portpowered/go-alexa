package main

import (
	"fmt"
	"go/parser"
	"go/token"
)

const wirePrimitiveOutput = "pkg/dependencymodels/wire_constants.gen.go"

func checkWirePrimitiveInventory() ([]inventoryRow, error) {
	primitives, err := readWirePrimitiveSchemas()
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(token.NewFileSet(), wirePrimitiveOutput, mustRead(wirePrimitiveOutput), parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse wire constant projections: %w", err)
	}

	err = verifySDKPrimitiveDeclarations(file, primitives)
	if err != nil {
		return nil, err
	}

	usages, err := findPrimitiveCallSites(primitives, primitivePackage{ImportPath: wireImportPath, Name: "alexamodels", Directory: "pkg/dependencymodels"})
	if err != nil {
		return nil, err
	}

	rows := make([]inventoryRow, 0, len(primitives))
	for name, primitive := range primitives {
		rows = append(rows, inventoryRow{
			Schema: primitive.Schema, Component: primitive.Component + "." + name,
			GoType: "alexamodels." + name + " (constant)", Definition: wirePrimitiveOutput,
			Generator: "go run ./tools/wireconstants; checked schema constant projection",
			CallSites: usages[name],
		})
	}

	return rows, nil
}

func readWirePrimitiveSchemas() (map[string]sdkPrimitive, error) {
	primitives := make(map[string]sdkPrimitive)
	seen := make(map[string]bool)

	for _, set := range generatedSets() {
		if seen[set.Schema] {
			continue
		}

		seen[set.Schema] = true

		document, err := readSchemaDocument(set.Schema)
		if err != nil {
			return nil, err
		}

		for name, schema := range document.Components.Schemas {
			err = collectWirePrimitiveSchema(set.Schema, name, schema, primitives)
			if err != nil {
				return nil, err
			}
		}

		for name, parameter := range document.Components.Parameters {
			err = collectWirePrimitiveSchema(set.Schema, "parameter "+name, parameter.Schema, primitives)
			if err != nil {
				return nil, err
			}
		}
	}

	return primitives, nil
}

func collectWirePrimitiveSchema(path, component string, schema schemaNode, primitives map[string]sdkPrimitive) error {
	values := make(map[string]string)
	for name, value := range schema.CompatibilityConstants {
		values[name] = value
	}

	if len(schema.CompatibilityNames) != len(schema.Enum) && len(schema.CompatibilityNames) != 0 {
		return diagnosticf("wire primitive component %s has incomplete enum names", component)
	}

	for index, name := range schema.CompatibilityNames {
		value, isString := schema.Enum[index].(string)
		if !isString {
			return diagnosticf("wire primitive component %s has a non-string enum", component)
		}

		values[name] = value
	}

	if schema.CompatibilityName != "" {
		values[schema.CompatibilityName] = schema.Default
	}

	for name, value := range values {
		err := addWirePrimitiveSchema(path, component, name, value, schema.CompatibilityType, primitives)
		if err != nil {
			return err
		}
	}

	return collectWirePropertyPrimitiveSchemas(path, component, schema, primitives)
}

func collectWirePropertyPrimitiveSchemas(path, component string, schema schemaNode, primitives map[string]sdkPrimitive) error {
	for name, property := range schema.Properties {
		child := component + "." + name
		if property.WireKeyName != "" {
			err := addWirePrimitiveSchema(path, child, property.WireKeyName, name, "", primitives)
			if err != nil {
				return err
			}
		}

		err := collectWirePrimitiveSchema(path, child, property, primitives)
		if err != nil {
			return err
		}
	}

	return nil
}

func addWirePrimitiveSchema(path, component, name, value, goType string, primitives map[string]sdkPrimitive) error {
	if _, exists := primitives[name]; exists {
		return diagnosticf("wire primitive %s has duplicate schema owners", name)
	}

	primitives[name] = sdkPrimitive{Schema: path, Component: component, Value: value, Type: goType}

	return nil
}
