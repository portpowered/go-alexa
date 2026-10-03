//nolint:testpackage // Verifies private endpoint parsing and feature mapping behavior.
package alexa

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestDetermineSupportedFeatures(t *testing.T) {
	t.Parallel()

	session := &Session{}

	// helper to check if a feature is present in the result
	hasFeature := func(features []alexaapimodels.Feature, name alexaapimodels.FeatureName) bool {
		for _, f := range features {
			if f.Name == name {
				return true
			}
		}

		return false
	}

	t.Run("GraphQL connectivity feature", func(t *testing.T) {
		t.Parallel()

		features := session.determineSupportedFeatures([]string{"connectivity"}, nil, "")
		if !hasFeature(features, alexaapimodels.FeatureNameConnectivity) {
			t.Errorf("expected FeatureNameConnectivity in output, got %v", features)
		}
	})

	t.Run("GraphQL location feature", func(t *testing.T) {
		t.Parallel()

		features := session.determineSupportedFeatures([]string{"location"}, nil, "")
		if !hasFeature(features, alexaapimodels.FeatureNameLocation) {
			t.Errorf("expected FeatureNameLocation in output, got %v", features)
		}
	})

	t.Run("GraphQL locationTracker feature", func(t *testing.T) {
		t.Parallel()

		features := session.determineSupportedFeatures([]string{"locationTracker"}, nil, "")
		if !hasFeature(features, alexaapimodels.FeatureNameLocationTracker) {
			t.Errorf("expected FeatureNameLocationTracker in output, got %v", features)
		}
	})

	t.Run("REST alexa.location capability", func(t *testing.T) {
		t.Parallel()

		features := session.determineSupportedFeatures(nil, []string{"alexa.location"}, "")
		if !hasFeature(features, alexaapimodels.FeatureNameLocation) {
			t.Errorf("expected FeatureNameLocation in output, got %v", features)
		}
	})

	t.Run("REST alexa.location.tracker capability", func(t *testing.T) {
		t.Parallel()

		features := session.determineSupportedFeatures(nil, []string{"alexa.location.tracker"}, "")
		if !hasFeature(features, alexaapimodels.FeatureNameLocationTracker) {
			t.Errorf("expected FeatureNameLocationTracker in output, got %v", features)
		}
	})

	t.Run("REST mixed-case Alexa.Location.Tracker capability", func(t *testing.T) {
		t.Parallel()

		features := session.determineSupportedFeatures(nil, []string{"Alexa.Location.Tracker"}, "")
		if !hasFeature(features, alexaapimodels.FeatureNameLocationTracker) {
			t.Errorf("expected FeatureNameLocationTracker for mixed-case input, got %v", features)
		}
	})
}

type endpointEnumerationTransport struct {
	t                        *testing.T
	airQualityMonitorPath    string
	genericListEndpointsPath string
	requestBodies            []string
	airQualityMonitorQueries int
}

func (t *endpointEnumerationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte

	if req.Body != nil {
		var err error

		body, err = io.ReadAll(req.Body)
		if err != nil {
			t.t.Fatalf("failed to read request body: %v", err)
		}
	}

	bodyText := string(body)
	t.requestBodies = append(t.requestBodies, bodyText)

	switch {
	case strings.Contains(req.URL.Path, "/api/devices-v2/device"):
		return fixtureResponse(t.t, "testdata/synthetic_devices_v2_empty.json"), nil
	case strings.Contains(bodyText, "query Endpoints"):
		return fixtureResponse(t.t, "testdata/synthetic_endpoints_empty.json"), nil
	case strings.Contains(bodyText, "query ListEndpoints") && strings.Contains(bodyText, `"displayCategory":"AIR_QUALITY_MONITOR"`):
		t.airQualityMonitorQueries++

		return fixtureResponse(t.t, t.airQualityMonitorPath), nil
	case strings.Contains(bodyText, "query ListEndpoints"):
		path := t.genericListEndpointsPath
		if path == "" {
			path = "testdata/synthetic_list_endpoints_with_states_empty.json"
		}

		return fixtureResponse(t.t, path), nil
	default:
		t.t.Fatalf("unexpected request url=%s body=%s", req.URL.String(), bodyText)

		return nil, staticError("unexpected request reached fixture transport")
	}
}

