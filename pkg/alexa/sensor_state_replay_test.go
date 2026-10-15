package alexa_test

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

func TestListEndpoints_SensorStatesPairedSyntheticReplay(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-sensor-states.json")
	player, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	client, err := alexa.NewClient(alexa.WithGraphQLBaseURL("https://graphql.synthetic.test"),
		alexa.WithHTTPClient(&http.Client{Transport: player}))
	if err != nil {
		t.Fatal(err)
	}

	session, err := client.NewSession(alexa.WithBearerToken("synthetic-graphql-token"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if closeErr := session.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	})

	for _, test := range []struct {
		name, feature, value, errorType string
	}{
		{"motion detected", "motionSensor", `"DETECTED"`, ""},
		{"contact not detected", "contactSensor", `"NOT_DETECTED"`, ""},
		{"unknown detection", "motionSensor", `"UNKNOWN"`, ""},
		{"future detection", "contactSensor", `"FUTURE_STATE"`, ""},
		{"null detection", "motionSensor", `null`, ""},
		{"missing detection", "contactSensor", `null`, ""},
		{"unavailable detection", "motionSensor", `null`, "ENDPOINT_UNREACHABLE"},
		{"value-only illuminance", "lightSensor", `{"value":235.5}`, ""},
	} {
		// The shared fixture orders exchanges; subtests deliberately run serially.
		t.Run(test.name, func(t *testing.T) {
			response, readErr := session.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{
				EndpointIDs: []string{"synthetic-sensor"},
				IncludeFields: &alexaapimodels.EndpointIncludeFields{
					Properties: true,
					Features:   true,
				},
			})
			if readErr != nil {
				t.Fatal(readErr)
			}

			property := replaySensorProperty(t, response, test.feature)
			value, marshalErr := json.Marshal(property.StateValue)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}

			if string(value) != test.value {
				t.Errorf("state value = %s, want %s", value, test.value)
			}

			assertReplaySensorMetadata(t, property, test.errorType)
		})
	}

	if err := player.AssertConsumed(); err != nil {
		t.Fatal(err)
	}
}

func replaySensorProperty(t *testing.T, response *alexaapimodels.UnifiedEndpointListResponse, feature string) alexaapimodels.FeatureProperty {
	t.Helper()

	if len(response.Results) != 1 || response.Results[0].EndpointID != "synthetic-sensor" {
		t.Fatalf("endpoint results = %#v", response.Results)
	}

	for _, candidate := range response.Results[0].Features {
		if string(candidate.Name) == feature {
			if len(candidate.Properties) != 1 {
				t.Fatalf("%s properties = %#v", feature, candidate.Properties)
			}

			expectedName := "detectionState"
			if feature == "lightSensor" {
				expectedName = "illuminance"
			}

			if candidate.Properties[0].Name != expectedName {
				t.Errorf("%s property name = %q, want %q", feature, candidate.Properties[0].Name, expectedName)
			}

			return candidate.Properties[0]
		}
	}

	t.Fatalf("missing %s in %#v", feature, response.Results[0].Features)

	return alexaapimodels.FeatureProperty{}
}

func assertReplaySensorMetadata(t *testing.T, property alexaapimodels.FeatureProperty, errorType string) {
	t.Helper()

	if property.TimeOfSample != "2026-10-10T13:00:00Z" || property.TimeOfLastChange != "2026-10-10T12:59:00Z" {
		t.Errorf("sensor timestamps were not preserved: %#v", property)
	}

	if property.Accuracy != "HIGH" || property.Type != "RETRIEVABLE" {
		t.Errorf("sensor metadata was not preserved: %#v", property)
	}

	if errorType == "" {
		if property.Error != nil {
			t.Errorf("unexpected property error: %#v", property.Error)
		}

		return
	}

	if property.Error == nil || property.Error.Type != errorType || property.Error.Message != "synthetic unavailable" {
		t.Errorf("property error = %#v, want %s", property.Error, errorType)
	}
}
