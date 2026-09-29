//nolint:testpackage // Exercises private feature-operation serialization helpers.
package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

const syntheticEndpointID = "synthetic-endpoint"

type featureControlTestCase struct {
	name      string
	feature   string
	operation string
	payload   map[string]any
	call      func(*Client) error
}

func featureControlTestCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	cases := make([]featureControlTestCase, 0, 25)
	cases = append(cases, speakerFeatureCases(base)...)
	cases = append(cases, brightnessFeatureCases(base)...)
	cases = append(cases, colorAndTemperatureFeatureCases(base)...)
	cases = append(cases, lockAndModeFeatureCases(base)...)
	cases = append(cases, rangeAndToggleFeatureCases(base)...)
	cases = append(cases, percentageFeatureCases(base)...)
	cases = append(cases, powerLevelFeatureCases(base)...)
	cases = append(cases, actionAndThermostatFeatureCases(base)...)

	return cases
}

func speakerFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "speaker set volume",
			feature:   string(FeatureNameSpeaker),
			operation: string(FeatureOperationNameSetvolume),
			payload:   map[string]any{"volume": 32},
			call: func(client *Client) error {
				_, err := client.ControlSpeakerFeature(
					context.Background(),
					&alexamodels.SpeakerControlRequest{
						BaseFeatureRequest: base,
						SetVolume:          true,
						Volume:             featureTestPointer(32),
					},
				)

				return err
			},
		},
		{
			name:      "speaker adjust volume",
			feature:   string(FeatureNameSpeaker),
			operation: string(FeatureOperationNameAdjustvolume),
			payload:   map[string]any{"volumeDelta": -4},
			call: func(client *Client) error {
				_, err := client.ControlSpeakerFeature(
					context.Background(),
					&alexamodels.SpeakerControlRequest{BaseFeatureRequest: base, Delta: featureTestPointer(-4)},
				)

				return err
			},
		},
	}
}

func brightnessFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "brightness set",
			feature:   string(FeatureNameBrightness),
			operation: string(FeatureOperationNameSetbrightness),
			payload:   map[string]any{"brightness": 75},
			call: func(client *Client) error {
				_, err := client.ControlBrightnessFeature(
					context.Background(),
					&alexamodels.BrightnessControlRequest{
						BaseFeatureRequest: base,
						SetBrightness:      true,
						Brightness:         featureTestPointer(75),
					},
				)

				return err
			},
		},
		{
			name:      "brightness adjust",
			feature:   string(FeatureNameBrightness),
			operation: string(FeatureOperationNameAdjustbrightness),
			payload:   map[string]any{"brightnessDelta": -15},
			call: func(client *Client) error {
				_, err := client.ControlBrightnessFeature(
					context.Background(),
					&alexamodels.BrightnessControlRequest{BaseFeatureRequest: base, Delta: featureTestPointer(-15)},
				)

				return err
			},
		},
	}
}

func colorAndTemperatureFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "color",
			feature:   string(FeatureNameColor),
			operation: string(FeatureOperationNameSetcolor),
			payload:   map[string]any{"color": map[string]any{"hue": 190.0, "saturation": 0.6, "brightness": 0.8}},
			call: func(client *Client) error {
				_, err := client.ControlColorFeature(
					context.Background(),
					&alexamodels.ColorControlRequest{
						BaseFeatureRequest: base,
						Hue:                190,
						Saturation:         0.6,
						Brightness:         0.8,
					},
				)

				return err
			},
		},
		{
			name:      "color temperature set",
			feature:   string(FeatureNameColortemperature),
			operation: string(FeatureOperationNameSetcolortemperature),
			payload:   map[string]any{"colorTemperatureInKelvin": 3200},
			call: func(client *Client) error {
				_, err := client.ControlColorTemperatureFeature(
					context.Background(),
					&alexamodels.ColorTemperatureControlRequest{
						BaseFeatureRequest: base,
						SetTemperature:     true,
						ColorTemperature:   featureTestPointer(3200),
					},
				)

				return err
			},
		},
		{
			name:      "color temperature increase",
			feature:   string(FeatureNameColortemperature),
			operation: string(FeatureOperationNameIncreasecolortemperature),
			payload:   map[string]any{},
			call: func(client *Client) error {
				_, err := client.ControlColorTemperatureFeature(
					context.Background(),
					&alexamodels.ColorTemperatureControlRequest{BaseFeatureRequest: base, Increase: true},
				)

				return err
			},
		},
		{
			name:      "color temperature decrease",
			feature:   string(FeatureNameColortemperature),
			operation: string(FeatureOperationNameDecreasecolortemperature),
			payload:   map[string]any{},
			call: func(client *Client) error {
				_, err := client.ControlColorTemperatureFeature(
					context.Background(),
					&alexamodels.ColorTemperatureControlRequest{BaseFeatureRequest: base},
				)

				return err
			},
		},
	}
}

func lockAndModeFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "locked",
			feature:   string(FeatureNameLock),
			operation: string(FeatureOperationNameLock),
			payload:   map[string]any{"lockState": "LOCKED"},
			call: func(client *Client) error {
				_, err := client.ControlLockFeature(
					context.Background(),
					&alexamodels.LockControlRequest{BaseFeatureRequest: base, State: "LOCKED"},
				)

				return err
			},
		},
		{
			name:      "unlocked",
			feature:   string(FeatureNameLock),
			operation: string(FeatureOperationNameUnlock),
			payload:   map[string]any{"lockState": "UNLOCKED"},
			call: func(client *Client) error {
				_, err := client.ControlLockFeature(
					context.Background(),
					&alexamodels.LockControlRequest{BaseFeatureRequest: base, State: "UNLOCKED"},
				)

				return err
			},
		},
		{
			name:      "mode set",
			feature:   string(FeatureNameMode),
			operation: string(FeatureOperationNameSetmode),
			payload:   map[string]any{"mode": "ECO"},
			call: func(client *Client) error {
				_, err := client.ControlModeFeature(
					context.Background(),
					&alexamodels.ModeControlRequest{
						BaseFeatureRequest: base,
						SetMode:            true,
						Mode:               featureTestPointer("ECO"),
					},
				)

				return err
			},
		},
		{
			name:      "mode adjust",
			feature:   string(FeatureNameMode),
			operation: string(FeatureOperationNameAdjustmode),
			payload:   map[string]any{"modeDelta": "NEXT"},
			call: func(client *Client) error {
				_, err := client.ControlModeFeature(
					context.Background(),
					&alexamodels.ModeControlRequest{BaseFeatureRequest: base, Delta: featureTestPointer("NEXT")},
				)

				return err
			},
		},
	}
}

func rangeAndToggleFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "range set",
			feature:   string(FeatureNameRange),
			operation: string(FeatureOperationNameSetrangevalue),
			payload:   map[string]any{"rangeValue": 12.5},
			call: func(client *Client) error {
				_, err := client.ControlRangeFeature(
					context.Background(),
					&alexamodels.RangeControlRequest{
						BaseFeatureRequest: base,
						SetValue:           true,
						RangeValue:         featureTestPointer(12.5),
					},
				)

				return err
			},
		},
		{
			name:      "range adjust",
			feature:   string(FeatureNameRange),
			operation: string(FeatureOperationNameAdjustrangevalue),
			payload:   map[string]any{"rangeValueDelta": -2.5},
			call: func(client *Client) error {
				_, err := client.ControlRangeFeature(
					context.Background(),
					&alexamodels.RangeControlRequest{BaseFeatureRequest: base, Delta: featureTestPointer(-2.5)},
				)

				return err
			},
		},
		{
			name:      "toggle on",
			feature:   string(FeatureNameToggle),
			operation: string(FeatureOperationNameTurnon),
			payload:   map[string]any{"toggleState": "ON"},
			call: func(client *Client) error {
				_, err := client.ControlToggleFeature(
					context.Background(),
					&alexamodels.ToggleControlRequest{BaseFeatureRequest: base, State: "ON"},
				)

				return err
			},
		},
		{
			name:      "toggle off",
			feature:   string(FeatureNameToggle),
			operation: string(FeatureOperationNameTurnoff),
			payload:   map[string]any{"toggleState": "OFF"},
			call: func(client *Client) error {
				_, err := client.ControlToggleFeature(
					context.Background(),
					&alexamodels.ToggleControlRequest{BaseFeatureRequest: base, State: "OFF"},
				)

				return err
			},
		},
	}
}

func percentageFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "percentage set",
			feature:   string(FeatureNamePercentage),
			operation: string(FeatureOperationNameSetpercentage),
			payload:   map[string]any{"percentage": 25.0},
			call: func(client *Client) error {
				_, err := client.ControlPercentageFeature(
					context.Background(),
					&alexamodels.PercentageControlRequest{
						BaseFeatureRequest: base,
						SetPercentage:      true,
						Percentage:         featureTestPointer(25.0),
					},
				)

				return err
			},
		},
		{
			name:      "percentage adjust",
			feature:   string(FeatureNamePercentage),
			operation: string(FeatureOperationNameAdjustpercentage),
			payload:   map[string]any{"percentageDelta": -5.0},
			call: func(client *Client) error {
				_, err := client.ControlPercentageFeature(
					context.Background(),
					&alexamodels.PercentageControlRequest{BaseFeatureRequest: base, Delta: featureTestPointer(-5.0)},
				)

				return err
			},
		},
	}
}

func powerLevelFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "power level set",
			feature:   string(FeatureNamePowerlevel),
			operation: string(FeatureOperationNameSetpercentage),
			payload:   map[string]any{"powerLevel": 68},
			call: func(client *Client) error {
				_, err := client.ControlPowerLevelFeature(
					context.Background(),
					&alexamodels.PowerLevelControlRequest{
						BaseFeatureRequest: base,
						SetPowerLevel:      true,
						PowerLevel:         featureTestPointer(68),
					},
				)

				return err
			},
		},
		{
			name:      "power level adjust",
			feature:   string(FeatureNamePowerlevel),
			operation: string(FeatureOperationNameAdjustpercentage),
			payload:   map[string]any{"powerLevelDelta": -8},
			call: func(client *Client) error {
				_, err := client.ControlPowerLevelFeature(
					context.Background(),
					&alexamodels.PowerLevelControlRequest{BaseFeatureRequest: base, Delta: featureTestPointer(-8)},
				)

				return err
			},
		},
	}
}

func actionAndThermostatFeatureCases(base alexamodels.BaseFeatureRequest) []featureControlTestCase {
	return []featureControlTestCase{
		{
			name:      "action without parameters",
			feature:   string(FeatureNameAction),
			operation: string(FeatureOperationNamePerformaction),
			payload:   map[string]any{"action": "open"},
			call: func(client *Client) error {
				_, err := client.ControlActionFeature(
					context.Background(),
					&alexamodels.ActionControlRequest{BaseFeatureRequest: base, Action: "open"},
				)

				return err
			},
		},
		{
			name:      "action with parameters",
			feature:   string(FeatureNameAction),
			operation: string(FeatureOperationNamePerformaction),
			payload:   map[string]any{"action": "open", "duration": 12},
			call: func(client *Client) error {
				_, err := client.ControlActionFeature(
					context.Background(),
					&alexamodels.ActionControlRequest{
						BaseFeatureRequest: base,
						Action:             "open",
						Params:             map[string]any{"duration": 12},
					},
				)

				return err
			},
		},
		{
			name:      "thermostat setpoint",
			feature:   string(FeatureNameThermostat),
			operation: string(FeatureOperationNameSettargetsetpoint),
			payload:   map[string]any{"targetSetpoint": map[string]any{"value": 21.5, "scale": "CELSIUS"}},
			call: func(client *Client) error {
				_, err := client.ControlThermostatFeature(
					context.Background(),
					&alexamodels.ThermostatControlRequest{
						BaseFeatureRequest: base,
						SetSetpoint:        true,
						Value:              featureTestPointer(21.5),
						Scale:              "CELSIUS",
					},
				)

				return err
			},
		},
		{
			name:      "thermostat delta",
			feature:   string(FeatureNameThermostat),
			operation: string(FeatureOperationNameAdjusttargetsetpoint),
			payload:   map[string]any{"targetSetpointDelta": map[string]any{"value": -1.5, "scale": "CELSIUS"}},
			call: func(client *Client) error {
				_, err := client.ControlThermostatFeature(
					context.Background(),
					&alexamodels.ThermostatControlRequest{
						BaseFeatureRequest: base,
						Delta:              featureTestPointer(-1.5),
						Scale:              "CELSIUS",
					},
				)

				return err
			},
		},
		{
			name:      "thermostat mode",
			feature:   string(FeatureNameThermostat),
			operation: string(FeatureOperationNameSetthermostatmode),
			payload:   map[string]any{"thermostatMode": "HEAT"},
			call: func(client *Client) error {
				_, err := client.ControlThermostatModeFeature(
					context.Background(),
					&alexamodels.ThermostatModeControlRequest{BaseFeatureRequest: base, Mode: "HEAT"},
				)

				return err
			},
		},
	}
}

