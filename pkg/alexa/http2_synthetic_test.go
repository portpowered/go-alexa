//nolint:testpackage // Exercises private HTTP/2 lifecycle and stream-state transitions.
package alexa

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type syntheticEventTransport func(*http.Request) (*http.Response, error)

const syntheticBearerToken = "synthetic-token"

func (f syntheticEventTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func eventResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// Every edge response here is hand-authored; no account traffic is recorded.
func TestSyntheticEventStreamTransport(t *testing.T) {
	t.Parallel()

	var directives, pings int

	pinged := make(chan struct{}, 1)
	transport := syntheticEventTransport(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "Bearer "+syntheticBearerToken {
			t.Errorf("authorization header = %q", req.Header.Get("Authorization"))
		}

		switch req.URL.Path {
		case "/v20160207/directives":
			directives++

			if req.URL.Host != "events.example.test" {
				t.Errorf("event authority = %q", req.URL.Host)
			}

			return eventResponse(
				http.StatusOK,
				"------synthetic-boundary\nContent-Type: application/json\n\nnot-json\n",
			), nil
		case "/ping":
			pings++

			pinged <- struct{}{}

			return eventResponse(http.StatusNoContent, ""), nil
		default:
			t.Errorf("unexpected path %q", req.URL.Path)

			return eventResponse(http.StatusNotFound, ""), nil
		}
	})
	client := &http.Client{Transport: transport}

	conn := newSessionHTTP2Connection(
		"events.example.test",
		func(context.Context) (string, error) { return syntheticBearerToken, nil },
		client,
	)

	if conn.client != client || conn.transport != nil {
		t.Fatal("event HTTP client was not injected")
	}

	err := conn.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	{
		_, err := conn.Receive()
		if err == nil || !alexaapimodels.IsConnectionError(err) {
			t.Fatalf("stream EOF = %v", err)
		}
	}

	select {
	case <-pinged:
	case <-time.After(time.Second):
		t.Fatal("initial ping did not use injected transport")
	}

	if directives != 1 || pings != 1 {
		t.Fatalf("requests: directives=%d pings=%d", directives, pings)
	}

	assertClosedEventConnection(t, conn)
}

func assertClosedEventConnection(t *testing.T, conn *HTTP2Connection) {
	t.Helper()

	err := conn.Send(nil)

	if err == nil || !alexaapimodels.IsConnectionError(err) {
		t.Fatalf("unsupported send = %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	{
		_, err := conn.Receive()
		if !alexaapimodels.IsClosedError(err) {
			t.Fatalf("receive after close = %v", err)
		}
	}

	err = conn.Send(nil)

	if !alexaapimodels.IsClosedError(err) {
		t.Fatalf("send after close = %v", err)
	}

	err = conn.Connect(context.Background())

	if !alexaapimodels.IsClosedError(err) {
		t.Fatalf("connect after close = %v", err)
	}
}

func TestSyntheticEventStreamErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		status int
		want   func(error) bool
	}{
		{"unauthorized", http.StatusUnauthorized, alexaapimodels.IsAuthenticationError},
		{"forbidden", http.StatusForbidden, alexaapimodels.IsAuthenticationError},
		{"server", http.StatusServiceUnavailable, alexaapimodels.IsHTTPError},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			conn := newSessionHTTP2Connection(
				"events.example.test",
				func(context.Context) (string, error) { return syntheticBearerToken, nil },
				&http.Client{Transport: syntheticEventTransport(func(*http.Request) (*http.Response, error) {
					return eventResponse(test.status, "synthetic failure"), nil
				})},
			)

			t.Cleanup(func() { closeTestHTTP2Connection(t, conn) })

			err := conn.Connect(context.Background())

			if !test.want(err) {
				t.Fatalf("Connect error = %v", err)
			}
		})
	}

	conn := newSessionHTTP2Connection(
		"events.example.test",
		func(context.Context) (string, error) { return "", staticError("synthetic token failure") },
		&http.Client{Transport: syntheticEventTransport(func(*http.Request) (*http.Response, error) {
			t.Fatal("unexpected request")

			return nil, staticError("unexpected event stream request")
		})},
	)

	err := conn.Connect(context.Background())
	if !alexaapimodels.IsTokenError(err) {
		t.Fatalf("token getter error = %v", err)
	}

	closeTestHTTP2Connection(t, conn)

	conn = newSessionHTTP2Connection("events.example.test", nil, nil)
	{
		_, err := conn.getToken(context.Background())
		if !alexaapimodels.IsTokenError(err) {
			t.Fatalf("missing token = %v", err)
		}
	}

	conn.bearerToken = "synthetic-fixed-token"
	{
		token, err := conn.getToken(context.Background())
		if err != nil || token != "synthetic-fixed-token" {
			t.Fatalf("fixed token = %q, %v", token, err)
		}
	}

	closeTestHTTP2Connection(t, conn)
}

