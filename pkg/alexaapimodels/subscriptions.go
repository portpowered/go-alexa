package alexaapimodels

// QualityOfServiceExperience represents the type of quality of service experience.
type QualityOfServiceExperience string

const (
	// QualityOfServiceExperienceBackgroundEphemeral prioritizes short-lived background polling.
	QualityOfServiceExperienceBackgroundEphemeral QualityOfServiceExperience = "BACKGROUND_EPHEMERAL"
	// QualityOfServiceExperienceBackgroundPersistent keeps background polling active.
	QualityOfServiceExperienceBackgroundPersistent QualityOfServiceExperience = "BACKGROUND_PERSISTENT"
	// QualityOfServiceExperienceForegroundPersistent keeps foreground polling active.
	QualityOfServiceExperienceForegroundPersistent QualityOfServiceExperience = "FOREGROUND_PERSISTENT"
)

// QualityOfServiceRequest represents a request to request quality of service for endpoints.
type QualityOfServiceRequest struct {
	// The list of endpoints to configure quality of service for.
	Endpoints     []string                      `json:"endpoints"`
	Configuration QualityOfServiceConfiguration `json:"configuration"`
}

// QualityOfServiceConfiguration represents the quality of service configuration.
type QualityOfServiceConfiguration struct {
	// The type of experience we want to optimize data polling for state should go.
	TypeOfExperience QualityOfServiceExperience `json:"typeOfExperience"`
	// How long polling should occur
	DurationInSeconds int `json:"durationInSeconds"`
}

// SubscribeRequest represents a request to subscribe to endpoint events.
type SubscribeRequest struct {
	Entities          []SubscribeEntity `json:"entities"`
	DurationInMinutes int               `json:"durationInMinutes"`
}

// SubscribeEntity selects an entity type for an event subscription.
type SubscribeEntity struct {
	// Type of entity to subscribe to.
	EntityType EntityType `json:"entityType"`
}

// EntityType identifies the kind of events requested by a subscription.
type EntityType string

const (
	// EntityTypeEndpoint selects endpoint lifecycle events.
	EntityTypeEndpoint EntityType = "Endpoint"
	// EntityTypeState selects endpoint state-change events.
	EntityTypeState EntityType = "State"
)

// QualityOfServiceResponse represents the response from a quality of service request.
type QualityOfServiceResponse struct {
	DurationInSeconds int                     `json:"durationInSeconds"`
	Errors            []QualityOfServiceError `json:"errors,omitempty"`
}

// QualityOfServiceError represents an error from a quality of service operation.
type QualityOfServiceError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SubscribeResponse represents the response from a subscribe request.
type SubscribeResponse struct {
	Entities          []string         `json:"entities,omitempty"`
	Errors            []SubscribeError `json:"errors,omitempty"`
	DurationInMinutes int              `json:"durationInMinutes"`
}

// SubscribeError represents an error from a subscribe operation.
type SubscribeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
