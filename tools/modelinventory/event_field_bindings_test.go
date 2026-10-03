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

func TestControlFieldBindingsRejectDroppedEnumReferences(t *testing.T) {
	t.Parallel()

	for component, field := range map[string]string{
		"LockPayload": "lockState", "TogglePayload": "toggleState",
		"ThermostatSetpoint": "scale", "ThermostatModePayload": "thermostatMode",
	} {
		t.Run(component+"."+field, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile("../../api/feature-controls.yaml")
			if err != nil {
				t.Fatalf("read controls: %v", err)
			}

			var document schemaDocument

			err = yaml.Unmarshal(data, &document)
			if err != nil {
				t.Fatalf("decode controls: %v", err)
			}

			err = checkControlFieldBindings("api/feature-controls.yaml", document.Components.Schemas)
			if err != nil {
				t.Fatalf("valid binding rejected: %v", err)
			}

			property := emptySchemaNode()
			property.Type = schemaStringType
			document.Components.Schemas[component].Properties[field] = property

			err = checkControlFieldBindings("api/feature-controls.yaml", document.Components.Schemas)
			if err == nil {
				t.Fatal("dropped control enum binding accepted")
			}
		})
	}
}
