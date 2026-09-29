package alexaapimodels

// FeaturePropertyName names a property reported by a feature.
type FeaturePropertyName string

// FeatureOperationName represents the name of a control operation.
type FeatureOperationName string

const (
	// FeatureOperationNameSend sends a notification.
	FeatureOperationNameSend FeatureOperationName = "send"

	// FeatureOperationNameSpeak speaks the supplied text.
	FeatureOperationNameSpeak FeatureOperationName = "speak"

	// FeatureOperationNamePlay starts audio playback.
	FeatureOperationNamePlay FeatureOperationName = "play"
	// FeatureOperationNamePlayURI plays media from a URI.
	FeatureOperationNamePlayURI FeatureOperationName = "playURI"

	// FeatureOperationNameNavigateHome navigates to the home view.
	FeatureOperationNameNavigateHome FeatureOperationName = "navigateHome"

	// FeatureOperationNameSetBrightness sets brightness.
	FeatureOperationNameSetBrightness FeatureOperationName = "setBrightness"
	// FeatureOperationNameAdjustBrightness adjusts brightness by a delta.
	FeatureOperationNameAdjustBrightness FeatureOperationName = "adjustBrightness"

	// FeatureOperationNameSetColor sets a color value.
	FeatureOperationNameSetColor FeatureOperationName = "setColor"

	// FeatureOperationNameSetColorTemperature sets color temperature.
	FeatureOperationNameSetColorTemperature FeatureOperationName = "setColorTemperature"
	// FeatureOperationNameIncreasColorTemperature adjusts color temperature upward.
	FeatureOperationNameIncreasColorTemperature FeatureOperationName = "adjustColorTemperature"
	// FeatureOperationNameDecreaseColorTemperature decreases color temperature.
	FeatureOperationNameDecreaseColorTemperature FeatureOperationName = "decreaseColorTemperature"

	// FeatureOperationNameSetLockState sets the lock state.
	FeatureOperationNameSetLockState FeatureOperationName = "setLockState"

	// FeatureOperationNameSetMode sets a mode value.
	FeatureOperationNameSetMode FeatureOperationName = "setMode"
	// FeatureOperationNameAdjustMode adjusts a mode value.
	FeatureOperationNameAdjustMode FeatureOperationName = "adjustMode"

	// FeatureOperationNameSetRangeValue sets a range value.
	FeatureOperationNameSetRangeValue FeatureOperationName = "setRangeValue"
	// FeatureOperationNameAdjustRangeValue adjusts a range value by a delta.
	FeatureOperationNameAdjustRangeValue FeatureOperationName = "adjustRangeValue"

	// FeatureOperationNameSetToggleState sets a toggle state.
	FeatureOperationNameSetToggleState FeatureOperationName = "setToggleState"

	// FeatureOperationNameSetPercentage sets a percentage value.
	FeatureOperationNameSetPercentage FeatureOperationName = "setPercentage"
	// FeatureOperationNameAdjustPercentage changes a percentage feature by a delta.
	FeatureOperationNameAdjustPercentage FeatureOperationName = "adjustPercentage"

	// FeatureOperationNameSetPowerLevel sets a power level.
	FeatureOperationNameSetPowerLevel FeatureOperationName = "setPowerLevel"
	// FeatureOperationNameAdjustPowerLevel changes a power level by a delta.
	FeatureOperationNameAdjustPowerLevel FeatureOperationName = "adjustPowerLevel"

	// FeatureOperationNamePerformAction performs an action.
	FeatureOperationNamePerformAction FeatureOperationName = "performAction"

	// FeatureOperationNamePause pauses playback.
	FeatureOperationNamePause FeatureOperationName = "pause"
	// FeatureOperationNameResume resumes playback.
	FeatureOperationNameResume FeatureOperationName = "resume"
	// FeatureOperationNameNext identifies this feature operation.
	FeatureOperationNameNext FeatureOperationName = "next"
	// FeatureOperationNamePrevious identifies this feature operation.
	FeatureOperationNamePrevious FeatureOperationName = "previous"
	// FeatureOperationNameStop identifies this feature operation.
	FeatureOperationNameStop FeatureOperationName = "stop"
	// FeatureOperationNameForward identifies this feature operation.
	FeatureOperationNameForward FeatureOperationName = "forward"
	// FeatureOperationNameRewind identifies this feature operation.
	FeatureOperationNameRewind FeatureOperationName = "rewind"
	// FeatureOperationNameShuffle identifies this feature operation.
	FeatureOperationNameShuffle FeatureOperationName = "shuffle"
	// FeatureOperationNameRepeat identifies this feature operation.
	FeatureOperationNameRepeat FeatureOperationName = "repeat"

	// FeatureOperationNameTurnOn turns a device on.
	FeatureOperationNameTurnOn FeatureOperationName = "turnOn"
	// FeatureOperationNameTurnOff identifies this feature operation.
	FeatureOperationNameTurnOff FeatureOperationName = "turnOff"

	// FeatureOperationNameSetVolume sets the playback volume.
	FeatureOperationNameSetVolume FeatureOperationName = "setVolume"
	// FeatureOperationNameAdjustVolume identifies this feature operation.
	FeatureOperationNameAdjustVolume FeatureOperationName = "adjustVolume"

	// FeatureOperationNameSetTargetSetpoint sets a thermostat target.
	FeatureOperationNameSetTargetSetpoint FeatureOperationName = "setTargetSetpoint"
	// FeatureOperationNameSetThermostatMode identifies this feature operation.
	FeatureOperationNameSetThermostatMode FeatureOperationName = "setThermostatMode"
	// FeatureOperationNameAdjustTargetSetpoint identifies this feature operation.
	FeatureOperationNameAdjustTargetSetpoint FeatureOperationName = "adjustTargetSetpoint"
)

