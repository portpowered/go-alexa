//nolint:testpackage // Verifies private query conversion and generated operation behavior.
package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type queryOperationTestCase struct {
	name         string
	queryPrefix  string
	response     string
	variableText string
	call         func(*Client) error
}

func TestClientQueryMethodsIssueGeneratedOperations(t *testing.T) {
	t.Parallel()

	tests := []queryOperationTestCase{
		{
			name: "get endpoint", queryPrefix: "query GetEndpoint", response: `{"endpoint":{"endpointId":"synthetic-endpoint"}}`, variableText: "synthetic-endpoint-id",
			call: func(client *Client) error {
				_, err := client.GetEndpoint(context.Background(), "synthetic-endpoint-id")

				return err
			},
		},
		{
			name: "list endpoints", queryPrefix: "query ListEndpoints", response: `{"listEndpoints":{"endpoints":[],"paginationInfo":{"totalCount":0,"nextToken":""},"completeResult":true}}`, variableText: "synthetic-endpoint-id",
			call: func(client *Client) error {
				_, err := client.ListEndpoints(context.Background(), ListEndpointsInput{EndpointIds: []string{"synthetic-endpoint-id"}})

				return err
			},
		},
		{
			name: "list endpoints with pagination", queryPrefix: "query ListEndpoints", response: `{"listEndpoints":{"endpoints":[],"paginationInfo":{"totalCount":0,"nextToken":""},"completeResult":true}}`, variableText: "\"pageSize\":50",
			call: func(client *Client) error {
				_, err := client.ListEndpointsWithPagination(context.Background(), ListEndpointsInput{PaginationParams: PaginationParams{DisablePagination: true}})

				return err
			},
		},
		{
			name: "list endpoints without states", queryPrefix: "query Endpoints", response: `{"endpoints":{"items":[]}}`, variableText: "synthetic-next-token",
			call: func(client *Client) error {
				_, err := client.ListEndpointsWithoutStates(context.Background(), EndpointsQueryParams{PaginationParams: PaginationParams{NextToken: "synthetic-next-token"}})

				return err
			},
		},
		{
			name: "list endpoints with states", queryPrefix: "query ListEndpointsWithStates", response: `{"listEndpoints":{"completeResult":true,"endpoints":[]}}`, variableText: "synthetic-endpoint-id",
			call: func(client *Client) error {
				_, err := client.ListEndpointsWithStates(context.Background(), ListEndpointsInput{EndpointIds: []string{"synthetic-endpoint-id"}})

				return err
			},
		},
		{
			name: "quality of service", queryPrefix: "mutation RequestEndpointQualityOfService", response: `{"requestEndpointQualityOfService":{"durationInSeconds":45,"errors":[]}}`, variableText: "synthetic-endpoint-id",
			call: func(client *Client) error {
				_, err := client.RequestEndpointQualityOfService(context.Background(), EndpointQualityOfServiceInput{Endpoints: []string{"synthetic-endpoint-id"}, Configuration: QualityOfServiceConfiguration{DurationInSeconds: 45}})

				return err
			},
		},
		{
			name: "subscribe", queryPrefix: "mutation Subscribe", response: `{"subscribe":{"entities":["synthetic-endpoint-id"],"durationInMinutes":5,"errors":[]}}`, variableText: "synthetic-endpoint-id",
			call: func(client *Client) error {
				_, err := client.Subscribe(context.Background(), SubscribeConfiguration{Entities: []SubscriptionFilter{{EntityType: "endpoint", Ids: []string{"synthetic-endpoint-id"}}}, DurationInMinutes: 5})

				return err
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			called := false

			client := NewClient(
				WithBaseURL("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-query-token"),
				WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
					called = true

					return assertGeneratedQueryRequest(t, testCase, request), nil
				})}),
			)

			err := testCase.call(client)
			if err != nil {
				t.Fatalf("query method returned an error: %v", err)
			}

			if !called {
				t.Fatal("query method did not submit an HTTP request")
			}
		})
	}
}

