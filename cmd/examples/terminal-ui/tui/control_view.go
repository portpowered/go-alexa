package tui

import (
	"context"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

const (
	volumeAdjustmentStep = 10
	keyEnter             = "enter"
	keyEscape            = "esc"
	keyDown              = "down"
)

type controlModel struct {
	endpoint *alexaapimodels.Endpoint
	client   alexa.ClientInterface
	//nolint:containedctx // Control commands inherit cancellation from the app model.
	ctx         context.Context
	cursor      int
	width       int
	height      int
	showInput   bool
	inputValue  string
	inputPrompt string
	lastResult  string
	lastError   string
}

func newControlModel() controlModel {
	return controlModel{
		endpoint:    nil,
		client:      nil,
		ctx:         nil,
		cursor:      0,
		width:       0,
		height:      0,
		showInput:   false,
		inputValue:  "",
		inputPrompt: "",
		lastResult:  "",
		lastError:   "",
	}
}

func (m controlModel) Init() tea.Cmd {
	return nil
}

func (m controlModel) Update(msg tea.Msg) (controlModel, tea.Cmd) {
	switch msg := msg.(type) {
	case endpointSelectedMsg:
		m.endpoint = msg.endpoint
		m.cursor = 0

		return m, nil

	case controlExecutedMsg:
		if msg.success {
			m.lastResult = msg.message
			m.lastError = ""
		} else {
			m.lastError = msg.message
			m.lastResult = ""
		}

		m.showInput = false

		return m, nil

	case tea.KeyMsg:
		if m.showInput {
			return updateControlInput(m, msg)
		}

		return updateControlSelection(m, msg)
	}

	return m, nil
}

func (m controlModel) View() string {
	if m.endpoint == nil {
		return "No endpoint selected"
	}

	view := titleStyle.Render("Control: "+m.endpoint.EndpointID) + "\n"
	view += helpStyle.Render("Device: "+m.endpoint.DeviceType) + "\n\n"

	if m.showInput {
		view += m.inputPrompt + ": " + m.inputValue + "\n"
		view += helpStyle.Render("Enter: Confirm | Esc: Cancel")

		return view
	}

	options := getControlOptions(m.endpoint)
	for optionIndex, option := range options {
		cursor := " "
		if optionIndex == m.cursor {
			cursor = ">"
		}

		line := fmt.Sprintf("%s %s", cursor, option.label)
		if optionIndex == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = lipgloss.NewStyle().Render(line)
		}

		view += line + "\n"
	}

	if m.lastResult != "" {
		view += "\n" + successStyle.Render("Success: "+m.lastResult)
	}

	if m.lastError != "" {
		view += "\n" + errorStyle.Render("Error: "+m.lastError)
	}

	view += "\n" + helpStyle.Render("↑/↓: Navigate | Enter: Select | b: Back | s: View State | q: Quit")

	return view
}

func updateControlInput(model controlModel, msg tea.KeyMsg) (controlModel, tea.Cmd) {
	switch msg.String() {
	case keyEnter:
		return model, model.executeControlWithInput(model.inputValue)
	case keyEscape:
		model.showInput = false
		model.inputValue = ""
	default:
		if len(msg.Runes) > 0 {
			model.inputValue += string(msg.Runes)
		} else if msg.String() == "backspace" && len(model.inputValue) > 0 {
			model.inputValue = model.inputValue[:len(model.inputValue)-1]
		}
	}

	return model, nil
}

func updateControlSelection(model controlModel, msg tea.KeyMsg) (controlModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if model.cursor > 0 {
			model.cursor--
		}
	case keyDown, "j":
		maxCursor := len(getControlOptions(model.endpoint)) - 1
		if model.cursor < maxCursor {
			model.cursor++
		}
	case keyEnter:
		return selectControlOption(model)
	case "b", keyEscape:
		return model, func() tea.Msg { return backToEndpointsMsg{} }
	case "s":
		return model, func() tea.Msg { return viewStateMsg{} }
	}

	return model, nil
}

func selectControlOption(model controlModel) (controlModel, tea.Cmd) {
	options := getControlOptions(model.endpoint)
	if model.cursor >= len(options) {
		return model, nil
	}

	option := options[model.cursor]
	if !option.needsInput {
		return model, model.executeControl(option)
	}

	model.showInput = true
	model.inputPrompt = option.prompt
	model.inputValue = ""

	return model, nil
}

type controlOption struct {
	label      string
	needsInput bool
	prompt     string
	execute    func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd
}

func getControlOptions(endpoint *alexaapimodels.Endpoint) []controlOption {
	options := make([]controlOption, 0)
	options = appendPlaybackControlOptions(endpoint, options)
	options = appendFeatureControlOptions(endpoint, options)

	return appendCommunicationControlOptions(endpoint, options)
}

