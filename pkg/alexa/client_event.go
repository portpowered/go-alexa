package alexa

import (
	"encoding/json"
	"fmt"

	directivewire "github.com/portpowered/go-alexa/pkg/alexa/internal/wire"
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
	"EndpointColorTemperature": {
		namespace: alexaapimodels.FeatureNameColorTemperature.EventNamespace(),
		name:      alexaapimodels.EventNameColorTemperatureState,
		converter: parsePayload[alexamodels.ColorTemperatureProperty],
	},
	"EndpointPower": {
		namespace: alexaapimodels.FeatureNamePower.EventNamespace(),
		name:      alexaapimodels.EventNamePowerState,
		converter: parsePayload[alexamodels.PowerProperty],
	},
	"EndpointSpeaker": {
		namespace: alexaapimodels.FeatureNameSpeaker.EventNamespace(),
		name:      alexaapimodels.EventNameSpeakerState,
		converter: parsePayload[alexamodels.SpeakerProperty],
	},
	"EndpointBrightness": {
		namespace: alexaapimodels.FeatureNameBrightness.EventNamespace(),
		name:      alexaapimodels.EventNameBrightnessState,
		converter: parsePayload[alexamodels.BrightnessProperty],
	},
	"EndpointColor": {
		namespace: alexaapimodels.FeatureNameColor.EventNamespace(),
		name:      alexaapimodels.EventNameColorState,
		converter: parsePayload[alexamodels.ColorProperty],
	},
	"EndpointLock": {
		namespace: alexaapimodels.FeatureNameLock.EventNamespace(),
		name:      alexaapimodels.EventNameLockState,
		converter: parsePayload[alexamodels.LockProperty],
	},
	"EndpointMode": {
		namespace: alexaapimodels.FeatureNameMode.EventNamespace(),
		name:      alexaapimodels.EventNameModeState,
		converter: parsePayload[alexamodels.ModeProperty],
	},
	"EndpointRange": {
		namespace: alexaapimodels.FeatureNameRange.EventNamespace(),
		name:      alexaapimodels.EventNameRangeState,
		converter: parsePayload[alexamodels.RangeProperty],
	},
	"EndpointToggle": {
		namespace: alexaapimodels.FeatureNameToggle.EventNamespace(),
		name:      alexaapimodels.EventNameToggleState,
		converter: parsePayload[alexamodels.ToggleProperty],
	},
	"EndpointPercentage": {
		namespace: alexaapimodels.FeatureNamePercentage.EventNamespace(),
		name:      alexaapimodels.EventNamePercentageState,
		converter: parsePayload[alexamodels.PercentageProperty],
	},
	"EndpointPowerLevel": {
		namespace: alexaapimodels.FeatureNamePowerLevel.EventNamespace(),
		name:      alexaapimodels.EventNamePowerLevelState,
		converter: parsePayload[alexamodels.PowerLevelProperty],
	},
	"EndpointThermostat": {
		namespace: alexaapimodels.FeatureNameThermostat.EventNamespace(),
		name:      alexaapimodels.EventNameThermostatModeState,
		converter: parsePayload[alexamodels.ThermostatModeProperty],
	},
	"EndpointTemperatureSensor": {
		namespace: alexaapimodels.FeatureNameTemperatureSensor.EventNamespace(),
		name:      alexaapimodels.EventNameTemperatureState,
		converter: parsePayload[alexamodels.TemperatureSensorProperty],
	},
	"EndpointMotionSensor": {
		namespace: alexaapimodels.FeatureNameMotionSensor.EventNamespace(),
		name:      alexaapimodels.EventNameDetectionState,
		converter: parsePayload[alexamodels.DetectionStateProperty],
	},
	"EndpointContactSensor": {
		namespace: alexaapimodels.FeatureNameContactSensor.EventNamespace(),
		name:      alexaapimodels.EventNameDetectionState,
		converter: parsePayload[alexamodels.DetectionStateProperty],
	},
	"EndpointAction": {
		namespace: alexaapimodels.FeatureNameAction.EventNamespace(),
		name:      alexaapimodels.EventNameActionState,
		converter: parsePayload[alexamodels.ActionStateProperty],
	},
	"EndpointConnectivity": {
		namespace: alexaapimodels.FeatureNameConnectivity.EventNamespace(),
		name:      alexaapimodels.EventNameReachabilityState,
		converter: parsePayload[alexamodels.ReachabilityProperty],
	},
	"EndpointEndpointHealth": {
		namespace: alexaapimodels.FeatureNameEndpointHealth.EventNamespace(),
		name:      alexaapimodels.EventNameBatteryState,
		converter: parsePayload[alexamodels.BatteryProperty],
	},
	"EndpointLightSensor": {
		namespace: alexaapimodels.FeatureNameLightSensor.EventNamespace(),
		name:      alexaapimodels.EventNameIlluminanceState,
		converter: parsePayload[alexamodels.IlluminanceProperty],
	},
	"EndpointLocation": {
		namespace: alexaapimodels.FeatureNameLocation.EventNamespace(),
		name:      alexaapimodels.EventNameGeolocationState,
		converter: parsePayload[alexamodels.GeolocationProperty],
	},
	"EndpointStatusCode": {
		namespace: alexaapimodels.FeatureNameStatusCode.EventNamespace(),
		name:      alexaapimodels.EventNameStatusCodeState,
		converter: parsePayload[alexamodels.StatusCodeProperty],
	},
	"EndpointSecurityPanel": {
		namespace: alexaapimodels.FeatureNameSecurityPanel.EventNamespace(),
		name:      alexaapimodels.EventNameArmState,
		converter: parsePayload[alexamodels.ArmStateProperty],
	},
	"EndpointHumiditySensor": {
		namespace: alexaapimodels.FeatureNameHumiditySensor.EventNamespace(),
		name:      alexaapimodels.EventNameRelativeHumidityState,
		converter: parsePayload[alexamodels.RelativeHumidityProperty],
	},
}

