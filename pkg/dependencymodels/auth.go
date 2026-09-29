// Package alexamodels provides dependency API request and response models.
package alexamodels

import (
	"time"
)

// Token represents an OAuth access token with expiration tracking.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
}

// IsExpired checks if the token is expired or will expire soon (within 1 minute).
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	// Consider token expired if it expires within 1 minute
	return time.Now().Add(1 * time.Minute).After(t.ExpiresAt)
}

// AuthConfig represents OAuth configuration for standard flow.
type AuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}

// DeviceAuthConfig represents OAuth device authorization configuration.
type DeviceAuthConfig struct {
	ClientID      string
	ClientSecret  string
	DeviceAuthURL string
	TokenURL      string
	Scopes        []string
}

// DeviceAuthResponse represents the response from device authorization endpoint.
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval,omitempty"`
}

// MAP (Mobile Access Protocol) Authentication Models

// RegistrationData represents device registration information for MAP authentication.
type RegistrationData struct {
	AppName      string `json:"app_name"`
	AppVersion   string `json:"app_version"`
	DeviceType   string `json:"device_type"`
	Domain       string `json:"domain"`
	DeviceModel  string `json:"device_model"`
	OSVersion    string `json:"os_version"`
	DeviceSerial string `json:"device_serial"`
	DeviceName   string `json:"device_name"`
}

// EmailPasswordAuth represents email/password authentication data.
type EmailPasswordAuth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CodePairAuth represents code-based linking authentication data.
type CodePairAuth struct {
	PublicCode  string `json:"public_code"`
	PrivateCode string `json:"private_code"`
}

// RegistrationRequest represents a MAP registration request.
type RegistrationRequest struct {
	RequestedTokenType []string         `json:"requested_token_type"`
	RegistrationData   RegistrationData `json:"registration_data"`
	AuthData           AuthData         `json:"auth_data"`
}

// AuthData contains the authentication method used for device registration.
type AuthData struct {
	CodePairAuth      *CodePairAuth      `json:"code_pair,omitempty"`
	EmailPasswordAuth *EmailPasswordAuth `json:"email_password,omitempty"`
}

// RegistrationResponse represents a successful MAP registration response.
type RegistrationResponse struct {
	Response struct {
		Success struct {
			Tokens struct {
				Bearer struct {
					AccessToken  string `json:"access_token"`
					RefreshToken string `json:"refresh_token"`
				} `json:"bearer"`
			} `json:"tokens"`
		} `json:"success"`
	} `json:"response"`
}

// ChallengeResponse represents a challenge response from MAP registration.
type ChallengeResponse struct {
	Response struct {
		Challenge struct {
			ChallengeReason              string `json:"challenge_reason"`
			RequiredAuthenticationMethod string `json:"required_authentication_method,omitempty"`
		} `json:"challenge"`
	} `json:"response"`
}

// CodePairRequest represents a request to generate CBL codes.
type CodePairRequest struct {
	CodeData struct {
		AppName               string `json:"app_name"`
		AppVersion            string `json:"app_version"`
		DeviceType            string `json:"device_type"`
		Domain                string `json:"domain"`
		DeviceModel           string `json:"device_model"`
		OSVersion             string `json:"os_version"`
		DeviceSerial          string `json:"device_serial"`
		DeviceName            string `json:"device_name"`
		SecondaryRegistration string `json:"secondary_registration"`
	} `json:"code_data"`
	Scopes []string `json:"scopes"`
}

// CodePairResponse represents the response from CBL code generation.
type CodePairResponse struct {
	PublicCode  string `json:"public_code"`
	PrivateCode string `json:"private_code"`
}

// TokenRefreshRequest represents a token refresh request.
type TokenRefreshRequest struct {
	AppName            string `json:"app_name"`
	AppVersion         string `json:"app_version"`
	SourceTokenType    string `json:"source_token_type"`
	SourceToken        string `json:"source_token"`
	RequestedTokenType string `json:"requested_token_type"`
	DeviceMetadata     struct {
		DeviceType   string `json:"device_type"`
		DeviceModel  string `json:"device_model"`
		OSVersion    string `json:"os_version"`
		DeviceSerial string `json:"device_serial"`
		Manufacturer string `json:"manufacturer"`
	} `json:"device_metadata"`
}

// TokenRefreshResponse represents a token refresh response.
type TokenRefreshResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresInSeconds int    `json:"expires_in"`
}