// String returns the string representation of the operation name.
func (n FeatureOperationName) String() string {
	return string(n)
}

// ControlRequest represents a unified request for all control operations
// This replaces the individual control request types with a single dispatchable interface.
type ControlRequest struct {
	// Target is the endpoint to control
	Target EndpointInterface `json:"-"` // Not serialized, used to extract device info

	// Namespace is the namespace of the operation
	Namespace FeatureName `json:"namespace"`

	// Name is the name of the operation
	Name FeatureOperationName `json:"name"`

	// Instance is the optional instance identifier for multi-instance capabilities
	// (e.g., "Air Quality.indoorAirQuality" or "Vacuum.CleaningMode").
	// When set, it is propagated to the underlying feature request so the correct
	// instance on the endpoint is targeted.
	Instance string `json:"instance,omitempty"`

	// Payload is the request payload that would normally be passed in.
	// This should always be a struct with some values inside of it.
	// The struct type depends on the namespace and name combination.
	Payload any `json:"payload"`
}

// Control payload types - these are the payload structures for ControlRequest

// ControlNotificationPayload represents the payload for notification operations.
type ControlNotificationPayload struct {
	Message string `json:"message"`
	Title   string `json:"title"`
}

// ControlAnnouncementPayload represents the payload for announcement operations.
type ControlAnnouncementPayload struct {
	Message string `json:"message"`
	Method  string `json:"method"` // "speak", "show", or "all"
	Title   string `json:"title,omitempty"`
	Locale  string `json:"locale,omitempty"`
}

// ControlSpeechSynthesizerPayload represents the payload for text-to-speech operations.
type ControlSpeechSynthesizerPayload struct {
	Message string `json:"message"`
}

// ControlAudioPlayerPayload represents the payload for music playback operations.
type ControlAudioPlayerPayload struct {
	ProviderID   ProviderID `json:"providerId"`
	SearchPhrase string     `json:"searchPhrase"`
	TimerSeconds *int       `json:"timerSeconds,omitempty"`
}

// ControlAudioPlayerURIPayload represents the payload for playing audio from a public HTTPS URI.
type ControlAudioPlayerURIPayload struct {
	URI                  string `json:"uri"`
	Title                string `json:"title,omitempty"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds,omitempty"`
}

// ControlNavigationPayload represents the payload for navigation operations (typically empty).
type ControlNavigationPayload struct{}

// ControlBrightnessSetPayload represents the payload for setting brightness.
type ControlBrightnessSetPayload struct {
	Brightness int `json:"brightness"`
}

// ControlBrightnessAdjustPayload represents the payload for adjusting brightness.
type ControlBrightnessAdjustPayload struct {
	Delta int `json:"delta"`
}

// ControlColorPayload represents the payload for setting color.
type ControlColorPayload struct {
	Hue        float64 `json:"hue"`        // Hue value (0-360)
	Saturation float64 `json:"saturation"` // Saturation value (0-1)
	Brightness float64 `json:"brightness"` // Brightness value (0-1)
}