func fixtureResponse(t *testing.T, path string) *http.Response {
	t.Helper()

	//nolint:gosec // The path points to a checked-in synthetic fixture.
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestListEndpoints_IncludesAirQualityMonitorAbsentFromGenericEndpointList(t *testing.T) {
	t.Parallel()

	transport := &endpointEnumerationTransport{
		t:                     t,
		airQualityMonitorPath: "testdata/synthetic_list_endpoints_air_quality_monitor.json",
	}
	client := newTestSession(t, &http.Client{Transport: transport}, WithBearerToken("test-token"))

	response, err := client.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{})
	if err != nil {
		t.Fatalf("ListEndpoints returned error: %v", err)
	}

	if transport.airQualityMonitorQueries != 1 {
		t.Fatalf("expected one dedicated AQM query, got %d", transport.airQualityMonitorQueries)
	}

	if len(response.Results) != 1 {
		t.Fatalf("expected one endpoint, got %d", len(response.Results))
	}

	endpoint := response.Results[0]
	if endpoint.EndpointID != "amzn1.alexa.endpoint.synthetic-aqm-001" {
		t.Fatalf("endpoint ID = %q", endpoint.EndpointID)
	}

	if endpoint.FriendlyName == nil || endpoint.FriendlyName.Value != "Synthetic Air Quality Monitor" {
		t.Fatalf("friendly name = %#v", endpoint.FriendlyName)
	}

	if endpoint.DeviceSerialNumber != "synthetic-serial-001" {
		t.Fatalf("device serial number = %q", endpoint.DeviceSerialNumber)
	}

	if endpoint.DeviceType != "AMAZON_AIRQUALITYMONITOR" {
		t.Fatalf("device type = %q", endpoint.DeviceType)
	}

	if endpoint.Manufacturer == nil || endpoint.Manufacturer.Value != "Amazon" {
		t.Fatalf("manufacturer = %#v", endpoint.Manufacturer)
	}

	if endpoint.Model == nil || endpoint.Model.Value != "Smart Air Quality Monitor" {
		t.Fatalf("model = %#v", endpoint.Model)
	}

	if endpoint.SoftwareVersion == nil || endpoint.SoftwareVersion.Value != "1.2.3" {
		t.Fatalf("software version = %#v", endpoint.SoftwareVersion)
	}

	if endpoint.DisplayCategories.Primary != alexaapimodels.EndpointDisplayCategoryAirQualityMonitor {
		t.Fatalf("primary display category = %q", endpoint.DisplayCategories.Primary)
	}

	if !endpoint.HasFeature(alexaapimodels.FeatureNameRange) {
		t.Fatalf("expected range feature, got %#v", endpoint.Features)
	}
}

func TestListEndpoints_EmptyAirQualityMonitorResultDoesNotFailGenericEnumeration(t *testing.T) {
	t.Parallel()

	transport := &endpointEnumerationTransport{
		t:                     t,
		airQualityMonitorPath: "testdata/synthetic_list_endpoints_air_quality_monitor_empty.json",
	}
	client := newTestSession(t, &http.Client{Transport: transport}, WithBearerToken("test-token"))

	response, err := client.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{})
	if err != nil {
		t.Fatalf("ListEndpoints returned error: %v", err)
	}

	if transport.airQualityMonitorQueries != 1 {
		t.Fatalf("expected one dedicated AQM query, got %d", transport.airQualityMonitorQueries)
	}

	if len(response.Results) != 0 {
		t.Fatalf("expected no endpoints, got %d", len(response.Results))
	}
}

func TestListEndpointsWithStates_PreservesAirQualityMonitorRangeInstancesAndValues(t *testing.T) {
	t.Parallel()

	transport := &endpointEnumerationTransport{
		t:                     t,
		airQualityMonitorPath: "testdata/synthetic_list_endpoints_air_quality_monitor_with_states.json",
	}
	client := newTestSession(t, &http.Client{Transport: transport}, WithBearerToken("test-token"))

	response, err := client.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{
		IncludeFields: &alexaapimodels.EndpointIncludeFields{
			Properties: true,
			Features:   true,
		},
	})
	if err != nil {
		t.Fatalf("ListEndpoints returned error: %v", err)
	}

	if len(response.Results) != 1 {
		t.Fatalf("expected one endpoint, got %d", len(response.Results))
	}

	featuresByInstance := rangeFeaturesByInstance(response.Results[0].Features)
	if len(featuresByInstance) != 5 {
		t.Fatalf("expected five range instances, got %d: %#v", len(featuresByInstance), response.Results[0].Features)
	}

	assertRangeFeature(t, featuresByInstance["Air Quality.humidity"], "Humidity", "%", 42)
	assertRangeFeature(t, featuresByInstance["Air Quality.pm25"], "PM2.5", "ug/m3", 7.5)
	assertRangeFeature(t, featuresByInstance["Air Quality.indoorAirQuality"], "Air quality", "AQI", 37)
	assertRangeFeature(t, featuresByInstance["Air Quality.voc"], "VOC", "ppb", 125)

	pm10 := featuresByInstance["Air Quality.pm10"]
	if pm10.Instance == "" {
		t.Fatalf("expected pm10 range instance to be preserved")
	}

	if len(pm10.Properties) != 1 {
		t.Fatalf("expected one pm10 property, got %d", len(pm10.Properties))
	}

	if pm10.Properties[0].Error == nil || pm10.Properties[0].Error.Type != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND pm10 property error, got %#v", pm10.Properties[0].Error)
	}

	if pm10.Properties[0].StateValue != nil {
		t.Fatalf("expected NOT_FOUND pm10 state value to be nil, got %#v", pm10.Properties[0].StateValue)
	}
}

