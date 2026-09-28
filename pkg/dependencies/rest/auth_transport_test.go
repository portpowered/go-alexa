package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestRefreshAccessTokenSendsExplicitRefreshInputAndReturnsToken(t *testing.T) {
	client := NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.String() != "https://auth.synthetic.test/auth/token" {
				t.Errorf("unexpected request target: %s %s", request.Method, request.URL)
			}
			if request.Header.Get("Authorization") != "" {
				t.Errorf("token refresh unexpectedly used Authorization: %q", request.Header.Get("Authorization"))
			}
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode refresh request: %v", err)
			}
			if body["source_token"] != "synthetic-refresh-token" || body["source_token_type"] != "refresh_token" || body["requested_token_type"] != "access_token" {
				t.Errorf("unexpected token exchange request: %#v", body)
			}
			metadata, ok := body["device_metadata"].(map[string]any)
			if !ok || metadata["device_serial"] != "synthetic-device" || metadata["manufacturer"] != "Synthetic" {
				t.Errorf("unexpected device metadata: %#v", body["device_metadata"])
			}
			return syntheticResponse(request, http.StatusOK, `{"access_token":"synthetic-access-token","expires_in":3600}`), nil
		})}),
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
	)

	response, err := client.RefreshAccessToken(context.Background(), "synthetic-refresh-token", &DeviceRegistrationConfig{
		AppName:      "synthetic-client",
		AppVersion:   "1.2.3",
		DeviceType:   "synthetic-device-type",
		DeviceModel:  "synthetic-model",
		OSVersion:    "synthetic-os",
		DeviceSerial: "synthetic-device",
		Manufacturer: "Synthetic",
	})
	if err != nil {
		t.Fatalf("RefreshAccessToken returned an error: %v", err)
	}
	if response.AccessToken != "synthetic-access-token" || response.ExpiresInSeconds != 3600 {
		t.Fatalf("unexpected token response: %#v", response)
	}
}

func TestRefreshAccessTokenRequiresConfigurationAndReportsHTTPFailure(t *testing.T) {
	client := NewClient()
	if _, err := client.RefreshAccessToken(context.Background(), "synthetic-refresh-token", nil); !alexaapimodels.IsBadRequestError(err) {
		t.Fatalf("expected missing-config BadRequestError, got %v", err)
	}

	client = NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusUnauthorized, `{"message":"synthetic rejection"}`), nil
		})}),
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
	)
	if _, err := client.RefreshAccessToken(context.Background(), "synthetic-refresh-token", &DeviceRegistrationConfig{}); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected an error that reports HTTP 401, got %v", err)
	}
}

func TestExchangeRefreshTokenForCookiesReturnsSyntheticCookieMaterial(t *testing.T) {
	client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.String() != "https://api.synthetic.amazon.test/ap/exchangetoken/cookies" {
			t.Errorf("unexpected cookie-exchange target: %s %s", request.Method, request.URL)
		}
		if request.Header.Get("x-amzn-identity-auth-domain") != "api.synthetic.amazon.test" {
			t.Errorf("unexpected identity-auth-domain header: %q", request.Header.Get("x-amzn-identity-auth-domain"))
		}
		if request.Header.Get("Authorization") != "" {
			t.Errorf("cookie exchange unexpectedly used Authorization: %q", request.Header.Get("Authorization"))
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode cookie-exchange request: %v", err)
		}
		if body["source_token"] != "synthetic-refresh-token" || body["requested_token_type"] != "auth_cookies" || body["domain"] != "synthetic.amazon.test" {
			t.Errorf("unexpected cookie-exchange input: %#v", body)
		}
		return syntheticResponse(request, http.StatusOK, `{"response":{"tokens":{"cookies":{".alexa.amazon.test":[{"Name":"session-id","Value":"synthetic-session","Path":"/","Secure":true,"HttpOnly":true,"Expires":"Mon, 02 Jan 2006 15:04:05 GMT"}]}}}}`), nil
	})}))

	cookies, err := client.ExchangeRefreshTokenForCookies(context.Background(), "synthetic-refresh-token", "synthetic.amazon.test")
	if err != nil {
		t.Fatalf("ExchangeRefreshTokenForCookies returned an error: %v", err)
	}
	cookie := cookies["alexa.amazon.test:session-id"]
	if cookie == nil {
		t.Fatalf("session cookie missing from result: %#v", cookies)
	}
	if cookie.Value != "synthetic-session" || cookie.Path != "/" || !cookie.Secure || !cookie.HttpOnly || cookie.Expires.IsZero() {
		t.Errorf("unexpected session cookie: %#v", cookie)
	}
}

func TestExchangeRefreshTokenForCookiesHandlesEmptyAndInvalidResponses(t *testing.T) {
	t.Run("no cookies", func(t *testing.T) {
		client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusOK, `{"response":{"tokens":{}}}`), nil
		})}))
		cookies, err := client.ExchangeRefreshTokenForCookies(context.Background(), "synthetic-refresh-token", "synthetic.amazon.test")
		if err != nil || len(cookies) != 0 {
			t.Fatalf("expected empty cookie result without error, got %#v, %v", cookies, err)
		}
	})

	t.Run("bad status", func(t *testing.T) {
		client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusForbidden, "synthetic denial"), nil
		})}))
		if _, err := client.ExchangeRefreshTokenForCookies(context.Background(), "synthetic-refresh-token", "synthetic.amazon.test"); err == nil || !strings.Contains(err.Error(), "synthetic denial") {
			t.Fatalf("expected exchange rejection details, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusOK, "not-json"), nil
		})}))
		if _, err := client.ExchangeRefreshTokenForCookies(context.Background(), "synthetic-refresh-token", "synthetic.amazon.test"); err == nil {
			t.Fatal("expected invalid JSON to fail")
		}
	})
}
