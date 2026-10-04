//nolint:testpackage // Exercises the private event parser and conversion boundary.
package alexa

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	eventFixtureSampleTime     = "2025-12-02T05:17:49Z"
	eventFixtureLastChangeTime = "2025-12-01T05:17:49Z"
	eventFixtureEndpointID     = "event-fixture-endpoint"
	eventFixtureInstance       = "event-fixture-instance"
	eventFixtureAccuracy       = "HIGH"
	eventFixturePropertyType   = "RETRIEVABLE"
)

type eventPayloadCase struct {
	metricName string
	metadata   string
	wantType   reflect.Type
	want       map[string]any
	lastChange bool
}

func eventPayloadCases(t *testing.T) []eventPayloadCase {
	t.Helper()

	cases := make([]eventPayloadCase, 0, len(eventParserRegistry))
	cases = append(cases, scalarEventPayloadCases(t)...)
	cases = append(cases, sensorEventPayloadCases(t)...)
	cases = append(cases, compositeEventPayloadCases(t)...)

	return cases
}

func scalarEventPayloadCases(t *testing.T) []eventPayloadCase {
	t.Helper()

	return []eventPayloadCase{
		newEventPayloadCase(t, "EndpointColorTemperature", map[string]any{
			"colorTemperatureInKelvinStateValue": 2700,
		}, (*alexaapimodels.ColorTemperaturePayload)(nil), map[string]any{
			"colorTemperatureInKelvin": 2700,
		}, false),
		newEventPayloadCase(t, "EndpointPower", map[string]any{
			"powerStateValue": "ON",
		}, (*alexaapimodels.PowerPayload)(nil), map[string]any{
			"powerState": "ON",
		}, false),
		newEventPayloadCase(t, "EndpointSpeaker", map[string]any{
			"volume": 6, "muted": true,
		}, (*alexaapimodels.SpeakerPayload)(nil), map[string]any{
			"volume": 6, "muted": true,
		}, false),
		newEventPayloadCase(t, "EndpointBrightness", map[string]any{
			"brightnessStateValue": 55,
		}, (*alexaapimodels.BrightnessPayload)(nil), map[string]any{
			"brightness": 55,
		}, false),
		newEventPayloadCase(t, "EndpointColor", map[string]any{
			"colorStateValue": map[string]any{
				"hue": 120.0, "saturation": 0.5, "brightness": 0.75,
			},
		}, (*alexaapimodels.ColorPayload)(nil), map[string]any{
			"hue": 120.0, "saturation": 0.5, "brightness": 0.75,
		}, false),
		newEventPayloadCase(t, "EndpointLock", map[string]any{
			"lockState": "LOCKED",
		}, (*alexaapimodels.LockPayload)(nil), map[string]any{
			"lockState": "LOCKED",
		}, false),
		newEventPayloadCase(t, "EndpointMode", map[string]any{
			"modeValue": map[string]any{"value": "HEAT"},
		}, (*alexaapimodels.ModePayload)(nil), map[string]any{
			"instance": eventFixtureInstance, "mode": "HEAT",
		}, false),
		newEventPayloadCase(t, "EndpointRange", map[string]any{
			"rangeValue": map[string]any{"value": 12.5},
		}, (*alexaapimodels.RangePayload)(nil), map[string]any{
			"instance": eventFixtureInstance, "rangeValue": 12.5,
		}, false),
		newEventPayloadCase(t, "EndpointToggle", map[string]any{
			"toggleStateValue": "ON",
		}, (*alexaapimodels.TogglePayload)(nil), map[string]any{
			"instance": eventFixtureInstance, "toggleState": "ON",
		}, false),
		newEventPayloadCase(t, "EndpointPercentage", map[string]any{
			"percentageValue": 75,
		}, (*alexaapimodels.PercentagePayload)(nil), map[string]any{
			"percentage": 75.0,
		}, false),
		newEventPayloadCase(t, "EndpointPowerLevel", map[string]any{
			"powerLevelValue": 8,
		}, (*alexaapimodels.PowerLevelPayload)(nil), map[string]any{
			"powerLevel": 8,
		}, false),
	}
}

