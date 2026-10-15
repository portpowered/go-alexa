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

func TestListEndpoints_PreservesErrorsAcrossSelectedStateProperties(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-property-errors.json")

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
		closeErr := session.Close()
		if closeErr != nil {
			t.Error(closeErr)
		}
	})

	// The two exchanges are ordered: absent errors, then errors alongside values.
	for _, expectErrors := range []bool{false, true} {
		properties := readPropertyErrorMatrix(t, session)
		assertPropertyErrorMatrix(t, properties, expectErrors)
		assertErrorMatrixValues(t, properties)
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}
}

func readPropertyErrorMatrix(t *testing.T, session *alexa.Session) []alexaapimodels.FeatureProperty {
	t.Helper()

	response, err := session.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{
		EndpointIDs:   []string{"synthetic-property-errors"},
		IncludeFields: &alexaapimodels.EndpointIncludeFields{Properties: true, Features: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Results) != 1 || len(response.Results[0].Features) != 1 {
		t.Fatalf("property matrix endpoints = %#v", response.Results)
	}

	return response.Results[0].Features[0].Properties
}

func assertPropertyErrorMatrix(t *testing.T, properties []alexaapimodels.FeatureProperty, expectErrors bool) {
	t.Helper()

	names := []string{
		"ActionState", "ArmState", "Battery", "Brightness", "BurglaryAlarm", "CarbonMonoxideAlarm",
		"Color", "ColorTemperature", "Reachability", "RangeValue", "PowerLevel", "Power", "Percentage",
		"NetworkThroughput", "Mode", "Lock", "DetectionState", "Illuminance", "Geolocation", "FireAlarm",
		"RadioDiagnostics", "WaterAlarm", "Volume", "ToggleState", "ThermostatMode", "TemperatureSensor",
		"StatusCode", "Snapshot", "SirenState", "Setpoint", "RelativeHumidity",
	}
	if len(properties) != len(names) {
		t.Fatalf("selected properties = %d, want %d", len(properties), len(names))
	}

	byName := make(map[string]alexaapimodels.FeatureProperty, len(properties))
	for _, property := range properties {
		byName[property.Name] = property
	}

	for _, name := range names {
		property, present := byName[name]
		if !present {
			t.Fatalf("selected property %s missing", name)
		}

		if !expectErrors {
			if property.Error != nil {
				t.Errorf("%s invented error: %#v", name, property.Error)
			}

			continue
		}

		if property.Error == nil || property.Error.Type != "ENDPOINT_UNREACHABLE" ||
			property.Error.Message != "synthetic "+name+" unavailable" {
			t.Errorf("%s lost its error: %#v", name, property.Error)
		}
	}
}

func assertErrorMatrixValues(t *testing.T, properties []alexaapimodels.FeatureProperty) {
	t.Helper()

	expected := map[string]string{
		"Color":             `{"hue":120,"saturation":0.5,"brightness":0.6}`,
		"ColorTemperature":  `4100`,
		"Setpoint":          `{"value":21,"scale":"CELSIUS"}`,
		"RelativeHumidity":  `{"value":37.5}`,
		"NetworkThroughput": `null`,
	}
	for _, property := range properties {
		want, asserted := expected[property.Name]
		if !asserted {
			continue
		}

		value, err := json.Marshal(property.StateValue)
		if err != nil {
			t.Fatal(err)
		}

		if string(value) != want {
			t.Errorf("%s value = %s, want %s", property.Name, value, want)
		}
	}
}
