//nolint:testpackage // Verifies the private merge on decoded synthetic payloads.
package alexa

import (
	"encoding/json"
	"testing"

	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
)

func TestMergeEndpointWithStatesSkipsNullPropertiesAndFeatures(t *testing.T) {
	t.Parallel()

	wire := []byte(`{"id":"synthetic","endpointId":"synthetic","dmsIdentifier":{"deviceType":"T","dsn":"s"},"features":[
		null,
		{"name":"brightness","instance":"","properties":[null,
			{"__typename":"SyntheticFutureKind","name":"future"},
			{"__typename":"Brightness","name":"brightness","brightnessStateValue":40}]}]}`)

	var state graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint

	err := json.Unmarshal(wire, &state)
	if err != nil {
		t.Fatal(err)
	}

	merge := endpointMergeData{deviceV2Map: nil, includeFeatures: true, includeCapabilities: true}
	endpoint := (&Session{}).mergeEndpointWithStates(state, merge)

	found := false

	for _, feature := range endpoint.Features {
		for _, property := range feature.Properties {
			if property.Name == "brightness" {
				found = true
			}

			if property.Name == "future" || property.Name == "" {
				t.Fatalf("undecodable property leaked into the result: %#v", property)
			}
		}
	}

	if !found {
		t.Fatalf("known property was lost: %#v", endpoint.Features)
	}

	// A directly constructed nil entry must not panic either.
	state.Features[1].Properties = append(state.Features[1].Properties, nil)
	_ = (&Session{}).mergeEndpointWithStates(state, merge)
}
