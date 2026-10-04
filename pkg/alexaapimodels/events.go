// Package alexaapimodels provides unified API models for Alexa services.
package alexaapimodels

import (
	"time"
)

// Event represents a decomposed event from the Alexa event stream.
//
//modelinventory:semantic Event: public envelope for a parsed event namespace, name, endpoint, message ID, and payload.
type Event struct {
	Namespace string `json:"namespace"`
	// The name of the event within the context of the namespace
	Name string `json:"name"`
	// The target endpoint ID for the target of the event
	EndpointID string `json:"endpointId"`
	// A unique identifier for the event
	MessageID string      `json:"messageId,omitempty"`
	Payload   interface{} `json:"payload"` // One of the specific payload types below
}

// ColorTemperaturePayload represents a color temperature update event.
//
//modelinventory:semantic ColorTemperaturePayload: color-temperature event value with normalized sample time and property metadata.
type ColorTemperaturePayload struct {
	ColorTemperatureInKelvin int       `json:"colorTemperatureInKelvin"`
	TimeOfSample             time.Time `json:"timeOfSample"`
	Accuracy                 string    `json:"accuracy"`
	Type                     string    `json:"type"`
	Error                    *Error    `json:"error,omitempty"`
}

// PowerPayload represents a power state update event.
//
//modelinventory:semantic PowerPayload: power state event value with normalized sample time and property metadata.
type PowerPayload struct {
	PowerState   string    `json:"powerState"` // "ON" or "OFF"
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// SpeakerPayload represents a speaker/volume update event.
//
//modelinventory:semantic SpeakerPayload: speaker volume and mute event values with optional-value preservation.
type SpeakerPayload struct {
	Volume       *int      `json:"volume,omitempty"`
	Muted        *bool     `json:"muted,omitempty"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// BrightnessPayload represents a brightness update event.
//
//modelinventory:semantic BrightnessPayload: brightness event value with optional numeric state preserved.
type BrightnessPayload struct {
	Brightness   *int      `json:"brightness,omitempty"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// ColorPayload represents a color update event.
//
//modelinventory:semantic ColorPayload: hue, saturation, and brightness event values with optional state preserved.
type ColorPayload struct {
	Hue          *float64  `json:"hue,omitempty"`
	Saturation   *float64  `json:"saturation,omitempty"`
	Brightness   *float64  `json:"brightness,omitempty"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// LockPayload represents a lock state update event.
//
//modelinventory:semantic LockPayload: lock state event value with normalized sample time and property metadata.
type LockPayload struct {
	LockState    string    `json:"lockState"` // "LOCKED" or "UNLOCKED"
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// ModePayload represents a mode update event.
//
//modelinventory:semantic ModePayload: mode event value and optional instance identifier.
type ModePayload struct {
	Instance     *string   `json:"instance,omitempty"` // Instance identifier for multiple modes on same endpoint
	Mode         string    `json:"mode"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// RangePayload represents a range value update event.
//
//modelinventory:semantic RangePayload: range event value and optional feature instance identifier.
type RangePayload struct {
	Instance     *string   `json:"instance,omitempty"` // Instance identifier for multiple ranges on same endpoint
	RangeValue   *float64  `json:"rangeValue,omitempty"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// TogglePayload represents a toggle state update event.
//
//modelinventory:semantic TogglePayload: toggle state event value and optional feature instance identifier.
type TogglePayload struct {
	Instance     *string   `json:"instance,omitempty"` // Instance identifier for multiple toggles on same endpoint
	ToggleState  string    `json:"toggleState"`        // "ON" or "OFF"
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// PercentagePayload represents a percentage update event.
//
//modelinventory:semantic PercentagePayload: percentage event value with optional zero-preserving numeric state.
type PercentagePayload struct {
	Percentage   *float64  `json:"percentage,omitempty"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// PowerLevelPayload represents a power level update event.
//
//modelinventory:semantic PowerLevelPayload: power-level event value with optional zero-preserving numeric state.
type PowerLevelPayload struct {
	PowerLevel   *int      `json:"powerLevel,omitempty"`
	TimeOfSample time.Time `json:"timeOfSample"`
	Accuracy     string    `json:"accuracy"`
	Type         string    `json:"type"`
	Error        *Error    `json:"error,omitempty"`
}

// Error represents an error in an event payload.
//
//modelinventory:semantic Error: normalized event property error exposed through typed event payloads.
type Error struct {
	Type string `json:"type"`
}

// ThermostatModePayload represents a thermostat mode update event.
//
//modelinventory:semantic ThermostatModePayload: thermostat mode state converted from the registered endpoint event.
type ThermostatModePayload struct {
	ThermostatMode   string    `json:"thermostatMode"` // "COOL", "HEAT", "OFF", "AUTO", "ECO", "EM_HEAT"
	TimeOfSample     time.Time `json:"timeOfSample"`
	TimeOfLastChange time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy         string    `json:"accuracy"`
	Type             string    `json:"type"`
	Error            *Error    `json:"error,omitempty"`
}

// TemperatureSensorPayload represents a temperature sensor update event.
//
//modelinventory:semantic TemperatureSensorPayload: temperature sensor reading with scale and timestamp metadata.
type TemperatureSensorPayload struct {
	Value            *float64  `json:"value,omitempty"` // Temperature value
	Scale            string    `json:"scale,omitempty"` // "FAHRENHEIT", "CELSIUS", "KELVIN"
	TimeOfSample     time.Time `json:"timeOfSample"`
	TimeOfLastChange time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy         string    `json:"accuracy"`
	Type             string    `json:"type"`
	Error            *Error    `json:"error,omitempty"`
}

// DetectionStatePayload represents a detection state update event
// Used for contact sensors and motion sensors.
//
//modelinventory:semantic DetectionStatePayload: motion or contact detection state converted from registered endpoint events.
type DetectionStatePayload struct {
	DetectionState   string    `json:"detectionState"` // "DETECTED", "NOT_DETECTED", "UNKNOWN"
	TimeOfSample     time.Time `json:"timeOfSample"`
	TimeOfLastChange time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy         string    `json:"accuracy"`
	Type             string    `json:"type"`
	Error            *Error    `json:"error,omitempty"`
}

// ActionStatePayload represents an action state update event.
//
//modelinventory:semantic ActionStatePayload: action status, IDs, target IDs, and timing converted from an endpoint event.
type ActionStatePayload struct {
	Instance         *string       `json:"instance,omitempty"` // Instance identifier for multiple actions on same endpoint
	Status           string        `json:"status"`             // "IDLE", "UNAVAILABLE", "RUNNING", "PAUSED", "COMPLETED", "INCOMPLETE", "UNKNOWN"
	TimeInterval     *TimeInterval `json:"timeInterval,omitempty"`
	ActionID         string        `json:"actionId,omitempty"`
	TargetIDs        []string      `json:"targetIds,omitempty"`
	TimeOfSample     time.Time     `json:"timeOfSample"`
	TimeOfLastChange time.Time     `json:"timeOfLastChange,omitempty"`
	Accuracy         string        `json:"accuracy"`
	Type             string        `json:"type"`
	Error            *Error        `json:"error,omitempty"`
}

// TimeInterval represents a time interval for action states.
//
//modelinventory:semantic TimeInterval: optional action start, end, and ISO duration fields preserved in the public event model.
type TimeInterval struct {
	Start    *time.Time `json:"start,omitempty"`
	End      *time.Time `json:"end,omitempty"`
	Duration *string    `json:"duration,omitempty"`
}

// ReachabilityPayload represents a reachability update event.
//
//modelinventory:semantic ReachabilityPayload: endpoint reachability state and optional change time converted for consumers.
type ReachabilityPayload struct {
	ReachabilityStatus string    `json:"reachabilityStatus"` // "REACHABLE", "UNREACHABLE", "UNKNOWN"
	TimeOfSample       time.Time `json:"timeOfSample"`
	TimeOfLastChange   time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy           string    `json:"accuracy"`
	Type               string    `json:"type"`
	Error              *Error    `json:"error,omitempty"`
}

// BatteryPayload represents a battery update event.
//
//modelinventory:semantic BatteryPayload: battery percentage, health, and charging status converted from the registered event.
type BatteryPayload struct {
	LevelPercentage  *int                   `json:"levelPercentage,omitempty"`
	Health           *BatteryHealth         `json:"health,omitempty"`
	ChargingHealth   *BatteryChargingHealth `json:"chargingHealth,omitempty"`
	TimeOfSample     time.Time              `json:"timeOfSample"`
	TimeOfLastChange time.Time              `json:"timeOfLastChange,omitempty"`
	Accuracy         string                 `json:"accuracy"`
	Type             string                 `json:"type"`
	Error            *Error                 `json:"error,omitempty"`
}

// BatteryHealth represents battery health information.
//
//modelinventory:semantic BatteryHealth: battery state and reason list nested in the public battery event payload.
type BatteryHealth struct {
	State   string   `json:"state"` // "OK", "WARNING", "CRITICAL", "UNKNOWN"
	Reasons []string `json:"reasons"`
}

// BatteryChargingHealth represents battery charging health information.
//
//modelinventory:semantic BatteryChargingHealth: charging health state and reason nested in the public battery event payload.
type BatteryChargingHealth struct {
	State  string `json:"state"` // "OK", "WARNING", "CRITICAL", "UNKNOWN"
	Reason string `json:"reason,omitempty"`
}

// IlluminancePayload represents an illuminance update event.
//
//modelinventory:semantic IlluminancePayload: illuminance reading with normalized timestamps and optional change time.
type IlluminancePayload struct {
	Value            *float64  `json:"value,omitempty"`
	TimeOfSample     time.Time `json:"timeOfSample"`
	TimeOfLastChange time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy         string    `json:"accuracy"`
	Type             string    `json:"type"`
	Error            *Error    `json:"error,omitempty"`
}

// GeolocationPayload represents a geolocation update event.
//
//modelinventory:semantic GeolocationPayload: location reading and optional coordinate, altitude, heading, and speed details.
type GeolocationPayload struct {
	Coordinate       *GeolocationCoordinate `json:"coordinate,omitempty"`
	Altitude         *GeolocationAltitude   `json:"altitude,omitempty"`
	Heading          *GeolocationHeading    `json:"heading,omitempty"`
	Speed            *GeolocationSpeed      `json:"speed,omitempty"`
	Source           string                 `json:"source,omitempty"` // "SELF_REPORTED", "CUSTOMER_SET", "UNKNOWN"
	TimeOfSample     time.Time              `json:"timeOfSample"`
	TimeOfLastChange time.Time              `json:"timeOfLastChange,omitempty"`
	Accuracy         string                 `json:"accuracy"`
	Type             string                 `json:"type"`
	Error            *Error                 `json:"error,omitempty"`
}

// GeolocationCoordinate represents a geolocation coordinate.
//
//modelinventory:semantic GeolocationCoordinate: latitude and longitude values nested in the public geolocation event.
type GeolocationCoordinate struct {
	LatitudeInDegrees  float64  `json:"latitudeInDegrees"`
	LongitudeInDegrees float64  `json:"longitudeInDegrees"`
	AccuracyInMeters   *float64 `json:"accuracyInMeters,omitempty"`
}

// GeolocationAltitude represents geolocation altitude information.
//
//modelinventory:semantic GeolocationAltitude: altitude and accuracy nested in the public geolocation event.
type GeolocationAltitude struct {
	AltitudeInMeters float64  `json:"altitudeInMeters"`
	AccuracyInMeters *float64 `json:"accuracyInMeters,omitempty"`
}

// GeolocationHeading represents geolocation heading information.
//
//modelinventory:semantic GeolocationHeading: direction and accuracy nested in the public geolocation event.
type GeolocationHeading struct {
	DirectionInDegrees float64  `json:"directionInDegrees"`
	AccuracyInDegrees  *float64 `json:"accuracyInDegrees,omitempty"`
}

// GeolocationSpeed represents geolocation speed information.
//
//modelinventory:semantic GeolocationSpeed: speed and accuracy nested in the public geolocation event.
type GeolocationSpeed struct {
	SpeedInMetersPerSecond    float64  `json:"speedInMetersPerSecond"`
	AccuracyInMetersPerSecond *float64 `json:"accuracyInMetersPerSecond,omitempty"`
}

// StatusCodePayload represents a status code update event.
//
//modelinventory:semantic StatusCodePayload: status-code event collection with normalized timestamps and typed code entries.
type StatusCodePayload struct {
	Codes            []StatusCodeValue `json:"codes"`
	TimeOfSample     time.Time         `json:"timeOfSample"`
	TimeOfLastChange time.Time         `json:"timeOfLastChange,omitempty"`
	Accuracy         string            `json:"accuracy"`
	Type             string            `json:"type"`
	Error            *Error            `json:"error,omitempty"`
}

// StatusCodeValue represents a single status code value.
//
//modelinventory:semantic StatusCodeValue: individual detected status code and optional detection time in a status event.
type StatusCodeValue struct {
	Code            string     `json:"code,omitempty"`
	TimeOfDetection *time.Time `json:"timeOfDetection,omitempty"`
}

// ArmStatePayload represents an arm state update event (security panel).
//
//modelinventory:semantic ArmStatePayload: security panel arm state converted from the registered endpoint event.
type ArmStatePayload struct {
	ArmState         string    `json:"armState"` // "ARMED_AWAY", "ARMED_STAY", "ARMED_NIGHT", "DISARMED", "UNKNOWN"
	TimeOfSample     time.Time `json:"timeOfSample"`
	TimeOfLastChange time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy         string    `json:"accuracy"`
	Type             string    `json:"type"`
	Error            *Error    `json:"error,omitempty"`
}

// RelativeHumidityPayload represents a relative humidity update event.
//
//modelinventory:semantic RelativeHumidityPayload: humidity reading with normalized sample time and optional change time.
type RelativeHumidityPayload struct {
	Value            *float64  `json:"value,omitempty"`
	TimeOfSample     time.Time `json:"timeOfSample"`
	TimeOfLastChange time.Time `json:"timeOfLastChange,omitempty"`
	Accuracy         string    `json:"accuracy"`
	Type             string    `json:"type"`
	Error            *Error    `json:"error,omitempty"`
}
