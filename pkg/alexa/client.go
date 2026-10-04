// Package alexa provides the main client for interacting with Alexa services.
// It orchestrates authentication, API clients (REST and GraphQL), and event connections.
// It supports code-pair registration, bearer tokens, and explicit token refresh.
// Browser authorization and callback validation belong to the calling application.
package alexa

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	"github.com/portpowered/go-alexa/pkg/dependencies/rest"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

const defaultClientTimeout = 30 * time.Second

// Client contains immutable service endpoints and network transports. It is
// safe to reuse for multiple Alexa accounts; account credentials live on Session.
type Client struct {
	region            alexaapimodels.Region
	apiBaseURL        string
	amazonBaseURL     string
	webBaseURL        string
	graphqlBaseURL    string
	eventAuthority    string
	restHTTPClient    *http.Client
	graphqlHTTPClient *http.Client
	eventHTTPClient   *http.Client
	eventTransport    http.RoundTripper
}

// Session contains account credentials and owns event connections created from it.
type Session struct {
	client           *Client
	restClient       *rest.Client
	graphqlClient    *graphql.Client
	mu               sync.RWMutex
	accessToken      string
	refreshToken     string
	csrfToken        string
	cookies          map[string]*http.Cookie
	customerID       string
	closed           bool
	eventConnections map[*HTTP2Connection]struct{}
}

// Option configures a reusable client. Options are validated when NewClient runs.
type Option func(*clientConfig) error

// SessionOption configures credentials and account-specific state.
type SessionOption func(*sessionConfig) error

type clientConfig struct {
	region                                                                alexaapimodels.Region
	apiBaseURL, amazonBaseURL, webBaseURL, graphqlBaseURL, eventAuthority string
	commonHTTPClient, restHTTPClient, graphqlHTTPClient, eventHTTPClient  *http.Client
	eventTransport                                                        http.RoundTripper
	timeout                                                               time.Duration
	set                                                                   map[string]string
}

type sessionConfig struct {
	accessToken, refreshToken, csrfToken, customerID string
	cookies                                          map[string]*http.Cookie
	set                                              map[string]string
	cookiesSet                                       bool
}

// NewClient creates a reusable service client from validated endpoint and transport options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := clientConfig{
		region:            alexaapimodels.RegionUS,
		apiBaseURL:        "",
		amazonBaseURL:     "",
		webBaseURL:        "",
		graphqlBaseURL:    "",
		eventAuthority:    "",
		commonHTTPClient:  nil,
		restHTTPClient:    nil,
		graphqlHTTPClient: nil,
		eventHTTPClient:   nil,
		eventTransport:    nil,
		timeout:           defaultClientTimeout,
		set:               make(map[string]string),
	}

	for _, opt := range opts {
		if opt == nil {
			return nil, errClientOptionNil
		}

		err := opt(&cfg)
		if err != nil {
			return nil, fmt.Errorf("invalid client option: %w", err)
		}
	}

	defaults, err := regionEndpoints(cfg.region)
	if err != nil {
		return nil, err
	}

	if cfg.apiBaseURL == "" {
		cfg.apiBaseURL = defaults.api
	}

	if cfg.amazonBaseURL == "" {
		cfg.amazonBaseURL = defaults.amazon
	}

	if cfg.webBaseURL == "" {
		cfg.webBaseURL = defaults.web
	}

	if cfg.graphqlBaseURL == "" {
		cfg.graphqlBaseURL = defaults.graphql
	}

	if cfg.eventAuthority == "" {
		cfg.eventAuthority = defaults.events
	}

	common := cloneHTTPClient(cfg.commonHTTPClient)
	if common == nil {
		common = &http.Client{}
	}

	restClient := cloneHTTPClient(cfg.restHTTPClient)
	if restClient == nil {
		restClient = cloneHTTPClient(common)
	}

	graphqlClient := cloneHTTPClient(cfg.graphqlHTTPClient)
	if graphqlClient == nil {
		graphqlClient = cloneHTTPClient(common)
	}

	if cfg.timeout > 0 {
		if cfg.restHTTPClient == nil {
			restClient.Timeout = cfg.timeout
		}

		if cfg.graphqlHTTPClient == nil {
			graphqlClient.Timeout = cfg.timeout
		}
	}

	eventClient := cloneHTTPClient(cfg.eventHTTPClient)
	if cfg.eventTransport != nil {
		if eventClient != nil {
			return nil, errConflictingEventOptions
		}

		eventClient = &http.Client{Transport: cfg.eventTransport}
	}

	return &Client{region: cfg.region, apiBaseURL: cfg.apiBaseURL, amazonBaseURL: cfg.amazonBaseURL,
		webBaseURL: cfg.webBaseURL, graphqlBaseURL: cfg.graphqlBaseURL, eventAuthority: cfg.eventAuthority,
		restHTTPClient: restClient, graphqlHTTPClient: graphqlClient, eventHTTPClient: eventClient,
		eventTransport: cfg.eventTransport}, nil
}

func cloneHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		return nil
	}

	clientCopy := *client

	return &clientCopy
}

type endpointSet struct{ api, amazon, web, graphql, events string }

func regionEndpoints(region alexaapimodels.Region) (endpointSet, error) {
	switch region {
	case alexaapimodels.RegionUS:
		return endpointSet{
			alexaapimodels.ApiServiceUriNa,
			alexaapimodels.AmazonApiServiceUriNa,
			alexaapimodels.AlexaAmazonBaseUriNa,
			alexaapimodels.AlexaAmazonBaseUriNa,
			alexaapimodels.Http2ConnectionUriNa,
		}, nil
	case alexaapimodels.RegionEU:
		return endpointSet{
			alexaapimodels.ApiServiceUriEu,
			alexaapimodels.AmazonApiServiceUriEu,
			alexaapimodels.AlexaAmazonBaseUriEu,
			alexaapimodels.AlexaAmazonBaseUriEu,
			alexaapimodels.Http2ConnectionUriEu,
		}, nil
	case alexaapimodels.RegionJP:
		return endpointSet{
			alexaapimodels.ApiServiceUriJp,
			alexaapimodels.AmazonApiServiceUriJp,
			alexaapimodels.AlexaAmazonBaseUriJp,
			alexaapimodels.AlexaAmazonBaseUriJp,
			alexaapimodels.Http2ConnectionUriJp,
		}, nil
	default:
		return endpointSet{}, fmt.Errorf("%w %q", errUnsupportedAlexaRegion, region)
	}
}

func setOptionValue(cfg *clientConfig, name, value string) error {
	if previous, ok := cfg.set[name]; ok && previous != value {
		return fmt.Errorf("%w %s", errConflictingValues, name)
	}

	cfg.set[name] = value

	return nil
}

// NewSession creates account state isolated from other sessions using this Client.
func (c *Client) NewSession(opts ...SessionOption) (*Session, error) {
	if c == nil {
		return nil, errClientNil
	}

	cfg := sessionConfig{
		accessToken:  "",
		refreshToken: "",
		csrfToken:    "",
		customerID:   "",
		cookies:      make(map[string]*http.Cookie),
		set:          make(map[string]string),
		cookiesSet:   false,
	}

	for _, opt := range opts {
		if opt == nil {
			return nil, errSessionOptionNil
		}

		err := opt(&cfg)
		if err != nil {
			return nil, fmt.Errorf("invalid session option: %w", err)
		}
	}

	session := &Session{
		client:           c,
		restClient:       nil,
		graphqlClient:    nil,
		mu:               sync.RWMutex{},
		accessToken:      cfg.accessToken,
		refreshToken:     cfg.refreshToken,
		csrfToken:        cfg.csrfToken,
		cookies:          cfg.cookies,
		customerID:       cfg.customerID,
		closed:           false,
		eventConnections: make(map[*HTTP2Connection]struct{}),
	}
	session.restClient = c.newRESTClient()
	session.restClient.Apply(rest.WithTokenGetter(session.accessTokenForRequest), rest.WithCustomerID(session.customerID))

	if len(session.cookies) > 0 {
		session.restClient.Apply(rest.WithCookies(cloneCookies(session.cookies)))
	}

	if session.csrfToken != "" {
		session.restClient.Apply(rest.WithCSRFToken(session.csrfToken))
	}

	session.graphqlClient = graphql.NewClient(
		graphql.WithBaseURL(c.graphqlBaseURL),
		graphql.WithHTTPClient(c.graphqlHTTPClient),
		graphql.WithTokenGetter(session.accessTokenForRequest),
	)

	return session, nil
}

