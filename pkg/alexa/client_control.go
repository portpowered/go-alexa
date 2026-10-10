package alexa

import (
	"context"
	"fmt"
	"html"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// Control provides a unified interface for all control operations.
// It dispatches to the appropriate control method based on the namespace and name in the request.
//
//nolint:cyclop // Each supported provider namespace has a distinct control contract.
func (c *Session) Control(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	if req.Target == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "target endpoint is required",
		}
	}

	// Keep provider-specific typed dispatch and the unsupported namespace fallback together.
	//nolint:exhaustive // The control facade intentionally handles only supported namespaces.
	switch req.Namespace {
	case alexaapimodels.FeatureNameNotification:
		return c.controlNotification(ctx, req)
	case alexaapimodels.FeatureNameAnnouncement:
		return c.controlAnnouncement(ctx, req)
	case alexaapimodels.FeatureNameSpeechSynthesizer:
		return c.controlTTS(ctx, req)
	case alexaapimodels.FeatureNameAudioPlayer:
		switch req.Name {
		case alexaapimodels.FeatureOperationNamePlayURI:
			return c.controlAudioPlayerURI(ctx, req)
		default:
			return c.controlMusic(ctx, req)
		}
	case alexaapimodels.FeatureNameNavigation:
		return c.controlNavigateHome(ctx, req)
	case alexaapimodels.FeatureNameBrightness:
		return c.controlBrightness(ctx, req)
	case alexaapimodels.FeatureNameColor:
		return c.controlColor(ctx, req)
	case alexaapimodels.FeatureNameColorTemperature:
		return c.controlColorTemperature(ctx, req)
	case alexaapimodels.FeatureNameLock:
		return c.controlLock(ctx, req)
	case alexaapimodels.FeatureNameMode:
		return c.controlMode(ctx, req)
	case alexaapimodels.FeatureNameRange:
		return c.controlRange(ctx, req)
	case alexaapimodels.FeatureNameToggle:
		return c.controlToggle(ctx, req)
	case alexaapimodels.FeatureNamePercentage:
		return c.controlPercentage(ctx, req)
	case alexaapimodels.FeatureNamePowerLevel:
		return c.controlPowerLevel(ctx, req)
	case alexaapimodels.FeatureNameAction:
		return c.controlAction(ctx, req)
	case alexaapimodels.FeatureNamePlayback:
		return c.controlPlaybackWithContext(ctx, req)
	case alexaapimodels.FeatureNamePower:
		return c.controlPower(ctx, req)
	case alexaapimodels.FeatureNameSpeaker:
		return c.controlVolume(ctx, req)
	case alexaapimodels.FeatureNameThermostat:
		return c.controlThermostat(ctx, req)
	default:
		return &alexaapimodels.ControlResponse{
				Errors: []alexaapimodels.FeatureControlError{
					{Message: fmt.Sprintf("unknown namespace: %s", req.Namespace)},
				},
			}, &alexaapimodels.BadRequestError{
				Message: fmt.Sprintf("unknown namespace: %s", req.Namespace),
			}
	}
}

// Helper methods to convert ControlRequest to specific request types

func (c *Session) controlNotification(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlNotificationPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSend,
	)
	if err != nil {
		return nil, err
	}

	notificationReq := &alexamodels.SendNotificationRequest{
		Endpoint:   req.Target,
		CustomerID: "",
		Message:    payload.Message,
		Title:      payload.Title,
	}

	err = c.restClient.SendNotification(ctx, notificationReq)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func parseAndValidate[T any](
	_ context.Context,
	req alexaapimodels.ControlRequest,
	expectedName alexaapimodels.FeatureOperationName,
) (*T, error) {
	if req.Name != expectedName {
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name: expected %s, got %s", expectedName, req.Name),
		}
	}

	// Type assert payload to the expected type
	payload, err := cast[T](req)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func setOrAdjustControl[TSet any, TAdjust any](
	ctx context.Context,
	req alexaapimodels.ControlRequest,
	setName alexaapimodels.FeatureOperationName,
	adjustName alexaapimodels.FeatureOperationName,
	featureName string,
	onSet func(TSet),
	onAdjust func(TAdjust),
) error {
	// The shared dispatcher accepts a caller-selected set/adjust pair.
	//nolint:exhaustive
	switch req.Name {
	case setName:
		payload, err := parseAndValidate[TSet](ctx, req, setName)
		if err != nil {
			return err
		}

		onSet(*payload)
	case adjustName:
		payload, err := parseAndValidate[TAdjust](ctx, req, adjustName)
		if err != nil {
			return err
		}

		onAdjust(*payload)
	default:
		return &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for %s: %s", featureName, req.Name),
		}
	}

	return nil
}

