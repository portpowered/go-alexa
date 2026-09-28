// Package models provides data structures for Alexa API interactions.
package alexamodels

import (
	"encoding/json"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// Event represents an event received from Alexa
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
}

// resourceMetadataPayload represents the structure of resourceMetadata JSON
type ResourceMetadataPayload struct {
	MetricName string          `json:"metricName"`
	Payload    json.RawMessage `json:"payload"`
	Type       string          `json:"type"`
	Timestamp  string          `json:"timestamp"`
}

// resourceMetadataPayloadData represents the data within the payload
type ResourceMetadataPayloadData struct {
	Data struct {
		TypeName string `json:"__typename"`
		Features []struct {
			Name       string          `json:"name"`
			TypeName   string          `json:"__typename"`
			Properties json.RawMessage `json:"properties"`
			Instance   *string         `json:"instance"`
		} `json:"features"`
	} `json:"data"`
	Entity struct {
		TypeName string `json:"__typename"`
		ID       string `json:"id"`
	} `json:"entity"`
}

// colorTemperatureProperty represents the ColorTemperature property structure
type ColorTemperatureProperty struct {
	TypeName                           string                `json:"__typename"`
	Name                               string                `json:"name"`
	ColorTemperatureInKelvinStateValue *int                  `json:"colorTemperatureInKelvinStateValue,omitempty"`
	TimeOfSample                       string                `json:"timeOfSample"`
	Accuracy                           string                `json:"accuracy"`
	Type                               string                `json:"type"`
	Error                              *alexaapimodels.Error `json:"error"`
}

// powerProperty represents the Power property structure
type PowerProperty struct {
	TypeName        string                `json:"__typename"`
	Name            string                `json:"name"`
	PowerStateValue string                `json:"powerStateValue"`
	TimeOfSample    string                `json:"timeOfSample"`
	Accuracy        string                `json:"accuracy"`
	Type            string                `json:"type"`
	Error           *alexaapimodels.Error `json:"error"`
}

// speakerProperty represents the Speaker property structure
type SpeakerProperty struct {
	TypeName     string                `json:"__typename"`
	Name         string                `json:"name"`
	Volume       *int                  `json:"volume,omitempty"`
	Muted        *bool                 `json:"muted,omitempty"`
	TimeOfSample string                `json:"timeOfSample"`
	Accuracy     string                `json:"accuracy"`
	Type         string                `json:"type"`
	Error        *alexaapimodels.Error `json:"error"`
}

// brightnessProperty represents the Brightness property structure
type BrightnessProperty struct {
	TypeName             string                `json:"__typename"`
	Name                 string                `json:"name"`
	BrightnessStateValue *int                  `json:"brightnessStateValue,omitempty"`
	TimeOfSample         string                `json:"timeOfSample"`
	Accuracy             string                `json:"accuracy"`
	Type                 string                `json:"type"`
	Error                *alexaapimodels.Error `json:"error"`
}

// colorProperty represents the Color property structure
// ColorStateValue is a nested object matching GraphQL ColorValue type
type ColorProperty struct {
	TypeName        string `json:"__typename"`
	Name            string `json:"name"`
	ColorStateValue *struct {
		Hue        float64 `json:"hue"`
		Saturation float64 `json:"saturation"`
		Brightness float64 `json:"brightness"`
	} `json:"colorStateValue,omitempty"`
	TimeOfSample string                `json:"timeOfSample"`
	Accuracy     string                `json:"accuracy"`
	Type         string                `json:"type"`
	Error        *alexaapimodels.Error `json:"error"`
}

// lockProperty represents the Lock property structure
type LockProperty struct {
	TypeName     string                `json:"__typename"`
	Name         string                `json:"name"`
	LockState    string                `json:"lockState"`
	TimeOfSample string                `json:"timeOfSample"`
	Accuracy     string                `json:"accuracy"`
	Type         string                `json:"type"`
	Error        *alexaapimodels.Error `json:"error"`
}

