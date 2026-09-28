package alexa

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// ListEndpoints retrieves all endpoints by joining GraphQL endpoint data
// with DeviceV2 data using a left outer join on DMSIdentifier.
// The query parameter can specify filters and which optional fields to include
// (states, capabilities, features). When States is true in IncludeFields,
// the function will retrieve endpoints with detailed property states.
func (c *Session) ListEndpoints(ctx context.Context, q alexaapimodels.EndpointQuery) (*alexaapimodels.UnifiedEndpointListResponse, error) {

	// Determine what optional fields to include
	includeStates := false
	includeCapabilities := false
	includeFeatures := false
	if q.IncludeFields != nil {
		includeStates = q.IncludeFields.Properties
		includeCapabilities = q.IncludeFields.Features
	}
	// Always include capabilities if not explicitly disabled, as we need them for operation detection
	// We'll fetch capabilities to determine supported operations
	if !includeCapabilities {
		includeCapabilities = true
	}

	// Build GraphQL input from query parameters
	graphqlInput := graphql.ListEndpointsInput{
		PaginationParams: graphql.PaginationParams{
			PageSize:          50,
			DisablePagination: true,
		},
	}

	var graphqlEndpointsWithoutStates []graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint
	var graphqlEndpointsWithStates []graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint
	var err error

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
				PageSize:          50,
				DisablePagination: true,
			},
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

func mergeDistinctEndpoints(primary []*alexaapimodels.Endpoint, extra []*alexaapimodels.Endpoint) []*alexaapimodels.Endpoint {
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

// endpointMergeData holds the data structures needed for merging endpoints
type endpointMergeData struct {
	deviceV2Map         map[string]*alexamodels.DeviceV2
	includeFeatures     bool
	includeCapabilities bool
}

// mergeEndpoints performs a left outer join of GraphQL endpoints with DeviceV2 and REST data
// It handles both endpoints with and without states based on which list is provided
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

// mergeEndpointWithStates merges a single GraphQL endpoint (with states) with DeviceV2 and REST data
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
			Type:  "PLAIN", // Default type for string field
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
		if deviceV2.Language != nil {
			unified.Locale = *deviceV2.Language
		}
	}

	// Extract supported features from feature names
	var graphqlFeatures []string
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

// mergeEndpointWithoutStates merges a single GraphQL endpoint (without states) with DeviceV2 and REST data
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
			Type:  "PLAIN", // Default type for string field
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
		if deviceV2.Language != nil {
			unified.Locale = *deviceV2.Language
		}
	}

	// Extract supported features from feature names
	var graphqlFeatures []string
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
			Type:  "PLAIN",
			Value: gqlEndpoint.FriendlyName,
		}
	}
	unified.Manufacturer = convertNameValueObject(gqlEndpoint.Manufacturer.Type, gqlEndpoint.Manufacturer.Value.Text)
	unified.Model = convertNameValueObject(gqlEndpoint.Model.Type, gqlEndpoint.Model.Value.Text)
	unified.SerialNumber = convertNameValueObject(gqlEndpoint.SerialNumber.Type, gqlEndpoint.SerialNumber.Value.Text)
	unified.Description = convertNameValueObject(gqlEndpoint.Description.Type, gqlEndpoint.Description.Value.Text)
	unified.SoftwareVersion = convertNameValueObject(gqlEndpoint.SoftwareVersion.Type, gqlEndpoint.SoftwareVersion.Value.Text)
	unified.DisplayCategories = convertListEndpointDisplayCategories(gqlEndpoint.DisplayCategories)

	var graphqlFeatures []string
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
	input graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeatureConfiguration,
) *alexaapimodels.FeatureConfig {
	rangeConfig, ok := input.(*graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeatureConfigurationRangeConfiguration)
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
	inputs []graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeatureConfigurationRangeConfigurationPresetsPreset,
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
	rangeConfig, ok := input.(*graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeatureConfigurationRangeConfiguration)
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
	inputs []graphql.EndpointsEndpointsEndpointsResponseItemsEndpointFeaturesFeatureConfigurationRangeConfigurationPresetsPreset,
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
	inputs graphql.ListEndpointsListEndpointsListEndpointsResponseEndpointsEndpointDisplayCategories,
) alexaapimodels.DisplayCategories {
	var secondaryCategories []alexaapimodels.EndpointDisplayCategory
	for _, secondary := range inputs.GetAll() {
		secondaryCategories = append(secondaryCategories, mapCategory(secondary.Value))
	}
	return alexaapimodels.DisplayCategories{
		Primary:   mapCategory(inputs.Primary.Value),
		Secondary: secondaryCategories,
	}
}

