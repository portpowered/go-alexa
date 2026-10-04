package alexa

import (
	"context"
	"fmt"
	"strings"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
)

// Selected state reads use provider endpoint IDs and require no account-wide
// DevicesV2 metadata join. Discovery remains responsible for that metadata.
func (c *Session) selectedEndpointStates(ctx context.Context, endpointIDs []string) (*alexaapimodels.UnifiedEndpointListResponse, error) {
	ids, err := selectedStateEndpointIDs(endpointIDs)
	if err != nil {
		return nil, err
	}

	input := listEndpointsInput(true, ids)

	response, err := c.graphqlClient.ListEndpointsWithStates(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("read selected endpoint states: %w", err)
	}

	endpoints := c.mergeEndpoints(nil, response.ListEndpoints.Endpoints, endpointMergeData{deviceV2Map: nil, includeFeatures: true, includeCapabilities: true})
	endpoints = filterEndpointsByIDs(endpoints, ids)
	missing := missingStateEndpointIDs(ids, endpoints)

	missing = append(missing, incompleteAQMStateIDs(ids, response.ListEndpoints.Endpoints)...)

	if len(missing) > 0 {
		// Amazon can omit AQM identities or their sensor properties. Supplement
		// only absent/incomplete selected identities, never the whole account.
		input.EndpointIds = missing

		extra, err := c.listAirQualityMonitorEndpoints(ctx, input, true)
		if err != nil {
			return nil, err
		}

		endpoints = supplementSelectedStateFeatures(endpoints, filterEndpointsByIDs(extra, missing))
	}

	return &alexaapimodels.UnifiedEndpointListResponse{Results: endpoints}, nil
}

func incompleteAQMStateIDs(ids []string, endpoints []graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint) []string {
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}

	var incomplete []string

	for _, endpoint := range endpoints {
		if !selected[endpoint.EndpointId] || endpoint.DisplayCategories.Primary.Value != string(alexaapimodels.EndpointDisplayCategoryAirQualityMonitor) {
			continue
		}

		hasSensorProperties := false

		for _, feature := range endpoint.Features {
			if feature.Name == string(alexaapimodels.FeatureNameRange) && len(feature.Properties) > 0 {
				hasSensorProperties = true
			}
		}

		if !hasSensorProperties {
			incomplete = append(incomplete, endpoint.EndpointId)
		}
	}

	return incomplete
}

func supplementSelectedStateFeatures(primary, extra []*alexaapimodels.Endpoint) []*alexaapimodels.Endpoint {
	byID := make(map[string]*alexaapimodels.Endpoint, len(primary))
	for _, endpoint := range primary {
		byID[endpoint.EndpointID] = endpoint
	}

	for _, endpoint := range extra {
		original := byID[endpoint.EndpointID]
		if original == nil {
			primary = append(primary, endpoint)
			byID[endpoint.EndpointID] = endpoint

			continue
		}

		for _, feature := range endpoint.Features {
			found := false

			for i, existing := range original.Features {
				if existing.Name == feature.Name && existing.Instance == feature.Instance {
					found = true

					if len(existing.Properties) == 0 {
						original.Features[i] = feature
					}

					break
				}
			}

			if !found {
				original.Features = append(original.Features, feature)
			}
		}
	}

	return primary
}

func selectedStateEndpointIDs(endpointIDs []string) ([]string, error) {
	ids := make([]string, 0, len(endpointIDs))

	seen := make(map[string]bool, len(endpointIDs))

	for _, endpointID := range endpointIDs {
		if strings.TrimSpace(endpointID) == "" {
			return nil, &alexaapimodels.BadRequestError{Message: "selected state query requires nonempty endpoint IDs"}
		}

		if !seen[endpointID] {
			ids = append(ids, endpointID)
			seen[endpointID] = true
		}
	}

	return ids, nil
}

func missingStateEndpointIDs(ids []string, endpoints []*alexaapimodels.Endpoint) []string {
	returned := make(map[string]bool, len(endpoints))
	for _, endpoint := range endpoints {
		returned[endpoint.EndpointID] = true
	}

	missing := make([]string, 0)

	for _, id := range ids {
		if !returned[id] {
			missing = append(missing, id)
		}
	}

	return missing
}

func filterEndpointsByIDs(endpoints []*alexaapimodels.Endpoint, ids []string) []*alexaapimodels.Endpoint {
	if len(ids) == 0 {
		return endpoints
	}

	allowed := make(map[string]bool, len(ids))
	for _, id := range ids {
		allowed[id] = true
	}

	filtered := make([]*alexaapimodels.Endpoint, 0, len(endpoints))

	for _, endpoint := range endpoints {
		if endpoint != nil && allowed[endpoint.EndpointID] {
			filtered = append(filtered, endpoint)
		}
	}

	return filtered
}

func listEndpointsInput(includeStates bool, endpointIDs []string) graphql.ListEndpointsInput {
	input := graphql.ListEndpointsInput{
		MaxAgeInMillis:          0,
		LatencyTolerance:        "",
		DisplayCategory:         "",
		AllDisplayCategories:    "",
		Enablement:              "",
		Filters:                 nil,
		FilterExpressions:       nil,
		QueryExpression:         nil,
		MaxPagesToFetch:         0,
		EndpointIds:             append([]string(nil), endpointIDs...),
		IncludeHouseholdDevices: false,
		PaginationParams: graphql.PaginationParams{
			PageSize:          defaultEndpointPageSize,
			DisablePagination: true,
			NextToken:         "",
		},
	}
	if includeStates {
		input.MaxAgeInMillis = 1
		input.LatencyTolerance = graphql.LatencyToleranceValueLow
	}

	return input
}
