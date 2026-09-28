package rest

import (
	"context"
	"fmt"
	"net/url"

	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// GetUserInfoOptions contains options for getting user info
type GetUserInfoOptions struct {
	Platform  string // Platform identifier (e.g., "ios", "android")
	Version   string // App version (e.g., "2.2.556530.0")
	CSRFToken string // CSRF token for cookie header
}

// GetUserInfo retrieves user information from the /api/users/me endpoint
// GET {alexaAmazonBaseUri}/api/users/me?platform={platform}&version={version}
func (c *Client) GetUserInfo(ctx context.Context, opts *GetUserInfoOptions) (*alexamodels.UserInfo, error) {
	if opts == nil {
		opts = &GetUserInfoOptions{}
	}

	// Build the full URL using the client's base URI
	baseURL := c.alexaAmazonBaseUri + "/api/users/me"
	params := url.Values{}
	if opts.Platform != "" {
		params.Set("platform", opts.Platform)
	}
	if opts.Version != "" {
		params.Set("version", opts.Version)
	}
	if len(params) > 0 {
		baseURL += "?" + params.Encode()
	}

	// Build custom headers
	customHeaders := make(map[string]string)
	if opts.CSRFToken != "" {
		customHeaders["Cookie"] = fmt.Sprintf("csrf=%s", opts.CSRFToken)
	}

	// Use the client's helper method to perform the request
	var userInfo alexamodels.UserInfo
	if err := c.doJSONRequestWithFullURL(ctx, "GET", baseURL, nil, customHeaders, &userInfo, true); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