func convertDisplayCategories(inputs graphql.EndpointsEndpointsEndpointsResponseItemsEndpointDisplayCategories) alexaapimodels.DisplayCategories {
	var secondaryCategories []alexaapimodels.EndpointDisplayCategory
	for _, secondary := range inputs.GetAll() {
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
		// Ignore transformation
	}
	return alexaapimodels.EndpointDisplayCategory(input)
}

// extractTimeOfSample extracts timeOfSample from a property using map conversion
func (c *Session) extractTimeOfSample(prop graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFeatureProperty) string {
	data, err := json.Marshal(prop)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if val, ok := m["timeOfSample"].(string); ok {
		return val
	}
	return ""
}

// extractTimeOfLastChange extracts timeOfLastChange from a property using map conversion
func (c *Session) extractTimeOfLastChange(prop graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFeatureProperty) string {
	data, err := json.Marshal(prop)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if val, ok := m["timeOfLastChange"].(string); ok {
		return val
	}
	return ""
}

// extractStateValue extracts the state value from a property using type assertion
func (c *Session) extractStateValue(prop graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFeatureProperty) interface{} {
	// Extract state values based on property type
	switch p := prop.(type) {
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesVolume:
		// Volume doesn't have a value field in the query, return nil or the struct itself
		return nil
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPower:
		return p.PowerStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBrightness:
		return p.BrightnessStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesColor:
		return p.ColorStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesColorTemperature:
		return p.ColorTemperatureInKelvinStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesReachability:
		return p.ReachabilityStatusValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPercentage:
		return p.PercentageValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPowerLevel:
		return p.PowerLevelValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesLock:
		return p.LockState
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesMode:
		return p.ModeValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRangeValue:
		rangeState := extractRangeStateValue(p)
		if rangeState == nil {
			return nil
		}
		return rangeState
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesToggleState:
		return p.ToggleStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesThermostatMode:
		return p.ThermostatModeValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesTemperatureSensor:
		return p.Value
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesArmState:
		return p.ArmStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesActionState:
		return p.ActionStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBattery:
		return p.BatteryValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBurglaryAlarm:
		return p.BurglaryAlarmValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesCarbonMonoxideAlarm:
		return p.CarbonMonoxideAlarmValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFireAlarm:
		return p.FireAlarmValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesWaterAlarm:
		return p.WaterAlarmValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSirenState:
		return p.SirenStateValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRadioDiagnostics:
		return p.RadioDiagnosticsValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesNetworkThroughput:
		// NetworkThroughput doesn't have a value field in the query, return nil
		return nil
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSnapshot:
		return p.SnapshotValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesStatusCode:
		return p.StatusCodeValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSetpoint:
		return p.Value
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRelativeHumidity:
		return p.RelativeHumidityValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesIlluminance:
		return p.IlluminanceValue
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesGeolocation:
		return p.GeolocationValue
	}
	return nil
}

