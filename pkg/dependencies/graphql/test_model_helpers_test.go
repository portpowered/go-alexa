//nolint:testpackage // Shared helpers are used by in-package transport tests.
package graphql

func testListEndpointsInput(endpointIDs []string, pagination *PaginationParams) ListEndpointsInput {
	var input ListEndpointsInput

	input.EndpointIds = endpointIDs

	if pagination != nil {
		input.PaginationParams = *pagination
	}

	return input
}

func testEndpointsQueryParams(pagination PaginationParams) EndpointsQueryParams {
	var params EndpointsQueryParams

	params.PaginationParams = pagination

	return params
}

func testPaginationParams(nextToken string, disablePagination bool) PaginationParams {
	var params PaginationParams

	params.NextToken = nextToken
	params.DisablePagination = disablePagination

	return params
}

func testFeatureControlRequest(endpointID string, featureName FeatureName, operationName FeatureOperationName) FeatureControlRequest {
	var request FeatureControlRequest

	request.EndpointId = endpointID
	request.FeatureName = featureName
	request.FeatureOperationName = operationName

	return request
}

func testQualityOfServiceConfiguration(durationInSeconds int) QualityOfServiceConfiguration {
	var configuration QualityOfServiceConfiguration

	configuration.DurationInSeconds = durationInSeconds

	return configuration
}

func testSubscriptionFilter(entityType string, ids []string) SubscriptionFilter {
	var filter SubscriptionFilter

	filter.EntityType = entityType
	filter.Ids = ids

	return filter
}

func testSubscribeConfiguration(entities []SubscriptionFilter, durationInMinutes int) SubscribeConfiguration {
	var configuration SubscribeConfiguration

	configuration.Entities = entities
	configuration.DurationInMinutes = durationInMinutes

	return configuration
}
