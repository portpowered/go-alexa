package graphql

import (
	"context"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// ControlPowerFeature controls power on an endpoint.
func (c *Client) ControlPowerFeature(
	ctx context.Context,
	req *alexamodels.PowerControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	operationName := FeatureOperationNameTurnoff
	if req.State == alexaapimodels.PowerStateOn {
		operationName = FeatureOperationNameTurnon
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.Endpoint.GetEndpointId(),
		EntityId:             "",
		FeatureName:          FeatureNamePower,
		Instance:             "",
		FeatureOperationName: operationName,
		Payload:              nil,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlSpeakerFeature controls speaker volume on an endpoint.
func (c *Client) ControlSpeakerFeature(
	ctx context.Context,
	req *alexamodels.SpeakerControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetVolume && req.Volume != nil:
		operationName = FeatureOperationNameSetvolume
		payload = alexamodels.SpeakerSetVolumePayload{Volume: *req.Volume}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjustvolume
		payload = alexamodels.SpeakerAdjustVolumePayload{VolumeDelta: *req.Delta}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either volume or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		FeatureName:          FeatureNameSpeaker,
		FeatureOperationName: operationName,
		Instance:             req.Instance,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlBrightnessFeature controls brightness on an endpoint.
func (c *Client) ControlBrightnessFeature(
	ctx context.Context,
	req *alexamodels.BrightnessControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetBrightness && req.Brightness != nil:
		operationName = FeatureOperationNameSetbrightness
		payload = alexamodels.BrightnessSetPayload{Brightness: *req.Brightness}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjustbrightness
		payload = alexamodels.BrightnessAdjustPayload{BrightnessDelta: *req.Delta}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either brightness or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameBrightness,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlColorFeature controls color on an endpoint.
func (c *Client) ControlColorFeature(
	ctx context.Context,
	req *alexamodels.ColorControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	payload := alexamodels.ColorSetPayload{
		Color: alexamodels.FeatureColorValue{
			Hue:        req.Hue,
			Saturation: req.Saturation,
			Brightness: req.Brightness,
		},
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameColor,
		FeatureOperationName: FeatureOperationNameSetcolor,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlColorTemperatureFeature controls color temperature on an endpoint.
func (c *Client) ControlColorTemperatureFeature(
	ctx context.Context,
	req *alexamodels.ColorTemperatureControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetTemperature && req.ColorTemperature != nil:
		operationName = FeatureOperationNameSetcolortemperature
		payload = alexamodels.ColorTemperatureSetPayload{ColorTemperatureInKelvin: *req.ColorTemperature}
	case req.Increase:
		operationName = FeatureOperationNameIncreasecolortemperature
		payload = alexamodels.FeatureEmptyPayload{}
	case !req.Increase:
		operationName = FeatureOperationNameDecreasecolortemperature
		payload = alexamodels.FeatureEmptyPayload{}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either colorTemperature or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameColortemperature,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlLockFeature controls lock state on an endpoint.
func (c *Client) ControlLockFeature(
	ctx context.Context,
	req *alexamodels.LockControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	operationName := FeatureOperationNameUnlock
	if req.State == string(alexamodels.LockStateValueLocked) {
		operationName = FeatureOperationNameLock
	}

	payload := alexamodels.LockPayload{LockState: req.State}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameLock,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlModeFeature controls mode on an endpoint.
func (c *Client) ControlModeFeature(
	ctx context.Context,
	req *alexamodels.ModeControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetMode && req.Mode != nil:
		operationName = FeatureOperationNameSetmode
		payload = alexamodels.ModeSetPayload{Mode: *req.Mode}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjustmode
		payload = alexamodels.ModeAdjustPayload{ModeDelta: *req.Delta}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either mode or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameMode,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlRangeFeature controls range value on an endpoint.
func (c *Client) ControlRangeFeature(
	ctx context.Context,
	req *alexamodels.RangeControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetValue && req.RangeValue != nil:
		operationName = FeatureOperationNameSetrangevalue
		payload = alexamodels.RangeSetPayload{RangeValue: *req.RangeValue}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjustrangevalue
		payload = alexamodels.RangeAdjustPayload{RangeValueDelta: *req.Delta}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either rangeValue or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameRange,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlToggleFeature controls toggle state on an endpoint.
func (c *Client) ControlToggleFeature(
	ctx context.Context,
	req *alexamodels.ToggleControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	operationName := FeatureOperationNameTurnoff
	if req.State == string(alexamodels.ToggleStateValueOn) {
		operationName = FeatureOperationNameTurnon
	}

	payload := alexamodels.TogglePayload{ToggleState: req.State}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameToggle,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlPercentageFeature controls percentage on an endpoint.
func (c *Client) ControlPercentageFeature(
	ctx context.Context,
	req *alexamodels.PercentageControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetPercentage && req.Percentage != nil:
		operationName = FeatureOperationNameSetpercentage
		payload = alexamodels.PercentageSetPayload{Percentage: *req.Percentage}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjustpercentage
		payload = alexamodels.PercentageAdjustPayload{PercentageDelta: *req.Delta}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either percentage or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNamePercentage,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlPowerLevelFeature controls power level on an endpoint.
func (c *Client) ControlPowerLevelFeature(
	ctx context.Context,
	req *alexamodels.PowerLevelControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetPowerLevel && req.PowerLevel != nil:
		operationName = FeatureOperationNameSetpercentage
		payload = alexamodels.PowerLevelSetPayload{PowerLevel: *req.PowerLevel}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjustpercentage
		payload = alexamodels.PowerLevelAdjustPayload{PowerLevelDelta: *req.Delta}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either powerLevel or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNamePowerlevel,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlActionFeature controls action on an endpoint.
func (c *Client) ControlActionFeature(
	ctx context.Context,
	req *alexamodels.ActionControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	payload := alexamodels.ActionPayload{
		Action:               req.Action,
		AdditionalProperties: req.Params,
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameAction,
		FeatureOperationName: FeatureOperationNamePerformaction,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlThermostatFeature controls thermostat target temperature on an endpoint.
func (c *Client) ControlThermostatFeature(
	ctx context.Context,
	req *alexamodels.ThermostatControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	var (
		operationName FeatureOperationName
		payload       interface{}
	)

	switch {
	case req.SetSetpoint && req.Value != nil:
		operationName = FeatureOperationNameSettargetsetpoint
		payload = alexamodels.ThermostatSetpointSetPayload{
			TargetSetpoint: alexamodels.ThermostatSetpoint{Value: *req.Value, Scale: req.Scale},
		}
	case req.Delta != nil:
		operationName = FeatureOperationNameAdjusttargetsetpoint
		payload = alexamodels.ThermostatSetpointAdjustPayload{
			TargetSetpointDelta: alexamodels.ThermostatSetpoint{Value: *req.Delta, Scale: req.Scale},
		}
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: "either value or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameThermostat,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

// ControlThermostatModeFeature controls thermostat mode on an endpoint.
func (c *Client) ControlThermostatModeFeature(
	ctx context.Context,
	req *alexamodels.ThermostatModeControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	payload := alexamodels.ThermostatModePayload{ThermostatMode: req.Mode}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameThermostat,
		FeatureOperationName: FeatureOperationNameSetthermostatmode,
		Payload:              payload,
	}

	return c.sendFeatureControlRequest(ctx, featureReq)
}

func (c *Client) sendFeatureControlRequest(
	ctx context.Context,
	featureReq FeatureControlRequest,
) (*alexaapimodels.FeatureControlResponse, error) {
	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}

	return ConvertToFeatureControlResponse(gqlResp), nil
}
