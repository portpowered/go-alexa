package alexa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	replay "github.com/portpowered/go-alexa/pkg/testing"
)

func syntheticDirectiveFrame(t *testing.T) string {
	t.Helper()
	metadata := `{"metricName":"EndpointPower","payload":{"data":{"features":[{"name":"power","properties":[{"name":"powerState","powerStateValue":"OFF"}]}]},"entity":{"id":"synthetic-endpoint"}}}`
	directive := map[string]any{"directive": map[string]any{"header": map[string]any{"messageId": "synthetic-message", "namespace": "Alexa.Mobile.Push", "name": "RenderUpdate"}, "payload": map[string]any{"renderingUpdates": []any{map[string]any{"resourceMetadata": metadata}}}}}
	body, err := json.Marshal(directive)
	if err != nil {
		t.Fatal(err)
	}
	return "------synthetic-boundary\nContent-Type: application/json\n\n" + string(body) + "\n"
}

func TestEventPairedSyntheticReplay(t *testing.T) {
	fixture := filepath.Join("..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-events.json")
	frame := syntheticDirectiveFrame(t)
	responses := []struct {
		operation string
		status    int
		body      string
	}{
		{"openDirectiveStream", 200, frame},
		{"pingDirectiveStream", 204, ""},
	}
	var recorded []replay.SyntheticExchange
	var player *replay.SyntheticReplay
	pinged := make(chan struct{}, 1)
	var transport http.RoundTripper
	if os.Getenv("UPDATE_SYNTHETIC_REPLAY") == "1" {
		transport = syntheticEventTransport(func(req *http.Request) (*http.Response, error) {
			if len(recorded) >= len(responses) {
				return nil, fmt.Errorf("unexpected event request after %d exchanges", len(recorded))
			}
			current := responses[len(recorded)]
			responseHeaders := make(http.Header)
			if current.operation == "openDirectiveStream" {
				responseHeaders.Set("Content-Type", "multipart/mixed; boundary=--synthetic-boundary")
			}
			pair, err := replay.RecordSyntheticExchange(current.operation, req, current.status, responseHeaders, current.body)
			if err != nil {
				return nil, err
			}
			recorded = append(recorded, pair)
			if current.operation == "pingDirectiveStream" {
				pinged <- struct{}{}
			}
			response := eventResponse(current.status, current.body)
			response.Header = responseHeaders
			return response, nil
		})
	} else {
		var err error
		player, err = replay.LoadSyntheticReplay(fixture)
		if err != nil {
			t.Fatal(err)
		}
		transport = syntheticEventTransport(func(req *http.Request) (*http.Response, error) {
			resp, err := player.RoundTrip(req)
			if err == nil && req.URL.Path == "/ping" {
				pinged <- struct{}{}
			}
			return resp, err
		})
	}
	conn := newSessionHTTP2Connection("events.synthetic.test", func(context.Context) (string, error) { return "synthetic-event-token", nil }, &http.Client{Transport: transport})
	defer conn.Close()
	if err := conn.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-pinged:
	case <-time.After(time.Second):
		t.Fatal("initial ping was not sent")
	}
	gotEvent := false
	for i := 0; i < 2; i++ {
		result := make(chan bool, 1)
		go func() {
			event, _ := conn.Receive()
			result <- event != nil && event.MessageID == "synthetic-message" && event.EndpointID == "synthetic-endpoint"
		}()
		select {
		case valid := <-result:
			gotEvent = gotEvent || valid
		case <-time.After(time.Second):
			t.Fatal("ordered directive frame was not consumed")
		}
	}
	if !gotEvent {
		t.Fatal("no decoded synthetic directive event")
	}
	if player != nil {
		if err := player.AssertConsumed(); err != nil {
			t.Fatal(err)
		}
		ping, err := http.NewRequest(http.MethodGet, "https://events.synthetic.test/ping", nil)
		if err != nil {
			t.Fatal(err)
		}
		ping.Header.Set("Authorization", "Bearer synthetic-event-token")
		if _, err := player.RoundTrip(ping); err == nil {
			t.Fatal("duplicate event ping received a replay response")
		}
		outOfOrder, err := replay.LoadSyntheticReplay(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := outOfOrder.RoundTrip(ping); err == nil {
			t.Fatal("out-of-order event ping received a replay response")
		}
		if err := outOfOrder.AssertConsumed(); err == nil {
			t.Fatal("out-of-order event script was marked consumed")
		}
	} else {
		if len(recorded) != len(responses) {
			t.Fatalf("recorded %d of %d", len(recorded), len(responses))
		}
		if err := os.MkdirAll(filepath.Dir(fixture), 0755); err != nil {
			t.Fatal(err)
		}
		if err := replay.WriteSyntheticReplay(fixture, recorded); err != nil {
			t.Fatal(err)
		}
	}
}
