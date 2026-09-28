package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type stateModel struct {
	endpoint *alexaapimodels.Endpoint
	state    *alexaapimodels.Endpoint
	width    int
	height   int
	loading  bool
}

func newStateModel() stateModel {
	return stateModel{
		loading: false,
	}
}

func (m stateModel) Init() tea.Cmd {
	return nil
}

func (m stateModel) Update(msg tea.Msg) (stateModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case endpointSelectedMsg:
		m.endpoint = msg.endpoint
		m.loading = true
		return m, nil

	case stateLoadedMsg:
		// Find the endpoint in the response
		for _, ep := range msg.response.Results {
			if ep.ID == m.endpoint.ID {
				m.state = ep
				break
			}
		}
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Refresh state
			m.loading = true
			return m, func() tea.Msg {
				return refreshStateMsg{}
			}

		case "b", "esc":
			return m, func() tea.Msg {
				return backToEndpointsMsg{}
			}

		case "c":
			// Go to control view
			return m, func() tea.Msg {
				return endpointSelectedMsg{endpoint: m.endpoint}
			}
		}
	}

	return m, cmd
}

func (m stateModel) View() string {
	if m.endpoint == nil {
		return "No endpoint selected"
	}

	view := titleStyle.Render(fmt.Sprintf("State: %s", m.endpoint.EndpointID)) + "\n"
	view += helpStyle.Render(fmt.Sprintf("Device: %s", m.endpoint.DeviceType)) + "\n\n"

	if m.loading {
		return view + "Loading state..."
	}

	if m.state == nil {
		return view + "No state data available. Press 'r' to refresh."
	}

	// Display features with states (filter features that have properties)
	featuresWithStates := make([]alexaapimodels.Feature, 0)
	for _, feature := range m.state.Features {
		if len(feature.Properties) > 0 {
			featuresWithStates = append(featuresWithStates, feature)
		}
	}

	if len(featuresWithStates) == 0 {
		view += "No features with states found.\n"
	} else {
		for _, feature := range featuresWithStates {
			view += lipgloss.NewStyle().Bold(true).Render(string(feature.Name)) + "\n"

			if len(feature.Properties) == 0 {
				view += "  No properties\n"
			} else {
				for _, prop := range feature.Properties {
					view += fmt.Sprintf("  %s: ", prop.Name)

					// Display state value
					if prop.StateValue != nil {
						view += fmt.Sprintf("%v", prop.StateValue)
					} else {
						view += "N/A"
					}

					// Display accuracy if available
					if prop.Accuracy != "" {
						view += fmt.Sprintf(" (accuracy: %s)", prop.Accuracy)
					}

					// Display error if present
					if prop.Error != nil {
						view += " " + errorStyle.Render(fmt.Sprintf("[ERROR: %s]", prop.Error.Type))
					}

					view += "\n"
				}
			}
			view += "\n"
		}
	}

	view += "\n" + helpStyle.Render("r: Refresh | b: Back | c: Control | q: Quit")
	return view
}

type refreshStateMsg struct{}
