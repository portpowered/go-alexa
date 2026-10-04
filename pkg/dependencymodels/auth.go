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

// AuthConfig retains the legacy shape for caller-owned OAuth configuration.
// SDK operations do not consume it or handle browser callbacks.
//
// Deprecated: configure browser authorization in the calling application.
//
//modelinventory:legacy-nonwire AuthConfig: unused caller configuration for source compatibility; no SDK consumer, JSON tags, or codec.
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}

// DeviceAuthConfig retains the legacy shape for caller-owned device OAuth configuration.
// SDK operations do not consume it; account linking uses code-pair registration.
//
// Deprecated: use the session's code-pair registration methods for account linking.
//
//modelinventory:legacy-nonwire DeviceAuthConfig: unused caller configuration for source compatibility; no SDK consumer, JSON tags, or codec.
type DeviceAuthConfig struct {
	ClientID      string
	ClientSecret  string
	DeviceAuthURL string
	TokenURL      string
	Scopes        []string
}
