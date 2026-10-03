//nolint:testpackage // Verifies private event parsing and payload conversion behavior.
package alexa

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// generateRandomUUID generates a random UUID-like string.
func generateRandomUUID() string {
	return uuid.New().String()
}

// generateRandomEndpointID generates a random Alexa endpoint ID.
func generateRandomEndpointID() string {
	return "amzn1.alexa.endpoint." + generateRandomUUID()
}

type parseEventTestCase struct {
	name            string
	eventJSON       string
	expectedNS      string
	expectedName    string
	expectedMsgID   string
	expectedEpID    string
	validatePayload func(t *testing.T, payload interface{})
}

func TestParseEvent(t *testing.T) {
	t.Parallel()

	messageID1 := generateRandomUUID()
	endpointID1 := generateRandomEndpointID()
	resourceID1 := generateRandomUUID()

	messageID2 := generateRandomUUID()
	endpointID2 := generateRandomEndpointID()
	resourceID2 := generateRandomUUID()
	temperatureMetadata := fmt.Sprintf(
		`{"metricName":"EndpointTemperatureSensor","payload":{"data":{"__typename":"Endpoint","features":[{"n`+
			`ame":"temperatureSensor","__typename":"Feature","properties":[{"__typename":"TemperatureSensor","nam`+
			`e":"temperature","value":{"__typename":"Temperature","value":20.8,"scale":"CELSIUS"},"type":"RETRIEV`+
			`ABLE","timeOfSample":"2025-12-02T05:17:29.25Z","accuracy":"HIGH","error":null}],"instance":null}]},"`+
			`entity":{"__typename":"Endpoint","id":"%s"},"fragment":"fragment EndpointTemperatureSensor on Endpoi`+
			`nt{features{name instance properties{...on TemperatureSensor{name value{value scale} timeOfSample ac`+
			`curacy type error{type}}}}}"},"type":"UPDATE_ENTITY","timestamp":"2025-12-02T05:17:29.25Z"}`,
		endpointID1,
	)
	powerMetadata := fmt.Sprintf(
		`{"metricName":"EndpointPower","payload":{"data":{"__typename":"Endpoint","features":[{"name":"power"`+
			`,"__typename":"Feature","properties":[{"__typename":"Power","name":"powerState","powerStateValue":"O`+
			`FF","type":"RETRIEVABLE","timeOfSample":"2025-12-02T05:17:49Z","accuracy":"HIGH","error":null}],"ins`+
			`tance":null}]},"entity":{"__typename":"Endpoint","id":"%s"},"fragment":"fragment EndpointPower on En`+
			`dpoint{features{name instance properties{...on Power{name powerStateValue timeOfSample accuracy type`+
			` error{type}}}}}"},"type":"UPDATE_ENTITY","timestamp":"2025-12-02T05:17:49Z"}`,
		endpointID2,
	)

	tests := []parseEventTestCase{
		{
			name: "TemperatureSensor",
			eventJSON: fmt.Sprintf(`{
				"data": {
					"directive": {
						"header": {
							"namespace": "Alexa.Mobile.Push",
							"name": "RenderUpdate",
							"messageId": "%s"
						},
						"payload": {
							"renderingUpdates": [
								{
									"route": "EventBus:AlexaMobile::FDAL",
									"resourceId": "%s",
									"resourceMetadata": %q
								}
							]
						}
					}
				}
			}`, messageID1, resourceID1, temperatureMetadata),
			expectedNS:      alexaapimodels.FeatureNameTemperatureSensor.EventNamespace(),
			expectedName:    alexaapimodels.EventNameTemperatureState,
			expectedMsgID:   messageID1,
			expectedEpID:    endpointID1,
			validatePayload: assertTemperaturePayload,
		},
		{
			name: "Power",
			eventJSON: fmt.Sprintf(`{
				"data": {
					"directive": {
						"header": {
							"namespace": "Alexa.Mobile.Push",
							"name": "RenderUpdate",
							"messageId": "%s"
						},
						"payload": {
							"renderingUpdates": [
								{
									"route": "EventBus:AlexaMobile::FDAL",
									"resourceId": "%s",
									"resourceMetadata": %q
								}
							]
						}
					}
				}
			}`, messageID2, resourceID2, powerMetadata),
			expectedNS:      alexaapimodels.FeatureNamePower.EventNamespace(),
			expectedName:    alexaapimodels.EventNamePowerState,
			expectedMsgID:   messageID2,
			expectedEpID:    endpointID2,
			validatePayload: assertPowerPayload,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertParsedEvent(t, testCase)
		})
	}
}

