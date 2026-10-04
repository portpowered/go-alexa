package integration_test

import (
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestPowerEventClassificationRequiresTypedEndpointPayload(t *testing.T) {
	t.Parallel()

	payload := &alexaapimodels.PowerPayload{
		PowerState: "ON", TimeOfSample: time.Time{}, Accuracy: "", Type: "", Error: nil,
	}

	correct := &alexaapimodels.Event{
		Namespace: "Alexa.PowerController", Name: "StateReport", EndpointID: "synthetic-endpoint",
		MessageID: "synthetic-message", Payload: payload,
	}
	if !isIntegrationPowerEvent(t, correct, "synthetic-endpoint") {
		t.Fatal("typed power event was rejected")
	}

	if isIntegrationPowerEvent(t, correct, "another-endpoint") || isIntegrationPowerEvent(t, nil, "synthetic-endpoint") {
		t.Fatal("unrelated or absent event was accepted")
	}

	for _, name := range []string{"ChangeReport", "StateReport"} {
		untyped := &alexaapimodels.Event{
			Namespace: "Alexa", Name: name, EndpointID: "synthetic-endpoint", MessageID: "", Payload: nil,
		}
		if isIntegrationPowerEvent(t, untyped, "synthetic-endpoint") {
			t.Fatal("a generic report name cannot prove a power event")
		}
	}
}
