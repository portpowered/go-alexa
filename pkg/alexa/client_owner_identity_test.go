//nolint:testpackage // Exercises both provider-data join paths before facade dispatch.
package alexa

import (
	"encoding/json"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"

	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestEndpointJoinPreservesDeviceOwnerWithAndWithoutStates(t *testing.T) {
	t.Parallel()

	const endpointJSON = `{"id":"synthetic-id","endpointId":"synthetic-endpoint","dmsIdentifier":{"deviceType":"FIRE_TV_DEVICE","dsn":"synthetic-fire-tv"}}`

	for _, withStates := range []bool{false, true} {
		name := "without-states"
		if withStates {
			name = "with-states"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data := endpointMergeData{includeFeatures: false, includeCapabilities: false, deviceV2Map: map[string]*alexamodels.DeviceV2{
				"FIRE_TV_DEVICE:synthetic-fire-tv": {
					DeviceFamily:          alexamodels.DeviceFamilyFireTV,
					DeviceAccountId:       "synthetic-device-account",
					DeviceOwnerCustomerId: "synthetic-device-owner",
				},
			}}
			session := &Session{}

			var endpoint *alexaapimodels.Endpoint

			if withStates {
				var source graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint

				err := json.Unmarshal([]byte(endpointJSON), &source)
				if err != nil {
					t.Fatal(err)
				}

				endpoint = session.mergeEndpointWithStates(source, data)
			} else {
				var source graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint

				err := json.Unmarshal([]byte(endpointJSON), &source)
				if err != nil {
					t.Fatal(err)
				}

				endpoint = session.mergeEndpointWithoutStates(source, data)
			}

			if endpoint.GetDeviceOwnerCustomerID() != "synthetic-device-owner" || endpoint.DeviceAccountId != "synthetic-device-account" {
				t.Fatalf("joined identity = %#v", endpoint)
			}
		})
	}
}
