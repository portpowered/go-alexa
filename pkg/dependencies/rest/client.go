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
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

const defaultClientTimeout = 30 * time.Second

// Client is a REST API client for Alexa services.
type Client struct {
	httpClient            *http.Client
	bearerToken           string
	amazonalexaapiBaseURI string
	amazonapiBaseURI      string
	alexaAmazonBaseURI    string
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
		return nil, fmt.Errorf("marshal wire model: %w", err)
	}

	var result T
	{
		err := json.Unmarshal(data, &result)
		if err != nil {
			return nil, fmt.Errorf("unmarshal wire model: %w", err)
		}
	}

	return &result, nil
}

//nolint:ireturn // This generic helper returns the caller's selected concrete value type.
func valueOrZero[T any](value *T) T {
	if value != nil {
		return *value
	}

	var zero T

	return zero
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithRegion selects the Alexa service region.
func WithRegion(region alexaapimodels.Region) ClientOption {
	return func(client *Client) {
		switch region {
		case alexaapimodels.RegionUS:
			client.amazonalexaapiBaseURI = alexaapimodels.ApiServiceUriNa
			client.amazonapiBaseURI = alexaapimodels.AmazonApiServiceUriNa
			client.alexaAmazonBaseURI = alexaapimodels.AlexaAmazonBaseUriNa
		case alexaapimodels.RegionEU:
			client.amazonalexaapiBaseURI = alexaapimodels.ApiServiceUriEu
			client.amazonapiBaseURI = alexaapimodels.AmazonApiServiceUriEu
			client.alexaAmazonBaseURI = alexaapimodels.AlexaAmazonBaseUriEu
		case alexaapimodels.RegionJP:
			client.amazonalexaapiBaseURI = alexaapimodels.ApiServiceUriJp
			client.amazonapiBaseURI = alexaapimodels.AmazonApiServiceUriJp
			client.alexaAmazonBaseURI = alexaapimodels.AlexaAmazonBaseUriJp
		}
	}
}

// WithAmazonalexaAPIBaseURI sets the base URL for the Amazon Alexa API.
func WithAmazonalexaAPIBaseURI(baseURL string) ClientOption {
	return func(c *Client) {
		c.amazonalexaapiBaseURI = baseURL
	}
}

// WithAmazonapiBaseURI sets the base URL for the Amazon API.
func WithAmazonapiBaseURI(baseURL string) ClientOption {
	return func(c *Client) {
		c.amazonapiBaseURI = baseURL
	}
}

// WithAlexaAmazonBaseURI sets the base URL for the Alexa Amazon web domain.
func WithAlexaAmazonBaseURI(baseURL string) ClientOption {
	return func(c *Client) {
		c.alexaAmazonBaseURI = baseURL
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

// WithRefreshToken sets a refresh token for cookie-based authentication.
func WithRefreshToken(refreshToken string) ClientOption {
	return func(c *Client) {
		c.refreshToken = refreshToken
		c.useCookieAuth = true
	}
}

// WithCSRFToken sets a CSRF token for cookie-based authentication.
func WithCSRFToken(csrfToken string) ClientOption {
	return func(c *Client) {
		c.csrfToken = csrfToken
		c.useCookieAuth = true
	}
}

// WithCustomerID sets the customer ID used by customer-scoped requests.
func WithCustomerID(customerID string) ClientOption {
	return func(c *Client) {
		c.customerID = customerID
	}
}

// WithCSRFTokenGetter sets a function to retrieve CSRF tokens dynamically.
func WithCSRFTokenGetter(getter func(ctx context.Context) (string, error)) ClientOption {
	return func(c *Client) {
		c.csrfTokenGetter = getter
		c.useCookieAuth = true
	}
}

// WithCookies sets cookies for cookie-based authentication.
func WithCookies(cookies map[string]*http.Cookie) ClientOption {
	return func(c *Client) {
		c.cookies = cookies
		c.useCookieAuth = true
	}
}

// NewClient creates a new REST API client.
func NewClient(opts ...ClientOption) *Client {
	jar, _ := cookiejar.New(nil)
	client := &Client{
		bearerToken:           "",
		amazonalexaapiBaseURI: alexaapimodels.ApiServiceUriNa,
		amazonapiBaseURI:      alexaapimodels.AmazonApiServiceUriNa,
		alexaAmazonBaseURI:    alexaapimodels.AlexaAmazonBaseUriNa,
		tokenGetter:           nil,
		cookies:               make(map[string]*http.Cookie),
		csrfToken:             "",
		csrfTokenGetter:       nil,
		refreshToken:          "",
		useCookieAuth:         false,
		customerID:            "",
		httpClient: &http.Client{
			Timeout: defaultClientTimeout,
			Jar:     jar,
		},
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

// GetCSRFToken retrieves a CSRF token for cookie-based authentication.
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
// This should be called before making requests if using cookie-based authentication.
func (c *Client) InitializeCookieAuth(ctx context.Context) error {
	if c.refreshToken == "" {
		return &alexaapimodels.BadRequestError{
			Message: "refresh token is required for cookie authentication",
		}
	}

	domain := cookieAuthDomain(c.amazonapiBaseURI)

	// Exchange refresh token for cookies
	cookies, err := c.ExchangeRefreshTokenForCookies(ctx, c.refreshToken, domain)
	if err != nil {
		return err
	}

	// Store cookies
	c.cookies = cookies

	return nil
}

func cookieAuthDomain(baseURI string) string {
	if len(baseURI) <= len("https://") {
		return "amazon.com"
	}

	domain := baseURI[len("https://"):]
	if idx := len(domain) - 1; idx >= 0 && domain[idx] == '/' {
		domain = domain[:idx]
	}

	domain = strings.TrimPrefix(domain, "api.")

	return domain
}

// getToken retrieves the bearer token, either from direct token or token getter.
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

// cachedCSRFToken returns already configured CSRF material without making an
// HTTP request. Request methods use this helper so credential setup stays explicit.
func (c *Client) cachedCSRFToken() (string, error) {
	if c.csrfToken != "" {
		return c.csrfToken, nil
	}

	for key, cookie := range c.cookies {
		if cookie == nil {
			continue
		}

		if cookie.Name == "csrf" || key == "alexa.amazon.com:csrf" || key == ".alexa.amazon.com:csrf" {
			return cookie.Value, nil
		}
	}

	return "", &alexaapimodels.TokenError{Message: "no CSRF token available; call GetCSRFToken explicitly or configure one"}
}

// fetchCSRFTokenFromAPI fetches CSRF token by visiting Alexa API endpoints.
func (c *Client) fetchCSRFTokenFromAPI(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, apiroutes.MethodFetchCsrfCookie, c.alexaAmazonBaseURI+apiroutes.PathFetchCsrfCookie, nil)
	if err != nil {
		return "", fmt.Errorf("create CSRF request: %w", err)
	}

	// Add cookies to request, preserving their stored bytes.
	cookiePairs := make([]string, 0, len(c.cookies)+1)
	for _, cookie := range c.cookies {
		cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	if len(cookiePairs) > 0 {
		req.Header.Set(apiroutes.HeaderCookie, strings.Join(cookiePairs, "; "))
	}

	req.Header.Set(apiroutes.HeaderUserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:1.0) bash-script/1.0")
	req.Header.Set(apiroutes.HeaderDNT, "1")
	req.Header.Set(apiroutes.HeaderReferer, "https://alexa.amazon.com/spa/index.html")
	req.Header.Set(apiroutes.HeaderOrigin, "https://alexa.amazon.com")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch CSRF cookie: %w", err)
	}

	// Check response cookies for CSRF token
	for _, cookie := range resp.Cookies() {
		if cookie.Name == apiroutes.HeaderCsrf {
			_ = resp.Body.Close()

			return cookie.Value, nil
		}
	}

	_ = resp.Body.Close()

	return "", errFailedToRetrieveCSRFToken
}

// doRequest performs an HTTP request with retry logic.
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

	url := c.amazonalexaapiBaseURI + path

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

	req.Header.Set(apiroutes.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(apiroutes.HeaderContentType, "application/json")
	req.Header.Set(apiroutes.HeaderAccept, "application/json")

	// Retry logic
	maxRetries := 3

	var resp *http.Response

	for attempt := range maxRetries {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if attempt < maxRetries-1 {
			// Exponential backoff
			backoff := time.Duration(attempt+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("request retry canceled: %w", ctx.Err())
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

// doJSONRequest performs a request and unmarshals the JSON response.
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
		err := json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
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
// useCookieAuth controls whether cookie-based authentication should be used for this request.
func (c *Client) doRequestWithFullURL(
	ctx context.Context,
	method string,
	fullURL string,
	body interface{},
	customHeaders map[string]string,
	useCookieAuth bool,
) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		var err error

		bodyReader, err = marshalFullURLBody(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	err = c.setFullURLAuthentication(ctx, req, useCookieAuth)
	if err != nil {
		return nil, err
	}

	req.Header.Set(apiroutes.HeaderAccept, "application/json")

	if body != nil {
		req.Header.Set(apiroutes.HeaderContentType, "application/json; charset=UTF-8")
	}

	for key, value := range customHeaders {
		if !apiroutes.IsKnownRequestHeader(key) {
			return nil, &alexaapimodels.BadRequestError{Message: "request header is not declared in the schema"}
		}

		req.Header.Set(key, value)
	}

	const maxRetries = 3

	var response *http.Response

	var requestErr error

	for attempt := range maxRetries {
		response, requestErr = c.httpClient.Do(req)
		if requestErr != nil {
			return nil, &alexaapimodels.NetworkError{Message: "request failed", Err: requestErr}
		}

		if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
			return response, nil
		}

		statusErr := fullURLResponseError(response)
		if statusErr != nil {
			return nil, statusErr
		}

		if attempt < maxRetries-1 {
			backoff := time.Duration(attempt+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("request retry canceled: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}
	}

	if requestErr != nil {
		return nil, &alexaapimodels.NetworkError{Message: "request failed after retries", Err: requestErr}
	}

	return response, nil
}

func marshalFullURLBody(body interface{}) (io.Reader, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "failed to marshal request body",
			Err:     err,
		}
	}

	return bytes.NewReader(bodyBytes), nil
}

func (c *Client) setFullURLAuthentication(ctx context.Context, req *http.Request, useCookieAuth bool) error {
	if !useCookieAuth || !c.useCookieAuth || len(c.cookies) == 0 {
		token, err := c.getToken(ctx)
		if err != nil {
			return &alexaapimodels.TokenError{
				Message: "failed to get token",
				Err:     err,
			}
		}

		req.Header.Set(apiroutes.HeaderAuthorization, "Bearer "+token)

		return nil
	}

	cookiePairs := make([]string, 0, len(c.cookies)+1)
	for _, cookie := range c.cookies {
		cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	if len(cookiePairs) > 0 {
		req.Header.Set(apiroutes.HeaderCookie, strings.Join(cookiePairs, "; "))
	}

	csrfToken, err := c.cachedCSRFToken()
	if err != nil {
		return err
	}

	if csrfToken != "" {
		req.Header.Set(apiroutes.HeaderCsrf, csrfToken)
	}

	req.Header.Set(apiroutes.HeaderUserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:1.0) bash-script/1.0")
	req.Header.Set(apiroutes.HeaderDNT, "1")
	req.Header.Set(apiroutes.HeaderReferer, "https://alexa.amazon.com/spa/index.html")
	req.Header.Set(apiroutes.HeaderOrigin, "https://alexa.amazon.com")

	return nil
}

func fullURLResponseError(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return &alexaapimodels.UnauthorizedError{Message: "unauthorized", Err: errUnauthorizedResponse}
	case http.StatusNotFound:
		return &alexaapimodels.NotFoundError{Message: "not found", Err: errNotFoundResponse}
	}

	if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError {
		return &alexaapimodels.BadRequestError{Message: "bad request", Err: errBadRequestResponse}
	}

	if response.StatusCode >= http.StatusInternalServerError {
		return &alexaapimodels.InternalServerError{Message: "internal server error", Err: errServerErrorResponse}
	}

	return nil
}

// doJSONRequestWithFullURL performs a request with a full URL and unmarshals the JSON response
// useCookieAuth controls whether cookie-based authentication should be used for this request.
func (c *Client) doJSONRequestWithFullURL(
	ctx context.Context,
	method string,
	fullURL string,
	body interface{},
	customHeaders map[string]string,
	result interface{},
	useCookieAuth bool,
) error {
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
			Err:     errors.New(string(bodyBytes)), //nolint:err113 // Preserve the provider response body verbatim in its error text.
		}
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		bodyBytes, _ := io.ReadAll(resp.Body)

		return &alexaapimodels.InternalServerError{
			Message: "internal server error",
			Err:     errors.New(string(bodyBytes)), //nolint:err113 // Preserve the provider response body verbatim in its error text.
		}
	}

	if result != nil {
		err := json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to decode response",
				Err:     err,
			}
		}
	}

	return nil
}

// doUnauthenticatedRequest performs an HTTP request without authentication headers
// This is used for authentication endpoints that don't require a bearer token.
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

	req.Header.Set(apiroutes.HeaderContentType, "application/json")
	req.Header.Set(apiroutes.HeaderAccept, "application/json")

	// Retry logic
	maxRetries := 3

	var resp *http.Response

	for attempt := range maxRetries {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if attempt < maxRetries-1 {
			// Exponential backoff
			backoff := time.Duration(attempt+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("request retry canceled: %w", ctx.Err())
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

// doUnauthenticatedJSONRequest performs an unauthenticated request and unmarshals the JSON response.
//
//nolint:unparam // The explicit generated method lets apiroutes verify every caller's schema operation.
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
		err := json.Unmarshal(bodyBytes, result)
		if err != nil {
			return &alexaapimodels.BadRequestError{
				Message: "failed to decode response",
				Err:     err,
			}
		}
	}

	return nil
}

// UnauthenticatedRequestError represents an error from an unauthenticated request.
type UnauthenticatedRequestError struct {
	StatusCode int
	Body       string
}

func (e *UnauthenticatedRequestError) Error() string {
	return fmt.Sprintf("request failed with status %d: %s", e.StatusCode, e.Body)
}
