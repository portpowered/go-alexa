package graphql

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	genqlient "github.com/Khan/genqlient/graphql"
)

type recordingGeneratedClient struct {
	responseData string
	request      *genqlient.Request
	calls        int
	err          error
}

func (client *recordingGeneratedClient) MakeRequest(_ context.Context, request *genqlient.Request, response *genqlient.Response) error {
	client.calls++
	client.request = request
	if client.err != nil {
		return client.err
	}
	if client.responseData == "" || response.Data == nil {
		return nil
	}
	return json.Unmarshal([]byte(client.responseData), response.Data)
}

func TestGeneratedOperationWrappersBuildRequestsAndDecodeSyntheticData(t *testing.T) {
	tests := []struct {
		name           string
		operation      string
		responseData   string
		variableMarker string
		call           func(context.Context, genqlient.Client) error
	}{
		{
			name:           "get endpoint",
			operation:      "GetEndpoint",
			responseData:   `{"endpoint":{"endpointId":"synthetic-endpoint","friendlyName":"Synthetic outlet"}}`,
			variableMarker: "synthetic-endpoint-id",
			call: func(ctx context.Context, client genqlient.Client) error {
				response, err := GetEndpoint(ctx, client, "synthetic-endpoint-id")
				if err == nil {
					endpoint := response.GetEndpoint()
					if endpoint.GetEndpointId() != "synthetic-endpoint" {
						t.Errorf("unexpected generated endpoint response: %#v", response)
					}
				}
				return err
			},
		},
		{
			name:           "list endpoints",
			operation:      "ListEndpoints",
			responseData:   `{"listEndpoints":{"endpoints":[],"paginationInfo":{"totalCount":0,"nextToken":"synthetic-next"},"completeResult":true}}`,
			variableMarker: "synthetic-endpoint-id",
			call: func(ctx context.Context, client genqlient.Client) error {
				_, err := ListEndpoints(ctx, client, ListEndpointsInput{
					EndpointIds:      []string{"synthetic-endpoint-id"},
					PaginationParams: PaginationParams{NextToken: "synthetic-next"},
				})
				return err
			},
		},
		{
			name:           "list endpoints with states",
			operation:      "ListEndpointsWithStates",
			responseData:   `{"listEndpoints":{"completeResult":true,"endpoints":[]}}`,
			variableMarker: "synthetic-endpoint-id",
			call: func(ctx context.Context, client genqlient.Client) error {
				_, err := ListEndpointsWithStates(ctx, client, ListEndpointsInput{EndpointIds: []string{"synthetic-endpoint-id"}})
				return err
			},
		},
		{
			name:           "quality of service",
			operation:      "RequestEndpointQualityOfService",
			responseData:   `{"requestEndpointQualityOfService":{"durationInSeconds":45,"errors":[]}}`,
			variableMarker: "synthetic-endpoint-id",
			call: func(ctx context.Context, client genqlient.Client) error {
				_, err := RequestEndpointQualityOfService(ctx, client, EndpointQualityOfServiceInput{
					Endpoints:     []string{"synthetic-endpoint-id"},
					Configuration: QualityOfServiceConfiguration{DurationInSeconds: 45},
				})
				return err
			},
		},
		{
			name:           "set endpoint features",
			operation:      "SetEndpointFeatures",
			responseData:   `{"setEndpointFeatures":{"featureControlResponses":[],"errors":[]}}`,
			variableMarker: "synthetic-endpoint-id",
			call: func(ctx context.Context, client genqlient.Client) error {
				_, err := SetEndpointFeatures(ctx, client, SetEndpointFeaturesInput{
					FeatureControlRequests: []FeatureControlRequest{{
						EndpointId:           "synthetic-endpoint-id",
						FeatureName:          FeatureNamePower,
						FeatureOperationName: FeatureOperationNameTurnon,
					}},
				})
				return err
			},
		},
		{
			name:           "subscribe",
			operation:      "Subscribe",
			responseData:   `{"subscribe":{"entities":["synthetic-endpoint-id"],"durationInMinutes":5,"errors":[]}}`,
			variableMarker: "synthetic-endpoint-id",
			call: func(ctx context.Context, client genqlient.Client) error {
				_, err := Subscribe(ctx, client, SubscribeConfiguration{
					Entities:          []SubscriptionFilter{{EntityType: "endpoint", Ids: []string{"synthetic-endpoint-id"}}},
					DurationInMinutes: 5,
				})
				return err
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			client := &recordingGeneratedClient{responseData: testCase.responseData}
			if err := testCase.call(context.Background(), client); err != nil {
				t.Fatalf("generated wrapper returned an error: %v", err)
			}
			if client.calls != 1 || client.request == nil {
				t.Fatalf("generated wrapper did not make exactly one request: calls=%d request=%#v", client.calls, client.request)
			}
			if client.request.OpName != testCase.operation || strings.TrimSpace(client.request.Query) == "" {
				t.Fatalf("unexpected generated operation: name=%q query-empty=%v", client.request.OpName, strings.TrimSpace(client.request.Query) == "")
			}
			variables, err := json.Marshal(client.request.Variables)
			if err != nil {
				t.Fatalf("marshal generated variables: %v", err)
			}
			if !strings.Contains(string(variables), testCase.variableMarker) {
				t.Errorf("generated variables omitted input marker %q: %s", testCase.variableMarker, variables)
			}
		})
	}
}

func TestGeneratedOperationWrapperReturnsClientError(t *testing.T) {
	wantErr := &syntheticGraphQLError{message: "synthetic transport rejection"}
	client := &recordingGeneratedClient{err: wantErr}
	_, err := GetEndpoint(context.Background(), client, "synthetic-endpoint-id")
	if err != wantErr {
		t.Fatalf("expected generated wrapper to return client error unchanged, got %v", err)
	}
}

type syntheticGraphQLError struct{ message string }

func (err *syntheticGraphQLError) Error() string { return err.message }
