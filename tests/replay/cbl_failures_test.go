package replay_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

func TestSyntheticCBLTypedFailures(t *testing.T) {
	t.Parallel()

	player, err := replay.LoadSyntheticReplay("fixtures/synthetic/alexa-cbl-errors.json")
	if err != nil {
		t.Fatal(err)
	}

	client, err := alexa.NewClient(alexa.WithHTTPClient(&http.Client{Transport: player}))
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		target any
	}{
		{"pending", new(*alexaapimodels.CodePairPendingError)},
		{"expired", new(*alexaapimodels.CodePairExpiredError)},
		{"invalid-pair", new(*alexaapimodels.CodePairExpiredError)},
		{"denied", new(*alexaapimodels.CodePairDeniedError)},
		{"invalid-device", new(*alexaapimodels.CodePairInvalidDeviceError)},
		{"duplicate-name", new(*alexaapimodels.CodePairDuplicateDeviceNameError)},
		{"invalid-name", new(*alexaapimodels.CodePairInvalidDeviceError)},
		{"prefixed-serial", new(*alexaapimodels.CodePairInvalidDeviceError)},
		{"unknown", new(*alexaapimodels.CodePairRegistrationError)},
		{"malformed", new(*alexaapimodels.CodePairRegistrationError)},
		{"generation-invalid-device", new(*alexaapimodels.CodePairInvalidDeviceError)},
		{"generation-rejected", new(*alexaapimodels.CodePairGenerationError)},
	}
	for _, test := range cases {
		config := alexaapimodels.DefaultDeviceRegistrationConfig("ABCDEF1234567", "go-alexa CLI")
		if test.name == "invalid-name" {
			config.DeviceName = ""
		}

		if test.name == "prefixed-serial" {
			config.DeviceSerial = "go-alexa-0123456789abcdef0123456789abcdef"
		}

		if strings.HasPrefix(test.name, "generation-") {
			_, err = client.GenerateCodePair(context.Background(), config)
		} else {
			_, err = client.RegisterWithCodePair(context.Background(), "PUBLIC", "synthetic-private", config)
		}

		if err == nil || !errors.As(fmt.Errorf("caller wrapped: %w", err), test.target) {
			t.Fatalf("%s: expected typed failure %T, got %T: %v", test.name, test.target, err, err)
		}

		for _, secret := range []string{"synthetic-private", "unknown_private_details", "not JSON"} {
			if strings.Contains(err.Error(), secret) {
				t.Fatalf("%s exposed provider diagnostics", test.name)
			}
		}

		assertCBLFailureStatus(t, test.target)
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}
}

func assertCBLFailureStatus(t *testing.T, target any) {
	t.Helper()

	switch typed := target.(type) {
	case **alexaapimodels.CodePairDuplicateDeviceNameError:
		if (*typed).StatusCode != http.StatusBadRequest {
			t.Fatal("duplicate device name lost HTTP status")
		}
	case **alexaapimodels.CodePairInvalidDeviceError:
		if (*typed).StatusCode != http.StatusBadRequest {
			t.Fatal("invalid device lost HTTP status")
		}
	case **alexaapimodels.CodePairRegistrationError:
		if (*typed).StatusCode != http.StatusBadRequest && (*typed).StatusCode != http.StatusUnauthorized {
			t.Fatal("registration rejection lost HTTP status")
		}
	case **alexaapimodels.CodePairGenerationError:
		if (*typed).StatusCode != http.StatusUnauthorized {
			t.Fatal("generation rejection lost HTTP status")
		}
	}
}
