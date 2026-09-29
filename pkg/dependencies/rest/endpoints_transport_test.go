//nolint:testpackage // Exercises private endpoint request construction and decoding behavior.
package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func TestEndpointQueryMutationsAndDevicesV2UseExpectedRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		call  func(*Client) error
		check func(*testing.T, *http.Request)
	}{
		{
			name: "query endpoints",
			call: func(client *Client) error {
				_, err := client.QueryEndpoints(context.Background(), &alexamodels.EndpointQueryRequest{
					Query: alexamodels.EndpointQuery{AND: []alexamodels.EndpointQueryClause{{AssociatedUnits: &alexamodels.AssociatedUnitsFilter{ID: "synthetic-unit"}, Manufacturer: nil, Model: nil}}},
				}, &ListEndpointsOptions{NextToken: "synthetic-next", MaxResults: 4, Expand: []string{"all"}})

				return err
			},
			check: func(t *testing.T, request *http.Request) {
				t.Helper()
				assertEndpointQueryRequest(t, request)
			},
		},
		{
			name: "forget endpoint",
			call: func(client *Client) error { return client.ForgetEndpoint(context.Background(), "synthetic-endpoint") },
			check: func(t *testing.T, request *http.Request) {
				t.Helper()
				assertEndpointActionRequest(t, request, "/v2/endpoints/synthetic-endpoint/forget", "forget")
			},
		},
		{
			name: "deregister endpoint",
			call: func(client *Client) error {
				return client.DeregisterEndpoint(context.Background(), "synthetic-endpoint")
			},
			check: func(t *testing.T, request *http.Request) {
				t.Helper()
				assertEndpointActionRequest(t, request, "/v2/endpoints/synthetic-endpoint/deregister", "deregister")
			},
		},
		{
			name: "friendly name",
			call: func(client *Client) error {
				return client.UpdateFriendlyName(context.Background(), "synthetic-endpoint", "Kitchen light")
			},
			check: func(t *testing.T, request *http.Request) {
				t.Helper()
				assertFriendlyNameRequest(t, request)
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-rest-token"),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					if request.Header.Get("Authorization") != "Bearer synthetic-rest-token" {
						t.Errorf("unexpected Authorization header: %q", request.Header.Get("Authorization"))
					}
					testCase.check(t, request)

					return syntheticResponse(request, http.StatusOK, `{"results":[]}`), nil
				})}),
			)

			err := testCase.call(client)
			if err != nil {
				t.Fatalf("endpoint route returned an error: %v", err)
			}
		})
	}

	t.Run("devices-v2", func(t *testing.T) {
		client := NewClient(
			WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
			WithBearerToken("synthetic-rest-token"),
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				assertDevicesV2Request(t, request)

				return syntheticResponse(request, http.StatusOK, `{}`), nil
			})}),
		)

		_, devicesErr := client.GetDevicesV2(context.Background(), &GetDevicesV2Options{CSRFToken: "synthetic-csrf"})
		if devicesErr != nil {
			t.Fatalf("GetDevicesV2 returned an error: %v", devicesErr)
		}
	})
}

func assertEndpointQueryRequest(t *testing.T, request *http.Request) {
	t.Helper()

	if request.Method != http.MethodPost || request.URL.Path != "/v2/endpoint-query" ||
		request.URL.Query().Get("nextToken") != "synthetic-next" ||
		request.URL.Query().Get("maxResults") != "4" || request.URL.Query().Get("expand") != "all" {
		t.Errorf("unexpected endpoint-query route: %s %s", request.Method, request.URL)
	}

	var body map[string]any

	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		t.Fatalf("decode endpoint-query input: %v", err)
	}

	query, ok := body["query"].(map[string]any)
	if !ok || query["and"] == nil {
		t.Errorf("endpoint query was not serialized: %#v", body)
	}
}

func assertEndpointActionRequest(t *testing.T, request *http.Request, path, action string) {
	t.Helper()

	if request.Method != http.MethodPost || request.URL.Path != path {
		t.Errorf("unexpected %s route: %s %s", action, request.Method, request.URL)
	}
}

