package alexa

import (
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestGetFeatureDefaults_Connectivity(t *testing.T) {
	props, ops := getFeatureDefaults(alexaapimodels.FeatureNameConnectivity)
	if len(props) != 1 {
		t.Fatalf("expected 1 property, got %d", len(props))
	}
	if props[0].Name != "reachabilityState" {
		t.Errorf("expected property name 'reachabilityState', got %q", props[0].Name)
	}
	if len(ops) != 0 {
		t.Errorf("expected 0 operations, got %d", len(ops))
	}
}

func TestGetFeatureDefaults_Speaker(t *testing.T) {
	props, ops := getFeatureDefaults(alexaapimodels.FeatureNameSpeaker)
	if len(props) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(props))
	}
	names := map[string]bool{}
	for _, p := range props {
		names[p.Name] = true
	}
	if !names["volume"] {
		t.Error("expected 'volume' property")
	}
	if !names["muted"] {
		t.Error("expected 'muted' property")
	}
	if len(ops) != 0 {
		t.Errorf("expected 0 operations for speaker, got %d", len(ops))
	}
}

func TestGetFeatureDefaults_Playback(t *testing.T) {
	props, ops := getFeatureDefaults(alexaapimodels.FeatureNamePlayback)
	if len(props) != 0 {
		t.Errorf("expected 0 properties, got %d", len(props))
	}
	expectedOps := map[string]bool{
		"Play": true, "Pause": true, "Next": true, "Previous": true,
		"Stop": true, "FastForward": true, "Rewind": true, "StartOver": true,
	}
	if len(ops) != len(expectedOps) {
		t.Fatalf("expected %d operations, got %d", len(expectedOps), len(ops))
	}
	for _, op := range ops {
		if !expectedOps[op.Name] {
			t.Errorf("unexpected operation %q", op.Name)
		}
	}
}

func TestGetFeatureDefaults_UnknownFeature(t *testing.T) {
	props, ops := getFeatureDefaults(alexaapimodels.FeatureName("unknownFeature"))
	if props == nil {
		t.Error("expected non-nil properties slice")
	}
	if ops == nil {
		t.Error("expected non-nil operations slice")
	}
	if len(props) != 0 {
		t.Errorf("expected 0 properties, got %d", len(props))
	}
	if len(ops) != 0 {
		t.Errorf("expected 0 operations, got %d", len(ops))
	}
}

func TestDetermineSupportedFeatures_PopulatesPropertiesAndOperations(t *testing.T) {
	c := &Client{}
	features := c.determineSupportedFeatures(
		[]string{"connectivity", "speaker", "playback"},
		[]string{},
		"",
	)

	featureMap := make(map[string]alexaapimodels.Feature)
	for _, f := range features {
		featureMap[string(f.Name)] = f
	}

	// Verify connectivity has reachabilityState property
	conn, ok := featureMap["connectivity"]
	if !ok {
		t.Fatal("expected connectivity feature")
	}
	if len(conn.Properties) != 1 || conn.Properties[0].Name != "reachabilityState" {
		t.Errorf("connectivity: expected [reachabilityState] property, got %v", conn.Properties)
	}

	// Verify speaker has volume and muted properties
	spk, ok := featureMap["speaker"]
	if !ok {
		t.Fatal("expected speaker feature")
	}
	if len(spk.Properties) != 2 {
		t.Fatalf("speaker: expected 2 properties, got %d", len(spk.Properties))
	}
	propNames := map[string]bool{}
	for _, p := range spk.Properties {
		propNames[p.Name] = true
	}
	if !propNames["volume"] || !propNames["muted"] {
		t.Errorf("speaker: expected volume and muted, got %v", spk.Properties)
	}

	// Verify playback has operations
	pb, ok := featureMap["playback"]
	if !ok {
		t.Fatal("expected playback feature")
	}
	if len(pb.Operations) == 0 {
		t.Error("playback: expected operations to be populated")
	}
	opNames := map[string]bool{}
	for _, op := range pb.Operations {
		opNames[op.Name] = true
	}
	for _, expected := range []string{"Play", "Pause", "Next", "Previous"} {
		if !opNames[expected] {
			t.Errorf("playback: missing expected operation %q", expected)
		}
	}

	// Verify all features have non-nil slices (not null in JSON)
	for _, f := range features {
		if f.Properties == nil {
			t.Errorf("feature %q: Properties is nil, expected empty slice", f.Name)
		}
		if f.Operations == nil {
			t.Errorf("feature %q: Operations is nil, expected empty slice", f.Name)
		}
	}
}

