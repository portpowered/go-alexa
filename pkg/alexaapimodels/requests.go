// Package alexaapimodels provides unified API models for Alexa services.
// These models consolidate structures from both GraphQL and REST APIs.
package alexaapimodels

// ProviderID identifies a supported media provider.
type ProviderID string

// FeatureControlResponse represents a unified response for feature control operations.
//
//modelinventory:semantic FeatureControlResponse: public result of a feature-control GraphQL operation.
type FeatureControlResponse struct {
	FeatureControlResponses []FeatureControlResponseItem `json:"featureControlResponses,omitempty"`
	Errors                  []FeatureControlError        `json:"errors,omitempty"`
}

// FeatureControlResponseItem represents a single feature control response.
//
//modelinventory:semantic FeatureControlResponseItem: successful operation result with endpoint, feature, operation, and payload fields.
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
//
//modelinventory:semantic FeatureControlError: failed feature-control result with typed operation context and message.
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
//
//modelinventory:client-input DeviceRegistrationConfig: device identity and app metadata accepted by code-pair, registration, and token-refresh calls.
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
