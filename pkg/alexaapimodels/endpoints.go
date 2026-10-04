// Package alexaapimodels provides unified API models for Alexa services.
package alexaapimodels

// EndpointInterface represents a device endpoint that can be used for behavior operations
// This interface provides unified access to endpoint information from different sources.
type EndpointInterface interface {
	GetDeviceType() string
	GetDeviceSerialNumber() string
	GetLocale() string
	GetEndpointId() string
	GetDeviceFamily() string
	GetDeviceAccountId() string
}

// Endpoint represents a unified endpoint that combines GraphQL endpoint data
// with DeviceV2 data through a left outer join on DMSIdentifier
// The primary key is the ID from the GraphQL API response.
//
//modelinventory:semantic Endpoint: unified device identity and capability data returned by endpoint enumeration.
type Endpoint struct {
	// Primary key from GraphQL (always present)
	ID string

	// Primary data from GraphQL
	EndpointID         string
	DeviceType         string // From DMSIdentifier
	DeviceSerialNumber string // From DMSIdentifier (device serial number)

	// Endpoint metadata from GraphQL
	FriendlyName      *NameValue        `json:"friendlyName,omitempty"` // User-friendly name of the endpoint
	Manufacturer      *NameValue        `json:"manufacturer,omitempty"` // Manufacturer name
	Model             *NameValue        `json:"model,omitempty"`        // Device model
	Description       *NameValue        `json:"description,omitempty"`  // Device description
	SerialNumber      *NameValue        `json:"serialNumber,omitempty"` // Serial number (may differ from DeviceSerialNumber)
	SoftwareVersion   *NameValue        `json:"softwareVersion,omitempty"`
	DisplayCategories DisplayCategories `json:"displayCategories,omitempty"`
	// Secondary data from DeviceV2 API (may be empty if no match found)
	DeviceFamily    string
	DeviceAccountId string
	Locale          string

	// Features lists the features this endpoint supports
	// This is computed by merging GraphQL features, REST capabilities, and device family information
	// Feature names correspond to the operations available in the ClientInterface
	// When states are requested, features include their detailed property states
	Features []Feature `json:"features,omitempty"`
}

// DisplayCategories contains the primary and additional categories of an endpoint.
//
//modelinventory:semantic DisplayCategories: normalized primary and secondary device categories exposed on an endpoint.
type DisplayCategories struct {
	Primary   EndpointDisplayCategory   `json:"primary,omitempty"`
	Secondary []EndpointDisplayCategory `json:"secondary,omitempty"`
}

// NameValue represents a name-value pair with type information.
//
//modelinventory:semantic NameValue: optional named metadata value used for endpoint labels and details.
type NameValue struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// Feature describes a capability exposed by an endpoint.
//
//modelinventory:semantic Feature: capability name, instances, properties, operations, and configuration in the unified endpoint result.
type Feature struct {
	Name       FeatureName        `json:"name"`
	Instance   string             `json:"instance,omitempty"`
	Config     *FeatureConfig     `json:"config,omitempty"`
	Properties []FeatureProperty  `json:"properties,omitempty"`
	Operations []FeatureOperation `json:"operations,omitempty"`
}

// FeatureConfig contains configuration details for a feature.
//
//modelinventory:semantic FeatureConfig: public configuration for a feature after generated GraphQL state conversion.
type FeatureConfig struct {
	Range *RangeConfig `json:"range,omitempty"`
}

// RangeConfig describes the units, supported bounds, and presets for a range feature.
//
//modelinventory:semantic RangeConfig: range-specific configuration attached to a converted feature.
type RangeConfig struct {
	FriendlyName   *NameValue      `json:"friendlyName,omitempty"`
	UnitOfMeasure  *NameValue      `json:"unitOfMeasure,omitempty"`
	SupportedRange *SupportedRange `json:"supportedRange,omitempty"`
	Presets        []RangePreset   `json:"presets,omitempty"`
}

// SupportedRange describes the minimum, maximum, and precision of a range feature.
//
//modelinventory:semantic SupportedRange: minimum, maximum, and precision metadata for a supported range feature.
type SupportedRange struct {
	MinimumValue float64 `json:"minimumValue"`
	MaximumValue float64 `json:"maximumValue"`
	Precision    float64 `json:"precision"`
}

// RangePreset pairs a supported range value with its display name.
//
//modelinventory:semantic RangePreset: named range preset converted from endpoint capability data.
type RangePreset struct {
	RangeValue   float64    `json:"rangeValue"`
	FriendlyName *NameValue `json:"friendlyName,omitempty"`
}

// FeatureProperty represents a property of a feature with its state.
//
//modelinventory:semantic FeatureProperty: state and error value attached to a public feature property.
type FeatureProperty struct {
	Name             string         `json:"name"`
	Type             string         `json:"type,omitempty"`
	Accuracy         string         `json:"accuracy,omitempty"`
	TimeOfSample     string         `json:"timeOfSample,omitempty"`
	TimeOfLastChange string         `json:"timeOfLastChange,omitempty"`
	Error            *PropertyError `json:"error,omitempty"`
	// StateValue holds the actual state value as JSON, which can be deserialized
	// into specific types like VolumeState, PowerState, etc. based on the property name
	StateValue interface{} `json:"stateValue,omitempty"`
}

// RangeValueState is the current numeric value of a range feature.
//
//modelinventory:semantic RangeValueState: present range state value, including an explicit zero returned by enumeration.
type RangeValueState struct {
	Value float64 `json:"value"`
}

// PropertyError represents an error for a property.
//
//modelinventory:semantic PropertyError: normalized error type and message attached to an unavailable endpoint property.
type PropertyError struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

// FeatureOperation represents an operation available on a feature.
//
//modelinventory:semantic FeatureOperation: supported control operation attached to a public feature.
type FeatureOperation struct {
	Name string `json:"name"`
}

// IsType checks if the feature matches the given control namespace.
func (f Feature) IsType(namespace FeatureName) bool {
	return f.Name == namespace
}

// UnifiedEndpointListResponse represents the response from listing unified endpoints.
//
//modelinventory:semantic UnifiedEndpointListResponse: paginated endpoint enumeration result with its source completeness metadata.
type UnifiedEndpointListResponse struct {
	Results   []*Endpoint `json:"results"`
	NextToken string      `json:"nextToken,omitempty"`
}

// GetDeviceType returns the device type.
func (e *Endpoint) GetDeviceType() string {
	return e.DeviceType
}

// GetDeviceSerialNumber returns the device serial number (DSN).
func (e *Endpoint) GetDeviceSerialNumber() string {
	return e.DeviceSerialNumber
}

// GetLocale returns the locale.
func (e *Endpoint) GetLocale() string {
	if e.Locale == "" {
		return DefaultLocale
	}

	return e.Locale
}

// GetEndpointId returns the endpoint ID from GraphQL.
func (e *Endpoint) GetEndpointId() string {
	return e.EndpointID
}

// GetDeviceFamily returns the device family from DeviceV2.
func (e *Endpoint) GetDeviceFamily() string {
	return e.DeviceFamily
}

// GetDeviceAccountId returns the device account ID from DeviceV2.
func (e *Endpoint) GetDeviceAccountId() string {
	return e.DeviceAccountId
}

// HasFeature reports whether the endpoint exposes the named feature.
func (e *Endpoint) HasFeature(featureName FeatureName) bool {
	for _, feature := range e.Features {
		if feature.IsType(featureName) {
			return true
		}
	}

	return false
}
