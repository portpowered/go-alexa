package graphql

import (
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// ConvertToFeatureControlResponse converts a GraphQL SetEndpointFeaturesResponse to alexaapimodels format.
func ConvertToFeatureControlResponse(gqlResp *SetEndpointFeaturesResponse) *alexaapimodels.FeatureControlResponse {
	if gqlResp == nil {
		return nil
	}

	response := &alexaapimodels.FeatureControlResponse{
		FeatureControlResponses: make([]alexaapimodels.FeatureControlResponseItem, 0),
		Errors:                  make([]alexaapimodels.FeatureControlError, 0),
	}

	// Convert feature control responses
	setEndpointFeatures := gqlResp.GetSetEndpointFeatures()
	for _, resp := range setEndpointFeatures.GetFeatureControlResponses() {
		// Convert response payload to interface{} for unified format
		var responsePayload interface{}
		if respPayload := resp.GetResponsePayload(); respPayload != nil {
			// The response payload is a complex type, we'll convert it to a generic interface
			responsePayload = respPayload
		}

		item := alexaapimodels.FeatureControlResponseItem{
			EndpointID:           resp.GetEndpointId(),
			FeatureName:          string(resp.GetFeatureName()),
			Instance:             resp.GetInstance(),
			FeatureOperationName: string(resp.GetFeatureOperationName()),
			Payload:              resp.GetPayload(),
			Code:                 resp.GetCode(),
			ResponsePayload:      responsePayload,
		}
		response.FeatureControlResponses = append(response.FeatureControlResponses, item)
	}

	// Convert errors
	for _, err := range setEndpointFeatures.GetErrors() {
		errorItem := alexaapimodels.FeatureControlError{
			EndpointID:           err.GetEndpointId(),
			FeatureName:          string(err.GetFeatureName()),
			Instance:             err.GetInstance(),
			FeatureOperationName: string(err.GetFeatureOperationName()),
			Payload:              err.GetPayload(),
			Code:                 err.GetCode(),
			Message:              err.GetMessage(),
		}
		response.Errors = append(response.Errors, errorItem)
	}

	return response
}

// ConvertQualityOfServiceRequest converts alexaapimodels.QualityOfServiceRequest to graphql.EndpointQualityOfServiceInput.
func ConvertQualityOfServiceRequest(req *alexaapimodels.QualityOfServiceRequest) EndpointQualityOfServiceInput {
	return EndpointQualityOfServiceInput{
		Endpoints: req.Endpoints,
		Configuration: QualityOfServiceConfiguration{
			TypeOfExperience:  QualityOfServiceExperience(req.Configuration.TypeOfExperience),
			DurationInSeconds: req.Configuration.DurationInSeconds,
		},
	}
}

// ConvertToQualityOfServiceResponse converts graphql.RequestEndpointQualityOfServiceResponse to alexaapimodels.QualityOfServiceResponse.
func ConvertToQualityOfServiceResponse(gqlResp *RequestEndpointQualityOfServiceResponse) *alexaapimodels.QualityOfServiceResponse {
	if gqlResp == nil {
		return nil
	}

	response := gqlResp.GetRequestEndpointQualityOfService()
	result := &alexaapimodels.QualityOfServiceResponse{
		DurationInSeconds: response.GetDurationInSeconds(),
		Errors:            make([]alexaapimodels.QualityOfServiceError, 0),
	}

	// Convert errors
	for _, err := range response.GetErrors() {
		errorItem := alexaapimodels.QualityOfServiceError{
			Type:    err.GetType(),
			Message: err.GetMessage(),
		}
		result.Errors = append(result.Errors, errorItem)
	}

	return result
}

// ConvertSubscribeRequest converts alexaapimodels.SubscribeRequest to graphql.SubscribeConfiguration.
func ConvertSubscribeRequest(req *alexaapimodels.SubscribeRequest) SubscribeConfiguration {
	return SubscribeConfiguration{
		// NOTE: the GraphQL format allows us to know which endpoint the event is coming from, this not possible with the default format

		Format:            SubscriptionEventFormatGraphql,
		Entities:          ConvertSubscribeEntities(req.Entities),
		Reset:             true,
		Version:           "2",
		DurationInMinutes: req.DurationInMinutes,
	}
}

// ConvertSubscribeEntities converts alexaapimodels.SubscribeEntity to graphql.SubscriptionFilter.
func ConvertSubscribeEntities(entities []alexaapimodels.SubscribeEntity) []SubscriptionFilter {
	filters := make([]SubscriptionFilter, 0, len(entities))
	for _, entity := range entities {
		filters = append(filters, SubscriptionFilter{
			EntityType: string(entity.EntityType),
			Ids:        nil,
		})
	}

	return filters
}

// ConvertToSubscribeResponse converts graphql.SubscribeResponse to alexaapimodels.SubscribeResponse.
func ConvertToSubscribeResponse(gqlResp *SubscribeResponse) *alexaapimodels.SubscribeResponse {
	if gqlResp == nil {
		return nil
	}

	response := gqlResp.GetSubscribe()
	result := &alexaapimodels.SubscribeResponse{
		Entities:          response.GetEntities(),
		DurationInMinutes: response.GetDurationInMinutes(),
		Errors:            make([]alexaapimodels.SubscribeError, 0),
	}

	// Convert errors
	for _, err := range response.GetErrors() {
		errorItem := alexaapimodels.SubscribeError{
			Type:    err.GetType(),
			Message: err.GetMessage(),
		}
		result.Errors = append(result.Errors, errorItem)
	}

	return result
}
