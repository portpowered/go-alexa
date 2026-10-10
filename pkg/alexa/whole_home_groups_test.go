//nolint:testpackage // Tests private discovery merging of legacy-only groups.
package alexa

import (
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"testing"
)

func TestWholeHomeGroupsSurviveGraphQLLeftJoin(t *testing.T) {
	t.Parallel()

	var device alexamodels.DeviceV2
	device.DeviceFamily = string(alexamodels.WholeHomeDeviceFamily)
	device.DeviceType, device.DeviceAccountId, device.SerialNumber = "synthetic-group-type", "synthetic-group-account", "synthetic-group"
	device.AccountName = "Synthetic audio group"
	device.ClusterMembers = []string{"synthetic-speaker"}
	device.Capabilities = []string{alexamodels.AudioPlayerCapability, alexamodels.VolumeSettingCapability}
	var other alexamodels.DeviceV2
	other.DeviceFamily, other.DeviceType, other.SerialNumber = "ECHO", "echo", "missing-echo"

	endpoints := mergeWholeHomeGroups(nil, []alexamodels.DeviceV2{device, other})
	if len(endpoints) != 1 || endpoints[0].DeviceFamily != "WHA" || len(endpoints[0].ClusterMembers) != 1 {
		t.Fatalf("legacy group not preserved: %+v", endpoints)
	}

	if len(endpoints[0].Features) != 3 {
		t.Fatalf("expected playback, search, speaker: %+v", endpoints[0].Features)
	}

	endpoints = mergeWholeHomeGroups(endpoints, []alexamodels.DeviceV2{device})
	if len(endpoints) != 1 {
		t.Fatal("duplicate group")
	}

	endpoints[0].ClusterMembers[0] = "changed"

	if device.ClusterMembers[0] != "synthetic-speaker" {
		t.Fatal("group members alias vendor response")
	}
}
