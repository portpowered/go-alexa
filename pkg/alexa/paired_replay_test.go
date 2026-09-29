//nolint:testpackage // Exercises private event transport state using paired synthetic replay.
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

const syntheticPingPath = "/ping"

func syntheticDirectiveFrame(t *testing.T) string {
	t.Helper()

	metadata := `{"metricName":"EndpointPower","payload":{"data":{"features":[{"name":"power","properties":[{"name":"powerState","powerStateValue":"OFF"}]}]},` +
		`"entity":{"id":"synthetic-endpoint"}}}`
	directive := map[string]any{
		"directive": map[string]any{
			"header": map[string]any{
				"messageId": "synthetic-message",
				"namespace": "Alexa.Mobile.Push",
				"name":      "RenderUpdate",
			},
			"payload": map[string]any{"renderingUpdates": []any{map[string]any{"resourceMetadata": metadata}}},
		},
	}

	body, err := json.Marshal(directive)
	if err != nil {
		t.Fatal(err)
	}

	return "------synthetic-boundary\nContent-Type: application/json\n\n" + string(body) + "\n"
}

func TestEventPairedSyntheticReplay(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-events.json")
	frame := syntheticDirectiveFrame(t)
	responses := eventReplayResponses(frame)
	pinged := make(chan struct{}, 1)
	transport, player, capture := eventReplayTransport(t, fixture, responses, pinged)
	conn := newSessionHTTP2Connection(
		"events.synthetic.test",
		func(context.Context) (string, error) { return "synthetic-event-token", nil },
		&http.Client{Transport: transport},
	)

	t.Cleanup(func() { closeTestHTTP2Connection(t, conn) })

	err := conn.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	awaitInitialEventPing(t, pinged)

	gotEvent := false
	for range 2 {
		gotEvent = gotEvent || receiveSyntheticDirective(t, conn)
	}

	if !gotEvent {
		t.Fatal("no decoded synthetic directive event")
	}

	if player != nil {
		assertEventReplayPlayer(t, fixture, player)

		return
	}

	writeEventReplayFixture(t, fixture, responses, capture.exchanges)
}

type eventReplayResponse struct {
	operation string
	status    int
	body      string
}

type eventReplayCapture struct {
	exchanges []replay.SyntheticExchange
}

func eventReplayResponses(frame string) []eventReplayResponse {
	return []eventReplayResponse{
		{operation: "openDirectiveStream", status: http.StatusOK, body: frame},
		{operation: "pingDirectiveStream", status: http.StatusNoContent, body: ""},
	}
}

func eventReplayTransport(
	t *testing.T,
	fixture string,
	responses []eventReplayResponse,
	pinged chan<- struct{},
) (http.RoundTripper, *replay.SyntheticReplay, *eventReplayCapture) {
	t.Helper()

	if os.Getenv("UPDATE_SYNTHETIC_REPLAY") == "1" {
		capture := &eventReplayCapture{exchanges: make([]replay.SyntheticExchange, 0, len(responses))}

		return eventRecordingTransport(responses, capture, pinged), nil, capture
	}

	player, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	return eventPlayerTransport(player, pinged), player, nil
}

func eventRecordingTransport(
	responses []eventReplayResponse,
	capture *eventReplayCapture,
	pinged chan<- struct{},
) http.RoundTripper {
	return syntheticEventTransport(func(req *http.Request) (*http.Response, error) {
		if len(capture.exchanges) >= len(responses) {
			// Report the extra request count during fixture generation.
			//nolint:err113
			return nil, fmt.Errorf("unexpected event request after %d exchanges", len(capture.exchanges))
		}

		current := responses[len(capture.exchanges)]

		responseHeaders := make(http.Header)
		if current.operation == "openDirectiveStream" {
			responseHeaders.Set("Content-Type", "multipart/mixed; boundary=--synthetic-boundary")
		}

		pair, err := replay.RecordSyntheticExchange(
			current.operation,
			req,
			current.status,
			responseHeaders,
			current.body,
		)
		if err != nil {
			return nil, fmt.Errorf("record synthetic event exchange: %w", err)
		}

		capture.exchanges = append(capture.exchanges, pair)

		if current.operation == "pingDirectiveStream" {
			pinged <- struct{}{}
		}

		response := eventResponse(current.status, current.body)
		response.Header = responseHeaders

		return response, nil
	})
}

func eventPlayerTransport(player *replay.SyntheticReplay, pinged chan<- struct{}) http.RoundTripper {
	return syntheticEventTransport(func(req *http.Request) (*http.Response, error) {
		response, err := player.RoundTrip(req)
		if err == nil && req.URL.Path == syntheticPingPath {
			pinged <- struct{}{}
		}

		if err != nil {
			return nil, fmt.Errorf("replay synthetic event exchange: %w", err)
		}

		return response, nil
	})
}

func awaitInitialEventPing(t *testing.T, pinged <-chan struct{}) {
	t.Helper()

	select {
	case <-pinged:
	case <-time.After(time.Second):
		t.Fatal("initial ping was not sent")
	}
}

func receiveSyntheticDirective(t *testing.T, conn *HTTP2Connection) bool {
	t.Helper()

	result := make(chan bool, 1)

	go func() {
		event, _ := conn.Receive()
		result <- event != nil && event.MessageID == "synthetic-message" && event.EndpointID == "synthetic-endpoint"
	}()

	select {
	case valid := <-result:
		return valid
	case <-time.After(time.Second):
		t.Fatal("ordered directive frame was not consumed")

		return false
	}
}

func assertEventReplayPlayer(t *testing.T, fixture string, player *replay.SyntheticReplay) {
	t.Helper()

	err := player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}

	ping, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://events.synthetic.test/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	ping.Header.Set("Authorization", "Bearer synthetic-event-token")
	assertEventRequestRejected(t, player, ping, "duplicate event ping received a replay response")

	outOfOrder, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	assertEventRequestRejected(t, outOfOrder, ping, "out-of-order event ping received a replay response")
	assertOutOfOrderEventReplayIsIncomplete(t, outOfOrder)
}

func assertOutOfOrderEventReplayIsIncomplete(t *testing.T, player *replay.SyntheticReplay) {
	t.Helper()

	err := player.AssertConsumed()
	if err == nil {
		t.Fatal("out-of-order event script was marked consumed")
	}
}

func assertEventRequestRejected(t *testing.T, player *replay.SyntheticReplay, req *http.Request, failureMessage string) {
	t.Helper()

	response, err := player.RoundTrip(req)
	if response != nil {
		closeErr := response.Body.Close()
		if closeErr != nil {
			t.Errorf("close rejected event response body: %v", closeErr)
		}
	}

	if err == nil {
		t.Fatal(failureMessage)
	}
}

func writeEventReplayFixture(
	t *testing.T,
	fixture string,
	responses []eventReplayResponse,
	exchanges []replay.SyntheticExchange,
) {
	t.Helper()

	if len(exchanges) != len(responses) {
		t.Fatalf("recorded %d of %d", len(exchanges), len(responses))
	}

	err := os.MkdirAll(filepath.Dir(fixture), 0o750)
	if err != nil {
		t.Fatal(err)
	}

	err = replay.WriteSyntheticReplay(fixture, exchanges)
	if err != nil {
		t.Fatal(err)
	}
}