func assertFriendlyNameRequest(t *testing.T, request *http.Request) {
	t.Helper()

	assertEndpointActionRequest(t, request, "/v2/endpoints/synthetic-endpoint/friendlyName", "friendly-name")

	var body map[string]any

	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		t.Fatalf("decode friendly-name input: %v", err)
	}

	friendlyName := requireTestObject(t, body["friendlyName"])
	value := requireTestObject(t, friendlyName["value"])

	if friendlyName["type"] != "PLAIN" || value["text"] != "Kitchen light" {
		t.Errorf("unexpected friendly-name body: %#v", body)
	}
}

func assertDevicesV2Request(t *testing.T, request *http.Request) {
	t.Helper()

	if request.Method != http.MethodGet || request.URL.Path != "/api/devices-v2/device" {
		t.Errorf("unexpected devices-v2 route: %s %s", request.Method, request.URL)
	}

	if request.Header.Get("Authorization") != "Bearer synthetic-rest-token" ||
		request.Header.Get("Cookie") != "csrf=synthetic-csrf" {
		t.Errorf("unexpected devices-v2 credentials: %#v", request.Header)
	}
}

func TestMediaControlRoutesAndPlayerStateUseEndpointIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		call    func(*Client, *alexamodels.MediaControlRequest) error
		flags   map[string]any
	}{
		{name: "pause", command: "PauseCommand", flags: nil, call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.PausePlayback(context.Background(), r)
		}},
		{name: "resume", command: "PlayCommand", flags: nil, call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.ResumePlayback(context.Background(), r)
		}},
		{name: "next", command: "NextCommand", flags: nil, call: func(c *Client, r *alexamodels.MediaControlRequest) error { return c.NextTrack(context.Background(), r) }},
		{name: "previous", command: "PreviousCommand", flags: nil, call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.PreviousTrack(context.Background(), r)
		}},
		{name: "forward", command: "ForwardCommand", flags: nil, call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.ForwardMedia(context.Background(), r)
		}},
		{name: "rewind", command: "RewindCommand", flags: nil, call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.RewindMedia(context.Background(), r)
		}},
		{name: "shuffle", command: "ShuffleCommand", call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.SetShuffle(context.Background(), r, true)
		}, flags: map[string]any{"shuffle": true}},
		{name: "repeat", command: "RepeatCommand", call: func(c *Client, r *alexamodels.MediaControlRequest) error {
			return c.SetRepeat(context.Background(), r, true)
		}, flags: map[string]any{"repeat": true}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-media-token"),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					assertMediaControlRequest(t, request, testCase.command, testCase.flags)

					return syntheticResponse(request, http.StatusNoContent, ""), nil
				})}),
			)

			request := &alexamodels.MediaControlRequest{Endpoint: syntheticEndpoint{device: "synthetic-device-type", serial: "synthetic-device-serial"}}

			err := testCase.call(client, request)
			if err != nil {
				t.Fatalf("media command returned an error: %v", err)
			}
		})
	}

	t.Run("player state", func(t *testing.T) {
		client := NewClient(
			WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
			WithBearerToken("synthetic-media-token"),
			WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
				assertPlayerStateRequest(t, request)

				return syntheticResponse(request, http.StatusOK, `{}`), nil
			})}),
		)

		response, err := client.GetPlayerState(context.Background(), &alexamodels.PlayerStateRequest{Endpoint: syntheticEndpoint{device: "synthetic-device-type", serial: "synthetic-device-serial"}})
		if err != nil || response == nil {
			t.Fatalf("GetPlayerState returned %#v, %v", response, err)
		}
	})
}

func assertMediaControlRequest(t *testing.T, request *http.Request, command string, flags map[string]any) {
	t.Helper()

	if request.Method != http.MethodPost || request.URL.Path != "/api/np/command" ||
		request.URL.Query().Get("deviceType") != "synthetic-device-type" ||
		request.URL.Query().Get("deviceSerialNumber") != "synthetic-device-serial" {
		t.Errorf("unexpected media command route: %s %s", request.Method, request.URL)
	}

	if request.Header.Get("Authorization") != "Bearer synthetic-media-token" {
		t.Errorf("unexpected media command authorization: %q", request.Header.Get("Authorization"))
	}

	var body map[string]any

	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		t.Fatalf("decode media command: %v", err)
	}

	if body["type"] != command {
		t.Errorf("unexpected media command body: %#v", body)
	}

	for key, value := range flags {
		if body[key] != value {
			t.Errorf("unexpected command flag %s: %#v", key, body[key])
		}
	}
}

