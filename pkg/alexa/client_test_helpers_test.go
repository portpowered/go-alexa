//nolint:testpackage // Shared test helpers require private client and session state.
package alexa

import (
	"net/http"
	"testing"
)

func newTestSession(t *testing.T, httpClient *http.Client, options ...SessionOption) *Session {
	t.Helper()

	client, err := NewClient(WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	session, err := client.NewSession(options...)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	t.Cleanup(func() { closeTestSession(t, session) })

	return session
}

func closeTestSession(t *testing.T, session *Session) {
	t.Helper()

	closeErr := session.Close()
	if closeErr != nil {
		t.Errorf("close test session: %v", closeErr)
	}
}

func closeTestHTTP2Connection(t *testing.T, connection *HTTP2Connection) {
	t.Helper()

	closeErr := connection.Close()
	if closeErr != nil {
		t.Errorf("close test HTTP/2 connection: %v", closeErr)
	}
}
