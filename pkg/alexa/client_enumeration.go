package alexa

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

const defaultEndpointPageSize = 50

type (
	gqlEndpointFeaturesFeatureConfigurationRangeConfiguration              = graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeatureConfigurationRangeConfiguration
	gqlEndpointFeaturesFeatureConfigurationRangeConfigurationPresetsPreset = graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeatureConfigurationRangeConfigurationPresetsPreset
	gqlListEndpointDisplayCategories                                       = graphql.ListEndpointsListEndpointsListEndpointsResponseEndpointsEndpointDisplayCategories
	gqlStateFeatureConfiguration                                           = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeatureConfiguration
	gqlStateRangeConfiguration                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeatureConfigurationRangeConfiguration
	gqlStateRangePreset                                                    = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeatureConfigurationRangeConfigurationPresetsPreset
	gqlStatePropertyActionState                                            = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesActionState
	gqlStatePropertyArmState                                               = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesArmState
	gqlStatePropertyBattery                                                = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBattery
	gqlStatePropertyBrightness                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBrightness
	gqlStatePropertyBurglaryAlarm                                          = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBurglaryAlarm
	gqlStatePropertyCarbonMonoxideAlarm                                    = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesCarbonMonoxideAlarm
	gqlStatePropertyColor                                                  = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesColor
	gqlStatePropertyColorTemperature                                       = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesColorTemperature
	gqlStatePropertyFeatureProperty                                        = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFeatureProperty
	gqlStatePropertyFireAlarm                                              = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFireAlarm
	gqlStatePropertyGeolocation                                            = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesGeolocation
	gqlStatePropertyIlluminance                                            = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesIlluminance
	gqlStatePropertyLock                                                   = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesLock
	gqlStatePropertyMode                                                   = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesMode
	gqlStatePropertyNetworkThroughput                                      = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesNetworkThroughput
	gqlStatePropertyPercentage                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPercentage
	gqlStatePropertyPower                                                  = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPower
	gqlStatePropertyPowerLevel                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPowerLevel
	gqlStatePropertyRadioDiagnostics                                       = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRadioDiagnostics
	gqlStatePropertyRangeValue                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRangeValue
	gqlStatePropertyReachability                                           = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesReachability
	gqlStatePropertyRelativeHumidity                                       = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRelativeHumidity
	gqlStatePropertySetpoint                                               = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSetpoint
	gqlStatePropertySirenState                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSirenState
	gqlStatePropertySnapshot                                               = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSnapshot
	gqlStatePropertyStatusCode                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesStatusCode
	gqlStatePropertyTemperatureSensor                                      = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesTemperatureSensor
	gqlStatePropertyThermostatMode                                         = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesThermostatMode
	gqlStatePropertyToggleState                                            = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesToggleState
	gqlStatePropertyVolume                                                 = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesVolume
	gqlStatePropertyWaterAlarm                                             = graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesWaterAlarm
)

// ListEndpoints retrieves all endpoints by joining GraphQL endpoint data
// with DeviceV2 data using a left outer join on DMSIdentifier.
// The query parameter can specify filters and which optional fields to include
// (states, capabilities, features). When States is true in IncludeFields,
// the function will retrieve endpoints with detailed property states.
func (c *Session) ListEndpoints(
	ctx context.Context,
	query alexaapimodels.EndpointQuery,
) (*alexaapimodels.UnifiedEndpointListResponse, error) {
	// Determine what optional fields to include
	includeStates := false
	includeCapabilities := false
	includeFeatures := false

	if query.IncludeFields != nil {
		includeStates = query.IncludeFields.Properties
		includeCapabilities = query.IncludeFields.Features
	}
	// Always include capabilities if not explicitly disabled, as we need them for operation detection
	// We'll fetch capabilities to determine supported operations
	if !includeCapabilities {
		includeCapabilities = true
	}

	// Build GraphQL input from query parameters
	graphqlInput := graphql.ListEndpointsInput{
		MaxAgeInMillis:          0,
		LatencyTolerance:        "",
		DisplayCategory:         "",
		AllDisplayCategories:    "",
		Enablement:              "",
		Filters:                 nil,
		FilterExpressions:       nil,
		QueryExpression:         nil,
		MaxPagesToFetch:         0,
		EndpointIds:             nil,
		IncludeHouseholdDevices: false,
		PaginationParams: graphql.PaginationParams{
			PageSize:          defaultEndpointPageSize,
			DisablePagination: true,
			NextToken:         "",
		},
	}

	var (
		graphqlEndpointsWithoutStates []graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint
		graphqlEndpointsWithStates    []graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint
		err                           error
	)

	// Fetch GraphQL endpoints based on whether states are requested

	if includeStates {
		// Execute the GraphQL query with states
		gqlResp, err := c.graphqlClient.ListEndpointsWithStates(ctx, graphqlInput)
		if err != nil {
			return nil, err
		}

		graphqlEndpointsWithStates = gqlResp.ListEndpoints.Endpoints
	} else {
		// Fetch GraphQL endpoints without states (primary source)
		graphqlWithoutStatesResp, err := c.graphqlClient.ListEndpointsWithoutStates(ctx, graphql.EndpointsQueryParams{
			FeatureLatencyTolerance: graphql.FeatureLatencyToleranceValueLow,
			PaginationParams: graphql.PaginationParams{
				PageSize:          defaultEndpointPageSize,
				DisablePagination: true,
				NextToken:         "",
			},
			Filters:                 nil,
			IncludeHouseholdDevices: false,
		})
		if err != nil {
			return nil, err
		}

		graphqlEndpointsWithoutStates = graphqlWithoutStatesResp.Endpoints.Items
	}

	airQualityMonitorEndpoints, err := c.listAirQualityMonitorEndpoints(ctx, graphqlInput, includeStates)
	if err != nil {
		return nil, err
	}

	// Fetch DeviceV2 devices (secondary source)
	devicesV2Resp, err := c.restClient.GetDevicesV2(ctx, nil)
	if err != nil {
		// If DeviceV2 fetch fails, we still return GraphQL endpoints (left outer join)
		devicesV2Resp = &alexamodels.DevicesV2Response{Devices: []alexamodels.DeviceV2{}}
	}

	// Create a map for fast lookup of DeviceV2 by DeviceType + SerialNumber
	deviceV2Map := make(map[string]*alexamodels.DeviceV2)

	for i := range devicesV2Resp.Devices {
		device := &devicesV2Resp.Devices[i]
		key := device.DeviceType + ":" + device.SerialNumber
		deviceV2Map[key] = device
	}

	// Prepare merge data
	mergeData := endpointMergeData{
		deviceV2Map:         deviceV2Map,
		includeFeatures:     includeFeatures,
		includeCapabilities: includeCapabilities,
	}

	// Perform merge using the helper function
	unifiedEndpoints := c.mergeEndpoints(
		graphqlEndpointsWithoutStates,
		graphqlEndpointsWithStates,
		mergeData,
	)
	unifiedEndpoints = mergeDistinctEndpoints(unifiedEndpoints, airQualityMonitorEndpoints)

	return &alexaapimodels.UnifiedEndpointListResponse{
		Results: unifiedEndpoints,
	}, nil
}

