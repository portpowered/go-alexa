package alexaapimodels_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	alexaapimodels "github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// These values are synthetic contract examples, not Amazon account captures.
func TestSyntheticEndpointModel(t *testing.T) {
	t.Parallel()

	endpoint := &alexaapimodels.Endpoint{
		EndpointID:         "endpoint-1",
		DeviceType:         "type-1",
		DeviceSerialNumber: "serial-1",
		DeviceFamily:       "family-1",
		DeviceAccountId:    "account-1",
		Features:           []alexaapimodels.Feature{{Name: alexaapimodels.FeatureNamePower}},
	}
	if endpoint.GetEndpointId() != "endpoint-1" || endpoint.GetDeviceType() != "type-1" ||
		endpoint.GetDeviceSerialNumber() != "serial-1" ||
		endpoint.GetDeviceFamily() != "family-1" ||
		endpoint.GetDeviceAccountId() != "account-1" {
		t.Fatalf("endpoint model lost identifying fields: %+v", endpoint)
	}

	if endpoint.GetLocale() != "en-US" {
		t.Fatalf("default locale = %q", endpoint.GetLocale())
	}

	endpoint.Locale = "fr-FR"
	if endpoint.GetLocale() != "fr-FR" {
		t.Fatalf("locale = %q", endpoint.GetLocale())
	}

	if !endpoint.HasFeature(alexaapimodels.FeatureNamePower) || endpoint.HasFeature(alexaapimodels.FeatureNameBrightness) {
		t.Fatal("feature detection returned wrong result")
	}

	if !alexaapimodels.ValidDisplayCategory(string(alexaapimodels.EndpointDisplayCategoryLight)) ||
		alexaapimodels.ValidDisplayCategory("SYNTHETIC_UNKNOWN") {
		t.Fatal("display category validation returned wrong result")
	}

	if alexaapimodels.FeatureNamePower.String() != "power" ||
		alexaapimodels.FeatureNamePower.EventNamespace() != "alexa.endpoint.feature.power" ||
		alexaapimodels.FeatureOperationNameTurnOn.String() != "turnOn" {
		t.Fatal("feature identifiers changed")
	}

	config := alexaapimodels.DefaultDeviceRegistrationConfig("serial-1", "Synthetic speaker")
	if config.DeviceSerial != "serial-1" || config.DeviceName != "Synthetic speaker" || config.DeviceType == "" {
		t.Fatalf("registration defaults = %+v", config)
	}
}

func TestSyntheticErrorContracts(t *testing.T) {
	t.Parallel()

	cause := syntheticFailureError("synthetic transport failure")

	wrapping := []struct {
		name     string
		err      error
		identify func(error) bool
	}{
		{"connection", alexaapimodels.NewConnectionError("connect", cause), alexaapimodels.IsConnectionError},
		{"network", alexaapimodels.NewNetworkError("request", cause), alexaapimodels.IsNetworkError},
		{"token", alexaapimodels.NewTokenError("token", cause), alexaapimodels.IsTokenError},
		{"bad request", &alexaapimodels.BadRequestError{Message: "bad", Err: cause}, alexaapimodels.IsBadRequestError},
		{"unauthorized", &alexaapimodels.UnauthorizedError{Message: "denied", Err: cause}, alexaapimodels.IsUnauthorizedError},
		{"server", &alexaapimodels.InternalServerError{Message: "server", Err: cause}, alexaapimodels.IsInternalServerError},
		{"ping", alexaapimodels.NewPingError(0, "ping", cause), alexaapimodels.IsPingError},
	}
	for _, test := range wrapping {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if !test.identify(test.err) || !errors.Is(test.err, cause) ||
				!strings.Contains(test.err.Error(), test.name[:1]) {
				t.Fatalf("error contract = %v", test.err)
			}

			if test.identify(cause) {
				t.Fatalf("plain error identified as %s", test.name)
			}
		})
	}

	assertSyntheticAuthenticationError(t, cause)
	assertSyntheticHTTPError(t, cause)
	assertSyntheticClosedAndPingErrors(t, cause)
}

func assertSyntheticAuthenticationError(t *testing.T, cause error) {
	t.Helper()

	got := alexaapimodels.NewAuthenticationError("", 401)
	if got.Error() != "authentication error (status: 401)" || !alexaapimodels.IsAuthenticationError(got) {
		t.Fatalf("authentication fallback = %v", got)
	}

	got = alexaapimodels.NewAuthenticationError("expired", 401)
	if !strings.Contains(got.Error(), "expired") || alexaapimodels.IsAuthenticationError(cause) {
		t.Fatalf("authentication message = %v", got)
	}
}

func assertSyntheticHTTPError(t *testing.T, cause error) {
	t.Helper()

	httpError := alexaapimodels.NewHTTPError(
		&http.Response{StatusCode: http.StatusTooManyRequests, Status: "429 Too Many Requests"},
		"slow down",
	)
	if !alexaapimodels.IsHTTPStatusCode(httpError, 429) || alexaapimodels.IsHTTPStatusCode(httpError, 500) ||
		!alexaapimodels.IsHTTPError(httpError) ||
		!strings.Contains(httpError.Error(), "slow down") {
		t.Fatalf("HTTP error = %v", httpError)
	}

	if alexaapimodels.IsHTTPStatusCode(cause, 429) || alexaapimodels.IsHTTPError(cause) {
		t.Fatal("plain error identified as HTTP")
	}
}

func assertSyntheticClosedAndPingErrors(t *testing.T, cause error) {
	t.Helper()

	closedError := alexaapimodels.NewClosedError("")
	if closedError.Error() != "connection closed" || !alexaapimodels.IsClosedError(closedError) ||
		alexaapimodels.IsClosedError(cause) {
		t.Fatalf("closed error = %v", closedError)
	}

	pingError := alexaapimodels.NewPingError(503, "offline", cause)
	if !strings.Contains(pingError.Error(), "503") || !errors.Is(pingError, cause) {
		t.Fatalf("status ping error = %v", pingError)
	}
}
