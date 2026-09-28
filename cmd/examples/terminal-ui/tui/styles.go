package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("219")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

func splitView(left, right string, width int) string {
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1 // Account for separator

	leftBox := borderStyle.
		Width(leftWidth).
		Height(0). // Auto height
		Render(left)

	rightBox := borderStyle.
		Width(rightWidth).
		Height(0). // Auto height
		Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)
}
