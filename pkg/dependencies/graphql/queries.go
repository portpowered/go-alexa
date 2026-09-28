package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	genqlient "github.com/Khan/genqlient/graphql"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/internal/wire"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

const NexusGraphqlEndpoint = apiroutes.PathExecuteNexusGraphQL

// ListEndpoints executes a GraphQL query to list endpoints using genqlient
func (c *Client) ListEndpoints(ctx context.Context, input ListEndpointsInput) (*ListEndpointsResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}

	return ListEndpoints(ctx, genqlientClient, input)
}

// GetEndpoint executes a GraphQL query to get a specific endpoint using genqlient
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*GetEndpointResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}

	return GetEndpoint(ctx, genqlientClient, endpointID)
}

// ListEndpointsWithPagination executes a GraphQL query to list endpoints with pagination support
func (c *Client) ListEndpointsWithPagination(ctx context.Context, input ListEndpointsInput) (*ListEndpointsResponse, error) {
	// If pagination is disabled, enable it with default page size
	if input.PaginationParams.DisablePagination {
		input.PaginationParams.DisablePagination = false
		input.PaginationParams.PageSize = 50
	}

	return c.ListEndpoints(ctx, input)
}

func (c *Client) ListEndpointsWithoutStates(ctx context.Context, input EndpointsQueryParams) (*EndpointsResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}
	return Endpoints(ctx, genqlientClient, input)
}

// ListAllEndpoints fetches all endpoints by automatically handling pagination
func (c *Client) ListAllEndpoints(ctx context.Context, input ListEndpointsInput) ([]ListEndpointsListEndpointsListEndpointsResponseEndpointsEndpoint, error) {
	var allEndpoints []ListEndpointsListEndpointsListEndpointsResponseEndpointsEndpoint

	// Initialize pagination
	input.PaginationParams.DisablePagination = false
	if input.PaginationParams.PageSize == 0 {
		input.PaginationParams.PageSize = 50
	}

	// Make a copy of input to avoid modifying the original
	inputCopy := input
	currentInput := &inputCopy

	for {
		response, err := c.ListEndpointsWithPagination(ctx, *currentInput)
		if err != nil {
			return nil, err
		}

		allEndpoints = append(allEndpoints, response.ListEndpoints.Endpoints...)

		// Check if there's a next page
		nextToken := response.ListEndpoints.PaginationInfo.NextToken
		if nextToken == "" {
			break
		}

		// Update pagination token for next request
		currentInput.PaginationParams.NextToken = nextToken
	}

	return allEndpoints, nil
}

// ListEndpointsWithStates executes a GraphQL query to list endpoints with state information using genqlient
// This query includes detailed property states such as Volume, Power, Brightness, etc.
func (c *Client) ListEndpointsWithStates(ctx context.Context, input ListEndpointsInput) (*ListEndpointsWithStatesResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}

	return ListEndpointsWithStates(ctx, genqlientClient, input)
}

// SetEndpointFeatures executes a GraphQL mutation to set endpoint features using genqlient
func (c *Client) SetEndpointFeatures(ctx context.Context, input SetEndpointFeaturesInput) (*SetEndpointFeaturesResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}

	// The entity seemingly needs to be set for the operation to work.
	for i := range input.FeatureControlRequests {
		if input.FeatureControlRequests[i].EntityId == "" {
			trimmedId := strings.TrimPrefix(input.FeatureControlRequests[i].EndpointId, "amzn1.alexa.endpoint.")
			input.FeatureControlRequests[i].EntityId = trimmedId
		}
	}
	return SetEndpointFeatures(ctx, genqlientClient, input)
}

// RequestEndpointQualityOfService executes a GraphQL mutation to request endpoint quality of service using genqlient
func (c *Client) RequestEndpointQualityOfService(ctx context.Context, input EndpointQualityOfServiceInput) (*RequestEndpointQualityOfServiceResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}

	return RequestEndpointQualityOfService(ctx, genqlientClient, input)
}

// Subscribe executes a GraphQL mutation to subscribe to endpoint events using genqlient
func (c *Client) Subscribe(ctx context.Context, input SubscribeConfiguration) (*SubscribeResponse, error) {
	// Create a genqlient client adapter
	genqlientClient := &genqlientClientAdapter{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		getToken:   c.getToken,
	}

	return Subscribe(ctx, genqlientClient, input)
}

// genqlientClientAdapter adapts our Client to genqlient's graphql.Client interface
type genqlientClientAdapter struct {
	baseURL    string
	httpClient *http.Client
	getToken   func(ctx context.Context) (string, error)
}

func (a *genqlientClientAdapter) MakeRequest(ctx context.Context, req *genqlient.Request, resp *genqlient.Response) error {
	// Get token
	token, err := a.getToken(ctx)
	if err != nil {
		return &alexaapimodels.TokenError{
			Message: "failed to get token",
			Err:     err,
		}
	}

	// Marshal request
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return &alexaapimodels.BadRequestError{
			Message: "failed to marshal request",
			Err:     err,
		}
	}
	var wireRequest wire.WireGraphQLRequest
	if err := json.Unmarshal(bodyBytes, &wireRequest); err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to decode request", Err: err}
	}
	bodyBytes, err = json.Marshal(wireRequest)
	if err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to marshal generated request", Err: err}
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, apiroutes.MethodExecuteNexusGraphQL, a.baseURL+apiroutes.PathExecuteNexusGraphQL, bytes.NewReader(bodyBytes))
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	httpResp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "request failed",
			Err:     err,
		}
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	// Read response
	bodyBytes, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to read response",
			Err:     err,
		}
	}

	// Check status code
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return alexaapimodels.NewHTTPError(httpResp, string(bodyBytes))
	}

	// Decode through the schema-generated HTTP response model before adapting
	// it to genqlient's response type.
	var wireResponse wire.WireGraphQLResponse
	if err := json.Unmarshal(bodyBytes, &wireResponse); err != nil {
		return &alexaapimodels.BadRequestError{
			Message: "failed to decode response",
			Err:     err,
		}
	}
	bodyBytes, err = json.Marshal(wireResponse)
	if err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to marshal generated response", Err: err}
	}
	if err := json.Unmarshal(bodyBytes, resp); err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to adapt response", Err: err}
	}

	// Check for GraphQL errors
	if len(resp.Errors) > 0 {
		return &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("graphql errors: %v", resp.Errors),
		}
	}

	return nil
}
