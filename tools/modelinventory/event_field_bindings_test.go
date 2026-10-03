package main

import (
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

func TestEventFieldBindingsRejectLostEnumAndAnonymousObject(t *testing.T) {
	t.Parallel()

	graphql, err := readGraphQLSchema()
	if err != nil {
		t.Fatalf("read SDL: %v", err)
	}

	tests := []struct {
		component string
		field     string
		inline    string
	}{
		{component: "LockProperty", field: "lockState", inline: ""},
		{component: "ColorProperty", field: "colorStateValue", inline: "EventColorValue"},
		{component: "EventGeolocationValue", field: "source", inline: ""},
		{component: "RangeProperty", field: "type", inline: ""},
		{component: "ToggleProperty", field: "toggleStateValue", inline: ""},
	}

	for _, test := range tests {
		t.Run(test.component+"."+test.field, func(t *testing.T) {
			t.Parallel()

			data, readErr := os.ReadFile("../../api/feature-events.yaml")
			if readErr != nil {
				t.Fatalf("read event schema: %v", readErr)
			}

			var document schemaDocument

			decodeErr := yaml.Unmarshal(data, &document)
			if decodeErr != nil {
				t.Fatalf("decode event schema: %v", decodeErr)
			}

			components := document.Components.Schemas
			property := emptySchemaNode()

			property.Type = schemaStringType

			if test.inline != "" {
				property = components[test.inline]
			}

			components[test.component].Properties[test.field] = property

			checkErr := checkEventFieldBindings("api/feature-events.yaml", components, graphql)
			if checkErr == nil {
				t.Fatal("lost schema binding accepted")
			}
		})
	}
}