func (c *Session) controlAnnouncement(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlAnnouncementPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSend,
	)
	if err != nil {
		return nil, err
	}

	announcementReq := &alexamodels.SendAnnouncementRequest{
		Endpoint:      req.Target,
		CustomerID:    "",
		Locale:        payload.Locale,
		Message:       payload.Message,
		Method:        payload.Method,
		TargetDevices: nil,
		Title:         payload.Title,
	}

	err = c.restClient.SendAnnouncement(ctx, announcementReq)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Session) controlTTS(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlSpeechSynthesizerPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSpeak,
	)
	if err != nil {
		return nil, err
	}

	ttsReq := &alexamodels.SendTTSRequest{
		Endpoint:      req.Target,
		CustomerID:    "",
		Message:       payload.Message,
		TargetDevices: nil,
	}

	err = c.restClient.SendTTS(ctx, ttsReq)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Session) controlMusic(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlAudioPlayerPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNamePlay,
	)
	if err != nil {
		return nil, err
	}

	musicReq := &alexamodels.PlayMusicRequest{
		Endpoint:     req.Target,
		CustomerID:   "",
		ProviderID:   string(payload.ProviderID),
		SearchPhrase: payload.SearchPhrase,
		TimerSeconds: payload.TimerSeconds,
	}

	err = c.restClient.PlayMusic(ctx, musicReq)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Session) controlAudioPlayerURI(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlAudioPlayerURIPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNamePlayURI,
	)
	if err != nil {
		return nil, err
	}

	messageString := fmt.Sprintf(alexamodels.BehaviorAudioSSMLTemplate, html.EscapeString(payload.URI))

	uriReq := &alexamodels.SendTTSRequest{
		Endpoint:      req.Target,
		CustomerID:    "",
		Message:       messageString,
		TargetDevices: nil,
	}

	err = c.restClient.SendTTS(ctx, uriReq)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Session) controlNavigateHome(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	_, err := parseAndValidate[alexaapimodels.ControlNavigationPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameNavigateHome,
	)
	if err != nil {
		return nil, err
	}

	navigateReq := &alexamodels.FireTVRequest{
		Endpoint: req.Target,
	}

	err = c.restClient.FireTVNavigateHome(ctx, navigateReq)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Session) controlBrightness(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	brightnessReq := &alexamodels.BrightnessControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   "",
		},
		Brightness:    nil,
		Delta:         nil,
		SetBrightness: false,
	}

	err := setOrAdjustControl(
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetBrightness,
		alexaapimodels.FeatureOperationNameAdjustBrightness,
		string(alexaapimodels.FeatureNameBrightness),
		func(payload alexaapimodels.ControlBrightnessSetPayload) {
			brightnessReq.Brightness = &payload.Brightness
			brightnessReq.SetBrightness = true
		},
		func(payload alexaapimodels.ControlBrightnessAdjustPayload) {
			brightnessReq.Delta = &payload.Delta
			brightnessReq.SetBrightness = false
		},
	)
	if err != nil {
		return nil, err
	}

	return c.graphqlClient.ControlBrightnessFeature(ctx, brightnessReq)
}

func (c *Session) controlColor(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlColorPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetColor,
	)
	if err != nil {
		return nil, err
	}

	colorReq := &alexamodels.ColorControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   "",
		},
		Hue:        payload.Hue,
		Saturation: payload.Saturation,
		Brightness: payload.Brightness,
	}

	return c.graphqlClient.ControlColorFeature(ctx, colorReq)
}

//nolint:ireturn // The generic payload parser returns the caller-selected concrete payload type.
func cast[T any](req alexaapimodels.ControlRequest) (T, error) {
	var zero T
	// Try pointer type first
	if ptr, ok := req.Payload.(*T); ok {
		return *ptr, nil
	}
	// Try value type
	if val, ok := req.Payload.(T); ok {
		return val, nil
	}

	return zero, &alexaapimodels.BadRequestError{
		Message: fmt.Sprintf("invalid payload type for %s: expected %T", req.Name, zero),
	}
}

