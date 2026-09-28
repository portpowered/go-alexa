package tui

import (
	"context"
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type controlModel struct {
	endpoint    *alexaapimodels.Endpoint
	client      alexa.ClientInterface
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
		cursor: 0,
	}
}

func (m controlModel) Init() tea.Cmd {
	return nil
}

func (m controlModel) Update(msg tea.Msg) (controlModel, tea.Cmd) {
	var cmd tea.Cmd

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
			switch msg.String() {
			case "enter":
				// Execute control with input value
				return m, m.executeControlWithInput(m.inputValue)

			case "esc":
				m.showInput = false
				m.inputValue = ""
				return m, nil

			default:
				// Handle text input
				if len(msg.Runes) > 0 {
					m.inputValue += string(msg.Runes)
				} else if msg.String() == "backspace" {
					if len(m.inputValue) > 0 {
						m.inputValue = m.inputValue[:len(m.inputValue)-1]
					}
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			maxCursor := len(getControlOptions(m.endpoint)) - 1
			if m.cursor < maxCursor {
				m.cursor++
			}
			return m, nil

		case "enter":
			options := getControlOptions(m.endpoint)
			if m.cursor < len(options) {
				option := options[m.cursor]
				if option.needsInput {
					m.showInput = true
					m.inputPrompt = option.prompt
					m.inputValue = ""
					return m, nil
				}
				return m, m.executeControl(option)
			}
			return m, nil

		case "b", "esc":
			return m, func() tea.Msg {
				return backToEndpointsMsg{}
			}

		case "s":
			// View state
			return m, func() tea.Msg {
				return viewStateMsg{}
			}
		}
	}

	return m, cmd
}