func cloneCookies(cookies map[string]*http.Cookie) map[string]*http.Cookie {
	copiedCookies := make(map[string]*http.Cookie, len(cookies))

	for name, cookie := range cookies {
		if cookie != nil {
			item := *cookie
			copiedCookies[name] = &item
		}
	}

	return copiedCookies
}

// Credentials returns a copy of the credentials currently configured on the session.
func (s *Session) Credentials() SessionCredentials {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SessionCredentials{
		AccessToken:  s.accessToken,
		RefreshToken: s.refreshToken,
		CSRFToken:    s.csrfToken,
		Cookies:      cloneCookies(s.cookies),
	}
}

// SessionCredentials are account credentials managed by the caller.
type SessionCredentials struct {
	AccessToken, RefreshToken, CSRFToken string
	Cookies                              map[string]*http.Cookie
}

// SetAccessToken installs a caller-provided access token on this session.
func (s *Session) SetAccessToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return errAccessCredentialEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return alexaapimodels.NewClosedError("session is closed")
	}

	s.accessToken = token

	return nil
}

func setSessionOptionValue(cfg *sessionConfig, name, value string) error {
	if previous, ok := cfg.set[name]; ok && previous != value {
		return fmt.Errorf("%w %s", errConflictingValues, name)
	}

	cfg.set[name] = value

	return nil
}

// WithBearerToken sets the session's caller-managed access token.
func WithBearerToken(token string) SessionOption {
	return func(cfg *sessionConfig) error {
		if strings.TrimSpace(token) == "" {
			return errBearerCredentialEmpty
		}

		err := setSessionOptionValue(cfg, "access token", token)
		if err != nil {
			return err
		}

		cfg.accessToken = token

		return nil
	}
}

// WithRefreshToken stores a refresh token for explicit session refresh operations.
func WithRefreshToken(token string) SessionOption {
	return func(cfg *sessionConfig) error {
		if strings.TrimSpace(token) == "" {
			return errRefreshCredentialEmpty
		}

		err := setSessionOptionValue(cfg, "refresh token", token)
		if err != nil {
			return err
		}

		cfg.refreshToken = token

		return nil
	}
}

// WithCustomerID sets the account customer ID on the session.
func WithCustomerID(customerID string) SessionOption {
	return func(cfg *sessionConfig) error {
		if strings.TrimSpace(customerID) == "" {
			return errCustomerIDEmpty
		}

		err := setSessionOptionValue(cfg, "customer ID", customerID)
		if err != nil {
			return err
		}

		cfg.customerID = customerID

		return nil
	}
}

// WithCookies sets pre-obtained cookies on a new session.
func WithCookies(cookies map[string]*http.Cookie) SessionOption {
	return func(cfg *sessionConfig) error {
		if cookies == nil {
			return errCookiesNil
		}

		for name, cookie := range cookies {
			if cookie == nil {
				return fmt.Errorf("%w %q must not be nil", errCookieNil, name)
			}
		}

		if cfg.cookiesSet && !reflect.DeepEqual(cfg.cookies, cookies) {
			return errConflictingCookies
		}

		cfg.cookiesSet = true
		cfg.cookies = cloneCookies(cookies)

		return nil
	}
}

