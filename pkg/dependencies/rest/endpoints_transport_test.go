//nolint:testpackage // Exercises private endpoint request construction and decoding behavior.
package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
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
					Query: alexamodels.EndpointQuery{OR: nil, IncludeFields: nil, PaginationContext: nil, AND: []alexamodels.EndpointQueryClause{{AssociatedUnits: &alexamodels.AssociatedUnitsFilter{ID: "synthetic-unit"}, Manufacturer: nil, Model: nil}}},
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

//nolint:funlen,maintidx // Keep this synthetic operation matrix together under one preview wire contract.
func TestBehaviorMethodsSerializeSyntheticSequences(t *testing.T) {
	t.Parallel()

	endpoint := syntheticEndpoint{id: "synthetic-endpoint", device: "synthetic-device-type", serial: "synthetic-device-serial", locale: "fr-FR", family: alexamodels.DeviceFamilyFireTV, account: "synthetic-account", owner: "synthetic-device-owner"}

	type behaviorCase struct {
		name string
		node any
		call func(*Client) error
	}

	seconds := 30
	volume := 45
	behaviorNodeType := alexamodels.ComAmazonAlexaBehaviorsModelOpaquePayloadOperationNode
	deviceTarget := alexamodels.DeviceTarget{DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial"}
	sharedTarget := alexamodels.Target{
		CustomerID: "synthetic-client-customer",
		Devices:    []alexamodels.DeviceTarget{deviceTarget},
	}
	announcementContent := func(locale, title, message, spoken string) []alexamodels.BehaviorAnnouncementContent {
		return []alexamodels.BehaviorAnnouncementContent{{
			Locale: locale,
			Display: alexamodels.BehaviorAnnouncementDisplay{
				Title: title,
				Body:  message,
			},
			Speak: alexamodels.BehaviorAnnouncementSpeak{
				Type:  alexamodels.BehaviorAnnouncementSpeakTypeText,
				Value: spoken,
			},
		}}
	}

	tests := []behaviorCase{
		{
			name: "send sequence", node: alexamodels.OpaquePayloadOperationNode{
				Type: behaviorNodeType, OperationType: "synthetic.Operation",
				OperationPayload: map[string]interface{}{"deviceType": "synthetic-device-type", "deviceSerialNumber": "synthetic-device-serial"},
			},
			call: func(client *Client) error {
				return client.SendSequence(context.Background(), "synthetic.Operation", map[string]any{"deviceType": "synthetic-device-type", "deviceSerialNumber": "synthetic-device-serial"})
			},
		},
		{
			name: "stop playback", node: alexamodels.DeviceControlsStopOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.DeviceControlsStopOperationTypeStop,
				OperationPayload: alexamodels.DeviceControlsStopPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					SkillID: alexamodels.BehaviorSkillIDAlexaDeviceControls,
				},
			},
			call: func(client *Client) error {
				return client.StopPlayback(context.Background(), &alexamodels.StopPlaybackRequest{
					Endpoint: endpoint, AllDevices: false, CustomerID: "synthetic-customer",
				})
			},
		},
		{
			name: "volume behavior", node: alexamodels.DeviceControlsVolumeOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.DeviceControlsVolumeOperationTypeVolume,
				OperationPayload: alexamodels.DeviceControlsVolumePayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					Value: &volume, SkillID: alexamodels.BehaviorSkillIDAlexaDeviceControls,
				},
			},
			call: func(client *Client) error {
				return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{
					Endpoint: endpoint, CustomerID: "synthetic-customer", Delta: nil, SetVolume: false, Volume: &volume,
				})
			},
		},
		{
			name: "volume behavior with null value", node: alexamodels.DeviceControlsVolumeOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.DeviceControlsVolumeOperationTypeVolume,
				OperationPayload: alexamodels.DeviceControlsVolumePayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					Value: nil, SkillID: alexamodels.BehaviorSkillIDAlexaDeviceControls,
				},
			},
			call: func(client *Client) error {
				return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{
					Endpoint: endpoint, CustomerID: "synthetic-customer", Delta: nil, SetVolume: false, Volume: nil,
				})
			},
		},
		{
			name: "notification", node: alexamodels.NotificationsSendMobilePushOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.NotificationsSendMobilePushOperationTypeSendMobilePush,
				OperationPayload: alexamodels.NotificationsSendMobilePushPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					NotificationMessage: "Synthetic alert", Title: "Synthetic title",
					AlexaURL: alexamodels.BehaviorAlexaURL(alexamodels.AlexaURLBehaviors),
					SkillID:  alexamodels.BehaviorSkillIDRoutinesMessaging,
				},
			},
			call: func(client *Client) error {
				return client.SendNotification(context.Background(), &alexamodels.SendNotificationRequest{Endpoint: endpoint, Message: "Synthetic alert", Title: "Synthetic title", CustomerID: "request-customer"})
			},
		},
		{
			name: "announcement speak", node: alexamodels.AnnouncementOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.AnnouncementOperationTypeAlexaAnnouncement,
				OperationPayload: alexamodels.AnnouncementPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: "de-DE", CustomerID: "synthetic-client-customer",
					ExpireAfter: alexamodels.BehaviorAnnouncementExpireAfter(alexamodels.DefaultAnnouncementExpireAfter),
					Content:     announcementContent("de-DE", "", "", "Synthetic words"),
					Target:      sharedTarget, SkillID: alexamodels.BehaviorSkillIDAlexaNotifications,
				},
			},
			call: func(client *Client) error {
				return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{
					Endpoint: endpoint, CustomerID: "request-customer", Locale: "de-DE", Message: "Synthetic words",
					Method: alexamodels.AnnouncementMethodSpeak, TargetDevices: nil, Title: "",
				})
			},
		},
		{
			name: "announcement show", node: alexamodels.AnnouncementOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.AnnouncementOperationTypeAlexaAnnouncement,
				OperationPayload: alexamodels.AnnouncementPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: "fr-FR", CustomerID: "synthetic-client-customer", ExpireAfter: alexamodels.BehaviorAnnouncementExpireAfter(alexamodels.DefaultAnnouncementExpireAfter),
					Content: announcementContent("fr-FR", "Synthetic title", "Synthetic display", ""),
					Target:  sharedTarget, SkillID: alexamodels.BehaviorSkillIDAlexaNotifications,
				},
			},
			call: func(client *Client) error {
				return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{
					Endpoint: endpoint, CustomerID: "", Locale: "", Message: "Synthetic display",
					Method: alexamodels.AnnouncementMethodShow, TargetDevices: nil, Title: "Synthetic title",
				})
			},
		},
		{
			name: "announcement all", node: alexamodels.AnnouncementOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.AnnouncementOperationTypeAlexaAnnouncement,
				OperationPayload: alexamodels.AnnouncementPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: "fr-FR", CustomerID: "synthetic-client-customer", ExpireAfter: alexamodels.BehaviorAnnouncementExpireAfter(alexamodels.DefaultAnnouncementExpireAfter),
					Content: announcementContent("fr-FR", "", "Synthetic display and speech", "Synthetic display and speech"),
					Target:  sharedTarget, SkillID: alexamodels.BehaviorSkillIDAlexaNotifications,
				},
			},
			call: func(client *Client) error {
				return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{
					Endpoint: endpoint, CustomerID: "", Locale: "", Message: "Synthetic display and speech",
					Method: alexamodels.AnnouncementMethodAll, TargetDevices: nil, Title: "",
				})
			},
		},
		{
			name: "text to speech", node: alexamodels.SpeakOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.SpeakOperationTypeAlexaSpeak,
				OperationPayload: alexamodels.SpeakPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					TextToSpeak: "Synthetic spoken text", SkillID: alexamodels.BehaviorSkillIDSaySomething,
				},
			},
			call: func(client *Client) error {
				return client.SendTTS(context.Background(), &alexamodels.SendTTSRequest{
					Endpoint: endpoint, CustomerID: "", Message: "Synthetic spoken text", TargetDevices: nil,
				})
			},
		},
		{
			name: "play music with timer", node: alexamodels.MusicPlaySearchPhraseOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.MusicPlaySearchPhraseOperationTypePlaySearchPhrase,
				OperationPayload: alexamodels.MusicPlaySearchPhrasePayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					SearchPhrase: "Synthetic artist", SanitizedSearchPhrase: "Synthetic artist",
					MusicProviderID: "synthetic-provider", WaitTimeInSeconds: seconds,
				},
			},
			call: func(client *Client) error {
				return client.PlayMusic(context.Background(), &alexamodels.PlayMusicRequest{
					Endpoint: endpoint, CustomerID: "", ProviderID: "synthetic-provider", SearchPhrase: "Synthetic artist", TimerSeconds: &seconds,
				})
			},
		},
		{
			name: "play music omits nonpositive timer", node: alexamodels.MusicPlaySearchPhraseOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.MusicPlaySearchPhraseOperationTypePlaySearchPhrase,
				OperationPayload: alexamodels.MusicPlaySearchPhrasePayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					SearchPhrase: "Synthetic artist", SanitizedSearchPhrase: "Synthetic artist", MusicProviderID: "synthetic-provider",
					WaitTimeInSeconds: 0,
				},
			},
			call: func(client *Client) error {
				zero := 0

				return client.PlayMusic(context.Background(), &alexamodels.PlayMusicRequest{
					Endpoint: endpoint, CustomerID: "", ProviderID: "synthetic-provider", SearchPhrase: "Synthetic artist", TimerSeconds: &zero,
				})
			},
		},
		{
			name: "play audio uri", node: alexamodels.SoundOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.SoundOperationTypeAlexaSound,
				OperationPayload: alexamodels.SoundPayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					SoundStringID: "https://media.example.invalid/synthetic.mp3",
				},
			},
			call: func(client *Client) error {
				return client.PlayAudioURI(context.Background(), &alexamodels.PlayAudioURIRequest{
					Endpoint: endpoint, CustomerID: "", URI: "https://media.example.invalid/synthetic.mp3",
				})
			},
		},
		{
			name: "play video with provider", node: alexamodels.VideoPlaySearchPhraseOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.VideoPlaySearchPhraseOperationTypePlaySearchPhrase,
				OperationPayload: alexamodels.VideoPlaySearchPhrasePayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					SearchPhrase: "Synthetic title on Synthetic Video", SanitizedSearchPhrase: "Synthetic title on Synthetic Video",
					WaitTimeInSeconds: seconds,
				},
			},
			call: func(client *Client) error {
				return client.PlayVideo(context.Background(), &alexamodels.PlayVideoRequest{
					Endpoint: endpoint, CustomerID: "", SearchPhrase: "Synthetic title", TimerSeconds: &seconds, VideoProviderID: "Synthetic Video",
				})
			},
		},
		{
			name: "play video without provider", node: alexamodels.VideoPlaySearchPhraseOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.VideoPlaySearchPhraseOperationTypePlaySearchPhrase,
				OperationPayload: alexamodels.VideoPlaySearchPhrasePayload{
					DeviceType: "synthetic-device-type", DeviceSerialNumber: "synthetic-device-serial",
					Locale: alexamodels.DefaultLocale, CustomerID: "synthetic-client-customer",
					SearchPhrase: "Synthetic title", SanitizedSearchPhrase: "Synthetic title",
					WaitTimeInSeconds: 0,
				},
			},
			call: func(client *Client) error {
				return client.PlayVideo(context.Background(), &alexamodels.PlayVideoRequest{
					Endpoint: endpoint, CustomerID: "", SearchPhrase: "Synthetic title", TimerSeconds: nil, VideoProviderID: "",
				})
			},
		},
		{
			name: "fire tv sequence", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVPauseVideo,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-client-customer", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.SendFireTVSequence(context.Background(), "synthetic-account", alexamodels.OperationTypeFireTVPauseVideo)
			},
		},
		{
			name: "fire tv sequence preserves arbitrary operation type", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: "synthetic.CustomFireTVOperation",
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-client-customer", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.SendFireTVSequence(context.Background(), "synthetic-account", "synthetic.CustomFireTVOperation")
			},
		},
		{
			name: "fire tv on", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVTurnOn,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVTurnOn(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv off", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVTurnOff,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVTurnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv turn on/off true", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVTurnOn,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVTurnOnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint}, true)
			},
		},
		{
			name: "fire tv turn on/off false", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVTurnOff,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVTurnOnOff(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint}, false)
			},
		},
		{
			name: "fire tv pause", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVPauseVideo,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVPauseVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv resume", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVResumeVideo,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVResumeVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
		{
			name: "fire tv home", node: alexamodels.FireTVOperationNode{
				Type: behaviorNodeType, OperationType: alexamodels.OperationTypeFireTVNavigateHome,
				OperationPayload: alexamodels.FireTVOperationPayload{DeviceAccountID: "synthetic-account", CustomerID: "synthetic-device-owner", SkillID: alexamodels.BehaviorSkillIDRoutinesFireTV},
			},
			call: func(client *Client) error {
				return client.FireTVNavigateHome(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			requestCount := 0

			client := NewClient(
				WithAlexaAmazonBaseURI("https://alexa.synthetic.test"),
				WithBearerToken("synthetic-behavior-token"),
				WithCustomerID("synthetic-client-customer"),
				WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(request *http.Request) (*http.Response, error) {
					requestCount++
					assertBehaviorSequenceRequest(t, request, testCase.node)

					return syntheticResponse(request, http.StatusOK, `{}`), nil
				})}),
			)

			err := testCase.call(client)
			if err != nil {
				t.Fatalf("behavior operation returned an error: %v", err)
			}

			if requestCount != 1 {
				t.Fatalf("behavior operation made %d requests, want exactly one", requestCount)
			}
		})
	}
}

