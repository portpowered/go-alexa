// Package alexa provides the main client for interacting with Alexa services.
// It orchestrates authentication, API clients (REST and GraphQL), and event connections.
// The client supports both full OAuth flows and simple bearer token authentication.
package alexa

import (
	"context"
	"net/http"
	"sync"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	"github.com/portpowered/go-alexa/pkg/dependencies/rest"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// Client is the main client for interacting with Alexa services
type Client struct {
	restClient         *rest.Client
	graphqlClient      *graphql.Client
	connection         *HTTP2Connection
	authURI            string
	apiServiceURI      string
	http2ConnectionURI string
	bearerToken        string
	tokenGetter        func(ctx context.Context) (string, error)
	mu                 sync.RWMutex
	closed             bool
}

// Apply options to the client.
func (c *Client) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt.Apply(c); err != nil {
			return err
		}
	}
	return nil
}

// NewClient creates a new Alexa client
func NewClient(opts ...Option) (*Client, error) {
	client := &Client{
		authURI:            alexaapimodels.AuthorizationUriNa,
		apiServiceURI:      alexaapimodels.ApiServiceUriNa,
		http2ConnectionURI: alexaapimodels.Http2ConnectionUriNa,
		restClient:         rest.NewClient(),
		graphqlClient:      graphql.NewClient(),
		connection:         http2Connection(),
	}

	for _, opt := range opts {
		if err := opt.Apply(client); err != nil {
			return nil, &alexaapimodels.ConnectionError{
				Message: "failed to apply option",
				Err:     err,
			}
		}
	}

	return client, nil
}

type Option interface {
	Apply(c *Client) error
}

// WithBearerToken creates a client with a simple bearer token (bypasses OAuth setup)
func WithBearerToken(token string) Option {
	return withBearerToken(token)
}

type withBearerToken string

func (w withBearerToken) Apply(c *Client) error {
	token := string(w)
	c.bearerToken = token
	c.restClient.Apply(
		rest.WithBearerToken(token),
	)
	c.graphqlClient.Apply(
		graphql.WithBearerToken(token),
	)
	return nil
}

func WithCustomerID(customerID string) Option {
	return withCustomerID(customerID)
}

type withCustomerID string

func (w withCustomerID) Apply(c *Client) error {
	c.restClient.Apply(
		rest.WithCustomerID(string(w)),
	)
	return nil
}

func WithRegion(region alexaapimodels.Region) Option {
	return withRegion(region)
}

type withRegion alexaapimodels.Region

func (w withRegion) Apply(c *Client) error {
	switch alexaapimodels.Region(w) {
	case alexaapimodels.RegionUS:
		c.authURI = alexaapimodels.AuthorizationUriNa
		c.apiServiceURI = alexaapimodels.ApiServiceUriNa
		c.http2ConnectionURI = alexaapimodels.Http2ConnectionUriNa
		c.restClient.Apply(
			rest.WithRegion(alexaapimodels.RegionUS),
		)
	case alexaapimodels.RegionEU:
		c.authURI = alexaapimodels.AuthorizationUriEu
		c.apiServiceURI = alexaapimodels.ApiServiceUriEu
		c.http2ConnectionURI = alexaapimodels.Http2ConnectionUriEu
		c.restClient.Apply(
			rest.WithRegion(alexaapimodels.RegionEU),
		)
	case alexaapimodels.RegionJP:
		c.authURI = alexaapimodels.AuthorizationUriJp
		c.apiServiceURI = alexaapimodels.ApiServiceUriJp
		c.http2ConnectionURI = alexaapimodels.Http2ConnectionUriJp
		c.restClient.Apply(
			rest.WithRegion(alexaapimodels.RegionJP),
		)
	}
	return nil
}

// NewClientWithToken creates a client with a simple bearer token (convenience function)
func NewClientWithToken(bearerToken string, opts ...Option) (*Client, error) {
	opts = append([]Option{WithBearerToken(bearerToken)}, opts...)
	return NewClient(opts...)
}

