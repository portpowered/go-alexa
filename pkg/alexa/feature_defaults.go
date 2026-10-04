package alexa

import (
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// featureDefaults defines the canonical default properties and operations for each
// supported Alexa Smart Home feature type. Property names align with the Alexa Smart
// Home API specification and the schema-generated discovery descriptor names.
//
// Only the Name field is set for default properties — Type, Accuracy, TimeOfSample,
// TimeOfLastChange, Error, and StateValue are populated at runtime when state data
// is available from the GraphQL API.
//
// Reference: https://developer.amazon.com/en-US/docs/alexa/device-apis/list-of-interfaces.html
var featureDefaults = map[alexaapimodels.FeatureName]featureDefaultEntry{
	// connectivity — Alexa.EndpointHealth connectivity property
	alexaapimodels.FeatureNameConnectivity: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameReachabilityState},
		},
	},

	// location — Alexa.Location geolocation properties
	alexaapimodels.FeatureNameLocation: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameGeolocation},
		},
	},

	// locationTracker — Alexa.Location.Tracker operations
	alexaapimodels.FeatureNameLocationTracker: {
		Properties: nil,
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameLocate},
		},
	},

	// playback — Alexa.PlaybackController operations
	alexaapimodels.FeatureNamePlayback: {
		Properties: nil,
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNamePlay},
			{Name: alexamodels.FeatureDefaultOperationNamePause},
			{Name: alexamodels.FeatureDefaultOperationNameNext},
			{Name: alexamodels.FeatureDefaultOperationNamePrevious},
			{Name: alexamodels.FeatureDefaultOperationNameStop},
			{Name: alexamodels.FeatureDefaultOperationNameFastForward},
			{Name: alexamodels.FeatureDefaultOperationNameRewind},
			{Name: alexamodels.FeatureDefaultOperationNameStartOver},
		},
	},

	// speaker — Alexa.Speaker properties
	alexaapimodels.FeatureNameSpeaker: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameVolume},
			{Name: alexamodels.FeatureDefaultPropertyNameMuted},
		},
	},

	// power — Alexa.PowerController property
	alexaapimodels.FeatureNamePower: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNamePowerState},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameTurnOn},
			{Name: alexamodels.FeatureDefaultOperationNameTurnOff},
		},
	},

	// brightness — Alexa.BrightnessController property
	alexaapimodels.FeatureNameBrightness: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameBrightness},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetBrightness},
			{Name: alexamodels.FeatureDefaultOperationNameAdjustBrightness},
		},
	},

	// color — Alexa.ColorController property
	alexaapimodels.FeatureNameColor: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameColor},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetColor},
		},
	},

	// colorTemperature — Alexa.ColorTemperatureController property
	alexaapimodels.FeatureNameColorTemperature: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameColorTemperatureInKelvin},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetColorTemperature},
			{Name: alexamodels.FeatureDefaultOperationNameIncreaseColorTemperature},
			{Name: alexamodels.FeatureDefaultOperationNameDecreaseColorTemperature},
		},
	},

	// lock — Alexa.LockController property
	alexaapimodels.FeatureNameLock: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameLockState},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameLock},
			{Name: alexamodels.FeatureDefaultOperationNameUnlock},
		},
	},

	// mode — Alexa.ModeController property
	alexaapimodels.FeatureNameMode: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameMode},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetMode},
			{Name: alexamodels.FeatureDefaultOperationNameAdjustMode},
		},
	},

	// range — Alexa.RangeController property
	alexaapimodels.FeatureNameRange: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameRangeValue},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetRangeValue},
			{Name: alexamodels.FeatureDefaultOperationNameAdjustRangeValue},
		},
	},

	// toggle — Alexa.ToggleController property
	alexaapimodels.FeatureNameToggle: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameToggleState},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameTurnOn},
			{Name: alexamodels.FeatureDefaultOperationNameTurnOff},
		},
	},

	// percentage — Alexa.PercentageController property
	alexaapimodels.FeatureNamePercentage: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNamePercentage},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetPercentage},
			{Name: alexamodels.FeatureDefaultOperationNameAdjustPercentage},
		},
	},

	// powerLevel — Alexa.PowerLevelController property
	alexaapimodels.FeatureNamePowerLevel: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNamePowerLevel},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetPowerLevel},
			{Name: alexamodels.FeatureDefaultOperationNameAdjustPowerLevel},
		},
	},

	// thermostat — Alexa.ThermostatController properties
	alexaapimodels.FeatureNameThermostat: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameThermostatMode},
			{Name: alexamodels.FeatureDefaultPropertyNameTargetSetpoint},
			{Name: alexamodels.FeatureDefaultPropertyNameLowerSetpoint},
			{Name: alexamodels.FeatureDefaultPropertyNameUpperSetpoint},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameSetThermostatMode},
			{Name: alexamodels.FeatureDefaultOperationNameSetTargetTemperature},
			{Name: alexamodels.FeatureDefaultOperationNameAdjustTargetTemperature},
		},
	},

	// temperatureSensor — Alexa.TemperatureSensor property (read-only)
	alexaapimodels.FeatureNameTemperatureSensor: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameTemperature},
		},
	},

	// endpointHealth — Alexa.EndpointHealth battery property
	alexaapimodels.FeatureNameEndpointHealth: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameBattery},
		},
	},

	// securityPanel — Alexa.SecurityPanelController property
	alexaapimodels.FeatureNameSecurityPanel: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameArmState},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameArm},
			{Name: alexamodels.FeatureDefaultOperationNameDisarm},
		},
	},

	// humiditySensor — Alexa.HumiditySensor property (read-only)
	alexaapimodels.FeatureNameHumiditySensor: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameRelativeHumidity},
		},
	},

	// lightSensor — Alexa.LightSensor property (read-only)
	alexaapimodels.FeatureNameLightSensor: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameIlluminance},
		},
	},

	// motionSensor — Alexa.MotionSensor property (read-only)
	alexaapimodels.FeatureNameMotionSensor: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameDetectionState},
		},
	},

	// contactSensor — Alexa.ContactSensor property (read-only)
	alexaapimodels.FeatureNameContactSensor: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameDetectionState},
		},
	},

	// action — Alexa.SceneController property
	alexaapimodels.FeatureNameAction: {
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameActionState},
		},
		Operations: []alexaapimodels.FeatureOperation{
			{Name: alexamodels.FeatureDefaultOperationNameActivate},
			{Name: alexamodels.FeatureDefaultOperationNameDeactivate},
		},
	},

	// statusCode — read-only status code property
	alexaapimodels.FeatureNameStatusCode: {
		Operations: nil,
		Properties: []alexaapimodels.FeatureProperty{
			{Name: alexamodels.FeatureDefaultPropertyNameStatusCode},
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
func getFeatureDefaults(
	name alexaapimodels.FeatureName,
) ([]alexaapimodels.FeatureProperty, []alexaapimodels.FeatureOperation) {
	if entry, ok := featureDefaults[name]; ok {
		props := make([]alexaapimodels.FeatureProperty, len(entry.Properties))
		copy(props, entry.Properties)
		ops := make([]alexaapimodels.FeatureOperation, len(entry.Operations))
		copy(ops, entry.Operations)

		return props, ops
	}

	return []alexaapimodels.FeatureProperty{}, []alexaapimodels.FeatureOperation{}
}
