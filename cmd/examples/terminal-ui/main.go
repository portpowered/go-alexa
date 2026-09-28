package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/portpowered/go-alexa/cmd/examples/terminal-ui/tui"
)

func main() {
	// Check for token as command-line argument or environment variable
	if len(os.Args) > 1 {
		if err := os.Setenv("ALEXA_ACCESS_TOKEN", "x"); err != nil {
			os.Exit(1)
		}
	}

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
