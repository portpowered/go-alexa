package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type endpointsModel struct {
	endpoints    []*alexaapimodels.Endpoint
	allEndpoints []*alexaapimodels.Endpoint // Store all endpoints for filtering
	cursor       int
	startIdx     int // Starting index for pagination
	width        int
	height       int
	loading      bool
	searching    bool
	searchQuery  string
}

const maxVisibleEndpoints = 20

func newEndpointsModel() endpointsModel {
	return endpointsModel{
		endpoints:    nil,
		allEndpoints: nil,
		cursor:       0,
		startIdx:     0,
		width:        0,
		height:       0,
		loading:      false,
		searching:    false,
		searchQuery:  "",
	}
}

func (m *endpointsModel) Init() tea.Cmd {
	return nil
}

func (m *endpointsModel) Update(msg tea.Msg) (endpointsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case endpointsLoadedMsg:
		return updateLoadedEndpoints(m, msg)

	case tea.KeyMsg:
		if m.searching {
			return updateEndpointSearch(m, msg)
		}

		return updateEndpointSelection(m, msg)
	}

	return *m, nil
}

func (m *endpointsModel) View() string {
	if m.loading {
		return "Loading endpoints..."
	}

	view := titleStyle.Render("Endpoints") + "\n\n"

	// Show search bar if searching
	if m.searching {
		searchBar := "/" + m.searchQuery
		view += helpStyle.Render(searchBar) + "\n\n"
	} else if m.searchQuery != "" {
		searchInfo := fmt.Sprintf("Filter: %s (Press '/' to modify)", m.searchQuery)
		view += helpStyle.Render(searchInfo) + "\n\n"
	}

	if len(m.endpoints) == 0 {
		if m.searchQuery != "" {
			return view + "No endpoints match your search. Press Esc to clear filter."
		}

		return view + "No endpoints found. Press 'r' to refresh."
	}

	// Calculate visible range
	endIdx := m.startIdx + maxVisibleEndpoints
	if endIdx > len(m.endpoints) {
		endIdx = len(m.endpoints)
	}

	// Show only visible endpoints
	for endpointIndex := m.startIdx; endpointIndex < endIdx; endpointIndex++ {
		endpoint := m.endpoints[endpointIndex]

		cursor := " "
		if endpointIndex == m.cursor {
			cursor = ">"
		}

		// Get endpoint name (use FriendlyName if available, otherwise fallback to EndpointID)
		endpointName := getEndpointDisplayName(endpoint)

		endpointLine := fmt.Sprintf("%s %s (%s)", cursor, endpointName, endpoint.EndpointID)
		if endpointIndex == m.cursor {
			endpointLine = selectedStyle.Render(endpointLine)
		} else {
			endpointLine = lipgloss.NewStyle().Render(endpointLine)
		}

		view += endpointLine + "\n"

		// Show supported features if selected
		if endpointIndex == m.cursor && len(endpoint.Features) > 0 {
			features := "  Features: "

			for j, feat := range endpoint.Features {
				if j > 0 {
					features += ", "
				}

				features += string(feat.Name)
			}

			view += helpStyle.Render(features) + "\n"
		}
	}

	// Show pagination info if there are more endpoints
	paginationInfo := ""
	if len(m.endpoints) > maxVisibleEndpoints {
		paginationInfo = fmt.Sprintf("\nShowing %d-%d of %d endpoints", m.startIdx+1, endIdx, len(m.endpoints))
	}

	view += paginationInfo + "\n"

	helpText := "↑/↓: Navigate | Enter: Select | r: Refresh | s: View State"
	if m.searching {
		helpText += " | Esc: Exit search"
	} else {
		helpText += " | /: Search"
	}

	helpText += " | q: Quit"

	view += helpStyle.Render(helpText)

	return view
}