func (c *Session) listAirQualityMonitorEndpoints(
	ctx context.Context,
	baseInput graphql.ListEndpointsInput,
	includeStates bool,
) ([]*alexaapimodels.Endpoint, error) {
	input := baseInput
	input.DisplayCategory = string(alexaapimodels.EndpointDisplayCategoryAirQualityMonitor)

	if includeStates {
		gqlResp, err := c.graphqlClient.ListEndpointsWithStates(ctx, input)
		if err != nil {
			return nil, err
		}

		mergeData := endpointMergeData{
			deviceV2Map:         map[string]*alexamodels.DeviceV2{},
			includeFeatures:     true,
			includeCapabilities: true,
		}

		return c.mergeEndpoints(nil, gqlResp.ListEndpoints.Endpoints, mergeData), nil
	}

	gqlResp, err := c.graphqlClient.ListEndpoints(ctx, input)
	if err != nil {
		return nil, err
	}

	endpoints := make([]*alexaapimodels.Endpoint, 0, len(gqlResp.ListEndpoints.Endpoints))
	for _, gqlEndpoint := range gqlResp.ListEndpoints.Endpoints {
		endpoints = append(endpoints, c.convertListEndpoint(gqlEndpoint))
	}

	return endpoints, nil
}

func mergeDistinctEndpoints(
	primary []*alexaapimodels.Endpoint,
	extra []*alexaapimodels.Endpoint,
) []*alexaapimodels.Endpoint {
	if len(extra) == 0 {
		return primary
	}

	seen := make(map[string]struct{}, len(primary)+len(extra))

	for _, endpoint := range primary {
		key := endpointDeduplicationKey(endpoint)
		if key != "" {
			seen[key] = struct{}{}
		}
	}

	merged := primary

	for _, endpoint := range extra {
		key := endpointDeduplicationKey(endpoint)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
		}

		merged = append(merged, endpoint)
	}

	return merged
}

func endpointDeduplicationKey(endpoint *alexaapimodels.Endpoint) string {
	if endpoint == nil {
		return ""
	}

	if endpoint.EndpointID != "" {
		return "endpoint:" + endpoint.EndpointID
	}

	if endpoint.ID != "" {
		return "id:" + endpoint.ID
	}

	if endpoint.DeviceType != "" || endpoint.DeviceSerialNumber != "" {
		return "dms:" + endpoint.DeviceType + ":" + endpoint.DeviceSerialNumber
	}

	return ""
}

// endpointMergeData holds the data structures needed for merging endpoints.
type endpointMergeData struct {
	deviceV2Map         map[string]*alexamodels.DeviceV2
	includeFeatures     bool
	includeCapabilities bool
}

// mergeEndpoints performs a left outer join of GraphQL endpoints with DeviceV2 and REST data
// It handles both endpoints with and without states based on which list is provided.
func (c *Session) mergeEndpoints(
	graphqlEndpointsWithoutStates []graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint,
	graphqlEndpointsWithStates []graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint,
	mergeData endpointMergeData,
) []*alexaapimodels.Endpoint {
	var unifiedEndpoints []*alexaapimodels.Endpoint

	// Handle endpoints with states
	if len(graphqlEndpointsWithStates) > 0 {
		unifiedEndpoints = make([]*alexaapimodels.Endpoint, 0, len(graphqlEndpointsWithStates))

		for _, gqlEndpoint := range graphqlEndpointsWithStates {
			unified := c.mergeEndpointWithStates(gqlEndpoint, mergeData)
			unifiedEndpoints = append(unifiedEndpoints, unified)
		}
	} else if len(graphqlEndpointsWithoutStates) > 0 {
		// Handle endpoints without states
		unifiedEndpoints = make([]*alexaapimodels.Endpoint, 0, len(graphqlEndpointsWithoutStates))

		for _, gqlEndpoint := range graphqlEndpointsWithoutStates {
			unified := c.mergeEndpointWithoutStates(gqlEndpoint, mergeData)
			unifiedEndpoints = append(unifiedEndpoints, unified)
		}
	}

	return unifiedEndpoints
}