func TestMergeEndpointWithoutStates_IncludesProperties(t *testing.T) {
	c := &Client{}

	// Build a GraphQL endpoint with a feature that has properties via operations
	gqlEndpoint := graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint{
		Id:         "test-id",
		EndpointId: "test-endpoint",
		Features: []graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeature{
			{
				Name:     "power",
				Instance: "test-instance",
				Operations: []graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeatureOperationsFeatureOperation{
					{Name: "CustomTurnOn"},
				},
			},
		},
	}

	mergeData := endpointMergeData{
		deviceV2Map:         map[string]*alexamodels.DeviceV2{},
		includeFeatures:     true,
		includeCapabilities: true,
	}

	result := c.mergeEndpointWithoutStates(gqlEndpoint, mergeData)

	// Find the power feature
	var powerFeature *alexaapimodels.Feature
	for i, f := range result.Features {
		if f.Name == alexaapimodels.FeatureNamePower {
			powerFeature = &result.Features[i]
			break
		}
	}
	if powerFeature == nil {
		t.Fatal("expected power feature in result")
	}

	// Verify that Operations from GraphQL override defaults
	if len(powerFeature.Operations) != 1 || powerFeature.Operations[0].Name != "CustomTurnOn" {
		t.Errorf("expected GraphQL operations to override defaults, got %v", powerFeature.Operations)
	}
}

func TestMergeEndpointWithoutStates_DefaultPropertiesWhenNoGraphQL(t *testing.T) {
	c := &Client{}

	// Endpoint with power feature but no GraphQL metadata for it
	// Power is inferred from device family (FireTV)
	gqlEndpoint := graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint{
		Id:         "test-id",
		EndpointId: "test-endpoint",
		Features:   []graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeature{},
	}

	mergeData := endpointMergeData{
		deviceV2Map: map[string]*alexamodels.DeviceV2{
			":": {DeviceFamily: "FIRE_TV"},
		},
		includeFeatures:     true,
		includeCapabilities: true,
	}

	result := c.mergeEndpointWithoutStates(gqlEndpoint, mergeData)

	// FireTV should have power feature with default properties
	var powerFeature *alexaapimodels.Feature
	for i, f := range result.Features {
		if f.Name == alexaapimodels.FeatureNamePower {
			powerFeature = &result.Features[i]
			break
		}
	}
	if powerFeature == nil {
		t.Fatal("expected power feature for FireTV device")
	}
	if len(powerFeature.Properties) != 1 || powerFeature.Properties[0].Name != "powerState" {
		t.Errorf("expected default powerState property, got %v", powerFeature.Properties)
	}
}

func TestGetFeatureDefaults_ReturnsCopies(t *testing.T) {
	// Verify that modifying returned slices doesn't affect the canonical mapping
	props1, ops1 := getFeatureDefaults(alexaapimodels.FeatureNamePower)
	props1[0].Name = "modified"
	if len(ops1) > 0 {
		ops1[0].Name = "modified"
	}

	props2, _ := getFeatureDefaults(alexaapimodels.FeatureNamePower)
	if props2[0].Name == "modified" {
		t.Error("getFeatureDefaults should return copies, not references to the canonical data")
	}
}
