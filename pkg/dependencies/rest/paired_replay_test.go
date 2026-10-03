//nolint:testpackage // Exercises private REST transport state with paired synthetic replay.
package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

// These fixed inputs and paired outputs are synthetic. UPDATE_SYNTHETIC_REPLAY=1
// refreshes the checked-in expectations after an intentional wire change.
type restReplayResponse struct {
	operation, body string
	status          int
	headers         http.Header
	call            func(*Client) error
}

func restPairedReplayResponses(
	ctx context.Context,
	config *DeviceRegistrationConfig,
	endpoint syntheticEndpoint,
	media *alexamodels.MediaControlRequest,
) []restReplayResponse {
	responses := make([]restReplayResponse, 0, 26)
	responses = append(responses, restAuthenticationReplayCases(ctx, config)...)
	responses = append(responses, restEndpointReplayCases(ctx, endpoint)...)
	responses = append(responses, restAccountReplayCases(ctx)...)
	responses = append(responses, restMediaReplayCases(ctx, endpoint, media)...)

	return responses
}

func restAuthenticationReplayCases(
	ctx context.Context,
	config *DeviceRegistrationConfig,
) []restReplayResponse {
	return []restReplayResponse{
		{"createCodePair", "{}", 200, nil, func(c *Client) error {
			_, err := c.GenerateCodePair(ctx, config)

			return err
		}},
		{"registerDevice_email", "{}", 200, nil, func(c *Client) error {
			_, err := c.RegisterWithEmailPassword(ctx, "synthetic@example.invalid", "synthetic-password", config)

			return err
		}},
		{"registerDevice_codePair", "{}", 200, nil, func(c *Client) error {
			_, err := c.RegisterWithCodePair(ctx, "synthetic-public", "synthetic-private", config)

			return err
		}},
		{"refreshAccessToken", `{"access_token":"synthetic-access","expires_in":3600}`, 200, nil, func(c *Client) error {
			_, err := c.RefreshAccessToken(ctx, "synthetic-refresh", config)

			return err
		}},
		{"exchangeRefreshTokenForCookies", `{"response":{"tokens":{"cookies":{}}}}`, 200, nil, func(c *Client) error {
			_, err := c.ExchangeRefreshTokenForCookies(ctx, "synthetic-refresh", "synthetic.amazon.test")

			return err
		}},
	}
}

func restEndpointReplayCases(
	ctx context.Context,
	endpoint syntheticEndpoint,
) []restReplayResponse {
	return []restReplayResponse{
		{"listRestEndpoints", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetEndpoints(ctx, &ListEndpointsOptions{Owner: "~caller", Expand: []string{"all", "feature:power"}})

			return err
		}},
		{"getRestEndpoint", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetEndpointByID(ctx, "synthetic-endpoint", []string{"all", "feature:power"})

			return err
		}},
		{"queryRestEndpoints", "{}", 200, nil, func(c *Client) error {
			_, err := c.QueryEndpoints(ctx, &alexamodels.EndpointQueryRequest{
				Query: alexamodels.EndpointQuery{AND: nil, OR: nil, IncludeFields: nil, PaginationContext: nil},
			}, nil)

			return err
		}},
		{"forgetEndpoint", "{}", 200, nil, func(c *Client) error { return c.ForgetEndpoint(ctx, "synthetic-endpoint") }},
		{"deregisterEndpoint", "{}", 200, nil, func(c *Client) error { return c.DeregisterEndpoint(ctx, "synthetic-endpoint") }},
		{"updateEndpointFriendlyName", "{}", 200, nil, func(c *Client) error { return c.UpdateFriendlyName(ctx, "synthetic-endpoint", "Synthetic light") }},
		{"controlRestEndpoint", "{}", 200, nil, func(c *Client) error {
			return c.ControlEndpoint(
				ctx,
				"synthetic-endpoint",
				alexamodels.Command{
					DeviceID: "", Type: "",
					Namespace: "Alexa.PowerController",
					Name:      "TurnOn",
					Payload:   map[string]any{"powerState": "ON"},
				},
			)
		}},
		{"sendEndpointInterfaceMessage", "{}", 200, nil, func(c *Client) error {
			return c.SendInterfaceMessage(ctx, &alexamodels.InterfaceMessageRequest{
				Endpoint:      endpoint,
				FeatureName:   "power",
				OperationName: "turnOn",
				Payload:       map[string]any{"state": "ON"},
			})
		}},
	}
}

func restAccountReplayCases(
	ctx context.Context,
) []restReplayResponse {
	return []restReplayResponse{
		{"listFirstPartyDevices", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetDevicesV2(ctx, &GetDevicesV2Options{CSRFToken: "synthetic-csrf"})

			return err
		}},
		{"getUserInfo", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetUserInfo(ctx, &GetUserInfoOptions{Platform: "synthetic", Version: "1", CSRFToken: "synthetic-csrf"})

			return err
		}},
		{"fetchCsrfCookie", "{}", 200, http.Header{"Set-Cookie": []string{"csrf=synthetic-csrf; Path=/"}}, func(c *Client) error {
			_, err := c.fetchCSRFTokenFromAPI(ctx)

			return err
		}},
		{"submitBehaviorPreview", "{}", 200, nil, func(c *Client) error {
			return c.SendSequence(ctx, "synthetic.Operation", map[string]any{"deviceType": "synthetic-device"})
		}},
	}
}

