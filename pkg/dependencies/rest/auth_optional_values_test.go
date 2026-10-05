package rest_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	rest "github.com/portpowered/go-alexa/pkg/dependencies/rest"
)

type optionalValuesRoundTripper func(*http.Request) (*http.Response, error)

func (roundTripper optionalValuesRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTripper(request)
}

func TestOptionalAuthValuesKeepZeroDefaults(t *testing.T) {
	t.Parallel()

	t.Run("cookie fields", func(t *testing.T) {
		t.Parallel()

		client := rest.NewClient(rest.WithHTTPClient(&http.Client{
			Transport: optionalValuesRoundTripper(func(request *http.Request) (*http.Response, error) {
				return optionalValuesResponse(request, http.StatusOK, `{"response":{"tokens":{"cookies":{".alexa.amazon.test":[{}]}}}}`), nil
			}),
		}))

		cookies, err := client.ExchangeRefreshTokenForCookies(context.Background(), "synthetic-refresh", "synthetic.amazon.test")
		if err != nil {
			t.Fatalf("ExchangeRefreshTokenForCookies returned an error: %v", err)
		}

		cookie := cookies["alexa.amazon.test:"]
		if cookie == nil {
			t.Fatalf("cookie with missing optional fields was not returned: %#v", cookies)
		}

		if cookie.Name != "" || cookie.Value != "" || cookie.Path != "" || cookie.Secure || cookie.HttpOnly || !cookie.Expires.IsZero() {
			t.Errorf("missing optional fields did not keep their zero values: %#v", cookie)
		}
	})

	t.Run("challenge authentication method", func(t *testing.T) {
		t.Parallel()

		client := rest.NewClient(rest.WithHTTPClient(&http.Client{
			Transport: optionalValuesRoundTripper(func(request *http.Request) (*http.Response, error) {
				return optionalValuesResponse(request, http.StatusBadRequest, `{"response":{"challenge":{"challenge_reason":"MissingRequiredAuthenticationData"}}}`), nil
			}),
		}))

		_, err := client.RegisterWithEmailPassword(
			context.Background(),
			"user@example.invalid",
			"synthetic-password",
			&rest.DeviceRegistrationConfig{AppName: "synthetic-client"},
		)

		var challenge *rest.RegistrationChallengeError

		if !errors.As(err, &challenge) {
			t.Fatalf("expected registration challenge, got %T: %v", err, err)
		}

		if challenge.ChallengeReason != "MissingRequiredAuthenticationData" || challenge.RequiredAuthenticationMethod != "" {
			t.Errorf("unexpected challenge values: %#v", challenge)
		}
	})
}

func optionalValuesResponse(request *http.Request, statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode:    statusCode,
		Header:        make(http.Header),
		Body:          io.NopCloser(strings.NewReader(body)),
		Request:       request,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(body)),
	}
}
