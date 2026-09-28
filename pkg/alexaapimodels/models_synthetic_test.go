package alexaapimodels

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// These values are synthetic contract examples, not Amazon account captures.
func TestSyntheticEndpointModel(t *testing.T) {
	endpoint := &Endpoint{EndpointID: "endpoint-1", DeviceType: "type-1", DeviceSerialNumber: "serial-1", DeviceFamily: "family-1", DeviceAccountId: "account-1", Features: []Feature{{Name: FeatureNamePower}}}
	if endpoint.GetEndpointId() != "endpoint-1" || endpoint.GetDeviceType() != "type-1" || endpoint.GetDeviceSerialNumber() != "serial-1" || endpoint.GetDeviceFamily() != "family-1" || endpoint.GetDeviceAccountId() != "account-1" {
		t.Fatalf("endpoint model lost identifying fields: %+v", endpoint)
	}
	if endpoint.GetLocale() != "en-US" {
		t.Fatalf("default locale = %q", endpoint.GetLocale())
	}
	endpoint.Locale = "fr-FR"
	if endpoint.GetLocale() != "fr-FR" {
		t.Fatalf("locale = %q", endpoint.GetLocale())
	}
	if !endpoint.HasFeature(FeatureNamePower) || endpoint.HasFeature(FeatureNameBrightness) {
		t.Fatal("feature detection returned wrong result")
	}
	if !ValidDisplayCategory(string(EndpointDisplayCategoryLight)) || ValidDisplayCategory("SYNTHETIC_UNKNOWN") {
		t.Fatal("display category validation returned wrong result")
	}
	if FeatureNamePower.String() != "power" || FeatureNamePower.EventNamespace() != "alexa.endpoint.feature.power" || FeatureOperationNameTurnOn.String() != "turnOn" {
		t.Fatal("feature identifiers changed")
	}
	config := DefaultDeviceRegistrationConfig("serial-1", "Synthetic speaker")
	if config.DeviceSerial != "serial-1" || config.DeviceName != "Synthetic speaker" || config.DeviceType == "" {
		t.Fatalf("registration defaults = %+v", config)
	}
}

func TestSyntheticErrorContracts(t *testing.T) {
	cause := errors.New("synthetic transport failure")
	wrapping := []struct {
		name     string
		err      error
		identify func(error) bool
	}{
		{"connection", NewConnectionError("connect", cause), IsConnectionError},
		{"network", NewNetworkError("request", cause), IsNetworkError},
		{"token", NewTokenError("token", cause), IsTokenError},
		{"bad request", &BadRequestError{Message: "bad", Err: cause}, IsBadRequestError},
		{"unauthorized", &UnauthorizedError{Message: "denied", Err: cause}, IsUnauthorizedError},
		{"server", &InternalServerError{Message: "server", Err: cause}, IsInternalServerError},
		{"ping", NewPingError(0, "ping", cause), IsPingError},
	}
	for _, test := range wrapping {
		t.Run(test.name, func(t *testing.T) {
			if !test.identify(test.err) || !errors.Is(test.err, cause) || !strings.Contains(test.err.Error(), test.name[:1]) {
				t.Fatalf("error contract = %v", test.err)
			}
			if test.identify(cause) {
				t.Fatalf("plain error identified as %s", test.name)
			}
		})
	}
	if got := NewAuthenticationError("", 401); got.Error() != "authentication error (status: 401)" || !IsAuthenticationError(got) {
		t.Fatalf("authentication fallback = %v", got)
	}
	if got := NewAuthenticationError("expired", 401); !strings.Contains(got.Error(), "expired") || IsAuthenticationError(cause) {
		t.Fatalf("authentication message = %v", got)
	}
	if got := NewHTTPError(&http.Response{StatusCode: 429, Status: "429 Too Many Requests"}, "slow down"); !IsHTTPStatusCode(got, 429) || IsHTTPStatusCode(got, 500) || !IsHTTPError(got) || !strings.Contains(got.Error(), "slow down") {
		t.Fatalf("HTTP error = %v", got)
	}
	if IsHTTPStatusCode(cause, 429) || IsHTTPError(cause) {
		t.Fatal("plain error identified as HTTP")
	}
	if got := NewClosedError(""); got.Error() != "connection closed" || !IsClosedError(got) || IsClosedError(cause) {
		t.Fatalf("closed error = %v", got)
	}
	if got := NewPingError(503, "offline", cause); !strings.Contains(got.Error(), "503") || !errors.Is(got, cause) {
		t.Fatalf("status ping error = %v", got)
	}
}
