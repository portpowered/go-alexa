package alexa

import (
	"context"
	"fmt"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// Control provides a unified interface for all control operations.
// It dispatches to the appropriate control method based on the namespace and name in the request.
func (c *Client) Control(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {

	if req.Target == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "target endpoint is required",
		}
	}

	// Dispatch based on namespace and name
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
		return c.controlPlayback(ctx, req)
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

func (c *Client) controlNotification(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlNotificationPayload](ctx, req, alexaapimodels.FeatureOperationNameSend)
	if err != nil {
		return nil, err
	}
	notificationReq := &alexamodels.SendNotificationRequest{
		Endpoint: req.Target,
		Message:  payload.Message,
		Title:    payload.Title,
	}
	err = c.restClient.SendNotification(ctx, notificationReq)
	if err != nil {
		return nil, err
	}
	return &alexaapimodels.ControlResponse{}, nil
}

func parseAndValidate[T any](ctx context.Context, req alexaapimodels.ControlRequest, expectedName alexaapimodels.FeatureOperationName) (*T, error) {
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
func (c *Client) controlAnnouncement(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlAnnouncementPayload](ctx, req, alexaapimodels.FeatureOperationNameSend)
	if err != nil {
		return nil, err
	}

	announcementReq := &alexamodels.SendAnnouncementRequest{
		Endpoint: req.Target,
		Message:  payload.Message,
		Method:   payload.Method,
		Title:    payload.Title,
		Locale:   payload.Locale,
	}

	err = c.restClient.SendAnnouncement(ctx, announcementReq)
	if err != nil {
		return nil, err
	}
	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Client) controlTTS(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlSpeechSynthesizerPayload](ctx, req, alexaapimodels.FeatureOperationNameSpeak)
	if err != nil {
		return nil, err
	}

	ttsReq := &alexamodels.SendTTSRequest{
		Endpoint: req.Target,
		Message:  payload.Message,
	}

	err = c.restClient.SendTTS(ctx, ttsReq)
	if err != nil {
		return nil, err
	}
	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Client) controlMusic(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlAudioPlayerPayload](ctx, req, alexaapimodels.FeatureOperationNamePlay)
	if err != nil {
		return nil, err
	}

	musicReq := &alexamodels.PlayMusicRequest{
		Endpoint:     req.Target,
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

func (c *Client) controlAudioPlayerURI(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlAudioPlayerURIPayload](ctx, req, alexaapimodels.FeatureOperationNamePlayURI)
	if err != nil {
		return nil, err
	}

	messageString := fmt.Sprintf("<audio src='%s'/>", payload.URI)

	uriReq := &alexamodels.SendTTSRequest{
		Endpoint: req.Target,
		Message:  messageString,
	}

	err = c.restClient.SendTTS(ctx, uriReq)
	if err != nil {
		return nil, err
	}
	return &alexaapimodels.ControlResponse{}, nil
}

func (c *Client) controlNavigateHome(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	_, err := parseAndValidate[alexaapimodels.ControlNavigationPayload](ctx, req, alexaapimodels.FeatureOperationNameNavigateHome)
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

func (c *Client) controlBrightness(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	brightnessReq := &alexamodels.BrightnessControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
		},
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetBrightness:
		payload, err := parseAndValidate[alexaapimodels.ControlBrightnessSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetBrightness)
		if err != nil {
			return nil, err
		}
		brightnessReq.Brightness = &payload.Brightness
		brightnessReq.SetBrightness = true
	case alexaapimodels.FeatureOperationNameAdjustBrightness:
		payload, err := parseAndValidate[alexaapimodels.ControlBrightnessAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustBrightness)
		if err != nil {
			return nil, err
		}
		brightnessReq.Delta = &payload.Delta
		brightnessReq.SetBrightness = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for brightness: %s", req.Name),
		}
	}

	return c.graphqlClient.ControlBrightnessFeature(ctx, brightnessReq)
}

func (c *Client) controlColor(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlColorPayload](ctx, req, alexaapimodels.FeatureOperationNameSetColor)
	if err != nil {
		return nil, err
	}

	colorReq := &alexamodels.ColorControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
		},
		Hue:        payload.Hue,
		Saturation: payload.Saturation,
		Brightness: payload.Brightness,
	}

	return c.graphqlClient.ControlColorFeature(ctx, colorReq)
}

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

