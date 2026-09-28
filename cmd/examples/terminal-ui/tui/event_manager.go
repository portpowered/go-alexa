package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type eventManager struct {
	client    alexa.ClientInterface
	conn      alexa.Connection
	ctx       context.Context
	connected bool
}

func newEventManager(client alexa.ClientInterface, ctx context.Context) *eventManager {
	return &eventManager{
		client:    client,
		ctx:       ctx,
		connected: false,
	}
}

func (em *eventManager) connect(eventChan chan<- *alexaapimodels.Event) tea.Cmd {
	return func() tea.Msg {
		if em.client == nil {
			return eventConnectionErrorMsg{err: fmt.Errorf("client not initialized")}
		}

		conn, err := em.client.ConnectEvents(em.ctx)
		if err != nil {
			return eventConnectionErrorMsg{err: fmt.Errorf("failed to connect: %w", err)}
		}

		em.conn = conn
		em.connected = true

		// Subscribe to events
		subscribeReq := alexaapimodels.SubscribeRequest{
			DurationInMinutes: 60,
		}

		_, err = em.client.Subscribe(em.ctx, subscribeReq)
		if err != nil {
			_ = em.conn.Close()
			em.connected = false
			return eventConnectionErrorMsg{err: fmt.Errorf("failed to subscribe: %w", err)}
		}

		// Start receiving events
		go em.receiveEvents(eventChan)

		return eventConnectedMsg{}
	}
}

func (em *eventManager) receiveEvents(eventChan chan<- *alexaapimodels.Event) {
	for {
		if em.conn == nil {
			return
		}

		event, err := em.conn.Receive()
		if err != nil {
			em.connected = false
			return
		}

		// Send event to channel
		select {
		case eventChan <- event:
		case <-em.ctx.Done():
			return
		default:
			// Channel full, skip this event
		}
	}
}

// Commands for event management
func connectEventsCmd(client alexa.ClientInterface, ctx context.Context, eventChan chan<- *alexaapimodels.Event) tea.Cmd {
	return func() tea.Msg {
		em := newEventManager(client, ctx)
		return em.connect(eventChan)
	}
}

// Command to listen for events from channel
func listenForEvents(eventChan <-chan *alexaapimodels.Event) tea.Cmd {
	return func() tea.Msg {
		event := <-eventChan
		return eventReceivedMsg{event: event}
	}
}

type eventConnectedMsg struct{}

type eventConnectionErrorMsg struct {
	err error
}

// Note: The actual event receiving needs to be integrated with bubbletea's message system
// This requires a different approach - we'll need to use a channel or similar mechanism
// to send events from the goroutine to the bubbletea program
