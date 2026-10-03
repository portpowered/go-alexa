//nolint:testpackage // Exercises private dispatch paths with synthetic provider responses.
package alexa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	m "github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// Synthetic control requests verify that public dispatch reaches the expected
// GraphQL feature edge with a concrete endpoint and operation.
type syntheticFeatureDispatchCase struct {
	name      string
	feature   m.FeatureName
	operation m.FeatureOperationName
	payload   any
}

func syntheticFeatureDispatchCases() []syntheticFeatureDispatchCase {
	cases := make([]syntheticFeatureDispatchCase, 0, 24)
	cases = append(cases, syntheticFeatureDispatchCasesFirst()...)
	cases = append(cases, syntheticFeatureDispatchCasesSecond()...)

	return cases
}

func syntheticFeatureDispatchCasesFirst() []syntheticFeatureDispatchCase {
	return []syntheticFeatureDispatchCase{
		{
			"brightness set",
			m.FeatureNameBrightness,
			m.FeatureOperationNameSetBrightness,
			m.ControlBrightnessSetPayload{Brightness: 75},
		},
		{
			"brightness adjust",
			m.FeatureNameBrightness,
			m.FeatureOperationNameAdjustBrightness,
			m.ControlBrightnessAdjustPayload{Delta: -5},
		},
		{
			"color",
			m.FeatureNameColor,
			m.FeatureOperationNameSetColor,
			m.ControlColorPayload{Hue: 180, Saturation: .5, Brightness: .8},
		},
		{
			"color temperature set",
			m.FeatureNameColorTemperature,
			m.FeatureOperationNameSetColorTemperature,
			m.ControlColorTemperatureSetPayload{ColorTemperature: 3200},
		},
		{
			"color temperature increase",
			m.FeatureNameColorTemperature,
			m.FeatureOperationNameIncreasColorTemperature,
			m.ControlColorTemperatureAdjustPayload{Increase: true},
		},
		{
			"color temperature decrease",
			m.FeatureNameColorTemperature,
			m.FeatureOperationNameDecreaseColorTemperature,
			m.ControlColorTemperatureAdjustPayload{},
		},
		{"lock", m.FeatureNameLock, m.FeatureOperationNameSetLockState, m.ControlLockPayload{State: "LOCKED"}},
		{"mode set", m.FeatureNameMode, m.FeatureOperationNameSetMode, m.ControlModeSetPayload{Mode: "ECO"}},
		{"mode adjust", m.FeatureNameMode, m.FeatureOperationNameAdjustMode, m.ControlModeAdjustPayload{Delta: "NEXT"}},
		{
			"range set",
			m.FeatureNameRange,
			m.FeatureOperationNameSetRangeValue,
			m.ControlRangeSetPayload{RangeValue: 12.5},
		},
		{
			"range adjust",
			m.FeatureNameRange,
			m.FeatureOperationNameAdjustRangeValue,
			m.ControlRangeAdjustPayload{Delta: -2.5},
		},
		{"toggle", m.FeatureNameToggle, m.FeatureOperationNameSetToggleState, m.ControlTogglePayload{State: "ON"}},
	}
}

