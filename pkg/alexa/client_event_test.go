package alexa

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// generateRandomUUID generates a random UUID-like string
func generateRandomUUID() string {
	return uuid.New().String()
}

// generateRandomEndpointID generates a random Alexa endpoint ID
func generateRandomEndpointID() string {
	return "amzn1.alexa.endpoint." + generateRandomUUID()
}

func TestParseEvent(t *testing.T) {
	messageID1 := generateRandomUUID()
	endpointID1 := generateRandomEndpointID()
	resourceID1 := generateRandomUUID()

	messageID2 := generateRandomUUID()
	endpointID2 := generateRandomEndpointID()
	resourceID2 := generateRandomUUID()

	tests := []struct {
		name            string
		eventJSON       string
		expectedNS      string
		expectedName    string
		expectedMsgID   string
		expectedEpID    string
		validatePayload func(t *testing.T, payload interface{})
	}{
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
									"resourceMetadata": "{\"metricName\":\"EndpointTemperatureSensor\",\"payload\":{\"data\":{\"__typename\":\"Endpoint\",\"features\":[{\"name\":\"temperatureSensor\",\"__typename\":\"Feature\",\"properties\":[{\"__typename\":\"TemperatureSensor\",\"name\":\"temperature\",\"value\":{\"__typename\":\"Temperature\",\"value\":20.8,\"scale\":\"CELSIUS\"},\"type\":\"RETRIEVABLE\",\"timeOfSample\":\"2025-12-02T05:17:29.25Z\",\"accuracy\":\"HIGH\",\"error\":null}],\"instance\":null}]},\"entity\":{\"__typename\":\"Endpoint\",\"id\":\"%s\"},\"fragment\":\"fragment EndpointTemperatureSensor on Endpoint{features{name instance properties{...on TemperatureSensor{name value{value scale} timeOfSample accuracy type error{type}}}}}\"},\"type\":\"UPDATE_ENTITY\",\"timestamp\":\"2025-12-02T05:17:29.25Z\"}"
								}
							]
						}
					}
				}
			}`, messageID1, resourceID1, endpointID1),
			expectedNS:    alexaapimodels.FeatureNameTemperatureSensor.EventNamespace(),
			expectedName:  alexaapimodels.EventNameTemperatureState,
			expectedMsgID: messageID1,
			expectedEpID:  endpointID1,
			validatePayload: func(t *testing.T, payload interface{}) {
				p, ok := payload.(*alexamodels.TemperatureSensorProperty)
				if !ok {
					t.Fatalf("Expected payload type *alexamodels.TemperatureSensorProperty, got %T", payload)
				}

				if p.Name != "temperature" {
					t.Errorf("Expected property name 'temperature', got %s", p.Name)
				}

				if p.Value == nil {
					t.Fatal("Expected value to be non-nil")
				}

				if p.Value.Value != 20.8 {
					t.Errorf("Expected temperature value 20.8, got %f", p.Value.Value)
				}

				if p.Value.Scale != "CELSIUS" {
					t.Errorf("Expected temperature scale 'CELSIUS', got %s", p.Value.Scale)
				}

				if p.TimeOfSample != "2025-12-02T05:17:29.25Z" {
					t.Errorf("Expected timeOfSample '2025-12-02T05:17:29.25Z', got %s", p.TimeOfSample)
				}

				if p.Accuracy != "HIGH" {
					t.Errorf("Expected accuracy 'HIGH', got %s", p.Accuracy)
				}

				if p.Type != "RETRIEVABLE" {
					t.Errorf("Expected type 'RETRIEVABLE', got %s", p.Type)
				}
			},
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
									"resourceMetadata": "{\"metricName\":\"EndpointPower\",\"payload\":{\"data\":{\"__typename\":\"Endpoint\",\"features\":[{\"name\":\"power\",\"__typename\":\"Feature\",\"properties\":[{\"__typename\":\"Power\",\"name\":\"powerState\",\"powerStateValue\":\"OFF\",\"type\":\"RETRIEVABLE\",\"timeOfSample\":\"2025-12-02T05:17:49Z\",\"accuracy\":\"HIGH\",\"error\":null}],\"instance\":null}]},\"entity\":{\"__typename\":\"Endpoint\",\"id\":\"%s\"},\"fragment\":\"fragment EndpointPower on Endpoint{features{name instance properties{...on Power{name powerStateValue timeOfSample accuracy type error{type}}}}}\"},\"type\":\"UPDATE_ENTITY\",\"timestamp\":\"2025-12-02T05:17:49Z\"}"
								}
							]
						}
					}
				}
			}`, messageID2, resourceID2, endpointID2),
			expectedNS:    alexaapimodels.FeatureNamePower.EventNamespace(),
			expectedName:  alexaapimodels.EventNamePowerState,
			expectedMsgID: messageID2,
			expectedEpID:  endpointID2,
			validatePayload: func(t *testing.T, payload interface{}) {
				p, ok := payload.(*alexamodels.PowerProperty)
				if !ok {
					t.Fatalf("Expected payload type *alexamodels.PowerProperty, got %T", payload)
				}

				if p.Name != "powerState" {
					t.Errorf("Expected property name 'powerState', got %s", p.Name)
				}

				if p.PowerStateValue != "OFF" {
					t.Errorf("Expected powerStateValue 'OFF', got %s", p.PowerStateValue)
				}

				if p.TimeOfSample != "2025-12-02T05:17:49Z" {
					t.Errorf("Expected timeOfSample '2025-12-02T05:17:49Z', got %s", p.TimeOfSample)
				}

				if p.Accuracy != "HIGH" {
					t.Errorf("Expected accuracy 'HIGH', got %s", p.Accuracy)
				}

				if p.Type != "RETRIEVABLE" {
					t.Errorf("Expected type 'RETRIEVABLE', got %s", p.Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg alexamodels.Message
			if err := json.Unmarshal([]byte(tt.eventJSON), &msg); err != nil {
				t.Fatalf("Failed to unmarshal test event: %v", err)
			}

			event, err := ParseEvent(&msg)
			if err != nil {
				t.Fatalf("ParseEvent failed: %v", err)
			}

			// Verify event structure
			if event.Namespace != tt.expectedNS {
				t.Errorf("Expected namespace %s, got %s", tt.expectedNS, event.Namespace)
			}

			if event.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, event.Name)
			}

			if event.EndpointID != tt.expectedEpID {
				t.Errorf("Expected endpoint ID %s, got %s", tt.expectedEpID, event.EndpointID)
			}

			if event.MessageID != tt.expectedMsgID {
				t.Errorf("Expected message ID %s, got %s", tt.expectedMsgID, event.MessageID)
			}

			// Validate payload
			tt.validatePayload(t, event.Payload)
		})
	}
}
