package main

import (
	"strings"

	graphqlast "github.com/vektah/gqlparser/v2/ast"
)

func checkEventFieldBindings(path string, components map[string]schemaNode, graphql *graphqlast.Schema) error {
	if path != "api/feature-events.yaml" {
		return nil
	}

	for name, schema := range components {
		if !strings.HasSuffix(name, "Property") {
			continue
		}

		if name == "SpeakerProperty" {
			continue // Implementation-derived flat volume/muted events have no direct SDL counterpart.
		}

		definition := graphql.Types[eventPropertyGraphQLType(name)]
		if definition == nil {
			return diagnosticf("event property %s lacks an SDL object mapping", name)
		}

		err := checkGraphQLWireFields(name, schema, definition, components, graphql)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkGraphQLWireFields(
	path string, schema schemaNode, definition *graphqlast.Definition,
	components map[string]schemaNode, graphql *graphqlast.Schema,
) error {
	for name, property := range schema.Properties {
		field := definition.Fields.ForName(name)
		if field == nil {
			continue
		}

		target := graphql.Types[field.Type.Name()]
		if target == nil {
			continue
		}

		if field.Type.Elem != nil && property.Items != nil {
			property = *property.Items
		}

		if target.Kind == graphqlast.Enum {
			bound := components[strings.TrimPrefix(property.Ref, "#/components/schemas/")]
			if bound.GraphQLEnum != target.Name {
				return diagnosticf("wire field %s.%s lacks its SDL enum binding %s", path, name, target.Name)
			}

			continue
		}

		if target.Kind == graphqlast.Object && len(property.Properties) > 0 {
			return diagnosticf("wire field %s.%s must reference a named component for SDL object %s", path, name, target.Name)
		}

		if property.Ref != "" {
			property = components[strings.TrimPrefix(property.Ref, "#/components/schemas/")]
		}

		if len(property.Properties) > 0 {
			err := checkGraphQLWireFields(path+"."+name, property, target, components, graphql)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func eventPropertyGraphQLType(name string) string {
	switch name {
	case "RangeProperty":
		return "RangeValue"
	case "ToggleProperty":
		return "ToggleState"
	case "SpeakerProperty":
		return "" // Flat volume/muted events have no direct SDL object counterpart.
	}

	return strings.TrimSuffix(name, "Property")
}