func assertGeneratedQueryRequest(
	t *testing.T,
	testCase queryOperationTestCase,
	request *http.Request,
) *http.Response {
	t.Helper()

	if request.URL.Path != NexusGraphqlEndpoint || request.Header.Get("Authorization") != "Bearer synthetic-query-token" {
		t.Errorf("unexpected GraphQL request target or authorization: %s %q", request.URL, request.Header.Get("Authorization"))
	}

	var body map[string]any

	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		t.Fatalf("decode generated GraphQL request: %v", err)
	}

	query, _ := body["query"].(string)
	if !strings.HasPrefix(strings.TrimSpace(query), testCase.queryPrefix) {
		t.Errorf("unexpected operation query: %q", query[:min(len(query), 80)])
	}

	variables, err := json.Marshal(body["variables"])
	if err != nil {
		t.Fatalf("marshal GraphQL variables: %v", err)
	}

	if !strings.Contains(string(variables), testCase.variableText) {
		t.Errorf("variables omitted %q: %s", testCase.variableText, variables)
	}

	return graphqlSyntheticResponse(request, http.StatusOK, `{"data":`+testCase.response+`}`)
}

func TestListAllEndpointsContinuesUntilPaginationTokenIsEmpty(t *testing.T) {
	t.Parallel()

	calls := 0
	client := NewClient(
		WithBaseURL("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-pagination-token"),
		WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
			calls++
			var body map[string]any
			err := json.NewDecoder(request.Body).Decode(&body)
			if err != nil {
				t.Fatalf("decode page request: %v", err)
			}
			variables, variablesAreMap := body["variables"].(map[string]any)
			if !variablesAreMap {
				t.Fatalf("query variables are not an object: %#v", body["variables"])
			}
			input, inputIsMap := variables["input"].(map[string]any)
			if !inputIsMap {
				t.Fatalf("query input is not an object: %#v", variables["input"])
			}
			pagination, paginationIsMap := input["paginationParams"].(map[string]any)
			if !paginationIsMap {
				t.Fatalf("pagination parameters are not an object: %#v", input["paginationParams"])
			}
			if pagination["disablePagination"] != false || pagination["pageSize"] != float64(50) {
				t.Errorf("unexpected page configuration: %#v", pagination)
			}
			wantToken := ""
			responseToken := "synthetic-next-token"
			if calls == 2 {
				wantToken = "synthetic-next-token"
				responseToken = ""
			}
			if pagination["nextToken"] != wantToken {
				t.Errorf("page %d used nextToken %#v, want %#v", calls, pagination["nextToken"], wantToken)
			}

			return graphqlSyntheticResponse(request, http.StatusOK, `{"data":{"listEndpoints":{"endpoints":[],"paginationInfo":{"totalCount":0,"nextToken":"`+responseToken+`"},"completeResult":true}}}`), nil
		})}),
	)

	endpoints, err := client.ListAllEndpoints(context.Background(), ListEndpointsInput{})
	if err != nil {
		t.Fatalf("ListAllEndpoints returned an error: %v", err)
	}

	if calls != 2 || len(endpoints) != 0 {
		t.Fatalf("expected two empty synthetic pages, calls=%d endpoints=%#v", calls, endpoints)
	}
}

func TestClientQueryAdapterReturnsHTTPAndGraphQLErrors(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		status int
		body   string
	}{
		{name: "HTTP failure", status: http.StatusForbidden, body: `{"message":"synthetic forbidden"}`},
		{name: "GraphQL failure", status: http.StatusOK, body: `{"errors":[{"message":"synthetic query failure"}]}`},
		{name: "invalid response", status: http.StatusOK, body: "not-json"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithBaseURL("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-query-token"),
				WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
					return graphqlSyntheticResponse(request, testCase.status, testCase.body), nil
				})}),
			)
			{
				_, err := client.GetEndpoint(context.Background(), "synthetic-endpoint-id")
				if err == nil {
					t.Fatal("expected query adapter error")
				}
			}
		})
	}
}

func TestResponseConvertersMapFeatureQoSAndSubscriptionData(t *testing.T) {
	t.Parallel()

	assertFeatureResponseConversion(t)
	assertQoSConversions(t)
	assertSubscriptionConversions(t)
}

func assertFeatureResponseConversion(t *testing.T) {
	t.Helper()

	var featureResponse SetEndpointFeaturesResponse

	err := json.Unmarshal([]byte(`{"setEndpointFeatures":{"featureControlResponses":[{"endpointId":"synthetic-endpoint","featureName":"power","instance":"main","featureOperationName":"turnOn","code":"SUCCESS"}],"errors":[{"endpointId":"other-endpoint","featureName":"power","instance":"main","featureOperationName":"turnOff","code":"REJECTED","message":"synthetic rejection"}]}}`), &featureResponse)
	if err != nil {
		t.Fatalf("decode feature response: %v", err)
	}

	convertedFeature := ConvertToFeatureControlResponse(&featureResponse)
	if convertedFeature == nil || len(convertedFeature.FeatureControlResponses) != 1 || convertedFeature.FeatureControlResponses[0].EndpointID != "synthetic-endpoint" || len(convertedFeature.Errors) != 1 || convertedFeature.Errors[0].Message != "synthetic rejection" {
		t.Fatalf("unexpected feature conversion: %#v", convertedFeature)
	}

	if ConvertToFeatureControlResponse(nil) != nil {
		t.Fatal("nil feature response should convert to nil")
	}
}

