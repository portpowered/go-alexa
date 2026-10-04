package main

import "strings"

func checkControlFieldBindings(path string, components map[string]schemaNode) error {
	if path != "api/feature-controls.yaml" {
		return nil
	}

	bindings := map[string]map[string]string{
		"LockPayload":           {"lockState": "LockStateValue"},
		"TogglePayload":         {"toggleState": "ToggleStateValue"},
		"ThermostatSetpoint":    {"scale": "TemperatureScale"},
		"ThermostatModePayload": {"thermostatMode": "ThermostatModeValue"},
	}

	for component, fields := range bindings {
		for name, enum := range fields {
			property := components[component].Properties[name]

			bound := components[strings.TrimPrefix(property.Ref, "#/components/schemas/")]
			if bound.GraphQLEnum != enum {
				return diagnosticf("control wire field %s.%s lacks its SDL enum binding %s", component, name, enum)
			}
		}
	}

	return nil
}
