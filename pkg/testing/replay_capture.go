package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// ReplayCaptureRoundTripper is an http.RoundTripper that replays captured request/response pairs
// from a directory of capture files.
type ReplayCaptureRoundTripper struct {
	dirPath    string
	captures   []CapturePair
	capturesMu sync.RWMutex
	counter    int
	counterMu  sync.Mutex
}

var _ http.RoundTripper = (*ReplayCaptureRoundTripper)(nil)

// NewReplayCaptureRoundTripper creates a new ReplayCaptureRoundTripper that loads
// captured request/response pairs from the specified directory and replays them.
func NewReplayCaptureRoundTripper(dirPath string) (*ReplayCaptureRoundTripper, error) {

	rt := &ReplayCaptureRoundTripper{
		dirPath:  dirPath,
		captures: make([]CapturePair, 0),
	}

	if err := rt.loadCaptures(); err != nil {
		return nil, fmt.Errorf("failed to load captures: %w", err)
	}

	return rt, nil
}

// loadCaptures loads all capture files from the directory
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
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read capture file %s: %w", entry.Name(), err)
		}

		var pair CapturePair
		if err := json.Unmarshal(data, &pair); err != nil {
			return fmt.Errorf("failed to parse capture file %s: %w", entry.Name(), err)
		}

		captures = append(captures, pair)
	}

	// Sort captures by filename to maintain order (capture_0001.json, capture_0002.json, etc.)
	sort.Slice(captures, func(i, j int) bool {
		// Extract numbers from filenames for sorting
		// This is a simple approach - if filenames follow the pattern, they'll sort correctly
		return i < j
	})

	t.capturesMu.Lock()
	t.captures = captures
	t.capturesMu.Unlock()

	return nil
}

// RoundTrip matches the incoming request to a captured request and returns the corresponding
// captured response. It uses sequential matching by default (returns the next capture in order).
func (t *ReplayCaptureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t.capturesMu.RLock()
	captures := t.captures
	t.capturesMu.RUnlock()

	if len(captures) == 0 {
		return nil, fmt.Errorf("no captures available for replay")
	}

	// Try to match by URL and method first (smart matching)
	capturedResp := t.findMatchingCapture(req, captures)

	// If no match found, use sequential matching
	if capturedResp == nil {
		t.counterMu.Lock()
		if t.counter >= len(captures) {
			t.counterMu.Unlock()
			return nil, fmt.Errorf("no more captures available for replay (used %d of %d)", t.counter, len(captures))
		}
		idx := t.counter
		t.counter++
		t.counterMu.Unlock()

		capturedResp = &captures[idx].Response
	}

	// Build HTTP response from captured response
	return t.buildResponse(req, capturedResp), nil
}

// findMatchingCapture attempts to find a matching capture by URL and method
func (t *ReplayCaptureRoundTripper) findMatchingCapture(req *http.Request, captures []CapturePair) *CapturedResponse {
	reqURL := req.URL.String()
	reqMethod := req.Method

	for i := range captures {
		capturedReq := captures[i].Request
		if capturedReq.Method == reqMethod && capturedReq.URL == reqURL {
			return &captures[i].Response
		}
	}

	return nil
}

// buildResponse constructs an http.Response from a captured response
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