func (c *Client) controlColorTemperature(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	colorTempReq := &alexamodels.ColorTemperatureControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
		},
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetColorTemperature:
		payload, err := parseAndValidate[alexaapimodels.ControlColorTemperatureSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetColorTemperature)
		if err != nil {
			return nil, err
		}
		colorTempReq.ColorTemperature = &payload.ColorTemperature
		colorTempReq.SetTemperature = true
	case alexaapimodels.FeatureOperationNameIncreasColorTemperature:
		payload, err := parseAndValidate[alexaapimodels.ControlColorTemperatureAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameIncreasColorTemperature)
		if err != nil {
			return nil, err
		}
		colorTempReq.Increase = payload.Increase
		colorTempReq.SetTemperature = false
	case alexaapimodels.FeatureOperationNameDecreaseColorTemperature:
		payload, err := parseAndValidate[alexaapimodels.ControlColorTemperatureAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameDecreaseColorTemperature)
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

func (c *Client) controlLock(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlLockPayload](ctx, req, alexaapimodels.FeatureOperationNameSetLockState)
	if err != nil {
		return nil, err
	}

	lockReq := &alexamodels.LockControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
		},
		State: payload.State,
	}

	return c.graphqlClient.ControlLockFeature(ctx, lockReq)
}

func (c *Client) controlMode(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	modeReq := &alexamodels.ModeControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			Instance:   req.Instance,
		},
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetMode:
		payload, err := parseAndValidate[alexaapimodels.ControlModeSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetMode)
		if err != nil {
			return nil, err
		}
		modeReq.Mode = &payload.Mode
		modeReq.SetMode = true
	case alexaapimodels.FeatureOperationNameAdjustMode:
		payload, err := parseAndValidate[alexaapimodels.ControlModeAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustMode)
		if err != nil {
			return nil, err
		}
		modeReq.Delta = &payload.Delta
		modeReq.SetMode = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for mode: %s", req.Name),
		}
	}

	return c.graphqlClient.ControlModeFeature(ctx, modeReq)
}

func (c *Client) controlRange(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	rangeReq := &alexamodels.RangeControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			Instance:   req.Instance,
		},
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetRangeValue:
		payload, err := parseAndValidate[alexaapimodels.ControlRangeSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetRangeValue)
		if err != nil {
			return nil, err
		}
		rangeReq.RangeValue = &payload.RangeValue
		rangeReq.SetValue = true
	case alexaapimodels.FeatureOperationNameAdjustRangeValue:
		payload, err := parseAndValidate[alexaapimodels.ControlRangeAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustRangeValue)
		if err != nil {
			return nil, err
		}
		rangeReq.Delta = &payload.Delta
		rangeReq.SetValue = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for range: %s", req.Name),
		}
	}

	return c.graphqlClient.ControlRangeFeature(ctx, rangeReq)
}

func (c *Client) controlToggle(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlTogglePayload](ctx, req, alexaapimodels.FeatureOperationNameSetToggleState)
	if err != nil {
		return nil, err
	}

	toggleReq := &alexamodels.ToggleControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			Instance:   req.Instance,
		},
		State: payload.State,
	}

	return c.graphqlClient.ControlToggleFeature(ctx, toggleReq)
}

func (c *Client) controlPercentage(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	percentageReq := &alexamodels.PercentageControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
			Instance:   req.Instance,
		},
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetPercentage:
		payload, err := parseAndValidate[alexaapimodels.ControlPercentageSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetPercentage)
		if err != nil {
			return nil, err
		}
		percentageReq.Percentage = &payload.Percentage
		percentageReq.SetPercentage = true
	case alexaapimodels.FeatureOperationNameAdjustPercentage:
		payload, err := parseAndValidate[alexaapimodels.ControlPercentageAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustPercentage)
		if err != nil {
			return nil, err
		}
		percentageReq.Delta = &payload.Delta
		percentageReq.SetPercentage = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for percentage: %s", req.Name),
		}
	}
	return c.graphqlClient.ControlPercentageFeature(ctx, percentageReq)
}

