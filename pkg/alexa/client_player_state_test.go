package alexa

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type playerStateTransport struct {
	t *testing.T
}

func (transport playerStateTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.t.Helper()
	if request.Method != http.MethodGet {
		transport.t.Fatalf("method = %s, want GET", request.Method)
	}
	if request.URL.Path != "/api/np/player" {
		transport.t.Fatalf("path = %s, want /api/np/player", request.URL.Path)
	}
	query := request.URL.Query()
	if got := query.Get("deviceType"); got != "synthetic-device-type" {
		transport.t.Fatalf("device type = %q", got)
	}
	if got := query.Get("deviceSerialNumber"); got != "synthetic-device-serial" {
		transport.t.Fatalf("device serial = %q", got)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: io.NopCloser(strings.NewReader(
			`{"playerInfo":{"state":"PLAYING","provider":{"providerName":"Synthetic source"},"progress":{"mediaLength":120,"mediaProgress":45}}}`,
		)),
		Request: request,
	}, nil
}

func TestGetPlayerStateUsesEndpointIdentityAndConvertsResponse(t *testing.T) {
	client, err := NewClient(
		WithHttpClient(&http.Client{Transport: playerStateTransport{t: t}}),
		WithBearerToken("synthetic-token"),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	response, err := client.GetPlayerState(context.Background(), alexaapimodels.PlayerStateRequest{
		Target: &alexaapimodels.Endpoint{
			DeviceType:         "synthetic-device-type",
			DeviceSerialNumber: "synthetic-device-serial",
		},
	})
	if err != nil {
		t.Fatalf("get player state: %v", err)
	}
	if response.PlayerInfo == nil {
		t.Fatal("player info is nil")
	}
	if response.PlayerInfo.State != "PLAYING" {
		t.Fatalf("state = %q, want PLAYING", response.PlayerInfo.State)
	}
	if response.PlayerInfo.Provider == nil || response.PlayerInfo.Provider.ProviderName != "Synthetic source" {
		t.Fatalf("provider = %#v", response.PlayerInfo.Provider)
	}
	if response.PlayerInfo.Progress == nil || response.PlayerInfo.Progress.MediaProgress != 45 {
		t.Fatalf("progress = %#v", response.PlayerInfo.Progress)
	}
}