func assertPlayerStateRequest(t *testing.T, request *http.Request) {
	t.Helper()

	if request.Method != http.MethodGet || request.URL.Path != "/api/np/player" ||
		request.URL.Query().Get("deviceType") != "synthetic-device-type" ||
		request.URL.Query().Get("deviceSerialNumber") != "synthetic-device-serial" {
		t.Errorf("unexpected player-state route: %s %s", request.Method, request.URL)
	}
}

//nolint:funlen // Keep this synthetic operation matrix together under one preview wire contract.
func TestBehaviorMethodsSerializeSyntheticSequences(t *testing.T) {
	t.Parallel()

	endpoint := syntheticEndpoint{id: "synthetic-endpoint", device: "synthetic-device-type", serial: "synthetic-device-serial", locale: "fr-FR", family: alexamodels.DeviceFamilyFireTV, account: "synthetic-account"}

	type behaviorCase struct {
		name      string
		operation string
		payload   map[string]any
		call      func(*Client) error
	}

	seconds := 30
	volume := 45

	tests := []behaviorCase{
		{
			name: "send sequence", operation: "synthetic.Operation", payload: map[string]any{"deviceType": "synthetic-device-type", "deviceSerialNumber": "synthetic-device-serial"},
			call: func(client *Client) error {
				return client.SendSequence(context.Background(), "synthetic.Operation", map[string]any{"deviceType": "synthetic-device-type", "deviceSerialNumber": "synthetic-device-serial"})
			},
		},
		{
			name: "stop playback", operation: alexamodels.OperationTypeDeviceControlsStop, payload: map[string]any{"deviceType": "synthetic-device-type", "deviceSerialNumber": "synthetic-device-serial", "customerId": "synthetic-client-customer", "skillId": alexamodels.SkillIDAlexaDeviceControls},
			call: func(client *Client) error {
				return client.StopPlayback(context.Background(), &alexamodels.StopPlaybackRequest{Endpoint: endpoint, CustomerID: "synthetic-customer"})
			},
		},
		{
			name: "volume behavior", operation: alexamodels.OperationTypeDeviceControlsVolume, payload: map[string]any{"deviceSerialNumber": "synthetic-device-serial", "value": 45},
			call: func(client *Client) error {
				return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{Endpoint: endpoint, Volume: &volume, CustomerID: "synthetic-customer"})
			},
		},
		{
			name: "notification", operation: alexamodels.OperationTypeNotificationsSendMobilePush, payload: map[string]any{"deviceType": "synthetic-device-type", "deviceSerialNumber": "synthetic-device-serial", "notificationMessage": "Synthetic alert", "title": "Synthetic title", "customerId": "synthetic-client-customer"},
			call: func(client *Client) error {
				return client.SendNotification(context.Background(), &alexamodels.SendNotificationRequest{Endpoint: endpoint, Message: "Synthetic alert", Title: "Synthetic title", CustomerID: "request-customer"})
			},
		},
		{
			name: "announcement speak", operation: alexamodels.OperationTypeAnnouncement, payload: map[string]any{"deviceSerialNumber": "synthetic-device-serial", "locale": "de-DE", "customerId": "synthetic-client-customer"},
			call: func(client *Client) error {
				return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{Endpoint: endpoint, Message: "Synthetic words", Method: alexamodels.AnnouncementMethodSpeak, Locale: "de-DE", CustomerID: "request-customer"})
			},
		},
		{
			name: "announcement show", operation: alexamodels.OperationTypeAnnouncement, payload: map[string]any{"deviceSerialNumber": "synthetic-device-serial", "locale": "fr-FR"},
			call: func(client *Client) error {
				return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{Endpoint: endpoint, Message: "Synthetic display", Method: alexamodels.AnnouncementMethodShow, Title: "Synthetic title"})
			},
		},
		{
			name: "announcement all", operation: alexamodels.OperationTypeAnnouncement, payload: map[string]any{"deviceSerialNumber": "synthetic-device-serial", "locale": "fr-FR"},
			call: func(client *Client) error {
				return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{Endpoint: endpoint, Message: "Synthetic display and speech", Method: alexamodels.AnnouncementMethodAll})
			},
		},
		{
			name: "text to speech", operation: alexamodels.OperationTypeSpeak, payload: map[string]any{"deviceSerialNumber": "synthetic-device-serial", "textToSpeak": "Synthetic spoken text"},
			call: func(client *Client) error {
				return client.SendTTS(context.Background(), &alexamodels.SendTTSRequest{Endpoint: endpoint, Message: "Synthetic spoken text"})
			},
		},
		{
			name: "play music with timer", operation: alexamodels.OperationTypeMusicPlaySearchPhrase, payload: map[string]any{"searchPhrase": "Synthetic artist", "sanitizedSearchPhrase": "Synthetic artist", "musicProviderId": "synthetic-provider", "waitTimeInSeconds": 30},
			call: func(client *Client) error {
				return client.PlayMusic(context.Background(), &alexamodels.PlayMusicRequest{Endpoint: endpoint, SearchPhrase: "Synthetic artist", ProviderID: "synthetic-provider", TimerSeconds: &seconds})
			},
		},
		{
			name: "play audio uri", operation: alexamodels.OperationTypeSound, payload: map[string]any{"soundStringId": "https://media.example.invalid/synthetic.mp3"},
			call: func(client *Client) error {
				return client.PlayAudioURI(context.Background(), &alexamodels.PlayAudioURIRequest{Endpoint: endpoint, URI: "https://media.example.invalid/synthetic.mp3"})
			},
		},
		{
			name: "play video with provider", operation: alexamodels.OperationTypeVideoPlaySearchPhrase, payload: map[string]any{"searchPhrase": "Synthetic title on Synthetic Video", "sanitizedSearchPhrase": "Synthetic title on Synthetic Video", "waitTimeInSeconds": 30},
			call: func(client *Client) error {
				return client.PlayVideo(context.Background(), &alexamodels.PlayVideoRequest{Endpoint: endpoint, SearchPhrase: "Synthetic title", VideoProviderID: "Synthetic Video", TimerSeconds: &seconds})
			},
		},
		{
			name: "play video without provider", operation: alexamodels.OperationTypeVideoPlaySearchPhrase, payload: map[string]any{"searchPhrase": "Synthetic title", "sanitizedSearchPhrase": "Synthetic title"},
			call: func(client *Client) error {
				return client.PlayVideo(context.Background(), &alexamodels.PlayVideoRequest{Endpoint: endpoint, SearchPhrase: "Synthetic title"})
			},
		},
		{
			name: "fire tv sequence", operation: alexamodels.OperationTypeFireTVPauseVideo, payload: map[string]any{"deviceAccountId": "synthetic-account", "skillId": alexamodels.SkillIDRoutinesFireTV},
			call: func(client *Client) error {
				return client.SendFireTVSequence(context.Background(), "synthetic-account", alexamodels.OperationTypeFireTVPauseVideo)
			},
		},
		{
			name: "fire tv on", operation: alexamodels.OperationTypeFireTVTurnOn, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVTurnOn(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv off", operation: alexamodels.OperationTypeFireTVTurnOff, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVTurnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv turn on/off true", operation: alexamodels.OperationTypeFireTVTurnOn, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVTurnOnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint}, true)
			},
		},
		{
			name: "fire tv turn on/off false", operation: alexamodels.OperationTypeFireTVTurnOff, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVTurnOnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint}, false)
			},
		},
		{
			name: "fire tv pause", operation: alexamodels.OperationTypeFireTVPauseVideo, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVPauseVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv resume", operation: alexamodels.OperationTypeFireTVResumeVideo, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVResumeVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv home", operation: alexamodels.OperationTypeFireTVNavigateHome, payload: map[string]any{"deviceAccountId": "synthetic-account"},
			call: func(client *Client) error {
				return client.FireTVNavigateHome(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(
				WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-behavior-token"),
				WithCustomerID("synthetic-client-customer"),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					assertBehaviorSequenceRequest(t, request, testCase.operation, testCase.payload)

					return syntheticResponse(request, http.StatusOK, `{}`), nil
				})}),
			)

			err := testCase.call(client)
			if err != nil {
				t.Fatalf("behavior operation returned an error: %v", err)
			}
		})
	}
}

