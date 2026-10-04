package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	graphqlparser "github.com/vektah/gqlparser/v2"
	graphqlast "github.com/vektah/gqlparser/v2/ast"
	graphqlsyntax "github.com/vektah/gqlparser/v2/parser"
)

func TestGraphQLModelEvidenceCoversOperationsAndSelections(t *testing.T) {
	t.Parallel()

	goModels, err := readGraphQLGoModels()
	if err != nil {
		t.Fatal(err)
	}

	models := make(map[string]generatedModel, len(goModels))
	for name := range goModels {
		models[name] = generatedModel{
			Name: name, File: graphQLOutputPath, Schema: "", Generator: "",
			Fields: nil, Type: nil, Alias: false, Generated: false,
		}
	}

	evidence, err := graphqlModelEvidence(models)
	if err != nil {
		t.Fatal(err)
	}

	if len(evidence) == 0 {
		t.Fatal("GraphQL evidence is empty")
	}

	for name, model := range models {
		if model.File != graphQLOutputPath {
			continue
		}

		if len(evidence[name]) == 0 {
			t.Errorf("generated GraphQL model %s has no operation/selection evidence", name)
		}
	}

	const expectedType = "ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPower"
	for _, trace := range evidence[expectedType] {
		if trace.Operation == "ListEndpointsWithStates" &&
			trace.Document == "pkg/schemas/graphql/listEndpointsWithStates.graphql" &&
			strings.Contains(trace.SelectionPath, "listEndpoints.endpoints[].features[].properties[].{Power}.name") {
			return
		}
	}

	t.Fatalf("model %s lacks its source selection trace: %#v", expectedType, evidence[expectedType])
}

func TestGeneratedGraphQLOperationMustMatchSourceAndSDL(t *testing.T) {
	t.Parallel()

	const schemaSource = `
schema { query: Query }
interface Node { id: ID! }
type Device implements Node { id: ID!, value: String!, label: String! }
type Other implements Node { id: ID! }
type Query { node: Node! }
`

	schema := mustLoadTestGraphQLSchema(t, schemaSource)
	source := mustParseTestGraphQLOperation(t, `query Device { node { id ... on Device { value } } }`)

	validGenerated := mustParseTestGraphQLOperation(t, `query Device { node { __typename id ... on Device { value } } }`)

	validationErr := validateGeneratedGraphQLOperation(source, validGenerated, schema)
	if validationErr != nil {
		t.Fatalf("valid generated operation rejected: %v", validationErr)
	}

	tests := []struct {
		name      string
		generated string
		schema    string
	}{
		{
			name:      "missing source selection",
			generated: `query Device { node { __typename ... on Device { value } } }`,
			schema:    schemaSource,
		},
		{
			name:      "extra generated selection",
			generated: `query Device { node { __typename id label ... on Device { value } } }`,
			schema:    schemaSource,
		},
		{
			name:      "wrong case generated field",
			generated: `query Device { node { __typename id ... on Device { Value } } }`,
			schema:    schemaSource,
		},
		{
			name:      "SDL drift",
			generated: `query Device { node { __typename id ... on Device { value } } }`,
			schema: `
schema { query: Query }
interface Node { id: ID! }
type Device implements Node { id: ID! }
type Other implements Node { id: ID! }
type Query { node: Node! }
`,
		},
		{
			name:      "missing abstract discriminator",
			generated: `query Device { node { id ... on Device { value } } }`,
			schema:    schemaSource,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			generated := mustParseTestGraphQLOperation(t, test.generated)

			testSchema := mustLoadTestGraphQLSchema(t, test.schema)

			validationErr := validateGeneratedGraphQLOperation(source, generated, testSchema)
			if validationErr == nil {
				t.Fatal("invalid generated operation was accepted")
			}
		})
	}
}

func TestGraphQLInterfaceDecoderEvidenceMatchesSDLAndGeneratedModels(t *testing.T) {
	t.Parallel()

	const schemaSource = `
schema { query: Query }
interface Node { id: ID! }
type Device implements Node { id: ID! }
type Other implements Node { id: ID! }
type Query { node: Node! }
`

	schema := mustLoadTestGraphQLSchema(t, schemaSource)
	models := map[string]graphqlGoModel{
		"NodeModel": {
			Kind: "interface", Fields: nil, Implementors: []string{"DeviceModel", "OtherModel"},
			Comment: "GraphQL interface Node.", CustomJSON: false,
		},
		"DeviceModel": {
			Kind: "struct", Fields: nil, Implementors: nil,
			Comment: "GraphQL type Device.", CustomJSON: false,
		},
		"OtherModel": {
			Kind: "struct", Fields: nil, Implementors: nil,
			Comment: "GraphQL type Other.", CustomJSON: false,
		},
	}

	valid := graphqlInterfaceDecoderCases{
		"NodeModel": {
			Cases: map[string]graphqlInterfaceDecoderCase{
				"Device": {Model: "DeviceModel", JSONUnmarshal: true},
				"Other":  {Model: "OtherModel", JSONUnmarshal: true},
			},
			ReadsTypeName:      true,
			RejectsMissingType: true,
			RejectsUnknownType: true,
		},
	}

	validationErr := validateGraphQLInterfaceDecoderCases(models, schema, valid)
	if validationErr != nil {
		t.Fatalf("valid decoder evidence rejected: %v", validationErr)
	}

	assertGraphQLInterfaceDecoderFailures(t, models, valid, schemaSource)
}

