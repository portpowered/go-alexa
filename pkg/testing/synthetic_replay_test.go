//nolint:testpackage // Exercises private replay matching and exhaustion state.
package testing

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyntheticReplayRejectsMismatchAndExhaustion(t *testing.T) {
	t.Parallel()

	makeRequest := func() *http.Request {
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"https://example.invalid/a%2Fb?expand=all&expand=power",
			strings.NewReader(`{"id":"synthetic-id"}`),
		)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer synthetic-token")

		return req
	}

	pair, err := RecordSyntheticExchange(
		"syntheticOperation",
		makeRequest(),
		http.StatusCreated,
		http.Header{"Content-Type": []string{"application/json"}},
		`{"ok":true}`,
	)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "synthetic.json")
	{
		err := WriteSyntheticReplay(path, []SyntheticExchange{pair})
		if err != nil {
			t.Fatal(err)
		}
	}

	player, err := LoadSyntheticReplay(path)
	if err != nil {
		t.Fatal(err)
	}

	{
		err := player.AssertConsumed()
		if err == nil {
			t.Fatal("unconsumed pair was accepted")
		}
	}

	assertSyntheticReplayRejectsMismatches(t, player, makeRequest)

	response, err := player.RoundTrip(makeRequest())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		closeErr := response.Body.Close()
		if closeErr != nil {
			t.Errorf("close synthetic replay response body: %v", closeErr)
		}
	})

	if response.StatusCode != http.StatusCreated || response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected paired response: %#v", response)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil || string(data) != `{"ok":true}` {
		t.Fatalf("paired body = %q, %v", data, err)
	}

	{
		err := player.AssertConsumed()
		if err != nil {
			t.Fatal(err)
		}
	}

	duplicateResponse, err := player.RoundTrip(makeRequest())
	if duplicateResponse != nil {
		closeErr := duplicateResponse.Body.Close()
		if closeErr != nil {
			t.Errorf("close duplicate synthetic response body: %v", closeErr)
		}
	}

	if err == nil {
		t.Fatal("duplicate request received a response")
	}
}

func assertSyntheticReplayRejectsMismatches(
	t *testing.T,
	player *SyntheticReplay,
	makeRequest func() *http.Request,
) {
	t.Helper()

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
			t.Parallel()

			req := makeRequest()
			test.mutate(req)

			response, err := player.RoundTrip(req)
			if response != nil {
				closeErr := response.Body.Close()
				if closeErr != nil {
					t.Errorf("close mismatched response body: %v", closeErr)
				}
			}

			if err == nil {
				t.Fatal("mismatched request received a response")
			}
		})
	}
}
