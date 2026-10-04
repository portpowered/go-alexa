package alexamodels_test

import (
	"encoding/json"
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

func TestGeneratedWireStateBindingsPreserveFutureValues(t *testing.T) {
	t.Parallel()

	var control alexamodels.LockPayload

	err := json.Unmarshal([]byte(`{"lockState":"FUTURE_LOCK_STATE"}`), &control)
	if err != nil || control.LockState != "FUTURE_LOCK_STATE" {
		t.Fatalf("future control state = %+v, %v", control, err)
	}

	var event alexamodels.PowerProperty

	err = json.Unmarshal([]byte(`{"powerStateValue":"FUTURE_POWER_STATE","accuracy":"FUTURE_ACCURACY"}`), &event)
	if err != nil || event.PowerStateValue != "FUTURE_POWER_STATE" || event.Accuracy != "FUTURE_ACCURACY" {
		t.Fatalf("future event state = %+v, %v", event, err)
	}

	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("encode future event state: %v", err)
	}

	var roundTrip alexamodels.PowerProperty

	err = json.Unmarshal(encoded, &roundTrip)
	if err != nil || roundTrip.PowerStateValue != event.PowerStateValue || roundTrip.Accuracy != event.Accuracy {
		t.Fatalf("future state round trip = %+v, %v", roundTrip, err)
	}
}
