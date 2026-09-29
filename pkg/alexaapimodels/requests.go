// Package alexaapimodels provides unified API models for Alexa services.
// These models consolidate structures from both GraphQL and REST APIs.
package alexaapimodels

// ProviderID identifies a supported media provider.
type ProviderID string

const (
	// ProviderIDAmazon identifies Amazon Music.
	ProviderIDAmazon ProviderID = "AMAZON_MUSIC"
	// ProviderIdDefault identifies the default media provider.
	ProviderIdDefault ProviderID = "DEFAULT"
	// ProviderIDSpotify identifies Spotify.
	ProviderIDSpotify ProviderID = "SPOTIFY"
	// ProviderIDAudible identifies Audible.
	ProviderIDAudible ProviderID = "AUDIBLE"
)

// FeatureControlResponse represents a unified response for feature control operations.
type FeatureControlResponse struct {
	FeatureControlResponses []FeatureControlResponseItem `json:"featureControlResponses,omitempty"`
	Errors                  []FeatureControlError        `json:"errors,omitempty"`
}

// FeatureControlResponseItem represents a single feature control response.
type FeatureControlResponseItem struct {
	EndpointID           string      `json:"endpointId,omitempty"`
	FeatureName          string      `json:"featureName,omitempty"`
	Instance             string      `json:"instance,omitempty"`
	FeatureOperationName string      `json:"featureOperationName,omitempty"`
	Payload              interface{} `json:"payload,omitempty"`
	Code                 string      `json:"code,omitempty"`
	ResponsePayload      interface{} `json:"responsePayload,omitempty"`
}

// FeatureControlError represents an error from a feature control operation.
type FeatureControlError struct {
	EndpointID           string      `json:"endpointId,omitempty"`
	FeatureName          string      `json:"featureName,omitempty"`
	Instance             string      `json:"instance,omitempty"`
	FeatureOperationName string      `json:"featureOperationName,omitempty"`
	Payload              interface{} `json:"payload,omitempty"`
	Code                 string      `json:"code,omitempty"`
	Message              string      `json:"message,omitempty"`
}

// ControlResponse is an alias for FeatureControlResponse
// This provides a unified response type for the Control method.
type ControlResponse = FeatureControlResponse

// DeviceRegistrationConfig contains device information for MAP authentication.
type DeviceRegistrationConfig struct {
	AppName      string `json:"appName"`
	AppVersion   string `json:"appVersion"`
	DeviceType   string `json:"deviceType"`
	Domain       string `json:"domain"`
	DeviceModel  string `json:"deviceModel"`
	OSVersion    string `json:"osVersion"`
	DeviceSerial string `json:"deviceSerial"`
	DeviceName   string `json:"deviceName"`
	Manufacturer string `json:"manufacturer"`
}