func appendPlaybackControlOptions(endpoint *alexaapimodels.Endpoint, options []controlOption) []controlOption {
	if hasFeature(endpoint, alexaapimodels.FeatureNamePower) {
		options = append(options,
			controlOption{label: "Power: ON", needsInput: false, prompt: "", execute: executePowerOn},
			controlOption{label: "Power: OFF", needsInput: false, prompt: "", execute: executePowerOff},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameSpeaker) {
		options = append(
			options,
			controlOption{
				label:      "Volume: Set",
				needsInput: true,
				prompt:     "Enter volume (0-100)",
				execute:    executeSetVolume,
			},
			controlOption{label: "Volume: +10", needsInput: false, prompt: "", execute: executeVolumeDelta(volumeAdjustmentStep)},
			controlOption{label: "Volume: -10", needsInput: false, prompt: "", execute: executeVolumeDelta(-volumeAdjustmentStep)},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNamePlayback) {
		options = append(options,
			controlOption{label: "Playback: Play", needsInput: false, prompt: "", execute: executePlayback("play")},
			controlOption{label: "Playback: Pause", needsInput: false, prompt: "", execute: executePlayback("pause")},
			controlOption{label: "Playback: Resume", needsInput: false, prompt: "", execute: executePlayback("resume")},
			controlOption{label: "Playback: Next", needsInput: false, prompt: "", execute: executePlayback("next")},
			controlOption{label: "Playback: Previous", needsInput: false, prompt: "", execute: executePlayback("previous")},
			controlOption{label: "Playback: Stop", needsInput: false, prompt: "", execute: executePlayback("stop")},
		)
	}

	return options
}

func appendFeatureControlOptions(endpoint *alexaapimodels.Endpoint, options []controlOption) []controlOption {
	if hasFeature(endpoint, alexaapimodels.FeatureNameBrightness) {
		options = append(
			options,
			controlOption{
				label:      "Brightness: Set",
				needsInput: true,
				prompt:     "Enter brightness (0-100)",
				execute:    executeBrightness,
			},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameColor) {
		options = append(
			options,
			controlOption{
				label:      "Color: Set",
				needsInput: true,
				prompt:     "Enter color (hue 0-360)",
				execute:    executeColor,
			},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameColorTemperature) {
		options = append(
			options,
			controlOption{
				label:      "Color Temp: Set",
				needsInput: true,
				prompt:     "Enter temperature (K)",
				execute:    executeColorTemp,
			},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameLock) {
		options = append(options,
			controlOption{label: "Lock: Lock", needsInput: false, prompt: "", execute: executeLock(true)},
			controlOption{label: "Lock: Unlock", needsInput: false, prompt: "", execute: executeLock(false)},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameToggle) {
		options = append(options,
			controlOption{label: "Toggle: ON", needsInput: false, prompt: "", execute: executeToggle(true)},
			controlOption{label: "Toggle: OFF", needsInput: false, prompt: "", execute: executeToggle(false)},
		)
	}

	return options
}

func appendCommunicationControlOptions(endpoint *alexaapimodels.Endpoint, options []controlOption) []controlOption {
	if hasFeature(endpoint, alexaapimodels.FeatureNameNotification) {
		options = append(
			options,
			controlOption{
				label:      "Send Notification",
				needsInput: true,
				prompt:     "Enter message",
				execute:    executeNotification,
			},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameAnnouncement) {
		options = append(
			options,
			controlOption{
				label:      "Announcement: Speak",
				needsInput: true,
				prompt:     "Enter message",
				execute:    executeAnnouncement("speak"),
			},
			controlOption{
				label:      "Announcement: Show",
				needsInput: true,
				prompt:     "Enter message",
				execute:    executeAnnouncement("show"),
			},
			controlOption{
				label:      "Announcement: All",
				needsInput: true,
				prompt:     "Enter message",
				execute:    executeAnnouncement("all"),
			},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameSpeechSynthesizer) {
		options = append(options,
			controlOption{label: "Speak", needsInput: true, prompt: "Enter message to speak", execute: executeSpeak},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameAudioPlayer) {
		options = append(
			options,
			controlOption{
				label:      "Audio Player: Play",
				needsInput: true,
				prompt:     "Enter search phrase",
				execute:    executeAudioPlayer,
			},
		)
	}

	if hasFeature(endpoint, alexaapimodels.FeatureNameNavigation) {
		options = append(options,
			controlOption{label: "Navigate Home", needsInput: false, prompt: "", execute: executeNavigateHome},
		)
	}

	return options
}

func hasFeature(endpoint *alexaapimodels.Endpoint, namespace alexaapimodels.FeatureName) bool {
	for _, feat := range endpoint.Features {
		if feat.IsType(namespace) {
			return true
		}
	}

	return false
}

func (m controlModel) executeControl(option controlOption) tea.Cmd {
	if m.client == nil || m.endpoint == nil {
		return func() tea.Msg {
			return controlExecutedMsg{success: false, message: "Client or endpoint not initialized"}
		}
	}

	return option.execute(m.ctx, m.client, m.endpoint, "")
}

func (m controlModel) executeControlWithInput(value string) tea.Cmd {
	if m.client == nil || m.endpoint == nil {
		return func() tea.Msg {
			return controlExecutedMsg{success: false, message: "Client or endpoint not initialized"}
		}
	}

	options := getControlOptions(m.endpoint)
	if m.cursor < len(options) {
		return options[m.cursor].execute(m.ctx, m.client, m.endpoint, value)
	}

	return func() tea.Msg {
		return controlExecutedMsg{success: false, message: "Invalid option"}
	}
}

// Control execution functions.
func executePowerOn(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	_ string,
) tea.Cmd {
	return func() tea.Msg {
		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNamePower,
			Name:      alexaapimodels.FeatureOperationNameTurnOn,
			Payload:   alexaapimodels.ControlPowerPayload{},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: "Power turned ON"}
	}
}

func executePowerOff(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	_ string,
) tea.Cmd {
	return func() tea.Msg {
		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNamePower,
			Name:      alexaapimodels.FeatureOperationNameTurnOff,
			Payload:   alexaapimodels.ControlPowerPayload{},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: "Power turned OFF"}
	}
}

func executeSetVolume(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		vol, err := strconv.Atoi(value)
		if err != nil || vol < 0 || vol > 100 {
			return controlExecutedMsg{success: false, message: "Invalid volume (0-100)"}
		}

		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameSpeaker,
			Name:      alexaapimodels.FeatureOperationNameSetVolume,
			Payload: alexaapimodels.ControlVolumeSetPayload{
				Volume: vol,
			},
		}

		return runControlRequest(ctx, client, req, fmt.Sprintf("Volume set to %d", vol))
	}
}

func executeVolumeDelta(
	delta int,
) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
		return func() tea.Msg {
			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNameSpeaker,
				Name:      alexaapimodels.FeatureOperationNameAdjustVolume,
				Payload: alexaapimodels.ControlVolumeAdjustPayload{
					Delta: delta,
				},
			}

			resp, err := client.Control(ctx, req)
			if err != nil {
				return controlExecutedMsg{success: false, message: err.Error()}
			}

			if len(resp.Errors) > 0 {
				return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
			}

			return controlExecutedMsg{success: true, message: fmt.Sprintf("Volume adjusted by %d", delta)}
		}
	}
}

func executePlayback(operationName string) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
		return func() tea.Msg {
			var operation alexaapimodels.FeatureOperationName

			switch operationName {
			case "play":
				// "play" maps to "resume" for playback operations
				operation = alexaapimodels.FeatureOperationNameResume
			case "pause":
				operation = alexaapimodels.FeatureOperationNamePause
			case "resume":
				operation = alexaapimodels.FeatureOperationNameResume
			case "next":
				operation = alexaapimodels.FeatureOperationNameNext
			case "previous":
				operation = alexaapimodels.FeatureOperationNamePrevious
			case "stop":
				operation = alexaapimodels.FeatureOperationNameStop
			default:
				return controlExecutedMsg{success: false, message: "Unknown playback operation: " + operationName}
			}

			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNamePlayback,
				Name:      operation,
				Payload:   alexaapimodels.ControlPlaybackPayload{},
			}

			resp, err := client.Control(ctx, req)
			if err != nil {
				return controlExecutedMsg{success: false, message: err.Error()}
			}

			if len(resp.Errors) > 0 {
				return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
			}

			return controlExecutedMsg{success: true, message: "Playback: " + operationName}
		}
	}
}

