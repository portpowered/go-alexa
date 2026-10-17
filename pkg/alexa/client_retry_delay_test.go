package alexa_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestWithRetryDelayReachesClientRESTRequests(t *testing.T) {
	t.Parallel()

	var attempts, delays atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	client, err := alexa.NewClient(
		alexa.WithAlexaAPIBaseURL(server.URL),
		alexa.WithAmazonAPIBaseURL(server.URL),
		alexa.WithAlexaWebBaseURL(server.URL),
		alexa.WithRESTHTTPClient(server.Client()),
		alexa.WithRetryDelay(func(int) time.Duration {
			delays.Add(1)

			return 0
		}),
	)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	started := time.Now()

	_, err = client.RefreshAccessToken(context.Background(), alexaapimodels.TokenRefreshRequest{
		RefreshToken: "synthetic-refresh-token",
		Config: alexaapimodels.DeviceRegistrationConfig{
			AppName: "", AppVersion: "", DeviceType: "", Domain: "", DeviceModel: "",
			OSVersion: "", DeviceSerial: "", DeviceName: "", Manufacturer: "",
		},
	})
	if err == nil {
		t.Fatal("expected the 5xx response to fail")
	}

	if attempts.Load() != 3 || delays.Load() != 2 {
		t.Fatalf("attempts=%d delays=%d, want 3 attempts and 2 delays", attempts.Load(), delays.Load())
	}

	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("zero retry delay still waited %v", elapsed)
	}
}
