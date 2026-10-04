package alexa

import (
	"encoding/json"
	"fmt"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// eventParserMetadata holds metadata for parsing a specific event type.
type eventParserMetadata[T any] struct {
	namespace string
	name      string
	converter func(*alexamodels.ResourceMetadataPayloadData) (T, error)
}

// eventParserRegistry maps metric names to their parsing metadata.
var eventParserRegistry = map[string]eventParserMetadata[any]{
	alexamodels.EventMetricNameEndpointColorTemperature: {
		namespace: alexaapimodels.FeatureNameColorTemperature.EventNamespace(),
		name:      alexaapimodels.EventNameColorTemperatureState,
		converter: eventPayloadConverter(convertColorTemperaturePayload),
	},
	alexamodels.EventMetricNameEndpointPower: {
		namespace: alexaapimodels.FeatureNamePower.EventNamespace(),
		name:      alexaapimodels.EventNamePowerState,
		converter: eventPayloadConverter(convertPowerPayload),
	},
	alexamodels.EventMetricNameEndpointSpeaker: {
		namespace: alexaapimodels.FeatureNameSpeaker.EventNamespace(),
		name:      alexaapimodels.EventNameSpeakerState,
		converter: eventPayloadConverter(convertSpeakerPayload),
	},
	alexamodels.EventMetricNameEndpointBrightness: {
		namespace: alexaapimodels.FeatureNameBrightness.EventNamespace(),
		name:      alexaapimodels.EventNameBrightnessState,
		converter: eventPayloadConverter(convertBrightnessPayload),
	},
	alexamodels.EventMetricNameEndpointColor: {
		namespace: alexaapimodels.FeatureNameColor.EventNamespace(),
		name:      alexaapimodels.EventNameColorState,
		converter: eventPayloadConverter(convertColorPayload),
	},
	alexamodels.EventMetricNameEndpointLock: {
		namespace: alexaapimodels.FeatureNameLock.EventNamespace(),
		name:      alexaapimodels.EventNameLockState,
		converter: eventPayloadConverter(convertLockPayload),
	},
	alexamodels.EventMetricNameEndpointMode: {
		namespace: alexaapimodels.FeatureNameMode.EventNamespace(),
		name:      alexaapimodels.EventNameModeState,
		converter: eventPayloadConverter(convertModePayload),
	},
	alexamodels.EventMetricNameEndpointRange: {
		namespace: alexaapimodels.FeatureNameRange.EventNamespace(),
		name:      alexaapimodels.EventNameRangeState,
		converter: eventPayloadConverter(convertRangePayload),
	},
	alexamodels.EventMetricNameEndpointToggle: {
		namespace: alexaapimodels.FeatureNameToggle.EventNamespace(),
		name:      alexaapimodels.EventNameToggleState,
		converter: eventPayloadConverter(convertTogglePayload),
	},
	alexamodels.EventMetricNameEndpointPercentage: {
		namespace: alexaapimodels.FeatureNamePercentage.EventNamespace(),
		name:      alexaapimodels.EventNamePercentageState,
		converter: eventPayloadConverter(convertPercentagePayload),
	},
	alexamodels.EventMetricNameEndpointPowerLevel: {
		namespace: alexaapimodels.FeatureNamePowerLevel.EventNamespace(),
		name:      alexaapimodels.EventNamePowerLevelState,
		converter: eventPayloadConverter(convertPowerLevelPayload),
	},
	alexamodels.EventMetricNameEndpointThermostat: {
		namespace: alexaapimodels.FeatureNameThermostat.EventNamespace(),
		name:      alexaapimodels.EventNameThermostatModeState,
		converter: eventPayloadConverter(convertThermostatModePayload),
	},
	alexamodels.EventMetricNameEndpointTemperatureSensor: {
		namespace: alexaapimodels.FeatureNameTemperatureSensor.EventNamespace(),
		name:      alexaapimodels.EventNameTemperatureState,
		converter: eventPayloadConverter(convertTemperatureSensorPayload),
	},
	alexamodels.EventMetricNameEndpointMotionSensor: {
		namespace: alexaapimodels.FeatureNameMotionSensor.EventNamespace(),
		name:      alexaapimodels.EventNameDetectionState,
		converter: eventPayloadConverter(convertDetectionStatePayload),
	},
	alexamodels.EventMetricNameEndpointContactSensor: {
		namespace: alexaapimodels.FeatureNameContactSensor.EventNamespace(),
		name:      alexaapimodels.EventNameDetectionState,
		converter: eventPayloadConverter(convertDetectionStatePayload),
	},
	alexamodels.EventMetricNameEndpointAction: {
		namespace: alexaapimodels.FeatureNameAction.EventNamespace(),
		name:      alexaapimodels.EventNameActionState,
		converter: eventPayloadConverter(convertActionStatePayload),
	},
	alexamodels.EventMetricNameEndpointConnectivity: {
		namespace: alexaapimodels.FeatureNameConnectivity.EventNamespace(),
		name:      alexaapimodels.EventNameReachabilityState,
		converter: eventPayloadConverter(convertReachabilityPayload),
	},
	alexamodels.EventMetricNameEndpointEndpointHealth: {
		namespace: alexaapimodels.FeatureNameEndpointHealth.EventNamespace(),
		name:      alexaapimodels.EventNameBatteryState,
		converter: eventPayloadConverter(convertBatteryPayload),
	},
	alexamodels.EventMetricNameEndpointLightSensor: {
		namespace: alexaapimodels.FeatureNameLightSensor.EventNamespace(),
		name:      alexaapimodels.EventNameIlluminanceState,
		converter: eventPayloadConverter(convertIlluminancePayload),
	},
	alexamodels.EventMetricNameEndpointLocation: {
		namespace: alexaapimodels.FeatureNameLocation.EventNamespace(),
		name:      alexaapimodels.EventNameGeolocationState,
		converter: eventPayloadConverter(convertGeolocationPayload),
	},
	alexamodels.EventMetricNameEndpointStatusCode: {
		namespace: alexaapimodels.FeatureNameStatusCode.EventNamespace(),
		name:      alexaapimodels.EventNameStatusCodeState,
		converter: eventPayloadConverter(convertStatusCodePayload),
	},
	alexamodels.EventMetricNameEndpointSecurityPanel: {
		namespace: alexaapimodels.FeatureNameSecurityPanel.EventNamespace(),
		name:      alexaapimodels.EventNameArmState,
		converter: eventPayloadConverter(convertArmStatePayload),
	},
	alexamodels.EventMetricNameEndpointHumiditySensor: {
		namespace: alexaapimodels.FeatureNameHumiditySensor.EventNamespace(),
		name:      alexaapimodels.EventNameRelativeHumidityState,
		converter: eventPayloadConverter(convertRelativeHumidityPayload),
	},
}

// ParseEvent converts a compatibility Message through the generated directive model.
func ParseEvent(msg *alexamodels.Message) (*alexaapimodels.Event, error) {
	if msg == nil || msg.Data == nil {
		return nil, errEventMessageMissing
	}

	data, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("encode compatibility event: %w", err)
	}

	var directive alexamodels.DirectiveMessage

	err = json.Unmarshal(data, &directive)
	if err != nil {
		return nil, fmt.Errorf("decode compatibility event: %w", err)
	}

	return parseDirectiveMessage(&directive)
}