func TestExtractRangeStateValuePreservesPresentZeroAndOmitsMissingValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wireJSON  string
		wantNil   bool
		wantValue float64
	}{
		{
			name:      "present zero",
			wireJSON:  `{"rangeValue":{"__typename":"RangeValueNumber","value":0}}`,
			wantNil:   false,
			wantValue: 0,
		},
		{
			name:      "present nonzero",
			wireJSON:  `{"rangeValue":{"__typename":"RangeValueNumber","value":12.5}}`,
			wantNil:   false,
			wantValue: 12.5,
		},
		{
			name:      "explicit null",
			wireJSON:  `{"rangeValue":null}`,
			wantNil:   true,
			wantValue: 0,
		},
		{
			name:      "omitted",
			wireJSON:  `{}`,
			wantNil:   true,
			wantValue: 0,
		},
		{
			name:      "provider error",
			wireJSON:  `{"error":{"type":"NOT_FOUND","message":"missing"},"rangeValue":{"__typename":"RangeValueNumber","value":0}}`,
			wantNil:   true,
			wantValue: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var property gqlStatePropertyRangeValue

			err := json.Unmarshal([]byte(test.wireJSON), &property)
			if err != nil {
				t.Fatalf("decode generated GraphQL property: %v", err)
			}

			got := extractRangeStateValue(&property)
			if test.wantNil {
				if got != nil {
					t.Fatalf("extractRangeStateValue() = %#v, want nil", got)
				}

				return
			}

			if got == nil || got.Value != test.wantValue {
				t.Fatalf("extractRangeStateValue() = %#v, want value %v", got, test.wantValue)
			}
		})
	}
}

func rangeFeaturesByInstance(features []alexaapimodels.Feature) map[string]alexaapimodels.Feature {
	result := make(map[string]alexaapimodels.Feature)

	for _, feature := range features {
		if feature.Name == alexaapimodels.FeatureNameRange {
			result[feature.Instance] = feature
		}
	}

	return result
}

func assertRangeFeature(
	t *testing.T,
	feature alexaapimodels.Feature,
	expectedLabel string,
	expectedUnit string,
	expectedValue float64,
) {
	t.Helper()

	if feature.Instance == "" {
		t.Fatalf("expected range feature for %s", expectedLabel)
	}

	if feature.Config == nil || feature.Config.Range == nil {
		t.Fatalf("expected range config for %s, got %#v", expectedLabel, feature.Config)
	}

	if feature.Config.Range.FriendlyName == nil || feature.Config.Range.FriendlyName.Value != expectedLabel {
		t.Fatalf("range label = %#v, want %q", feature.Config.Range.FriendlyName, expectedLabel)
	}

	if feature.Config.Range.UnitOfMeasure == nil || feature.Config.Range.UnitOfMeasure.Value != expectedUnit {
		t.Fatalf("range unit = %#v, want %q", feature.Config.Range.UnitOfMeasure, expectedUnit)
	}

	if len(feature.Properties) != 1 {
		t.Fatalf("expected one property for %s, got %d", expectedLabel, len(feature.Properties))
	}

	value, ok := feature.Properties[0].StateValue.(*alexaapimodels.RangeValueState)
	if !ok {
		t.Fatalf("state value for %s = %#v, want RangeValueState", expectedLabel, feature.Properties[0].StateValue)
	}

	if value.Value != expectedValue {
		t.Fatalf("state value for %s = %v, want %v", expectedLabel, value.Value, expectedValue)
	}
}
