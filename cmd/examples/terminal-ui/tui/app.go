// Package tui implements the terminal Alexa client interface.
package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	eventChannelBufferSize = 100
	splitPaneCount         = 2
	clearErrorDelay        = 3 * time.Second
)

type config struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	CustomerID   string `json:"customerID"`
}

type viewMode int

const (
	viewTokenInput viewMode = iota
	viewEndpoints
	viewControl
	viewState
)

// Model is the stateful Bubble Tea model for the terminal client.
type Model struct {
	// Client and connection
	client    alexa.ClientInterface
	eventChan chan *alexaapimodels.Event
	//nolint:containedctx // Asynchronous commands share the Bubble Tea model lifetime.
	ctx             context.Context
	cancel          context.CancelFunc
	eventsConnected bool

	// Views
	tokenInput tokenInputModel
	endpoints  endpointsModel
	control    controlModel
	state      stateModel
	events     eventsModel

	// Current state
	currentView      viewMode
	selectedEndpoint *alexaapimodels.Endpoint
	errorMsg         string
	width            int
	height           int
}

// NewModel creates an initialized terminal client model.
func NewModel() *Model {
	ctx, cancel := context.WithCancel(context.Background())
	eventChan := make(chan *alexaapimodels.Event, eventChannelBufferSize)
	model := &Model{
		client:           nil,
		ctx:              ctx,
		cancel:           cancel,
		eventChan:        eventChan,
		eventsConnected:  false,
		currentView:      viewTokenInput,
		selectedEndpoint: nil,
		errorMsg:         "",
		width:            0,
		height:           0,
		tokenInput:       newTokenInputModel(),
		endpoints:        newEndpointsModel(),
		control:          newControlModel(),
		state:            newStateModel(),
		events:           newEventsModel(),
	}

	return model
}

func getConfigFromFile() *config {
	// Look for config.json in the current directory
	configPath := "config.json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist or can't be read - that's okay
		return nil
	}

	var cfg config
	{
		err := json.Unmarshal(data, &cfg)
		if err != nil {
			// Config file exists but is invalid - that's okay, we'll prompt for input
			return nil
		}
	}

	// Validate that all required fields are present
	if cfg.AccessToken == "" || cfg.RefreshToken == "" || cfg.CustomerID == "" {
		return nil
	}

	return &cfg
}

// Init starts model initialization and optional configuration validation.
func (m *Model) Init() tea.Cmd {
	// If config is provided via file, auto-validate it
	if cfg := getConfigFromFile(); cfg != nil {
		return tea.Batch(m.tokenInput.Init(), validateToken(cfg.AccessToken, cfg.RefreshToken, cfg.CustomerID))
	}

	return m.tokenInput.Init()
}

// Update applies a message and returns the next command.
//
//nolint:ireturn // Bubble Tea requires Model.Update to return the tea.Model interface.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := updateEventMessage(m, msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.endpoints.width = msg.Width / splitPaneCount
		m.endpoints.height = msg.Height
		m.events.width = msg.Width / splitPaneCount
		m.events.height = msg.Height
		m.control.width = msg.Width / splitPaneCount
		m.control.height = msg.Height
		m.state.width = msg.Width / splitPaneCount
		m.state.height = msg.Height

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || (key == "q" && m.currentView == viewTokenInput) {
			m.closeSession()

			return m, tea.Quit
		}

	case tokenValidatedMsg:
		m.client = msg.client
		m.control.client = msg.client
		m.control.ctx = m.ctx
		m.currentView = viewEndpoints
		// Load endpoints
		// Also connect to events
		return m, tea.Batch(m.loadEndpoints(), connectEventsCmd(m.ctx, m.client, m.eventChan), listenForEvents(m.eventChan))

	case endpointSelectedMsg:
		m.selectedEndpoint = msg.endpoint
		m.control.endpoint = msg.endpoint
		m.state.endpoint = msg.endpoint
		m.currentView = viewControl

	case backToEndpointsMsg:
		m.currentView = viewEndpoints

	case viewStateMsg:
		m.currentView = viewState

		return m, m.loadState()

	case errorMsg:
		m.errorMsg = msg.error
		// Clear error after a delay
		cmds = append(cmds, clearErrorAfterDelay())

	case controlExecutedMsg:
		// Show result in control view
		_, cmd := m.control.Update(msg)
		cmds = append(cmds, cmd)

	case refreshEndpointsMsg:
		return m, m.loadEndpoints()

	case refreshStateMsg:
		return m, m.loadState()

	case stateLoadedMsg:
		m.state.state = findEndpointByID(msg.response, m.selectedEndpoint.ID)

		return m, nil

	case tokenValidationErrorMsg:
		m.tokenInput.err = msg.err

		return m, nil

	case clearErrorMsg:
		m.errorMsg = ""

		return m, nil
	}

	cmds = append(cmds, updateCurrentView(m, msg)...)

	return m, tea.Batch(cmds...)
}

