//nolint:testpackage // Exercises private request validation and transport behavior.
package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

func TestCustomHeaderMustBeDeclaredInSchema(t *testing.T) {
	t.Parallel()

	client := NewClient(WithBearerToken("synthetic-token"), WithHTTPClient(&http.Client{
		Transport: syntheticRoundTripper(func(*http.Request) (*http.Response, error) {
			t.Fatal("undeclared request header reached transport")

			return nil, syntheticFailureError("undeclared request header reached transport")
		}),
	}))

	response, err := client.doRequestWithFullURL(context.Background(), apiroutes.MethodGetUserInfo,
		"https://alexa.synthetic.test"+apiroutes.PathGetUserInfo, nil,
		map[string]string{"X-Undeclared": "synthetic"}, false)
	if response != nil {
		closeErr := response.Body.Close()
		if closeErr != nil {
			t.Errorf("close undeclared-header response body: %v", closeErr)
		}
	}

	if !alexaapimodels.IsBadRequestError(err) {
		t.Fatalf("expected undeclared header error, got %v", err)
	}
}

type syntheticEndpoint struct {
	id      string
	device  string
	serial  string
	locale  string
	family  string
	account string
}

func (endpoint syntheticEndpoint) GetDeviceType() string         { return endpoint.device }
func (endpoint syntheticEndpoint) GetDeviceSerialNumber() string { return endpoint.serial }
func (endpoint syntheticEndpoint) GetLocale() string             { return endpoint.locale }
func (endpoint syntheticEndpoint) GetEndpointId() string         { return endpoint.id }
func (endpoint syntheticEndpoint) GetDeviceFamily() string       { return endpoint.family }
func (endpoint syntheticEndpoint) GetDeviceAccountId() string    { return endpoint.account }

func TestGetEndpointByIDUsesBearerAndDecodesSyntheticDevice(t *testing.T) {
	t.Parallel()

	client := NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodGet || request.URL.Path != "/v2/endpoints/synthetic-endpoint" {
				t.Errorf("unexpected endpoint request: %s %s", request.Method, request.URL)
			}
			if got := request.URL.Query()["expand"]; len(got) != 2 || got[0] != expandAll || got[1] != "feature:power" {
				t.Errorf("unexpected expand query: %#v", request.URL.Query())
			}
			if request.Header.Get("Authorization") != "Bearer synthetic-access-token" {
				t.Errorf("unexpected Authorization header: %q", request.Header.Get("Authorization"))
			}
			if request.Header.Get("Accept") != "application/json" {
				t.Errorf("unexpected Accept header: %q", request.Header.Get("Accept"))
			}

			return syntheticResponse(request, http.StatusOK,
				`{"id":"synthetic-endpoint","name":"Test lamp","type":"LIGHT",`+
					`"state":{"on":true},"metadata":{"room":"synthetic"}}`), nil
		})}),
		WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-access-token"),
	)

	device, err := client.GetEndpointByID(context.Background(), "synthetic-endpoint", []string{expandAll, "feature:power"})
	if err != nil {
		t.Fatalf("GetEndpointByID returned an error: %v", err)
	}

	if device.ID != "synthetic-endpoint" || device.Name != "Test lamp" ||
		device.Type != "LIGHT" || device.State["on"] != true ||
		device.Metadata["room"] != "synthetic" {
		t.Fatalf("unexpected decoded device: %#v", device)
	}
}

func TestGetUserInfoUsesCookieAndCSRFHeaders(t *testing.T) {
	t.Parallel()

	client := NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.URL.Path != "/api/users/me" ||
				request.URL.Query().Get("platform") != "synthetic-app" ||
				request.URL.Query().Get("version") != "1.2.3" {
				t.Errorf("unexpected user-info target: %s", request.URL)
			}
			if got := request.Header.Get("Cookie"); !strings.Contains(got, "session-id=synthetic-cookie") {
				t.Errorf("session cookie missing: %q", got)
			}
			if got := request.Header.Get("Csrf"); got != "synthetic-csrf" {
				t.Errorf("unexpected CSRF header: %q", got)
			}
			if request.Header.Get("Authorization") != "" {
				t.Errorf("cookie-auth request also sent Authorization: %q", request.Header.Get("Authorization"))
			}
			if request.Header.Get("User-Agent") == "" || request.Header.Get("Referer") == "" || request.Header.Get("Origin") == "" {
				t.Errorf("expected browser-compatible cookie-auth headers, got %#v", request.Header)
			}

			return syntheticResponse(request, http.StatusOK, `{"id":"synthetic-user","email":"user@example.invalid","fullName":"Synthetic User"}`), nil
		})}),
		WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
		WithCookies(map[string]*http.Cookie{
			"alexa.synthetic.test:session-id": {Name: "session-id", Value: "synthetic-cookie"},
		}),
		WithCSRFToken("synthetic-csrf"),
	)

	user, err := client.GetUserInfo(context.Background(), &GetUserInfoOptions{Platform: "synthetic-app", Version: "1.2.3"})
	if err != nil {
		t.Fatalf("GetUserInfo returned an error: %v", err)
	}

	if user.ID != "synthetic-user" || user.Email != "user@example.invalid" || user.FullName != "Synthetic User" {
		t.Fatalf("unexpected synthetic user info: %#v", user)
	}
}

