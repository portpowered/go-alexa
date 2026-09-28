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

func TestGraphQLPairedSyntheticReplay(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		operation, response string
		call                func(*Client) error
	}{
		{"GetEndpoint", `{"data":{"endpoint":{"endpointId":"synthetic-endpoint"}}}`, func(c *Client) error { _, err := c.GetEndpoint(ctx, "synthetic-endpoint"); return err }},
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
	fixture := filepath.Join("..", "..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-graphql.json")
	var transport http.RoundTripper
	var recorded []replay.SyntheticExchange
	var player *replay.SyntheticReplay
	if os.Getenv("UPDATE_SYNTHETIC_REPLAY") == "1" {
		transport = graphqlRoundTripper(func(req *http.Request) (*http.Response, error) {
			if len(recorded) >= len(cases) {
				return nil, fmt.Errorf("unexpected GraphQL request after %d operations", len(recorded))
			}
			current := cases[len(recorded)]
			responseHeaders := http.Header{"Content-Type": []string{"application/json"}}
			pair, err := replay.RecordSyntheticExchange(current.operation, req, 200, responseHeaders, current.response)
			if err != nil {
				return nil, err
			}
			recorded = append(recorded, pair)
			response := graphqlSyntheticResponse(req, 200, current.response)
			response.Header = responseHeaders
			return response, nil
		})
	} else {
		var err error
		player, err = replay.LoadSyntheticReplay(fixture)
		if err != nil {
			t.Fatal(err)
		}
		transport = player
	}
	client := NewClient(WithBaseURL("https://graphql.synthetic.test"), WithBearerToken("synthetic-graphql-token"), WithHTTPClient(&http.Client{Transport: transport}))
	for _, current := range cases {
		if err := current.call(client); err != nil {
			t.Fatalf("%s: %v", current.operation, err)
		}
	}
	if player != nil {
		if err := player.AssertConsumed(); err != nil {
			t.Fatal(err)
		}
		if _, err := client.GetEndpoint(ctx, "synthetic-endpoint"); err == nil {
			t.Fatal("duplicate GraphQL request received a replay response")
		}
	} else {
		if len(recorded) != len(cases) {
			t.Fatalf("recorded %d of %d", len(recorded), len(cases))
		}
		if err := os.MkdirAll(filepath.Dir(fixture), 0755); err != nil {
			t.Fatal(err)
		}
		if err := replay.WriteSyntheticReplay(fixture, recorded); err != nil {
			t.Fatal(err)
		}
	}
}
