package alexa

import "github.com/portpowered/go-alexa/pkg/alexaapimodels"

// featureDefaults defines the canonical default properties and operations for each
// supported Alexa Smart Home feature type. Property names align with the Alexa Smart
// Home API specification and the property models in dependencymodels/event.go.
//
// Only the Name field is set for default properties — Type, Accuracy, TimeOfSample,
// TimeOfLastChange, Error, and StateValue are populated at runtime when state data
// is available from the GraphQL API.
//
// Reference: https://developer.amazon.com/en-US/docs/alexa/device-apis/list-of-interfaces.html
var featureDefaults = map[alexaapimodels.FeatureName]featureDefaultEntry{
	// connectivity — Alexa.EndpointHealth connectivity property
	alexaapimodels.FeatureNameConnectivity: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "reachabilityState"},
		},
	},

	// location — Alexa.Location geolocation properties
	alexaapimodels.FeatureNameLocation: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "geolocation"},
		},
	},

	// locationTracker — Alexa.Location.Tracker operations
	alexaapimodels.FeatureNameLocationTracker: {
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "Locate"},
		},
	},

	// playback — Alexa.PlaybackController operations
	alexaapimodels.FeatureNamePlayback: {
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "Play"},
			{Name: "Pause"},
			{Name: "Next"},
			{Name: "Previous"},
			{Name: "Stop"},
			{Name: "FastForward"},
			{Name: "Rewind"},
			{Name: "StartOver"},
		},
	},

	// speaker — Alexa.Speaker properties
	alexaapimodels.FeatureNameSpeaker: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "volume"},
			{Name: "muted"},
		},
	},

	// power — Alexa.PowerController property
	alexaapimodels.FeatureNamePower: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "powerState"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "TurnOn"},
			{Name: "TurnOff"},
		},
	},

	// brightness — Alexa.BrightnessController property
	alexaapimodels.FeatureNameBrightness: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "brightness"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetBrightness"},
			{Name: "AdjustBrightness"},
		},
	},

	// color — Alexa.ColorController property
	alexaapimodels.FeatureNameColor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "color"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetColor"},
		},
	},

	// colorTemperature — Alexa.ColorTemperatureController property
	alexaapimodels.FeatureNameColorTemperature: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "colorTemperatureInKelvin"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetColorTemperature"},
			{Name: "IncreaseColorTemperature"},
			{Name: "DecreaseColorTemperature"},
		},
	},

	// lock — Alexa.LockController property
	alexaapimodels.FeatureNameLock: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "lockState"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "Lock"},
			{Name: "Unlock"},
		},
	},

	// mode — Alexa.ModeController property
	alexaapimodels.FeatureNameMode: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "mode"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetMode"},
			{Name: "AdjustMode"},
		},
	},

	// range — Alexa.RangeController property
	alexaapimodels.FeatureNameRange: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "rangeValue"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetRangeValue"},
			{Name: "AdjustRangeValue"},
		},
	},

	// toggle — Alexa.ToggleController property
	alexaapimodels.FeatureNameToggle: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "toggleState"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "TurnOn"},
			{Name: "TurnOff"},
		},
	},

	// percentage — Alexa.PercentageController property
	alexaapimodels.FeatureNamePercentage: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "percentage"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetPercentage"},
			{Name: "AdjustPercentage"},
		},
	},

	// powerLevel — Alexa.PowerLevelController property
	alexaapimodels.FeatureNamePowerLevel: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "powerLevel"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetPowerLevel"},
			{Name: "AdjustPowerLevel"},
		},
	},

	// thermostat — Alexa.ThermostatController properties
	alexaapimodels.FeatureNameThermostat: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "thermostatMode"},
			{Name: "targetSetpoint"},
			{Name: "lowerSetpoint"},
			{Name: "upperSetpoint"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "SetThermostatMode"},
			{Name: "SetTargetTemperature"},
			{Name: "AdjustTargetTemperature"},
		},
	},

	// temperatureSensor — Alexa.TemperatureSensor property (read-only)
	alexaapimodels.FeatureNameTemperatureSensor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "temperature"},
		},
	},

	// endpointHealth — Alexa.EndpointHealth battery property
	alexaapimodels.FeatureNameEndpointHealth: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "battery"},
		},
	},

	// securityPanel — Alexa.SecurityPanelController property
	alexaapimodels.FeatureNameSecurityPanel: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "armState"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "Arm"},
			{Name: "Disarm"},
		},
	},

	// humiditySensor — Alexa.HumiditySensor property (read-only)
	alexaapimodels.FeatureNameHumiditySensor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "relativeHumidity"},
		},
	},

	// lightSensor — Alexa.LightSensor property (read-only)
	alexaapimodels.FeatureNameLightSensor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "illuminance"},
		},
	},

	// motionSensor — Alexa.MotionSensor property (read-only)
	alexaapimodels.FeatureNameMotionSensor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "detectionState"},
		},
	},

	// contactSensor — Alexa.ContactSensor property (read-only)
	alexaapimodels.FeatureNameContactSensor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "detectionState"},
		},
	},

	// action — Alexa.SceneController property
	alexaapimodels.FeatureNameAction: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "actionState"},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: "Activate"},
			{Name: "Deactivate"},
		},
	},

	// statusCode — read-only status code property
	alexaapimodels.FeatureNameStatusCode: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: "statusCode"},
		},
	},
}

// featureDefaultEntry holds the default properties and operations for a feature type.
type featureDefaultEntry struct {
	Properties []alexaapimodels.FeatureProperty
	Operations []alexaapimodels.FeatureOperation
}

// getFeatureDefaults returns the default properties and operations for a given feature name.
// If no defaults are defined, it returns empty (non-nil) slices.
func getFeatureDefaults(name alexaapimodels.FeatureName) ([]alexaapimodels.FeatureProperty, []alexaapimodels.FeatureOperation) {
	if entry, ok := featureDefaults[name]; ok {
		props := make([]alexaapimodels.FeatureProperty, len(entry.Properties))
		copy(props, entry.Properties)
		ops := make([]alexaapimodels.FeatureOperation, len(entry.Operations))
		copy(ops, entry.Operations)
		return props, ops
	}
	return []alexaapimodels.FeatureProperty{}, []alexaapimodels.FeatureOperation{}
}