// applySearchFilter filters endpoints based on the search query.
func (m *endpointsModel) applySearchFilter() {
	if m.searchQuery == "" {
		m.endpoints = m.allEndpoints
	} else {
		query := strings.ToLower(m.searchQuery)
		filtered := []*alexaapimodels.Endpoint{}

		for _, endpoint := range m.allEndpoints {
			name := strings.ToLower(getEndpointDisplayName(endpoint))

			endpointID := strings.ToLower(endpoint.EndpointID)

			if strings.Contains(name, query) || strings.Contains(endpointID, query) {
				filtered = append(filtered, endpoint)
			}
		}

		m.endpoints = filtered
	}

	// Ensure cursor is within bounds after filtering
	if m.cursor >= len(m.endpoints) {
		if len(m.endpoints) > 0 {
			m.cursor = len(m.endpoints) - 1
		} else {
			m.cursor = 0
		}
	}

	if m.startIdx >= len(m.endpoints) {
		m.startIdx = 0
	}
}

type refreshEndpointsMsg struct{}

// getEndpointDisplayName returns the display name for an endpoint
// Uses FriendlyName if available, otherwise falls back to EndpointID.
func getEndpointDisplayName(endpoint *alexaapimodels.Endpoint) string {
	if endpoint.FriendlyName != nil && endpoint.FriendlyName.Value != "" {
		return endpoint.FriendlyName.Value
	}

	return endpoint.EndpointID
}

func updateLoadedEndpoints(model *endpointsModel, msg endpointsLoadedMsg) (endpointsModel, tea.Cmd) {
	model.allEndpoints = msg.endpoints
	sort.Slice(model.allEndpoints, func(i, j int) bool {
		nameI := getEndpointDisplayName(model.allEndpoints[i])
		nameJ := getEndpointDisplayName(model.allEndpoints[j])

		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})
	model.applySearchFilter()
	model.loading = false
	model.startIdx = 0

	if len(model.endpoints) > 0 && model.cursor >= len(model.endpoints) {
		model.cursor = len(model.endpoints) - 1
	}

	if model.cursor >= model.startIdx+maxVisibleEndpoints {
		model.startIdx = model.cursor - maxVisibleEndpoints + 1
	}

	return *model, nil
}

func updateEndpointSearch(model *endpointsModel, msg tea.KeyMsg) (endpointsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		model.searching = false
		model.searchQuery = ""
		model.applySearchFilter()
		model.cursor = 0
		model.startIdx = 0
	case "backspace":
		if len(model.searchQuery) > 0 {
			model.searchQuery = model.searchQuery[:len(model.searchQuery)-1]
			model.applySearchFilter()
			model.cursor = 0
			model.startIdx = 0
		}
	case "enter":
		model.searching = false
	default:
		if len(msg.Runes) > 0 {
			model.searchQuery += string(msg.Runes)
			model.applySearchFilter()
			model.cursor = 0
			model.startIdx = 0
		}
	}

	return *model, nil
}

func updateEndpointSelection(model *endpointsModel, msg tea.KeyMsg) (endpointsModel, tea.Cmd) {
	switch msg.String() {
	case "/":
		model.searching = true
	case "up", "k":
		if model.cursor > 0 {
			model.cursor--
			if model.cursor < model.startIdx {
				model.startIdx = model.cursor
			}
		}
	case "down", "j":
		if model.cursor < len(model.endpoints)-1 {
			model.cursor++
			if model.cursor >= model.startIdx+maxVisibleEndpoints {
				model.startIdx = model.cursor - maxVisibleEndpoints + 1
			}
		}
	case "enter", "s":
		return *model, selectedEndpointCommand(model)
	case "r":
		model.loading = true

		return *model, func() tea.Msg { return refreshEndpointsMsg{} }
	}

	return *model, nil
}

func selectedEndpointCommand(model *endpointsModel) tea.Cmd {
	if len(model.endpoints) == 0 || model.cursor >= len(model.endpoints) {
		return nil
	}

	return func() tea.Msg {
		return endpointSelectedMsg{endpoint: model.endpoints[model.cursor]}
	}
}