func (c *Session) controlColorTemperature(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	colorTempReq := &alexamodels.ColorTemperatureControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   "",
		},
		ColorTemperature: nil,
		Increase:         false,
		SetTemperature:   false,
	}

	//nolint:exhaustive // This handler intentionally supports only selected operations.
	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetColorTemperature:
		payload, err := parseAndValidate[alexaapimodels.ControlColorTemperatureSetPayload](
			ctx,
			req,
			alexaapimodels.FeatureOperationNameSetColorTemperature,
		)
		if err != nil {
			return nil, err
		}

		colorTempReq.ColorTemperature = &payload.ColorTemperature
		colorTempReq.SetTemperature = true
	case alexaapimodels.FeatureOperationNameIncreasColorTemperature:
		payload, err := parseAndValidate[alexaapimodels.ControlColorTemperatureAdjustPayload](
			ctx,
			req,
			alexaapimodels.FeatureOperationNameIncreasColorTemperature,
		)
		if err != nil {
			return nil, err
		}

		colorTempReq.Increase = payload.Increase
		colorTempReq.SetTemperature = false
	case alexaapimodels.FeatureOperationNameDecreaseColorTemperature:
		payload, err := parseAndValidate[alexaapimodels.ControlColorTemperatureAdjustPayload](
			ctx,
			req,
			alexaapimodels.FeatureOperationNameDecreaseColorTemperature,
		)
		if err != nil {
			return nil, err
		}

		colorTempReq.Increase = payload.Increase
		colorTempReq.SetTemperature = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for colorTemperature: %s", req.Name),
		}
	}

	return c.graphqlClient.ControlColorTemperatureFeature(ctx, colorTempReq)
}

func (c *Session) controlLock(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlLockPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetLockState,
	)
	if err != nil {
		return nil, err
	}

	lockReq := &alexamodels.LockControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   "",
		},
		State: payload.State,
	}

	return c.graphqlClient.ControlLockFeature(ctx, lockReq)
}

func (c *Session) controlMode(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	modeReq := &alexamodels.ModeControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   req.Instance,
		},
		Delta:   nil,
		Mode:    nil,
		SetMode: false,
	}

	err := setOrAdjustControl(
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetMode,
		alexaapimodels.FeatureOperationNameAdjustMode,
		string(alexaapimodels.FeatureNameMode),
		func(payload alexaapimodels.ControlModeSetPayload) {
			modeReq.Mode = &payload.Mode
			modeReq.SetMode = true
		},
		func(payload alexaapimodels.ControlModeAdjustPayload) {
			modeReq.Delta = &payload.Delta
			modeReq.SetMode = false
		},
	)
	if err != nil {
		return nil, err
	}

	return c.graphqlClient.ControlModeFeature(ctx, modeReq)
}

func (c *Session) controlRange(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	rangeReq := &alexamodels.RangeControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   req.Instance,
		},
		Delta:      nil,
		RangeValue: nil,
		SetValue:   false,
	}

	err := setOrAdjustControl(
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetRangeValue,
		alexaapimodels.FeatureOperationNameAdjustRangeValue,
		string(alexaapimodels.FeatureNameRange),
		func(payload alexaapimodels.ControlRangeSetPayload) {
			rangeReq.RangeValue = &payload.RangeValue
			rangeReq.SetValue = true
		},
		func(payload alexaapimodels.ControlRangeAdjustPayload) {
			rangeReq.Delta = &payload.Delta
			rangeReq.SetValue = false
		},
	)
	if err != nil {
		return nil, err
	}

	return c.graphqlClient.ControlRangeFeature(ctx, rangeReq)
}

func (c *Session) controlToggle(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlTogglePayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetToggleState,
	)
	if err != nil {
		return nil, err
	}

	toggleReq := &alexamodels.ToggleControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   req.Instance,
		},
		State: payload.State,
	}

	return c.graphqlClient.ControlToggleFeature(ctx, toggleReq)
}

