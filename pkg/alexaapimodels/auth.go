package alexaapimodels

// CodePairResponse represents the response from CBL code generation.
//modelinventory:semantic CodePairResponse: browser login codes returned by the code-pair endpoint.
type CodePairResponse struct {
	PublicCode  string `json:"publicCode"`
	PrivateCode string `json:"privateCode"`
}

// RegistrationResponse represents a successful MAP registration response.
//modelinventory:semantic RegistrationResponse: access and refresh credentials returned after device pairing.
type RegistrationResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// TokenRefreshRequest represents a token refresh request.
//modelinventory:client-input TokenRefreshRequest: refresh credential and device registration options passed into the explicit token exchange.
type TokenRefreshRequest struct {
	RefreshToken string                   `json:"refreshToken"`
	Config       DeviceRegistrationConfig `json:"config"`
}

// TokenRefreshResponse represents a token refresh response.
//modelinventory:semantic TokenRefreshResponse: caller-facing access token and expiry projection from the generated OAuth response.
type TokenRefreshResponse struct {
	AccessToken      string `json:"accessToken"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// DefaultDeviceRegistrationConfig returns a default device registration config
// This is a convenience function for creating a DeviceRegistrationConfig with sensible defaults.
func DefaultDeviceRegistrationConfig(deviceSerial, deviceName string) DeviceRegistrationConfig {
	return DeviceRegistrationConfig{
		AppName:      "Client SDK",
		AppVersion:   "1.0",
		DeviceType:   DeviceTypeIphone, // DeviceTypeSimulator
		Domain:       "Device",
		DeviceModel:  "Client SDK",
		OSVersion:    "0",
		DeviceSerial: deviceSerial,
		DeviceName:   deviceName,
		Manufacturer: "Amazon",
	}
}

// UserInfoRequest contains options for getting user info.
//modelinventory:client-input UserInfoRequest: optional platform, client version, and CSRF values mapped to the profile request.
type UserInfoRequest struct {
	// Platform identifier (e.g., "ios", "android")
	Platform string `json:"platform,omitempty"`
	// App version (e.g., "2.2.556530.0")
	Version string `json:"version,omitempty"`
	// CSRF token for cookie header
	CSRFToken string `json:"csrfToken,omitempty"`
}

// UserInfo represents user information from the /api/users/me endpoint.
//modelinventory:semantic UserInfo: account profile fields returned by the SDK profile lookup.
type UserInfo struct {
	// This is the country code for country of residence for the customer.
	CountryOfResidence string `json:"countryOfResidence"`
	// This is the amazon marketplace ID for the user.
	EffectiveMarketPlaceID string   `json:"effectiveMarketPlaceId"`
	Email                  string   `json:"email"`
	EulaAcceptance         bool     `json:"eulaAcceptance"`
	Features               []string `json:"features"`
	FullName               string   `json:"fullName"`
	HasActiveDopplers      bool     `json:"hasActiveDopplers"`
	// This ID is the user's unique identifier in the amazon account system.
	ID                    string `json:"id"`
	MarketPlaceDomainName string `json:"marketPlaceDomainName"`
	MarketPlaceID         string `json:"marketPlaceId"`
	MarketPlaceLocale     string `json:"marketPlaceLocale"`
}
