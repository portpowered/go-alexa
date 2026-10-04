package alexaapimodels

// FeaturePropertyName names a property reported by a feature.
type FeaturePropertyName string

// FeatureOperationName represents the name of a control operation.
type FeatureOperationName string

// String returns the string representation of the operation name.
func (n FeatureOperationName) String() string {
	return string(n)
}

// ControlRequest represents a unified request for all control operations
// This replaces the individual control request types with a single dispatchable interface.
//
//modelinventory:client-input ControlRequest: common caller command envelope routed by namespace, operation, endpoint, and payload.
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
//
//modelinventory:client-input ControlNotificationPayload: notification message and title consumed by the send operation.
type ControlNotificationPayload struct {
	Message string `json:"message"`
	Title   string `json:"title"`
}

// ControlAnnouncementPayload represents the payload for announcement operations.
//
//modelinventory:client-input ControlAnnouncementPayload: announcement text, display method, title, and locale consumed by the announcement operation.
type ControlAnnouncementPayload struct {
	Message string `json:"message"`
	Method  string `json:"method"` // "speak", "show", or "all"
	Title   string `json:"title,omitempty"`
	Locale  string `json:"locale,omitempty"`
}

// ControlSpeechSynthesizerPayload represents the payload for text-to-speech operations.
//
//modelinventory:client-input ControlSpeechSynthesizerPayload: speech text passed to the text-to-speech operation.
type ControlSpeechSynthesizerPayload struct {
	Message string `json:"message"`
}

// ControlAudioPlayerPayload represents the payload for music playback operations.
//
//modelinventory:client-input ControlAudioPlayerPayload: provider, search phrase, and optional timer passed to music playback.
type ControlAudioPlayerPayload struct {
	ProviderID   ProviderID `json:"providerId"`
	SearchPhrase string     `json:"searchPhrase"`
	TimerSeconds *int       `json:"timerSeconds,omitempty"`
}

// ControlAudioPlayerURIPayload represents the payload for playing audio from a public HTTPS URI.
//
//modelinventory:client-input ControlAudioPlayerURIPayload: public media URI and optional title or playback offset passed to URI playback.
type ControlAudioPlayerURIPayload struct {
	URI                  string `json:"uri"`
	Title                string `json:"title,omitempty"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds,omitempty"`
}

// ControlNavigationPayload represents the payload for navigation operations (typically empty).
type ControlNavigationPayload struct{}

// ControlBrightnessSetPayload represents the payload for setting brightness.
//
//modelinventory:client-input ControlBrightnessSetPayload: absolute brightness value consumed by the set-brightness operation.
type ControlBrightnessSetPayload struct {
	Brightness int `json:"brightness"`
}

// ControlBrightnessAdjustPayload represents the payload for adjusting brightness.
//
//modelinventory:client-input ControlBrightnessAdjustPayload: brightness delta consumed by the adjust-brightness operation.
type ControlBrightnessAdjustPayload struct {
	Delta int `json:"delta"`
}

// ControlColorPayload represents the payload for setting color.
//
//modelinventory:client-input ControlColorPayload: hue, saturation, and brightness values consumed by the color operation.
type ControlColorPayload struct {
	Hue        float64 `json:"hue"`        // Hue value (0-360)
	Saturation float64 `json:"saturation"` // Saturation value (0-1)
	Brightness float64 `json:"brightness"` // Brightness value (0-1)
}

// ControlColorTemperatureSetPayload represents the payload for setting color temperature.
//
//modelinventory:client-input ControlColorTemperatureSetPayload: absolute Kelvin value consumed by color-temperature setting.
type ControlColorTemperatureSetPayload struct {
	ColorTemperature int `json:"colorTemperature"`
}

// ControlColorTemperatureAdjustPayload represents the payload for adjusting color temperature
// There is no delta, you can just up and down it.
//
//modelinventory:client-input ControlColorTemperatureAdjustPayload: direction flag consumed by color-temperature adjustment.
type ControlColorTemperatureAdjustPayload struct {
	// If increase, go up, else go down.
	Increase bool `json:"increase"`
}

// ControlLockPayload represents the payload for setting lock state.
//
//modelinventory:client-input ControlLockPayload: requested lock state consumed by the lock operation.
type ControlLockPayload struct {
	State string `json:"state"` // "LOCKED" or "UNLOCKED"
}

// ControlModeSetPayload represents the payload for setting mode
// Mode is opaque values that are unique to each capability interface.
//
//modelinventory:client-input ControlModeSetPayload: opaque mode value consumed by the set-mode operation.
type ControlModeSetPayload struct {
	Mode string `json:"mode"`
}

// ControlModeAdjustPayload represents the payload for adjusting mode
// Mode is opaque values that are unique to each capability interface.
//
//modelinventory:client-input ControlModeAdjustPayload: opaque mode delta consumed by the adjust-mode operation.
type ControlModeAdjustPayload struct {
	Delta string `json:"delta"`
}

