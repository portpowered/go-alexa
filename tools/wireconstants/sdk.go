package main

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// renderSDKConstants projects schema values into the existing SDK semantic types.
// These declarations share schema inputs with the wire constants without creating
// a dependency cycle between compatibility wire models and SDK projections.
func renderSDKConstants(contents [][]byte) ([]byte, error) {
	constants := make([]compatibilityConstant, 0)
	names := make(map[string]struct{})

	for index, contents := range contents {
		var document schemaDocument

		err := yaml.Unmarshal(contents, &document)
		if err != nil {
			return nil, fmt.Errorf("decode SDK schema %d: %w", index, err)
		}

		for componentName, schema := range document.Components.Schemas {
			err = collectSDKSchemaConstants(componentName, schema, &constants, names)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(constants) == 0 {
		return nil, fmt.Errorf("%w: schemas define no SDK constant projections", errInvalidMetadata)
	}

	return renderConstants(constants, "alexaapimodels")
}

func collectSDKSchemaConstants(component string, schema schemaComponent, constants *[]compatibilityConstant, names map[string]struct{}) error {
	for name, value := range schema.CompatibilityConstants {
		if !strings.HasPrefix(name, "SDK") {
			continue
		}

		err := addCompatibilityConstant(component, strings.TrimPrefix(name, "SDK"), value, schema.SDKType, constants, names)
		if err != nil {
			return err
		}
	}

	for name, property := range schema.Properties {
		err := collectSDKSchemaConstants(component+"."+name, property, constants, names)
		if err != nil {
			return err
		}
	}

	return nil
}