// parseDirectiveMessage converts the generated representation of the current
// HTTP/2 parser input into the public event model.
func parseDirectiveMessage(message *alexamodels.DirectiveMessage) (*alexaapimodels.Event, error) {
	if message == nil || message.Directive == nil {
		return nil, errDirectiveInvalid
	}

	if message.Directive.Header == nil {
		return nil, errHeaderInvalid
	}

	if message.Directive.Payload == nil {
		return nil, errPayloadInvalid
	}

	if len(message.Directive.Payload.RenderingUpdates) == 0 {
		return nil, errRenderingUpdatesEmpty
	}

	update := message.Directive.Payload.RenderingUpdates[0]
	if update.ResourceMetadata == "" {
		return nil, errResourceMetadataInvalid
	}

	return parseEventResourceMetadata(message.Directive.Header.MessageId, update.ResourceMetadata)
}

func parseEventResourceMetadata(messageID, resourceMetadataStr string) (*alexaapimodels.Event, error) {
	// Parse resourceMetadata JSON string
	var resourceMetadata alexamodels.ResourceMetadataPayload
	{
		err := json.Unmarshal([]byte(resourceMetadataStr), &resourceMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to parse resourceMetadata: %w", err)
		}
	}

	// Parse the payload data
	var payloadData alexamodels.ResourceMetadataPayloadData
	{
		err := json.Unmarshal(resourceMetadata.Payload, &payloadData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse payload data: %w", err)
		}
	}

	// Extract endpoint ID
	endpointID := payloadData.Entity.ID

	// Look up parser metadata
	metadata, ok := eventParserRegistry[resourceMetadata.MetricName]
	if !ok {
		// For unknown types, return the raw data
		var rawData map[string]interface{}

		err := json.Unmarshal(resourceMetadata.Payload, &rawData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse unknown payload: %w", err)
		}

		return nil, &alexaapimodels.SdkError{
			Message: "unknown payload type: " + resourceMetadata.MetricName,
		}
	}

	// Convert payload using the registered converter
	eventPayload, err := metadata.converter(&payloadData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payload for %s: %w", resourceMetadata.MetricName, err)
	}

	return &alexaapimodels.Event{
		Namespace:  metadata.namespace,
		Name:       metadata.name,
		EndpointID: endpointID,
		MessageID:  messageID,
		Payload:    eventPayload,
	}, nil
}

// parsePayload extracts and parses a property from the payload data.
func parsePayload[T any](data *alexamodels.ResourceMetadataPayloadData) (*T, *string, error) {
	if data == nil || len(data.Data.Features) == 0 {
		return nil, nil, errNoFeaturesFound
	}

	feature := data.Data.Features[0]
	if len(feature.Properties) == 0 {
		return nil, nil, errNoPropertiesFound
	}

	property, arrayPayload, arrayErr := parsePayloadArray[T](feature.Properties)
	if arrayErr != nil {
		return nil, nil, arrayErr
	}

	if arrayPayload {
		return property, cloneString(feature.Instance), nil
	}

	property = new(T)

	propertyErr := json.Unmarshal(feature.Properties, property)
	if propertyErr != nil {
		return nil, nil, fmt.Errorf("failed to parse properties: %w", propertyErr)
	}

	return property, cloneString(feature.Instance), nil
}

func parsePayloadArray[T any](raw json.RawMessage) (*T, bool, error) {
	var properties []json.RawMessage

	propertiesErr := json.Unmarshal(raw, &properties)
	if propertiesErr == nil {
		if len(properties) == 0 {
			return nil, true, errNoPropertiesFound
		}

		property := new(T)

		propertyErr := json.Unmarshal(properties[0], property)
		if propertyErr != nil {
			return nil, true, fmt.Errorf("failed to parse properties: %w", propertyErr)
		}

		return property, true, nil
	}

	return nil, false, nil
}
