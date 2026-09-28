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
func TestRESTPairedSyntheticReplay(t *testing.T) {
	ctx := context.Background()
	config := &DeviceRegistrationConfig{AppName: "synthetic-app", AppVersion: "1", DeviceType: "synthetic-device", DeviceSerial: "synthetic-serial", Manufacturer: "Synthetic"}
	endpoint := syntheticEndpoint{id: "synthetic-endpoint", device: "synthetic-device", serial: "synthetic-serial", locale: "en-US", family: alexamodels.DeviceFamilyFireTV, account: "synthetic-account"}
	media := &alexamodels.MediaControlRequest{Endpoint: endpoint}
	responses := []struct {
		operation, body string
		status          int
		headers         http.Header
		call            func(*Client) error
	}{
		{"createCodePair", "{}", 200, nil, func(c *Client) error { _, err := c.GenerateCodePair(ctx, config); return err }},
		{"registerDevice_email", "{}", 200, nil, func(c *Client) error {
			_, err := c.RegisterWithEmailPassword(ctx, "synthetic@example.invalid", "synthetic-password", config)
			return err
		}},
		{"registerDevice_codePair", "{}", 200, nil, func(c *Client) error {
			_, err := c.RegisterWithCodePair(ctx, "synthetic-public", "synthetic-private", config)
			return err
		}},
		{"refreshAccessToken", `{"access_token":"synthetic-access","expires_in":3600}`, 200, nil, func(c *Client) error { _, err := c.RefreshAccessToken(ctx, "synthetic-refresh", config); return err }},
		{"exchangeRefreshTokenForCookies", `{"response":{"tokens":{"cookies":{}}}}`, 200, nil, func(c *Client) error {
			_, err := c.ExchangeRefreshTokenForCookies(ctx, "synthetic-refresh", "synthetic.amazon.test")
			return err
		}},
		{"listRestEndpoints", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetEndpoints(ctx, &ListEndpointsOptions{Owner: "~caller", Expand: []string{"all", "feature:power"}})
			return err
		}},
		{"getRestEndpoint", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetEndpointByID(ctx, "synthetic-endpoint", []string{"all", "feature:power"})
			return err
		}},
		{"queryRestEndpoints", "{}", 200, nil, func(c *Client) error {
			_, err := c.QueryEndpoints(ctx, &alexamodels.EndpointQueryRequest{}, nil)
			return err
		}},
		{"forgetEndpoint", "{}", 200, nil, func(c *Client) error { return c.ForgetEndpoint(ctx, "synthetic-endpoint") }},
		{"deregisterEndpoint", "{}", 200, nil, func(c *Client) error { return c.DeregisterEndpoint(ctx, "synthetic-endpoint") }},
		{"updateEndpointFriendlyName", "{}", 200, nil, func(c *Client) error { return c.UpdateFriendlyName(ctx, "synthetic-endpoint", "Synthetic light") }},
		{"controlRestEndpoint", "{}", 200, nil, func(c *Client) error {
			return c.ControlEndpoint(ctx, "synthetic-endpoint", alexamodels.Command{Namespace: "Alexa.PowerController", Name: "TurnOn", Payload: map[string]any{"powerState": "ON"}})
		}},
		{"sendEndpointInterfaceMessage", "{}", 200, nil, func(c *Client) error {
			return c.SendInterfaceMessage(ctx, &alexamodels.InterfaceMessageRequest{Endpoint: endpoint, FeatureName: "power", OperationName: "turnOn", Payload: map[string]any{"state": "ON"}})
		}},
		{"listFirstPartyDevices", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetDevicesV2(ctx, &GetDevicesV2Options{CSRFToken: "synthetic-csrf"})
			return err
		}},
		{"getUserInfo", "{}", 200, nil, func(c *Client) error {
			_, err := c.GetUserInfo(ctx, &GetUserInfoOptions{Platform: "synthetic", Version: "1", CSRFToken: "synthetic-csrf"})
			return err
		}},
		{"fetchCsrfCookie", "{}", 200, http.Header{"Set-Cookie": []string{"csrf=synthetic-csrf; Path=/"}}, func(c *Client) error { _, err := c.fetchCSRFTokenFromAPI(ctx); return err }},
		{"submitBehaviorPreview", "{}", 200, nil, func(c *Client) error {
			return c.SendSequence(ctx, "synthetic.Operation", map[string]any{"deviceType": "synthetic-device"})
		}},
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
	fixture := filepath.Join("..", "..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-rest.json")
	var transport http.RoundTripper
	var recorded []replay.SyntheticExchange
	var player *replay.SyntheticReplay
	if os.Getenv("UPDATE_SYNTHETIC_REPLAY") == "1" {
		transport = syntheticRoundTripper(func(req *http.Request) (*http.Response, error) {
			if len(recorded) >= len(responses) {
				return nil, fmt.Errorf("unexpected request after %d operations", len(recorded))
			}
			expected := responses[len(recorded)]
			responseHeaders := expected.headers.Clone()
			if responseHeaders == nil {
				responseHeaders = make(http.Header)
			}
			responseHeaders.Set("Content-Type", "application/json")
			pair, err := replay.RecordSyntheticExchange(expected.operation, req, expected.status, responseHeaders, expected.body)
			if err != nil {
				return nil, err
			}
			recorded = append(recorded, pair)
			response := syntheticResponse(req, expected.status, expected.body)
			response.Header = responseHeaders
			return response, nil
		})
	} else {
		var err error
		player, err = replay.LoadSyntheticReplay(fixture)
		if err != nil {
			t.Fatal(err)
		}
		transport = player
	}
	client := NewClient(WithAmazonalexaAPIBaseURI("https://alexa.synthetic.test"), WithAmazonapiBaseURI("https://auth.synthetic.test"), WithAlexaAmazonBaseURI("https://web.synthetic.test"), WithBearerToken("synthetic-bearer"), WithCSRFToken("synthetic-csrf"), WithCookies(map[string]*http.Cookie{"web.synthetic.test:session-id": {Name: "session-id", Value: "synthetic-session"}}), WithHTTPClient(&http.Client{Transport: transport}))
	for _, test := range responses {
		if err := test.call(client); err != nil {
			t.Fatalf("%s: %v", test.operation, err)
		}
	}
	if player != nil {
		if err := player.AssertConsumed(); err != nil {
			t.Fatal(err)
		}
		if _, err := client.GetEndpoints(ctx, nil); err == nil {
			t.Fatal("duplicate REST request received a replay response")
		}
	} else {
		if len(recorded) != len(responses) {
			t.Fatalf("recorded %d of %d", len(recorded), len(responses))
		}
		if err := os.MkdirAll(filepath.Dir(fixture), 0755); err != nil {
			t.Fatal(err)
		}
		if err := replay.WriteSyntheticReplay(fixture, recorded); err != nil {
			t.Fatal(err)
		}
	}
}
