package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestExecuteSendsBearerGraphQLRequestAndDecodesData(t *testing.T) {
	type resultShape struct {
		Endpoint struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"endpoint"`
	}

	client := NewClient(
		WithBaseURL("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-graphql-token"),
		WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.Path != "/nexus/v1/graphql" {
				t.Errorf("unexpected GraphQL target: %s %s", request.Method, request.URL)
			}
			if request.Header.Get("Authorization") != "Bearer synthetic-graphql-token" {
				t.Errorf("unexpected Authorization header: %q", request.Header.Get("Authorization"))
			}
			if request.Header.Get("Content-Type") != "application/json" || request.Header.Get("Accept") != "application/json" {
				t.Errorf("unexpected content negotiation headers: %#v", request.Header)
			}
			var body Request
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode GraphQL request: %v", err)
			}
			if body.Query != GetEndpoint_Operation || body.Variables["id"] != "endpoint-11" {
				t.Errorf("unexpected GraphQL operation: %#v", body)
			}
			return graphqlSyntheticResponse(request, http.StatusOK, `{"data":{"endpoint":{"id":"endpoint-11","name":"Synthetic lamp"}}}`), nil
		})}),
	)

	var result resultShape
	err := client.Execute(context.Background(), GetEndpoint_Operation, map[string]any{"id": "endpoint-11"}, &result)
	if err != nil {
		t.Fatalf("Execute returned an error: %v", err)
	}
	if result.Endpoint.ID != "endpoint-11" || result.Endpoint.Name != "Synthetic lamp" {
		t.Fatalf("unexpected GraphQL data: %#v", result)
	}
}

func TestQueryAndMutateDelegateToExecute(t *testing.T) {
	requests := 0
	client := NewClient(
		WithBaseURL("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-token"),
		WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
			requests++
			return graphqlSyntheticResponse(request, http.StatusOK, `{"data":{"accepted":true}}`), nil
		})}),
	)
	for _, operation := range []struct {
		name string
		run  func(context.Context, string, map[string]interface{}, interface{}) error
	}{
		{name: "query", run: client.Query},
		{name: "mutation", run: client.Mutate},
	} {
		t.Run(operation.name, func(t *testing.T) {
			var result struct {
				Accepted bool `json:"accepted"`
			}
			if err := operation.run(context.Background(), GetEndpoint_Operation, nil, &result); err != nil {
				t.Fatalf("%s returned an error: %v", operation.name, err)
			}
			if !result.Accepted {
				t.Fatalf("%s response was not decoded: %#v", operation.name, result)
			}
		})
	}
	if requests != 2 {
		t.Fatalf("expected Query and Mutate to issue two requests, got %d", requests)
	}
}

func TestExecuteReportsHTTPGraphQLDecodeAndTokenErrors(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body       string
		tokenError error
		wantText   string
	}{
		{name: "HTTP status", status: http.StatusUnauthorized, body: `{"error":"synthetic unauthorized"}`, wantText: "401"},
		{name: "invalid JSON", status: http.StatusOK, body: "not-json", wantText: "decode response"},
		{name: "GraphQL errors", status: http.StatusOK, body: `{"errors":[{"message":"synthetic query failure"}]}`, wantText: "graphql errors"},
		{name: "invalid result data", status: http.StatusOK, body: `{"data":{"count":"not-an-integer"}}`, wantText: "unmarshal result"},
		{name: "token getter", tokenError: errors.New("synthetic token failure"), wantText: "failed to get token"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			transportCalled := false
			options := []ClientOption{
				WithBaseURL("https://alexa.synthetic.test"),
				WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
					transportCalled = true
					return graphqlSyntheticResponse(request, testCase.status, testCase.body), nil
				})}),
			}
			if testCase.tokenError != nil {
				options = append(options, WithTokenGetter(func(context.Context) (string, error) { return "", testCase.tokenError }))
			} else {
				options = append(options, WithBearerToken("synthetic-token"))
			}
			client := NewClient(options...)
			var result struct {
				Count int `json:"count"`
			}
			err := client.Execute(context.Background(), GetEndpoint_Operation, nil, &result)
			if err == nil || !strings.Contains(err.Error(), testCase.wantText) {
				t.Fatalf("expected error containing %q, got %v", testCase.wantText, err)
			}
			if (testCase.tokenError != nil) == transportCalled {
				t.Fatalf("unexpected transport invocation for token error case: called=%v", transportCalled)
			}
			if testCase.tokenError != nil && !alexaapimodels.IsTokenError(err) {
				t.Fatalf("expected TokenError, got %T: %v", err, err)
			}
		})
	}
}

func TestExecuteRejectsUnschematizedGraphQLOperationBeforeTransport(t *testing.T) {
	client := NewClient(WithBearerToken("synthetic-token"), WithHTTPClient(&http.Client{
		Transport: graphqlRoundTripper(func(*http.Request) (*http.Response, error) {
			t.Fatal("unschematized operation reached transport")
			return nil, nil
		}),
	}))
	err := client.Execute(context.Background(), "query Unlisted { unknownField }", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "not in the generated schema set") {
		t.Fatalf("expected schema gate error, got %v", err)
	}
}

func TestControlPowerFeatureBuildsSyntheticFeatureMutation(t *testing.T) {
	client := NewClient(
		WithBaseURL("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-control-token"),
		WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode generated mutation request: %v", err)
			}
			if !strings.Contains(body["query"].(string), "mutation SetEndpointFeatures") {
				t.Errorf("expected SetEndpointFeatures operation, got %#v", body["query"])
			}
			variables, ok := body["variables"].(map[string]any)
			if !ok {
				t.Fatalf("missing mutation variables: %#v", body["variables"])
			}
			input, ok := variables["input"].(map[string]any)
			if !ok {
				t.Fatalf("missing feature input: %#v", variables)
			}
			requests, ok := input["featureControlRequests"].([]any)
			if !ok || len(requests) != 1 {
				t.Fatalf("unexpected feature requests: %#v", input["featureControlRequests"])
			}
			feature := requests[0].(map[string]any)
			if feature["endpointId"] != "amzn1.alexa.endpoint.synthetic" || feature["entityId"] != "synthetic" || feature["featureOperationName"] != "turnOn" {
				t.Errorf("unexpected power feature request: %#v", feature)
			}
			return graphqlSyntheticResponse(request, http.StatusOK, `{"data":{"setEndpointFeatures":{"featureControlResponses":[{"endpointId":"amzn1.alexa.endpoint.synthetic","featureName":"power","featureOperationName":"turnOn","code":"SUCCESS"}],"errors":[]}}}`), nil
		})}),
	)

	response, err := client.ControlPowerFeature(context.Background(), &alexamodels.PowerControlRequest{
		Endpoint: graphqlSyntheticEndpoint{id: "amzn1.alexa.endpoint.synthetic"},
		State:    alexaapimodels.PowerStateOn,
	})
	if err != nil {
		t.Fatalf("ControlPowerFeature returned an error: %v", err)
	}
	if response == nil || len(response.FeatureControlResponses) != 1 || response.FeatureControlResponses[0].Code != "SUCCESS" {
		t.Fatalf("unexpected feature-control result: %#v", response)
	}
}

// Keep GraphQL transport tests independent from network access and service data.
type graphqlRoundTripper func(*http.Request) (*http.Response, error)

func (roundTrip graphqlRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func graphqlSyntheticResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}

type graphqlSyntheticEndpoint struct{ id string }

func (endpoint graphqlSyntheticEndpoint) GetDeviceType() string { return "synthetic-device-type" }
func (endpoint graphqlSyntheticEndpoint) GetDeviceSerialNumber() string {
	return "synthetic-device-serial"
}
func (endpoint graphqlSyntheticEndpoint) GetLocale() string          { return "en-US" }
func (endpoint graphqlSyntheticEndpoint) GetEndpointId() string      { return endpoint.id }
func (endpoint graphqlSyntheticEndpoint) GetDeviceFamily() string    { return "synthetic-device-family" }
func (endpoint graphqlSyntheticEndpoint) GetDeviceAccountId() string { return "synthetic-account-id" }