// ControlRangeSetPayload represents the payload for setting range value.
//
//modelinventory:client-input ControlRangeSetPayload: absolute range value consumed by the set-range operation.
type ControlRangeSetPayload struct {
	RangeValue float64 `json:"rangeValue"`
}

// ControlRangeAdjustPayload represents the payload for adjusting range value.
//
//modelinventory:client-input ControlRangeAdjustPayload: range delta consumed by the adjust-range operation.
type ControlRangeAdjustPayload struct {
	Delta float64 `json:"delta"`
}

// ControlTogglePayload represents the payload for setting toggle state.
//
//modelinventory:client-input ControlTogglePayload: requested on or off state consumed by the toggle operation.
type ControlTogglePayload struct {
	State string `json:"state"` // "ON" or "OFF"
}

// ControlPercentageSetPayload represents the payload for setting percentage.
//
//modelinventory:client-input ControlPercentageSetPayload: absolute percentage consumed by the set-percentage operation.
type ControlPercentageSetPayload struct {
	Percentage float64 `json:"percentage"`
}

// ControlPercentageAdjustPayload represents the payload for adjusting percentage.
//
//modelinventory:client-input ControlPercentageAdjustPayload: percentage delta consumed by the adjust-percentage operation.
type ControlPercentageAdjustPayload struct {
	Delta float64 `json:"delta"`
}

// ControlPowerLevelSetPayload represents the payload for setting power level.
//
//modelinventory:client-input ControlPowerLevelSetPayload: absolute power level consumed by the set-power-level operation.
type ControlPowerLevelSetPayload struct {
	PowerLevel int `json:"powerLevel"`
}

// ControlPowerLevelAdjustPayload represents the payload for adjusting power level.
//
//modelinventory:client-input ControlPowerLevelAdjustPayload: power-level delta consumed by the adjust-power-level operation.
type ControlPowerLevelAdjustPayload struct {
	Delta int `json:"delta"`
}

// ControlActionPayload represents the payload for performing an action.
//
//modelinventory:client-input ControlActionPayload: action name and optional parameters passed to a feature action.
type ControlActionPayload struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ControlPlaybackPayload represents the payload for playback operations.
//
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type ControlPlaybackPayload struct {
}

// ControlForwardPayload represents the payload for forward operations.
//
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type ControlForwardPayload struct {
}

// ControlRewindPayload represents the payload for rewind operations.
//
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type ControlRewindPayload struct {
}

// ControlShufflePayload represents the payload for shuffle operations.
//
//modelinventory:client-input ControlShufflePayload: shuffle toggle consumed by playback control operations.
type ControlShufflePayload struct {
	Shuffle bool `json:"shuffle"` // true or false
}

// ControlRepeatPayload represents the payload for repeat operations.
//
//modelinventory:client-input ControlRepeatPayload: repeat toggle consumed by playback control operations.
type ControlRepeatPayload struct {
	Repeat bool `json:"repeat"` // true or false
}

// ControlPowerPayload represents the payload for power operations (state is determined by operation name).
type ControlPowerPayload struct{}

// ControlVolumeSetPayload represents the payload for setting volume.
//
//modelinventory:client-input ControlVolumeSetPayload: absolute playback volume converted into the selected device volume request.
type ControlVolumeSetPayload struct {
	Volume int `json:"volume"`
}

// ControlVolumeAdjustPayload represents the payload for adjusting volume.
//
//modelinventory:client-input ControlVolumeAdjustPayload: playback volume delta converted into the selected device volume request.
type ControlVolumeAdjustPayload struct {
	Delta int `json:"delta"`
}

// ControlThermostatSetpointPayload represents the payload for setting target temperature.
//
//modelinventory:client-input ControlThermostatSetpointPayload: target temperature and scale consumed by thermostat setpoint control.
type ControlThermostatSetpointPayload struct {
	Value float64 `json:"value"` // Temperature value
	Scale string  `json:"scale"` // "FAHRENHEIT", "CELSIUS", or "KELVIN"
}

// ControlThermostatSetpointAdjustPayload represents the payload for adjusting target temperature.
//
//modelinventory:client-input ControlThermostatSetpointAdjustPayload: temperature delta and scale consumed by thermostat setpoint adjustment.
type ControlThermostatSetpointAdjustPayload struct {
	Delta float64 `json:"delta"` // Temperature delta
	Scale string  `json:"scale"` // "FAHRENHEIT", "CELSIUS", or "KELVIN"
}

// ControlThermostatModePayload represents the payload for setting thermostat mode.
//
//modelinventory:client-input ControlThermostatModePayload: requested HVAC mode consumed by thermostat mode control.
type ControlThermostatModePayload struct {
	Mode string `json:"mode"` // "COOL", "HEAT", "OFF", "AUTO", "ECO", "EM_HEAT"
}

// PowerState represents whether a device is on or off.
type PowerState string