func syntheticFeatureDispatchCasesSecond() []syntheticFeatureDispatchCase {
	return []syntheticFeatureDispatchCase{
		{
			"percentage set",
			m.FeatureNamePercentage,
			m.FeatureOperationNameSetPercentage,
			m.ControlPercentageSetPayload{Percentage: 25},
		},
		{
			"percentage adjust",
			m.FeatureNamePercentage,
			m.FeatureOperationNameAdjustPercentage,
			m.ControlPercentageAdjustPayload{Delta: -5},
		},
		{
			"power level set",
			m.FeatureNamePowerLevel,
			m.FeatureOperationNameSetPowerLevel,
			m.ControlPowerLevelSetPayload{PowerLevel: 68},
		},
		{
			"power level adjust",
			m.FeatureNamePowerLevel,
			m.FeatureOperationNameAdjustPowerLevel,
			m.ControlPowerLevelAdjustPayload{Delta: -8},
		},
		{
			"action",
			m.FeatureNameAction,
			m.FeatureOperationNamePerformAction,
			m.ControlActionPayload{Action: "open", Params: map[string]any{"duration": 12}},
		},
		{"power on", m.FeatureNamePower, m.FeatureOperationNameTurnOn, m.ControlPowerPayload{}},
		{"power off", m.FeatureNamePower, m.FeatureOperationNameTurnOff, m.ControlPowerPayload{}},
		{"volume set", m.FeatureNameSpeaker, m.FeatureOperationNameSetVolume, m.ControlVolumeSetPayload{Volume: 32}},
		{
			"volume adjust",
			m.FeatureNameSpeaker,
			m.FeatureOperationNameAdjustVolume,
			m.ControlVolumeAdjustPayload{Delta: -4},
		},
		{
			"thermostat mode",
			m.FeatureNameThermostat,
			m.FeatureOperationNameSetThermostatMode,
			m.ControlThermostatModePayload{Mode: "ECO"},
		},
		{
			"thermostat setpoint",
			m.FeatureNameThermostat,
			m.FeatureOperationNameSetTargetSetpoint,
			m.ControlThermostatSetpointPayload{Value: 21.5, Scale: "CELSIUS"},
		},
		{
			"thermostat adjust",
			m.FeatureNameThermostat,
			m.FeatureOperationNameAdjustTargetSetpoint,
			m.ControlThermostatSetpointAdjustPayload{Delta: -1, Scale: "CELSIUS"},
		},
	}
}

func TestSyntheticFeatureControlDispatch(t *testing.T) {
	t.Parallel()

	tests := syntheticFeatureDispatchCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			requests := 0
			session := newTestSession(
				t,
				&http.Client{Transport: syntheticEventTransport(func(request *http.Request) (*http.Response, error) {
					requests++
					if request.Header.Get("Authorization") != "Bearer synthetic-token" {
						t.Errorf("missing bearer token")
					}
					var envelope struct {
						Variables struct {
							Input struct {
								FeatureControlRequests []map[string]any `json:"featureControlRequests"`
							} `json:"input"`
						} `json:"variables"`
					}
					err := json.NewDecoder(request.Body).Decode(&envelope)
					if err != nil {
						t.Errorf("decode control request: %v", err)
					}
					if len(envelope.Variables.Input.FeatureControlRequests) != 1 {
						t.Errorf("control requests = %v", envelope.Variables.Input.FeatureControlRequests)
					} else {
						control := envelope.Variables.Input.FeatureControlRequests[0]
						if control["endpointId"] != "synthetic-endpoint" ||
							control["featureName"] != string(test.feature) ||
							control["featureOperationName"] == "" {
							t.Errorf("control target or operation = %#v", control)
						}
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(
							strings.NewReader(
								`{"data":{"setEndpointFeatures":{"featureControlResponses":[],"errors":[]}}}`,
							),
						),
						Header: make(http.Header),
					}, nil
				})},
				WithBearerToken("synthetic-token"),
			)

			response, err := session.Control(
				context.Background(),
				m.ControlRequest{
					Target: &m.Endpoint{
						EndpointID:         "synthetic-endpoint",
						DeviceSerialNumber: "synthetic-serial",
						DeviceType:         "synthetic-type",
					},
					Namespace: test.feature,
					Name:      test.operation,
					Payload:   test.payload,
					Instance:  "synthetic-instance",
				},
			)
			if err != nil || response == nil {
				t.Fatalf("Control(%s, %s) = %+v, %v", test.feature, test.operation, response, err)
			}

			if requests != 1 {
				t.Fatalf("request count = %d", requests)
			}
		})
	}
}

type syntheticRESTDispatchCase struct {
	name      string
	feature   m.FeatureName
	operation m.FeatureOperationName
	payload   any
	family    string
	account   string
}

