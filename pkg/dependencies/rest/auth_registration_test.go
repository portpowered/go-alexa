package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestRegistrationFlowsSendSyntheticAuthDataAndReturnTokens(t *testing.T) {
	config := &DeviceRegistrationConfig{
		AppName:      "synthetic-registration-client",
		AppVersion:   "1.0-test",
		DeviceType:   "synthetic-device-type",
		Domain:       "Device",
		DeviceModel:  "synthetic-model",
		OSVersion:    "synthetic-os",
		DeviceSerial: "synthetic-serial",
		DeviceName:   "Synthetic device",
	}
	tests := []struct {
		name  string
		call  func(*Client) error
		check func(*testing.T, map[string]any)
	}{
		{
			name: "email password",
			call: func(client *Client) error {
				response, err := client.RegisterWithEmailPassword(context.Background(), "user@example.invalid", "synthetic-password", config)
				if err == nil && (response.Response.Success.Tokens.Bearer.AccessToken != "synthetic-access-token" || response.Response.Success.Tokens.Bearer.RefreshToken != "synthetic-refresh-token") {
					t.Errorf("unexpected registration response: %#v", response)
				}
				return err
			},
			check: func(t *testing.T, body map[string]any) {
				data := body["registration_data"].(map[string]any)
				auth := body["auth_data"].(map[string]any)["email_password"].(map[string]any)
				if data["app_name"] != "synthetic-registration-client" || data["device_serial"] != "synthetic-serial" || auth["email"] != "user@example.invalid" || auth["password"] != "synthetic-password" {
					t.Errorf("unexpected synthetic email-registration request: %#v", body)
				}
			},
		},
		{
			name: "code pair",
			call: func(client *Client) error {
				response, err := client.RegisterWithCodePair(context.Background(), "synthetic-public-code", "synthetic-private-code", config)
				if err == nil && (response.Response.Success.Tokens.Bearer.AccessToken != "synthetic-access-token" || response.Response.Success.Tokens.Bearer.RefreshToken != "synthetic-refresh-token") {
					t.Errorf("unexpected registration response: %#v", response)
				}
				return err
			},
			check: func(t *testing.T, body map[string]any) {
				data := body["registration_data"].(map[string]any)
				auth := body["auth_data"].(map[string]any)["code_pair"].(map[string]any)
				if data["device_model"] != "synthetic-model" || auth["public_code"] != "synthetic-public-code" || auth["private_code"] != "synthetic-private-code" {
					t.Errorf("unexpected synthetic code-pair request: %#v", body)
				}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			client := NewClient(
				WithAmazonapiBaseURI("https://auth.synthetic.test"),
				WithBearerToken("must-not-be-used-by-registration"),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPost || request.URL.Path != "/auth/register" || request.Header.Get("Authorization") != "" {
						t.Errorf("unexpected unauthenticated registration request: %s %s Authorization=%q", request.Method, request.URL, request.Header.Get("Authorization"))
					}
					var body map[string]any
					if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
						t.Fatalf("decode registration request: %v", err)
					}
					testCase.check(t, body)
					return syntheticResponse(request, http.StatusOK, `{"response":{"success":{"tokens":{"bearer":{"access_token":"synthetic-access-token","refresh_token":"synthetic-refresh-token"}}}}}`), nil
				})}),
			)
			if err := testCase.call(client); err != nil {
				t.Fatalf("registration returned an error: %v", err)
			}
		})
	}
}