func TestCookieCredentialsAreExplicitAndCSRFMustBeRetrievedOrConfigured(t *testing.T) {
	t.Parallel()

	t.Run("refresh token is not exchanged implicitly", func(t *testing.T) {
		t.Parallel()

		transportCalls := 0

		client := NewClient(
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				transportCalls++

				return syntheticResponse(request, http.StatusOK, `{"response":{"tokens":{"cookies":{}}}}`), nil
			})}),
			WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
			WithRefreshToken("synthetic-refresh-token"),
		)
		{
			_, err := client.GetUserInfo(context.Background(), nil)
			if !alexaapimodels.IsTokenError(err) {
				t.Fatalf("expected configured cookie auth without an exchange to fail for missing material, got %v", err)
			}
		}

		if transportCalls != 0 {
			t.Fatalf("request performed an implicit credential exchange: %d transport calls", transportCalls)
		}
	})

	t.Run("missing CSRF fails until explicit retrieval", func(t *testing.T) {
		t.Parallel()

		transportCalls := 0
		client := NewClient(
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				transportCalls++
				switch request.URL.Path {
				case "/api/language":
					if !strings.Contains(request.Header.Get("Cookie"), "session-id=synthetic-session") {
						t.Errorf("explicit CSRF lookup omitted the configured session cookie: %q", request.Header.Get("Cookie"))
					}
					response := syntheticResponse(request, http.StatusOK, `{}`)
					response.Header.Add("Set-Cookie", "csrf=synthetic-fetched-csrf; Path=/")

					return response, nil
				case "/api/users/me":
					if request.Header.Get("Csrf") != "synthetic-fetched-csrf" {
						t.Errorf("request did not use explicitly retrieved CSRF token: %q", request.Header.Get("Csrf"))
					}

					return syntheticResponse(request, http.StatusOK, `{"id":"synthetic-user"}`), nil
				default:
					t.Errorf("unexpected request path: %s", request.URL.Path)

					return syntheticResponse(request, http.StatusNotFound, `{}`), nil
				}
			})}),
			WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
			WithCookies(map[string]*http.Cookie{
				"alexa.synthetic.test:session-id": {Name: "session-id", Value: "synthetic-session"},
			}),
			WithRefreshToken("synthetic-refresh-token"),
		)

		{
			_, err := client.GetUserInfo(context.Background(), nil)
			if !alexaapimodels.IsTokenError(err) {
				t.Fatalf("expected missing CSRF to fail before HTTP request, got %v", err)
			}
		}

		if transportCalls != 0 {
			t.Fatalf("missing CSRF triggered network access: %d calls", transportCalls)
		}

		csrf, err := client.GetCSRFToken(context.Background())
		if err != nil || csrf != "synthetic-fetched-csrf" {
			t.Fatalf("explicit GetCSRFToken returned %q, %v", csrf, err)
		}

		{
			_, err := client.GetUserInfo(context.Background(), nil)
			if err != nil {
				t.Fatalf("GetUserInfo after explicit CSRF retrieval failed: %v", err)
			}
		}

		if transportCalls != 2 {
			t.Fatalf("expected one explicit CSRF lookup and one API request, got %d calls", transportCalls)
		}
	})
}

func TestSendInterfaceMessageUsesEndpointFeatureOperationAndPayload(t *testing.T) {
	t.Parallel()

	client := NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.Path != "/v2/endpoints/endpoint-42/interfaces/power/turnOn/" {
				t.Errorf("unexpected interface message request: %s %s", request.Method, request.URL)
			}
			if request.Header.Get("Authorization") != "Bearer synthetic-control-token" {
				t.Errorf("unexpected Authorization header: %q", request.Header.Get("Authorization"))
			}
			var body map[string]any
			err := json.NewDecoder(request.Body).Decode(&body)
			if err != nil {
				t.Fatalf("decode interface message body: %v", err)
			}
			if body["state"] != "ON" || body["origin"] != "synthetic-test" {
				t.Errorf("unexpected interface message body: %#v", body)
			}

			return syntheticResponse(request, http.StatusOK, `{}`), nil
		})}),
		WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-control-token"),
	)

	err := client.SendInterfaceMessage(context.Background(), &alexamodels.InterfaceMessageRequest{
		Endpoint:      syntheticEndpoint{id: "endpoint-42"},
		FeatureName:   "power",
		OperationName: "turnOn",
		Payload:       map[string]any{"state": "ON", "origin": "synthetic-test"},
	})
	if err != nil {
		t.Fatalf("SendInterfaceMessage returned an error: %v", err)
	}
}

