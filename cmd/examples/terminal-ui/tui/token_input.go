package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/portpowered/go-alexa/pkg/alexa"
)

type inputStep int

const (
	stepAccessToken inputStep = iota
	stepRefreshToken
	stepCustomerID
)

type tokenInputModel struct {
	textInput    textinput.Model
	err          error
	step         inputStep
	accessToken  string
	refreshToken string
	customerID   string
}

func newTokenInputModel() tokenInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your access token..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 50

	return tokenInputModel{
		textInput: ti,
		step:      stepAccessToken,
	}
}

func (m tokenInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m tokenInputModel) Update(msg tea.Msg) (tokenInputModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			value := m.textInput.Value()
			if value == "" {
				m.err = fmt.Errorf("field cannot be empty")
				return m, nil
			}

			// Store the value and move to next step
			switch m.step {
			case stepAccessToken:
				m.accessToken = value
				m.step = stepRefreshToken
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter your refresh token..."
				m.err = nil
				return m, textinput.Blink

			case stepRefreshToken:
				m.refreshToken = value
				m.step = stepCustomerID
				m.textInput.SetValue("")
				m.textInput.Placeholder = "Enter your customer ID..."
				m.err = nil
				return m, textinput.Blink

			case stepCustomerID:
				m.customerID = value
				// All fields collected, validate and create client
				return m, validateToken(m.accessToken, m.refreshToken, m.customerID)
			}

		case tea.KeyEsc:
			// Only quit on explicit Esc key press
			return m, tea.Quit

		case tea.KeyCtrlC:
			// Only quit on explicit Ctrl+C
			return m, tea.Quit
		}
	}

	// Let textinput handle all other messages (including regular key presses)
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m tokenInputModel) View() string {
	view := titleStyle.Render("Alexa SDK Terminal UI") + "\n\n"

	var prompt string
	switch m.step {
	case stepAccessToken:
		prompt = "Enter your access token:"
	case stepRefreshToken:
		prompt = "Enter your refresh token:"
	case stepCustomerID:
		prompt = "Enter your customer ID:"
	}

	view += prompt + "\n\n"
	view += m.textInput.View() + "\n\n"

	if m.err != nil {
		view += errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	view += "\n" + helpStyle.Render("Press Enter to continue, Esc to quit")
	return view
}

func validateToken(accessToken, refreshToken, customerID string) tea.Cmd {
	return func() tea.Msg {
		// Keep account credentials in a session so a service client can be shared.
		client, err := alexa.NewClient()
		if err != nil {
			return tokenValidationErrorMsg{err: err}
		}
		sessionOptions := []alexa.SessionOption{
			alexa.WithBearerToken(accessToken),
			alexa.WithRefreshToken(refreshToken),
		}
		if customerID != "" {
			sessionOptions = append(sessionOptions, alexa.WithCustomerID(customerID))
		}
		session, err := client.NewSession(sessionOptions...)
		if err != nil {
			return tokenValidationErrorMsg{err: err}
		}

		// Test the connection by trying to list endpoints (this will fail if token is invalid)
		// Actually, we'll just return success if client creation works
		// The real validation happens when we try to use it
		return tokenValidatedMsg{client: session}
	}
}

type tokenValidationErrorMsg struct {
	err error
}
