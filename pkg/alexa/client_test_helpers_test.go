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
	t.Cleanup(func() { _ = session.Close() })
	return session
}