func executeBrightness(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		brightness, err := strconv.Atoi(value)
		if err != nil || brightness < 0 || brightness > 100 {
			return controlExecutedMsg{success: false, message: "Invalid brightness (0-100)"}
		}

		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameBrightness,
			Name:      alexaapimodels.FeatureOperationNameSetBrightness,
			Payload: alexaapimodels.ControlBrightnessSetPayload{
				Brightness: brightness,
			},
		}

		return runControlRequest(ctx, client, req, fmt.Sprintf("Brightness set to %d", brightness))
	}
}

func executeColor(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		hue, err := strconv.ParseFloat(value, 64)
		if err != nil || hue < 0 || hue > 360 {
			return controlExecutedMsg{success: false, message: "Invalid hue (0-360)"}
		}

		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameColor,
			Name:      alexaapimodels.FeatureOperationNameSetColor,
			Payload: alexaapimodels.ControlColorPayload{
				Hue:        hue,
				Saturation: 1.0, // Default saturation
				Brightness: 1.0, // Default brightness
			},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: fmt.Sprintf("Color set to hue %v", hue)}
	}
}

func executeColorTemp(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		temp, err := strconv.Atoi(value)
		if err != nil || temp < 0 {
			return controlExecutedMsg{success: false, message: "Invalid temperature"}
		}

		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameColorTemperature,
			Name:      alexaapimodels.FeatureOperationNameSetColorTemperature,
			Payload: alexaapimodels.ControlColorTemperatureSetPayload{
				ColorTemperature: temp,
			},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: fmt.Sprintf("Color temperature set to %dK", temp)}
	}
}

