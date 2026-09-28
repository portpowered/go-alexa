package testing

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyntheticReplayRejectsMismatchAndExhaustion(t *testing.T) {
	makeRequest := func() *http.Request {
		req, err := http.NewRequest(http.MethodPost, "https://example.invalid/a%2Fb?expand=all&expand=power", strings.NewReader(`{"id":"synthetic-id"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer synthetic-token")
		return req
	}
	pair, err := RecordSyntheticExchange("syntheticOperation", makeRequest(), http.StatusCreated, http.Header{"Content-Type": []string{"application/json"}}, `{"ok":true}`)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "synthetic.json")
	if err := WriteSyntheticReplay(path, []SyntheticExchange{pair}); err != nil {
		t.Fatal(err)
	}
	player, err := LoadSyntheticReplay(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := player.AssertConsumed(); err == nil {
		t.Fatal("unconsumed pair was accepted")
	}
	for _, test := range []struct {
		name   string
		mutate func(*http.Request)
	}{
		{"method", func(r *http.Request) { r.Method = http.MethodGet }},
		{"origin", func(r *http.Request) { r.URL.Host = "wrong.invalid" }},
		{"escaped path", func(r *http.Request) { r.URL.RawPath = ""; r.URL.Path = "/other" }},
		{"repeated query", func(r *http.Request) { r.URL.RawQuery = "expand=all" }},
		{"header", func(r *http.Request) { r.Header.Set("Authorization", "Bearer wrong") }},
		{"body", func(r *http.Request) { r.Body = io.NopCloser(strings.NewReader(`{"id":"wrong"}`)) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := makeRequest()
			test.mutate(req)
			if _, err := player.RoundTrip(req); err == nil {
				t.Fatal("mismatched request received a response")
			}
		})
	}
	response, err := player.RoundTrip(makeRequest())
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated || response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected paired response: %#v", response)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil || string(data) != `{"ok":true}` {
		t.Fatalf("paired body = %q, %v", data, err)
	}
	if err := player.AssertConsumed(); err != nil {
		t.Fatal(err)
	}
	if _, err := player.RoundTrip(makeRequest()); err == nil {
		t.Fatal("duplicate request received a response")
	}
}