func (c *Client) controlPowerLevel(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	powerLevelReq := &alexamodels.PowerLevelControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
		},
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetPowerLevel:
		payload, err := parseAndValidate[alexaapimodels.ControlPowerLevelSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetPowerLevel)
		if err != nil {
			return nil, err
		}
		powerLevelReq.PowerLevel = &payload.PowerLevel
		powerLevelReq.SetPowerLevel = true
	case alexaapimodels.FeatureOperationNameAdjustPowerLevel:
		payload, err := parseAndValidate[alexaapimodels.ControlPowerLevelAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustPowerLevel)
		if err != nil {
			return nil, err
		}
		powerLevelReq.Delta = &payload.Delta
		powerLevelReq.SetPowerLevel = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for powerLevel: %s", req.Name),
		}
	}

	return c.graphqlClient.ControlPowerLevelFeature(ctx, powerLevelReq)
}

func (c *Client) controlAction(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	payload, err := parseAndValidate[alexaapimodels.ControlActionPayload](ctx, req, alexaapimodels.FeatureOperationNamePerformAction)
	if err != nil {
		return nil, err
	}

	actionReq := &alexamodels.ActionControlRequest{
		BaseFeatureRequest: alexamodels.BaseFeatureRequest{
			EndpointID: req.Target.GetEndpointId(),
		},
		Action: payload.Action,
		Params: payload.Params,
	}

	return c.graphqlClient.ControlActionFeature(ctx, actionReq)
}

func (c *Client) controlPlayback(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
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

	// Check if this is a FireTV device
	deviceAccountID := req.Target.GetDeviceAccountId()
	isFireTV := deviceAccountID != "" && req.Target.GetDeviceFamily() == alexamodels.DeviceFamilyFireTV
	isEcho := req.Target.GetDeviceSerialNumber() != ""
	if isFireTV {
		// Route to FireTV operations
		var err error
		switch req.Name {
		case alexaapimodels.FeatureOperationNamePause:
			err = c.restClient.FireTVPauseVideo(ctx, &alexamodels.FireTVRequest{
				Endpoint: req.Target,
			})
		case alexaapimodels.FeatureOperationNameResume:
			err = c.restClient.FireTVResumeVideo(ctx, &alexamodels.FireTVRequest{
				Endpoint: req.Target,
			})
		default:
			return nil, &alexaapimodels.BadRequestError{
				Message: fmt.Sprintf("FireTV playback operation %s not supported", req.Name),
			}
		}
		if err != nil {
			return nil, err
		}
		return &alexaapimodels.ControlResponse{}, nil
	} else if isEcho {
		// Route to regular media operations
		var err error
		switch req.Name {
		case alexaapimodels.FeatureOperationNamePlay:
			err = c.restClient.ResumePlayback(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNamePause:
			err = c.restClient.PausePlayback(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameResume:
			err = c.restClient.ResumePlayback(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameNext:
			err = c.restClient.NextTrack(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNamePrevious:
			err = c.restClient.PreviousTrack(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameStop:
			err = c.restClient.StopPlayback(ctx, &alexamodels.StopPlaybackRequest{
				Endpoint: req.Target,
			})
		case alexaapimodels.FeatureOperationNameForward:
			err = c.restClient.ForwardMedia(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameRewind:
			err = c.restClient.RewindMedia(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameShuffle:
			payload, parseErr := parseAndValidate[alexaapimodels.ControlShufflePayload](ctx, req, alexaapimodels.FeatureOperationNameShuffle)
			if parseErr != nil {
				return nil, parseErr
			}
			err = c.restClient.SetShuffle(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target}, payload.Shuffle)
		case alexaapimodels.FeatureOperationNameRepeat:
			payload, parseErr := parseAndValidate[alexaapimodels.ControlRepeatPayload](ctx, req, alexaapimodels.FeatureOperationNameRepeat)
			if parseErr != nil {
				return nil, parseErr
			}
			err = c.restClient.SetRepeat(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target}, payload.Repeat)
		default:
			return nil, &alexaapimodels.BadRequestError{
				Message: fmt.Sprintf("media playback operation %s not supported", req.Name),
			}
		}
		if err != nil {
			return nil, err
		}
		return &alexaapimodels.ControlResponse{}, nil
	} else {
		// likely means, a third party device like roku.
		// Route to regular media operations
		var err error
		switch req.Name {
		case alexaapimodels.FeatureOperationNamePause:
			err = c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNamePlayback),
					OperationName: string(alexaapimodels.FeatureOperationNamePause), Payload: map[string]interface{}{}})
		case alexaapimodels.FeatureOperationNamePlay:
			err = c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNamePlayback),
					OperationName: string(alexaapimodels.FeatureOperationNamePlay), Payload: map[string]interface{}{}})
		case alexaapimodels.FeatureOperationNameResume:
			err = c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNamePlayback),
					OperationName: string(alexaapimodels.FeatureOperationNamePlay), Payload: map[string]interface{}{}})
		case alexaapimodels.FeatureOperationNameNext:
			err = c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNamePlayback),
					OperationName: string(alexaapimodels.FeatureOperationNameNext), Payload: map[string]interface{}{}})
		case alexaapimodels.FeatureOperationNamePrevious:
			err = c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNamePlayback),
					OperationName: string(alexaapimodels.FeatureOperationNamePrevious), Payload: map[string]interface{}{}})
		case alexaapimodels.FeatureOperationNameStop:
			err = c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNamePlayback),
					OperationName: string(alexaapimodels.FeatureOperationNameStop), Payload: map[string]interface{}{}})
		case alexaapimodels.FeatureOperationNameForward:
			err = c.restClient.ForwardMedia(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameRewind:
			err = c.restClient.RewindMedia(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target})
		case alexaapimodels.FeatureOperationNameShuffle:
			payload, parseErr := parseAndValidate[alexaapimodels.ControlShufflePayload](ctx, req, alexaapimodels.FeatureOperationNameShuffle)
			if parseErr != nil {
				return nil, parseErr
			}
			err = c.restClient.SetShuffle(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target}, payload.Shuffle)
		case alexaapimodels.FeatureOperationNameRepeat:
			payload, parseErr := parseAndValidate[alexaapimodels.ControlRepeatPayload](ctx, req, alexaapimodels.FeatureOperationNameRepeat)
			if parseErr != nil {
				return nil, parseErr
			}
			err = c.restClient.SetRepeat(ctx, &alexamodels.MediaControlRequest{Endpoint: req.Target}, payload.Repeat)
		default:
			return nil, &alexaapimodels.BadRequestError{
				Message: fmt.Sprintf("media playback operation %s not supported", req.Name),
			}
		}
		if err != nil {
			return nil, err
		}
		return &alexaapimodels.ControlResponse{}, nil
	}
}

