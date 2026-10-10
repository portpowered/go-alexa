//nolint:testpackage // Exercises private auth transport and registration state.
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

const (
	syntheticAccessToken  = "synthetic-access-token"
	syntheticCSRFToken    = "synthetic-csrf-from-getter" //nolint:gosec // This fixed token exists only in synthetic test fixtures.
	syntheticRefreshToken = "synthetic-refresh-token"
)

func requireTestObject(t *testing.T, value any) map[string]any {
	t.Helper()

	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", value)
	}

	return object
}

func TestRegistrationFlowsSendSyntheticAuthDataAndReturnTokens(t *testing.T) {
	t.Parallel()

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
				if err == nil && (response.Response.Success.Tokens.Bearer.AccessToken != syntheticAccessToken || response.Response.Success.Tokens.Bearer.RefreshToken != syntheticRefreshToken) {
					t.Errorf("unexpected registration response: %#v", response)
				}

				return err
			},
			check: func(t *testing.T, body map[string]any) {
				t.Helper()

				data := requireTestObject(t, body["registration_data"])
				authData := requireTestObject(t, body["auth_data"])
				auth := requireTestObject(t, authData["email_password"])
				if data["app_name"] != "synthetic-registration-client" || data["device_serial"] != "synthetic-serial" || auth["email"] != "user@example.invalid" || auth["password"] != "synthetic-password" {
					t.Errorf("unexpected synthetic email-registration request: %#v", body)
				}
			},
		},
		{
			name: "code pair",
			call: func(client *Client) error {
				response, err := client.RegisterWithCodePair(context.Background(), "synthetic-public-code", "synthetic-private-code", config)
				if err == nil && (response.Response.Success.Tokens.Bearer.AccessToken != syntheticAccessToken || response.Response.Success.Tokens.Bearer.RefreshToken != syntheticRefreshToken) {
					t.Errorf("unexpected registration response: %#v", response)
				}

				return err
			},
			check: func(t *testing.T, body map[string]any) {
				t.Helper()

				data := requireTestObject(t, body["registration_data"])
				authData := requireTestObject(t, body["auth_data"])
				auth := requireTestObject(t, authData["code_pair"])
				if data["device_model"] != "synthetic-model" || auth["public_code"] != "synthetic-public-code" || auth["private_code"] != "synthetic-private-code" {
					t.Errorf("unexpected synthetic code-pair request: %#v", body)
				}
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithAmazonapiBaseURI("https://auth.synthetic.test"),
				WithBearerToken("must-not-be-used-by-registration"),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					if request.Method != http.MethodPost || request.URL.Path != "/auth/register" || request.Header.Get("Authorization") != "" {
						t.Errorf("unexpected unauthenticated registration request: %s %s Authorization=%q", request.Method, request.URL, request.Header.Get("Authorization"))
					}
					var body map[string]any
					err := json.NewDecoder(request.Body).Decode(&body)
					if err != nil {
						t.Fatalf("decode registration request: %v", err)
					}
					testCase.check(t, body)

					responseBody := `{"response":{"success":{"tokens":{"bearer":{"access_token":"` + syntheticAccessToken + `","refresh_token":"` + syntheticRefreshToken + `"}}}}}`

					return syntheticResponse(request, http.StatusOK, responseBody), nil
				})}),
			)

			err := testCase.call(client)
			if err != nil {
				t.Fatalf("registration returned an error: %v", err)
			}
		})
	}
}

func TestEmailRegistrationMapsChallengeAndCodePairErrors(t *testing.T) {
	t.Parallel()

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

	_, codePairErr := client.RegisterWithCodePair(context.Background(), "public", "private", config)

	var rejection *alexaapimodels.CodePairRegistrationError

	if !errors.As(codePairErr, &rejection) || rejection.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected typed code-pair rejection, got %T: %v", codePairErr, codePairErr)
	}

	client = NewClient(
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusBadRequest, `{"message":"not a registration challenge"}`), nil
		})}),
	)

	_, emailRegistrationErr := client.RegisterWithEmailPassword(context.Background(), "user@example.invalid", "synthetic-password", config)
	if !alexaapimodels.IsNetworkError(emailRegistrationErr) {
		t.Fatalf("expected non-challenge email failure to return NetworkError, got %T: %v", emailRegistrationErr, emailRegistrationErr)
	}
}

func TestRegistrationConfigurationDefaultsChallengePredicatesAndOptions(t *testing.T) {
	t.Parallel()

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
		{reason: "MissingRequiredAuthenticationData", method: "", otp: true, cbl: false, failure: false},
		{reason: "HandleOnWebView", method: "", otp: false, cbl: true, failure: false},
		{reason: "AuthenticationFailed", method: "GenericClaimPassword", otp: false, cbl: false, failure: true},
		{reason: "UnknownChallenge", method: "", otp: false, cbl: false, failure: false},
	} {
		err := &RegistrationChallengeError{ChallengeReason: testCase.reason, RequiredAuthenticationMethod: testCase.method}
		if err.IsOTPRequired() != testCase.otp || err.IsCBLRequired() != testCase.cbl || err.IsAuthenticationFailed() != testCase.failure || !strings.Contains(err.Error(), testCase.reason) {
			t.Errorf("challenge predicates returned the wrong result for %#v", testCase)
		}
	}

	client := NewClient()
	client.Apply(WithRegion(alexaapimodels.RegionEU), WithCustomerID("synthetic-customer"), WithCSRFTokenGetter(func(context.Context) (string, error) {
		return syntheticCSRFToken, nil
	}))

	if client.amazonalexaapiBaseURI != alexaapimodels.ApiServiceUriEu || client.amazonapiBaseURI != alexaapimodels.AmazonApiServiceUriEu || client.alexaAmazonBaseURI != alexaapimodels.AlexaAmazonBaseUriEu || client.customerID != "synthetic-customer" {
		t.Fatalf("client options were not applied: %#v", client)
	}

	csrf, err := client.GetCSRFToken(context.Background())
	if err != nil || csrf != syntheticCSRFToken {
		t.Fatalf("CSRF token getter returned %q, %v", csrf, err)
	}

	token, cachedTokenErr := client.GetCSRFToken(context.Background())
	if cachedTokenErr != nil || token != syntheticCSRFToken {
		t.Fatalf("cached CSRF token getter returned %q, %v", token, cachedTokenErr)
	}
}

func TestRegistrationMethodsRejectMissingConfiguration(t *testing.T) {
	t.Parallel()

	client := NewClient()

	_, emailConfigErr := client.RegisterWithEmailPassword(context.Background(), "email", "password", nil)
	if !alexaapimodels.IsBadRequestError(emailConfigErr) {
		t.Fatalf("expected email registration to reject nil config, got %v", emailConfigErr)
	}

	_, codePairConfigErr := client.RegisterWithCodePair(context.Background(), "public", "private", nil)
	if !alexaapimodels.IsBadRequestError(codePairConfigErr) {
		t.Fatalf("expected code-pair registration to reject nil config, got %v", codePairConfigErr)
	}

	{
		_, err := client.GenerateCodePair(context.Background(), nil)
		if !alexaapimodels.IsBadRequestError(err) {
			t.Fatalf("expected code generation to reject nil config, got %v", err)
		}
	}
}