// mergeEndpointWithStates merges a single GraphQL endpoint (with states) with DeviceV2 and REST data.
func (c *Session) mergeEndpointWithStates(
	gqlEndpoint graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint,
	mergeData endpointMergeData,
) *alexaapimodels.Endpoint {
	unified := &alexaapimodels.Endpoint{
		ID:         gqlEndpoint.Id,
		EndpointID: gqlEndpoint.EndpointId,
	}

	// Extract DMSIdentifier from GraphQL endpoint
	dmsInfo := gqlEndpoint.DmsIdentifier
	unified.DeviceType = dmsInfo.DeviceType
	unified.DeviceSerialNumber = dmsInfo.Dsn

	// Extract endpoint metadata
	// For FriendlyName, prefer FriendlyNameObject if available, otherwise use the string field
	if gqlEndpoint.FriendlyName != "" {
		unified.FriendlyName = &alexaapimodels.NameValue{
			Type:  alexamodels.DefaultFriendlyNameType, // Default type for string field
			Value: gqlEndpoint.FriendlyName,
		}
	}

	unified.Model = convertNameValueObject(gqlEndpoint.Model.Type, gqlEndpoint.Model.Value.Text)
	unified.SerialNumber = convertNameValueObject(gqlEndpoint.SerialNumber.Type, gqlEndpoint.SerialNumber.Value.Text)
	unified.Description = convertNameValueObject(gqlEndpoint.Description.Type, gqlEndpoint.Description.Value.Text)
	// Note: Manufacturer is not available in ListEndpointsWithStates response
	unified.Manufacturer = convertNameValueObject(gqlEndpoint.Manufacturer.Type, gqlEndpoint.Manufacturer.Value.Text)

	// Try to find matching DeviceV2 by DeviceType + SerialNumber
	key := unified.DeviceType + ":" + unified.DeviceSerialNumber
	if deviceV2, found := mergeData.deviceV2Map[key]; found {
		unified.DeviceFamily = deviceV2.DeviceFamily

		unified.DeviceAccountId = deviceV2.DeviceAccountId

		unified.DeviceOwnerCustomerID = deviceV2.DeviceOwnerCustomerId
		if deviceV2.Language != nil {
			unified.Locale = *deviceV2.Language
		}
	}

	// Extract supported features from feature names
	graphqlFeatures := make([]string, 0, len(gqlEndpoint.Features))
	for _, feat := range gqlEndpoint.Features {
		graphqlFeatures = append(graphqlFeatures, feat.Name)
	}

	supportedFeatures := c.determineSupportedFeatures(graphqlFeatures, []string{}, unified.DeviceFamily)

	// Create a map of GraphQL features with states for merging
	graphqlFeaturesMap := make(map[string]alexaapimodels.Feature)

	for _, gqlFeature := range gqlEndpoint.Features {
		feature := alexaapimodels.Feature{
			Name:     alexaapimodels.FeatureName(gqlFeature.Name),
			Instance: gqlFeature.Instance,
			Config:   convertFeatureConfigWithStates(gqlFeature.Configuration),
		}

		// Convert operations
		operations := make([]alexaapimodels.FeatureOperation, 0)
		for _, op := range gqlFeature.Operations {
			operations = append(operations, alexaapimodels.FeatureOperation{
				Name: op.Name,
			})
		}

		feature.Operations = operations

		// Convert properties with states
		properties := make([]alexaapimodels.FeatureProperty, 0)

		for _, gqlProp := range gqlFeature.Properties {
			prop := alexaapimodels.FeatureProperty{
				Name:             gqlProp.GetName(),
				Type:             string(gqlProp.GetType()),
				Accuracy:         string(gqlProp.GetAccuracy()),
				TimeOfSample:     c.extractTimeOfSample(gqlProp),
				TimeOfLastChange: c.extractTimeOfLastChange(gqlProp),
				StateValue:       c.extractStateValue(gqlProp),
			}

			// Extract error if present using type assertion
			prop.Error = c.extractError(gqlProp)

			properties = append(properties, prop)
		}

		feature.Properties = properties

		graphqlFeaturesMap[featureIdentity(feature)] = feature
	}

	// Merge supported features with GraphQL features that have states
	// Start with supported features and enrich with state information when available
	unified.Features = mergeSupportedAndGraphQLFeatures(supportedFeatures, graphqlFeaturesMap)

	return unified
}

