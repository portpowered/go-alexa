package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const maxEvents = 100

type eventsModel struct {
	events     []eventEntry
	scrollPos  int
	width      int
	height     int
	autoScroll bool
}

type eventEntry struct {
	timestamp time.Time
	event     *alexaapimodels.Event
}

func newEventsModel() eventsModel {
	return eventsModel{
		events:     make([]eventEntry, 0),
		autoScroll: true,
	}
}

func (m eventsModel) Init() tea.Cmd {
	return nil
}

func (m eventsModel) Update(msg tea.Msg) (eventsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case eventReceivedMsg:
		entry := eventEntry{
			timestamp: time.Now(),
			event:     msg.event,
		}
		m.events = append(m.events, entry)

		// Keep only last maxEvents
		if len(m.events) > maxEvents {
			m.events = m.events[len(m.events)-maxEvents:]
		}

		// Auto-scroll to bottom if enabled
		if m.autoScroll {
			m.scrollPos = len(m.events) - 1
			if m.scrollPos < 0 {
				m.scrollPos = 0
			}
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scrollPos > 0 {
				m.scrollPos--
				m.autoScroll = false
			}
			return m, nil

		case "down", "j":
			if m.scrollPos < len(m.events)-1 {
				m.scrollPos++
			} else {
				m.autoScroll = true
			}
			return m, nil

		case "g":
			// Go to top
			m.scrollPos = 0
			m.autoScroll = false
			return m, nil

		case "G":
			// Go to bottom
			m.scrollPos = len(m.events) - 1
			if m.scrollPos < 0 {
				m.scrollPos = 0
			}
			m.autoScroll = true
			return m, nil
		}
	}

	return m, cmd
}

func (m eventsModel) View() string {
	view := titleStyle.Render("Events") + "\n"
	view += helpStyle.Render(fmt.Sprintf("Total: %d", len(m.events))) + "\n\n"

	if len(m.events) == 0 {
		view += "No events received yet.\n"
		view += helpStyle.Render("Connect to events stream to see events here.")
		return view
	}

	// Calculate visible range
	visibleHeight := m.height - 5 // Account for header
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	start := m.scrollPos
	if start < 0 {
		start = 0
	}
	if start >= len(m.events) {
		start = len(m.events) - 1
	}

	end := start + visibleHeight
	if end > len(m.events) {
		end = len(m.events)
	}

	// Display events
	for i := start; i < end; i++ {
		entry := m.events[i]
		view += m.formatEvent(entry, i == m.scrollPos)
		view += "\n"
	}

	view += "\n" + helpStyle.Render("↑/↓: Scroll | g/G: Top/Bottom | Auto-scroll: "+fmt.Sprintf("%v", m.autoScroll))
	return view
}

func (m eventsModel) formatEvent(entry eventEntry, selected bool) string {
	timeStr := entry.timestamp.Format("15:04:05")
	event := entry.event

	line := fmt.Sprintf("[%s] %s::%s", timeStr, event.Namespace, event.Name)
	if event.EndpointID != "" {
		line += fmt.Sprintf(" @ %s", event.EndpointID)
	}

	if selected {
		line = selectedStyle.Render(line)
	} else {
		line = lipgloss.NewStyle().Render(line)
	}

	// Add payload summary
	payloadStr := formatPayload(event.Payload)
	if payloadStr != "" {
		line += "\n  " + helpStyle.Render(payloadStr)
	}

	return line
}

func formatPayload(payload interface{}) string {
	if payload == nil {
		return ""
	}

	switch p := payload.(type) {
	case *alexaapimodels.PowerPayload:
		return fmt.Sprintf("Power: %s", p.PowerState)

	case *alexaapimodels.SpeakerPayload:
		result := "Speaker: "
		if p.Volume != nil {
			result += fmt.Sprintf("Volume=%d", *p.Volume)
		}
		if p.Muted != nil {
			result += fmt.Sprintf(" Muted=%v", *p.Muted)
		}
		return result

	case *alexaapimodels.BrightnessPayload:
		if p.Brightness != nil {
			return fmt.Sprintf("Brightness: %d", *p.Brightness)
		}
		return "Brightness: N/A"

	case *alexaapimodels.ColorPayload:
		result := "Color: "
		if p.Hue != nil {
			result += fmt.Sprintf("Hue=%.1f", *p.Hue)
		}
		if p.Saturation != nil {
			result += fmt.Sprintf(" Sat=%.1f", *p.Saturation)
		}
		if p.Brightness != nil {
			result += fmt.Sprintf(" Bright=%.1f", *p.Brightness)
		}
		return result

	case *alexaapimodels.ColorTemperaturePayload:
		return fmt.Sprintf("Color Temp: %dK", p.ColorTemperatureInKelvin)

	case *alexaapimodels.LockPayload:
		return fmt.Sprintf("Lock: %s", p.LockState)

	case *alexaapimodels.TogglePayload:
		return fmt.Sprintf("Toggle: %s", p.ToggleState)

	case *alexaapimodels.ModePayload:
		return fmt.Sprintf("Mode: %s", p.Mode)

	case *alexaapimodels.RangePayload:
		if p.RangeValue != nil {
			return fmt.Sprintf("Range: %.2f", *p.RangeValue)
		}
		return "Range: N/A"

	case *alexaapimodels.PercentagePayload:
		if p.Percentage != nil {
			return fmt.Sprintf("Percentage: %.1f%%", *p.Percentage)
		}
		return "Percentage: N/A"

	case *alexaapimodels.PowerLevelPayload:
		if p.PowerLevel != nil {
			return fmt.Sprintf("Power Level: %d", *p.PowerLevel)
		}
		return "Power Level: N/A"

	default:
		return fmt.Sprintf("Payload: %T", p)
	}
}