func assertGraphQLInterfaceDecoderFailures(
	t *testing.T,
	models map[string]graphqlGoModel,
	valid graphqlInterfaceDecoderCases,
	schemaSource string,
) {
	t.Helper()

	tests := []struct {
		name   string
		mutate func(graphqlInterfaceDecoderCases)
		schema string
	}{
		{
			name: "missing concrete case",
			mutate: func(decoders graphqlInterfaceDecoderCases) {
				delete(decoders["NodeModel"].Cases, "Other")
			},
			schema: schemaSource,
		},
		{
			name: "extra concrete case",
			mutate: func(decoders graphqlInterfaceDecoderCases) {
				decoders["NodeModel"].Cases["Unknown"] = graphqlInterfaceDecoderCase{Model: "OtherModel", JSONUnmarshal: true}
			},
			schema: schemaSource,
		},
		{
			name: "wrong model for typename",
			mutate: func(decoders graphqlInterfaceDecoderCases) {
				decoders["NodeModel"].Cases["Device"] = graphqlInterfaceDecoderCase{Model: "OtherModel", JSONUnmarshal: true}
			},
			schema: schemaSource,
		},
		{
			name:   "SDL possible type drift",
			mutate: func(graphqlInterfaceDecoderCases) {},
			schema: `
schema { query: Query }
interface Node { id: ID! }
type Device implements Node { id: ID! }
type Other { id: ID! }
type Query { node: Node! }
`,
		},
		{
			name: "decoder does not read typename",
			mutate: func(decoders graphqlInterfaceDecoderCases) {
				decoders["NodeModel"] = graphqlInterfaceDecoder{
					Cases: decoders["NodeModel"].Cases, ReadsTypeName: false,
					RejectsMissingType: true, RejectsUnknownType: true,
				}
			},
			schema: schemaSource,
		},
		{
			name: "decoder does not unmarshal concrete model",
			mutate: func(decoders graphqlInterfaceDecoderCases) {
				decoders["NodeModel"].Cases["Device"] = graphqlInterfaceDecoderCase{Model: "DeviceModel", JSONUnmarshal: false}
			},
			schema: schemaSource,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			decoders := cloneGraphQLDecoderCases(valid)
			test.mutate(decoders)

			testSchema := mustLoadTestGraphQLSchema(t, test.schema)

			validationErr := validateGraphQLInterfaceDecoderCases(models, testSchema, decoders)
			if validationErr == nil {
				t.Fatal("invalid decoder evidence was accepted")
			}
		})
	}
}

func cloneGraphQLDecoderCases(source graphqlInterfaceDecoderCases) graphqlInterfaceDecoderCases {
	cloned := make(graphqlInterfaceDecoderCases, len(source))

	for name, decoder := range source {
		cases := make(map[string]graphqlInterfaceDecoderCase, len(decoder.Cases))
		for concreteType, model := range decoder.Cases {
			cases[concreteType] = model
		}

		decoder.Cases = cases
		cloned[name] = decoder
	}

	return cloned
}

func mustLoadTestGraphQLSchema(t *testing.T, input string) *graphqlast.Schema {
	t.Helper()

	schema, err := graphqlparser.LoadSchema(&graphqlast.Source{Name: "test.graphql", Input: input, BuiltIn: false})
	if err != nil {
		t.Fatalf("load test GraphQL schema: %v", err)
	}

	return schema
}

func mustParseTestGraphQLOperation(t *testing.T, input string) *graphqlast.OperationDefinition {
	t.Helper()

	document, err := graphqlsyntax.ParseQuery(&graphqlast.Source{Name: "test.graphql", Input: input, BuiltIn: false})
	if err != nil {
		t.Fatalf("parse test GraphQL operation: %v", err)
	}

	if len(document.Operations) != 1 {
		t.Fatalf("parsed %d operations, want one", len(document.Operations))
	}

	return document.Operations[0]
}

func TestGraphQLSelectionEvidenceRejectsUnselectedPath(t *testing.T) {
	t.Parallel()

	source := `query Device($id: ID!) { endpoint(id: $id) { id } }`

	document, err := graphqlsyntax.ParseQuery(&graphqlast.Source{Name: "synthetic.graphql", Input: source, BuiltIn: false})
	if err != nil {
		t.Fatal(err)
	}

	selection := document.Operations[0].SelectionSet
	if findSelectedGraphQLField(selection, []string{"endpoint", "id"}) == nil {
		t.Fatal("selected response field was not found")
	}

	if findSelectedGraphQLField(selection, []string{"endpoint", "friendlyName"}) != nil {
		t.Fatal("unselected response field was accepted")
	}
}

func TestGraphQLGoModelTraceUsesJSONFieldNames(t *testing.T) {
	t.Parallel()

	const source = "package models\ntype Model struct { EndpointID string `json:\"endpointId\"` }"

	file, err := parser.ParseFile(token.NewFileSet(), "model.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	var field *ast.Field

	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, specification := range group.Specs {
			typeSpec, isTypeSpec := specification.(*ast.TypeSpec)
			if !isTypeSpec {
				continue
			}

			structure, ok := typeSpec.Type.(*ast.StructType)
			if ok {
				field = structure.Fields.List[0]
			}
		}
	}

	if field == nil || graphqlJSONFieldName(field) != "endpointId" {
		t.Fatalf("JSON field name = %v, want endpointId", field)
	}
}