// mergeEndpointWithoutStates merges a single GraphQL endpoint (without states) with DeviceV2 and REST data.
func (c *Session) mergeEndpointWithoutStates(
	gqlEndpoint graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint,
	mergeData endpointMergeData,
) *alexaapimodels.Endpoint {
	unified := &alexaapimodels.Endpoint{
		ID:         gqlEndpoint.Id,
		EndpointID: gqlEndpoint.EndpointId,
	}

	// Extract DMSIdentifier from GraphQL endpoint
	dmsInfo := gqlEndpoint.DmsIdentifier
	unified.DeviceType = dmsInfo.DeviceType
	unified.DeviceSerialNumber = dmsInfo.Dsn

	// Extract endpoint metadata
	// For FriendlyName, prefer FriendlyNameObject if available, otherwise use the string field
	if gqlEndpoint.FriendlyName != "" {
		unified.FriendlyName = &alexaapimodels.NameValue{
			Type:  alexamodels.DefaultFriendlyNameType, // Default type for string field
			Value: gqlEndpoint.FriendlyName,
		}
	}

	unified.Manufacturer = convertNameValueObject(gqlEndpoint.Manufacturer.Type, gqlEndpoint.Manufacturer.Value.Text)
	unified.Model = convertNameValueObject(gqlEndpoint.Model.Type, gqlEndpoint.Model.Value.Text)
	unified.SerialNumber = convertNameValueObject(gqlEndpoint.SerialNumber.Type, gqlEndpoint.SerialNumber.Value.Text)
	unified.Description = convertNameValueObject(gqlEndpoint.Description.Type, gqlEndpoint.Description.Value.Text)
	unified.DisplayCategories = convertDisplayCategories(gqlEndpoint.DisplayCategories)
	// Try to find matching DeviceV2 by DeviceType + SerialNumber
	key := unified.DeviceType + ":" + unified.DeviceSerialNumber
	if deviceV2, found := mergeData.deviceV2Map[key]; found {
		unified.DeviceFamily = deviceV2.DeviceFamily

		unified.DeviceAccountId = deviceV2.DeviceAccountId

		unified.DeviceOwnerCustomerID = deviceV2.DeviceOwnerCustomerId
		if deviceV2.Language != nil {
			unified.Locale = *deviceV2.Language
		}
	}

	// Extract supported features from feature names
	graphqlFeatures := make([]string, 0, len(gqlEndpoint.Features))
	for _, feat := range gqlEndpoint.Features {
		graphqlFeatures = append(graphqlFeatures, feat.Name)
	}

	// Compute supported features (without state information)
	legacyCapabilityInterfaces := extractLegacyCapabilityInterfaces(gqlEndpoint.LegacyAppliance.Capabilities)
	supportedFeatures := c.determineSupportedFeatures(graphqlFeatures, legacyCapabilityInterfaces, unified.DeviceFamily)

	// Create features map from GraphQL for operations/instance info
	graphqlFeaturesMap := make(map[string]alexaapimodels.Feature)

	for _, gqlFeature := range gqlEndpoint.Features {
		feature := alexaapimodels.Feature{
			Name:     alexaapimodels.FeatureName(gqlFeature.Name),
			Instance: gqlFeature.Instance,
			Config:   convertFeatureConfigWithoutStates(gqlFeature.Configuration),
		}

		// Convert operations
		operations := make([]alexaapimodels.FeatureOperation, 0)
		for _, op := range gqlFeature.Operations {
			operations = append(operations, alexaapimodels.FeatureOperation{
				Name: op.Name,
			})
		}

		feature.Operations = operations

		graphqlFeaturesMap[featureIdentity(feature)] = feature
	}

	// Merge supported features with GraphQL feature metadata (operations, instance)
	unified.Features = mergeSupportedAndGraphQLFeatures(supportedFeatures, graphqlFeaturesMap)

	return unified
}

func (c *Session) convertListEndpoint(
	gqlEndpoint graphql.ListEndpointsListEndpointsListEndpointsResponseEndpointsEndpoint,
) *alexaapimodels.Endpoint {
	unified := &alexaapimodels.Endpoint{
		ID:                 gqlEndpoint.Id,
		EndpointID:         gqlEndpoint.EndpointId,
		DeviceType:         gqlEndpoint.DmsIdentifier.DeviceType,
		DeviceSerialNumber: gqlEndpoint.DmsIdentifier.Dsn,
	}

	if gqlEndpoint.FriendlyName != "" {
		unified.FriendlyName = &alexaapimodels.NameValue{
			Type:  alexamodels.DefaultFriendlyNameType,
			Value: gqlEndpoint.FriendlyName,
		}
	}

	unified.Manufacturer = convertNameValueObject(gqlEndpoint.Manufacturer.Type, gqlEndpoint.Manufacturer.Value.Text)
	unified.Model = convertNameValueObject(gqlEndpoint.Model.Type, gqlEndpoint.Model.Value.Text)
	unified.SerialNumber = convertNameValueObject(gqlEndpoint.SerialNumber.Type, gqlEndpoint.SerialNumber.Value.Text)
	unified.Description = convertNameValueObject(gqlEndpoint.Description.Type, gqlEndpoint.Description.Value.Text)
	unified.SoftwareVersion = convertNameValueObject(
		gqlEndpoint.SoftwareVersion.Type,
		gqlEndpoint.SoftwareVersion.Value.Text,
	)
	unified.DisplayCategories = convertListEndpointDisplayCategories(gqlEndpoint.DisplayCategories)

	graphqlFeatures := make([]string, 0, len(gqlEndpoint.Features))

	graphqlFeaturesMap := make(map[string]alexaapimodels.Feature)

	for _, gqlFeature := range gqlEndpoint.Features {
		graphqlFeatures = append(graphqlFeatures, gqlFeature.Name)

		feature := alexaapimodels.Feature{
			Name:     alexaapimodels.FeatureName(gqlFeature.Name),
			Instance: gqlFeature.Instance,
		}
		for _, op := range gqlFeature.Operations {
			feature.Operations = append(feature.Operations, alexaapimodels.FeatureOperation{
				Name: op.Name,
			})
		}

		graphqlFeaturesMap[featureIdentity(feature)] = feature
	}

	supportedFeatures := c.determineSupportedFeatures(graphqlFeatures, []string{}, unified.DeviceFamily)
	unified.Features = mergeSupportedAndGraphQLFeatures(supportedFeatures, graphqlFeaturesMap)

	return unified
}

