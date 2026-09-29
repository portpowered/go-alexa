//nolint:testpackage // Exercises GraphQL transport internals against synthetic replay pairs.
package graphql

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	replay "github.com/portpowered/go-alexa/pkg/testing"
)

type graphqlReplayOperation struct {
	operation, response string
	call                func(*Client) error
}

func graphqlReplayOperations(ctx context.Context) []graphqlReplayOperation {
	return []graphqlReplayOperation{
		{"GetEndpoint", `{"data":{"endpoint":{"endpointId":"synthetic-endpoint"}}}`, func(c *Client) error {
			_, err := c.GetEndpoint(ctx, "synthetic-endpoint")

			return err
		}},
		{"ListEndpoints", `{"data":{"listEndpoints":{"endpoints":[],"paginationInfo":{"totalCount":0,"nextToken":""},"completeResult":true}}}`, func(c *Client) error {
			_, err := c.ListEndpoints(ctx, ListEndpointsInput{EndpointIds: []string{"synthetic-endpoint"}})

			return err
		}},
		{"Endpoints", `{"data":{"endpoints":{"items":[]}}}`, func(c *Client) error {
			_, err := c.ListEndpointsWithoutStates(ctx, EndpointsQueryParams{PaginationParams: PaginationParams{NextToken: "synthetic-next"}})

			return err
		}},
		{"ListEndpointsWithStates", `{"data":{"listEndpoints":{"completeResult":true,"endpoints":[]}}}`, func(c *Client) error {
			_, err := c.ListEndpointsWithStates(ctx, ListEndpointsInput{EndpointIds: []string{"synthetic-endpoint"}})

			return err
		}},
		{"SetEndpointFeatures", `{"data":{"setEndpointFeatures":{"featureControlResponses":[],"errors":[]}}}`, func(c *Client) error {
			_, err := c.SetEndpointFeatures(ctx, SetEndpointFeaturesInput{FeatureControlRequests: []FeatureControlRequest{{EndpointId: "synthetic-endpoint", FeatureName: FeatureNamePower, FeatureOperationName: FeatureOperationNameTurnon}}})

			return err
		}},
		{"RequestEndpointQualityOfService", `{"data":{"requestEndpointQualityOfService":{"durationInSeconds":45,"errors":[]}}}`, func(c *Client) error {
			_, err := c.RequestEndpointQualityOfService(ctx, EndpointQualityOfServiceInput{Endpoints: []string{"synthetic-endpoint"}, Configuration: QualityOfServiceConfiguration{DurationInSeconds: 45}})

			return err
		}},
		{"Subscribe", `{"data":{"subscribe":{"entities":["synthetic-endpoint"],"durationInMinutes":5,"errors":[]}}}`, func(c *Client) error {
			_, err := c.Subscribe(ctx, SubscribeConfiguration{Entities: []SubscriptionFilter{{EntityType: "endpoint", Ids: []string{"synthetic-endpoint"}}}, DurationInMinutes: 5})

			return err
		}},
	}
}

type graphqlReplayCapture struct {
	exchanges []replay.SyntheticExchange
}

func TestGraphQLPairedSyntheticReplay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cases := graphqlReplayOperations(ctx)

	fixture := filepath.Join("..", "..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-graphql.json")
	transport, player, capture := graphqlReplayTransport(t, fixture, cases)

	client := NewClient(WithBaseURL("https://graphql.synthetic.test"), WithBearerToken("synthetic-graphql-token"), WithHTTPClient(&http.Client{Transport: transport}))
	for _, current := range cases {
		err := current.call(client)
		if err != nil {
			t.Fatalf("%s: %v", current.operation, err)
		}
	}

	if player != nil {
		assertGraphQLReplayConsumed(ctx, t, player, client)

		return
	}

	writeGraphQLReplayFixture(t, fixture, cases, capture.exchanges)
}

func graphqlReplayTransport(
	t *testing.T,
	fixture string,
	cases []graphqlReplayOperation,
) (http.RoundTripper, *replay.SyntheticReplay, *graphqlReplayCapture) {
	t.Helper()

	if os.Getenv("UPDATE_SYNTHETIC_REPLAY") == "1" {
		capture := &graphqlReplayCapture{exchanges: make([]replay.SyntheticExchange, 0, len(cases))}

		return graphqlRecordingTransport(cases, capture), nil, capture
	}

	player, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	return player, player, nil
}

func graphqlRecordingTransport(cases []graphqlReplayOperation, capture *graphqlReplayCapture) http.RoundTripper {
	return graphqlRoundTripper(func(req *http.Request) (*http.Response, error) {
		if len(capture.exchanges) >= len(cases) {
			return nil, fmt.Errorf("unexpected GraphQL request after %d operations", len(capture.exchanges)) //nolint:err113 // Report the extra request count during fixture generation.
		}

		current := cases[len(capture.exchanges)]
		responseHeaders := http.Header{"Content-Type": []string{"application/json"}}

		pair, err := replay.RecordSyntheticExchange(current.operation, req, http.StatusOK, responseHeaders, current.response)
		if err != nil {
			return nil, fmt.Errorf("record synthetic GraphQL exchange: %w", err)
		}

		capture.exchanges = append(capture.exchanges, pair)
		response := graphqlSyntheticResponse(req, http.StatusOK, current.response)
		response.Header = responseHeaders

		return response, nil
	})
}

func assertGraphQLReplayConsumed(ctx context.Context, t *testing.T, player *replay.SyntheticReplay, client *Client) {
	t.Helper()

	err := player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetEndpoint(ctx, "synthetic-endpoint")
	if err == nil {
		t.Fatal("duplicate GraphQL request received a replay response")
	}
}

func writeGraphQLReplayFixture(
	t *testing.T,
	fixture string,
	cases []graphqlReplayOperation,
	exchanges []replay.SyntheticExchange,
) {
	t.Helper()

	if len(exchanges) != len(cases) {
		t.Fatalf("recorded %d of %d", len(exchanges), len(cases))
	}

	err := os.MkdirAll(filepath.Dir(fixture), 0o750)
	if err != nil {
		t.Fatal(err)
	}

	err = replay.WriteSyntheticReplay(fixture, exchanges)
	if err != nil {
		t.Fatal(err)
	}
}