func sensorEventPayloadCases(t *testing.T) []eventPayloadCase {
	t.Helper()

	change := true
	cases := []eventPayloadCase{
		newEventPayloadCase(t, "EndpointThermostat", map[string]any{
			"thermostatModeValue": "HEAT",
		}, (*alexaapimodels.ThermostatModePayload)(nil), map[string]any{
			"thermostatMode": "HEAT",
		}, change),
		newEventPayloadCase(t, "EndpointTemperatureSensor", map[string]any{
			"value": map[string]any{"value": 21.5, "scale": "CELSIUS"},
		}, (*alexaapimodels.TemperatureSensorPayload)(nil), map[string]any{
			"value": 21.5, "scale": "CELSIUS",
		}, change),
		newEventPayloadCase(t, "EndpointMotionSensor", map[string]any{
			"detectionStateValue": "DETECTED",
		}, (*alexaapimodels.DetectionStatePayload)(nil), map[string]any{
			"detectionState": "DETECTED",
		}, change),
		newEventPayloadCase(t, "EndpointContactSensor", map[string]any{
			"detectionStateValue": "NOT_DETECTED",
		}, (*alexaapimodels.DetectionStatePayload)(nil), map[string]any{
			"detectionState": "NOT_DETECTED",
		}, change),
		newEventPayloadCase(t, "EndpointConnectivity", map[string]any{
			"reachabilityStatusValue": "UNREACHABLE",
		}, (*alexaapimodels.ReachabilityPayload)(nil), map[string]any{
			"reachabilityStatus": "UNREACHABLE",
		}, change),
		newEventPayloadCase(t, "EndpointSecurityPanel", map[string]any{
			"armStateValue": "ARMED_AWAY",
		}, (*alexaapimodels.ArmStatePayload)(nil), map[string]any{
			"armState": "ARMED_AWAY",
		}, change),
		newEventPayloadCase(t, "EndpointHumiditySensor", map[string]any{
			"relativeHumidityValue": map[string]any{"value": 42.75},
		}, (*alexaapimodels.RelativeHumidityPayload)(nil), map[string]any{
			"value": 42.75,
		}, change),
	}

	return cases
}

