package main

import "testing"

func TestOpenEnumBindingRejectsMissingValuesAndPreservesFutureStrings(t *testing.T) {
	t.Parallel()

	graphql, err := readGraphQLSchema()
	if err != nil {
		t.Fatalf("read SDL: %v", err)
	}

	schema := emptySchemaNode()
	schema.Type = schemaStringType
	schema.GoType = schemaStringType
	schema.GraphQLEnum = "PowerStateValue"
	schema.ExtensibleEnum = []string{"ON", "OFF", "UNKNOWN"}

	err = checkOpenGraphQLBindings("EventKnownPowerStateValue", schema, graphql)
	if err != nil {
		t.Fatalf("valid open enum binding: %v", err)
	}

	schema.ExtensibleEnum = []string{"ON", "OFF"}

	err = checkOpenGraphQLBindings("EventKnownPowerStateValue", schema, graphql)
	if err == nil {
		t.Fatal("missing known state accepted")
	}

	schema.ExtensibleEnum = []string{"ON", "OFF", "FUTURE"}

	err = checkOpenGraphQLBindings("EventKnownPowerStateValue", schema, graphql)
	if err == nil {
		t.Fatal("unregistered known state accepted")
	}

	schema.GraphQLEnum = ""

	err = checkOpenGraphQLBindings("EventKnownPowerStateValue", schema, graphql)
	if err == nil {
		t.Fatal("missing enum binding accepted")
	}
}

func TestLegacyCapabilityBindingRejectsMissingShapeAndLostOpenFields(t *testing.T) {
	t.Parallel()

	graphql, err := readGraphQLSchema()
	if err != nil {
		t.Fatalf("read SDL: %v", err)
	}

	schema := emptySchemaNode()
	schema.GraphQLJSONField = "LegacyAppliance.capabilities"
	schema.AdditionalProperties = true
	property := emptySchemaNode()
	property.Type = schemaStringType
	schema.Properties = map[string]schemaNode{"interfaceName": property}

	err = checkOpenGraphQLBindings("LegacyCapabilityPayload", schema, graphql)
	if err != nil {
		t.Fatalf("valid JSON scalar binding: %v", err)
	}

	schema.AdditionalProperties = false

	err = checkOpenGraphQLBindings("LegacyCapabilityPayload", schema, graphql)
	if err == nil {
		t.Fatal("lost open provider fields accepted")
	}

	schema.AdditionalProperties = true
	schema.Properties = nil

	err = checkOpenGraphQLBindings("LegacyCapabilityPayload", schema, graphql)
	if err == nil {
		t.Fatal("missing known interface field accepted")
	}
}
