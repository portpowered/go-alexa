// Package rest provides a REST API client for interacting with Alexa services.
// It supports device enumeration, control, and state management via HTTP/RESTful APIs.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

const defaultClientTimeout = 30 * time.Second

// defaultRetryDelay waits one second after the first failed attempt and two after the second.
func defaultRetryDelay(attempt int) time.Duration { return time.Duration(attempt+1) * time.Second }

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
	// retryDelay is the wait before retrying a request whose attempt (0-based) failed
	retryDelay func(attempt int) time.Duration
	// webOrigin is the selected region's Alexa web site, sent as Origin and Referer and
	// used to recognize CSRF cookies; base URI overrides do not change it.
	webOrigin regionalWebOrigin
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

func assignOptional[T any](destination *T, value *T) {
	if value != nil {
		*destination = *value
	}
}

// regionalWebOrigin is the Alexa web site a browser session for one region runs on.
type regionalWebOrigin struct {
	origin          string
	cookieDomain    string
	cookieSubdomain string
}

// regionWebOrigin returns the Alexa web site of a supported region. Any other value gets
// North America, the client's default.
func regionWebOrigin(region alexaapimodels.Region) regionalWebOrigin {
	switch region {
	case alexaapimodels.RegionEU:
		return regionalWebOrigin{
			origin:          alexaapimodels.AlexaAmazonBaseUriEu,
			cookieDomain:    alexamodels.AuthCSRFCookieDomainEu,
			cookieSubdomain: alexamodels.AuthCSRFCookieSubdomainEu,
		}
	case alexaapimodels.RegionJP:
		return regionalWebOrigin{
			origin:          alexaapimodels.AlexaAmazonBaseUriJp,
			cookieDomain:    alexamodels.AuthCSRFCookieDomainJp,
			cookieSubdomain: alexamodels.AuthCSRFCookieSubdomainJp,
		}
	case alexaapimodels.RegionUS:
		return regionalWebOrigin{
			origin:          alexaapimodels.AlexaAmazonBaseUriNa,
			cookieDomain:    alexamodels.AuthCSRFCookieDomain,
			cookieSubdomain: alexamodels.AuthCSRFCookieSubdomain,
		}
	}

	return regionWebOrigin(alexaapimodels.RegionUS)
}

// referer is the browser Referer of this web site.
func (origin regionalWebOrigin) referer() string {
	return origin.origin + alexamodels.AuthWebRefererPath
}