func executeLock(lock bool) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
		return func() tea.Msg {
			state := string(alexamodels.LockStateValueLocked)
			if !lock {
				state = string(alexamodels.LockStateValueUnlocked)
			}

			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNameLock,
				Name:      alexaapimodels.FeatureOperationNameSetLockState,
				Payload: alexaapimodels.ControlLockPayload{
					State: state,
				},
			}

			action := "locked"
			if !lock {
				action = "unlocked"
			}

			return runControlRequest(ctx, client, req, "Lock "+action)
		}
	}
}

func executeToggle(isOn bool) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
		return func() tea.Msg {
			state := string(alexamodels.ToggleStateValueOn)
			if !isOn {
				state = string(alexamodels.ToggleStateValueOff)
			}

			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNameToggle,
				Name:      alexaapimodels.FeatureOperationNameSetToggleState,
				Payload: alexaapimodels.ControlTogglePayload{
					State: state,
				},
			}

			action := "ON"
			if !isOn {
				action = "OFF"
			}

			return runControlRequest(ctx, client, req, "Toggle set to "+action)
		}
	}
}

func runControlRequest(
	ctx context.Context,
	client alexa.ClientInterface,
	req alexaapimodels.ControlRequest,
	successMessage string,
) controlExecutedMsg {
	response, err := client.Control(ctx, req)
	if err != nil {
		return controlExecutedMsg{success: false, message: err.Error()}
	}

	if len(response.Errors) > 0 {
		return controlExecutedMsg{success: false, message: response.Errors[0].Message}
	}

	return controlExecutedMsg{success: true, message: successMessage}
}

func executeNotification(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		if value == "" {
			return controlExecutedMsg{success: false, message: "Message cannot be empty"}
		}

		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameNotification,
			Name:      alexaapimodels.FeatureOperationNameSend,
			Payload: alexaapimodels.ControlNotificationPayload{
				Message: value,
			},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: "Notification sent"}
	}
}

func executeNavigateHome(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	_ string,
) tea.Cmd {
	return func() tea.Msg {
		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameNavigation,
			Name:      alexaapimodels.FeatureOperationNameNavigateHome,
			Payload:   alexaapimodels.ControlNavigationPayload{},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: "Navigate home command sent"}
	}
}

func executeAnnouncement(
	method string,
) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
		return func() tea.Msg {
			if value == "" {
				return controlExecutedMsg{success: false, message: "Message cannot be empty"}
			}

			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNameAnnouncement,
				Name:      alexaapimodels.FeatureOperationNameSend,
				Payload: alexaapimodels.ControlAnnouncementPayload{
					Message: value,
					Method:  method,
				},
			}

			resp, err := client.Control(ctx, req)
			if err != nil {
				return controlExecutedMsg{success: false, message: err.Error()}
			}

			if len(resp.Errors) > 0 {
				return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
			}

			return controlExecutedMsg{success: true, message: "Announcement sent via " + method}
		}
	}
}

func executeSpeak(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		if value == "" {
			return controlExecutedMsg{success: false, message: "Message cannot be empty"}
		}

		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameSpeechSynthesizer,
			Name:      alexaapimodels.FeatureOperationNameSpeak,
			Payload: alexaapimodels.ControlSpeechSynthesizerPayload{
				Message: value,
			},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: "Speech command sent"}
	}
}

func executeAudioPlayer(
	ctx context.Context,
	client alexa.ClientInterface,
	endpoint *alexaapimodels.Endpoint,
	value string,
) tea.Cmd {
	return func() tea.Msg {
		if value == "" {
			return controlExecutedMsg{success: false, message: "Search phrase cannot be empty"}
		}
		// Default provider ID - can be customized if needed
		req := alexaapimodels.ControlRequest{
			Target:    endpoint,
			Namespace: alexaapimodels.FeatureNameAudioPlayer,
			Name:      alexaapimodels.FeatureOperationNamePlay,
			Payload: alexaapimodels.ControlAudioPlayerPayload{
				ProviderID:   alexaapimodels.ProviderIDAmazon,
				SearchPhrase: value,
			},
		}

		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}

		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}

		return controlExecutedMsg{success: true, message: "Audio playback started: " + value}
	}
}