func assertBehaviorSequenceRequest(t *testing.T, request *http.Request, operation string, payload map[string]any) {
	t.Helper()

	if request.Method != http.MethodPost || request.URL.Path != "/api/behaviors/preview" ||
		request.Header.Get("Authorization") != "Bearer synthetic-behavior-token" {
		t.Errorf("unexpected behavior request: %s %s auth=%q", request.Method, request.URL, request.Header.Get("Authorization"))
	}

	var body struct {
		BehaviorID   string `json:"behaviorId"`
		SequenceJSON string `json:"sequenceJson"`
		Status       string `json:"status"`
	}

	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		t.Fatalf("decode behavior preview request: %v", err)
	}

	if body.BehaviorID != "PREVIEW" || body.Status != "ENABLED" {
		t.Errorf("unexpected behavior preview metadata: %#v", body)
	}

	var sequence map[string]any

	err = json.Unmarshal([]byte(body.SequenceJSON), &sequence)
	if err != nil {
		t.Fatalf("decode behavior sequence: %v", err)
	}

	if sequence["@type"] != alexamodels.ModelTypeSequence {
		t.Errorf("unexpected sequence type: %#v", sequence)
	}

	node, ok := sequence["startNode"].(map[string]any)
	if !ok || node["type"] != operation || node["@type"] != alexamodels.ModelTypeOpaquePayloadOperationNode {
		t.Errorf("unexpected operation node: %#v", sequence["startNode"])
	}

	if got, ok := node["operationPayload"].(map[string]any); !ok || !restJSONSubset(t, got, payload) {
		t.Errorf("unexpected operation payload: %#v", node["operationPayload"])
	}
}

