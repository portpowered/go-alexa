package alexamodels

// BaseFeatureRequest is the base request for all feature operations
type BaseFeatureRequest struct {
	EndpointID string `json:"endpointId"`
	EntityID   string `json:"entityId,omitempty"` // Optional entity ID
	Instance   string `json:"instance,omitempty"` // Optional instance identifier
}

// SpeakerControlRequest represents a request to control speaker volume
type SpeakerControlRequest struct {
	BaseFeatureRequest
	Volume    *int `json:"volume,omitempty"`    // Set volume (0-100)
	Delta     *int `json:"delta,omitempty"`     // Adjust volume by delta
	SetVolume bool `json:"setVolume,omitempty"` // If true, use volume; if false, use delta
}

// BrightnessControlRequest represents a request to control brightness
type BrightnessControlRequest struct {
	BaseFeatureRequest
	Brightness    *int `json:"brightness,omitempty"`    // Set brightness (0-100)
	Delta         *int `json:"delta,omitempty"`         // Adjust brightness by delta
	SetBrightness bool `json:"setBrightness,omitempty"` // If true, use brightness; if false, use delta
}

// ColorControlRequest represents a request to set color
type ColorControlRequest struct {
	BaseFeatureRequest
	Hue        float64 `json:"hue"`        // Hue value (0-360)
	Saturation float64 `json:"saturation"` // Saturation value (0-1)
	Brightness float64 `json:"brightness"` // Brightness value (0-1)
}

// ColorTemperatureControlRequest represents a request to control color temperature
type ColorTemperatureControlRequest struct {
	BaseFeatureRequest
	ColorTemperature *int `json:"colorTemperature,omitempty"` // Set color temperature in Kelvin
	Increase         bool `json:"increase,omitempty"`         // If true, increase; if false, decrease. only use if Color Temperature is not set.
	SetTemperature   bool `json:"setTemperature,omitempty"`   // If true, use colorTemperature; if false, use delta/increase
}

// LockControlRequest represents a request to control lock state
type LockControlRequest struct {
	BaseFeatureRequest
	State string `json:"state"` // "LOCKED" or "UNLOCKED"
}

// ModeControlRequest represents a request to control mode
type ModeControlRequest struct {
	BaseFeatureRequest
	Mode    *string `json:"mode,omitempty"`    // Set mode value
	Delta   *string `json:"delta,omitempty"`   // Adjust mode by delta
	SetMode bool    `json:"setMode,omitempty"` // If true, use mode; if false, use delta
}

// RangeControlRequest represents a request to control range value
type RangeControlRequest struct {
	BaseFeatureRequest
	RangeValue *float64 `json:"rangeValue,omitempty"` // Set range value
	Delta      *float64 `json:"delta,omitempty"`      // Adjust range by delta
	SetValue   bool     `json:"setValue,omitempty"`   // If true, use rangeValue; if false, use delta
}

// ToggleControlRequest represents a request to control toggle state
type ToggleControlRequest struct {
	BaseFeatureRequest
	State string `json:"state"` // "ON" or "OFF"
}

// PercentageControlRequest represents a request to control percentage
type PercentageControlRequest struct {
	BaseFeatureRequest
	Percentage    *float64 `json:"percentage,omitempty"`    // Set percentage (0-100)
	Delta         *float64 `json:"delta,omitempty"`         // Adjust percentage by delta
	SetPercentage bool     `json:"setPercentage,omitempty"` // If true, use percentage; if false, use delta
}

// PowerLevelControlRequest represents a request to control power level
type PowerLevelControlRequest struct {
	BaseFeatureRequest
	PowerLevel    *int `json:"powerLevel,omitempty"`    // Set power level
	Delta         *int `json:"delta,omitempty"`         // Adjust power level by delta
	SetPowerLevel bool `json:"setPowerLevel,omitempty"` // If true, use powerLevel; if false, use delta
}

// ActionControlRequest represents a request to perform an action
type ActionControlRequest struct {
	BaseFeatureRequest
	Action string                 `json:"action"`           // Action identifier
	Params map[string]interface{} `json:"params,omitempty"` // Action parameters
}

// ThermostatControlRequest represents a request to control thermostat
type ThermostatControlRequest struct {
	BaseFeatureRequest
	Value       *float64 `json:"value,omitempty"`       // Set target temperature value
	Delta       *float64 `json:"delta,omitempty"`       // Adjust target temperature by delta
	Scale       string   `json:"scale"`                 // Temperature scale: "FAHRENHEIT", "CELSIUS", or "KELVIN"
	SetSetpoint bool     `json:"setSetpoint,omitempty"` // If true, use value; if false, use delta
}

// ThermostatModeControlRequest represents a request to control thermostat mode
type ThermostatModeControlRequest struct {
	BaseFeatureRequest
	Mode string `json:"mode"` // Thermostat mode: "COOL", "HEAT", "OFF", "AUTO", "ECO", "EM_HEAT"
}
