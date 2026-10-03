// Package alexamodels provides dependency API request and response models.
package alexamodels

import "time"

// IsExpired checks if the token is expired or will expire soon (within 1 minute).
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}

	return time.Now().Add(time.Minute).After(t.ExpiresAt)
}

// AuthConfig represents caller-provided OAuth configuration.
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}

// DeviceAuthConfig represents caller-provided OAuth device configuration.
type DeviceAuthConfig struct {
	ClientID      string
	ClientSecret  string
	DeviceAuthURL string
	TokenURL      string
	Scopes        []string
}