func assertBehaviorSequenceRequest(t *testing.T, request *http.Request, expectedStartNode any) {
	t.Helper()

	if request.Method != http.MethodPost || request.URL.Path != "/api/behaviors/preview" ||
		request.Header.Get("Authorization") != "Bearer synthetic-behavior-token" {
		t.Errorf("unexpected behavior request: %s %s auth=%q", request.Method, request.URL, request.Header.Get("Authorization"))
	}

	var body map[string]any

	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		t.Fatalf("decode behavior preview request: %v", err)
	}

	sequenceJSON, ok := body["sequenceJson"].(string)
	if !ok {
		t.Fatalf("behavior preview sequenceJson has type %T, want string", body["sequenceJson"])
	}

	expectedSequence, err := json.Marshal(alexamodels.Sequence{
		Type:      alexamodels.SequenceType(alexamodels.ModelTypeSequence),
		StartNode: expectedStartNode,
	})
	if err != nil {
		t.Fatalf("marshal expected behavior sequence: %v", err)
	}

	if sequenceJSON != string(expectedSequence) {
		t.Errorf("unexpected behavior sequence bytes:\n got: %s\nwant: %s", sequenceJSON, expectedSequence)
	}

	expectedRequest := alexamodels.WireBehaviorPreviewRequest{
		BehaviorId:   alexamodels.DefaultBehaviorID,
		SequenceJson: string(expectedSequence),
		Status:       alexamodels.DefaultBehaviorStatus,
	}
	if !sameRestJSON(body, expectedRequest) {
		t.Errorf("behavior preview request got %#v, want %#v", body, expectedRequest)
	}
}