func mergeSupportedAndGraphQLFeatures(
	supportedFeatures []alexaapimodels.Feature,
	graphqlFeaturesMap map[string]alexaapimodels.Feature,
) []alexaapimodels.Feature {
	features := make([]alexaapimodels.Feature, 0, len(supportedFeatures)+len(graphqlFeaturesMap))
	merged := make(map[string]struct{}, len(graphqlFeaturesMap))

	for _, supportedFeature := range supportedFeatures {
		var matchingKeys []string

		for key, graphqlFeature := range graphqlFeaturesMap {
			if graphqlFeature.Name == supportedFeature.Name {
				matchingKeys = append(matchingKeys, key)
			}
		}

		sort.Strings(matchingKeys)

		if len(matchingKeys) == 0 {
			features = append(features, supportedFeature)

			continue
		}

		for _, key := range matchingKeys {
			graphqlFeature := graphqlFeaturesMap[key]
			feature := supportedFeature
			feature.Instance = graphqlFeature.Instance
			feature.Config = graphqlFeature.Config
			feature.Operations = graphqlFeature.Operations
			feature.Properties = graphqlFeature.Properties
			features = append(features, feature)
			merged[key] = struct{}{}
		}
	}

	var extraKeys []string

	for key := range graphqlFeaturesMap {
		if _, ok := merged[key]; !ok {
			extraKeys = append(extraKeys, key)
		}
	}

	sort.Strings(extraKeys)

	for _, key := range extraKeys {
		features = append(features, graphqlFeaturesMap[key])
	}

	return features
}

func featureIdentity(feature alexaapimodels.Feature) string {
	return string(feature.Name) + "\x00" + feature.Instance
}

func convertFeatureConfigWithStates(
	input gqlStateFeatureConfiguration,
) *alexaapimodels.FeatureConfig {
	rangeConfig, ok := input.(*gqlStateRangeConfiguration)
	if !ok {
		return nil
	}

	return &alexaapimodels.FeatureConfig{
		Range: &alexaapimodels.RangeConfig{
			FriendlyName:  convertNameValueObject(rangeConfig.FriendlyName.Type, rangeConfig.FriendlyName.Value.Text),
			UnitOfMeasure: convertNameValueObject(rangeConfig.UnitOfMeasure.Type, rangeConfig.UnitOfMeasure.Value.Text),
			SupportedRange: &alexaapimodels.SupportedRange{
				MinimumValue: rangeConfig.SupportedRange.MinimumValue,
				MaximumValue: rangeConfig.SupportedRange.MaximumValue,
				Precision:    rangeConfig.SupportedRange.Precision,
			},
			Presets: convertRangePresetsWithStates(rangeConfig.Presets),
		},
	}
}

func convertRangePresetsWithStates(
	inputs []gqlStateRangePreset,
) []alexaapimodels.RangePreset {
	presets := make([]alexaapimodels.RangePreset, 0, len(inputs))
	for _, input := range inputs {
		presets = append(presets, alexaapimodels.RangePreset{
			RangeValue: input.RangeValue,
			FriendlyName: convertNameValueObject(
				input.PresetResources.FriendlyName.Type,
				input.PresetResources.FriendlyName.Value.Text,
			),
		})
	}

	return presets
}

func convertFeatureConfigWithoutStates(
	input graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeatureConfiguration,
) *alexaapimodels.FeatureConfig {
	rangeConfig, ok := input.(*gqlEndpointFeaturesFeatureConfigurationRangeConfiguration)
	if !ok {
		return nil
	}

	return &alexaapimodels.FeatureConfig{
		Range: &alexaapimodels.RangeConfig{
			FriendlyName:  convertNameValueObject(rangeConfig.FriendlyName.Type, rangeConfig.FriendlyName.Value.Text),
			UnitOfMeasure: convertNameValueObject(rangeConfig.UnitOfMeasure.Type, rangeConfig.UnitOfMeasure.Value.Text),
			SupportedRange: &alexaapimodels.SupportedRange{
				MinimumValue: rangeConfig.SupportedRange.MinimumValue,
				MaximumValue: rangeConfig.SupportedRange.MaximumValue,
				Precision:    rangeConfig.SupportedRange.Precision,
			},
			Presets: convertRangePresetsWithoutStates(rangeConfig.Presets),
		},
	}
}

func convertRangePresetsWithoutStates(
	inputs []gqlEndpointFeaturesFeatureConfigurationRangeConfigurationPresetsPreset,
) []alexaapimodels.RangePreset {
	presets := make([]alexaapimodels.RangePreset, 0, len(inputs))
	for _, input := range inputs {
		presets = append(presets, alexaapimodels.RangePreset{
			RangeValue: input.RangeValue,
			FriendlyName: convertNameValueObject(
				input.PresetResources.FriendlyName.Type,
				input.PresetResources.FriendlyName.Value.Text,
			),
		})
	}

	return presets
}

func convertListEndpointDisplayCategories(
	inputs gqlListEndpointDisplayCategories,
) alexaapimodels.DisplayCategories {
	secondaryInputs := inputs.GetAll()

	secondaryCategories := make([]alexaapimodels.EndpointDisplayCategory, 0, len(secondaryInputs))
	for _, secondary := range secondaryInputs {
		secondaryCategories = append(secondaryCategories, mapCategory(secondary.Value))
	}

	return alexaapimodels.DisplayCategories{
		Primary:   mapCategory(inputs.Primary.Value),
		Secondary: secondaryCategories,
	}
}