func TestEmailRegistrationMapsChallengeAndCodePairErrors(t *testing.T) {
	config := &DeviceRegistrationConfig{AppName: "synthetic-client"}
	client := NewClient(
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusBadRequest, `{"response":{"challenge":{"challenge_reason":"MissingRequiredAuthenticationData","required_authentication_method":"OTP"}}}`), nil
		})}),
	)
	_, err := client.RegisterWithEmailPassword(context.Background(), "user@example.invalid", "synthetic-password", config)
	var challenge *RegistrationChallengeError
	if !errors.As(err, &challenge) || !challenge.IsOTPRequired() || challenge.IsCBLRequired() || challenge.IsAuthenticationFailed() {
		t.Fatalf("expected explicit OTP challenge, got %T: %v", err, err)
	}

	client = NewClient(
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusUnauthorized, `{"message":"synthetic code-pair rejection"}`), nil
		})}),
	)
	if _, err := client.RegisterWithCodePair(context.Background(), "public", "private", config); !alexaapimodels.IsNetworkError(err) {
		t.Fatalf("expected code-pair failure to return NetworkError, got %T: %v", err, err)
	}

	client = NewClient(
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusBadRequest, `{"message":"not a registration challenge"}`), nil
		})}),
	)
	if _, err := client.RegisterWithEmailPassword(context.Background(), "user@example.invalid", "synthetic-password", config); !alexaapimodels.IsNetworkError(err) {
		t.Fatalf("expected non-challenge email failure to return NetworkError, got %T: %v", err, err)
	}
}

func TestRegistrationConfigurationDefaultsChallengePredicatesAndOptions(t *testing.T) {
	defaults := DefaultDeviceRegistrationConfig("synthetic-serial", "Synthetic device")
	if defaults.DeviceSerial != "synthetic-serial" || defaults.DeviceName != "Synthetic device" || defaults.AppName == "" || defaults.DeviceType == "" {
		t.Fatalf("unexpected default registration configuration: %#v", defaults)
	}
	for _, testCase := range []struct {
		reason  string
		method  string
		otp     bool
		cbl     bool
		failure bool
	}{
		{reason: "MissingRequiredAuthenticationData", otp: true},
		{reason: "HandleOnWebView", cbl: true},
		{reason: "AuthenticationFailed", method: "GenericClaimPassword", failure: true},
		{reason: "UnknownChallenge"},
	} {
		err := &RegistrationChallengeError{ChallengeReason: testCase.reason, RequiredAuthenticationMethod: testCase.method}
		if err.IsOTPRequired() != testCase.otp || err.IsCBLRequired() != testCase.cbl || err.IsAuthenticationFailed() != testCase.failure || !strings.Contains(err.Error(), testCase.reason) {
			t.Errorf("challenge predicates returned the wrong result for %#v", testCase)
		}
	}

	client := NewClient()
	client.Apply(WithRegion(alexaapimodels.RegionEU), WithCustomerID("synthetic-customer"), WithCSRFTokenGetter(func(context.Context) (string, error) {
		return "synthetic-csrf-from-getter", nil
	}))
	if client.amazonalexaapiBaseUri != alexaapimodels.ApiServiceUriEu || client.amazonapiBaseUri != alexaapimodels.AmazonApiServiceUriEu || client.alexaAmazonBaseUri != alexaapimodels.AlexaAmazonBaseUriEu || client.customerID != "synthetic-customer" {
		t.Fatalf("client options were not applied: %#v", client)
	}
	csrf, err := client.GetCSRFToken(context.Background())
	if err != nil || csrf != "synthetic-csrf-from-getter" {
		t.Fatalf("CSRF token getter returned %q, %v", csrf, err)
	}
	if token, err := client.GetCSRFToken(context.Background()); err != nil || token != "synthetic-csrf-from-getter" {
		t.Fatalf("cached CSRF token getter returned %q, %v", token, err)
	}
}

func TestRegistrationMethodsRejectMissingConfiguration(t *testing.T) {
	client := NewClient()
	if _, err := client.RegisterWithEmailPassword(context.Background(), "email", "password", nil); !alexaapimodels.IsBadRequestError(err) {
		t.Fatalf("expected email registration to reject nil config, got %v", err)
	}
	if _, err := client.RegisterWithCodePair(context.Background(), "public", "private", nil); !alexaapimodels.IsBadRequestError(err) {
		t.Fatalf("expected code-pair registration to reject nil config, got %v", err)
	}
	if _, err := client.GenerateCodePair(context.Background(), nil); !alexaapimodels.IsBadRequestError(err) {
		t.Fatalf("expected code generation to reject nil config, got %v", err)
	}
}
