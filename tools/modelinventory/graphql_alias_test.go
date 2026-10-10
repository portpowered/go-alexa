package main

import "testing"

func TestSelectedSchemaFieldResolvesAliasWithoutInventingFields(t *testing.T) {
	t.Parallel()

	schema := mustLoadTestGraphQLSchema(t, `type Query { volume: Volume } type Volume { value: Int }`)
	operation := mustParseTestGraphQLOperation(t, `query Device { volume { volumeValue: value } }`)

	field := selectedGraphQLSchemaField(operation.SelectionSet, schema, "Volume", []string{"volume"}, "volumeValue")
	if field == nil || field.Name != "value" {
		t.Fatal("response alias did not resolve to the SDL source field")
	}

	missing := selectedGraphQLSchemaField(operation.SelectionSet, schema, "Volume", []string{"volume"}, "unselectedAlias")
	if missing != nil {
		t.Fatal("unselected field was invented")
	}

	invalid := mustParseTestGraphQLOperation(t, `query Device { volume { volumeValue: missing } }`)

	missing = selectedGraphQLSchemaField(invalid.SelectionSet, schema, "Volume", []string{"volume"}, "volumeValue")
	if missing != nil {
		t.Fatal("invalid aliased source field was accepted")
	}
}