func convertDisplayCategories(
	inputs graphql.EndpointsEndpointsEndpointsResponseItemsEndpointDisplayCategories,
) alexaapimodels.DisplayCategories {
	secondaryInputs := inputs.GetAll()

	secondaryCategories := make([]alexaapimodels.EndpointDisplayCategory, 0, len(secondaryInputs))
	for _, secondary := range secondaryInputs {
		secondaryCategories = append(secondaryCategories, mapCategory(secondary.Value))
	}

	return alexaapimodels.DisplayCategories{
		Primary:   mapCategory(inputs.Primary.Value),
		Secondary: secondaryCategories,
	}
}

func mapCategory(input string) alexaapimodels.EndpointDisplayCategory {
	out := alexaapimodels.ValidDisplayCategory(input)
	if !out {
		return alexaapimodels.EndpointDisplayCategoryOther
	}

	return alexaapimodels.EndpointDisplayCategory(input)
}

// extractTimeOfSample reads the schema-generated property accessor.
func (c *Session) extractTimeOfSample(prop gqlStatePropertyFeatureProperty) string {
	if isNilFeatureProperty(prop) {
		return ""
	}

	property, hasSample := prop.(interface{ GetTimeOfSample() string })
	if !hasSample {
		return ""
	}

	return property.GetTimeOfSample()
}

// extractTimeOfLastChange reads the schema-generated property accessor.
func (c *Session) extractTimeOfLastChange(prop gqlStatePropertyFeatureProperty) string {
	if isNilFeatureProperty(prop) {
		return ""
	}

	property, hasChange := prop.(interface{ GetTimeOfLastChange() string })
	if !hasChange {
		return ""
	}

	return property.GetTimeOfLastChange()
}

func isNilFeatureProperty(prop gqlStatePropertyFeatureProperty) bool {
	if prop == nil {
		return true
	}

	value := reflect.ValueOf(prop)

	return value.Kind() == reflect.Pointer && value.IsNil()
}

// extractStateValue extracts the state value from a property using type assertion.
func (c *Session) extractStateValue(
	prop gqlStatePropertyFeatureProperty,
) interface{} {
	if value, ok := extractCoreStateValue(prop); ok {
		return value
	}

	value, _ := extractExtendedStateValue(prop)

	return value
}

func extractCoreStateValue(prop gqlStatePropertyFeatureProperty) (interface{}, bool) {
	switch propertyValue := prop.(type) {
	case *gqlStatePropertyVolume:
		if propertyValue.VolumeValue == nil || propertyValue.VolumeValue.Value == nil {
			return nil, true
		}

		return propertyValue.VolumeValue.Value, true
	case *gqlStatePropertyPower:
		return propertyValue.PowerStateValue, true
	case *gqlStatePropertyBrightness:
		return propertyValue.BrightnessStateValue, true
	case *gqlStatePropertyColor:
		return propertyValue.ColorStateValue, true
	case *gqlStatePropertyColorTemperature:
		return propertyValue.ColorTemperatureInKelvinStateValue, true
	case *gqlStatePropertyReachability:
		return propertyValue.ReachabilityStatusValue, true
	case *gqlStatePropertyPercentage:
		return propertyValue.PercentageValue, true
	case *gqlStatePropertyPowerLevel:
		return propertyValue.PowerLevelValue, true
	case *gqlStatePropertyLock:
		return propertyValue.LockState, true
	case *gqlStatePropertyMode:
		return propertyValue.ModeValue, true
	case *gqlStatePropertyRangeValue:
		rangeState := extractRangeStateValue(propertyValue)
		if rangeState == nil {
			return nil, true
		}

		return rangeState, true
	case *gqlStatePropertyToggleState:
		return propertyValue.ToggleStateValue, true
	case *gqlStatePropertyThermostatMode:
		return propertyValue.ThermostatModeValue, true
	case *gqlStatePropertyTemperatureSensor:
		return propertyValue.Value, true
	default:
		return nil, false
	}
}

func extractExtendedStateValue(prop gqlStatePropertyFeatureProperty) (interface{}, bool) {
	switch propertyValue := prop.(type) {
	case *gqlStatePropertyArmState:
		return propertyValue.ArmStateValue, true
	case *gqlStatePropertyActionState:
		return propertyValue.ActionStateValue, true
	case *gqlStatePropertyBattery:
		return propertyValue.BatteryValue, true
	case *gqlStatePropertyBurglaryAlarm:
		return propertyValue.BurglaryAlarmValue, true
	case *gqlStatePropertyCarbonMonoxideAlarm:
		return propertyValue.CarbonMonoxideAlarmValue, true
	case *gqlStatePropertyFireAlarm:
		return propertyValue.FireAlarmValue, true
	case *gqlStatePropertyWaterAlarm:
		return propertyValue.WaterAlarmValue, true
	case *gqlStatePropertySirenState:
		return propertyValue.SirenStateValue, true
	case *gqlStatePropertyRadioDiagnostics:
		return propertyValue.RadioDiagnosticsValue, true
	case *gqlStatePropertyNetworkThroughput:
		return nil, true
	case *gqlStatePropertySnapshot:
		return propertyValue.SnapshotValue, true
	case *gqlStatePropertyStatusCode:
		return propertyValue.StatusCodeValue, true
	case *gqlStatePropertySetpoint:
		return propertyValue.Value, true
	case *gqlStatePropertyRelativeHumidity:
		return propertyValue.RelativeHumidityValue, true
	case *gqlStatePropertyIlluminance:
		return propertyValue.IlluminanceValue, true
	case *gqlStatePropertyGeolocation:
		return propertyValue.GeolocationValue, true
	default:
		return nil, false
	}
}