// WithCSRFToken sets the caller-managed CSRF token on a new session.
func WithCSRFToken(token string) SessionOption {
	return func(cfg *sessionConfig) error {
		if strings.TrimSpace(token) == "" {
			return errCSRFTokenEmpty
		}

		err := setSessionOptionValue(cfg, "CSRF token", token)
		if err != nil {
			return err
		}

		cfg.csrfToken = token

		return nil
	}
}

// WithRegion selects US, EU, or JP service endpoints.
func WithRegion(region alexaapimodels.Region) Option {
	return func(cfg *clientConfig) error {
		{
			_, err := regionEndpoints(region)
			if err != nil {
				return err
			}
		}

		err := setOptionValue(cfg, "region", string(region))
		if err != nil {
			return err
		}

		cfg.region = region

		return nil
	}
}

func baseURLOption(name string, assign func(*clientConfig, string), value string) Option {
	return func(cfg *clientConfig) error {
		baseURL, err := validateBaseURL(value)
		if err != nil {
			return fmt.Errorf("invalid %s: %w", name, err)
		}

		err = setOptionValue(cfg, name, baseURL)
		if err != nil {
			return err
		}

		assign(cfg, baseURL)

		return nil
	}
}

// WithGraphQLBaseURL overrides the GraphQL service base URL.
func WithGraphQLBaseURL(value string) Option {
	return baseURLOption("GraphQL base URL", func(c *clientConfig, v string) { c.graphqlBaseURL = v }, value)
}

// WithAlexaAPIBaseURL overrides the REST base URL used for Alexa API calls.
func WithAlexaAPIBaseURL(value string) Option {
	return baseURLOption("Alexa API base URL", func(c *clientConfig, v string) { c.apiBaseURL = v }, value)
}

// WithAmazonAPIBaseURL overrides the Amazon base URL used for auth calls.
func WithAmazonAPIBaseURL(value string) Option {
	return baseURLOption("Amazon API base URL", func(c *clientConfig, v string) { c.amazonBaseURL = v }, value)
}

// WithAlexaWebBaseURL overrides the Alexa web base URL used for profile calls.
func WithAlexaWebBaseURL(value string) Option {
	return baseURLOption("Alexa web base URL", func(c *clientConfig, v string) { c.webBaseURL = v }, value)
}

func validateBaseURL(baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" ||
		parsed.Fragment != "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errAbsoluteHTTPURLRequired
	}

	return strings.TrimRight(baseURL, "/"), nil
}

// WithEventAuthority overrides the host used to establish HTTP/2 event streams.
func WithEventAuthority(authority string) Option {
	return func(cfg *clientConfig) error {
		parsed, err := url.Parse("https://" + authority)
		if err != nil || authority == "" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" ||
			parsed.RawQuery != "" ||
			parsed.Fragment != "" {
			return fmt.Errorf("%w %q", errInvalidEventAuthority, authority)
		}

		{
			err := setOptionValue(cfg, "event authority", authority)
			if err != nil {
				return err
			}
		}

		cfg.eventAuthority = authority

		return nil
	}
}

// WithHTTPClient injects the shared HTTP client for REST and GraphQL.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) error {
		if client == nil {
			return errHTTPClientNil
		}

		cfg.commonHTTPClient = client

		return nil
	}
}

// WithHttpClient is a spelling-compatible alias for WithHTTPClient.
func WithHttpClient(client *http.Client) Option { return WithHTTPClient(client) }

// WithRESTHTTPClient injects the REST network edge independently.
func WithRESTHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) error {
		if client == nil {
			return errRESTHTTPClientNil
		}

		cfg.restHTTPClient = client

		return nil
	}
}

// WithGraphQLHTTPClient injects the GraphQL network edge independently.
func WithGraphQLHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) error {
		if client == nil {
			return errGraphQLHTTPClientNil
		}

		cfg.graphqlHTTPClient = client

		return nil
	}
}

// WithEventHTTPClient injects the long-lived HTTP/2 event-stream client.
func WithEventHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) error {
		if client == nil {
			return errEventHTTPClientNil
		}

		cfg.eventHTTPClient = client

		return nil
	}
}