func TestBehaviorMethodsValidateEndpointAndVolumeInputs(t *testing.T) {
	t.Parallel()

	client := NewClient()

	invalid := []struct {
		name string
		call func() error
	}{
		{name: "stop playback endpoint", call: func() error { return client.StopPlayback(context.Background(), &alexamodels.StopPlaybackRequest{}) }},
		{name: "set volume endpoint", call: func() error { return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{}) }},
		{name: "negative volume", call: func() error {
			value := -1

			return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{Endpoint: syntheticEndpoint{}, Volume: &value})
		}},
		{name: "volume over maximum", call: func() error {
			value := 101

			return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{Endpoint: syntheticEndpoint{}, Volume: &value})
		}},
		{name: "notification endpoint", call: func() error {
			return client.SendNotification(context.Background(), &alexamodels.SendNotificationRequest{})
		}},
		{name: "announcement endpoint", call: func() error {
			return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{})
		}},
		{name: "tts endpoint", call: func() error { return client.SendTTS(context.Background(), &alexamodels.SendTTSRequest{}) }},
		{name: "music endpoint", call: func() error { return client.PlayMusic(context.Background(), &alexamodels.PlayMusicRequest{}) }},
		{name: "audio endpoint", call: func() error { return client.PlayAudioURI(context.Background(), &alexamodels.PlayAudioURIRequest{}) }},
		{name: "video endpoint", call: func() error { return client.PlayVideo(context.Background(), &alexamodels.PlayVideoRequest{}) }},
		{name: "fire tv missing account", call: func() error {
			return client.FireTVTurnOn(context.Background(), &alexamodels.FireTVRequest{Endpoint: syntheticEndpoint{family: alexamodels.DeviceFamilyFireTV}})
		}},
		{name: "fire tv wrong family", call: func() error {
			return client.FireTVTurnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: syntheticEndpoint{account: "synthetic-account", family: "OTHER"}})
		}},
		{name: "fire tv pause wrong family", call: func() error {
			return client.FireTVPauseVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: syntheticEndpoint{account: "synthetic-account", family: "OTHER"}})
		}},
		{name: "fire tv resume wrong family", call: func() error {
			return client.FireTVResumeVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: syntheticEndpoint{account: "synthetic-account", family: "OTHER"}})
		}},
		{name: "fire tv home wrong family", call: func() error {
			return client.FireTVNavigateHome(context.Background(), &alexamodels.FireTVRequest{Endpoint: syntheticEndpoint{account: "synthetic-account", family: "OTHER"}})
		}},
	}
	for _, testCase := range invalid {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.call()
			if !alexaapimodels.IsBadRequestError(err) {
				t.Fatalf("expected BadRequestError, got %v", err)
			}
		})
	}
}

