// Package graphql provides a GraphQL API client for interacting with Alexa services.
// It supports querying and mutating device data using GraphQL queries and mutations.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/internal/wire"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

// Client is a GraphQL API client for Alexa services
type Client struct {
	baseURL     string
	httpClient  *http.Client
	bearerToken string
	tokenGetter func(ctx context.Context) (string, error)
}

// ClientOption is a function that configures a GraphQL Client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the GraphQL API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBearerToken sets a bearer token directly (simple mode)
func WithBearerToken(token string) ClientOption {
	return func(c *Client) {
		c.bearerToken = token
	}
}

// WithTokenGetter sets a function to retrieve tokens dynamically
func WithTokenGetter(getter func(ctx context.Context) (string, error)) ClientOption {
	return func(c *Client) {
		c.tokenGetter = getter
	}
}

// apply applies the options to the client
func (c *Client) Apply(opts ...ClientOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// NewClient creates a new GraphQL API client
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		baseURL: alexaapimodels.AlexaAmazonBaseUriNa,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// getToken retrieves the bearer token
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

// Request represents a GraphQL request
type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Response represents a GraphQL response
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []Error     `json:"errors,omitempty"`
}

// Error represents a GraphQL error
type Error struct {
	Message    string                 `json:"message"`
	Locations  []Location             `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Location represents an error location in a GraphQL query
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Execute executes a GraphQL query or mutation
func (c *Client) Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	if !schemaBackedGraphQLOperation(query) {
		return &alexaapimodels.BadRequestError{Message: "GraphQL operation is not in the generated schema set"}
	}
	wireRequest := wire.WireGraphQLRequest{
		Query: query,
	}
	if len(variables) > 0 {
		wireRequest.Variables = &variables
	}
	bodyBytes, err := json.Marshal(wireRequest)
	if err != nil {
		return &alexaapimodels.BadRequestError{
			Message: "failed to marshal request",
			Err:     err,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, apiroutes.MethodExecuteNexusGraphQL, c.baseURL+apiroutes.PathExecuteNexusGraphQL, bytes.NewReader(bodyBytes))
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	// Get token and set authorization header
	token, err := c.getToken(ctx)
	if err != nil {
		return &alexaapimodels.TokenError{
			Message: "failed to get token",
			Err:     err,
		}
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "request failed",
			Err:     err,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to read response",
			Err:     err,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return alexaapimodels.NewHTTPError(resp, string(bodyBytes))
	}

	var wireResponse wire.WireGraphQLResponse
	if err := json.Unmarshal(bodyBytes, &wireResponse); err != nil {
		return &alexaapimodels.BadRequestError{
			Message: "failed to decode response",
			Err:     err,
		}
	}
	encodedResponse, err := json.Marshal(wireResponse)
	if err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to marshal response", Err: err}
	}
	var graphqlResp Response
	if err := json.Unmarshal(encodedResponse, &graphqlResp); err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to convert response", Err: err}
	}

	if len(graphqlResp.Errors) > 0 {
		return &alexaapimodels.BadRequestError{
			Message: fmt.Sprintf("graphql errors: %v", graphqlResp.Errors),
		}
	}

	if result != nil && graphqlResp.Data != nil {
		dataBytes, err := json.Marshal(graphqlResp.Data)
		if err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to marshal data",
				Err:     err,
			}
		}
		if err := json.Unmarshal(dataBytes, result); err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to unmarshal result",
				Err:     err,
			}
		}
	}

	return nil
}

// Query executes a GraphQL query
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	return c.Execute(ctx, query, variables, result)
}

// Mutate executes a GraphQL mutation
func (c *Client) Mutate(ctx context.Context, mutation string, variables map[string]interface{}, result interface{}) error {
	return c.Execute(ctx, mutation, variables, result)
}
