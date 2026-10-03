package main

import (
	"strings"

	graphqlast "github.com/vektah/gqlparser/v2/ast"
)

func checkOpenGraphQLBindings(component string, schema schemaNode, graphql *graphqlast.Schema) error {
	if strings.HasPrefix(component, "EventKnown") || strings.HasPrefix(component, "ControlKnown") {
		if schema.GraphQLEnum == "" {
			return diagnosticf("known string component %s lacks its GraphQL enum binding", component)
		}
	}

	if schema.GraphQLEnum != "" {
		err := checkOpenStringEnumBinding(component, schema, graphql)
		if err != nil {
			return err
		}
	}

	if component == "LegacyCapabilityPayload" {
		return checkLegacyCapabilityScalarBinding(schema, graphql)
	}

	return nil
}

func checkOpenStringEnumBinding(component string, schema schemaNode, graphql *graphqlast.Schema) error {
	definition := graphql.Types[schema.GraphQLEnum]
	if definition == nil || definition.Kind != graphqlast.Enum || schema.Type != schemaStringType || schema.GoType != schemaStringType || len(schema.Enum) != 0 {
		return diagnosticf("component %s must preserve an open string bound to a known SDL enum", component)
	}

	known := make(map[string]bool)
	for _, value := range schema.ExtensibleEnum {
		if known[value] || definition.EnumValues.ForName(value) == nil {
			return diagnosticf("component %s has a duplicate or unknown SDL enum value %s", component, value)
		}

		known[value] = true
	}

	if len(known) != len(definition.EnumValues) {
		return diagnosticf("component %s does not list every known SDL enum value", component)
	}

	return nil
}

func checkLegacyCapabilityScalarBinding(schema schemaNode, graphql *graphqlast.Schema) error {
	definition := graphql.Types["LegacyAppliance"]
	if definition == nil || schema.GraphQLJSONField != "LegacyAppliance.capabilities" {
		return diagnosticf("legacy capability payload lacks its JSON scalar binding")
	}

	field := definition.Fields.ForName("capabilities")
	if field == nil || field.Type.Elem == nil || field.Type.Elem.NamedType != "JSON" {
		return diagnosticf("legacy capabilities no longer use the bound JSON scalar list")
	}

	property, exists := schema.Properties["interfaceName"]
	if !exists || property.Type != schemaStringType || schema.AdditionalProperties != true {
		return diagnosticf("legacy capability payload must type interfaceName and preserve open provider members")
	}

	return nil
}
