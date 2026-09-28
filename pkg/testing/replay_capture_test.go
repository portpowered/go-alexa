package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestReplayCaptureMatchesWholeRequestBeforeReturningResponse(t *testing.T) {
	dir := t.TempDir()
	for i, pair := range []CapturePair{
		{
			Request:  CapturedRequest{Method: http.MethodPost, URL: "https://example.invalid/one?a=1&a=2", Headers: http.Header{"X-Test": {"first"}}, Body: []byte(`{"id":1}`)},
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
		if err := os.WriteFile(filepath.Join(dir, []string{"capture_0001.json", "capture_0002.json"}[i]), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	replay, err := NewReplayCaptureRoundTripper(dir)
	if err != nil {
		t.Fatal(err)
	}
	request := func(target, body, header string) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, target, bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Test", header)
		return req
	}
	for _, bad := range []*http.Request{
		request("https://example.invalid/one?a=1", `{"id":1}`, "first"),
		request("https://example.invalid/one?a=1&a=2", `{"id":1}`, "wrong"),
		request("https://example.invalid/one?a=1&a=2", `{"id":2}`, "first"),
	} {
		if response, err := replay.RoundTrip(bad); err == nil || response != nil {
			t.Fatalf("mismatched request returned response %v, error %v", response, err)
		}
	}
	if err := replay.AssertConsumed(); err == nil {
		t.Fatal("unconsumed capture was not detected")
	}
	response, err := replay.RoundTrip(request("https://example.invalid/one?a=1&a=2", `{"id":1}`, "first"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusCreated || string(body) != `{"ok":true}` {
		t.Fatalf("response status=%d body=%q error=%v", response.StatusCode, body, err)
	}
	second, err := http.NewRequest(http.MethodGet, "https://example.invalid/two", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := replay.RoundTrip(second); err != nil {
		t.Fatal(err)
	}
	if err := replay.AssertConsumed(); err != nil {
		t.Fatal(err)
	}
	if response, err := replay.RoundTrip(second); err == nil || response != nil {
		t.Fatalf("duplicate request returned response %v, error %v", response, err)
	}
}
