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
		cursor:      0,
		startIdx:    0,
		loading:     false,
		searching:   false,
		searchQuery: "",
	}
}

func (m endpointsModel) Init() tea.Cmd {
	return nil
}

func (m endpointsModel) Update(msg tea.Msg) (endpointsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case endpointsLoadedMsg:
		m.allEndpoints = msg.endpoints
		// Sort endpoints by name (FriendlyName if available, otherwise EndpointID)
		sort.Slice(m.allEndpoints, func(i, j int) bool {
			nameI := getEndpointDisplayName(m.allEndpoints[i])
			nameJ := getEndpointDisplayName(m.allEndpoints[j])
			return strings.ToLower(nameI) < strings.ToLower(nameJ)
		})
		// Apply current search filter if any
		m.applySearchFilter()
		m.loading = false
		m.startIdx = 0
		if len(m.endpoints) > 0 && m.cursor >= len(m.endpoints) {
			m.cursor = len(m.endpoints) - 1
		}
		// Adjust startIdx if cursor is out of visible range
		if m.cursor >= m.startIdx+maxVisibleEndpoints {
			m.startIdx = m.cursor - maxVisibleEndpoints + 1
		}
		return m, nil

	case tea.KeyMsg:
		// Handle search mode
		if m.searching {
			switch msg.String() {
			case "esc":
				// Exit search mode
				m.searching = false
				m.searchQuery = ""
				m.applySearchFilter()
				m.cursor = 0
				m.startIdx = 0
				return m, nil

			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.applySearchFilter()
					m.cursor = 0
					m.startIdx = 0
				}
				return m, nil

			case "enter":
				// Exit search mode but keep filter
				m.searching = false
				return m, nil

			default:
				// Handle text input
				if len(msg.Runes) > 0 {
					m.searchQuery += string(msg.Runes)
					m.applySearchFilter()
					m.cursor = 0
					m.startIdx = 0
				}
				return m, nil
			}
		}

		// Normal mode key handling
		switch msg.String() {
		case "/":
			// Enter search mode (keep existing query if any)
			m.searching = true
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Scroll up if cursor moves above visible window
				if m.cursor < m.startIdx {
					m.startIdx = m.cursor
				}
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.endpoints)-1 {
				m.cursor++
				// Scroll down if cursor moves below visible window
				if m.cursor >= m.startIdx+maxVisibleEndpoints {
					m.startIdx = m.cursor - maxVisibleEndpoints + 1
				}
			}
			return m, nil

		case "enter":
			if len(m.endpoints) > 0 && m.cursor < len(m.endpoints) {
				return m, func() tea.Msg {
					return endpointSelectedMsg{endpoint: m.endpoints[m.cursor]}
				}
			}
			return m, nil

		case "r":
			// Refresh endpoints
			m.loading = true
			return m, func() tea.Msg {
				// This will be handled by the parent model
				return refreshEndpointsMsg{}
			}

		case "s":
			// View state
			if len(m.endpoints) > 0 && m.cursor < len(m.endpoints) {
				return m, func() tea.Msg {
					return endpointSelectedMsg{endpoint: m.endpoints[m.cursor]}
				}
			}
			return m, nil
		}
	}

	return m, cmd
}

func (m endpointsModel) View() string {
	if m.loading {
		return "Loading endpoints..."
	}

	view := titleStyle.Render("Endpoints") + "\n\n"

	// Show search bar if searching
	if m.searching {
		searchBar := fmt.Sprintf("/%s", m.searchQuery)
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
	for i := m.startIdx; i < endIdx; i++ {
		endpoint := m.endpoints[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Get endpoint name (use FriendlyName if available, otherwise fallback to EndpointID)
		endpointName := getEndpointDisplayName(endpoint)

		endpointLine := fmt.Sprintf("%s %s (%s)", cursor, endpointName, endpoint.EndpointID)
		if i == m.cursor {
			endpointLine = selectedStyle.Render(endpointLine)
		} else {
			endpointLine = lipgloss.NewStyle().Render(endpointLine)
		}

		view += endpointLine + "\n"

		// Show supported features if selected
		if i == m.cursor && len(endpoint.Features) > 0 {
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

// applySearchFilter filters endpoints based on the search query
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
// Uses FriendlyName if available, otherwise falls back to EndpointID
func getEndpointDisplayName(endpoint *alexaapimodels.Endpoint) string {
	if endpoint.FriendlyName != nil && endpoint.FriendlyName.Value != "" {
		return endpoint.FriendlyName.Value
	}
	return endpoint.EndpointID
}