// ControlColorTemperatureSetPayload represents the payload for setting color temperature.
type ControlColorTemperatureSetPayload struct {
	ColorTemperature int `json:"colorTemperature"`
}

// ControlColorTemperatureAdjustPayload represents the payload for adjusting color temperature
// There is no delta, you can just up and down it.
type ControlColorTemperatureAdjustPayload struct {
	// If increase, go up, else go down.
	Increase bool `json:"increase"`
}

// ControlLockPayload represents the payload for setting lock state.
type ControlLockPayload struct {
	State string `json:"state"` // "LOCKED" or "UNLOCKED"
}

// ControlModeSetPayload represents the payload for setting mode
// Mode is opaque values that are unique to each capability interface.
type ControlModeSetPayload struct {
	Mode string `json:"mode"`
}

// ControlModeAdjustPayload represents the payload for adjusting mode
// Mode is opaque values that are unique to each capability interface.
type ControlModeAdjustPayload struct {
	Delta string `json:"delta"`
}

// ControlRangeSetPayload represents the payload for setting range value.
type ControlRangeSetPayload struct {
	RangeValue float64 `json:"rangeValue"`
}

// ControlRangeAdjustPayload represents the payload for adjusting range value.
type ControlRangeAdjustPayload struct {
	Delta float64 `json:"delta"`
}

// ControlTogglePayload represents the payload for setting toggle state.
type ControlTogglePayload struct {
	State string `json:"state"` // "ON" or "OFF"
}

// ControlPercentageSetPayload represents the payload for setting percentage.
type ControlPercentageSetPayload struct {
	Percentage float64 `json:"percentage"`
}

// ControlPercentageAdjustPayload represents the payload for adjusting percentage.
type ControlPercentageAdjustPayload struct {
	Delta float64 `json:"delta"`
}

// ControlPowerLevelSetPayload represents the payload for setting power level.
type ControlPowerLevelSetPayload struct {
	PowerLevel int `json:"powerLevel"`
}

// ControlPowerLevelAdjustPayload represents the payload for adjusting power level.
type ControlPowerLevelAdjustPayload struct {
	Delta int `json:"delta"`
}

// ControlActionPayload represents the payload for performing an action.
type ControlActionPayload struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ControlPlaybackPayload represents the payload for playback operations.
type ControlPlaybackPayload struct {
}

// ControlForwardPayload represents the payload for forward operations.
type ControlForwardPayload struct {
}

// ControlRewindPayload represents the payload for rewind operations.
type ControlRewindPayload struct {
}

// ControlShufflePayload represents the payload for shuffle operations.
type ControlShufflePayload struct {
	Shuffle bool `json:"shuffle"` // true or false
}

// ControlRepeatPayload represents the payload for repeat operations.
type ControlRepeatPayload struct {
	Repeat bool `json:"repeat"` // true or false
}

// ControlPowerPayload represents the payload for power operations (state is determined by operation name).
type ControlPowerPayload struct{}

// ControlVolumeSetPayload represents the payload for setting volume.
type ControlVolumeSetPayload struct {
	Volume int `json:"volume"`
}

// ControlVolumeAdjustPayload represents the payload for adjusting volume.
type ControlVolumeAdjustPayload struct {
	Delta int `json:"delta"`
}

// ControlThermostatSetpointPayload represents the payload for setting target temperature.
type ControlThermostatSetpointPayload struct {
	Value float64 `json:"value"` // Temperature value
	Scale string  `json:"scale"` // "FAHRENHEIT", "CELSIUS", or "KELVIN"
}

// ControlThermostatSetpointAdjustPayload represents the payload for adjusting target temperature.
type ControlThermostatSetpointAdjustPayload struct {
	Delta float64 `json:"delta"` // Temperature delta
	Scale string  `json:"scale"` // "FAHRENHEIT", "CELSIUS", or "KELVIN"
}

// ControlThermostatModePayload represents the payload for setting thermostat mode.
type ControlThermostatModePayload struct {
	Mode string `json:"mode"` // "COOL", "HEAT", "OFF", "AUTO", "ECO", "EM_HEAT"
}

// PowerState represents whether a device is on or off.
type PowerState string

const (
	// PowerStateOn is the on state of a device.
	PowerStateOn PowerState = "ON"
	// PowerStateOff is the off state of a device.
	PowerStateOff PowerState = "OFF"
)