// WithEventTransport injects a custom RoundTripper for event streams.
func WithEventTransport(transport http.RoundTripper) Option {
	return func(cfg *clientConfig) error {
		if transport == nil {
			return errEventTransportNil
		}

		cfg.eventTransport = transport

		return nil
	}
}

// WithTimeout sets the default REST and GraphQL request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *clientConfig) error {
		if timeout <= 0 {
			return errTimeoutNotPositive
		}

		err := setOptionValue(cfg, "timeout", timeout.String())
		if err != nil {
			return err
		}

		cfg.timeout = timeout

		return nil
	}
}

// Close closes event connections owned by this session.
func (s *Session) Close() error {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()

		return nil
	}

	s.closed = true

	connections := make([]*HTTP2Connection, 0, len(s.eventConnections))

	for connection := range s.eventConnections {
		connections = append(connections, connection)
	}

	s.mu.Unlock()

	for _, connection := range connections {
		_ = connection.Close()
	}

	return nil
}

// RequestEndpointQualityOfService requests quality of service for endpoints.
func (s *Session) RequestEndpointQualityOfService(
	ctx context.Context,
	req alexaapimodels.QualityOfServiceRequest,
) (*alexaapimodels.QualityOfServiceResponse, error) {
	// Convert alexaapimodels request to graphql input
	input := graphql.ConvertQualityOfServiceRequest(&req)

	// Call the graphql client
	gqlResp, err := s.graphqlClient.RequestEndpointQualityOfService(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert graphql response to alexaapimodels response
	return graphql.ConvertToQualityOfServiceResponse(gqlResp), nil
}

// Subscribe subscribes the session's account to endpoint events.
func (s *Session) Subscribe(
	ctx context.Context,
	req alexaapimodels.SubscribeRequest,
) (*alexaapimodels.SubscribeResponse, error) {
	// Convert alexaapimodels request to graphql input
	input := graphql.ConvertSubscribeRequest(&req)

	// Call the graphql client
	gqlResp, err := s.graphqlClient.Subscribe(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert graphql response to alexaapimodels response
	return graphql.ConvertToSubscribeResponse(gqlResp), nil
}

// ConnectEvents opens a caller-owned HTTP/2 event stream for this account session.
func (s *Session) ConnectEvents(ctx context.Context) (*HTTP2Connection, error) {
	{
		_, err := s.accessTokenForRequest(ctx)
		if err != nil {
			return nil, err
		}
	}

	conn := newSessionHTTP2Connection(s.client.eventAuthority, s.accessTokenForRequest, s.client.eventHTTPClient)

	err := conn.Connect(ctx)
	if err != nil {
		_ = conn.Close()

		return nil, &alexaapimodels.ConnectionError{
			Message: "failed to connect",
			Err:     err,
		}
	}

	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()

		_ = conn.Close()

		return nil, alexaapimodels.NewClosedError("session is closed")
	}

	s.eventConnections[conn] = struct{}{}
	s.mu.Unlock()

	return conn, nil
}

// GenerateCodePair generates a code pair for code-based linking (CBL) authentication.
func (c *Client) GenerateCodePair(
	ctx context.Context,
	config alexaapimodels.DeviceRegistrationConfig,
) (*alexaapimodels.CodePairResponse, error) {
	// Convert alexaapimodels config to rest.DeviceRegistrationConfig
	restConfig := &rest.DeviceRegistrationConfig{
		AppName:      config.AppName,
		AppVersion:   config.AppVersion,
		DeviceType:   config.DeviceType,
		Domain:       config.Domain,
		DeviceModel:  config.DeviceModel,
		OSVersion:    config.OSVersion,
		DeviceSerial: config.DeviceSerial,
		DeviceName:   config.DeviceName,
		Manufacturer: config.Manufacturer,
	}

	// Call the rest client
	restResp, err := c.newRESTClient().GenerateCodePair(ctx, restConfig)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.CodePairResponse{
		PublicCode:  restResp.PublicCode,
		PrivateCode: restResp.PrivateCode,
	}, nil
}

// RegisterWithCodePair registers a device using code-based linking (CBL).
func (c *Client) RegisterWithCodePair(
	ctx context.Context,
	publicCode, privateCode string,
	config alexaapimodels.DeviceRegistrationConfig,
) (*alexaapimodels.RegistrationResponse, error) {
	// Convert alexaapimodels config to rest.DeviceRegistrationConfig
	restConfig := &rest.DeviceRegistrationConfig{
		AppName:      config.AppName,
		AppVersion:   config.AppVersion,
		DeviceType:   config.DeviceType,
		Domain:       config.Domain,
		DeviceModel:  config.DeviceModel,
		OSVersion:    config.OSVersion,
		DeviceSerial: config.DeviceSerial,
		DeviceName:   config.DeviceName,
		Manufacturer: config.Manufacturer,
	}

	// Call the rest client
	restResp, err := c.newRESTClient().RegisterWithCodePair(ctx, publicCode, privateCode, restConfig)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.RegistrationResponse{
		AccessToken:  restResp.Response.Success.Tokens.Bearer.AccessToken,
		RefreshToken: restResp.Response.Success.Tokens.Bearer.RefreshToken,
	}, nil
}

// RefreshAccessToken refreshes an access token using a refresh token.
func (c *Client) RefreshAccessToken(
	ctx context.Context,
	req alexaapimodels.TokenRefreshRequest,
) (*alexaapimodels.TokenRefreshResponse, error) {
	// Convert alexaapimodels config to rest.DeviceRegistrationConfig
	restConfig := &rest.DeviceRegistrationConfig{
		AppName:      req.Config.AppName,
		AppVersion:   req.Config.AppVersion,
		DeviceType:   req.Config.DeviceType,
		Domain:       req.Config.Domain,
		DeviceModel:  req.Config.DeviceModel,
		OSVersion:    req.Config.OSVersion,
		DeviceSerial: req.Config.DeviceSerial,
		DeviceName:   req.Config.DeviceName,
		Manufacturer: req.Config.Manufacturer,
	}

	// Call the rest client
	restResp, err := c.newRESTClient().RefreshAccessToken(ctx, req.RefreshToken, restConfig)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.TokenRefreshResponse{
		AccessToken:      restResp.AccessToken,
		ExpiresInSeconds: restResp.ExpiresInSeconds,
	}, nil
}

func (c *Client) newRESTClient() *rest.Client {
	return rest.NewClient(rest.WithRegion(c.region), rest.WithAmazonalexaAPIBaseURI(c.apiBaseURL),
		rest.WithAmazonapiBaseURI(c.amazonBaseURL), rest.WithAlexaAmazonBaseURI(c.webBaseURL),
		rest.WithHTTPClient(c.restHTTPClient))
}

// GetUserInfo retrieves user information from the /api/users/me endpoint.
func (s *Session) GetUserInfo(
	ctx context.Context,
	req alexaapimodels.UserInfoRequest,
) (*alexaapimodels.UserInfo, error) {
	// Convert alexaapimodels request to rest.GetUserInfoOptions
	restOpts := &rest.GetUserInfoOptions{
		Platform:  req.Platform,
		Version:   req.Version,
		CSRFToken: req.CSRFToken,
	}

	// Call the rest client
	restResp, err := s.restClient.GetUserInfo(ctx, restOpts)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.UserInfo{
		CountryOfResidence:     restResp.CountryOfResidence,
		EffectiveMarketPlaceID: restResp.EffectiveMarketPlaceID,
		Email:                  restResp.Email,
		EulaAcceptance:         restResp.EulaAcceptance,
		Features:               restResp.Features,
		FullName:               restResp.FullName,
		HasActiveDopplers:      restResp.HasActiveDopplers,
		ID:                     restResp.ID,
		MarketPlaceDomainName:  restResp.MarketPlaceDomainName,
		MarketPlaceID:          restResp.MarketPlaceID,
		MarketPlaceLocale:      restResp.MarketPlaceLocale,
	}, nil
}

// convertPlayerInfo converts alexamodels.PlayerInfo to alexaapimodels.PlayerInfo.
func convertPlayerInfo(src *alexamodels.PlayerInfo) *alexaapimodels.PlayerInfo {
	if src == nil {
		return nil
	}

	return &alexaapimodels.PlayerInfo{
		Hint:             src.Hint,
		InfoText:         convertInfoText(src.InfoText),
		IsPlayingInLemur: src.IsPlayingInLemur,
		LemurVolume:      src.LemurVolume,
		Lyrics:           src.Lyrics,
		MainArt:          convertArt(src.MainArt),
		MediaId:          src.MediaId,
		MiniArt:          convertArt(src.MiniArt),
		MiniInfoText:     convertInfoText(src.MiniInfoText),
		PlaybackSource:   src.PlaybackSource,
		PlayingInLemurId: src.PlayingInLemurId,
		Progress:         convertProgress(src.Progress),
		Provider:         convertProvider(src.Provider),
		Quality:          src.Quality,
		QueueId:          src.QueueId,
		State:            src.State,
		Template:         convertTemplate(src.Template),
		Transport:        convertTransport(src.Transport),
		UpNextItems:      src.UpNextItems,
		Volume:           convertVolume(src.Volume),
	}
}

// convertInfoText converts alexamodels.InfoText to alexaapimodels.InfoText.
func convertInfoText(src *alexamodels.InfoText) *alexaapimodels.InfoText {
	if src == nil {
		return nil
	}

	return &alexaapimodels.InfoText{
		Header:         src.Header,
		HeaderSubtext1: src.HeaderSubtext1,
		MultiLineMode:  src.MultiLineMode,
		SubText1:       src.SubText1,
		SubText2:       src.SubText2,
		Title:          src.Title,
	}
}

// convertArt converts alexamodels.Art to alexaapimodels.Art.
func convertArt(src *alexamodels.Art) *alexaapimodels.Art {
	if src == nil {
		return nil
	}

	return &alexaapimodels.Art{
		AltText:     src.AltText,
		ArtType:     src.ArtType,
		ContentType: src.ContentType,
		URL:         src.URL,
		IconId:      src.IconId,
		IconStyles:  src.IconStyles,
	}
}

// convertProgress converts alexamodels.Progress to alexaapimodels.Progress.
func convertProgress(src *alexamodels.Progress) *alexaapimodels.Progress {
	if src == nil {
		return nil
	}

	return &alexaapimodels.Progress{
		AllowScrubbing: src.AllowScrubbing,
		LocationInfo:   src.LocationInfo,
		MediaLength:    src.MediaLength,
		MediaProgress:  src.MediaProgress,
		ShowTiming:     src.ShowTiming,
		Visible:        src.Visible,
	}
}

// convertProvider converts alexamodels.Provider to alexaapimodels.Provider.
func convertProvider(src *alexamodels.Provider) *alexaapimodels.Provider {
	if src == nil {
		return nil
	}

	return &alexaapimodels.Provider{
		ArtOverlay:          convertArt(src.ArtOverlay),
		FallbackMainArt:     convertArt(src.FallbackMainArt),
		ProviderDisplayName: src.ProviderDisplayName,
		ProviderLogo:        convertArt(src.ProviderLogo),
		ProviderName:        src.ProviderName,
	}
}

// convertTemplate converts alexamodels.Template to alexaapimodels.Template.
func convertTemplate(src *alexamodels.Template) *alexaapimodels.Template {
	if src == nil {
		return nil
	}

	return &alexaapimodels.Template{
		Art:                convertArt(src.Art),
		BackgroundImageURL: src.BackgroundImageURL,
		TemplateType:       src.TemplateType,
	}
}

// convertTransport converts alexamodels.Transport to alexaapimodels.Transport.
func convertTransport(src *alexamodels.Transport) *alexaapimodels.Transport {
	if src == nil {
		return nil
	}

	return &alexaapimodels.Transport{
		ClosedCaptions:    src.ClosedCaptions,
		LayoutType:        src.LayoutType,
		Lyrics:            src.Lyrics,
		Next:              src.Next,
		PlayPause:         src.PlayPause,
		Previous:          src.Previous,
		RateContentAction: convertRateContentAction(src.RateContentAction),
		ThumbsDown:        src.ThumbsDown,
		ThumbsUp:          src.ThumbsUp,
	}
}

// convertRateContentAction converts alexamodels.RateContentAction to alexaapimodels.RateContentAction.
func convertRateContentAction(src *alexamodels.RateContentAction) *alexaapimodels.RateContentAction {
	if src == nil {
		return nil
	}

	return &alexaapimodels.RateContentAction{
		MediaOwnerCustomerId: src.MediaOwnerCustomerId,
		Rating:               src.Rating,
		Type:                 src.Type,
	}
}

// convertVolume converts alexamodels.Volume to alexaapimodels.Volume.
func convertVolume(src *alexamodels.Volume) *alexaapimodels.Volume {
	if src == nil {
		return nil
	}

	return &alexaapimodels.Volume{
		Muted:  src.Muted,
		Volume: src.Volume,
	}
}

// GetPlayerState retrieves the current player state for an endpoint.
func (s *Session) GetPlayerState(
	ctx context.Context,
	req alexaapimodels.PlayerStateRequest,
) (*alexaapimodels.PlayerStateResponse, error) {
	if req.Target == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "target endpoint is required",
		}
	}

	// Convert alexaapimodels request to dependencymodels request
	restReq := &alexamodels.PlayerStateRequest{
		Endpoint: req.Target,
	}

	// Call the rest client
	restResp, err := s.restClient.GetPlayerState(ctx, restReq)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	var playerInfo *alexaapimodels.PlayerInfo
	if restResp.PlayerInfo != nil {
		playerInfo = convertPlayerInfo(restResp.PlayerInfo)
	}

	return &alexaapimodels.PlayerStateResponse{
		PlayerInfo: playerInfo,
	}, nil
}