func (c *Client) controlPower(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	// Type assert payload to ControlPowerPayload (can be empty)
	if req.Payload != nil {
		_, err := cast[alexaapimodels.ControlPowerPayload](req)
		if err != nil {
			return nil, err
		}
	}

	var state alexaapimodels.PowerState
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
	isFireTV := deviceAccountID != "" && req.Target.GetDeviceFamily() == "FIRE_TV"

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
	} else {
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
}

func (c *Client) controlVolume(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	volumeReq := &alexamodels.VolumeControlRequest{
		Endpoint: req.Target,
	}

	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetVolume:
		payload, err := parseAndValidate[alexaapimodels.ControlVolumeSetPayload](ctx, req, alexaapimodels.FeatureOperationNameSetVolume)
		if err != nil {
			return nil, err
		}
		volumeReq.Volume = &payload.Volume
		volumeReq.SetVolume = true
	case alexaapimodels.FeatureOperationNameAdjustVolume:
		payload, err := parseAndValidate[alexaapimodels.ControlVolumeAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustVolume)
		if err != nil {
			return nil, err
		}
		volumeReq.Delta = &payload.Delta
		volumeReq.SetVolume = false
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for volume: %s", req.Name),
		}
	}

	// Determine if we should use GraphQL (if endpointId is provided or delta is set)
	// or REST API (if endpoint is provided and volume is set)
	useGraphQL := req.Target.GetDeviceSerialNumber() != "" && req.Target.GetDeviceFamily() != string(alexamodels.WholeHomeDeviceFamily)
	useNpCommandInterfaces := req.Target.GetDeviceSerialNumber() != ""
	if useGraphQL {
		// Route to GraphQL speaker feature control
		endpointID := req.Target.GetEndpointId()
		if endpointID == "" {
			return nil, &alexaapimodels.BadRequestError{
				Message: "endpointId is required for GraphQL volume control",
			}
		}

		speakerReq := &alexamodels.SpeakerControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: endpointID,
			},
			Volume:    volumeReq.Volume,
			Delta:     volumeReq.Delta,
			SetVolume: volumeReq.SetVolume,
		}

		return c.graphqlClient.ControlSpeakerFeature(ctx, speakerReq)
	} else if useNpCommandInterfaces {
		// Route to REST API volume control
		if volumeReq.Volume == nil {
			return nil, &alexaapimodels.BadRequestError{
				Message: "volume is required for REST volume control",
			}
		}

		volumeControlReq := &alexamodels.VolumeControlRequest{
			Endpoint:   req.Target,
			Volume:     volumeReq.Volume,
			CustomerID: volumeReq.CustomerID,
		}

		err := c.restClient.SetVolume(ctx, volumeControlReq)
		if err != nil {
			return nil, err
		}
		return &alexaapimodels.ControlResponse{}, nil
	} else {

		switch req.Name {
		case alexaapimodels.FeatureOperationNameSetVolume:

			err := c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNameSpeaker),
					OperationName: string(alexaapimodels.FeatureOperationNameSetVolume), Payload: map[string]interface{}{
						"volume": *volumeReq.Volume,
					}})
			if err != nil {
				return nil, err
			}
		case alexaapimodels.FeatureOperationNameAdjustVolume:
			err := c.restClient.SendInterfaceMessage(ctx,
				&alexamodels.InterfaceMessageRequest{Endpoint: req.Target,
					FeatureName:   string(alexaapimodels.FeatureNameSpeaker),
					OperationName: string(alexaapimodels.FeatureOperationNameAdjustVolume), Payload: map[string]interface{}{
						"volume": *volumeReq.Delta,
					}})
			if err != nil {
				return nil, err
			}
			return &alexaapimodels.ControlResponse{}, nil
		default:
			return nil, &alexaapimodels.BadRequestError{
				Message: fmt.Sprintf("unknown operation name for volume: %s", req.Name),
			}
		}
		return &alexaapimodels.ControlResponse{}, nil
	}
}

