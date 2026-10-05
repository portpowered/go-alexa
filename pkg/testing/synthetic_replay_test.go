//nolint:testpackage // Exercises private replay matching and exhaustion state.
package testing

import (
	"context"
	"io"
	"net/http"
	"net/url"
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
		req.Host = "service.example.invalid"

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

	if pair.Request.Host != "service.example.invalid" {
		t.Fatalf("request Host override was not recorded: %q", pair.Request.Host)
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

	assertSyntheticReplayRejectsMismatches(t, path, makeRequest)

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

type syntheticReplayMismatchCase struct {
	name   string
	field  string
	mutate func(*http.Request)
}

func assertSyntheticReplayRejectsMismatches(
	t *testing.T,
	fixturePath string,
	makeRequest func() *http.Request,
) {
	t.Helper()

	tests := []syntheticReplayMismatchCase{
		{
			name: "method", field: "method",
			mutate: func(r *http.Request) { r.Method = http.MethodGet },
		},
		{
			name: "origin", field: "origin",
			mutate: func(r *http.Request) { r.URL.Host = "wrong.invalid" },
		},
		{
			name: "effective authority", field: "effective authority",
			mutate: func(r *http.Request) { r.Host = "secret-authority.invalid" },
		},
		{
			name: "escaped path", field: "escaped path",
			mutate: func(r *http.Request) {
				r.URL.RawPath = ""
				r.URL.Path = "/other"
			},
		},
		{
			name: "repeated query", field: "query",
			mutate: func(r *http.Request) { r.URL.RawQuery = "expand=all" },
		},
		{
			name: "malformed query", field: "malformed query",
			mutate: func(r *http.Request) { r.URL.RawQuery = "token=secret-query%" },
		},
		{
			name: "URL user information", field: "URL user information",
			mutate: func(r *http.Request) {
				r.URL.User = url.UserPassword("credential-secret", "password-secret")
			},
		},
		{
			name: "opaque URL", field: "opaque URL",
			mutate: func(r *http.Request) { r.URL.Opaque = "//opaque-secret.invalid/private" },
		},
		{
			name: "header", field: "headers",
			mutate: func(r *http.Request) { r.Header.Set("Authorization", "Bearer secret-header") },
		},
		{
			name: "body", field: "body",
			mutate: func(r *http.Request) {
				r.Body = io.NopCloser(strings.NewReader(`{"id":"secret-body"}`))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertSyntheticReplayRejectsMismatch(t, fixturePath, makeRequest, test)
		})
	}
}

func assertSyntheticReplayRejectsMismatch(
	t *testing.T,
	fixturePath string,
	makeRequest func() *http.Request,
	test syntheticReplayMismatchCase,
) {
	t.Helper()

	player, err := LoadSyntheticReplay(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	req := makeRequest()
	test.mutate(req)

	response, err := player.RoundTrip(req)
	assertSyntheticReplayMismatchDiagnostic(t, response, err, test.field)
	assertSyntheticReplayPairStillAvailable(t, player, makeRequest)
}

func assertSyntheticReplayMismatchDiagnostic(t *testing.T, response *http.Response, err error, field string) {
	t.Helper()

	if response != nil {
		closeErr := response.Body.Close()
		if closeErr != nil {
			t.Errorf("close mismatched response body: %v", closeErr)
		}
	}

	if err == nil || response != nil {
		t.Fatalf("mismatched request returned response %v, error %v", response, err)
	}

	for _, secret := range []string{
		"synthetic-token",
		"synthetic-id",
		"secret-authority.invalid",
		"secret-query",
		"credential-secret",
		"password-secret",
		"opaque-secret.invalid",
		"secret-header",
		"secret-body",
	} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("mismatch diagnostic leaked sensitive value %q: %v", secret, err)
		}
	}

	if !strings.Contains(err.Error(), field) {
		t.Fatalf("mismatch diagnostic lacks field %q: %v", field, err)
	}
}

func assertSyntheticReplayPairStillAvailable(t *testing.T, player *SyntheticReplay, makeRequest func() *http.Request) {
	t.Helper()

	err := player.AssertConsumed()
	if err == nil {
		t.Fatal("mismatched request consumed the expected pair")
	}

	validResponse, err := player.RoundTrip(makeRequest())
	if err != nil {
		t.Fatalf("valid fixture request failed after mismatch: %v", err)
	}

	closeErr := validResponse.Body.Close()
	if closeErr != nil {
		t.Errorf("close valid fixture response body: %v", closeErr)
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}
}
