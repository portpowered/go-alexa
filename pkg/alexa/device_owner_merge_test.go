//nolint:testpackage // Verifies both private GraphQL enumeration projections.
package alexa

import (
	"encoding/json"
	"testing"

	"github.com/portpowered/go-alexa/pkg/dependencies/graphql"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestDeviceOwnerSurvivesDiscoveryAndStateProjection(t *testing.T) {
	t.Parallel()

	var device alexamodels.DeviceV2

	const deviceJSON = `{
		"deviceType":"tv", "serialNumber":"serial",
		"deviceOwnerCustomerId":"synthetic-owner", "deviceFamily":"FIRE_TV",
		"deviceAccountId":"synthetic-account", "language":"de-DE"
	}`

	err := json.Unmarshal([]byte(deviceJSON), &device)
	if err != nil {
		t.Fatal(err)
	}

	merge := endpointMergeData{deviceV2Map: map[string]*alexamodels.DeviceV2{"tv:serial": &device}, includeFeatures: true, includeCapabilities: true}
	session := &Session{}
	wire := []byte(`{"id":"synthetic-endpoint","endpointId":"synthetic-endpoint","dmsIdentifier":{"deviceType":"tv","dsn":"serial"}}`)

	var discovery graphql.EndpointsEndpointsEndpointsResponseItemsEndpoint

	err = json.Unmarshal(wire, &discovery)
	if err != nil {
		t.Fatal(err)
	}

	var state graphql.ListEndpointsWithStatesListEndpointsListEndpointsResponseEndpointsEndpoint

	err = json.Unmarshal(wire, &state)
	if err != nil {
		t.Fatal(err)
	}

	first := session.mergeEndpointWithoutStates(discovery, merge)
	second := session.mergeEndpointWithStates(state, merge)

	if first.DeviceOwnerCustomerID != "synthetic-owner" || second.DeviceOwnerCustomerID != "synthetic-owner" {
		t.Fatal("device owner was lost during GraphQL/DeviceV2 merge")
	}
}