func TestControlEndpointUsesPathAndWireCommand(t *testing.T) {
	t.Parallel()

	client := NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.Path != "/v2/endpoints/endpoint-7/control" {
				t.Errorf("unexpected control request: %s %s", request.Method, request.URL)
			}
			var body map[string]any
			err := json.NewDecoder(request.Body).Decode(&body)
			if err != nil {
				t.Fatalf("decode control command: %v", err)
			}
			if body["device_id"] != "endpoint-7" || body["namespace"] != "Alexa.PowerController" || body["name"] != "TurnOn" {
				t.Errorf("unexpected wire command: %#v", body)
			}
			payload, ok := body["payload"].(map[string]any)
			if !ok || payload["powerState"] != "ON" {
				t.Errorf("unexpected command payload: %#v", body["payload"])
			}

			return syntheticResponse(request, http.StatusNoContent, ""), nil
		})}),
		WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-control-token"),
	)

	err := client.ControlEndpoint(context.Background(), "endpoint-7", alexamodels.Command{
		Namespace: "Alexa.PowerController",
		Name:      "TurnOn",
		Payload:   map[string]any{"powerState": "ON"},
	})
	if err != nil {
		t.Fatalf("ControlEndpoint returned an error: %v", err)
	}
}

func TestRestResponseErrorsCoverStatusDecodeAndTokenFailure(t *testing.T) {
	t.Parallel()

	t.Run("HTTP status", func(t *testing.T) {
		t.Parallel()

		client := NewClient(
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				return syntheticResponse(request, http.StatusBadRequest, `{"message":"synthetic bad request"}`), nil
			})}),
			WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
			WithBearerToken("synthetic-token"),
		)

		_, err := client.GetEndpoints(context.Background(), &ListEndpointsOptions{Owner: "~caller"})
		if err == nil || !strings.Contains(err.Error(), "400") {
			t.Fatalf("expected HTTP 400 details, got %v", err)
		}
	})

	t.Run("invalid response JSON", func(t *testing.T) {
		t.Parallel()

		client := NewClient(
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				return syntheticResponse(request, http.StatusOK, "not-json"), nil
			})}),
			WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
			WithBearerToken("synthetic-token"),
		)

		_, err := client.GetEndpoints(context.Background(), &ListEndpointsOptions{Owner: "~caller"})
		if !alexaapimodels.IsBadRequestError(err) {
			t.Fatalf("expected response decode BadRequestError, got %v", err)
		}
	})

	t.Run("token provider error", func(t *testing.T) {
		t.Parallel()

		called := false
		client := NewClient(
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				called = true

				return syntheticResponse(request, http.StatusOK, `{}`), nil
			})}),
			WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
			WithTokenGetter(func(context.Context) (string, error) { return "", syntheticFailureError("synthetic token failure") }),
		)

		_, err := client.GetEndpoints(context.Background(), &ListEndpointsOptions{Owner: "~caller"})
		if !alexaapimodels.IsTokenError(err) || called {
			t.Fatalf("expected token failure before transport, err=%v, transport-called=%v", err, called)
		}
	})
}

func TestGetEndpointsEncodesOptions(t *testing.T) {
	t.Parallel()

	client := NewClient(
		WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-rest-token"),
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodGet || request.URL.Path != "/v2/endpoints" {
				t.Errorf("unexpected list route: %s %s", request.Method, request.URL)
			}

			query := request.URL.Query()
			if query.Get("owner") != "~caller" || query.Get("maxResults") != "7" || query.Get("nextToken") != "next page/one" {
				t.Errorf("unexpected list query: %s", request.URL.RawQuery)
			}

			if got := query["expand"]; len(got) != 2 || got[0] != expandAll || got[1] != "feature:power" {
				t.Errorf("unexpected expand options: %#v", got)
			}

			return syntheticResponse(request, http.StatusOK, `{"results":[]}`), nil
		})}),
	)

	_, err := client.GetEndpoints(context.Background(), &ListEndpointsOptions{
		Owner:      "~caller",
		Expand:     []string{expandAll, "feature:power"},
		MaxResults: 7,
		NextToken:  "next page/one",
	})
	if err != nil {
		t.Fatalf("GetEndpoints returned an error: %v", err)
	}
}