func syntheticRESTDispatchCases() []syntheticRESTDispatchCase {
	cases := make([]syntheticRESTDispatchCase, 0, 19)
	cases = append(cases, syntheticRESTDispatchCasesFirst()...)
	cases = append(cases, syntheticRESTDispatchCasesSecond()...)

	return cases
}

func syntheticRESTDispatchCasesFirst() []syntheticRESTDispatchCase {
	return []syntheticRESTDispatchCase{
		{
			"notification",
			m.FeatureNameNotification,
			m.FeatureOperationNameSend,
			m.ControlNotificationPayload{Message: "Synthetic reminder", Title: "Reminder"},
			"",
			"",
		},
		{
			"announcement",
			m.FeatureNameAnnouncement,
			m.FeatureOperationNameSend,
			m.ControlAnnouncementPayload{Message: "Synthetic notice", Method: "speak"},
			"",
			"",
		},
		{
			"speech",
			m.FeatureNameSpeechSynthesizer,
			m.FeatureOperationNameSpeak,
			m.ControlSpeechSynthesizerPayload{Message: "Synthetic speech"},
			"",
			"",
		},
		{
			"music",
			m.FeatureNameAudioPlayer,
			m.FeatureOperationNamePlay,
			m.ControlAudioPlayerPayload{ProviderID: m.ProviderIDAmazon, SearchPhrase: "Synthetic song"},
			"",
			"",
		},
		{
			"navigation",
			m.FeatureNameNavigation,
			m.FeatureOperationNameNavigateHome,
			m.ControlNavigationPayload{},
			"FIRE_TV",
			"synthetic-account",
		},
		{"playback play", m.FeatureNamePlayback, m.FeatureOperationNamePlay, m.ControlPlaybackPayload{}, "", ""},
		{"playback pause", m.FeatureNamePlayback, m.FeatureOperationNamePause, m.ControlPlaybackPayload{}, "", ""},
		{"playback resume", m.FeatureNamePlayback, m.FeatureOperationNameResume, m.ControlPlaybackPayload{}, "", ""},
		{"playback next", m.FeatureNamePlayback, m.FeatureOperationNameNext, m.ControlPlaybackPayload{}, "", ""},
		{
			"playback previous",
			m.FeatureNamePlayback,
			m.FeatureOperationNamePrevious,
			m.ControlPlaybackPayload{},
			"",
			"",
		},
	}
}

func syntheticRESTDispatchCasesSecond() []syntheticRESTDispatchCase {
	return []syntheticRESTDispatchCase{
		{"playback stop", m.FeatureNamePlayback, m.FeatureOperationNameStop, m.ControlPlaybackPayload{}, "", ""},
		{"playback forward", m.FeatureNamePlayback, m.FeatureOperationNameForward, m.ControlForwardPayload{}, "", ""},
		{"playback rewind", m.FeatureNamePlayback, m.FeatureOperationNameRewind, m.ControlRewindPayload{}, "", ""},
		{
			"playback shuffle",
			m.FeatureNamePlayback,
			m.FeatureOperationNameShuffle,
			m.ControlShufflePayload{Shuffle: true},
			"",
			"",
		},
		{
			"playback repeat",
			m.FeatureNamePlayback,
			m.FeatureOperationNameRepeat,
			m.ControlRepeatPayload{Repeat: true},
			"",
			"",
		},
		{
			"fire TV pause",
			m.FeatureNamePlayback,
			m.FeatureOperationNamePause,
			m.ControlPlaybackPayload{},
			"FIRE_TV",
			"synthetic-account",
		},
		{
			"fire TV resume",
			m.FeatureNamePlayback,
			m.FeatureOperationNameResume,
			m.ControlPlaybackPayload{},
			"FIRE_TV",
			"synthetic-account",
		},
		{
			"fire TV power on",
			m.FeatureNamePower,
			m.FeatureOperationNameTurnOn,
			m.ControlPowerPayload{},
			"FIRE_TV",
			"synthetic-account",
		},
		{
			"fire TV power off",
			m.FeatureNamePower,
			m.FeatureOperationNameTurnOff,
			m.ControlPowerPayload{},
			"FIRE_TV",
			"synthetic-account",
		},
	}
}

