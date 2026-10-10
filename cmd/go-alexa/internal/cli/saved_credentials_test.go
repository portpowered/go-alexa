//nolint:testpackage // Verify private credential storage and source selection without exporting secrets or test-only APIs.
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
	"regexp"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestLoginPersistsCBLCredentialsAndReusesThem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "credentials.json")
	t.Setenv("GO_ALEXA_CREDENTIALS", path)
	t.Setenv("ALEXA_ACCESS_TOKEN", "")
	t.Setenv("ALEXA_REFRESH_TOKEN", "")

	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++

		response.Header().Set("Content-Type", "application/json")

		switch request.URL.Path {
		case "/auth/create/codepair":
			_, _ = io.WriteString(response, `{"public_code":"PUBLIC","private_code":"PRIVATE"}`)
		case "/auth/register":
			var payload map[string]json.RawMessage

			err := json.NewDecoder(request.Body).Decode(&payload)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Contains(payload["auth_data"], []byte(`"code_pair"`)) ||
				bytes.Contains(payload["auth_data"], []byte(`"email_password"`)) {
				t.Errorf("registration must use CBL: %s", payload["auth_data"])
			}

			_, _ = io.WriteString(response, `{"response":{"success":{"tokens":{"bearer":{"access_token":"ACCESS","refresh_token":"REFRESH"}}}}}`)
		case "/auth/token":
			_, _ = io.WriteString(response, `{"access_token":"ACCESS","expires_in":3600}`)
		default:
			t.Errorf("unexpected request: %s", request.URL.Path)
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	var output bytes.Buffer

	app := New(nil, &output, io.Discard, alexa.WithAmazonAPIBaseURL(server.URL))
	for range 2 {
		app.input = strings.NewReader("\n")

		err := app.Run(context.Background(), []string{"auth", "login"})
		if err != nil {
			t.Fatal(err)
		}
	}

	if requests != 6 {
		t.Errorf("got %d requests, want code generation, registration and refresh per login", requests)
	}

	for _, secret := range []string{"PRIVATE", "ACCESS", "REFRESH"} {
		if strings.Contains(output.String(), secret) {
			t.Error("login output disclosed a secret")
		}
	}

	credentials, err := app.readCredentials(&credentialFlags{file: "", stdin: false})
	if err != nil || credentials.AccessToken != "ACCESS" || credentials.RefreshToken != "REFRESH" ||
		credentials.DeviceRegistration.DeviceSerial == "" {
		t.Fatalf("saved credentials could not be reused: %v", err)
	}

	if !regexp.MustCompile(`^[A-Z0-9]{13}$`).MatchString(credentials.DeviceRegistration.DeviceSerial) {
		t.Fatal("generated serial does not match the pre-extraction CBL example")
	}

	assertNoCredentialTemporaryFiles(t, filepath.Dir(path))
}

func TestRefreshUpdatesSavedLogin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	t.Setenv("GO_ALEXA_CREDENTIALS", path)
	t.Setenv("ALEXA_ACCESS_TOKEN", "")
	t.Setenv("ALEXA_REFRESH_TOKEN", "")

	credentials := credentialFile{AccessToken: "OLD", RefreshToken: "REFRESH",
		DeviceRegistration: alexaapimodels.DefaultDeviceRegistrationConfig("SERIAL", "Test device")}

	err := saveCredentials(path, credentials)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/auth/token" {
			t.Errorf("unexpected request: %s", request.URL.Path)
		}

		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"access_token":"NEW","expires_in":3600}`)
	}))
	defer server.Close()

	app := New(nil, io.Discard, io.Discard, alexa.WithAmazonAPIBaseURL(server.URL))

	err = app.Run(context.Background(), []string{"auth", "refresh"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := app.readCredentials(&credentialFlags{file: "", stdin: false})
	if err != nil || got.AccessToken != "NEW" || got.RefreshToken != "REFRESH" ||
		got.DeviceRegistration.DeviceSerial != "SERIAL" {
		t.Fatalf("refresh did not preserve the saved registration: %v", err)
	}

	err = app.Run(context.Background(), []string{"auth", "logout"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatal("logout did not remove the default credential file")
	}
}

func TestSavedCredentialSourcesAndDefaultPath(t *testing.T) {
	t.Parallel()

	app := New(strings.NewReader(`{"accessToken":"STDIN"}`), io.Discard, io.Discard)
	app.lookupEnv = func(key string) (string, bool) {
		if key == "ALEXA_ACCESS_TOKEN" {
			return "ENVIRONMENT", true
		}

		return "", false
	}

	directory, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := app.defaultCredentialPath()
	if err != nil || path != filepath.Join(directory, "go-alexa", "credentials.json") {
		t.Fatalf("default credential path = %q, %v", path, err)
	}

	got, err := app.readCredentials(&credentialFlags{file: "", stdin: false})
	if err != nil || got.AccessToken != "ENVIRONMENT" {
		t.Fatal("environment credentials were not preferred over saved login")
	}

	got, err = app.readCredentials(&credentialFlags{file: "", stdin: true})
	if err != nil || got.AccessToken != "STDIN" {
		t.Fatal("explicit stdin credentials were not preferred over environment")
	}
}

func TestCredentialReplacementRejectsDirectory(t *testing.T) {
	t.Parallel()

	path := t.TempDir()

	err := replaceCredentials(path, credentialFile{AccessToken: "NEW", RefreshToken: "",
		DeviceRegistration: alexaapimodels.DefaultDeviceRegistrationConfig("", "")})
	if err == nil {
		t.Fatal("replacement accepted a directory")
	}

	assertNoCredentialTemporaryFiles(t, filepath.Dir(path))
}

func TestFailedLoginPreservesSavedCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	t.Setenv("GO_ALEXA_CREDENTIALS", path)

	previous := credentialFile{AccessToken: "OLD", RefreshToken: "OLD-REFRESH",
		DeviceRegistration: alexaapimodels.DefaultDeviceRegistrationConfig("OLD-SERIAL", "Test")}

	err := saveCredentials(path, previous)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")

		if request.URL.Path == "/auth/create/codepair" {
			_, _ = io.WriteString(response, `{"public_code":"PUBLIC","private_code":"PRIVATE"}`)

			return
		}

		response.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(response, `{"error":"access_denied"}`)
	}))
	defer server.Close()

	app := New(strings.NewReader("\n"), io.Discard, io.Discard, alexa.WithAmazonAPIBaseURL(server.URL))

	err = app.Run(context.Background(), []string{"auth", "login"})
	if err == nil {
		t.Fatal("login accepted an Amazon rejection")
	}

	got, err := app.readCredentials(&credentialFlags{file: path, stdin: false})
	if err != nil || got.AccessToken != previous.AccessToken || got.RefreshToken != previous.RefreshToken ||
		got.DeviceRegistration.DeviceSerial != previous.DeviceRegistration.DeviceSerial {
		t.Fatalf("failed login changed saved credentials: %v", err)
	}

	assertNoCredentialTemporaryFiles(t, filepath.Dir(path))
}

func assertNoCredentialTemporaryFiles(t *testing.T, directory string) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(directory, "credentials.json.*.tmp"))
	if err != nil || len(files) != 0 {
		t.Fatalf("credential temporary files remain: %v, %v", files, err)
	}
}
