package main

import (
	"strings"
	"testing"
)

func TestBehaviorPayloadInventoryAcceptsProductionDefinitions(t *testing.T) {
	t.Parallel()

	err := checkBehaviorPayloadInventory()
	if err != nil {
		t.Fatalf("checkBehaviorPayloadInventory() rejected production payload schemas or builders: %v", err)
	}
}

func TestBehaviorPayloadSchemaRejectsMissingNestedPayloadKey(t *testing.T) {
	t.Parallel()

	document, err := readPayloadSchema(behaviorSchemaPath)
	if err != nil {
		t.Fatal(err)
	}

	keySchema := document.Components.Schemas["BehaviorPayloadKey"]
	keySchema.Enum = removeSchemaEnumValue(keySchema.Enum, "display")
	document.Components.Schemas["BehaviorPayloadKey"] = keySchema

	err = checkBehaviorPayloadSchema(document)
	if err == nil || !strings.Contains(err.Error(), `property "display" is missing from BehaviorPayloadKey`) {
		t.Fatalf("checkBehaviorPayloadSchema() error = %v; want missing nested payload key rejection", err)
	}
}

func TestBehaviorPayloadSchemaRejectsMissingNestedPayloadDefinition(t *testing.T) {
	t.Parallel()

	document, err := readPayloadSchema(behaviorSchemaPath)
	if err != nil {
		t.Fatal(err)
	}

	delete(document.Components.Schemas, "BehaviorAnnouncementDisplay")

	err = checkBehaviorPayloadSchema(document)
	if err == nil || !strings.Contains(err.Error(), "missing schema BehaviorAnnouncementDisplay") {
		t.Fatalf("checkBehaviorPayloadSchema() error = %v; want missing nested payload definition rejection", err)
	}
}

func TestFeaturePayloadSchemaKeepsOnlyActionParametersOpen(t *testing.T) {
	t.Parallel()

	featureSchema, err := readPayloadSchema(featureControlSchemaPath)
	if err != nil {
		t.Fatal(err)
	}

	compatSchema, err := readPayloadSchema(compatControlSchemaPath)
	if err != nil {
		t.Fatal(err)
	}

	payloadSchema := featureSchema.Components.Schemas["SpeakerSetVolumePayload"]
	payloadSchema.AdditionalProperties = true
	featureSchema.Components.Schemas["SpeakerSetVolumePayload"] = payloadSchema

	err = checkFeaturePayloadOpenBoundary(featureSchema, compatSchema)
	if err == nil || !strings.Contains(err.Error(), "SpeakerSetVolumePayload unexpectedly permits additional properties") {
		t.Fatalf("checkFeaturePayloadOpenBoundary() error = %v; want unexpected open-payload rejection", err)
	}
}

func TestDynamicPayloadMapGateRejectsMapsInUnlistedMethods(t *testing.T) {
	t.Parallel()

	source := []byte(`package rest
type Client struct{}
func (*Client) BuildFuturePayload() {
	payload := map[string]any{"syntheticKey": "synthetic-value"}
	_ = payload
}
`)

	err := checkDynamicPayloadMapSource("synthetic.go", source)
	if err == nil || !strings.Contains(err.Error(), "constructs a handwritten dynamic payload map") {
		t.Fatalf("checkDynamicPayloadMapSource() error = %v; want unlisted map-construction rejection", err)
	}
}

func TestDynamicPayloadMapGateRejectsMakeAndIndexedKeyWrites(t *testing.T) {
	t.Parallel()

	sources := []string{
		`package graphql
func buildDifferentFeature() {
	payload := make(((map[string]interface{})))
	payload["syntheticKey"] = "synthetic-value"
}
`,
		`package graphql
func buildDifferentFeature() {
	_ = new(((map[string]interface{})))
}
`,
		`package graphql
func buildDifferentFeature() {
	_ = ((map[string]interface{}))(nil)
}
`,
		`package graphql
func buildDifferentFeature() {
	var payload map[string]interface{}
	payload["syntheticKey"] = "synthetic-value"
}
`,
	}

	for _, source := range sources {
		err := checkDynamicPayloadMapSource("synthetic.go", []byte(source))
		if err == nil || (!strings.Contains(err.Error(), "constructs a handwritten dynamic payload map") &&
			!strings.Contains(err.Error(), "mutates or adds a key")) {
			t.Fatalf("checkDynamicPayloadMapSource() error = %v; want dynamic-map creation or write rejection", err)
		}
	}
}

func TestDynamicPayloadMapGateRejectsNamedMapAliasesAndCallerMapAliases(t *testing.T) {
	t.Parallel()

	sources := []string{
		`package graphql
type FeaturePayload = map[string]any
func buildDifferentFeature() {
	payload := FeaturePayload{"syntheticKey": true}
	_ = payload
}
`,
		`package graphql
type FeaturePayload map[string]interface{}
func buildDifferentFeature() {
	var payload FeaturePayload
	payload["syntheticKey"] = true
}
`,
		`package graphql
type request struct { Params map[string]interface{} }
func buildAction(req request) {
	payload := ((req.Params))
	payload["syntheticKey"] = "synthetic-value"
}
`,
		`package graphql
type request struct { Params map[string]interface{} }
func buildAction(req request) {
	payload := req.Params
	func() { payload["syntheticKey"] = "synthetic-value" }()
}
`,
	}

	for _, source := range sources {
		err := checkDynamicPayloadMapSource("synthetic.go", []byte(source))
		if err == nil {
			t.Fatal("checkDynamicPayloadMapSource() accepted a named or aliased dynamic payload map")
		}
	}
}

func TestDynamicPayloadMapGateRejectsIndexedWritesToOpenParams(t *testing.T) {
	t.Parallel()

	source := []byte(`package graphql
type request struct { Params map[string]interface{} }
func buildAction(req request) {
	req.Params["syntheticKey"] = "synthetic-value"
}
`)

	err := checkDynamicPayloadMapSource("synthetic.go", source)
	if err == nil || !strings.Contains(err.Error(), "mutates or adds a key") {
		t.Fatalf("checkDynamicPayloadMapSource() error = %v; want caller-map key-write rejection", err)
	}
}

func TestCallerOpenBoundaryKeepsGenericSendSequenceMap(t *testing.T) {
	t.Parallel()

	source := []byte(`package rest
type Client struct{}
func (*Client) SendSequence(operationPayload map[string]interface{}) {
	_ = operationPayload
}
`)

	err := checkSendSequenceOpenBoundary("synthetic.go", source)
	if err != nil {
		t.Fatalf("checkSendSequenceOpenBoundary() rejected the public generic input: %v", err)
	}
}

func removeSchemaEnumValue(values []any, unwanted string) []any {
	result := make([]any, 0, len(values))

	for _, value := range values {
		if value != unwanted {
			result = append(result, value)
		}
	}

	return result
}