// WithRefreshToken creates a client with cookie-based authentication using a refresh token
// This will automatically exchange the refresh token for cookies and retrieve CSRF tokens
func WithRefreshToken(refreshToken string) Option {
	return withRefreshToken(refreshToken)
}

type withRefreshToken string

func (w withRefreshToken) Apply(c *Client) error {
	token := string(w)
	c.restClient.Apply(
		rest.WithRefreshToken(token),
	)
	// Note: Cookie exchange happens automatically on first request
	return nil
}

// WithCookies creates a client with cookie-based authentication using pre-obtained cookies
func WithCookies(cookies map[string]*http.Cookie) Option {
	return withCookies(cookies)
}

type withCookies map[string]*http.Cookie

func (w withCookies) Apply(c *Client) error {
	cookies := map[string]*http.Cookie(w)
	c.restClient.Apply(
		rest.WithCookies(cookies),
	)
	return nil
}

// WithCSRFToken sets a CSRF token for cookie-based authentication
func WithCSRFToken(csrfToken string) Option {
	return withCSRFToken(csrfToken)
}

type withCSRFToken string

func (w withCSRFToken) Apply(c *Client) error {
	token := string(w)
	c.restClient.Apply(
		rest.WithCSRFToken(token),
	)
	return nil
}

// WithHttpClient sets a custom HTTP client for REST API calls.
// This allows full control over the HTTP client configuration, including
// setting a custom transport for features like network capture/replay.
func WithHttpClient(httpClient *http.Client) Option {
	return withRESTHTTPClient{httpClient: httpClient}
}

type withRESTHTTPClient struct {
	httpClient *http.Client
}

func (w withRESTHTTPClient) Apply(c *Client) error {
	c.restClient.Apply(rest.WithHTTPClient(w.httpClient))
	c.graphqlClient.Apply(graphql.WithHTTPClient(w.httpClient))
	return nil
}

