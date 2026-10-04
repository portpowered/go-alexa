package alexa

import (
	"fmt"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

const (
	errEventSampleTimeMissing staticError = "event payload timeOfSample is required"
	errEventTimestampInvalid  staticError = "event payload contains an invalid timestamp"
)

type eventPayloadMetadata struct {
	timeOfSample     time.Time
	timeOfLastChange time.Time
	accuracy         string
	typeName         string
	error            *alexaapimodels.Error
}

func eventPayloadConverter[WirePayload any, PublicPayload any](
	convert func(WirePayload, *string) (PublicPayload, error),
) func(*alexamodels.ResourceMetadataPayloadData) (any, error) {
	return func(data *alexamodels.ResourceMetadataPayloadData) (any, error) {
		property, instance, err := parsePayload[WirePayload](data)
		if err != nil {
			return nil, err
		}

		return convert(*property, instance)
	}
}

func eventPayloadFields(
	timeOfSample, timeOfLastChange string,
	providerError *alexamodels.EventPropertyError,
) (eventPayloadMetadata, error) {
	sample, err := parseEventTimestamp("timeOfSample", timeOfSample, true)
	if err != nil {
		return eventPayloadMetadata{}, err
	}

	lastChange, err := parseEventTimestamp("timeOfLastChange", timeOfLastChange, false)
	if err != nil {
		return eventPayloadMetadata{}, err
	}

	var convertedError *alexaapimodels.Error
	if providerError != nil {
		convertedError = &alexaapimodels.Error{Type: providerError.Type}
	}

	return eventPayloadMetadata{
		timeOfSample:     sample,
		timeOfLastChange: lastChange,
		accuracy:         "",
		typeName:         "",
		error:            convertedError,
	}, nil
}

func eventPropertyMetadata(
	timeOfSample, timeOfLastChange, accuracy, typeName string,
	providerError *alexamodels.EventPropertyError,
) (eventPayloadMetadata, error) {
	metadata, err := eventPayloadFields(timeOfSample, timeOfLastChange, providerError)
	if err != nil {
		return eventPayloadMetadata{}, err
	}

	metadata.accuracy = accuracy
	metadata.typeName = typeName

	return metadata, nil
}

func parseEventTimestamp(field, value string, required bool) (time.Time, error) {
	if value == "" {
		if required {
			return time.Time{}, fmt.Errorf("%w: %s", errEventSampleTimeMissing, field)
		}

		return time.Time{}, nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s %q: %w", errEventTimestampInvalid, field, value, err)
	}

	return parsed, nil
}

func optionalEventTimestamp(field string, value *string) (time.Time, bool, error) {
	if value == nil || *value == "" {
		return time.Time{}, false, nil
	}

	parsed, err := parseEventTimestamp(field, *value, true)
	if err != nil {
		return time.Time{}, false, err
	}

	return parsed, true, nil
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}

	return append([]string(nil), values...)
}

func convertColorTemperaturePayload(
	property alexamodels.ColorTemperatureProperty,
	_ *string,
) (*alexaapimodels.ColorTemperaturePayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	value := 0
	if property.ColorTemperatureInKelvinStateValue != nil {
		value = *property.ColorTemperatureInKelvinStateValue
	}

	return &alexaapimodels.ColorTemperaturePayload{
		ColorTemperatureInKelvin: value,
		TimeOfSample:             metadata.timeOfSample,
		Accuracy:                 metadata.accuracy,
		Type:                     metadata.typeName,
		Error:                    metadata.error,
	}, nil
}