func extractRangeStateValue(
	prop *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRangeValue,
) *alexaapimodels.RangeValueState {
	if prop == nil || prop.Error.Type != "" {
		return nil
	}

	data, err := json.Marshal(prop)
	if err != nil {
		return nil
	}
	var raw struct {
		RangeValue *struct {
			Value *float64 `json:"value"`
		} `json:"rangeValue"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	if raw.RangeValue == nil || raw.RangeValue.Value == nil {
		return nil
	}
	return &alexaapimodels.RangeValueState{Value: *raw.RangeValue.Value}
}

// extractError extracts error information from a property using type assertion
func (c *Session) extractError(prop graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFeatureProperty) *alexaapimodels.PropertyError {
	switch p := prop.(type) {
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesVolume:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPower:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBrightness:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesReachability:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPercentage:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesPowerLevel:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesLock:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesMode:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesRangeValue:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesToggleState:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesThermostatMode:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesTemperatureSensor:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesActionState:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesBurglaryAlarm:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesFireAlarm:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesNetworkThroughput:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesSnapshot:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	case *graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpointFeaturesFeaturePropertiesStatusCode:
		if p.Error.Type != "" {
			return &alexaapimodels.PropertyError{
				Type:    p.Error.Type,
				Message: p.Error.Message,
			}
		}
	}
	return nil
}

// determineSupportedFeatures merges GraphQL features, REST capabilities, and device family
// to determine which features this endpoint supports
func (c *Session) determineSupportedFeatures(graphqlFeatures []string, restCapabilities []string, deviceFamily string) []alexaapimodels.Feature {
	supported := make(map[string]bool)

	// Create sets for fast lookup
	featureSet := make(map[string]bool)
	for _, feat := range graphqlFeatures {
		featureSet[feat] = true
		// Add all GraphQL features to supported set
		supported[feat] = true
	}

	capabilitySet := make(map[string]bool)
	capabilitySetLower := make(map[string]bool)
	for _, cap := range restCapabilities {
		capabilitySet[cap] = true
		capabilitySetLower[strings.ToLower(cap)] = true
	}

	isFireTV := deviceFamily == string(alexamodels.FireTVDeviceFamily)
	isEcho := deviceFamily == string(alexamodels.EchoDeviceFamily) || deviceFamily == string(alexamodels.KnightDeviceFamily)

	// Map legacy capability interface names (and direct feature name strings) to feature names
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

		"alexa.playbackcontroller":      alexaapimodels.FeatureNamePlayback,
		"alexa.lightsensor":             alexaapimodels.FeatureNameLightSensor,
		"alexa.temperaturesensor":       alexaapimodels.FeatureNameTemperatureSensor,
		"alexa.humiditysensor":          alexaapimodels.FeatureNameHumiditySensor,
		"alexa.speaker":                 alexaapimodels.FeatureNameSpeaker,
		"alexa.location":                alexaapimodels.FeatureNameLocation,
		"alexa.location.tracker":        alexaapimodels.FeatureNameLocationTracker,
		"alexa.securitypanelcontroller": alexaapimodels.FeatureNameSecurityPanel,
	}

	// Power feature: supported if endpoint has power feature OR is FireTV
	if featureSet[string(alexaapimodels.FeatureNamePower)] || isFireTV {
		supported[string(alexaapimodels.FeatureNamePower)] = true
	}

	// Speaker/Volume feature: supported if endpoint has speaker/volume feature OR
	// is FireTV/Echo device family OR has VOLUME_SETTING capability
	if featureSet[string(alexaapimodels.FeatureNameSpeaker)] ||
		isFireTV || isEcho ||
		capabilitySet["VOLUME_SETTING"] || capabilitySet["DS_VOLUME_SETTING"] {
		if !featureSet[string(alexaapimodels.FeatureNameSpeaker)] {
			supported[string(alexaapimodels.FeatureNameSpeaker)] = true
		}
	}

	// Playback feature: supported if endpoint is FireTV OR has audio player capabilities
	// (AUDIO_PLAYER, AMAZON_MUSIC, TUNE_IN, etc.)
	hasAudioPlayer := capabilitySet["AUDIO_PLAYER"] ||
		capabilitySet["AMAZON_MUSIC"] ||
		capabilitySet["TUNE_IN"] ||
		capabilitySet["APPLE_MUSIC"] ||
		capabilitySet["PANDORA"] ||
		capabilitySet["I_HEART_RADIO"] ||
		capabilitySet["SIRIUSXM"] ||
		capabilitySet["DEEZER"] ||
		capabilitySet["TIDAL"]

	if isFireTV || hasAudioPlayer || isEcho {
		supported[string(alexaapimodels.FeatureNamePlayback)] = true
	}

	for capabilityName, featureName := range capabilityFeatureMap {
		if capabilitySetLower[capabilityName] {
			supported[string(featureName)] = true
		}
	}

	// All GraphQL features are already added above, so feature-specific features
	// (brightness, color, lock, mode, range, toggle, percentage, powerLevel, action)
	// are automatically included

	// Notification, announcement, TTS, and music features are generally available
	if isEcho {
		supported[string(alexaapimodels.FeatureNameNotification)] = true
		supported[string(alexaapimodels.FeatureNameAnnouncement)] = true
		supported[string(alexaapimodels.FeatureNameSpeechSynthesizer)] = true
		supported[string(alexaapimodels.FeatureNameAudioPlayer)] = true
	}

	// FireTV-specific navigation feature
	if isFireTV {
		supported[string(alexaapimodels.FeatureNameNavigation)] = true
	}

	// Convert map to sorted slice of Feature structs
	featureNames := make([]string, 0, len(supported))
	for feat := range supported {
		featureNames = append(featureNames, feat)
	}

	// Sort for consistent output
	sort.Strings(featureNames)

	// Convert to Feature structs with default properties and operations
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

// extractLegacyCapabilityInterfaces pulls interfaceName fields out of legacy capabilities for feature inference.
func extractLegacyCapabilityInterfaces(capabilities []interface{}) []string {
	var interfaces []string
	for _, cap := range capabilities {
		if capMap, ok := cap.(map[string]interface{}); ok {
			if iface, ok := capMap["interfaceName"].(string); ok && iface != "" {
				interfaces = append(interfaces, iface)
			}
		}
	}
	return interfaces
}

// convertNameValueObject converts a GraphQL NameValueObject to alexaapimodels NameValue
func convertNameValueObject(gqlType graphql.NameValueObjectType, text string) *alexaapimodels.NameValue {
	if text == "" {
		return nil
	}
	return &alexaapimodels.NameValue{
		Type:  string(gqlType),
		Value: text,
	}
}