func assertQoSConversions(t *testing.T) {
	t.Helper()

	qosInput := ConvertQualityOfServiceRequest(&alexaapimodels.QualityOfServiceRequest{
		Endpoints: []string{"synthetic-endpoint"},
		Configuration: alexaapimodels.QualityOfServiceConfiguration{
			TypeOfExperience:  alexaapimodels.QualityOfServiceExperienceForegroundPersistent,
			DurationInSeconds: 30,
		},
	})
	if len(qosInput.Endpoints) != 1 || qosInput.Configuration.DurationInSeconds != 30 || string(qosInput.Configuration.TypeOfExperience) != "FOREGROUND_PERSISTENT" {
		t.Fatalf("unexpected QoS request conversion: %#v", qosInput)
	}

	var qosResponse RequestEndpointQualityOfServiceResponse

	err := json.Unmarshal([]byte(`{"requestEndpointQualityOfService":{"durationInSeconds":30,"errors":[{"type":"LIMIT","message":"synthetic limit"}]}}`), &qosResponse)
	if err != nil {
		t.Fatalf("decode QoS response: %v", err)
	}

	convertedQoS := ConvertToQualityOfServiceResponse(&qosResponse)
	if convertedQoS.DurationInSeconds != 30 || len(convertedQoS.Errors) != 1 || convertedQoS.Errors[0].Type != "LIMIT" || ConvertToQualityOfServiceResponse(nil) != nil {
		t.Fatalf("unexpected QoS response conversion: %#v", convertedQoS)
	}
}

func assertSubscriptionConversions(t *testing.T) {
	t.Helper()

	subscribeInput := ConvertSubscribeRequest(&alexaapimodels.SubscribeRequest{
		Entities:          []alexaapimodels.SubscribeEntity{{EntityType: alexaapimodels.EntityTypeEndpoint}, {EntityType: alexaapimodels.EntityTypeState}},
		DurationInMinutes: 5,
	})
	if subscribeInput.Format != SubscriptionEventFormatGraphql || !subscribeInput.Reset || subscribeInput.Version != "2" || subscribeInput.DurationInMinutes != 5 || len(subscribeInput.Entities) != 2 {
		t.Fatalf("unexpected subscription conversion: %#v", subscribeInput)
	}

	if got := ConvertSubscribeEntities(nil); len(got) != 0 {
		t.Fatalf("nil entity list should convert to empty: %#v", got)
	}

	var subscribeResponse SubscribeResponse

	err := json.Unmarshal([]byte(`{"subscribe":{"entities":["synthetic-endpoint"],"durationInMinutes":5,"errors":[{"type":"LIMIT","message":"synthetic limit"}]}}`), &subscribeResponse)
	if err != nil {
		t.Fatalf("decode subscription response: %v", err)
	}

	convertedSubscribe := ConvertToSubscribeResponse(&subscribeResponse)
	if len(convertedSubscribe.Entities) != 1 || convertedSubscribe.DurationInMinutes != 5 || len(convertedSubscribe.Errors) != 1 || convertedSubscribe.Errors[0].Message != "synthetic limit" || ConvertToSubscribeResponse(nil) != nil {
		t.Fatalf("unexpected subscription response conversion: %#v", convertedSubscribe)
	}
}

func TestClientApplyAndTokenGetterContext(t *testing.T) {
	t.Parallel()

	type contextKey string

	ctx := context.WithValue(context.Background(), contextKey("synthetic"), "token-request")
	client := NewClient()
	client.Apply(WithTokenGetter(func(got context.Context) (string, error) {
		if got.Value(contextKey("synthetic")) != "token-request" {
			t.Errorf("token getter did not receive caller context")
		}

		return "synthetic-context-token", nil
	}))

	token, err := client.getToken(ctx)
	if err != nil || token != "synthetic-context-token" {
		t.Fatalf("token getter returned %q, %v", token, err)
	}

	{
		_, err := NewClient().getToken(context.Background())
		if !alexaapimodels.IsTokenError(err) {
			t.Fatalf("missing token configuration should return TokenError, got %v", err)
		}
	}
}
