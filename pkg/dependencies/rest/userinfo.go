package rest

import (
	"context"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
	"net/url"
)

// GetUserInfoOptions contains options for getting user info.
type GetUserInfoOptions struct {
	Platform  string // Platform identifier (e.g., "ios", "android")
	Version   string // App version (e.g., "2.2.556530.0")
	CSRFToken string // CSRF token for cookie header
}

// GetUserInfo retrieves user information from the /api/users/me endpoint
// GET {alexaAmazonBaseURI}/api/users/me?platform={platform}&version={version}.
func (c *Client) GetUserInfo(ctx context.Context, opts *GetUserInfoOptions) (*alexamodels.UserInfo, error) {
	if opts == nil {
		opts = &GetUserInfoOptions{}
	}

	// Build the full URL using the client's base URI
	baseURL := c.alexaAmazonBaseURI + apiroutes.PathGetUserInfo

	params := url.Values{}
	if opts.Platform != "" {
		params.Set(apiroutes.QueryParamPlatform, opts.Platform)
	}

	if opts.Version != "" {
		params.Set(apiroutes.QueryParamVersion, opts.Version)
	}

	if len(params) > 0 {
		baseURL += "?" + params.Encode()
	}

	// Build custom headers
	customHeaders := make(map[string]string)
	if opts.CSRFToken != "" {
		customHeaders[apiroutes.HeaderCookie] = alexamodels.AuthCSRFCookieName + "=" + opts.CSRFToken
	}

	// Use the client's helper method to perform the request
	var userInfo alexamodels.WireUserInfo

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodGetUserInfo, baseURL, nil, customHeaders, &userInfo, true)
	if err != nil {
		return nil, err
	}

	return convertWireModel[alexamodels.UserInfo](userInfo)
}