func (c *Session) controlPercentage(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	percentageReq := &alexamodels.PercentageControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   req.Instance,
		},
		Delta:         nil,
		Percentage:    nil,
		SetPercentage: false,
	}

	err := setOrAdjustControl(
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetPercentage,
		alexaapimodels.FeatureOperationNameAdjustPercentage,
		string(alexaapimodels.FeatureNamePercentage),
		func(payload alexaapimodels.ControlPercentageSetPayload) {
			percentageReq.Percentage = &payload.Percentage
			percentageReq.SetPercentage = true
		},
		func(payload alexaapimodels.ControlPercentageAdjustPayload) {
			percentageReq.Delta = &payload.Delta
			percentageReq.SetPercentage = false
		},
	)
	if err != nil {
		return nil, err
	}

	return c.graphqlClient.ControlPercentageFeature(ctx, percentageReq)
}

func (c *Session) controlPowerLevel(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	powerLevelReq := &alexamodels.PowerLevelControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   "",
		},
		Delta:         nil,
		PowerLevel:    nil,
		SetPowerLevel: false,
	}

	err := setOrAdjustControl(
		ctx,
		req,
		alexaapimodels.FeatureOperationNameSetPowerLevel,
		alexaapimodels.FeatureOperationNameAdjustPowerLevel,
		string(alexaapimodels.FeatureNamePowerLevel),
		func(payload alexaapimodels.ControlPowerLevelSetPayload) {
			powerLevelReq.PowerLevel = &payload.PowerLevel
			powerLevelReq.SetPowerLevel = true
		},
		func(payload alexaapimodels.ControlPowerLevelAdjustPayload) {
			powerLevelReq.Delta = &payload.Delta
			powerLevelReq.SetPowerLevel = false
		},
	)
	if err != nil {
		return nil, err
	}

	return c.graphqlClient.ControlPowerLevelFeature(ctx, powerLevelReq)
}

func (c *Session) controlAction(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlActionPayload](
		ctx,
		req,
		alexaapimodels.FeatureOperationNamePerformAction,
	)
	if err != nil {
		return nil, err
	}

	actionReq := &alexamodels.ActionControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			EntityID:   "",
			Instance:   "",
		},
		Action: payload.Action,
		Params: payload.Params,
	}

	return c.graphqlClient.ControlActionFeature(ctx, actionReq)
}

func (c *Session) controlPlayback(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	validOps := map[alexaapimodels.FeatureOperationName]bool{
		alexaapimodels.FeatureOperationNamePlay:     true,
		alexaapimodels.FeatureOperationNamePause:    true,
		alexaapimodels.FeatureOperationNameResume:   true,
		alexaapimodels.FeatureOperationNameNext:     true,
		alexaapimodels.FeatureOperationNamePrevious: true,
		alexaapimodels.FeatureOperationNameStop:     true,
		alexaapimodels.FeatureOperationNameForward:  true,
		alexaapimodels.FeatureOperationNameRewind:   true,
		alexaapimodels.FeatureOperationNameShuffle:  true,
		alexaapimodels.FeatureOperationNameRepeat:   true,
	}

	if !validOps[req.Name] {
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for playback: %s", req.Name),
		}
	}

	deviceAccountID := req.Target.GetDeviceAccountId()
	if deviceAccountID != "" && req.Target.GetDeviceFamily() == alexamodels.DeviceFamilyFireTV {
		return c.controlFireTVPlayback(ctx, req)
	}

	if req.Target.GetDeviceSerialNumber() != "" {
		return c.controlEchoPlayback(ctx, req)
	}

	return c.controlThirdPartyPlayback(ctx, req)
}

func (c *Session) controlPower(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	// Type assert payload to ControlPowerPayload (can be empty)
	if req.Payload != nil {
		_, err := cast[alexaapimodels.ControlPowerPayload](req)
		if err != nil {
			return nil, err
		}
	}

	var state alexaapimodels.PowerState

	//nolint:exhaustive // This handler intentionally supports only selected operations.
	switch req.Name {
	case alexaapimodels.FeatureOperationNameTurnOn:
		state = alexaapimodels.PowerStateOn
	case alexaapimodels.FeatureOperationNameTurnOff:
		state = alexaapimodels.PowerStateOff
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for power: %s", req.Name),
		}
	}

	deviceAccountID := req.Target.GetDeviceAccountId()
	isFireTV := deviceAccountID != "" && req.Target.GetDeviceFamily() == alexamodels.DeviceFamilyFireTV

	if isFireTV {
		// Route to FireTV power operations
		var err error

		switch state {
		case alexaapimodels.PowerStateOn:
			err = c.restClient.FireTVTurnOn(ctx, &alexamodels.FireTVRequest{
				Endpoint: req.Target,
			})
		case alexaapimodels.PowerStateOff:
			err = c.restClient.FireTVTurnOff(ctx, &alexamodels.FireTVRequest{
				Endpoint: req.Target,
			})
		default:
			return nil, &alexaapimodels.BadRequestError{
				Message: "invalid power state",
			}
		}

		if err != nil {
			return nil, err
		}

		return &alexaapimodels.ControlResponse{}, nil
	}

	// Route to GraphQL power feature control
	endpointID := req.Target.GetEndpointId()
	if endpointID == "" {
		return nil, &alexaapimodels.BadRequestError{
			Message: "endpointId is required for power control",
		}
	}

	powerControlReq := &alexamodels.PowerControlRequest{
		Endpoint: req.Target,
		State:    state,
	}

	return c.graphqlClient.ControlPowerFeature(ctx, powerControlReq)
}

