package alexamodels_test

import (
	"reflect"
	"testing"

	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestLegacyOAuthConfigurationHasNoWireTags(t *testing.T) {
	t.Parallel()

	for _, model := range []reflect.Type{
		reflect.TypeFor[alexamodels.AuthConfig](),
		reflect.TypeFor[alexamodels.DeviceAuthConfig](),
	} {
		for field := range model.NumField() {
			if _, present := model.Field(field).Tag.Lookup("json"); present {
				t.Fatalf("legacy caller configuration %s acquired a wire field", model.Name())
			}
		}
	}
}