func extractRangeStateValue(
	prop *gqlStatePropertyRangeValue,
) *alexaapimodels.RangeValueState {
	if prop == nil || prop.Error.Type != "" {
		return nil
	}

	if prop.RangeValue.Typename == "" {
		return nil
	}

	return &alexaapimodels.RangeValueState{Value: prop.RangeValue.Value}
}

// extractError extracts error information from a property using type assertion.
func (c *Session) extractError(
	prop gqlStatePropertyFeatureProperty,
) *alexaapimodels.PropertyError {
	switch propertyValue := prop.(type) {
	case *gqlStatePropertyVolume:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyPower:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyBrightness:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyReachability:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyPercentage:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyPowerLevel:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyLock:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyMode:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyRangeValue:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyToggleState:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyThermostatMode:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyTemperatureSensor:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyActionState:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyBurglaryAlarm:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyFireAlarm:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyNetworkThroughput:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertySnapshot:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	case *gqlStatePropertyStatusCode:
		return propertyErrorFromValues(propertyValue.Error.Type, propertyValue.Error.Message)
	}

	return nil
}

func propertyErrorFromValues(errorType, message string) *alexaapimodels.PropertyError {
	if errorType == "" {
		return nil
	}

	return &alexaapimodels.PropertyError{Type: errorType, Message: message}
}

// determineSupportedFeatures merges GraphQL features, REST capabilities, and device family
// to determine which features this endpoint supports.
func (c *Session) determineSupportedFeatures(
	graphqlFeatures []string,
	restCapabilities []string,
	deviceFamily string,
) []alexaapimodels.Feature {
	featureSet, supported := makeFeatureSets(graphqlFeatures)
	capabilitySet, capabilitySetLower := makeCapabilitySets(restCapabilities)
	isFireTV := deviceFamily == string(alexamodels.FireTVDeviceFamily)
	isEcho := deviceFamily == string(alexamodels.EchoDeviceFamily) ||
		deviceFamily == string(alexamodels.KnightDeviceFamily)

	// Map legacy capability interface names (and direct feature name strings) to feature names.
	capabilityFeatureMap := map[string]alexaapimodels.FeatureName{
		string(alexaapimodels.FeatureNamePlayback):          alexaapimodels.FeatureNamePlayback,
		string(alexaapimodels.FeatureNameLightSensor):       alexaapimodels.FeatureNameLightSensor,
		string(alexaapimodels.FeatureNameTemperatureSensor): alexaapimodels.FeatureNameTemperatureSensor,
		string(alexaapimodels.FeatureNameHumiditySensor):    alexaapimodels.FeatureNameHumiditySensor,
		string(alexaapimodels.FeatureNameMotionSensor):      alexaapimodels.FeatureNameMotionSensor,
		string(alexaapimodels.FeatureNameContactSensor):     alexaapimodels.FeatureNameContactSensor,
		string(alexaapimodels.FeatureNameEndpointHealth):    alexaapimodels.FeatureNameEndpointHealth,
		string(alexaapimodels.FeatureNameConnectivity):      alexaapimodels.FeatureNameConnectivity,
		string(alexaapimodels.FeatureNameSpeaker):           alexaapimodels.FeatureNameSpeaker,
		string(alexaapimodels.FeatureNamePower):             alexaapimodels.FeatureNamePower,
		string(alexaapimodels.FeatureNameBrightness):        alexaapimodels.FeatureNameBrightness,
		string(alexaapimodels.FeatureNameColor):             alexaapimodels.FeatureNameColor,
		string(alexaapimodels.FeatureNameColorTemperature):  alexaapimodels.FeatureNameColorTemperature,
		string(alexaapimodels.FeatureNameLock):              alexaapimodels.FeatureNameLock,
		string(alexaapimodels.FeatureNameMode):              alexaapimodels.FeatureNameMode,
		string(alexaapimodels.FeatureNameRange):             alexaapimodels.FeatureNameRange,
		string(alexaapimodels.FeatureNameToggle):            alexaapimodels.FeatureNameToggle,
		string(alexaapimodels.FeatureNamePercentage):        alexaapimodels.FeatureNamePercentage,
		string(alexaapimodels.FeatureNamePowerLevel):        alexaapimodels.FeatureNamePowerLevel,
		string(alexaapimodels.FeatureNameThermostat):        alexaapimodels.FeatureNameThermostat,
		string(alexaapimodels.FeatureNameLocation):          alexaapimodels.FeatureNameLocation,
		string(alexaapimodels.FeatureNameLocationTracker):   alexaapimodels.FeatureNameLocationTracker,
		string(alexaapimodels.FeatureNameSecurityPanel):     alexaapimodels.FeatureNameSecurityPanel,
		alexamodels.LegacyPlaybackControllerCapability:      alexaapimodels.FeatureNamePlayback,
		alexamodels.LegacyLightSensorCapability:             alexaapimodels.FeatureNameLightSensor,
		alexamodels.LegacyTemperatureSensorCapability:       alexaapimodels.FeatureNameTemperatureSensor,
		alexamodels.LegacyHumiditySensorCapability:          alexaapimodels.FeatureNameHumiditySensor,
		alexamodels.LegacySpeakerCapability:                 alexaapimodels.FeatureNameSpeaker,
		alexamodels.LegacyLocationCapability:                alexaapimodels.FeatureNameLocation,
		alexamodels.LegacyLocationTrackerCapability:         alexaapimodels.FeatureNameLocationTracker,
		alexamodels.LegacySecurityPanelControllerCapability: alexaapimodels.FeatureNameSecurityPanel,
	}

	addDeviceFamilyFeatures(supported, featureSet, capabilitySet, isFireTV, isEcho)
	addMappedCapabilityFeatures(supported, capabilitySetLower, capabilityFeatureMap)

	return buildSupportedFeatureList(supported)
}

