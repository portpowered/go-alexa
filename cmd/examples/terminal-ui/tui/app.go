package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
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

type model struct {
	// Client and connection
	client          alexa.ClientInterface
	eventChan       chan *alexaapimodels.Event
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

func NewModel() *model {
	ctx, cancel := context.WithCancel(context.Background())
	eventChan := make(chan *alexaapimodels.Event, 100)
	m := &model{
		ctx:         ctx,
		cancel:      cancel,
		eventChan:   eventChan,
		currentView: viewTokenInput,
		tokenInput:  newTokenInputModel(),
		endpoints:   newEndpointsModel(),
		control:     newControlModel(),
		state:       newStateModel(),
		events:      newEventsModel(),
	}

	// Check for config file in current directory
	if cfg := getConfigFromFile(); cfg != nil {
		// Auto-validate config if provided via file
		go func() {
			// Small delay to let the UI initialize
			time.Sleep(100 * time.Millisecond)
			// This will be handled by the Update method when tokenValidatedMsg is received
		}()
	}
	return m
}

func getConfigFromFile() *config {
	// Look for config.json in the current directory
	configPath := "config.json"

	// Get the absolute path to ensure we're reading from the current working directory
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		// Config file doesn't exist or can't be read - that's okay
		return nil
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Config file exists but is invalid - that's okay, we'll prompt for input
		return nil
	}

	// Validate that all required fields are present
	if cfg.AccessToken == "" || cfg.RefreshToken == "" || cfg.CustomerID == "" {
		return nil
	}

	return &cfg
}

func (m *model) Init() tea.Cmd {
	// If config is provided via file, auto-validate it
	if cfg := getConfigFromFile(); cfg != nil {
		return tea.Batch(m.tokenInput.Init(), validateToken(cfg.AccessToken, cfg.RefreshToken, cfg.CustomerID))
	}
	return m.tokenInput.Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.endpoints.width = msg.Width / 2
		m.endpoints.height = msg.Height
		m.events.width = msg.Width / 2
		m.events.height = msg.Height
		m.control.width = msg.Width / 2
		m.control.height = msg.Height
		m.state.width = msg.Width / 2
		m.state.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == viewTokenInput {
				m.closeSession()
				return m, tea.Quit
			}
			// Allow quitting from main views too
			if msg.String() == "ctrl+c" {
				m.closeSession()
				return m, tea.Quit
			}
		}

	case tokenValidatedMsg:
		m.client = msg.client
		m.control.client = msg.client
		m.control.ctx = m.ctx
		m.currentView = viewEndpoints
		// Load endpoints
		// Also connect to events
		return m, tea.Batch(m.loadEndpoints(), connectEventsCmd(m.client, m.ctx, m.eventChan), listenForEvents(m.eventChan))

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

	case eventReceivedMsg:
		// Add event to events view
		_, cmd := m.events.Update(msg)
		cmds = append(cmds, cmd)
		// Continue listening for more events
		cmds = append(cmds, listenForEvents(m.eventChan))

	case eventConnectedMsg:
		m.eventsConnected = true
		// Continue listening for events
		cmds = append(cmds, listenForEvents(m.eventChan))

	case eventConnectionErrorMsg:
		m.errorMsg = fmt.Sprintf("Event connection error: %v", msg.err)
		cmds = append(cmds, clearErrorAfterDelay())

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
		m.state.state = nil
		// Find the endpoint in the response
		for _, ep := range msg.response.Results {
			if ep.ID == m.selectedEndpoint.ID {
				m.state.state = ep
				break
			}
		}
		return m, nil

	case tokenValidationErrorMsg:
		m.tokenInput.err = msg.err
		return m, nil

	case clearErrorMsg:
		m.errorMsg = ""
		return m, nil
	}

	// Update current view
	switch m.currentView {
	case viewTokenInput:
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		cmds = append(cmds, cmd)

	case viewEndpoints:
		var cmd tea.Cmd
		m.endpoints, cmd = m.endpoints.Update(msg)
		cmds = append(cmds, cmd)

	case viewControl:
		var cmd tea.Cmd
		m.control, cmd = m.control.Update(msg)
		cmds = append(cmds, cmd)
		// Also update events view for display
		_, cmd = m.events.Update(tea.WindowSizeMsg{Width: m.width / 2, Height: m.height})
		cmds = append(cmds, cmd)

	case viewState:
		var cmd tea.Cmd
		m.state, cmd = m.state.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) closeSession() {
	m.cancel()
	if m.client != nil {
		_ = m.client.Close()
	}
}

func (m *model) View() string {
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

func (m *model) loadEndpoints() tea.Cmd {
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

func (m *model) loadState() tea.Cmd {
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

// Messages
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
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type clearErrorMsg struct{}