func compositeEventPayloadCases(t *testing.T) []eventPayloadCase {
	t.Helper()

	change := true

	return []eventPayloadCase{
		newEventPayloadCase(t, "EndpointAction", map[string]any{
			"actionStateValue": map[string]any{
				"status": "RUNNING", "actionId": "action-123",
				"targetIds": []string{"target-1", "target-2"},
				"timeInterval": map[string]any{
					"start":    eventFixtureSampleTime,
					"end":      eventFixtureLastChangeTime,
					"duration": "PT30M",
				},
			},
		}, (*alexaapimodels.ActionStatePayload)(nil), map[string]any{
			"instance": eventFixtureInstance,
			"status":   "RUNNING", "actionId": "action-123",
			"targetIds": []string{"target-1", "target-2"},
			"timeInterval": map[string]any{
				"start":    eventFixtureSampleTime,
				"end":      eventFixtureLastChangeTime,
				"duration": "PT30M",
			},
		}, change),
		newEventPayloadCase(t, "EndpointEndpointHealth", map[string]any{
			"batteryValue": map[string]any{
				"levelPercentage": 80,
				"health":          map[string]any{"state": "OK", "reasons": []string{"healthy"}},
				"chargingHealth":  map[string]any{"state": "CHARGING", "reason": "connected"},
			},
		}, (*alexaapimodels.BatteryPayload)(nil), map[string]any{
			"levelPercentage": 80,
			"health":          map[string]any{"state": "OK", "reasons": []string{"healthy"}},
			"chargingHealth":  map[string]any{"state": "CHARGING", "reason": "connected"},
		}, change),
		newEventPayloadCase(t, "EndpointLightSensor", map[string]any{
			"illuminanceValue": map[string]any{"value": 235.5},
		}, (*alexaapimodels.IlluminancePayload)(nil), map[string]any{
			"value": 235.5,
		}, change),
		newEventPayloadCase(t, "EndpointLocation", map[string]any{
			"geolocationValue": map[string]any{
				"source": "SELF_REPORTED",
				"coordinate": map[string]any{
					"latitudeInDegrees": 47.6, "longitudeInDegrees": -122.3,
					"accuracyInMeters": 4.0,
				},
				"altitude": map[string]any{"altitudeInMeters": 12.0},
				"heading":  map[string]any{"directionInDegrees": 90.0},
				"speed":    map[string]any{"speedInMetersPerSecond": 1.5},
			},
		}, (*alexaapimodels.GeolocationPayload)(nil), map[string]any{
			"source": "SELF_REPORTED",
			"coordinate": map[string]any{
				"latitudeInDegrees": 47.6, "longitudeInDegrees": -122.3,
				"accuracyInMeters": 4.0,
			},
			"altitude": map[string]any{"altitudeInMeters": 12.0},
			"heading":  map[string]any{"directionInDegrees": 90.0},
			"speed":    map[string]any{"speedInMetersPerSecond": 1.5},
		}, change),
		newEventPayloadCase(t, "EndpointStatusCode", map[string]any{
			"statusCodeValue": map[string]any{"value": []any{
				map[string]any{"code": "LOW_BATTERY", "timeOfDetection": eventFixtureSampleTime},
			}},
		}, (*alexaapimodels.StatusCodePayload)(nil), map[string]any{
			"codes": []any{map[string]any{
				"code": "LOW_BATTERY", "timeOfDetection": eventFixtureSampleTime,
			}},
		}, change),
	}
}

func newEventPayloadCase(
	t *testing.T,
	metricName string,
	fields map[string]any,
	wantType any,
	want map[string]any,
	lastChange bool,
) eventPayloadCase {
	t.Helper()

	property := map[string]any{
		"__typename":       "SyntheticProperty",
		"name":             "state",
		"timeOfSample":     eventFixtureSampleTime,
		"timeOfLastChange": eventFixtureLastChangeTime,
		"accuracy":         eventFixtureAccuracy,
		"type":             eventFixturePropertyType,
		"error":            map[string]any{"type": "ENDPOINT_UNREACHABLE"},
	}
	for name, value := range fields {
		property[name] = value
	}

	feature := map[string]any{
		"__typename": "Feature",
		"name":       "state",
		"instance":   eventFixtureInstance,
		"properties": []any{property},
	}
	resource := map[string]any{
		"metricName": metricName,
		"payload": map[string]any{
			"data": map[string]any{
				"__typename": "Endpoint",
				"features":   []any{feature},
			},
			"entity": map[string]any{
				"__typename": "Endpoint",
				"id":         eventFixtureEndpointID,
			},
		},
		"type":      "UPDATE_ENTITY",
		"timestamp": eventFixtureSampleTime,
	}

	metadata, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("marshal event metadata: %v", err)
	}

	return eventPayloadCase{
		metricName: metricName,
		metadata:   string(metadata),
		wantType:   reflect.TypeOf(wantType),
		want:       expectedEventPayload(want, lastChange),
		lastChange: lastChange,
	}
}

func expectedEventPayload(values map[string]any, lastChange bool) map[string]any {
	want := map[string]any{
		"timeOfSample": eventFixtureSampleTime,
		"accuracy":     eventFixtureAccuracy,
		"type":         eventFixturePropertyType,
		"error":        map[string]any{"type": "ENDPOINT_UNREACHABLE"},
	}
	if lastChange {
		want["timeOfLastChange"] = eventFixtureLastChangeTime
	}

	for name, value := range values {
		want[name] = value
	}

	return want
}