// RequestEndpointQualityOfService requests quality of service for endpoints
func (c *Client) RequestEndpointQualityOfService(ctx context.Context, req alexaapimodels.QualityOfServiceRequest) (*alexaapimodels.QualityOfServiceResponse, error) {
	// Convert alexaapimodels request to graphql input
	input := graphql.ConvertQualityOfServiceRequest(&req)

	// Call the graphql client
	gqlResp, err := c.graphqlClient.RequestEndpointQualityOfService(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert graphql response to alexaapimodels response
	return graphql.ConvertToQualityOfServiceResponse(gqlResp), nil
}

// Subscribe subscribes to endpoint events
func (c *Client) Subscribe(ctx context.Context, req alexaapimodels.SubscribeRequest) (*alexaapimodels.SubscribeResponse, error) {
	// Convert alexaapimodels request to graphql input
	input := graphql.ConvertSubscribeRequest(&req)

	// Call the graphql client
	gqlResp, err := c.graphqlClient.Subscribe(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert graphql response to alexaapimodels response
	return graphql.ConvertToSubscribeResponse(gqlResp), nil
}

// ConnectEvents establishes an HTTP/2 connection for receiving events
func (c *Client) ConnectEvents(ctx context.Context) (*HTTP2Connection, error) {
	c.mu.RLock()
	authority := c.http2ConnectionURI
	c.mu.RUnlock()

	// Create token getter function
	tokenGetter := func(ctx context.Context) (string, error) {
		return c.getTokenForConnection(ctx)
	}

	// Create HTTP/2 connection with options
	http2Conn := http2Connection(
		WithAuthority(authority),
		WithTokenGetter(tokenGetter),
	)

	// Connect to the server
	if err := http2Conn.Connect(ctx); err != nil {
		return nil, &alexaapimodels.ConnectionError{
			Message: "failed to connect",
			Err:     err,
		}
	}

	return http2Conn, nil
}

// getTokenForConnection retrieves the bearer token for the HTTP/2 connection
func (c *Client) getTokenForConnection(ctx context.Context) (string, error) {
	if c.bearerToken != "" {
		return c.bearerToken, nil
	}
	if c.tokenGetter != nil {
		return c.tokenGetter(ctx)
	}
	// Try to get token from REST client if available
	if c.restClient != nil {
		// We'll need to expose a method or use reflection, but for now
		// we'll require the token to be set via WithBearerToken
		return "", &alexaapimodels.TokenError{
			Message: "no token available - use WithBearerToken option",
		}
	}
	return "", &alexaapimodels.TokenError{
		Message: "no token available",
	}
}

// GenerateCodePair generates a code pair for code-based linking (CBL) authentication
func (c *Client) GenerateCodePair(ctx context.Context, config alexaapimodels.DeviceRegistrationConfig) (*alexaapimodels.CodePairResponse, error) {
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
	restResp, err := c.restClient.GenerateCodePair(ctx, restConfig)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.CodePairResponse{
		PublicCode:  restResp.PublicCode,
		PrivateCode: restResp.PrivateCode,
	}, nil
}

// RegisterWithCodePair registers a device using code-based linking (CBL)
func (c *Client) RegisterWithCodePair(ctx context.Context, publicCode, privateCode string, config alexaapimodels.DeviceRegistrationConfig) (*alexaapimodels.RegistrationResponse, error) {
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
	restResp, err := c.restClient.RegisterWithCodePair(ctx, publicCode, privateCode, restConfig)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.RegistrationResponse{
		AccessToken:  restResp.Response.Success.Tokens.Bearer.AccessToken,
		RefreshToken: restResp.Response.Success.Tokens.Bearer.RefreshToken,
	}, nil
}

// RefreshAccessToken refreshes an access token using a refresh token
func (c *Client) RefreshAccessToken(ctx context.Context, req alexaapimodels.TokenRefreshRequest) (*alexaapimodels.TokenRefreshResponse, error) {
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
	restResp, err := c.restClient.RefreshAccessToken(ctx, req.RefreshToken, restConfig)
	if err != nil {
		return nil, err
	}

	// Convert rest response to alexaapimodels response
	return &alexaapimodels.TokenRefreshResponse{
		AccessToken:      restResp.AccessToken,
		ExpiresInSeconds: restResp.ExpiresInSeconds,
	}, nil
}

// GetUserInfo retrieves user information from the /api/users/me endpoint
func (c *Client) GetUserInfo(ctx context.Context, req alexaapimodels.UserInfoRequest) (*alexaapimodels.UserInfo, error) {
	// Convert alexaapimodels request to rest.GetUserInfoOptions
	restOpts := &rest.GetUserInfoOptions{
		Platform:  req.Platform,
		Version:   req.Version,
		CSRFToken: req.CSRFToken,
	}

	// Call the rest client
	restResp, err := c.restClient.GetUserInfo(ctx, restOpts)
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

// convertPlayerInfo converts alexamodels.PlayerInfo to alexaapimodels.PlayerInfo
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

// convertInfoText converts alexamodels.InfoText to alexaapimodels.InfoText
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

// convertArt converts alexamodels.Art to alexaapimodels.Art
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

// convertProgress converts alexamodels.Progress to alexaapimodels.Progress
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

// convertProvider converts alexamodels.Provider to alexaapimodels.Provider
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

// convertTemplate converts alexamodels.Template to alexaapimodels.Template
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

// convertTransport converts alexamodels.Transport to alexaapimodels.Transport
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

// convertRateContentAction converts alexamodels.RateContentAction to alexaapimodels.RateContentAction
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

// convertVolume converts alexamodels.Volume to alexaapimodels.Volume
func convertVolume(src *alexamodels.Volume) *alexaapimodels.Volume {
	if src == nil {
		return nil
	}
	return &alexaapimodels.Volume{
		Muted:  src.Muted,
		Volume: src.Volume,
	}
}

// GetPlayerState retrieves the current player state for an endpoint
func (c *Client) GetPlayerState(ctx context.Context, req alexaapimodels.PlayerStateRequest) (*alexaapimodels.PlayerStateResponse, error) {
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
	restResp, err := c.restClient.GetPlayerState(ctx, restReq)
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

// Close closes the client and all connections
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.connection != nil {
		_ = c.connection.Close()
	}

	return nil
}