func restMediaReplayCases(
	ctx context.Context,
	endpoint syntheticEndpoint,
	media *alexamodels.MediaControlRequest,
) []restReplayResponse {
	return []restReplayResponse{
		{"sendMediaCommand_pause", "{}", 200, nil, func(c *Client) error { return c.PausePlayback(ctx, media) }},
		{"sendMediaCommand_resume", "{}", 200, nil, func(c *Client) error { return c.ResumePlayback(ctx, media) }},
		{"sendMediaCommand_next", "{}", 200, nil, func(c *Client) error { return c.NextTrack(ctx, media) }},
		{"sendMediaCommand_previous", "{}", 200, nil, func(c *Client) error { return c.PreviousTrack(ctx, media) }},
		{"sendMediaCommand_forward", "{}", 200, nil, func(c *Client) error { return c.ForwardMedia(ctx, media) }},
		{"sendMediaCommand_rewind", "{}", 200, nil, func(c *Client) error { return c.RewindMedia(ctx, media) }},
		{"sendMediaCommand_shuffle", "{}", 200, nil, func(c *Client) error { return c.SetShuffle(ctx, media, true) }},
		{"sendMediaCommand_repeat", "{}", 200, nil, func(c *Client) error { return c.SetRepeat(ctx, media, true) }},
		{"getMediaPlayerState", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetPlayerState(ctx, &alexamodels.PlayerStateRequest{Endpoint: endpoint})

			return err
		}},
	}
}

type restReplayCapture struct {
	exchanges []replay.SyntheticExchange
}

func TestRESTPairedSyntheticReplay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := &DeviceRegistrationConfig{
		AppName:      "synthetic-app",
		AppVersion:   "1",
		DeviceType:   "synthetic-device",
		DeviceSerial: "synthetic-serial",
		Manufacturer: "Synthetic",
	}
	endpoint := syntheticEndpoint{
		id:      "synthetic-endpoint",
		device:  "synthetic-device",
		serial:  "synthetic-serial",
		locale:  "en-US",
		family:  alexamodels.DeviceFamilyFireTV,
		account: "synthetic-account",
	}
	media := &alexamodels.MediaControlRequest{Endpoint: endpoint}
	responses := restPairedReplayResponses(ctx, config, endpoint, media)

	fixture := filepath.Join("..", "..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-rest.json")
	transport, player, capture := restReplayTransport(t, fixture, responses)
	client := restReplayClient(transport)

	for _, test := range responses {
		err := test.call(client)
		if err != nil {
			t.Fatalf("%s: %v", test.operation, err)
		}
	}

	if player != nil {
		assertRESTReplayConsumed(ctx, t, player, client)

		return
	}

	writeRESTReplayFixture(t, fixture, responses, capture.exchanges)
}

func restReplayTransport(
	t *testing.T,
	fixture string,
	responses []restReplayResponse,
) (http.RoundTripper, *replay.SyntheticReplay, *restReplayCapture) {
	t.Helper()

	if os.Getenv("UPDATE_SYNTHETIC_REPLAY") == "1" {
		capture := &restReplayCapture{exchanges: make([]replay.SyntheticExchange, 0, len(responses))}

		return restRecordingTransport(responses, capture), nil, capture
	}

	player, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	return player, player, nil
}

func restRecordingTransport(responses []restReplayResponse, capture *restReplayCapture) http.RoundTripper {
	return syntheticRoundTripper(func(req *http.Request) (*http.Response, error) {
		if len(capture.exchanges) >= len(responses) {
			return nil, fmt.Errorf("unexpected request after %d operations", len(capture.exchanges)) //nolint:err113 // Fixture-only overflow diagnostic.
		}

		expected := responses[len(capture.exchanges)]

		responseHeaders := expected.headers.Clone()
		if responseHeaders == nil {
			responseHeaders = make(http.Header)
		}

		responseHeaders.Set("Content-Type", "application/json")

		pair, err := replay.RecordSyntheticExchange(expected.operation, req, expected.status, responseHeaders, expected.body)
		if err != nil {
			return nil, fmt.Errorf("record synthetic REST exchange: %w", err)
		}

		capture.exchanges = append(capture.exchanges, pair)
		response := syntheticResponse(req, expected.status, expected.body)
		response.Header = responseHeaders

		return response, nil
	})
}

func restReplayClient(transport http.RoundTripper) *Client {
	clientOptions := []ClientOption{
		WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"),
		WithAmazonapiBaseURI("https://auth.synthetic.test"),
		WithAlexaAmazonBaseURI("https://web.synthetic.test"),
		WithBearerToken("synthetic-bearer"),
		WithCSRFToken("synthetic-csrf"),
		WithCookies(map[string]*http.Cookie{
			"web.synthetic.test:session-id": {Name: "session-id", Value: "synthetic-session"},
		}),
		WithHTTPClient(&http.Client{Transport: transport}),
	}

	return NewClient(clientOptions...)
}

func assertRESTReplayConsumed(ctx context.Context, t *testing.T, player *replay.SyntheticReplay, client *Client) {
	t.Helper()

	err := player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetEndpoints(ctx, nil)
	if err == nil {
		t.Fatal("duplicate REST request received a replay response")
	}
}

func writeRESTReplayFixture(
	t *testing.T,
	fixture string,
	responses []restReplayResponse,
	exchanges []replay.SyntheticExchange,
) {
	t.Helper()

	if len(exchanges) != len(responses) {
		t.Fatalf("recorded %d of %d", len(exchanges), len(responses))
	}

	err := os.MkdirAll(filepath.Dir(fixture), 0o750)
	if err != nil {
		t.Fatal(err)
	}

	err = replay.WriteSyntheticReplay(fixture, exchanges)
	if err != nil {
		t.Fatal(err)
	}
}
