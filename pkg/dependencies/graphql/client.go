// Package graphql provides a GraphQL API client for interacting with Alexa services.
// It supports querying and mutating device data using GraphQL queries and mutations.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

const defaultClientTimeout = 30 * time.Second

// Client is a GraphQL API client for Alexa services.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	bearerToken string
	tokenGetter func(ctx context.Context) (string, error)
}

// ClientOption is a function that configures a GraphQL Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the GraphQL API.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBearerToken sets a bearer token directly (simple mode).
func WithBearerToken(token string) ClientOption {
	return func(c *Client) {
		c.bearerToken = token
	}
}

// WithTokenGetter sets a function to retrieve tokens dynamically.
func WithTokenGetter(getter func(ctx context.Context) (string, error)) ClientOption {
	return func(c *Client) {
		c.tokenGetter = getter
	}
}

// NewClient creates a new GraphQL API client.
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		baseURL:     alexaapimodels.AlexaAmazonBaseUriNa,
		httpClient:  &http.Client{Timeout: defaultClientTimeout},
		bearerToken: "",
		tokenGetter: nil,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Apply applies options to the client.
func (c *Client) Apply(opts ...ClientOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// Execute executes a GraphQL query or mutation.
func (c *Client) Execute(
	ctx context.Context,
	query string,
	variables map[string]interface{},
	result interface{},
) error {
	body, err := executionRequestBody(query, variables)
	if err != nil {
		return graphQLRequestFailure(0, "request_encode", &alexaapimodels.BadRequestError{Message: "operation is not schema-backed or request could not be encoded"})
	}

	request, err := http.NewRequestWithContext(
		ctx,
		apiroutes.MethodExecuteNexusGraphQL,
		c.baseURL+apiroutes.PathExecuteNexusGraphQL,
		bytes.NewReader(body),
	)
	if err != nil {
		return graphQLRequestFailure(0, "request_create", &alexaapimodels.NetworkError{Message: "failed to create request"})
	}

	token, err := c.getToken(ctx)
	if err != nil {
		return graphQLRequestFailure(0, "authentication", &alexaapimodels.TokenError{Message: "failed to get token", Err: ctx.Err()})
	}

	request.Header.Set(apiroutes.HeaderAuthorization, "Bearer "+token)
	request.Header.Set(apiroutes.HeaderContentType, "application/json")
	request.Header.Set(apiroutes.HeaderAccept, "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return graphQLRequestFailure(0, "transport", &alexaapimodels.NetworkError{Message: "request failed", Err: graphQLTransportCause(err)})
	}

	defer func() {
		_ = response.Body.Close()
	}()

	body, err = io.ReadAll(response.Body)
	if err != nil {
		return graphQLRequestFailure(response.StatusCode, "response_read", &alexaapimodels.NetworkError{Message: "failed to read response", Err: ctx.Err()})
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return graphQLRequestFailure(response.StatusCode, "provider_response",
			&alexaapimodels.HTTPError{StatusCode: response.StatusCode, Status: http.StatusText(response.StatusCode)})
	}

	return decodeExecutionResponse(body, result, response.StatusCode)
}

// Query executes a GraphQL query.
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	return c.Execute(ctx, query, variables, result)
}

// Mutate executes a GraphQL mutation.
func (c *Client) Mutate(
	ctx context.Context,
	mutation string,
	variables map[string]interface{},
	result interface{},
) error {
	return c.Execute(ctx, mutation, variables, result)
}

// getToken retrieves the bearer token.
func (c *Client) getToken(ctx context.Context) (string, error) {
	if c.bearerToken != "" {
		return c.bearerToken, nil
	}

	if c.tokenGetter != nil {
		return c.tokenGetter(ctx)
	}

	return "", &alexaapimodels.TokenError{
		Message: "no token available",
	}
}

func executionRequestBody(
	query string,
	variables map[string]interface{},
) ([]byte, error) {
	if !schemaBackedGraphQLOperation(query) {
		return nil, &alexaapimodels.BadRequestError{Message: "GraphQL operation is not in the generated schema set"}
	}

	wireRequest := alexamodels.WireGraphQLRequest{
		OperationName:        nil,
		Query:                query,
		Variables:            nil,
		AdditionalProperties: nil,
	}
	if len(variables) > 0 {
		wireRequest.Variables = variables
	}

	body, err := json.Marshal(wireRequest)
	if err != nil {
		return nil, &alexaapimodels.BadRequestError{Message: "failed to marshal request", Err: err}
	}

	return body, nil
}

func decodeExecutionResponse(body []byte, result interface{}, status int) error {
	var wireResponse alexamodels.WireGraphQLResponse

	err := json.Unmarshal(body, &wireResponse)
	if err != nil {
		return graphQLRequestFailure(status, "response_decode", &alexaapimodels.BadRequestError{Message: "failed to decode response"})
	}

	if wireResponse.Errors != nil && len(*wireResponse.Errors) > 0 {
		return graphQLRequestFailure(status, "graphql_response", &alexaapimodels.BadRequestError{Message: "provider rejected GraphQL operation"})
	}

	if result == nil || wireResponse.Data == nil {
		return nil
	}

	data, err := json.Marshal(wireResponse.Data)
	if err != nil {
		return graphQLRequestFailure(status, "response_decode", &alexaapimodels.BadRequestError{Message: "failed to marshal data"})
	}

	err = json.Unmarshal(data, result)
	if err != nil {
		return graphQLRequestFailure(status, "response_decode", &alexaapimodels.BadRequestError{Message: "failed to unmarshal result"})
	}

	return nil
}
