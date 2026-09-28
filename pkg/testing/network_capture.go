package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
)

// CapturePair represents a captured HTTP request/response pair
type CapturePair struct {
	Request  CapturedRequest  `json:"request"`
	Response CapturedResponse `json:"response"`
}

// CapturedRequest represents a captured HTTP request
type CapturedRequest struct {
	Method  string      `json:"method"`
	URL     string      `json:"url"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body,omitempty"`
}

// CapturedResponse represents a captured HTTP response
type CapturedResponse struct {
	StatusCode int         `json:"status_code"`
	Status     string      `json:"status"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body,omitempty"`
}

// NetworkCaptureRoundTripper is an http.RoundTripper that captures request/response pairs
// and writes each pair to a separate file in the specified directory for later replay.
type NetworkCaptureRoundTripper struct {
	transport http.RoundTripper
	dirPath   string
	counter   atomic.Uint64
}

var _ http.RoundTripper = (*NetworkCaptureRoundTripper)(nil)

// NewNetworkCaptureRoundTripper creates a new NetworkCaptureRoundTripper that writes
// each captured request/response pair to a separate file in the specified directory.
// The directory will be created if it doesn't exist.
func NewNetworkCaptureRoundTripper(transport http.RoundTripper, dirPath string) (*NetworkCaptureRoundTripper, error) {
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create capture directory: %w", err)
	}

	return &NetworkCaptureRoundTripper{
		transport: transport,
		dirPath:   dirPath,
	}, nil
}

// RoundTrip executes the HTTP request, captures both the request and response,
// writes them to the capture file, and returns the response.
func (t *NetworkCaptureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Capture request
	capturedReq := t.captureRequest(req)

	// Execute the request
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		// Even on error, try to capture what we have
		_ = t.writeCapture(CapturePair{
			Request:  capturedReq,
			Response: CapturedResponse{},
		})
		return nil, err
	}

	// Capture response
	capturedResp := t.captureResponse(resp)

	// Write the capture pair to a separate file
	_ = t.writeCapture(CapturePair{
		Request:  capturedReq,
		Response: capturedResp,
	})

	// Replace the response body so the caller can read it
	resp.Body = io.NopCloser(bytes.NewReader(capturedResp.Body))

	return resp, nil
}

// captureRequest captures the details of an HTTP request
func (t *NetworkCaptureRoundTripper) captureRequest(req *http.Request) CapturedRequest {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		// Restore the request body for the actual request
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return CapturedRequest{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: req.Header.Clone(),
		Body:    body,
	}
}

// captureResponse captures the details of an HTTP response
func (t *NetworkCaptureRoundTripper) captureResponse(resp *http.Response) CapturedResponse {
	var body []byte
	if resp.Body != nil {
		body, _ = io.ReadAll(resp.Body)
		// Body will be replaced in RoundTrip
	}

	return CapturedResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header.Clone(),
		Body:       body,
	}
}

// writeCapture writes a capture pair to a new file (thread-safe)
func (t *NetworkCaptureRoundTripper) writeCapture(pair CapturePair) error {
	// Generate a unique filename using atomic counter
	counter := t.counter.Add(1)
	filename := fmt.Sprintf("capture_%04d.json", counter)
	filePath := filepath.Join(t.dirPath, filename)

	// Create and write to the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create capture file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON for readability
	if err := encoder.Encode(pair); err != nil {
		return fmt.Errorf("failed to encode capture pair: %w", err)
	}

	return nil
}
