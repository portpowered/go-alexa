// Package rest provides a REST API client for interacting with Alexa services.
// It supports device enumeration, control, and state management via HTTP/RESTful APIs.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// Client is a REST API client for Alexa services
type Client struct {
	httpClient            *http.Client
	bearerToken           string
	amazonalexaapiBaseUri string
	amazonapiBaseUri      string
	alexaAmazonBaseUri    string
	tokenGetter           func(ctx context.Context) (string, error)
	// Cookie-based authentication support
	cookies         map[string]*http.Cookie
	csrfToken       string
	csrfTokenGetter func(ctx context.Context) (string, error)
	refreshToken    string
	useCookieAuth   bool
	// customer ID (certain requests necessitate a customer ID to be set)
	customerID string
}

func convertWireModel[T any](source any) (*T, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func valueOrZero[T any](value *T) T {
	if value != nil {
		return *value
	}
	var zero T
	return zero
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithRegion(region alexaapimodels.Region) ClientOption {
	return func(c *Client) {
		switch region {
		case alexaapimodels.RegionUS:
			c.amazonalexaapiBaseUri = alexaapimodels.ApiServiceUriNa
			c.amazonapiBaseUri = alexaapimodels.AmazonApiServiceUriNa
			c.alexaAmazonBaseUri = alexaapimodels.AlexaAmazonBaseUriNa
		case alexaapimodels.RegionEU:
			c.amazonalexaapiBaseUri = alexaapimodels.ApiServiceUriEu
			c.amazonapiBaseUri = alexaapimodels.AmazonApiServiceUriEu
			c.alexaAmazonBaseUri = alexaapimodels.AlexaAmazonBaseUriEu
		case alexaapimodels.RegionJP:
			c.amazonalexaapiBaseUri = alexaapimodels.ApiServiceUriJp
			c.amazonapiBaseUri = alexaapimodels.AmazonApiServiceUriJp
			c.alexaAmazonBaseUri = alexaapimodels.AlexaAmazonBaseUriJp
		}
	}
}

// WithAmazonalexaAPIBaseURI sets the base URL for the Amazon Alexa API
func WithAmazonalexaAPIBaseURI(baseURL string) ClientOption {
	return func(c *Client) {
		c.amazonalexaapiBaseUri = baseURL
	}
}

// WithAmazonapiBaseURI sets the base URL for the Amazon API
func WithAmazonapiBaseURI(baseURL string) ClientOption {
	return func(c *Client) {
		c.amazonapiBaseUri = baseURL
	}
}

// WithAlexaAmazonBaseURI sets the base URL for the Alexa Amazon web domain
func WithAlexaAmazonBaseURI(baseURL string) ClientOption {
	return func(c *Client) {
		c.alexaAmazonBaseUri = baseURL
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

// WithRefreshToken sets a refresh token for cookie-based authentication
func WithRefreshToken(refreshToken string) ClientOption {
	return func(c *Client) {
		c.refreshToken = refreshToken
		c.useCookieAuth = true
	}
}

// WithCSRFToken sets a CSRF token for cookie-based authentication
func WithCSRFToken(csrfToken string) ClientOption {
	return func(c *Client) {
		c.csrfToken = csrfToken
		c.useCookieAuth = true
	}
}

func WithCustomerID(customerID string) ClientOption {
	return func(c *Client) {
		c.customerID = customerID
	}
}

// WithCSRFTokenGetter sets a function to retrieve CSRF tokens dynamically
func WithCSRFTokenGetter(getter func(ctx context.Context) (string, error)) ClientOption {
	return func(c *Client) {
		c.csrfTokenGetter = getter
		c.useCookieAuth = true
	}
}

// WithCookies sets cookies for cookie-based authentication
func WithCookies(cookies map[string]*http.Cookie) ClientOption {
	return func(c *Client) {
		c.cookies = cookies
		c.useCookieAuth = true
	}
}

// NewClient creates a new REST API client
func NewClient(opts ...ClientOption) *Client {
	jar, _ := cookiejar.New(nil)
	client := &Client{
		amazonalexaapiBaseUri: alexaapimodels.ApiServiceUriNa,
		amazonapiBaseUri:      alexaapimodels.AmazonApiServiceUriNa,
		alexaAmazonBaseUri:    alexaapimodels.AlexaAmazonBaseUriNa,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		cookies: make(map[string]*http.Cookie),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// apply applies the options to the client
func (c *Client) Apply(opts ...ClientOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// getToken retrieves the bearer token, either from direct token or token getter
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

// GetCSRFToken retrieves a CSRF token for cookie-based authentication
func (c *Client) GetCSRFToken(ctx context.Context) (string, error) {
	if c.csrfToken != "" {
		return c.csrfToken, nil
	}

	if c.csrfTokenGetter != nil {
		return c.csrfTokenGetter(ctx)
	}

	// Try to get CSRF from cookies if available
	if c.cookies != nil {
		// Check for csrf cookie
		for key, cookie := range c.cookies {
			if cookie.Name == "csrf" {
				c.csrfToken = cookie.Value
				return cookie.Value, nil
			}
			// Also check key format
			if key == "alexa.amazon.com:csrf" || key == ".alexa.amazon.com:csrf" {
				c.csrfToken = cookie.Value
				return cookie.Value, nil
			}
		}
	}

	// If we have cookies, try to get CSRF by making a request
	if len(c.cookies) > 0 {
		csrfToken, err := c.fetchCSRFTokenFromAPI(ctx)
		if err == nil && csrfToken != "" {
			c.csrfToken = csrfToken
			return csrfToken, nil
		}
	}

	return "", &alexaapimodels.TokenError{
		Message: "no CSRF token available",
	}
}

// InitializeCookieAuth exchanges refresh token for cookies and retrieves CSRF token
// This should be called before making requests if using cookie-based authentication
func (c *Client) InitializeCookieAuth(ctx context.Context) error {
	if c.refreshToken == "" {
		return &alexaapimodels.BadRequestError{
			Message: "refresh token is required for cookie authentication",
		}
	}

	// Determine domain
	domain := "amazon.com"
	if c.amazonapiBaseUri != "" {
		// Extract domain from URI
		if len(c.amazonapiBaseUri) > 8 {
			domain = c.amazonapiBaseUri[8:] // Skip "https://"
			if idx := len(domain) - 1; idx >= 0 && domain[idx] == '/' {
				domain = domain[:idx]
			}
			// Remove "api." prefix if present
			if len(domain) > 4 && domain[:4] == "api." {
				domain = domain[4:]
			}
		}
	}

	// Exchange refresh token for cookies
	cookies, err := c.ExchangeRefreshTokenForCookies(ctx, c.refreshToken, domain)
	if err != nil {
		return err
	}

	// Store cookies
	c.cookies = cookies

	return nil
}

// fetchCSRFTokenFromAPI fetches CSRF token by visiting Alexa API endpoints
func (c *Client) fetchCSRFTokenFromAPI(ctx context.Context) (string, error) {
	// Try multiple endpoints to get CSRF token
	endpoints := []string{
		c.alexaAmazonBaseUri + "/api/language",
	}

	for _, endpoint := range endpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			continue
		}

		// Add cookies to request
		// Use cookie-based authentication
		cookiePairs := make([]string, 0, len(c.cookies)+1)
		for _, cookie := range c.cookies {
			cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
		// Set the combined cookie header directly to preserve original bytes
		if len(cookiePairs) > 0 {
			req.Header.Set("Cookie", strings.Join(cookiePairs, "; "))
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:1.0) bash-script/1.0")
		req.Header.Set("DNT", "1")
		req.Header.Set("Referer", "https://alexa.amazon.com/spa/index.html")
		req.Header.Set("Origin", "https://alexa.amazon.com")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		// Check response cookies for CSRF token
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "csrf" {
				_ = resp.Body.Close()
				return cookie.Value, nil
			}
		}
		_ = resp.Body.Close()
	}

	return "", errors.New("failed to retrieve CSRF token from any endpoint")
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, &alexaapimodels.BadRequestError{
				Message: "failed to marshal request body",
				Err:     err,
			}
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	url := c.amazonalexaapiBaseUri + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	// Get token and set authorization header
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, &alexaapimodels.TokenError{
			Message: "failed to get token",
			Err:     err,
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Retry logic
	maxRetries := 3
	var resp *http.Response
	for i := 0; i < maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < maxRetries-1 {
			// Exponential backoff
			backoff := time.Duration(i+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "request failed after retries",
			Err:     err,
		}
	}

	return resp, nil
}

// doJSONRequest performs a request and unmarshals the JSON response
func (c *Client) doJSONRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	resp, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return alexaapimodels.NewHTTPError(resp, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to decode response",
				Err:     err,
			}
		}
	}

	return nil
}

// doRequestWithFullURL performs an HTTP request with a full URL and optional custom headers
// This is used for endpoints that don't use the standard base URI
// useCookieAuth controls whether cookie-based authentication should be used for this request
func (c *Client) doRequestWithFullURL(ctx context.Context, method, fullURL string, body interface{}, customHeaders map[string]string, useCookieAuth bool) (*http.Response, error) {
	// Initialize cookie auth if needed and requested
	if useCookieAuth && c.useCookieAuth && c.refreshToken != "" && len(c.cookies) == 0 {
		if err := c.InitializeCookieAuth(ctx); err != nil {
			return nil, err
		}
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, &alexaapimodels.BadRequestError{
				Message: "failed to marshal request body",
				Err:     err,
			}
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	// Set authentication headers
	if useCookieAuth && c.useCookieAuth && len(c.cookies) > 0 {
		// Use cookie-based authentication
		cookiePairs := make([]string, 0, len(c.cookies)+1)
		for _, cookie := range c.cookies {
			cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
		// Set the combined cookie header directly to preserve original bytes
		if len(cookiePairs) > 0 {
			req.Header.Set("Cookie", strings.Join(cookiePairs, "; "))
		}
		// Ensure CSRF token cookie is present
		csrfToken, err := c.GetCSRFToken(ctx)
		if err != nil {
			return nil, err
		}
		if csrfToken != "" {
			// cookiePairs = append(cookiePairs, fmt.Sprintf("csrf=%s", csrfToken))
			req.Header.Set("csrf", csrfToken)
		}

		// Set standard headers for cookie-based auth
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:1.0) bash-script/1.0")
		req.Header.Set("DNT", "1")
		req.Header.Set("Referer", "https://alexa.amazon.com/spa/index.html")
		req.Header.Set("Origin", "https://alexa.amazon.com")

	} else {
		// Use bearer token authentication
		token, err := c.getToken(ctx)
		if err != nil {
			return nil, &alexaapimodels.TokenError{
				Message: "failed to get token",
				Err:     err,
			}
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	// Set custom headers if provided (these override defaults)
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	// Retry logic
	maxRetries := 3
	var resp *http.Response
	for i := 0; i < maxRetries; i++ {
		resp, err = c.httpClient.Do(req)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		if err != nil {
			return nil, &alexaapimodels.NetworkError{
				Message: "request failed",
				Err:     err,
			}
		}

		if resp.StatusCode == 401 {
			return nil, &alexaapimodels.UnauthorizedError{
				Message: "unauthorized",
				Err:     errors.New("unauthorized"),
			}
		}

		if resp.StatusCode == 404 {
			return nil, &alexaapimodels.NotFoundError{
				Message: "not found",
				Err:     errors.New("not found on API call"),
			}
		}
		if err == nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, &alexaapimodels.BadRequestError{
				Message: "bad request",
				Err:     errors.New("bad request on API call"),
			}
		}
		if err == nil && resp.StatusCode >= 500 {
			return nil, &alexaapimodels.InternalServerError{
				Message: "internal server error",
				Err:     errors.New("internal server error on API call"),
			}
		}

		if i < maxRetries-1 {
			// Exponential backoff
			backoff := time.Duration(i+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "request failed after retries",
			Err:     err,
		}
	}

	return resp, nil
}

// doJSONRequestWithFullURL performs a request with a full URL and unmarshals the JSON response
// useCookieAuth controls whether cookie-based authentication should be used for this request
func (c *Client) doJSONRequestWithFullURL(ctx context.Context, method, fullURL string, body interface{}, customHeaders map[string]string, result interface{}, useCookieAuth bool) error {
	resp, err := c.doRequestWithFullURL(ctx, method, fullURL, body, customHeaders, useCookieAuth)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &alexaapimodels.BadRequestError{
			Message: "bad request",
			Err:     errors.New(string(bodyBytes)),
		}
	}

	if resp.StatusCode >= 500 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &alexaapimodels.InternalServerError{
			Message: "internal server error",
			Err:     errors.New(string(bodyBytes)),
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to decode response",
				Err:     err,
			}
		}
	}

	return nil
}

// doUnauthenticatedRequest performs an HTTP request without authentication headers
// This is used for authentication endpoints that don't require a bearer token
func (c *Client) doUnauthenticatedRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, &alexaapimodels.BadRequestError{
				Message: "failed to marshal request body",
				Err:     err,
			}
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Retry logic
	maxRetries := 3
	var resp *http.Response
	for i := 0; i < maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < maxRetries-1 {
			// Exponential backoff
			backoff := time.Duration(i+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "request failed after retries",
			Err:     err,
		}
	}

	return resp, nil
}

// doUnauthenticatedJSONRequest performs an unauthenticated request and unmarshals the JSON response
func (c *Client) doUnauthenticatedJSONRequest(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	resp, err := c.doUnauthenticatedRequest(ctx, method, url, body)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// For authentication endpoints, we need to handle different status codes
	// 200 = success, 400/401 = challenge or error
	if resp.StatusCode < 200 || resp.StatusCode >= 500 {
		return &UnauthenticatedRequestError{
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	// For 400/401, we might have a challenge response, but we'll let the caller handle it
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return &UnauthenticatedRequestError{
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	if result != nil {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to decode response",
				Err:     err,
			}
		}
	}

	return nil
}

// UnauthenticatedRequestError represents an error from an unauthenticated request
type UnauthenticatedRequestError struct {
	StatusCode int
	Body       string
}

func (e *UnauthenticatedRequestError) Error() string {
	return fmt.Sprintf("request failed with status %d: %s", e.StatusCode, e.Body)
}