// isCSRFCookieKey reports whether a cookie map key names this web site's CSRF cookie.
func (origin regionalWebOrigin) isCSRFCookieKey(key string) bool {
	return key == origin.cookieDomain+":"+alexamodels.AuthCSRFCookieName ||
		key == origin.cookieSubdomain+":"+alexamodels.AuthCSRFCookieName
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithRetryDelay sets the wait before retrying a request that failed with a transport
// error or a 5xx response; attempt is the 0-based index of the failed attempt.
func WithRetryDelay(delay func(attempt int) time.Duration) ClientOption {
	return func(c *Client) {
		c.retryDelay = delay
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
			client.webOrigin = regionWebOrigin(region)
		case alexaapimodels.RegionEU:
			client.amazonalexaapiBaseURI = alexaapimodels.ApiServiceUriEu
			client.amazonapiBaseURI = alexaapimodels.AmazonApiServiceUriEu
			client.alexaAmazonBaseURI = alexaapimodels.AlexaAmazonBaseUriEu
			client.webOrigin = regionWebOrigin(region)
		case alexaapimodels.RegionJP:
			client.amazonalexaapiBaseURI = alexaapimodels.ApiServiceUriJp
			client.amazonapiBaseURI = alexaapimodels.AmazonApiServiceUriJp
			client.alexaAmazonBaseURI = alexaapimodels.AlexaAmazonBaseUriJp
			client.webOrigin = regionWebOrigin(region)
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
		webOrigin:             regionWebOrigin(alexaapimodels.RegionUS),
		tokenGetter:           nil,
		cookies:               make(map[string]*http.Cookie),
		csrfToken:             "",
		csrfTokenGetter:       nil,
		refreshToken:          "",
		useCookieAuth:         false,
		customerID:            "",
		retryDelay:            defaultRetryDelay,
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
			if cookie.Name == alexamodels.AuthCSRFCookieName {
				c.csrfToken = cookie.Value

				return cookie.Value, nil
			}
			// Also check key format
			if c.webOrigin.isCSRFCookieKey(key) {
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

		if cookie.Name == alexamodels.AuthCSRFCookieName || c.webOrigin.isCSRFCookieKey(key) {
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
	req.Header.Set(apiroutes.HeaderReferer, c.webOrigin.referer())
	req.Header.Set(apiroutes.HeaderOrigin, c.webOrigin.origin)

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

const (
	diagnosticBodyLimit     = 16 * 1024
	authenticationBodyLimit = 64 * 1024
)

// doRequest performs an HTTP request with retry logic.
//
//nolint:dupl // The route inventory requires request construction and injected send in the same function.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, requestFailure(method, path, 0, "", "request_encode",
				&alexaapimodels.BadRequestError{Message: "failed to marshal request body"})
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	url := c.amazonalexaapiBaseURI + path

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, requestFailure(method, path, 0, "", "request_create",
			&alexaapimodels.NetworkError{Message: "failed to create request"})
	}

	// Get token and set authorization header
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, requestFailure(method, path, 0, "", "authentication",
			tokenFailureCause(ctx, err))
	}

	req.Header.Set(apiroutes.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(apiroutes.HeaderContentType, "application/json")
	req.Header.Set(apiroutes.HeaderAccept, "application/json")

	const maxAttempts = 3
	for attempt := range maxAttempts {
		response, err := c.httpClient.Do(req)
		if err == nil && response.StatusCode < http.StatusInternalServerError {
			return response, nil
		}

		if attempt == maxAttempts-1 {
			if err != nil {
				return nil, requestFailure(method, url, 0, "", "transport", networkFailureCause(ctx, err, "request failed after retries"))
			}

			return response, nil
		}

		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}

		// A canceled request never retries: select picks among ready cases at random, so a
		// short delay alone could still send the request again after cancellation.
		if ctx.Err() != nil {
			return nil, requestFailure(method, url, 0, "", "transport", networkFailureCause(ctx, ctx.Err(), "request retry canceled"))
		}

		select {
		case <-ctx.Done():
			return nil, requestFailure(method, url, 0, "", "transport", networkFailureCause(ctx, ctx.Err(), "request retry canceled"))
		case <-time.After(c.retryDelay(attempt)):
		}

		if req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, requestFailure(method, url, 0, "", "request_encode",
					&alexaapimodels.BadRequestError{Message: "failed to recreate request body"})
			}

			req.Body = body
		}
	}

	return nil, requestFailure(method, url, 0, "", "transport", &alexaapimodels.NetworkError{Message: "provider returned no response"})
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
		bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, diagnosticBodyLimit))
		reason := providerReason(bodyBytes)

		stage := "provider_response"
		if readErr != nil {
			stage = "response_read"
		}

		return requestFailure(method, path, resp.StatusCode, reason, stage,
			&alexaapimodels.HTTPError{StatusCode: resp.StatusCode, Status: http.StatusText(resp.StatusCode)})
	}

	if result != nil {
		err := json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return requestFailure(method, path, resp.StatusCode, "", "response_decode",
				&alexaapimodels.BadRequestError{Message: "failed to decode response"})
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
			return nil, requestFailure(method, fullURL, 0, "", "request_encode", &alexaapimodels.BadRequestError{Message: "failed to marshal request body"})
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, requestFailure(method, fullURL, 0, "", "request_create", &alexaapimodels.NetworkError{Message: "failed to create request"})
	}

	err = c.setFullURLAuthentication(ctx, req.Header, useCookieAuth)
	if err != nil {
		return nil, requestFailure(method, fullURL, 0, "", "authentication", tokenFailureCause(ctx, err))
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

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, requestFailure(method, fullURL, 0, "", "transport", networkFailureCause(ctx, err, "request failed"))
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fullURLResponseError(method, fullURL, response)
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

func (c *Client) setFullURLAuthentication(ctx context.Context, headers http.Header, useCookieAuth bool) error {
	if !useCookieAuth || !c.useCookieAuth || len(c.cookies) == 0 {
		token, err := c.getToken(ctx)
		if err != nil {
			return &alexaapimodels.TokenError{
				Message: "failed to get token",
				Err:     nil,
			}
		}

		headers.Set(apiroutes.HeaderAuthorization, "Bearer "+token)

		return nil
	}

	cookiePairs := make([]string, 0, len(c.cookies)+1)
	for _, cookie := range c.cookies {
		cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}

	if len(cookiePairs) > 0 {
		headers.Set(apiroutes.HeaderCookie, strings.Join(cookiePairs, "; "))
	}

	csrfToken, err := c.cachedCSRFToken()
	if err != nil {
		return err
	}

	if csrfToken != "" {
		headers.Set(apiroutes.HeaderCsrf, csrfToken)
	}

	headers.Set(apiroutes.HeaderUserAgent, "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:1.0) bash-script/1.0")
	headers.Set(apiroutes.HeaderDNT, "1")
	headers.Set(apiroutes.HeaderReferer, c.webOrigin.referer())
	headers.Set(apiroutes.HeaderOrigin, c.webOrigin.origin)

	return nil
}

func fullURLResponseError(method, fullURL string, response *http.Response) error {
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(response.Body, diagnosticBodyLimit))

	stage := "provider_response"

	if err != nil {
		stage = "response_read"
	}

	return requestFailure(method, fullURL, response.StatusCode, providerReason(body), stage, statusCause(response.StatusCode))
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

	if result != nil {
		err := json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return requestFailure(method, fullURL, resp.StatusCode, "", "response_decode", &alexaapimodels.BadRequestError{Message: "failed to decode response"})
		}
	}

	return nil
}

