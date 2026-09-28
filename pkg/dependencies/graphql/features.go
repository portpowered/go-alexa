package graphql

import (
	"context"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// ControlPowerFeature controls power on an endpoint
func (c *Client) ControlPowerFeature(ctx context.Context, req *alexamodels.PowerControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	operationName := FeatureOperationNameTurnoff
	if req.State == "ON" {
		operationName = FeatureOperationNameTurnon
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.Endpoint.GetEndpointId(),
		FeatureName:          FeatureNamePower,
		FeatureOperationName: operationName,
		Payload:              nil,
	}

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlSpeakerFeature controls speaker volume on an endpoint
func (c *Client) ControlSpeakerFeature(ctx context.Context, req *alexamodels.SpeakerControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetVolume && req.Volume != nil {
		operationName = FeatureOperationNameSetvolume
		payload = map[string]interface{}{
			"volume": *req.Volume,
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjustvolume
		payload = map[string]interface{}{
			"volumeDelta": *req.Delta,
		}
	} else {
		return nil, &alexaapimodels.BadRequestError{
			Message: "either volume or delta must be specified",
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		FeatureName:          FeatureNameSpeaker,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlBrightnessFeature controls brightness on an endpoint
func (c *Client) ControlBrightnessFeature(ctx context.Context, req *alexamodels.BrightnessControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetBrightness && req.Brightness != nil {
		operationName = FeatureOperationNameSetbrightness
		payload = map[string]interface{}{
			"brightness": *req.Brightness,
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjustbrightness
		payload = map[string]interface{}{
			"brightnessDelta": *req.Delta,
		}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlColorFeature controls color on an endpoint
func (c *Client) ControlColorFeature(ctx context.Context, req *alexamodels.ColorControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	payload := map[string]interface{}{
		"color": map[string]interface{}{
			"hue":        req.Hue,
			"saturation": req.Saturation,
			"brightness": req.Brightness,
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlColorTemperatureFeature controls color temperature on an endpoint
func (c *Client) ControlColorTemperatureFeature(ctx context.Context, req *alexamodels.ColorTemperatureControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetTemperature && req.ColorTemperature != nil {
		operationName = FeatureOperationNameSetcolortemperature
		payload = map[string]interface{}{
			"colorTemperatureInKelvin": *req.ColorTemperature,
		}
	} else if req.Increase {
		operationName = FeatureOperationNameIncreasecolortemperature
		payload = map[string]interface{}{}
	} else if !req.Increase {
		operationName = FeatureOperationNameDecreasecolortemperature
		payload = map[string]interface{}{}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlLockFeature controls lock state on an endpoint
func (c *Client) ControlLockFeature(ctx context.Context, req *alexamodels.LockControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	operationName := FeatureOperationNameUnlock
	if req.State == "LOCKED" {
		operationName = FeatureOperationNameLock
	}

	payload := map[string]interface{}{
		"lockState": req.State,
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameLock,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlModeFeature controls mode on an endpoint
func (c *Client) ControlModeFeature(ctx context.Context, req *alexamodels.ModeControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetMode && req.Mode != nil {
		operationName = FeatureOperationNameSetmode
		payload = map[string]interface{}{
			"mode": *req.Mode,
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjustmode
		payload = map[string]interface{}{
			"modeDelta": *req.Delta,
		}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlRangeFeature controls range value on an endpoint
func (c *Client) ControlRangeFeature(ctx context.Context, req *alexamodels.RangeControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetValue && req.RangeValue != nil {
		operationName = FeatureOperationNameSetrangevalue
		payload = map[string]interface{}{
			"rangeValue": *req.RangeValue,
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjustrangevalue
		payload = map[string]interface{}{
			"rangeValueDelta": *req.Delta,
		}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlToggleFeature controls toggle state on an endpoint
func (c *Client) ControlToggleFeature(ctx context.Context, req *alexamodels.ToggleControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	operationName := FeatureOperationNameTurnoff
	if req.State == "ON" {
		operationName = FeatureOperationNameTurnon
	}

	payload := map[string]interface{}{
		"toggleState": req.State,
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameToggle,
		FeatureOperationName: operationName,
		Payload:              payload,
	}

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlPercentageFeature controls percentage on an endpoint
func (c *Client) ControlPercentageFeature(ctx context.Context, req *alexamodels.PercentageControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetPercentage && req.Percentage != nil {
		operationName = FeatureOperationNameSetpercentage
		payload = map[string]interface{}{
			"percentage": *req.Percentage,
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjustpercentage
		payload = map[string]interface{}{
			"percentageDelta": *req.Delta,
		}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlPowerLevelFeature controls power level on an endpoint
func (c *Client) ControlPowerLevelFeature(ctx context.Context, req *alexamodels.PowerLevelControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetPowerLevel && req.PowerLevel != nil {
		operationName = FeatureOperationNameSetpercentage
		payload = map[string]interface{}{
			"powerLevel": *req.PowerLevel,
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjustpercentage
		payload = map[string]interface{}{
			"powerLevelDelta": *req.Delta,
		}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlActionFeature controls action on an endpoint
func (c *Client) ControlActionFeature(ctx context.Context, req *alexamodels.ActionControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	payload := map[string]interface{}{
		"action": req.Action,
	}
	if req.Params != nil {
		for k, v := range req.Params {
			payload[k] = v
		}
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameAction,
		FeatureOperationName: FeatureOperationNamePerformaction,
		Payload:              payload,
	}

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlThermostatFeature controls thermostat target temperature on an endpoint
func (c *Client) ControlThermostatFeature(ctx context.Context, req *alexamodels.ThermostatControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	var operationName FeatureOperationName
	var payload map[string]interface{}

	if req.SetSetpoint && req.Value != nil {
		operationName = FeatureOperationNameSettargetsetpoint
		payload = map[string]interface{}{
			"targetSetpoint": map[string]interface{}{
				"value": *req.Value,
				"scale": req.Scale,
			},
		}
	} else if req.Delta != nil {
		operationName = FeatureOperationNameAdjusttargetsetpoint
		payload = map[string]interface{}{
			"targetSetpointDelta": map[string]interface{}{
				"value": *req.Delta,
				"scale": req.Scale,
			},
		}
	} else {
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

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}

// ControlThermostatModeFeature controls thermostat mode on an endpoint
func (c *Client) ControlThermostatModeFeature(ctx context.Context, req *alexamodels.ThermostatModeControlRequest) (*alexaapimodels.FeatureControlResponse, error) {
	payload := map[string]interface{}{
		"thermostatMode": req.Mode,
	}

	featureReq := FeatureControlRequest{
		EndpointId:           req.EndpointID,
		EntityId:             req.EntityID,
		Instance:             req.Instance,
		FeatureName:          FeatureNameThermostat,
		FeatureOperationName: FeatureOperationNameSetthermostatmode,
		Payload:              payload,
	}

	input := SetEndpointFeaturesInput{
		FeatureControlRequests: []FeatureControlRequest{featureReq},
	}

	gqlResp, err := c.SetEndpointFeatures(ctx, input)
	if err != nil {
		return nil, err
	}
	return ConvertToFeatureControlResponse(gqlResp), nil
}