func (c *Session) controlVolume(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	volumeReq, err := volumeRequestFromControl(ctx, req)
	if err != nil {
		return nil, err
	}

	if usesGraphQLVolumeControl(req.Target) {
		return c.controlGraphQLVolume(ctx, req.Target, volumeReq)
	}

	if req.Target.GetDeviceSerialNumber() != "" {
		return c.controlRESTVolume(ctx, req.Target, volumeReq)
	}

	return c.controlInterfaceVolume(ctx, req.Target, req.Name, volumeReq)
}

func (c *Session) controlThermostat(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	// Check if this is a mode operation or temperature operation
	//nolint:exhaustive // This handler intentionally supports only selected operations.
	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetThermostatMode:
		// Handle thermostat mode
		payload, err := parseAndValidate[alexaapimodels.ControlThermostatModePayload](
			ctx,
			req,
			alexaapimodels.FeatureOperationNameSetThermostatMode,
		)
		if err != nil {
			return nil, err
		}

		modeReq := &alexamodels.ThermostatModeControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: req.Target.GetEndpointId(),
				EntityID:   "",
				Instance:   "",
			},
			Mode: payload.Mode,
		}

		return c.graphqlClient.ControlThermostatModeFeature(ctx, modeReq)
	case alexaapimodels.FeatureOperationNameSetTargetSetpoint:
		// Handle setting target temperature
		payload, err := parseAndValidate[alexaapimodels.ControlThermostatSetpointPayload](
			ctx,
			req,
			alexaapimodels.FeatureOperationNameSetTargetSetpoint,
		)
		if err != nil {
			return nil, err
		}

		thermostatReq := &alexamodels.ThermostatControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: req.Target.GetEndpointId(),
				EntityID:   "",
				Instance:   "",
			},
			Delta:       nil,
			Scale:       payload.Scale,
			SetSetpoint: true,
			Value:       &payload.Value,
		}

		return c.graphqlClient.ControlThermostatFeature(ctx, thermostatReq)
	case alexaapimodels.FeatureOperationNameAdjustTargetSetpoint:
		// Handle adjusting target temperature
		payload, err := parseAndValidate[alexaapimodels.ControlThermostatSetpointAdjustPayload](
			ctx,
			req,
			alexaapimodels.FeatureOperationNameAdjustTargetSetpoint,
		)
		if err != nil {
			return nil, err
		}

		thermostatReq := &alexamodels.ThermostatControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: req.Target.GetEndpointId(),
				EntityID:   "",
				Instance:   "",
			},
			Delta:       &payload.Delta,
			Scale:       payload.Scale,
			SetSetpoint: false,
			Value:       nil,
		}

		return c.graphqlClient.ControlThermostatFeature(ctx, thermostatReq)
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for thermostat: %s", req.Name),
		}
	}
}

func (c *Session) controlFireTVPlayback(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	var err error

	//nolint:exhaustive // FireTV supports only pause and resume in this facade.
	switch req.Name {
	case alexaapimodels.FeatureOperationNamePause:
		err = c.restClient.FireTVPauseVideo(ctx, &alexamodels.FireTVRequest{Endpoint: req.Target})
	case alexaapimodels.FeatureOperationNameResume:
		err = c.restClient.FireTVResumeVideo(ctx, &alexamodels.FireTVRequest{Endpoint: req.Target})
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("FireTV playback operation %s not supported", req.Name),
		}
	}

	return playbackControlResponse(err)
}

