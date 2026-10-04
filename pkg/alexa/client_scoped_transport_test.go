//nolint:testpackage // Verifies scoped public reads using the private synthetic transport.
package alexa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestSelectedStateReadFiltersProviderResultsWithoutAccountDiscovery(t *testing.T) {
	t.Parallel()

	calls := 0

	client, err := NewClient(WithHTTPClient(&http.Client{Transport: responseTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.URL.Path != "/nexus/v1/graphql" {
			t.Fatalf("selected state read fetched account metadata: %s", request.URL.Path)
		}

		var body map[string]json.RawMessage
		decodeErr := json.NewDecoder(request.Body).Decode(&body)
		if decodeErr != nil {
			t.Fatalf("decode GraphQL request: %v", decodeErr)
		}

		var variables map[string]json.RawMessage
		decodeErr = json.Unmarshal(body["variables"], &variables)
		if decodeErr != nil {
			t.Fatalf("decode GraphQL variables: %v", decodeErr)
		}

		if !strings.Contains(string(variables["input"]), `"endpointIds":["selected"]`) {
			t.Fatalf("request did not scope the provider IDs: %s", variables["input"])
		}

		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(
			`{"data":{"listEndpoints":{"endpoints":[{"id":"selected","endpointId":"selected","features":[]},{"id":"other","endpointId":"other","features":[]}]}}}`,
		))}, nil
	})}))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	session, err := client.NewSession(WithBearerToken("synthetic-token"))
	if err != nil {
		t.Fatalf("create account session: %v", err)
	}

	t.Cleanup(func() { _ = session.Close() })

	result, err := session.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{
		EndpointIDs: []string{"selected", "selected"}, IncludeFields: &alexaapimodels.EndpointIncludeFields{Properties: true},
	})
	if err != nil || len(result.Results) != 1 || result.Results[0].EndpointID != "selected" || calls != 1 {
		t.Fatalf("scoped state read: result=%#v calls=%d err=%v", result, calls, err)
	}
}
