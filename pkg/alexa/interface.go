// Package alexa provides the main client for interacting with Alexa services.
// This file defines the Client interface for dependency injection and testing.
package alexa

import (
	"context"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// A connection is representation of a bidirectional connection to the Alexa API.
type Connection interface {
	Receive() (*alexaapimodels.Event, error)
	Close() error
}

// ClientInterface defines the interface for the Alexa client.
type ClientInterface interface {
	// ListEndpoints retrieves all endpoints by joining GraphQL endpoint data
	// with endpointV2 data using a left outer join on DMSIdentifier.
	// The query parameter can specify filters and which optional fields to include
	// (states, capabilities, features). When States is true in IncludeFields,
	// the function will retrieve endpoints with detailed property states.
	ListEndpoints(ctx context.Context, q alexaapimodels.EndpointQuery) (*alexaapimodels.UnifiedEndpointListResponse, error)

	// Control provides a unified interface for all control operations.
	// It dispatches to the appropriate control method based on the namespace and name in the request.
	// All control operations are async fire-and-forget, meaning that we don't wait for a confirmation of success.
	Control(ctx context.Context, req alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error)

	// ConnectEvents establishes an HTTP/2 connection for receiving events
	// Note that no events will be received to the connect events endpoint unless a subscription has been created.
	ConnectEvents(ctx context.Context) (*HTTP2Connection, error)

	// RequestEndpointQualityOfService activates cloud side polling of endpoint states, so that when you query data cloud side, the data will exist.
	// As of right now, the data is not guaranteed to be correct if you query cloud side. This helps to make cloud side data more reliable
	// And reflect what is the state on the physical endpoint.
	RequestEndpointQualityOfService(ctx context.Context, req alexaapimodels.QualityOfServiceRequest) (*alexaapimodels.QualityOfServiceResponse, error)

	// Subscribe subscribes to endpoint events,
	// the events will be published to the connection object you create via the connect events method.
	Subscribe(ctx context.Context, req alexaapimodels.SubscribeRequest) (*alexaapimodels.SubscribeResponse, error)

	// GenerateCodePair generates a code pair for code-based linking (CBL) authentication.
	// The user must enter the public code on https://amazon.com/code before calling RegisterWithCodePair.
	GenerateCodePair(ctx context.Context, config alexaapimodels.DeviceRegistrationConfig) (*alexaapimodels.CodePairResponse, error)

	// RegisterWithCodePair registers a device using code-based linking (CBL).
	// This should be called after GenerateCodePair and after the user has entered the public code on Amazon's website.
	RegisterWithCodePair(ctx context.Context, publicCode, privateCode string, config alexaapimodels.DeviceRegistrationConfig) (*alexaapimodels.RegistrationResponse, error)

	// RefreshAccessToken refreshes an access token using a refresh token.
	RefreshAccessToken(ctx context.Context, req alexaapimodels.TokenRefreshRequest) (*alexaapimodels.TokenRefreshResponse, error)

	// GetUserInfo retrieves user information from the /api/users/me endpoint.
	// The request parameter can specify optional platform, version, and CSRF token.
	GetUserInfo(ctx context.Context, req alexaapimodels.UserInfoRequest) (*alexaapimodels.UserInfo, error)

	// GetPlayerState retrieves the current player state for an endpoint.
	// This returns information about the currently playing media, including state, title, artist, etc.
	// Note that this is not supported by spotify, and only seemingly works with amazon music.
	GetPlayerState(ctx context.Context, req alexaapimodels.PlayerStateRequest) (*alexaapimodels.PlayerStateResponse, error)

	// Close closes the client and all connections
	Close() error
}

// Ensure Client implements ClientInterface at compile time
var _ ClientInterface = (*Client)(nil)
