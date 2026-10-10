package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

// wireRecorder writes the SDK's existing captured replay format. The CLI owns
// private-file permissions; these raw traces are never synthetic fixtures.
type wireRecorder struct {
	transport http.RoundTripper
	directory string
	counter   atomic.Uint64
}

func (a *App) newRecordedClient(region, directory string) (*alexa.Client, error) {
	err := a.enableWireRecording(directory)
	if err != nil {
		return nil, err
	}

	return a.newClient(region)
}

func (a *App) enableWireRecording(directory string) error {
	if directory == "" {
		directory, _ = a.lookupEnv("GO_ALEXA_RECORD_DIR")
	}

	if directory == "" {
		return nil
	}

	suffix, err := randomDeviceSerial()
	if err != nil {
		return err
	}

	runDirectory := filepath.Join(directory, time.Now().UTC().Format("20060102T150405Z")+"-"+suffix)

	err = os.MkdirAll(runDirectory, credentialDirectoryMode)
	if err != nil {
		return fmt.Errorf("create wire capture directory: %w", err)
	}

	err = savePrivateJSON(filepath.Join(runDirectory, "provenance.json"), map[string]any{
		"source": "captured", "capturedAt": time.Now().UTC().Format(time.RFC3339),
		"format": "go-alexa CapturePair", "sanitized": false,
	})
	if err != nil {
		return err
	}

	a.recorder = &wireRecorder{transport: http.DefaultTransport, directory: runDirectory, counter: atomic.Uint64{}}

	_, err = fmt.Fprintf(a.errorOutput, "Private wire captures: %s\n", runDirectory)
	if err != nil {
		return fmt.Errorf("report wire capture directory: %w", err)
	}

	return nil
}

func (r *wireRecorder) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := copyWireBody(request.Body)
	if err != nil {
		return nil, fmt.Errorf("read wire request: %w", err)
	}

	if request.Body != nil {
		request.Body = io.NopCloser(bytes.NewReader(body))
	}

	pair := replay.CapturePair{
		Request:  replay.CapturedRequest{Method: request.Method, URL: request.URL.String(), Headers: request.Header.Clone(), Body: body},
		Response: replay.CapturedResponse{StatusCode: 0, Status: "", Headers: nil, Body: nil},
	}

	response, transportErr := r.transport.RoundTrip(request)
	if response != nil {
		responseBody, readErr := copyWireBody(response.Body)
		if readErr != nil {
			return nil, fmt.Errorf("read wire response: %w", readErr)
		}

		response.Body = io.NopCloser(bytes.NewReader(responseBody))
		pair.Response = replay.CapturedResponse{
			StatusCode: response.StatusCode, Status: response.Status, Headers: response.Header.Clone(), Body: responseBody,
		}
	}

	path := filepath.Join(r.directory, fmt.Sprintf("capture_%04d.json", r.counter.Add(1)))

	err = savePrivateJSON(path, pair)
	if err != nil {
		if response != nil {
			_ = response.Body.Close()
		}

		return nil, fmt.Errorf("save wire capture: %w", err)
	}

	if transportErr != nil {
		return nil, fmt.Errorf("recorded HTTP transport failure: %w", transportErr)
	}

	return response, nil
}

func copyWireBody(body io.ReadCloser) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	defer func() { _ = body.Close() }()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("copy HTTP body: %w", err)
	}

	return data, nil
}
