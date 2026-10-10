//nolint:testpackage // Verify private recording errors and file permissions without exposing CLI internals.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

const testCodePairPath = "/auth/create/codepair"

func TestRecordFailedCBLLoginAndReplayThroughSDK(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")

		if request.URL.Path == testCodePairPath {
			_, _ = io.WriteString(response, `{"public_code":"PUBLIC","private_code":"PRIVATE"}`)

			return
		}

		response.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(response, `{"response":{"error":{"code":"InvalidDevice","message":"The device information is invalid."}}}`)
	}))
	t.Cleanup(server.Close)
	baseDirectory := filepath.Join(t.TempDir(), "captures")
	credentialPath := filepath.Join(t.TempDir(), "credentials.json")

	var output, diagnostics bytes.Buffer

	app := New(strings.NewReader("\n"), &output, &diagnostics, alexa.WithAmazonAPIBaseURL(server.URL))
	app.lookupEnv = func(string) (string, bool) { return "", false }

	err := app.Run(context.Background(), []string{
		"auth", "login", "--device-serial", "ABCDEF1234567", "--device-name", "Test device",
		"--record-dir", baseDirectory, "--credentials-out", credentialPath,
	})
	if err == nil {
		t.Fatalf("recording changed the registration failure: %v", err)
	}

	directory := requireCaptureDirectory(t, baseDirectory)
	if !strings.Contains(diagnostics.String(), directory) {
		t.Fatal("CLI did not report the capture directory")
	}

	assertCapturedRejection(t, directory)
	server.Close()
	assertSDKReplaysCBLFailure(t, directory, server.URL)
}

func requireCaptureDirectory(t *testing.T, base string) string {
	t.Helper()

	entries, err := os.ReadDir(base)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("expected one recording directory: %v", err)
	}

	return filepath.Join(base, entries[0].Name())
}

func assertCapturedRejection(t *testing.T, directory string) {
	t.Helper()

	path := filepath.Join(directory, "capture_0002.json")

	//nolint:gosec // The trace is generated beneath this test's private temporary directory.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var pair replay.CapturePair

	err = json.Unmarshal(data, &pair)
	if err != nil || pair.Response.StatusCode != http.StatusBadRequest ||
		!bytes.Contains(pair.Response.Body, []byte("InvalidDevice")) ||
		!bytes.Contains(pair.Request.Body, []byte("ABCDEF1234567")) ||
		!bytes.Contains(pair.Request.Body, []byte("PRIVATE")) {
		t.Fatalf("capture lost registration wire data: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != credentialFileMode) {
		t.Fatal("wire capture does not have private file permissions")
	}
}

func assertSDKReplaysCBLFailure(t *testing.T, directory, baseURL string) {
	t.Helper()

	player, err := replay.NewReplayCaptureRoundTripper(directory)
	if err != nil {
		t.Fatal(err)
	}

	client, err := alexa.NewClient(alexa.WithAmazonAPIBaseURL(baseURL), alexa.WithHTTPClient(&http.Client{
		Transport: player, Timeout: 0, CheckRedirect: nil, Jar: nil,
	}))
	if err != nil {
		t.Fatal(err)
	}

	config := alexaapimodels.DefaultDeviceRegistrationConfig("ABCDEF1234567", "Test device")

	pair, err := client.GenerateCodePair(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.RegisterWithCodePair(context.Background(), pair.PublicCode, pair.PrivateCode, config)
	if err == nil {
		t.Fatalf("replay did not reproduce captured registration failure: %v", err)
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWireRecordingCanBeEnabledByEnvironment(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	app := New(nil, io.Discard, io.Discard)
	app.lookupEnv = func(key string) (string, bool) { return base, key == "GO_ALEXA_RECORD_DIR" }

	err := app.enableWireRecording("")

	if err != nil || app.recorder == nil {
		t.Fatalf("environment did not enable recording: %v", err)
	}

	err = app.enableWireRecording("")
	if err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(base)
	if err != nil || len(entries) != 2 {
		t.Fatal("recording reused a directory instead of preserving both runs")
	}
}

func TestWireCaptureWriteFailureIsReportedWithoutOverwrite(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "capture_0001.json")

	err := savePrivateJSON(path, "existing recording")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, "response")
	}))
	defer server.Close()

	app := New(nil, io.Discard, io.Discard)
	app.recorder = &wireRecorder{transport: http.DefaultTransport, directory: directory, counter: atomic.Uint64{}}

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := app.recorder.RoundTrip(request)
	if response != nil {
		_ = response.Body.Close()
	}

	if err == nil || response != nil || !strings.Contains(err.Error(), "save wire capture") {
		t.Fatalf("capture write failure was silently dropped: %v", err)
	}
	//nolint:gosec // The existing trace is in this test's private temporary directory.
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Contains(data, []byte("existing recording")) {
		t.Fatal("capture overwrote an existing recording")
	}
}
