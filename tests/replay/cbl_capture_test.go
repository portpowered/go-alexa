package replay_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

const cblCaptureDirectory = "fixtures/captured/cbl-login-20261010"

func TestCapturedCBLLogin(t *testing.T) {
	t.Parallel()

	player, err := replay.NewReplayCaptureRoundTripper(cblCaptureDirectory)
	if err != nil {
		t.Fatal(err)
	}

	client, err := alexa.NewClient(alexa.WithHTTPClient(&http.Client{Transport: player}))
	if err != nil {
		t.Fatal(err)
	}

	config := alexaapimodels.DefaultDeviceRegistrationConfig("ABCDEF1234567", "go-alexa CLI")

	pair, err := client.GenerateCodePair(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	registration, err := client.RegisterWithCodePair(context.Background(), pair.PublicCode, pair.PrivateCode, config)
	if err != nil {
		t.Fatal(err)
	}

	if registration.RefreshToken != "synthetic-refresh" {
		t.Fatal("registration lost the reusable token")
	}

	token, err := client.RefreshAccessToken(context.Background(), alexaapimodels.TokenRefreshRequest{
		RefreshToken: registration.RefreshToken, Config: config,
	})
	if err != nil {
		t.Fatal(err)
	}

	if token.AccessToken != "synthetic-access" || token.ExpiresInSeconds <= 0 {
		t.Fatal("refresh lost the access token or expiry")
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCapturedCBLDeviceMetadataIsConstrained(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, file, section, field, value string }{
		{"prefixed serial", "capture_0001.json", "code_data", "device_serial", "go-alexa-0123456789abcdef0123456789abcdef"},
		{"short serial", "capture_0001.json", "code_data", "device_serial", "ABCDEF"},
		{"lowercase serial", "capture_0001.json", "code_data", "device_serial", "abcdef1234567"},
		{"empty name", "capture_0001.json", "code_data", "device_name", ""},
		{"changed name", "capture_0001.json", "code_data", "device_name", "unsupported device name"},
		{"extra name parameter", "capture_0001.json", "code_data", "deviceName", "unsupported"},
		{"changed registration serial", "capture_0002.json", "registration_data", "device_serial", "ZYXWVU7654321"},
		{"changed registration name", "capture_0002.json", "registration_data", "device_name", "unsupported device name"},
		{"changed device type", "capture_0002.json", "registration_data", "device_type", "unsupported"},
		{"changed refresh serial", "capture_0003.json", "device_metadata", "device_serial", "ZYXWVU7654321"},
		{"extra refresh name", "capture_0003.json", "device_metadata", "device_name", "go-alexa CLI"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			pair := readCBLCapture(t, test.file)

			var bodySection map[string]json.RawMessage
			// Decode only the selected metadata object; other top-level fields include arrays.
			var root map[string]json.RawMessage

			err := json.Unmarshal(pair.Request.Body, &root)
			if err != nil {
				t.Fatal(err)
			}

			err = json.Unmarshal(root[test.section], &bodySection)
			if err != nil {
				t.Fatal(err)
			}

			value, err := json.Marshal(test.value)
			if err != nil {
				t.Fatal(err)
			}

			bodySection[test.field] = value

			root[test.section], err = json.Marshal(bodySection)
			if err != nil {
				t.Fatal(err)
			}

			mutated, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}

			assertCBLReplayRejects(t, test.file, mutated)
		})
	}
}

func readCBLCapture(t *testing.T, name string) replay.CapturePair {
	t.Helper()

	//nolint:gosec // Fixed fixture directory and filenames supplied exclusively by this test.
	data, err := os.ReadFile(filepath.Join(cblCaptureDirectory, name))
	if err != nil {
		t.Fatal(err)
	}

	var pair replay.CapturePair

	err = json.Unmarshal(data, &pair)
	if err != nil {
		t.Fatal(err)
	}

	return pair
}

func assertCBLReplayRejects(t *testing.T, target string, mutated []byte) {
	t.Helper()

	player, err := replay.NewReplayCaptureRoundTripper(cblCaptureDirectory)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"capture_0001.json", "capture_0002.json", "capture_0003.json"} {
		pair := readCBLCapture(t, name)

		request, err := http.NewRequestWithContext(context.Background(), pair.Request.Method, pair.Request.URL, bytes.NewReader(pair.Request.Body))
		if err != nil {
			t.Fatal(err)
		}

		request.Header = pair.Request.Headers.Clone()
		if name == target {
			request.Body = io.NopCloser(bytes.NewReader(mutated))

			response, rejectErr := player.RoundTrip(request)
			if response != nil {
				_ = response.Body.Close()

				t.Fatal("replay returned a response for changed device metadata")
			}

			if rejectErr == nil {
				t.Fatal("replay accepted changed device metadata")
			}

			err := player.AssertConsumed()
			if err == nil {
				t.Fatal("mismatch consumed captured responses")
			}

			request.Body = io.NopCloser(bytes.NewReader(pair.Request.Body))
		}

		response, err := player.RoundTrip(request)
		if err != nil {
			t.Fatal(err)
		}

		_ = response.Body.Close()
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}
}