func TestFeatureControlMethodsSerializeOperationPayloads(t *testing.T) {
	t.Parallel()

	base := alexamodels.BaseFeatureRequest{
		EndpointID: syntheticEndpointID,
		EntityID:   "synthetic-entity",
		Instance:   "synthetic-instance",
	}
	cases := featureControlTestCases(base)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			called := false

			client := NewClient(
				WithBaseURL("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-control-token"),
				WithHTTPClient(
					&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
						called = true
						var envelope struct {
							Query     string `json:"query"`
							Variables struct {
								Input struct {
									FeatureControlRequests []map[string]any `json:"featureControlRequests"`
								} `json:"input"`
							} `json:"variables"`
						}
						err := json.NewDecoder(request.Body).Decode(&envelope)
						if err != nil {
							t.Fatalf("decode feature request: %v", err)
						}
						if envelope.Query == "" || len(envelope.Variables.Input.FeatureControlRequests) != 1 {
							t.Fatalf("unexpected feature mutation envelope: %#v", envelope)
						}
						featureRequest := envelope.Variables.Input.FeatureControlRequests[0]
						if featureRequest["featureName"] != testCase.feature ||
							featureRequest["featureOperationName"] != testCase.operation {
							t.Errorf("unexpected feature operation: %#v", featureRequest)
						}
						if featureRequest["endpointId"] != syntheticEndpointID ||
							featureRequest["entityId"] != "synthetic-entity" {
							t.Errorf("feature endpoint or entity was lost: %#v", featureRequest)
						}
						if !sameJSONValue(t, featureRequest["payload"], testCase.payload) {
							t.Errorf(
								"unexpected feature payload: got %#v want %#v",
								featureRequest["payload"],
								testCase.payload,
							)
						}

						return graphqlSyntheticResponse(
							request,
							http.StatusOK,
							`{"data":{"setEndpointFeatures":{"featureControlResponses":[],"errors":[]}}}`,
						), nil
					})},
				),
			)

			err := testCase.call(client)
			if err != nil {
				t.Fatalf("feature control returned an error: %v", err)
			}

			if !called {
				t.Fatal("feature control did not submit a mutation")
			}
		})
	}
}

func TestFeatureControlMethodsRejectMissingSetOrDeltaValues(t *testing.T) {
	t.Parallel()

	client := NewClient()

	invalidRequests := []struct {
		name string
		call func() error
	}{
		{name: "speaker", call: func() error {
			_, err := client.ControlSpeakerFeature(context.Background(), &alexamodels.SpeakerControlRequest{})

			return err
		}},
		{name: "brightness", call: func() error {
			_, err := client.ControlBrightnessFeature(context.Background(), &alexamodels.BrightnessControlRequest{})

			return err
		}},
		{name: "mode", call: func() error {
			_, err := client.ControlModeFeature(context.Background(), &alexamodels.ModeControlRequest{})

			return err
		}},
		{name: "range", call: func() error {
			_, err := client.ControlRangeFeature(context.Background(), &alexamodels.RangeControlRequest{})

			return err
		}},
		{name: "percentage", call: func() error {
			_, err := client.ControlPercentageFeature(context.Background(), &alexamodels.PercentageControlRequest{})

			return err
		}},
		{name: "power level", call: func() error {
			_, err := client.ControlPowerLevelFeature(context.Background(), &alexamodels.PowerLevelControlRequest{})

			return err
		}},
		{name: "thermostat", call: func() error {
			_, err := client.ControlThermostatFeature(context.Background(), &alexamodels.ThermostatControlRequest{})

			return err
		}},
	}
	for _, testCase := range invalidRequests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.call()
			if !alexaapimodels.IsBadRequestError(err) {
				t.Fatalf("expected BadRequestError, got %v", err)
			}
		})
	}
}

func TestFeatureControlClientErrorsPropagate(t *testing.T) {
	t.Parallel()

	want := "synthetic GraphQL transport failure"

	client := NewClient(
		WithBaseURL("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-token"),
		WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(request *http.Request) (*http.Response, error) {
			return graphqlSyntheticResponse(request, http.StatusBadGateway, want), nil
		})}),
	)
	{
		_, err := client.ControlToggleFeature(
			context.Background(),
			&alexamodels.ToggleControlRequest{
				BaseFeatureRequest: alexamodels.BaseFeatureRequest{EndpointID: syntheticEndpointID},
				State:              "ON",
			},
		)
		if err == nil {
			t.Fatal("expected GraphQL transport error to propagate")
		}
	}
}

func featureTestPointer[T any](value T) *T { return &value }

func sameJSONValue(t *testing.T, actual, expected any) bool {
	t.Helper()

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Errorf("marshal actual JSON value: %v", err)

		return false
	}

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Errorf("marshal expected JSON value: %v", err)

		return false
	}

	var normalizedActual, normalizedExpected any
	{
		err := json.Unmarshal(actualJSON, &normalizedActual)
		if err != nil {
			t.Errorf("normalize actual JSON value: %v", err)

			return false
		}
	}

	{
		err := json.Unmarshal(expectedJSON, &normalizedExpected)
		if err != nil {
			t.Errorf("normalize expected JSON value: %v", err)

			return false
		}
	}

	return reflect.DeepEqual(normalizedActual, normalizedExpected)
}