func (c *Client) controlThermostat(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	// Check if this is a mode operation or temperature operation
	switch req.Name {
	case alexaapimodels.FeatureOperationNameSetThermostatMode:
		// Handle thermostat mode
		payload, err := parseAndValidate[alexaapimodels.ControlThermostatModePayload](ctx, req, alexaapimodels.FeatureOperationNameSetThermostatMode)
		if err != nil {
			return nil, err
		}

		modeReq := &alexamodels.ThermostatModeControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: req.Target.GetEndpointId(),
			},
			Mode: payload.Mode,
		}

		return c.graphqlClient.ControlThermostatModeFeature(ctx, modeReq)
	case alexaapimodels.FeatureOperationNameSetTargetSetpoint:
		// Handle setting target temperature
		payload, err := parseAndValidate[alexaapimodels.ControlThermostatSetpointPayload](ctx, req, alexaapimodels.FeatureOperationNameSetTargetSetpoint)
		if err != nil {
			return nil, err
		}

		thermostatReq := &alexamodels.ThermostatControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: req.Target.GetEndpointId(),
			},
			Value:       &payload.Value,
			Scale:       payload.Scale,
			SetSetpoint: true,
		}

		return c.graphqlClient.ControlThermostatFeature(ctx, thermostatReq)
	case alexaapimodels.FeatureOperationNameAdjustTargetSetpoint:
		// Handle adjusting target temperature
		payload, err := parseAndValidate[alexaapimodels.ControlThermostatSetpointAdjustPayload](ctx, req, alexaapimodels.FeatureOperationNameAdjustTargetSetpoint)
		if err != nil {
			return nil, err
		}

		thermostatReq := &alexamodels.ThermostatControlRequest{
			BaseFeatureRequest: alexamodels.BaseFeatureRequest{
				EndpointID: req.Target.GetEndpointId(),
			},
			Delta:       &payload.Delta,
			Scale:       payload.Scale,
			SetSetpoint: false,
		}

		return c.graphqlClient.ControlThermostatFeature(ctx, thermostatReq)
	default:
		return nil, &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("unknown operation name for thermostat: %s", req.Name),
		}
	}
}