func (c *Session) controlEchoPlayback(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	mediaRequest := &alexamodels.MediaControlRequest{Endpoint: req.Target}

	var err error

	//nolint:exhaustive // This handler intentionally supports selected operations.
	switch req.Name {
	case alexaapimodels.FeatureOperationNamePlay, alexaapimodels.FeatureOperationNameResume:
		err = c.restClient.ResumePlayback(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNamePause:
		err = c.restClient.PausePlayback(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNameNext:
		err = c.restClient.NextTrack(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNamePrevious:
		err = c.restClient.PreviousTrack(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNameStop:
		err = c.restClient.StopPlayback(ctx, &alexamodels.StopPlaybackRequest{
			Endpoint:   req.Target,
			AllDevices: false,
			CustomerID: "",
		})
	case alexaapimodels.FeatureOperationNameForward:
		err = c.restClient.ForwardMedia(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNameRewind:
		err = c.restClient.RewindMedia(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNameShuffle:
		payload, parseErr := parseAndValidate[alexaapimodels.ControlShufflePayload](ctx, req, alexaapimodels.FeatureOperationNameShuffle)
		if parseErr != nil {
			return nil, parseErr
		}

		err = c.restClient.SetShuffle(ctx, mediaRequest, payload.Shuffle)
	case alexaapimodels.FeatureOperationNameRepeat:
		payload, parseErr := parseAndValidate[alexaapimodels.ControlRepeatPayload](ctx, req, alexaapimodels.FeatureOperationNameRepeat)
		if parseErr != nil {
			return nil, parseErr
		}

		err = c.restClient.SetRepeat(ctx, mediaRequest, payload.Repeat)
	default:
		return nil, unsupportedPlaybackOperation("media", req.Name)
	}

	return playbackControlResponse(err)
}

func (c *Session) controlThirdPartyPlayback(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	mediaRequest := &alexamodels.MediaControlRequest{Endpoint: req.Target}

	var err error

	//nolint:exhaustive // This handler intentionally supports selected operations.
	switch req.Name {
	case alexaapimodels.FeatureOperationNamePause,
		alexaapimodels.FeatureOperationNamePlay,
		alexaapimodels.FeatureOperationNameResume,
		alexaapimodels.FeatureOperationNameNext,
		alexaapimodels.FeatureOperationNamePrevious,
		alexaapimodels.FeatureOperationNameStop:
		err = c.sendThirdPartyPlaybackMessage(ctx, req)
	case alexaapimodels.FeatureOperationNameForward:
		err = c.restClient.ForwardMedia(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNameRewind:
		err = c.restClient.RewindMedia(ctx, mediaRequest)
	case alexaapimodels.FeatureOperationNameShuffle:
		payload, parseErr := parseAndValidate[alexaapimodels.ControlShufflePayload](ctx, req, alexaapimodels.FeatureOperationNameShuffle)
		if parseErr != nil {
			return nil, parseErr
		}

		err = c.restClient.SetShuffle(ctx, mediaRequest, payload.Shuffle)
	case alexaapimodels.FeatureOperationNameRepeat:
		payload, parseErr := parseAndValidate[alexaapimodels.ControlRepeatPayload](ctx, req, alexaapimodels.FeatureOperationNameRepeat)
		if parseErr != nil {
			return nil, parseErr
		}

		err = c.restClient.SetRepeat(ctx, mediaRequest, payload.Repeat)
	default:
		return nil, unsupportedPlaybackOperation("media", req.Name)
	}

	return playbackControlResponse(err)
}

func (c *Session) sendThirdPartyPlaybackMessage(ctx context.Context, req alexaapimodels.ControlRequest) error {
	operation := req.Name
	if operation == alexaapimodels.FeatureOperationNameResume {
		operation = alexaapimodels.FeatureOperationNamePlay
	}

	return c.restClient.SendInterfaceMessage(ctx, &alexamodels.InterfaceMessageRequest{
		Endpoint:      req.Target,
		FeatureName:   string(alexaapimodels.FeatureNamePlayback),
		OperationName: string(operation),
		Payload:       alexamodels.FeatureEmptyPayload{},
	})
}

func playbackControlResponse(err error) (*alexaapimodels.ControlResponse, error) {
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func unsupportedPlaybackOperation(device string, name alexaapimodels.FeatureOperationName) error {
	return &alexaapimodels.BadRequestError{
		Message: fmt.Sprintf("%s playback operation %s not supported", device, name),
	}
}

func volumeRequestFromControl(
	ctx context.Context,
	req alexaapimodels.ControlRequest,
) (*alexamodels.VolumeControlRequest, error) {
	volumeReq := &alexamodels.VolumeControlRequest{
		Endpoint:   req.Target,
		CustomerID: "",
		Delta:      nil,
		SetVolume:  false,
		Volume:     nil,
	}

	//nolint:exhaustive // The volume handler supports only set and adjust operations.
	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetVolume:
		payload, err := parseAndValidate[alexaapimodels.ControlVolumeSetPayload](ctx, req, req.Name)
		if err != nil {
			return nil, err
		}

		volumeReq.Volume = &payload.Volume
		volumeReq.SetVolume = true
	case alexaapimodels.FeatureOperationNameAdjustVolume:
		payload, err := parseAndValidate[alexaapimodels.ControlVolumeAdjustPayload](ctx, req, req.Name)
		if err != nil {
			return nil, err
		}

		volumeReq.Delta = &payload.Delta
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for volume: %s", req.Name),
		}
	}

	return volumeReq, nil
}

func usesGraphQLVolumeControl(endpoint alexaapimodels.EndpointInterface) bool {
	return endpoint.GetDeviceSerialNumber() != "" &&
		endpoint.GetDeviceFamily() != string(alexamodels.WholeHomeDeviceFamily)
}

func (c *Session) controlGraphQLVolume(
	ctx context.Context,
	endpoint alexaapimodels.EndpointInterface,
	volumeReq *alexamodels.VolumeControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	endpointID := endpoint.GetEndpointId()
	if endpointID == "" {
		return nil, &alexaapimodels.BadRequestError{
			Message: "endpointId is required for GraphQL volume control",
		}
	}

	speakerReq := &alexamodels.SpeakerControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: endpointID,
			EntityID:   "",
			Instance:   "",
		},
		Volume:    volumeReq.Volume,
		Delta:     volumeReq.Delta,
		SetVolume: volumeReq.SetVolume,
	}

	return c.graphqlClient.ControlSpeakerFeature(ctx, speakerReq)
}

func (c *Session) controlRESTVolume(
	ctx context.Context,
	endpoint alexaapimodels.EndpointInterface,
	volumeReq *alexamodels.VolumeControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	if volumeReq.Volume == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "volume is required for REST volume control",
		}
	}

	restRequest := &alexamodels.VolumeControlRequest{
		Endpoint:   endpoint,
		CustomerID: volumeReq.CustomerID,
		Delta:      nil,
		SetVolume:  false,
		Volume:     volumeReq.Volume,
	}

	err := c.restClient.SetVolume(ctx, restRequest)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Session) controlInterfaceVolume(
	ctx context.Context,
	endpoint alexaapimodels.EndpointInterface,
	name alexaapimodels.FeatureOperationName,
	volumeReq *alexamodels.VolumeControlRequest,
) (*alexaapimodels.ControlResponse, error) {
	operation, value, err := volumeInterfaceOperation(name, volumeReq)
	if err != nil {
		return nil, err
	}

	request := &alexamodels.InterfaceMessageRequest{
		Endpoint:      endpoint,
		FeatureName:   string(alexaapimodels.FeatureNameSpeaker),
		OperationName: string(operation),
		Payload:       alexamodels.SpeakerSetVolumePayload{Volume: value},
	}

	err = c.restClient.SendInterfaceMessage(ctx, request)
	if err != nil {
		return nil, err
	}

	return &alexaapimodels.ControlResponse{}, nil
}

func volumeInterfaceOperation(
	name alexaapimodels.FeatureOperationName,
	volumeReq *alexamodels.VolumeControlRequest,
) (alexaapimodels.FeatureOperationName, int, error) {
	//nolint:exhaustive // The volume handler supports only set and adjust operations.
	switch name {
	case alexaapimodels.FeatureOperationNameSetVolume:
		return alexaapimodels.FeatureOperationNameSetVolume, *volumeReq.Volume, nil
	case alexaapimodels.FeatureOperationNameAdjustVolume:
		return alexaapimodels.FeatureOperationNameAdjustVolume, *volumeReq.Delta, nil
	default:
		return "", 0, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for volume: %s", name),
		}
	}
}