func assertParsedEvent(t *testing.T, testCase parseEventTestCase) {
	t.Helper()

	var message alexamodels.Message

	err := json.Unmarshal([]byte(testCase.eventJSON), &message)
	if err != nil {
		t.Fatalf("Failed to unmarshal test event: %v", err)
	}

	event, err := ParseEvent(&message)
	if err != nil {
		t.Fatalf("ParseEvent failed: %v", err)
	}

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}

	err = json.Unmarshal([]byte(testCase.eventJSON), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal synthetic event envelope: %v", err)
	}

	var directive alexamodels.DirectiveMessage

	err = json.Unmarshal(envelope.Data, &directive)
	if err != nil {
		t.Fatalf("Failed to unmarshal synthetic directive model: %v", err)
	}

	generatedEvent, err := parseDirectiveMessage(&directive)
	if err != nil {
		t.Fatalf("parseDirectiveMessage failed: %v", err)
	}

	if !reflect.DeepEqual(event, generatedEvent) {
		t.Fatalf("generated model result differs from legacy parser: %#v != %#v", generatedEvent, event)
	}

	assertParsedEventMetadata(t, event, testCase)
	testCase.validatePayload(t, event.Payload)
}

func assertParsedEventMetadata(t *testing.T, event *alexaapimodels.Event, testCase parseEventTestCase) {
	t.Helper()

	if event.Namespace != testCase.expectedNS {
		t.Errorf("Expected namespace %s, got %s", testCase.expectedNS, event.Namespace)
	}

	if event.Name != testCase.expectedName {
		t.Errorf("Expected name %s, got %s", testCase.expectedName, event.Name)
	}

	if event.EndpointID != testCase.expectedEpID {
		t.Errorf("Expected endpoint ID %s, got %s", testCase.expectedEpID, event.EndpointID)
	}

	if event.MessageID != testCase.expectedMsgID {
		t.Errorf("Expected message ID %s, got %s", testCase.expectedMsgID, event.MessageID)
	}
}

func assertTemperaturePayload(t *testing.T, payload interface{}) {
	t.Helper()

	property, ok := payload.(*alexaapimodels.TemperatureSensorPayload)
	if !ok {
		t.Fatalf("Expected payload type *alexaapimodels.TemperatureSensorPayload, got %T", payload)
	}

	if property.Value == nil {
		t.Fatal("Expected value to be non-nil")
	}

	if *property.Value != 20.8 {
		t.Errorf("Expected temperature value 20.8, got %f", *property.Value)
	}

	if property.Scale != "CELSIUS" {
		t.Errorf("Expected temperature scale 'CELSIUS', got %s", property.Scale)
	}

	if !property.TimeOfSample.Equal(time.Date(2025, time.December, 2, 5, 17, 29, 250000000, time.UTC)) {
		t.Errorf("Unexpected timeOfSample: %s", property.TimeOfSample)
	}

	if property.Accuracy != "HIGH" {
		t.Errorf("Expected accuracy 'HIGH', got %s", property.Accuracy)
	}

	if property.Type != "RETRIEVABLE" {
		t.Errorf("Expected type 'RETRIEVABLE', got %s", property.Type)
	}
}

func assertPowerPayload(t *testing.T, payload interface{}) {
	t.Helper()

	property, ok := payload.(*alexaapimodels.PowerPayload)
	if !ok {
		t.Fatalf("Expected payload type *alexaapimodels.PowerPayload, got %T", payload)
	}

	if property.PowerState != "OFF" {
		t.Errorf("Expected power state 'OFF', got %s", property.PowerState)
	}

	if !property.TimeOfSample.Equal(time.Date(2025, time.December, 2, 5, 17, 49, 0, time.UTC)) {
		t.Errorf("Unexpected timeOfSample: %s", property.TimeOfSample)
	}

	if property.Accuracy != "HIGH" {
		t.Errorf("Expected accuracy 'HIGH', got %s", property.Accuracy)
	}

	if property.Type != "RETRIEVABLE" {
		t.Errorf("Expected type 'RETRIEVABLE', got %s", property.Type)
	}
}

func TestCompatibilityEventRejectsMalformedGeneratedDirectiveShapes(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		`{}`, `{"directive":null}`, `{"directive":{"header":null}}`,
		`{"directive":{"header":{},"payload":null}}`,
		`{"directive":{"header":{},"payload":{"renderingUpdates":[]}}}`,
		`{"directive":{"header":{},"payload":{"renderingUpdates":[{"resourceMetadata":4}]}}}`,
	} {
		var message alexamodels.Message

		err := json.Unmarshal([]byte(`{"data":`+input+`}`), &message)
		if err != nil {
			t.Fatalf("decode synthetic compatibility input: %v", err)
		}

		_, err = ParseEvent(&message)
		if err == nil {
			t.Fatalf("malformed directive accepted: %s", input)
		}
	}
}
