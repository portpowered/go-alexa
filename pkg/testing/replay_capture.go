package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

// ReplayCaptureRoundTripper is an http.RoundTripper that replays captured request/response pairs
// from a directory of capture files.
type ReplayCaptureRoundTripper struct {
	dirPath  string
	captures []CapturePair
	mu       sync.Mutex
	next     int
}

var _ http.RoundTripper = (*ReplayCaptureRoundTripper)(nil)

// NewReplayCaptureRoundTripper creates a new ReplayCaptureRoundTripper that loads
// captured request/response pairs from the specified directory and replays them.
func NewReplayCaptureRoundTripper(dirPath string) (*ReplayCaptureRoundTripper, error) {
	replayRoundTripper := &ReplayCaptureRoundTripper{
		dirPath:  dirPath,
		captures: make([]CapturePair, 0),
		mu:       sync.Mutex{},
		next:     0,
	}

	err := replayRoundTripper.loadCaptures()
	if err != nil {
		return nil, fmt.Errorf("failed to load captures: %w", err)
	}

	return replayRoundTripper, nil
}

// RoundTrip returns only the response paired with the next exact captured request.
func (t *ReplayCaptureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte

	if req.Body != nil {
		var err error

		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read replay request: %w", err)
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.next >= len(t.captures) {
		return nil, fmt.Errorf("%w %d captures consumed", errReplayRequestsExhausted, len(t.captures))
	}

	pair := &t.captures[t.next]
	if req.Method != pair.Request.Method || req.URL.String() != pair.Request.URL ||
		!reflect.DeepEqual(req.Header, pair.Request.Headers) || !bytes.Equal(body, pair.Request.Body) {
		return nil, fmt.Errorf("%w %d mismatch in method, URL, headers, or body", errReplayRequestMismatch, t.next+1)
	}

	t.next++

	return t.buildResponse(req, &pair.Response), nil
}

// AssertConsumed reports unconsumed request/response pairs.
func (t *ReplayCaptureRoundTripper) AssertConsumed() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.next != len(t.captures) {
		return fmt.Errorf("%w %d of %d request/response pairs", errReplayPairsUnconsumed, t.next, len(t.captures))
	}

	return nil
}

// loadCaptures loads all capture files from the directory.
func (t *ReplayCaptureRoundTripper) loadCaptures() error {
	entries, err := os.ReadDir(t.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read capture directory: %w", err)
	}

	captures := make([]CapturePair, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		if !strings.HasPrefix(entry.Name(), "capture_") {
			continue
		}

		filePath := filepath.Join(t.dirPath, entry.Name())

		//nolint:gosec // The path is a child filename returned by ReadDir.
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read capture file %s: %w", entry.Name(), err)
		}

		var pair CapturePair
		{
			err := json.Unmarshal(data, &pair)
			if err != nil {
				return fmt.Errorf("failed to parse capture file %s: %w", entry.Name(), err)
			}
		}

		captures = append(captures, pair)
	}

	t.mu.Lock()
	t.captures = captures
	t.mu.Unlock()

	return nil
}

// buildResponse constructs an http.Response from a captured response.
func (t *ReplayCaptureRoundTripper) buildResponse(req *http.Request, capturedResp *CapturedResponse) *http.Response {
	resp := &http.Response{
		Status:     capturedResp.Status,
		StatusCode: capturedResp.StatusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(capturedResp.Body)),
		Request:    req,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	// Copy headers from captured response
	for key, values := range capturedResp.Headers {
		for _, value := range values {
			resp.Header.Add(key, value)
		}
	}

	// Set ContentLength based on body size
	resp.ContentLength = int64(len(capturedResp.Body))

	return resp
}