// modeProperty represents the Mode property structure
// ModeValue is a nested object matching GraphQL ModeValue type
type ModeProperty struct {
	TypeName  string `json:"__typename"`
	Name      string `json:"name"`
	ModeValue *struct {
		Value string `json:"value"`
	} `json:"modeValue,omitempty"`
	TimeOfSample string                `json:"timeOfSample"`
	Accuracy     string                `json:"accuracy"`
	Type         string                `json:"type"`
	Error        *alexaapimodels.Error `json:"error"`
}

// rangeProperty represents the Range property structure
// RangeValue is a nested object matching GraphQL RangeValueNumber type
type RangeProperty struct {
	TypeName   string `json:"__typename"`
	Name       string `json:"name"`
	RangeValue *struct {
		Value float64 `json:"value"`
	} `json:"rangeValue,omitempty"`
	TimeOfSample string                `json:"timeOfSample"`
	Accuracy     string                `json:"accuracy"`
	Type         string                `json:"type"`
	Error        *alexaapimodels.Error `json:"error"`
}

// toggleProperty represents the Toggle property structure
type ToggleProperty struct {
	TypeName         string                `json:"__typename"`
	Name             string                `json:"name"`
	ToggleStateValue string                `json:"toggleStateValue"`
	TimeOfSample     string                `json:"timeOfSample"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// percentageProperty represents the Percentage property structure
type PercentageProperty struct {
	TypeName        string                `json:"__typename"`
	Name            string                `json:"name"`
	PercentageValue *int                  `json:"percentageValue,omitempty"`
	TimeOfSample    string                `json:"timeOfSample"`
	Accuracy        string                `json:"accuracy"`
	Type            string                `json:"type"`
	Error           *alexaapimodels.Error `json:"error"`
}

// powerLevelProperty represents the PowerLevel property structure
type PowerLevelProperty struct {
	TypeName        string                `json:"__typename"`
	Name            string                `json:"name"`
	PowerLevelValue *int                  `json:"powerLevelValue,omitempty"`
	TimeOfSample    string                `json:"timeOfSample"`
	Accuracy        string                `json:"accuracy"`
	Type            string                `json:"type"`
	Error           *alexaapimodels.Error `json:"error"`
}

// thermostatModeProperty represents the ThermostatMode property structure
type ThermostatModeProperty struct {
	TypeName            string                `json:"__typename"`
	Name                string                `json:"name"`
	ThermostatModeValue string                `json:"thermostatModeValue"`
	TimeOfSample        string                `json:"timeOfSample"`
	TimeOfLastChange    string                `json:"timeOfLastChange"`
	Accuracy            string                `json:"accuracy"`
	Type                string                `json:"type"`
	Error               *alexaapimodels.Error `json:"error"`
}

// setpointProperty represents the Setpoint property structure
// Setpoint can represent lowerSetpoint, upperSetpoint, or targetSetpoint
type SetpointProperty struct {
	TypeName               string `json:"__typename"`
	Name                   string `json:"name"`
	DeviceNativeScaleValue string `json:"deviceNativeScaleValue"`
	Value                  *struct {
		Value float64 `json:"value"`
		Scale string  `json:"scale"`
	} `json:"value,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// temperatureSensorProperty represents the TemperatureSensor property structure
type TemperatureSensorProperty struct {
	TypeName string `json:"__typename"`
	Name     string `json:"name"`
	Value    *struct {
		Value float64 `json:"value"`
		Scale string  `json:"scale"`
	} `json:"value,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// detectionStateProperty represents the DetectionState property structure
// Used for contact sensors and motion sensors
type DetectionStateProperty struct {
	TypeName            string                `json:"__typename"`
	Name                string                `json:"name"`
	DetectionStateValue string                `json:"detectionStateValue"`
	TimeOfSample        string                `json:"timeOfSample"`
	TimeOfLastChange    string                `json:"timeOfLastChange"`
	Accuracy            string                `json:"accuracy"`
	Type                string                `json:"type"`
	Error               *alexaapimodels.Error `json:"error"`
}

// actionStateProperty represents the ActionState property structure
type ActionStateProperty struct {
	TypeName         string `json:"__typename"`
	Name             string `json:"name"`
	ActionStateValue *struct {
		Status       string `json:"status"`
		TimeInterval *struct {
			Start    string `json:"start,omitempty"`
			End      string `json:"end,omitempty"`
			Duration string `json:"duration,omitempty"`
		} `json:"timeInterval,omitempty"`
		ActionID  string   `json:"actionId,omitempty"`
		TargetIDs []string `json:"targetIds,omitempty"`
	} `json:"actionStateValue,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// reachabilityProperty represents the Reachability property structure
type ReachabilityProperty struct {
	TypeName                string                `json:"__typename"`
	Name                    string                `json:"name"`
	ReachabilityStatusValue string                `json:"reachabilityStatusValue"`
	TimeOfSample            string                `json:"timeOfSample"`
	TimeOfLastChange        string                `json:"timeOfLastChange"`
	Accuracy                string                `json:"accuracy"`
	Type                    string                `json:"type"`
	Error                   *alexaapimodels.Error `json:"error"`
}

// batteryProperty represents the Battery property structure
type BatteryProperty struct {
	TypeName     string `json:"__typename"`
	Name         string `json:"name"`
	BatteryValue *struct {
		LevelPercentage *int `json:"levelPercentage,omitempty"`
		Health          *struct {
			State   string   `json:"state"`
			Reasons []string `json:"reasons"`
		} `json:"health,omitempty"`
		ChargingHealth *struct {
			State  string `json:"state"`
			Reason string `json:"reason,omitempty"`
		} `json:"chargingHealth,omitempty"`
	} `json:"batteryValue,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// illuminanceProperty represents the Illuminance property structure
type IlluminanceProperty struct {
	TypeName         string `json:"__typename"`
	Name             string `json:"name"`
	IlluminanceValue *struct {
		Value *float64 `json:"value,omitempty"`
	} `json:"illuminanceValue,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// geolocationProperty represents the Geolocation property structure
type GeolocationProperty struct {
	TypeName         string `json:"__typename"`
	Name             string `json:"name"`
	GeolocationValue *struct {
		Coordinate *struct {
			LatitudeInDegrees  float64  `json:"latitudeInDegrees"`
			LongitudeInDegrees float64  `json:"longitudeInDegrees"`
			AccuracyInMeters   *float64 `json:"accuracyInMeters,omitempty"`
		} `json:"coordinate,omitempty"`
		Altitude *struct {
			AltitudeInMeters float64  `json:"altitudeInMeters"`
			AccuracyInMeters *float64 `json:"accuracyInMeters,omitempty"`
		} `json:"altitude,omitempty"`
		Heading *struct {
			DirectionInDegrees float64  `json:"directionInDegrees"`
			AccuracyInDegrees  *float64 `json:"accuracyInDegrees,omitempty"`
		} `json:"heading,omitempty"`
		Speed *struct {
			SpeedInMetersPerSecond    float64  `json:"speedInMetersPerSecond"`
			AccuracyInMetersPerSecond *float64 `json:"accuracyInMetersPerSecond,omitempty"`
		} `json:"speed,omitempty"`
		Source string `json:"source,omitempty"`
	} `json:"geolocationValue,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// statusCodeProperty represents the StatusCode property structure
type StatusCodeProperty struct {
	TypeName        string `json:"__typename"`
	Name            string `json:"name"`
	StatusCodeValue *struct {
		Value []struct {
			Code            string `json:"code,omitempty"`
			TimeOfDetection string `json:"timeOfDetection,omitempty"`
		} `json:"value"`
	} `json:"statusCodeValue,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// armStateProperty represents the ArmState property structure
type ArmStateProperty struct {
	TypeName         string                `json:"__typename"`
	Name             string                `json:"name"`
	ArmStateValue    string                `json:"armStateValue"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}

// relativeHumidityProperty represents the RelativeHumidity property structure
type RelativeHumidityProperty struct {
	TypeName              string `json:"__typename"`
	Name                  string `json:"name"`
	RelativeHumidityValue *struct {
		Value *float64 `json:"value,omitempty"`
	} `json:"relativeHumidityValue,omitempty"`
	TimeOfSample     string                `json:"timeOfSample"`
	TimeOfLastChange string                `json:"timeOfLastChange"`
	Accuracy         string                `json:"accuracy"`
	Type             string                `json:"type"`
	Error            *alexaapimodels.Error `json:"error"`
}