// RefreshAccessToken explicitly returns refreshed credentials without updating
// session state. Store the result and install the access token with SetAccessToken.
func (s *Session) RefreshAccessToken(
	ctx context.Context,
	config alexaapimodels.DeviceRegistrationConfig,
) (*alexaapimodels.TokenRefreshResponse, error) {
	credentials := s.Credentials()
	if credentials.RefreshToken == "" {
		return nil, &alexaapimodels.TokenError{Message: "session has no refresh token"}
	}

	return s.client.RefreshAccessToken(
		ctx,
		alexaapimodels.TokenRefreshRequest{RefreshToken: credentials.RefreshToken, Config: config},
	)
}

// ExchangeRefreshTokenForCookies explicitly exchanges the session refresh token
// and returns the resulting cookies for caller storage and a new session.
func (s *Session) ExchangeRefreshTokenForCookies(ctx context.Context, domain string) (map[string]*http.Cookie, error) {
	credentials := s.Credentials()
	if credentials.RefreshToken == "" {
		return nil, &alexaapimodels.TokenError{Message: "session has no refresh token"}
	}

	if strings.TrimSpace(domain) == "" {
		return nil, &alexaapimodels.BadRequestError{Message: "cookie exchange domain is required"}
	}

	return s.client.newRESTClient().ExchangeRefreshTokenForCookies(ctx, credentials.RefreshToken, domain)
}

// GetCSRFToken explicitly retrieves and returns a CSRF token for this session.
func (s *Session) GetCSRFToken(ctx context.Context) (string, error) {
	token, err := s.restClient.GetCSRFToken(ctx)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.csrfToken = token
	s.mu.Unlock()

	return token, nil
}

func (s *Session) accessTokenForRequest(context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return "", alexaapimodels.NewClosedError("session is closed")
	}

	if strings.TrimSpace(s.accessToken) == "" {
		return "", &alexaapimodels.TokenError{Message: "no access token available"}
	}

	return s.accessToken, nil
}