func restJSONSubset(t *testing.T, actual, expected map[string]any) bool {
	t.Helper()

	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok {
			t.Errorf("payload missing key %q: %#v", key, actual)

			return false
		}

		if !sameRestJSON(actualValue, expectedValue) {
			t.Errorf("payload key %q got %#v, want %#v", key, actualValue, expectedValue)

			return false
		}
	}

	return true
}

func sameRestJSON(actual, expected any) bool {
	actualData, err := json.Marshal(actual)
	if err != nil {
		return false
	}

	expectedData, err := json.Marshal(expected)
	if err != nil {
		return false
	}

	var actualNormalized, expectedNormalized any
	if json.Unmarshal(actualData, &actualNormalized) != nil || json.Unmarshal(expectedData, &expectedNormalized) != nil {
		return false
	}

	return reflect.DeepEqual(actualNormalized, expectedNormalized)
}

func TestMediaAndBehaviorErrorsPreserveSyntheticHTTPFailure(t *testing.T) {
	t.Parallel()

	client := NewClient(
		WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-error-token"),
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
			return syntheticResponse(request, http.StatusServiceUnavailable, "synthetic unavailable"), nil
		})}),
	)

	err := client.PausePlayback(context.Background(), &alexamodels.MediaControlRequest{Endpoint: syntheticEndpoint{device: "type", serial: "serial"}})
	if !alexaapimodels.IsNetworkError(err) {
		t.Fatalf("expected media wrapper NetworkError, got %T: %v", err, err)
	}

	err = client.RunBehavior(context.Background(), `{"@type":"com.amazon.alexa.behaviors.model.Sequence","startNode":{"@type":"com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode","type":"synthetic.Operation","operationPayload":{}}}`)
	if !alexaapimodels.IsInternalServerError(err) {
		t.Fatalf("expected behavior InternalServerError, got %T: %v", err, err)
	}
}

func TestRunBehaviorRejectsUnschematizedSequenceBeforeTransport(t *testing.T) {
	t.Parallel()

	client := NewClient(WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(*http.Request) (*http.Response, error) {
		t.Fatal("unschematized sequence reached transport")

		return nil, syntheticFailureError("unschematized sequence reached transport")
	})}))
	for _, sequence := range []string{
		`{}`,
		`{"@type":"com.amazon.alexa.behaviors.model.Sequence","startNode":{"@type":"unknown"}}`,
		`{"@type":"com.amazon.alexa.behaviors.model.Sequence","startNode":{"@type":"com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode","type":"","operationPayload":{}}}`,
	} {
		err := client.RunBehavior(context.Background(), sequence)
		if !alexaapimodels.IsBadRequestError(err) {
			t.Fatalf("expected schema rejection for %s, got %v", sequence, err)
		}
	}
}

func TestEndpointQueryRejectsNilInput(t *testing.T) {
	t.Parallel()

	client := NewClient()
	{
		_, err := client.QueryEndpoints(context.Background(), nil, nil)
		if !alexaapimodels.IsBadRequestError(err) {
			t.Fatalf("expected nil query to return BadRequestError, got %v", err)
		}
	}
}

func TestQueryEndpointPaginationParametersAreEncoded(t *testing.T) {
	t.Parallel()

	path := buildEndpointListPath("/v2/endpoints", &ListEndpointsOptions{NextToken: "synthetic token/1", MaxResults: 12})

	parsed, err := url.Parse(path)
	if err != nil {
		t.Fatalf("parse query path: %v", err)
	}

	if parsed.Query().Get("nextToken") != "synthetic token/1" || parsed.Query().Get("maxResults") != strconv.Itoa(12) {
		t.Fatalf("unexpected pagination query: %s", parsed.RawQuery)
	}
}