func TestSyntheticRESTControlDispatch(t *testing.T) {
	t.Parallel()

	tests := syntheticRESTDispatchCases()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			requests := 0
			session := newTestSession(
				t,
				&http.Client{Transport: syntheticEventTransport(func(request *http.Request) (*http.Response, error) {
					requests++
					if request.URL.Path == "" {
						t.Error("empty REST route")
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{}`)),
						Header:     make(http.Header),
					}, nil
				})},
				WithBearerToken("synthetic-token"),
			)

			response, err := session.Control(
				context.Background(),
				m.ControlRequest{
					Target: &m.Endpoint{
						EndpointID:         "synthetic-endpoint",
						DeviceSerialNumber: "synthetic-serial",
						DeviceType:         "synthetic-type",
						DeviceFamily:       test.family,
						DeviceAccountId:    test.account,
					},
					Namespace: test.feature,
					Name:      test.operation,
					Payload:   test.payload,
				},
			)
			if err != nil || response == nil {
				t.Fatalf("Control(%s, %s) = %+v, %v", test.feature, test.operation, response, err)
			}

			if requests != 1 {
				t.Fatalf("request count = %d", requests)
			}
		})
	}
}

func TestSyntheticRESTFeatureWireBodies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		feature   m.FeatureName
		operation m.FeatureOperationName
		payload   any
		path      string
		want      map[string]any
	}{
		{
			name:      "volume set includes zero",
			feature:   m.FeatureNameSpeaker,
			operation: m.FeatureOperationNameSetVolume,
			payload:   m.ControlVolumeSetPayload{Volume: 0},
			path:      "/v2/endpoints/synthetic-endpoint/interfaces/speaker/setVolume/",
			want:      map[string]any{"volume": float64(0)},
		},
		{
			name:      "playback pause sends empty object",
			feature:   m.FeatureNamePlayback,
			operation: m.FeatureOperationNamePause,
			payload:   m.ControlPlaybackPayload{},
			path:      "/v2/endpoints/synthetic-endpoint/interfaces/playback/pause/",
			want:      map[string]any{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			requests := 0
			session := newTestSession(
				t,
				&http.Client{Transport: syntheticEventTransport(func(request *http.Request) (*http.Response, error) {
					requests++
					if request.Method != http.MethodPost || request.URL.Path != test.path {
						t.Errorf("REST request = %s %s, want POST %s", request.Method, request.URL.Path, test.path)
					}

					var body map[string]any
					err := json.NewDecoder(request.Body).Decode(&body)
					if err != nil {
						t.Errorf("decode REST feature message: %v", err)
					}
					if !reflect.DeepEqual(body, test.want) {
						t.Errorf("REST payload = %#v, want %#v", body, test.want)
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{}`)),
						Header:     make(http.Header),
					}, nil
				})},
				WithBearerToken("synthetic-token"),
			)

			response, err := session.Control(
				context.Background(),
				m.ControlRequest{
					Target:    &m.Endpoint{EndpointID: "synthetic-endpoint"},
					Namespace: test.feature,
					Name:      test.operation,
					Payload:   test.payload,
				},
			)
			if err != nil || response == nil {
				t.Fatalf("Control(%s, %s) = %+v, %v", test.feature, test.operation, response, err)
			}

			if requests != 1 {
				t.Fatalf("request count = %d", requests)
			}
		})
	}
}

func TestSyntheticInterfaceControlDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		feature   m.FeatureName
		operation m.FeatureOperationName
		payload   any
	}{
		{"playback play", m.FeatureNamePlayback, m.FeatureOperationNamePlay, nil},
		{"playback pause", m.FeatureNamePlayback, m.FeatureOperationNamePause, nil},
		{"playback resume", m.FeatureNamePlayback, m.FeatureOperationNameResume, nil},
		{"playback next", m.FeatureNamePlayback, m.FeatureOperationNameNext, nil},
		{"playback previous", m.FeatureNamePlayback, m.FeatureOperationNamePrevious, nil},
		{"playback stop", m.FeatureNamePlayback, m.FeatureOperationNameStop, nil},
		{"playback forward", m.FeatureNamePlayback, m.FeatureOperationNameForward, nil},
		{"playback rewind", m.FeatureNamePlayback, m.FeatureOperationNameRewind, nil},
		{
			"playback shuffle",
			m.FeatureNamePlayback,
			m.FeatureOperationNameShuffle,
			m.ControlShufflePayload{Shuffle: false},
		},
		{"playback repeat", m.FeatureNamePlayback, m.FeatureOperationNameRepeat, m.ControlRepeatPayload{Repeat: false}},
		{"volume set", m.FeatureNameSpeaker, m.FeatureOperationNameSetVolume, m.ControlVolumeSetPayload{Volume: 20}},
		{
			"volume adjust",
			m.FeatureNameSpeaker,
			m.FeatureOperationNameAdjustVolume,
			m.ControlVolumeAdjustPayload{Delta: 5},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			requests := 0
			session := newTestSession(
				t,
				&http.Client{Transport: syntheticEventTransport(func(request *http.Request) (*http.Response, error) {
					requests++
					if request.URL.Path == "" {
						t.Error("empty interface route")
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{}`)),
						Header:     make(http.Header),
					}, nil
				})},
				WithBearerToken("synthetic-token"),
			)

			response, err := session.Control(
				context.Background(),
				m.ControlRequest{
					Target:    &m.Endpoint{EndpointID: "synthetic-third-party", DeviceType: "synthetic-type"},
					Namespace: test.feature,
					Name:      test.operation,
					Payload:   test.payload,
				},
			)
			if err != nil || response == nil {
				t.Fatalf("Control(%s, %s) = %+v, %v", test.feature, test.operation, response, err)
			}

			if requests != 1 {
				t.Fatalf("request count = %d", requests)
			}
		})
	}
}

func TestSyntheticControlRejectsUnknownOperationsBeforeNetwork(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: syntheticEventTransport(func(*http.Request) (*http.Response, error) {
		t.Error("invalid control reached network")

		return nil, staticError("invalid control reached transport")
	})}
	session := newTestSession(t, client, WithBearerToken("synthetic-token"))
	endpoint := &m.Endpoint{EndpointID: "synthetic-endpoint", DeviceSerialNumber: "synthetic-serial"}

	features := []m.FeatureName{
		m.FeatureNameNotification, m.FeatureNameAnnouncement, m.FeatureNameSpeechSynthesizer,
		m.FeatureNameAudioPlayer, m.FeatureNameNavigation, m.FeatureNameBrightness,
		m.FeatureNameColor, m.FeatureNameColorTemperature, m.FeatureNameLock,
		m.FeatureNameMode, m.FeatureNameRange, m.FeatureNameToggle,
		m.FeatureNamePercentage, m.FeatureNamePowerLevel, m.FeatureNameAction,
		m.FeatureNamePlayback, m.FeatureNamePower, m.FeatureNameSpeaker,
		m.FeatureNameThermostat,
	}
	for _, feature := range features {
		t.Run(feature.String(), func(t *testing.T) {
			t.Parallel()

			response, err := session.Control(
				context.Background(),
				m.ControlRequest{Target: endpoint, Namespace: feature, Name: "synthetic-unknown-operation"},
			)
			if !m.IsBadRequestError(err) || response != nil {
				t.Fatalf("invalid %s operation = %+v, %v", feature, response, err)
			}
		})
	}

	response, err := session.Control(
		context.Background(),
		m.ControlRequest{Target: endpoint, Namespace: "synthetic-unknown-feature", Name: "synthetic-unknown-operation"},
	)
	if !m.IsBadRequestError(err) || response == nil || len(response.Errors) != 1 {
		t.Fatalf("invalid namespace = %+v, %v", response, err)
	}
}