func makeFeatureSets(graphqlFeatures []string) (map[string]bool, map[string]bool) {
	featureSet := make(map[string]bool, len(graphqlFeatures))
	supported := make(map[string]bool, len(graphqlFeatures))

	for _, feature := range graphqlFeatures {
		featureSet[feature] = true
		supported[feature] = true
	}

	return featureSet, supported
}

func makeCapabilitySets(capabilities []string) (map[string]bool, map[string]bool) {
	capabilitySet := make(map[string]bool, len(capabilities))
	capabilitySetLower := make(map[string]bool, len(capabilities))

	for _, capability := range capabilities {
		capabilitySet[capability] = true
		capabilitySetLower[strings.ToLower(capability)] = true
	}

	return capabilitySet, capabilitySetLower
}

func addDeviceFamilyFeatures(
	supported map[string]bool,
	featureSet map[string]bool,
	capabilitySet map[string]bool,
	isFireTV bool,
	isEcho bool,
) {
	addCoreDeviceFeatures(supported, featureSet, capabilitySet, isFireTV, isEcho)
	addFamilyDefaultFeatures(supported, isFireTV, isEcho)
}

func addCoreDeviceFeatures(
	supported map[string]bool,
	featureSet map[string]bool,
	capabilitySet map[string]bool,
	isFireTV bool,
	isEcho bool,
) {
	powerFeature := string(alexaapimodels.FeatureNamePower)
	if featureSet[powerFeature] || isFireTV {
		supported[powerFeature] = true
	}

	speakerFeature := string(alexaapimodels.FeatureNameSpeaker)
	if featureSet[speakerFeature] || isFireTV || isEcho ||
		capabilitySet[alexamodels.VolumeSettingCapability] || capabilitySet[alexamodels.DsVolumeSettingCapability] {
		if !featureSet[speakerFeature] {
			supported[speakerFeature] = true
		}
	}

	if isFireTV || hasAudioPlayerCapability(capabilitySet) || isEcho {
		supported[string(alexaapimodels.FeatureNamePlayback)] = true
	}
}

func addFamilyDefaultFeatures(supported map[string]bool, isFireTV, isEcho bool) {
	if isEcho {
		supported[string(alexaapimodels.FeatureNameNotification)] = true
		supported[string(alexaapimodels.FeatureNameAnnouncement)] = true
		supported[string(alexaapimodels.FeatureNameSpeechSynthesizer)] = true
		supported[string(alexaapimodels.FeatureNameAudioPlayer)] = true
	}

	if isFireTV {
		supported[string(alexaapimodels.FeatureNameNavigation)] = true
	}
}

func hasAudioPlayerCapability(capabilitySet map[string]bool) bool {
	return capabilitySet[alexamodels.AudioPlayerCapability] ||
		capabilitySet[alexamodels.AmazonMusicCapability] ||
		capabilitySet[alexamodels.TuneInCapability] ||
		capabilitySet[alexamodels.AppleMusicCapability] ||
		capabilitySet[alexamodels.PandoraCapability] ||
		capabilitySet[alexamodels.IHeartRadioCapability] ||
		capabilitySet[alexamodels.SiriusxmCapability] ||
		capabilitySet[alexamodels.DeezerCapability] ||
		capabilitySet[alexamodels.TidalCapability]
}

func addMappedCapabilityFeatures(
	supported map[string]bool,
	capabilitySetLower map[string]bool,
	capabilityFeatureMap map[string]alexaapimodels.FeatureName,
) {
	for capabilityName, featureName := range capabilityFeatureMap {
		if capabilitySetLower[capabilityName] {
			supported[string(featureName)] = true
		}
	}
}

func buildSupportedFeatureList(supported map[string]bool) []alexaapimodels.Feature {
	featureNames := make([]string, 0, len(supported))
	for featureName := range supported {
		featureNames = append(featureNames, featureName)
	}

	sort.Strings(featureNames)

	result := make([]alexaapimodels.Feature, 0, len(featureNames))

	for _, name := range featureNames {
		featureName := alexaapimodels.FeatureName(name)
		props, ops := getFeatureDefaults(featureName)
		result = append(result, alexaapimodels.Feature{
			Name:       featureName,
			Properties: props,
			Operations: ops,
		})
	}

	return result
}

// extractLegacyCapabilityInterfaces decodes known members of the legacy JSON scalar.
func extractLegacyCapabilityInterfaces(capabilities []interface{}) []string {
	interfaces := make([]string, 0, len(capabilities))

	for _, capability := range capabilities {
		data, err := json.Marshal(capability)
		if err != nil {
			continue
		}

		var value alexamodels.LegacyCapabilityPayload

		err = json.Unmarshal(data, &value)
		if err != nil || value.InterfaceName == nil || *value.InterfaceName == "" {
			continue
		}

		interfaces = append(interfaces, *value.InterfaceName)
	}

	return interfaces
}

// convertNameValueObject converts a GraphQL NameValueObject to alexaapimodels NameValue.
func convertNameValueObject(gqlType graphql.NameValueObjectType, text string) *alexaapimodels.NameValue {
	if text == "" {
		return nil
	}

	return &alexaapimodels.NameValue{
		Type:  string(gqlType),
		Value: text,
	}
}