func TestEventPayloadRegistryProjectsPublicPayloads(t *testing.T) {
	t.Parallel()

	cases := eventPayloadCases(t)
	if len(cases) != len(eventParserRegistry) {
		t.Fatalf("got %d test cases for %d registered event metrics", len(cases), len(eventParserRegistry))
	}

	assertEventPayloadRegistryCoverage(t, cases)

	for _, testCase := range cases {
		t.Run(testCase.metricName, func(t *testing.T) {
			t.Parallel()

			event, err := parseEventResourceMetadata("event-fixture-message", testCase.metadata)
			if err != nil {
				t.Fatalf("parse event: %v", err)
			}

			if event.EndpointID != eventFixtureEndpointID {
				t.Errorf("endpoint ID = %q, want %q", event.EndpointID, eventFixtureEndpointID)
			}

			if event.MessageID != "event-fixture-message" {
				t.Errorf("message ID = %q, want event-fixture-message", event.MessageID)
			}

			if got := reflect.TypeOf(event.Payload); got != testCase.wantType {
				t.Fatalf("payload type = %v, want %v", got, testCase.wantType)
			}

			assertEventPayloadMetadata(t, event.Payload, testCase.lastChange)

			want := reflect.New(testCase.wantType.Elem())

			wantJSON, err := json.Marshal(testCase.want)
			if err != nil {
				t.Fatalf("marshal expected payload: %v", err)
			}

			err = json.Unmarshal(wantJSON, want.Interface())
			if err != nil {
				t.Fatalf("unmarshal expected payload: %v", err)
			}

			got := normalizeEventPayload(event.Payload)

			wantNormalized := normalizeEventPayload(want.Interface())
			if !reflect.DeepEqual(got, wantNormalized) {
				t.Errorf("payload values = %#v, want %#v", got, wantNormalized)
			}
		})
	}
}

func assertEventPayloadRegistryCoverage(t *testing.T, cases []eventPayloadCase) {
	t.Helper()

	seen := make(map[string]struct{}, len(cases))

	for _, testCase := range cases {
		if _, exists := eventParserRegistry[testCase.metricName]; !exists {
			t.Errorf("test case covers unregistered metric %q", testCase.metricName)
		}

		if _, exists := seen[testCase.metricName]; exists {
			t.Errorf("duplicate event payload case for %q", testCase.metricName)
		}

		seen[testCase.metricName] = struct{}{}
	}

	for metricName := range eventParserRegistry {
		if _, exists := seen[metricName]; !exists {
			t.Errorf("registered metric %q has no event payload case", metricName)
		}
	}
}

func assertEventPayloadMetadata(t *testing.T, payload any, lastChange bool) {
	t.Helper()

	value := reflect.ValueOf(payload).Elem()
	sampleTime := eventPayloadTime(t, value, "TimeOfSample")

	if !sampleTime.Equal(time.Date(2025, time.December, 2, 5, 17, 49, 0, time.UTC)) {
		t.Errorf("time of sample = %s", sampleTime)
	}

	if got := value.FieldByName("Accuracy").String(); got != eventFixtureAccuracy {
		t.Errorf("accuracy = %q, want HIGH", got)
	}

	if got := value.FieldByName("Type").String(); got != eventFixturePropertyType {
		t.Errorf("type = %q, want RETRIEVABLE", got)
	}

	if got := value.FieldByName("Error").Interface(); !reflect.DeepEqual(got, &alexaapimodels.Error{Type: "ENDPOINT_UNREACHABLE"}) {
		t.Errorf("error = %#v", got)
	}

	if lastChange {
		got := eventPayloadTime(t, value, "TimeOfLastChange")

		want := time.Date(2025, time.December, 1, 5, 17, 49, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("time of last change = %s, want %s", got, want)
		}
	}
}

func eventPayloadTime(t *testing.T, payload reflect.Value, fieldName string) time.Time {
	t.Helper()

	field := payload.FieldByName(fieldName)
	if !field.IsValid() {
		t.Fatalf("payload has no %s field", fieldName)
	}

	timestamp, isTime := field.Interface().(time.Time)
	if !isTime {
		t.Fatalf("%s is %T, want time.Time", fieldName, field.Interface())
	}

	return timestamp
}