// ParseEvent decomposes a raw Message into a structured Event.
func ParseEvent(msg *alexamodels.Message) (*alexaapimodels.Event, error) {
	if msg == nil || msg.Data == nil {
		return nil, errEventMessageMissing
	}

	// Extract directive
	directiveRaw, directiveValid := msg.Data["directive"].(map[string]interface{})
	if !directiveValid {
		return nil, errDirectiveInvalid
	}

	// Extract header
	headerRaw, headerValid := directiveRaw["header"].(map[string]interface{})
	if !headerValid {
		return nil, errHeaderInvalid
	}

	messageID, _ := headerRaw["messageId"].(string)

	// Extract payload
	payloadRaw, payloadValid := directiveRaw["payload"].(map[string]interface{})
	if !payloadValid {
		return nil, errPayloadInvalid
	}

	// Extract renderingUpdates
	renderingUpdatesRaw, updatesValid := payloadRaw["renderingUpdates"].([]interface{})
	if !updatesValid || len(renderingUpdatesRaw) == 0 {
		return nil, errRenderingUpdatesEmpty
	}

	// Process the first rendering update (most common case)
	updateRaw, updateValid := renderingUpdatesRaw[0].(map[string]interface{})
	if !updateValid {
		return nil, errRenderingUpdateInvalid
	}

	resourceMetadataStr, metadataValid := updateRaw["resourceMetadata"].(string)
	if !metadataValid {
		return nil, errResourceMetadataInvalid
	}

	return parseEventResourceMetadata(messageID, resourceMetadataStr)
}

// parseDirectiveMessage converts the generated representation of the current
// HTTP/2 parser input into the public event model.
func parseDirectiveMessage(message *directivewire.DirectiveMessage) (*alexaapimodels.Event, error) {
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
func parsePayload[T any](data *alexamodels.ResourceMetadataPayloadData) (any, error) {
	if len(data.Data.Features) == 0 {
		return nil, errNoFeaturesFound
	}

	feature := data.Data.Features[0]

	var props []T

	err := json.Unmarshal(feature.Properties, &props)
	if err != nil {
		var prop T

		err := json.Unmarshal(feature.Properties, &prop)
		if err != nil {
			return nil, fmt.Errorf("failed to parse properties: %w", err)
		}

		props = []T{prop}
	}

	if len(props) == 0 {
		return nil, errNoPropertiesFound
	}

	prop := props[0]

	return &prop, nil
}
