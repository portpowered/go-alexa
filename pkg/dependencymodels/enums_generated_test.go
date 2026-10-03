package alexamodels_test

import (
	"testing"

	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestGeneratedDeviceFamilyEnumPreservesLegacyValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		family alexamodels.DeviceFamily
		want   string
	}{
		{family: alexamodels.WholeHomeDeviceFamily, want: "WHA"},
		{family: alexamodels.FireTVDeviceFamily, want: "FIRE_TV"},
		{family: alexamodels.EchoDeviceFamily, want: "ECHO"},
		{family: alexamodels.KnightDeviceFamily, want: "KNIGHT"},
	}

	for _, test := range tests {
		if string(test.family) != test.want {
			t.Errorf("device family = %q, want %q", test.family, test.want)
		}

		if !test.family.Valid() {
			t.Errorf("generated enum rejects %q", test.family)
		}
	}
}

func TestGeneratedVideoProviderEnumPreservesLegacyValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider alexamodels.VideoProviderID
		want     string
	}{
		{provider: alexamodels.VideoProviderIDAmazon, want: "PRIME VIDEO"},
		{provider: alexamodels.VideoProviderIDYouTube, want: "YOUTUBE"},
		{provider: alexamodels.VideoProviderIDNetflix, want: "NETFLIX"},
		{provider: alexamodels.VideoProviderIDDailymotion, want: "DAILYMOTION"},
	}

	for _, test := range tests {
		if string(test.provider) != test.want {
			t.Errorf("video provider = %q, want %q", test.provider, test.want)
		}

		if !test.provider.Valid() {
			t.Errorf("generated enum rejects %q", test.provider)
		}
	}
}

func TestCapabilityInterfaceRemainsOpenStringAlias(t *testing.T) {
	t.Parallel()

	capability := alexamodels.TemperatureSensorCapability
	if capability != "TEMPERATURE_SENSOR" {
		t.Errorf("capability = %q, want TEMPERATURE_SENSOR", capability)
	}

	assertString := func(string) {}
	assertString(capability)

	var unknown = alexamodels.CapabilityInterface("SYNTHETIC_UNKNOWN_CAPABILITY")
	if unknown != "SYNTHETIC_UNKNOWN_CAPABILITY" {
		t.Errorf("unknown capability = %q", unknown)
	}
}
