// Package alexaapimodels provides unified API models for Alexa services.
package alexaapimodels

// FeatureName represents the namespace of a control operation
type FeatureName string

const (
	// FeatureNameNotification is for notification operations
	FeatureNameNotification FeatureName = "notification"
	// FeatureNameAnnouncement is for announcement operations
	FeatureNameAnnouncement FeatureName = "announcement"
	// FeatureNameSpeechSynthesizer is for text-to-speech operations
	FeatureNameSpeechSynthesizer FeatureName = "speechsynthesizer"
	// FeatureNameAudioPlayer is for music playback operations
	FeatureNameAudioPlayer FeatureName = "audioplayer"
	// FeatureNamespaceNavigation is for navigation operations
	FeatureNameNavigation FeatureName = "navigation"
	// FeatureNameBrightness is for brightness control operations
	FeatureNameBrightness FeatureName = "brightness"
	// FeatureNameColor is for color control operations
	FeatureNameColor FeatureName = "color"
	// FeatureNameColorTemperature is for color temperature control operations
	FeatureNameColorTemperature FeatureName = "colorTemperature"
	// FeatureNameLock is for lock control operations
	FeatureNameLock FeatureName = "lock"
	// FeatureNameMode is for mode control operations
	FeatureNameMode FeatureName = "mode"
	// FeatureNameRange is for range value control operations
	FeatureNameRange FeatureName = "range"
	// FeatureNameToggle is for toggle state control operations
	FeatureNameToggle FeatureName = "toggle"
	// FeatureNamePercentage is for percentage control operations
	FeatureNamePercentage FeatureName = "percentage"
	// FeatureNamePowerLevel is for power level control operations
	FeatureNamePowerLevel FeatureName = "powerLevel"
	// FeatureNameAction is for action control operations
	FeatureNameAction FeatureName = "action"
	// FeatureNamePlayback is for playback control operations
	FeatureNamePlayback FeatureName = "playback"
	// FeatureNamePower is for power control operations
	FeatureNamePower FeatureName = "power"
	// FeatureNameSpeaker is for speaker/volume control operations
	FeatureNameSpeaker FeatureName = "speaker"
	// FeatureNameThermostat is for thermostat control operations
	FeatureNameThermostat FeatureName = "thermostat"
	// FeatureNameTemperatureSensor is for temperature sensor operations (read-only)
	FeatureNameTemperatureSensor FeatureName = "temperatureSensor"
	// FeatureNameDetectionState is for detection state operations (read-only, contact/motion sensors)
	FeatureNameMotionSensor  FeatureName = "motionSensor"
	FeatureNameContactSensor FeatureName = "contactSensor"
	// FeatureNameReachability is for reachability operations (read-only)
	FeatureNameConnectivity FeatureName = "connectivity"
	// FeatureNameBattery is for battery operations (read-only)
	FeatureNameEndpointHealth FeatureName = "endpointHealth"
	// FeatureNameIlluminance is for illuminance operations (read-only)
	FeatureNameLightSensor FeatureName = "lightSensor"
	// FeatureNameGeolocation is for geolocation operations (read-only)
	FeatureNameLocation FeatureName = "location"
	// FeatureNameLocationTracker is for location tracker operations
	FeatureNameLocationTracker FeatureName = "locationTracker"
	// FeatureNameStatusCode is for status code operations (read-only)
	FeatureNameStatusCode FeatureName = "statusCode"
	// FeatureNameArmState is for arm state operations
	FeatureNameSecurityPanel FeatureName = "securityPanel"
	// FeatureNameRelativeHumidity is for relative humidity operations (read-only)
	FeatureNameHumiditySensor FeatureName = "humiditySensor"
)

// String returns the string representation of the namespace
func (n FeatureName) String() string {
	return string(n)
}

// EventNamespace returns the event namespace format for this control namespace
// Event namespaces use the format "alexa.endpoint.feature.{namespace}"
func (n FeatureName) EventNamespace() string {
	return "alexa.endpoint.feature." + string(n)
}
