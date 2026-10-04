//nolint:testpackage // Exercises private transport boundaries with synthetic failures.
package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestCodePairFailuresExposeStatesWithoutProviderPayload(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"authorization_pending", "expired_token", "access_denied", "unknown"} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				return syntheticResponse(request, http.StatusBadRequest, `{"error":"`+code+`","message":"private-provider-body"}`), nil
			})}))
			_, err := client.RegisterWithCodePair(context.Background(), "synthetic-public", "synthetic-private", DefaultDeviceRegistrationConfig("serial", "name"))
			assertSafeErrorChain(t, err)

			var (
				pending  *alexaapimodels.CodePairPendingError
				expired  *alexaapimodels.CodePairExpiredError
				denied   *alexaapimodels.CodePairDeniedError
				rejected *alexaapimodels.CodePairRegistrationError
			)

			if errors.As(err, &pending) != (code == "authorization_pending") ||
				errors.As(err, &expired) != (code == "expired_token") ||
				errors.As(err, &denied) != (code == "access_denied") ||
				errors.As(err, &rejected) != (code == "unknown") {
				t.Fatalf("wrong code-link state for %s: %T", code, err)
			}
		})
	}
}

func TestRefreshAndCookieFailuresDoNotRetainProviderPayload(t *testing.T) {
	t.Parallel()

	client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
		return syntheticResponse(request, http.StatusForbidden, `{"message":"private-provider-body"}`), nil
	})}))
	_, refreshErr := client.RefreshAccessToken(context.Background(), "private-refresh", DefaultDeviceRegistrationConfig("serial", "name"))
	assertSafeErrorChain(t, refreshErr)

	_, cookieErr := client.ExchangeRefreshTokenForCookies(context.Background(), "private-refresh", "synthetic.amazon.test")
	assertSafeErrorChain(t, cookieErr)

	_, passwordErr := client.RegisterWithEmailPassword(context.Background(),
		"synthetic@example.invalid", "private-password", DefaultDeviceRegistrationConfig("serial", "name"))
	assertSafeErrorChain(t, passwordErr)
	assertSafeErrorChain(t, &RegistrationChallengeError{ChallengeReason: "private-challenge", RequiredAuthenticationMethod: ""})
}

func assertSafeErrorChain(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected provider rejection")
	}

	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		if strings.Contains(cause.Error(), "private-") {
			t.Fatalf("provider payload survived in %T", cause)
		}
	}
}

func TestAuthenticationPOSTRetryRecreatesRequestBody(t *testing.T) {
	t.Parallel()

	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read retry request: %v", err)
		}

		requests <- string(body)

		if len(requests) == 1 {
			response.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		response.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(response, `{"access_token":"synthetic-access","token_type":"bearer","expires_in":3600}`)
	}))
	t.Cleanup(server.Close)
	client := NewClient(WithAmazonapiBaseURI(server.URL))

	token, err := client.RefreshAccessToken(context.Background(), "synthetic-refresh", DefaultDeviceRegistrationConfig("serial", "name"))
	if err != nil || token.AccessToken != "synthetic-access" {
		t.Fatalf("refresh after retry: token=%#v err=%v", token, err)
	}

	first, second := <-requests, <-requests
	if first == "" || first != second || !strings.Contains(second, "synthetic-refresh") {
		t.Fatalf("POST retry changed its body: first=%q second=%q", first, second)
	}
}