// doUnauthenticatedRequest performs an HTTP request without authentication headers
// This is used for authentication endpoints that don't require a bearer token.
//
//nolint:dupl // The route inventory requires request construction and injected send in the same function.
func (c *Client) doUnauthenticatedRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, &alexaapimodels.BadRequestError{
				Message: "failed to marshal request body",
				Err:     nil,
			}
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     nil,
		}
	}

	req.Header.Set(apiroutes.HeaderContentType, "application/json")
	req.Header.Set(apiroutes.HeaderAccept, "application/json")

	const maxAttempts = 3
	for attempt := range maxAttempts {
		response, err := c.httpClient.Do(req)
		if err == nil && response.StatusCode < http.StatusInternalServerError {
			return response, nil
		}

		if attempt == maxAttempts-1 {
			if err != nil {
				return nil, requestFailure(method, url, 0, "", "transport", networkFailureCause(ctx, err, "request failed after retries"))
			}

			return response, nil
		}

		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}

		// A canceled request never retries: select picks among ready cases at random, so a
		// short delay alone could still send the request again after cancellation.
		if ctx.Err() != nil {
			return nil, requestFailure(method, url, 0, "", "transport", networkFailureCause(ctx, ctx.Err(), "request retry canceled"))
		}

		select {
		case <-ctx.Done():
			return nil, requestFailure(method, url, 0, "", "transport", networkFailureCause(ctx, ctx.Err(), "request retry canceled"))
		case <-time.After(c.retryDelay(attempt)):
		}

		if req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, requestFailure(method, url, 0, "", "request_encode",
					&alexaapimodels.BadRequestError{Message: "failed to recreate request body"})
			}

			req.Body = body
		}
	}

	return nil, requestFailure(method, url, 0, "", "transport", &alexaapimodels.NetworkError{Message: "provider returned no response"})
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

	bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, authenticationBodyLimit))
	if readErr != nil {
		return &alexaapimodels.NetworkError{Message: "failed to read authentication response", Err: safeTransportCause(readErr)}
	}

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
				Err:     nil,
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
	return fmt.Sprintf("request failed with status %d", e.StatusCode)
}
