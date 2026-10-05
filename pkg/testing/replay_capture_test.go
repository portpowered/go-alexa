//nolint:testpackage // Exercises private replay-capture matching and exhaustion state.
package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplayCaptureMatchesWholeRequestBeforeReturningResponse(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	writeReplayCaptureFixtures(t, dir)

	replay, err := NewReplayCaptureRoundTripper(dir)
	if err != nil {
		t.Fatal(err)
	}

	request := replayTestRequestBuilder(t)
	assertReplayRejectsMismatchedRequests(t, replay, request)

	response, err := replay.RoundTrip(request("https://example.invalid/one?a=1&a=2", `{"id":1}`, "first"))
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusCreated || string(body) != `{"ok":true}` {
		t.Fatalf("response status=%d body=%q error=%v", response.StatusCode, body, err)
	}

	closeErr := response.Body.Close()
	if closeErr != nil {
		t.Errorf("close replay response body: %v", closeErr)
	}

	second, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/two", nil)
	if err != nil {
		t.Fatal(err)
	}

	secondResponse, err := replay.RoundTrip(second)
	if err != nil {
		t.Fatal(err)
	}

	secondCloseErr := secondResponse.Body.Close()
	if secondCloseErr != nil {
		t.Errorf("close second replay response body: %v", secondCloseErr)
	}

	err = replay.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}

	response, err = replay.RoundTrip(second)
	if response != nil {
		closeErr := response.Body.Close()
		if closeErr != nil {
			t.Errorf("close duplicate response body: %v", closeErr)
		}
	}

	if err == nil || response != nil {
		t.Fatalf("duplicate request returned response %v, error %v", response, err)
	}
}

func writeReplayCaptureFixtures(t *testing.T, dir string) {
	t.Helper()

	for pairIndex, pair := range []CapturePair{
		{
			Request: CapturedRequest{
				Method:  http.MethodPost,
				URL:     "https://example.invalid/one?a=1&a=2",
				Headers: http.Header{"X-Test": {"first"}},
				Body:    []byte(`{"id":1}`),
			},
			Response: CapturedResponse{StatusCode: http.StatusCreated, Body: []byte(`{"ok":true}`)},
		},
		{
			Request:  CapturedRequest{Method: http.MethodGet, URL: "https://example.invalid/two", Headers: http.Header{}},
			Response: CapturedResponse{StatusCode: http.StatusNoContent},
		},
	} {
		data, err := json.Marshal(pair)
		if err != nil {
			t.Fatal(err)
		}

		err = os.WriteFile(filepath.Join(dir, []string{"capture_0001.json", "capture_0002.json"}[pairIndex]), data, 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func replayTestRequestBuilder(t *testing.T) func(string, string, string) *http.Request {
	t.Helper()

	return func(target, body, header string) *http.Request {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, target, bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("X-Test", header)

		return req
	}
}

func assertReplayRejectsMismatchedRequests(
	t *testing.T,
	replay *ReplayCaptureRoundTripper,
	request func(string, string, string) *http.Request,
) {
	t.Helper()

	requestURIOverride := request("https://example.invalid/one?a=1&a=2", `{"id":1}`, "first")
	requestURIOverride.RequestURI = "/one?access_token=request-uri-secret"

	fragmentOverride := request("https://example.invalid/one?a=1&a=2", `{"id":1}`, "first")
	fragmentOverride.URL.Fragment = "fragment-secret"

	badRequests := []*http.Request{
		request("https://example.invalid/one?a=1", `{"id":1}`, "first"),
		request("https://example.invalid/one?a=1&a=2", `{"id":1}`, "wrong"),
		request("https://example.invalid/one?a=1&a=2", `{"id":2}`, "first"),
		requestURIOverride,
		fragmentOverride,
	}
	for _, bad := range badRequests {
		response, err := replay.RoundTrip(bad)
		if response != nil {
			closeErr := response.Body.Close()
			if closeErr != nil {
				t.Errorf("close mismatched response body: %v", closeErr)
			}
		}

		if err == nil || response != nil {
			t.Fatalf("mismatched request returned response %v, error %v", response, err)
		}

		assertReplayErrorHidesSensitiveValues(t, err)
	}

	err := replay.AssertConsumed()
	if err == nil {
		t.Fatal("unconsumed capture was not detected")
	}
}

func assertReplayErrorHidesSensitiveValues(t *testing.T, err error) {
	t.Helper()

	for _, secret := range []string{"request-uri-secret", "fragment-secret"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("replay mismatch diagnostic exposed %q: %v", secret, err)
		}
	}
}
