package alexa_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
	replay "github.com/portpowered/go-alexa/pkg/testing"
	"golang.org/x/net/http2"
)

// Exercise the injected connection beneath the real HTTP/2 codec, including
// multiplexed HEADERS/DATA and the directive application's line framing.
func TestEventNativeHTTP2PairedReplay(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-events-http2.json")

	player, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	clientPipe, serverPipe := net.Pipe()

	t.Cleanup(func() { _ = clientPipe.Close(); _ = serverPipe.Close() })

	pinged := make(chan struct{}, 1)
	serverDone := make(chan struct{})
	options := new(http2.ServeConnOpts)
	options.Handler = nativeEventReplayHandler(t, player, pinged)
	server := new(http2.Server)

	go func() {
		defer close(serverDone)

		server.ServeConn(serverPipe, options)
	}()

	transport := new(http2.Transport)
	transport.DisableCompression = true
	transport.DialTLSContext = func(_ context.Context, network, address string, _ *tls.Config) (net.Conn, error) {
		if network != "tcp" || address != "events.synthetic.test:443" {
			t.Errorf("unexpected connection target: %s %s", network, address)
		}

		return clientPipe, nil
	}
	t.Cleanup(transport.CloseIdleConnections)

	client, err := alexa.NewClient(alexa.WithEventAuthority("events.synthetic.test"), alexa.WithEventTransport(transport))
	if err != nil {
		t.Fatal(err)
	}

	session, err := client.NewSession(alexa.WithBearerToken("synthetic-event-token"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		closeErr := session.Close()
		if closeErr != nil {
			t.Errorf("close session: %v", closeErr)
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	connection, err := session.ConnectEvents(ctx)
	if err != nil {
		t.Fatal(err)
	}

	event, err := connection.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if event.MessageID != "synthetic-message" || event.EndpointID != "synthetic-endpoint" {
		t.Fatalf("unexpected event: %+v", event)
	}

	select {
	case <-pinged:
	case <-ctx.Done():
		t.Fatal("HTTP/2 keepalive exchange did not complete")
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}

	err = connection.Close()
	if err != nil {
		t.Fatal(err)
	}

	transport.CloseIdleConnections()

	_ = clientPipe.Close()

	select {
	case <-serverDone:
	case <-ctx.Done():
		t.Fatal("HTTP/2 server did not stop after connection closure")
	}
}

func nativeEventReplayHandler(
	t *testing.T,
	player *replay.SyntheticReplay,
	pinged chan<- struct{},
) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.ProtoMajor != 2 {
			t.Errorf("request used %s instead of HTTP/2", request.Proto)
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		replayRequest, ok := normalizeNativeEventReplayRequest(request)
		if !ok {
			t.Errorf("unexpected inbound HTTP/2 request identity")
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		response, err := player.RoundTrip(replayRequest)
		if err != nil {
			t.Errorf("paired HTTP/2 request mismatch: %v", err)
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		defer func() { _ = response.Body.Close() }()

		writer.Header()["Date"] = nil

		for key, values := range response.Header {
			writer.Header()[key] = values
		}

		writer.WriteHeader(response.StatusCode)

		_, err = io.Copy(writer, response.Body)
		if err != nil {
			t.Errorf("write paired HTTP/2 response: %v", err)

			return
		}

		if replayRequest.URL.Path == apiroutes.PathPingDirectiveStream {
			pinged <- struct{}{}

			return
		}

		err = http.NewResponseController(writer).Flush()
		if err != nil {
			t.Errorf("flush directive DATA frames: %v", err)

			return
		}

		<-request.Context().Done()
	})
}

func normalizeNativeEventReplayRequest(request *http.Request) (*http.Request, bool) {
	if !isExpectedNativeEventRequest(request) {
		return nil, false
	}

	replayRequest := request.Clone(request.Context())
	replayRequest.RequestURI = ""
	replayRequest.URL.Scheme = "https"
	replayRequest.URL.Host = replayRequest.Host

	return replayRequest, true
}

func isExpectedNativeEventRequest(request *http.Request) bool {
	if request == nil || request.URL == nil || request.Host != "events.synthetic.test" ||
		request.RequestURI == "" || request.RequestURI != request.URL.RequestURI() {
		return false
	}

	switch request.URL.Path {
	case apiroutes.ChannelDirectivesAddress:
		return request.Method == apiroutes.MethodOpenDirectiveStream
	case apiroutes.PathPingDirectiveStream:
		return request.Method == apiroutes.MethodPingDirectiveStream
	default:
		return false
	}
}