func (m controlModel) View() string {
	if m.endpoint == nil {
		return "No endpoint selected"
	}

	view := titleStyle.Render(fmt.Sprintf("Control: %s", m.endpoint.EndpointID)) + "\n"
	view += helpStyle.Render(fmt.Sprintf("Device: %s", m.endpoint.DeviceType)) + "\n\n"

	if m.showInput {
		view += m.inputPrompt + ": " + m.inputValue + "\n"
		view += helpStyle.Render("Enter: Confirm | Esc: Cancel")
		return view
	}

	options := getControlOptions(m.endpoint)
	for i, option := range options {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		line := fmt.Sprintf("%s %s", cursor, option.label)
		if i == m.cursor {
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

type controlOption struct {
	label      string
	needsInput bool
	prompt     string
	execute    func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd
}

func getControlOptions(endpoint *alexaapimodels.Endpoint) []controlOption {
	options := []controlOption{}

	// Check supported features
	hasPower := hasFeature(endpoint, alexaapimodels.FeatureNamePower)
	hasVolume := hasFeature(endpoint, alexaapimodels.FeatureNameSpeaker)
	hasPlayback := hasFeature(endpoint, alexaapimodels.FeatureNamePlayback)
	hasBrightness := hasFeature(endpoint, alexaapimodels.FeatureNameBrightness)
	hasColor := hasFeature(endpoint, alexaapimodels.FeatureNameColor)
	hasColorTemp := hasFeature(endpoint, alexaapimodels.FeatureNameColorTemperature)
	hasLock := hasFeature(endpoint, alexaapimodels.FeatureNameLock)
	hasToggle := hasFeature(endpoint, alexaapimodels.FeatureNameToggle)
	hasNotification := hasFeature(endpoint, alexaapimodels.FeatureNameNotification)
	hasAnnouncement := hasFeature(endpoint, alexaapimodels.FeatureNameAnnouncement)
	hasSpeechSynthesizer := hasFeature(endpoint, alexaapimodels.FeatureNameSpeechSynthesizer)
	hasAudioPlayer := hasFeature(endpoint, alexaapimodels.FeatureNameAudioPlayer)
	hasNavigateHome := hasFeature(endpoint, alexaapimodels.FeatureNameNavigation)

	if hasPower {
		options = append(options,
			controlOption{label: "Power: ON", execute: executePowerOn},
			controlOption{label: "Power: OFF", execute: executePowerOff},
		)
	}

	if hasVolume {
		options = append(options,
			controlOption{label: "Volume: Set", needsInput: true, prompt: "Enter volume (0-100)", execute: executeSetVolume},
			controlOption{label: "Volume: +10", execute: executeVolumeDelta(10)},
			controlOption{label: "Volume: -10", execute: executeVolumeDelta(-10)},
		)
	}

	if hasPlayback {
		options = append(options,
			controlOption{label: "Playback: Play", execute: executePlayback("play")},
			controlOption{label: "Playback: Pause", execute: executePlayback("pause")},
			controlOption{label: "Playback: Resume", execute: executePlayback("resume")},
			controlOption{label: "Playback: Next", execute: executePlayback("next")},
			controlOption{label: "Playback: Previous", execute: executePlayback("previous")},
			controlOption{label: "Playback: Stop", execute: executePlayback("stop")},
		)
	}

	if hasBrightness {
		options = append(options,
			controlOption{label: "Brightness: Set", needsInput: true, prompt: "Enter brightness (0-100)", execute: executeBrightness},
		)
	}

	if hasColor {
		options = append(options,
			controlOption{label: "Color: Set", needsInput: true, prompt: "Enter color (hue 0-360)", execute: executeColor},
		)
	}

	if hasColorTemp {
		options = append(options,
			controlOption{label: "Color Temp: Set", needsInput: true, prompt: "Enter temperature (K)", execute: executeColorTemp},
		)
	}

	if hasLock {
		options = append(options,
			controlOption{label: "Lock: Lock", execute: executeLock(true)},
			controlOption{label: "Lock: Unlock", execute: executeLock(false)},
		)
	}

	if hasToggle {
		options = append(options,
			controlOption{label: "Toggle: ON", execute: executeToggle(true)},
			controlOption{label: "Toggle: OFF", execute: executeToggle(false)},
		)
	}

	if hasNotification {
		options = append(options,
			controlOption{label: "Send Notification", needsInput: true, prompt: "Enter message", execute: executeNotification},
		)
	}

	if hasAnnouncement {
		options = append(options,
			controlOption{label: "Announcement: Speak", needsInput: true, prompt: "Enter message", execute: executeAnnouncement("speak")},
			controlOption{label: "Announcement: Show", needsInput: true, prompt: "Enter message", execute: executeAnnouncement("show")},
			controlOption{label: "Announcement: All", needsInput: true, prompt: "Enter message", execute: executeAnnouncement("all")},
		)
	}

	if hasSpeechSynthesizer {
		options = append(options,
			controlOption{label: "Speak", needsInput: true, prompt: "Enter message to speak", execute: executeSpeak},
		)
	}

	if hasAudioPlayer {
		options = append(options,
			controlOption{label: "Audio Player: Play", needsInput: true, prompt: "Enter search phrase", execute: executeAudioPlayer},
		)
	}

	if hasNavigateHome {
		options = append(options,
			controlOption{label: "Navigate Home", execute: executeNavigateHome},
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

// Control execution functions
func executePowerOn(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
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

func executePowerOff(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
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

func executeSetVolume(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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
		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}
		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}
		return controlExecutedMsg{success: true, message: fmt.Sprintf("Volume set to %d", vol)}
	}
}

func executeVolumeDelta(delta int) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
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

func executePlayback(op string) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
		return func() tea.Msg {
			var operation alexaapimodels.FeatureOperationName
			switch op {
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
				return controlExecutedMsg{success: false, message: fmt.Sprintf("Unknown playback operation: %s", op)}
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
			return controlExecutedMsg{success: true, message: fmt.Sprintf("Playback: %s", op)}
		}
	}
}

func executeBrightness(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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
		resp, err := client.Control(ctx, req)
		if err != nil {
			return controlExecutedMsg{success: false, message: err.Error()}
		}
		if len(resp.Errors) > 0 {
			return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
		}
		return controlExecutedMsg{success: true, message: fmt.Sprintf("Brightness set to %d", brightness)}
	}
}

func executeColor(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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

func executeColorTemp(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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
			state := "LOCKED"
			if !lock {
				state = "UNLOCKED"
			}
			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNameLock,
				Name:      alexaapimodels.FeatureOperationNameSetLockState,
				Payload: alexaapimodels.ControlLockPayload{
					State: state,
				},
			}
			resp, err := client.Control(ctx, req)
			if err != nil {
				return controlExecutedMsg{success: false, message: err.Error()}
			}
			if len(resp.Errors) > 0 {
				return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
			}
			action := "locked"
			if !lock {
				action = "unlocked"
			}
			return controlExecutedMsg{success: true, message: fmt.Sprintf("Lock %s", action)}
		}
	}
}

func executeToggle(on bool) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
	return func(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
		return func() tea.Msg {
			state := "ON"
			if !on {
				state = "OFF"
			}
			req := alexaapimodels.ControlRequest{
				Target:    endpoint,
				Namespace: alexaapimodels.FeatureNameToggle,
				Name:      alexaapimodels.FeatureOperationNameSetToggleState,
				Payload: alexaapimodels.ControlTogglePayload{
					State: state,
				},
			}
			resp, err := client.Control(ctx, req)
			if err != nil {
				return controlExecutedMsg{success: false, message: err.Error()}
			}
			if len(resp.Errors) > 0 {
				return controlExecutedMsg{success: false, message: resp.Errors[0].Message}
			}
			action := "ON"
			if !on {
				action = "OFF"
			}
			return controlExecutedMsg{success: true, message: fmt.Sprintf("Toggle set to %s", action)}
		}
	}
}

func executeNotification(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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

func executeNavigateHome(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, _ string) tea.Cmd {
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

func executeAnnouncement(method string) func(context.Context, alexa.ClientInterface, *alexaapimodels.Endpoint, string) tea.Cmd {
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
			return controlExecutedMsg{success: true, message: fmt.Sprintf("Announcement sent via %s", method)}
		}
	}
}

func executeSpeak(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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

func executeAudioPlayer(ctx context.Context, client alexa.ClientInterface, endpoint *alexaapimodels.Endpoint, value string) tea.Cmd {
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
		return controlExecutedMsg{success: true, message: fmt.Sprintf("Audio playback started: %s", value)}
	}
}