func TestSyntheticEventPingFailures(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		status int
		want   func(error) bool
	}{
		{http.StatusForbidden, alexaapimodels.IsAuthenticationError},
		{http.StatusInternalServerError, alexaapimodels.IsPingError},
	} {
		conn := newSessionHTTP2Connection(
			"events.example.test",
			nil,
			&http.Client{Transport: syntheticEventTransport(func(*http.Request) (*http.Response, error) {
				return eventResponse(test.status, "synthetic failure"), nil
			})},
		)

		err := conn.ping(context.Background(), syntheticBearerToken)
		if !test.want(err) {
			t.Fatalf("ping %d = %v", test.status, err)
		}

		closeTestHTTP2Connection(t, conn)
	}

	conn := newSessionHTTP2Connection(
		"events.example.test",
		nil,
		&http.Client{
			Transport: syntheticEventTransport(
				func(*http.Request) (*http.Response, error) { return nil, staticError("synthetic network failure") },
			),
		},
	)

	err := conn.ping(context.Background(), syntheticBearerToken)
	if !alexaapimodels.IsPingError(err) {
		t.Fatalf("network ping = %v", err)
	}

	closeTestHTTP2Connection(t, conn)
}

func TestSyntheticEventStreamCloseUnblocksLiveRead(t *testing.T) {
	t.Parallel()

	reader, writer := io.Pipe()

	t.Cleanup(func() {
		closeErr := writer.Close()
		if closeErr != nil {
			t.Errorf("close event stream pipe writer: %v", closeErr)
		}
	})

	conn := newSessionHTTP2Connection(
		"events.example.test",
		func(context.Context) (string, error) { return syntheticBearerToken, nil },
		&http.Client{Transport: syntheticEventTransport(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/ping" {
				return eventResponse(http.StatusNoContent, ""), nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       reader,
				Header:     make(http.Header),
			}, nil
		})},
	)

	err := conn.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)

	go func() { done <- conn.Close() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		// Release the blocked reader even when Close is broken, so the test can
		// report the lifecycle failure instead of hanging the package suite.
		closeErr := writer.Close()
		if closeErr != nil {
			t.Errorf("release blocked event stream reader: %v", closeErr)
		}

		<-done
		t.Fatal("Close did not unblock an idle event stream")
	}
}

func TestEventErrorsApplyBackpressureWithoutDroppingTerminalFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := &HTTP2Connection{ctx: ctx, errChan: make(chan error, 1)}
	conn.errChan <- staticError("synthetic prior error")

	terminal := staticError("synthetic terminal error")
	done := make(chan struct{})

	go func() {
		conn.reportStreamError(terminal)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("terminal failure was dropped when the error queue was full")
	case <-time.After(20 * time.Millisecond):
	}

	<-conn.errChan

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("terminal failure did not enter the queue after backpressure cleared")
	}

	got := <-conn.errChan

	if !errors.Is(got, terminal) {
		t.Fatalf("got %v, want terminal failure", got)
	}
}