func updateEventMessage(model *Model, msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case eventReceivedMsg:
		_, cmd := model.events.Update(msg)
		cmds = append(cmds, cmd, listenForEvents(model.eventChan))

	case eventConnectedMsg:
		model.eventsConnected = true
		cmds = append(cmds, listenForEvents(model.eventChan))

	case eventConnectionErrorMsg:
		model.errorMsg = fmt.Sprintf("Event connection error: %v", msg.err)
		cmd := clearErrorAfterDelay()
		cmds = append(cmds, cmd)
	}

	return cmds
}

func updateCurrentView(model *Model, msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	switch model.currentView {
	case viewTokenInput:
		var cmd tea.Cmd

		model.tokenInput, cmd = model.tokenInput.Update(msg)
		cmds = append(cmds, cmd)

	case viewEndpoints:
		var cmd tea.Cmd

		model.endpoints, cmd = model.endpoints.Update(msg)
		cmds = append(cmds, cmd)

	case viewControl:
		var cmd tea.Cmd

		model.control, cmd = model.control.Update(msg)
		cmds = append(cmds, cmd)
		// Also update events view for display
		_, cmd = model.events.Update(tea.WindowSizeMsg{Width: model.width / splitPaneCount, Height: model.height})
		cmds = append(cmds, cmd)

	case viewState:
		var cmd tea.Cmd

		model.state, cmd = model.state.Update(msg)
		cmds = append(cmds, cmd)
	}

	return cmds
}

func findEndpointByID(response *alexaapimodels.UnifiedEndpointListResponse, endpointID string) *alexaapimodels.Endpoint {
	for _, endpoint := range response.Results {
		if endpoint.ID == endpointID {
			return endpoint
		}
	}

	return nil
}

// View renders the active terminal view.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case viewTokenInput:
		return m.tokenInput.View()

	case viewEndpoints:
		left := m.endpoints.View()
		right := m.events.View()

		return splitView(left, right, m.width)

	case viewControl:
		left := m.control.View()
		right := m.events.View()

		view := splitView(left, right, m.width)
		if m.errorMsg != "" {
			view += "\n" + errorStyle.Render("Error: "+m.errorMsg)
		}

		return view

	case viewState:
		left := m.state.View()
		right := m.events.View()

		return splitView(left, right, m.width)
	}

	return ""
}

func (m *Model) closeSession() {
	m.cancel()

	if m.client != nil {
		_ = m.client.Close()
	}
}

func (m *Model) loadEndpoints() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return errorMsg{error: "Client not initialized"}
		}

		response, err := m.client.ListEndpoints(m.ctx, alexaapimodels.EndpointQuery{
			IncludeFields: &alexaapimodels.EndpointIncludeFields{
				Features: true,
			},
		})
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to load endpoints: %v", err)}
		}

		return endpointsLoadedMsg{endpoints: response.Results}
	}
}

func (m *Model) loadState() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil || m.selectedEndpoint == nil {
			return errorMsg{error: "Client or endpoint not initialized"}
		}

		response, err := m.client.ListEndpoints(m.ctx, alexaapimodels.EndpointQuery{
			IncludeFields: &alexaapimodels.EndpointIncludeFields{
				Properties: true,
			},
		})
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to load state: %v", err)}
		}

		// Find the endpoint in the response
		var endpointState *alexaapimodels.Endpoint

		for _, ep := range response.Results {
			if ep.ID == m.selectedEndpoint.ID {
				endpointState = ep

				break
			}
		}

		if endpointState == nil {
			return errorMsg{error: "Endpoint state not found"}
		}

		return stateLoadedMsg{response: response}
	}
}

// Messages.
type tokenValidatedMsg struct {
	client alexa.ClientInterface
}

type endpointsLoadedMsg struct {
	endpoints []*alexaapimodels.Endpoint
}

type endpointSelectedMsg struct {
	endpoint *alexaapimodels.Endpoint
}

type backToEndpointsMsg struct{}

type viewStateMsg struct{}

type stateLoadedMsg struct {
	response *alexaapimodels.UnifiedEndpointListResponse
}

type eventReceivedMsg struct {
	event *alexaapimodels.Event
}

type controlExecutedMsg struct {
	success bool
	message string
}

type errorMsg struct {
	error string
}

func clearErrorAfterDelay() tea.Cmd {
	return tea.Tick(clearErrorDelay, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type clearErrorMsg struct{}
