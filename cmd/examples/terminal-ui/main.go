// Package main starts the terminal Alexa client example.
package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/portpowered/go-alexa/cmd/examples/terminal-ui/tui"
)

func main() {
	// Check for token as command-line argument or environment variable
	if len(os.Args) > 1 {
		err := os.Setenv("ALEXA_ACCESS_TOKEN", "x")
		if err != nil {
			os.Exit(1)
		}
	}

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	{
		_, err := p.Run()
		if err != nil {
			os.Exit(1)
		}
	}
}