func convertPowerPayload(
	property alexamodels.PowerProperty,
	_ *string,
) (*alexaapimodels.PowerPayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.PowerPayload{
		PowerState:   property.PowerStateValue,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertSpeakerPayload(
	property alexamodels.SpeakerProperty,
	_ *string,
) (*alexaapimodels.SpeakerPayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.SpeakerPayload{
		Volume:       clonePointer(property.Volume),
		Muted:        clonePointer(property.Muted),
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertBrightnessPayload(
	property alexamodels.BrightnessProperty,
	_ *string,
) (*alexaapimodels.BrightnessPayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.BrightnessPayload{
		Brightness:   clonePointer(property.BrightnessStateValue),
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertColorPayload(
	property alexamodels.ColorProperty,
	_ *string,
) (*alexaapimodels.ColorPayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	var hue, saturation, brightness *float64
	if value := property.ColorStateValue; value != nil {
		hue = floatPointer(value.Hue)
		saturation = floatPointer(value.Saturation)
		brightness = floatPointer(value.Brightness)
	}

	return &alexaapimodels.ColorPayload{
		Hue:          hue,
		Saturation:   saturation,
		Brightness:   brightness,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertLockPayload(
	property alexamodels.LockProperty,
	_ *string,
) (*alexaapimodels.LockPayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.LockPayload{
		LockState:    property.LockState,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertModePayload(
	property alexamodels.ModeProperty,
	instance *string,
) (*alexaapimodels.ModePayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	mode := ""
	if property.ModeValue != nil {
		mode = property.ModeValue.Value
	}

	return &alexaapimodels.ModePayload{
		Instance:     cloneString(instance),
		Mode:         mode,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertRangePayload(
	property alexamodels.RangeProperty,
	instance *string,
) (*alexaapimodels.RangePayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	var rangeValue *float64
	if property.RangeValue != nil {
		rangeValue = floatPointer(property.RangeValue.Value)
	}

	return &alexaapimodels.RangePayload{
		Instance:     cloneString(instance),
		RangeValue:   rangeValue,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertTogglePayload(
	property alexamodels.ToggleProperty,
	instance *string,
) (*alexaapimodels.TogglePayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.TogglePayload{
		Instance:     cloneString(instance),
		ToggleState:  property.ToggleStateValue,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertPercentagePayload(
	property alexamodels.PercentageProperty,
	_ *string,
) (*alexaapimodels.PercentagePayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	var percentage *float64
	if property.PercentageValue != nil {
		percentage = floatPointer(float64(*property.PercentageValue))
	}

	return &alexaapimodels.PercentagePayload{
		Percentage:   percentage,
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertPowerLevelPayload(
	property alexamodels.PowerLevelProperty,
	_ *string,
) (*alexaapimodels.PowerLevelPayload, error) {
	metadata, err := eventPropertyMetadata(property.TimeOfSample, "", property.Accuracy, property.Type, property.Error)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.PowerLevelPayload{
		PowerLevel:   clonePointer(property.PowerLevelValue),
		TimeOfSample: metadata.timeOfSample,
		Accuracy:     metadata.accuracy,
		Type:         metadata.typeName,
		Error:        metadata.error,
	}, nil
}

func convertThermostatModePayload(
	property alexamodels.ThermostatModeProperty,
	_ *string,
) (*alexaapimodels.ThermostatModePayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ThermostatModePayload{
		ThermostatMode:   property.ThermostatModeValue,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}, nil
}

func convertTemperatureSensorPayload(
	property alexamodels.TemperatureSensorProperty,
	_ *string,
) (*alexaapimodels.TemperatureSensorPayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	var value *float64

	var scale string

	if property.Value != nil {
		value = floatPointer(property.Value.Value)
		scale = property.Value.Scale
	}

	return &alexaapimodels.TemperatureSensorPayload{
		Value:            value,
		Scale:            scale,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}, nil
}

func convertDetectionStatePayload(
	property alexamodels.DetectionStateProperty,
	_ *string,
) (*alexaapimodels.DetectionStatePayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.DetectionStatePayload{
		DetectionState:   property.DetectionStateValue,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}, nil
}

func convertActionStatePayload(
	property alexamodels.ActionStateProperty,
	instance *string,
) (*alexaapimodels.ActionStatePayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	payload := &alexaapimodels.ActionStatePayload{
		Status:           "",
		TimeInterval:     nil,
		ActionID:         "",
		TargetIDs:        nil,
		Instance:         cloneString(instance),
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}
	if property.ActionStateValue == nil {
		return payload, nil
	}

	payload.Status = property.ActionStateValue.Status
	if property.ActionStateValue.ActionID != nil {
		payload.ActionID = *property.ActionStateValue.ActionID
	}

	if property.ActionStateValue.TargetIDs != nil {
		payload.TargetIDs = cloneStrings(*property.ActionStateValue.TargetIDs)
	}

	if property.ActionStateValue.TimeInterval != nil {
		interval, err := convertActionTimeInterval(
			property.ActionStateValue.TimeInterval.Duration,
			property.ActionStateValue.TimeInterval.End,
			property.ActionStateValue.TimeInterval.Start,
		)
		if err != nil {
			return nil, err
		}

		payload.TimeInterval = interval
	}

	return payload, nil
}

func convertActionTimeInterval(duration, endTimestamp, startTimestamp *string) (*alexaapimodels.TimeInterval, error) {
	startValue, startPresent, err := optionalEventTimestamp("timeInterval.start", startTimestamp)
	if err != nil {
		return nil, err
	}

	endValue, endPresent, err := optionalEventTimestamp("timeInterval.end", endTimestamp)
	if err != nil {
		return nil, err
	}

	var start, end *time.Time
	if startPresent {
		start = &startValue
	}

	if endPresent {
		end = &endValue
	}

	return &alexaapimodels.TimeInterval{
		Start:    start,
		End:      end,
		Duration: cloneString(duration),
	}, nil
}

func convertReachabilityPayload(
	property alexamodels.ReachabilityProperty,
	_ *string,
) (*alexaapimodels.ReachabilityPayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ReachabilityPayload{
		ReachabilityStatus: property.ReachabilityStatusValue,
		TimeOfSample:       metadata.timeOfSample,
		TimeOfLastChange:   metadata.timeOfLastChange,
		Accuracy:           metadata.accuracy,
		Type:               metadata.typeName,
		Error:              metadata.error,
	}, nil
}

func convertBatteryPayload(
	property alexamodels.BatteryProperty,
	_ *string,
) (*alexaapimodels.BatteryPayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	payload := &alexaapimodels.BatteryPayload{
		LevelPercentage:  nil,
		Health:           nil,
		ChargingHealth:   nil,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}
	if property.BatteryValue != nil {
		payload.LevelPercentage = clonePointer(property.BatteryValue.LevelPercentage)
		if health := property.BatteryValue.Health; health != nil {
			payload.Health = &alexaapimodels.BatteryHealth{
				State:   health.State,
				Reasons: cloneStrings(health.Reasons),
			}
		}

		if health := property.BatteryValue.ChargingHealth; health != nil {
			payload.ChargingHealth = &alexaapimodels.BatteryChargingHealth{
				State:  health.State,
				Reason: stringValue(health.Reason),
			}
		}
	}

	return payload, nil
}

func convertIlluminancePayload(
	property alexamodels.IlluminanceProperty,
	_ *string,
) (*alexaapimodels.IlluminancePayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	var value *float64
	if property.IlluminanceValue != nil {
		value = clonePointer(property.IlluminanceValue.Value)
	}

	return &alexaapimodels.IlluminancePayload{
		Value:            value,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}, nil
}

func convertGeolocationPayload(
	property alexamodels.GeolocationProperty,
	_ *string,
) (*alexaapimodels.GeolocationPayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	payload := &alexaapimodels.GeolocationPayload{
		Coordinate:       nil,
		Altitude:         nil,
		Heading:          nil,
		Speed:            nil,
		Source:           "",
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}
	if value := property.GeolocationValue; value != nil {
		payload.Source = stringValue(value.Source)
		if coordinate := value.Coordinate; coordinate != nil {
			payload.Coordinate = &alexaapimodels.GeolocationCoordinate{
				LatitudeInDegrees:  coordinate.LatitudeInDegrees,
				LongitudeInDegrees: coordinate.LongitudeInDegrees,
				AccuracyInMeters:   clonePointer(coordinate.AccuracyInMeters),
			}
		}

		if altitude := value.Altitude; altitude != nil {
			payload.Altitude = &alexaapimodels.GeolocationAltitude{
				AltitudeInMeters: altitude.AltitudeInMeters,
				AccuracyInMeters: clonePointer(altitude.AccuracyInMeters),
			}
		}

		if heading := value.Heading; heading != nil {
			payload.Heading = &alexaapimodels.GeolocationHeading{
				DirectionInDegrees: heading.DirectionInDegrees,
				AccuracyInDegrees:  clonePointer(heading.AccuracyInDegrees),
			}
		}

		if speed := value.Speed; speed != nil {
			payload.Speed = &alexaapimodels.GeolocationSpeed{
				SpeedInMetersPerSecond:    speed.SpeedInMetersPerSecond,
				AccuracyInMetersPerSecond: clonePointer(speed.AccuracyInMetersPerSecond),
			}
		}
	}

	return payload, nil
}

func convertStatusCodePayload(
	property alexamodels.StatusCodeProperty,
	_ *string,
) (*alexaapimodels.StatusCodePayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	payload := &alexaapimodels.StatusCodePayload{
		Codes:            nil,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}
	if property.StatusCodeValue != nil {
		payload.Codes = make([]alexaapimodels.StatusCodeValue, 0, len(property.StatusCodeValue.Value))

		for _, code := range property.StatusCodeValue.Value {
			converted, err := convertStatusCodeValue(code.Code, code.TimeOfDetection)
			if err != nil {
				return nil, err
			}

			payload.Codes = append(payload.Codes, converted)
		}
	}

	return payload, nil
}

func convertStatusCodeValue(code, timeOfDetection *string) (alexaapimodels.StatusCodeValue, error) {
	converted := alexaapimodels.StatusCodeValue{Code: "", TimeOfDetection: nil}
	if code != nil {
		converted.Code = *code
	}

	if timeOfDetection != nil && *timeOfDetection != "" {
		timestamp, err := parseEventTimestamp("timeOfDetection", *timeOfDetection, true)
		if err != nil {
			return alexaapimodels.StatusCodeValue{}, err
		}

		converted.TimeOfDetection = &timestamp
	}

	return converted, nil
}

func convertArmStatePayload(
	property alexamodels.ArmStateProperty,
	_ *string,
) (*alexaapimodels.ArmStatePayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ArmStatePayload{
		ArmState:         property.ArmStateValue,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}, nil
}

func convertRelativeHumidityPayload(
	property alexamodels.RelativeHumidityProperty,
	_ *string,
) (*alexaapimodels.RelativeHumidityPayload, error) {
	metadata, err := eventPropertyMetadata(
		property.TimeOfSample,
		property.TimeOfLastChange,
		property.Accuracy,
		property.Type,
		property.Error,
	)
	if err != nil {
		return nil, err
	}

	var value *float64
	if property.RelativeHumidityValue != nil {
		value = clonePointer(property.RelativeHumidityValue.Value)
	}

	return &alexaapimodels.RelativeHumidityPayload{
		Value:            value,
		TimeOfSample:     metadata.timeOfSample,
		TimeOfLastChange: metadata.timeOfLastChange,
		Accuracy:         metadata.accuracy,
		Type:             metadata.typeName,
		Error:            metadata.error,
	}, nil
}

func floatPointer(value float64) *float64 {
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