func TestBehaviorMethodsValidateEndpointAndVolumeInputs(t *testing.T) {
	t.Parallel()

	client := NewClient()

	invalid := []struct {
		name string
		call func() error
	}{
		{name: "stop playback endpoint", call: func() error {
			return client.StopPlayback(context.Background(), &alexamodels.StopPlaybackRequest{
				Endpoint: nil, AllDevices: false, CustomerID: "",
			})
		}},
		{name: "set volume endpoint", call: func() error {
			return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{
				Endpoint: nil, CustomerID: "", Delta: nil, SetVolume: false, Volume: nil,
			})
		}},
		{name: "negative volume", call: func() error {
			value := -1

			return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{
				Endpoint: syntheticEndpoint{}, CustomerID: "", Delta: nil, SetVolume: false, Volume: &value,
			})
		}},
		{name: "volume over maximum", call: func() error {
			value := 101

			return client.SetVolume(context.Background(), &alexamodels.VolumeControlRequest{
				Endpoint: syntheticEndpoint{}, CustomerID: "", Delta: nil, SetVolume: false, Volume: &value,
			})
		}},
		{name: "notification endpoint", call: func() error {
			return client.SendNotification(context.Background(), &alexamodels.SendNotificationRequest{
				Endpoint: nil, CustomerID: "", Message: "", Title: "",
			})
		}},
		{name: "announcement endpoint", call: func() error {
			return client.SendAnnouncement(context.Background(), &alexamodels.SendAnnouncementRequest{
				Endpoint: nil, CustomerID: "", Locale: "", Message: "", Method: "", TargetDevices: nil, Title: "",
			})
		}},
		{name: "tts endpoint", call: func() error {
			return client.SendTTS(context.Background(), &alexamodels.SendTTSRequest{
				Endpoint: nil, CustomerID: "", Message: "", TargetDevices: nil,
			})
		}},
		{name: "music endpoint", call: func() error {
			return client.PlayMusic(context.Background(), &alexamodels.PlayMusicRequest{
				Endpoint: nil, CustomerID: "", ProviderID: "", SearchPhrase: "", TimerSeconds: nil,
			})
		}},
		{name: "audio endpoint", call: func() error {
			return client.PlayAudioURI(context.Background(), &alexamodels.PlayAudioURIRequest{
				Endpoint: nil, CustomerID: "", URI: "",
			})
		}},
		{name: "video endpoint", call: func() error {
			return client.PlayVideo(context.Background(), &alexamodels.PlayVideoRequest{
				Endpoint: nil, CustomerID: "", SearchPhrase: "", TimerSeconds: nil, VideoProviderID: "",
			})
		}},
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

func TestFireTVMissingOwnerRejectsRequestWithoutAccountFallback(t *testing.T) {
	t.Parallel()

	requests := 0
	client := NewClient(
		WithHTTPClient(&http.Client{Transport: syntheticRoundTripper(func(_ *http.Request) (*http.Response, error) {
			requests++

			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}, nil
		})}),
		WithCustomerID("synthetic-account-customer"),
	)
	endpoint := syntheticEndpoint{family: alexamodels.DeviceFamilyFireTV, account: "synthetic-device-account"}

	err := client.FireTVPauseVideo(context.Background(), &alexamodels.FireTVRequest{Endpoint: endpoint})
	if err == nil || !strings.Contains(err.Error(), "device owner customer ID is required") {
		t.Fatalf("expected missing-owner rejection, got %v", err)
	}

	if requests != 0 {
		t.Fatalf("missing owner made %d network requests", requests)
	}
}
