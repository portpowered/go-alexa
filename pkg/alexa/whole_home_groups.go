package alexa

import (
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// WHA groups can exist only in DevicesV2, so a GraphQL left join cannot discover them.
// The device-account identity is an SDK discovery key, never a GraphQL endpoint ID.
func mergeWholeHomeGroups(endpoints []*alexaapimodels.Endpoint, devices []alexamodels.DeviceV2) []*alexaapimodels.Endpoint {
	for _, device := range devices {
		if device.DeviceFamily != string(alexamodels.WholeHomeDeviceFamily) || device.DeviceType == "" || device.SerialNumber == "" || device.DeviceAccountId == "" {
			continue
		}

		existing := wholeHomeIdentity(endpoints, device)
		if existing != nil {
			existing.ClusterMembers = append([]string(nil), device.ClusterMembers...)

			continue
		}

		id := device.DeviceAccountId

		endpoint := &alexaapimodels.Endpoint{
			ID: id, EndpointID: id,
			DeviceType: device.DeviceType, DeviceSerialNumber: device.SerialNumber,
			DeviceFamily: device.DeviceFamily, DeviceAccountId: device.DeviceAccountId,
			DeviceOwnerCustomerID: device.DeviceOwnerCustomerId,
			ClusterMembers:        append([]string(nil), device.ClusterMembers...),
			FriendlyName:          &alexaapimodels.NameValue{Type: alexamodels.DefaultFriendlyNameType, Value: device.AccountName},
		}
		if device.Language != nil {
			endpoint.Locale = *device.Language
		}

		for _, capability := range device.Capabilities {
			switch capability {
			case alexamodels.AudioPlayerCapability:
				endpoint.Features = append(endpoint.Features, wholeHomeFeature(alexaapimodels.FeatureNameAudioPlayer), wholeHomeFeature(alexaapimodels.FeatureNamePlayback))
			case alexamodels.VolumeSettingCapability:
				endpoint.Features = append(endpoint.Features, wholeHomeFeature(alexaapimodels.FeatureNameSpeaker))
			}
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func wholeHomeIdentity(endpoints []*alexaapimodels.Endpoint, device alexamodels.DeviceV2) *alexaapimodels.Endpoint {
	for _, endpoint := range endpoints {
		if endpoint != nil && endpoint.DeviceType == device.DeviceType && endpoint.DeviceSerialNumber == device.SerialNumber {
			return endpoint
		}
	}

	return nil
}

func wholeHomeFeature(name alexaapimodels.FeatureName) alexaapimodels.Feature {
	properties, operations := getFeatureDefaults(name)

	return alexaapimodels.Feature{Name: name, Properties: properties, Operations: operations}
}
