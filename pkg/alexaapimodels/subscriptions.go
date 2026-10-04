package alexaapimodels

// QualityOfServiceExperience represents the type of quality of service experience.
type QualityOfServiceExperience string

// QualityOfServiceRequest represents a request to request quality of service for endpoints.
//
//modelinventory:client-input QualityOfServiceRequest: endpoint list and experience settings converted to the generated GraphQL request.
type QualityOfServiceRequest struct {
	// The list of endpoints to configure quality of service for.
	Endpoints     []string                      `json:"endpoints"`
	Configuration QualityOfServiceConfiguration `json:"configuration"`
}

// QualityOfServiceConfiguration represents the quality of service configuration.
//
//modelinventory:client-input QualityOfServiceConfiguration: experience duration and type nested in a caller quality-of-service request.
type QualityOfServiceConfiguration struct {
	// The type of experience we want to optimize data polling for state should go.
	TypeOfExperience QualityOfServiceExperience `json:"typeOfExperience"`
	// How long polling should occur
	DurationInSeconds int `json:"durationInSeconds"`
}

// SubscribeRequest represents a request to subscribe to endpoint events.
//
//modelinventory:client-input SubscribeRequest: requested subscription duration and entity filters converted to GraphQL.
type SubscribeRequest struct {
	Entities          []SubscribeEntity `json:"entities"`
	DurationInMinutes int               `json:"durationInMinutes"`
}

// SubscribeEntity selects an entity type for an event subscription.
//
//modelinventory:client-input SubscribeEntity: entity type selecting which endpoint events the subscription includes.
type SubscribeEntity struct {
	// Type of entity to subscribe to.
	EntityType EntityType `json:"entityType"`
}

// EntityType identifies the kind of events requested by a subscription.
type EntityType string

// QualityOfServiceResponse represents the response from a quality of service request.
//
//modelinventory:semantic QualityOfServiceResponse: caller-facing result and errors from a quality-of-service request.
type QualityOfServiceResponse struct {
	DurationInSeconds int                     `json:"durationInSeconds"`
	Errors            []QualityOfServiceError `json:"errors,omitempty"`
}

// QualityOfServiceError represents an error from a quality of service operation.
//
//modelinventory:semantic QualityOfServiceError: typed quality-of-service error message returned by the response converter.
type QualityOfServiceError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// SubscribeResponse represents the response from a subscribe request.
//
//modelinventory:semantic SubscribeResponse: caller-facing duration, entity results, and errors from a subscription request.
type SubscribeResponse struct {
	Entities          []string         `json:"entities,omitempty"`
	Errors            []SubscribeError `json:"errors,omitempty"`
	DurationInMinutes int              `json:"durationInMinutes"`
}

// SubscribeError represents an error from a subscribe operation.
//
//modelinventory:semantic SubscribeError: typed subscription error message returned by the response converter.
type SubscribeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