func normalizeEventPayload(payload any) any {
	value := reflect.ValueOf(payload)
	normalized := reflect.New(value.Elem().Type())
	normalized.Elem().Set(value.Elem())

	for _, fieldName := range []string{
		"TimeOfSample", "TimeOfLastChange", "Accuracy", "Type", "Error",
	} {
		field := normalized.Elem().FieldByName(fieldName)
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.Zero(field.Type()))
		}
	}

	return normalized.Interface()
}

func TestEventPayloadRejectsInvalidTimestamps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		metric    string
		fields    map[string]any
		wantError error
	}{
		{
			name:      "missing sample time",
			metric:    "EndpointPower",
			fields:    map[string]any{"timeOfSample": ""},
			wantError: errEventSampleTimeMissing,
		},
		{
			name:      "invalid sample time",
			metric:    "EndpointPower",
			fields:    map[string]any{"timeOfSample": "not-a-time"},
			wantError: errEventTimestampInvalid,
		},
		{
			name:   "invalid optional change time",
			metric: "EndpointAction",
			fields: map[string]any{
				"timeOfLastChange": "not-a-time",
			},
			wantError: errEventTimestampInvalid,
		},
		{
			name:   "invalid optional interval start",
			metric: "EndpointAction",
			fields: map[string]any{
				"actionStateValue": map[string]any{
					"timeInterval": map[string]any{"start": "not-a-time"},
				},
			},
			wantError: errEventTimestampInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			metadata := eventResourceMetadataForTest(t, test.metric, test.fields)

			_, err := parseEventResourceMetadata("event-fixture-message", metadata)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want errors.Is(_, %v)", err, test.wantError)
			}
		})
	}
}

func TestEventPayloadPreservesOptionalValues(t *testing.T) {
	t.Parallel()

	speakerMetadata := eventResourceMetadataForTest(t, "EndpointSpeaker", map[string]any{
		"error": nil,
	})

	speakerEvent, err := parseEventResourceMetadata("message", speakerMetadata)
	if err != nil {
		t.Fatalf("parse speaker event: %v", err)
	}

	speaker, payloadIsSpeaker := speakerEvent.Payload.(*alexaapimodels.SpeakerPayload)
	if !payloadIsSpeaker {
		t.Fatalf("speaker payload type = %T", speakerEvent.Payload)
	}

	if speaker.Volume != nil || speaker.Muted != nil || speaker.Error != nil {
		t.Errorf("absent speaker values were populated: %#v", speaker)
	}

	actionMetadata := eventResourceMetadataForTest(t, "EndpointAction", map[string]any{
		"error": nil,
		"actionStateValue": map[string]any{
			"timeInterval": map[string]any{
				"start":    "",
				"duration": "PT30M",
			},
		},
	})

	actionEvent, err := parseEventResourceMetadata("message", actionMetadata)
	if err != nil {
		t.Fatalf("parse action event: %v", err)
	}

	action, payloadIsAction := actionEvent.Payload.(*alexaapimodels.ActionStatePayload)
	if !payloadIsAction {
		t.Fatalf("action payload type = %T", actionEvent.Payload)
	}

	if action.TimeInterval == nil || action.TimeInterval.Start != nil || action.TimeInterval.End != nil {
		t.Fatalf("empty optional interval timestamps: %#v", action.TimeInterval)
	}

	if action.TimeInterval.Duration == nil || *action.TimeInterval.Duration != "PT30M" {
		t.Errorf("duration = %v, want PT30M", action.TimeInterval.Duration)
	}

	if action.Error != nil {
		t.Errorf("absent action error was populated: %#v", action.Error)
	}
}

func eventResourceMetadataForTest(t *testing.T, metricName string, fields map[string]any) string {
	t.Helper()

	return newEventPayloadCase(t, metricName, fields, (*alexaapimodels.PowerPayload)(nil), map[string]any{}, false).metadata
}
