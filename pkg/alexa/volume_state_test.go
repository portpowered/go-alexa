//nolint:testpackage // Checks private GraphQL projection and optional-value preservation.
package alexa

import (
	"encoding/json"
	"testing"
)

func TestGraphQLVolumeProjectionPreservesZeroAndMissingValues(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, input string
		present     bool
		volume      int
	}{
		{"zero", `{"volumeValue":{"value":0}}`, true, 0},
		{"nonzero", `{"volumeValue":{"value":43}}`, true, 43},
		{"missing", `{}`, false, 0},
		{"null", `{"volumeValue":null}`, false, 0},
		{"missing inner value", `{"volumeValue":{}}`, false, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var property gqlStatePropertyVolume

			err := json.Unmarshal([]byte(test.input), &property)
			if err != nil {
				t.Fatal(err)
			}

			value, supported := extractCoreStateValue(&property)

			if !supported {
				t.Fatal("volume property was not recognized")
			}

			volume, ok := value.(*int)

			if test.present && (!ok || volume == nil || *volume != test.volume) {
				t.Fatalf("wrong volume: %v", value)
			}

			if !test.present && value != nil {
				t.Fatalf("invented volume: %v", value)
			}
		})
	}
}
