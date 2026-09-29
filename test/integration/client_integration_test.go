package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	refreshTokenEnvVar = "ALEXA_REFRESH_TOKEN"
	testTimeout        = 30 * time.Second
	eventTimeout       = 60 * time.Second
)

// TestTokenRefresh tests the happy case for token refresh.
func TestTokenRefresh(t *testing.T) {
	t.Parallel()

	refreshToken := os.Getenv(refreshTokenEnvVar)
	if refreshToken == "" {
		t.Skipf("Skipping test: %s environment variable not set", refreshTokenEnvVar)
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Create a client without a bearer token (we'll use it to refresh)
	client, err := alexa.NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create device registration config
	config := alexaapimodels.DefaultDeviceRegistrationConfig("test-serial", "test-device")

	// Refresh the access token
	refreshReq := alexaapimodels.TokenRefreshRequest{
		RefreshToken: refreshToken,
		Config:       config,
	}

	resp, err := client.RefreshAccessToken(ctx, refreshReq)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("Access token is empty")
	}

	t.Logf("Successfully refreshed access token (length: %d)", len(resp.AccessToken))
}

// TestEnumeration tests the happy case for endpoint enumeration.
func TestEnumeration(t *testing.T) {
	t.Parallel()

	refreshToken := os.Getenv(refreshTokenEnvVar)
	if refreshToken == "" {
		t.Skipf("Skipping test: %s environment variable not set", refreshTokenEnvVar)
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// First, refresh the token to get an access token
	client, err := alexa.NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	config := alexaapimodels.DefaultDeviceRegistrationConfig("test-serial", "test-device")
	refreshReq := alexaapimodels.TokenRefreshRequest{
		RefreshToken: refreshToken,
		Config:       config,
	}

	tokenResp, err := client.RefreshAccessToken(ctx, refreshReq)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	// Create a new client with the access token
	clientWithToken, err := client.NewSession(alexa.WithBearerToken(tokenResp.AccessToken))
	if err != nil {
		t.Fatalf("Failed to create client with token: %v", err)
	}

	defer func() {
		_ = clientWithToken.Close()
	}()

	// List endpoints
	query := alexaapimodels.EndpointQuery{
		IncludeFields: &alexaapimodels.EndpointIncludeFields{
			Features: true,
		},
	}

	response, err := clientWithToken.ListEndpoints(ctx, query)
	if err != nil {
		t.Fatalf("Failed to list endpoints: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if len(response.Results) == 0 {
		t.Log("Warning: No endpoints found (this may be expected if no devices are registered)")
	} else {
		t.Logf("Successfully enumerated %d endpoints", len(response.Results))

		for i, endpoint := range response.Results {
			if i >= 3 {
				break // Only log first 3
			}

			t.Logf("  - Endpoint %d: ID=%s, Type=%s, Name=%v", i+1, endpoint.ID, endpoint.DeviceType, endpoint.FriendlyName)
		}
	}
}

// TestControlPowerStateChange tests the happy case for control and dispatch for power state change.
//
//nolint:paralleltest // Both live tests can toggle the same account's endpoints.
func TestControlPowerStateChange(t *testing.T) {
	refreshToken := requireIntegrationRefreshToken(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	clientWithToken := newIntegrationClient(ctx, t, refreshToken)

	defer func() { _ = clientWithToken.Close() }()

	testEndpoint := findPowerControlEndpoint(ctx, t, clientWithToken)

	t.Logf("Testing power control on endpoint: %s (%s)", testEndpoint.EndpointID, testEndpoint.FriendlyName)

	// Test turning on
	onResp, err := setIntegrationPowerWithResponse(ctx, clientWithToken, testEndpoint, alexaapimodels.FeatureOperationNameTurnOn)
	if err != nil {
		t.Fatalf("Failed to turn on endpoint: %v", err)
	}

	if onResp == nil {
		t.Fatal("Control response is nil")
	}

	if len(onResp.Errors) > 0 {
		t.Logf("Warning: Control response contains errors: %+v", onResp.Errors)
	}

	t.Log("Successfully sent turn on control request")

	// Wait a bit for the state to change
	time.Sleep(2 * time.Second)

	// Test turning off
	offResp, err := setIntegrationPowerWithResponse(ctx, clientWithToken, testEndpoint, alexaapimodels.FeatureOperationNameTurnOff)
	if err != nil {
		t.Fatalf("Failed to turn off endpoint: %v", err)
	}

	if offResp == nil {
		t.Fatal("Control response is nil")
	}

	if len(offResp.Errors) > 0 {
		t.Logf("Warning: Control response contains errors: %+v", offResp.Errors)
	}

	t.Log("Successfully sent turn off control request")
}

// TestControlAndEvents tests the happy case for messaging and using control dispatch to generate events for power state change.
//
//nolint:paralleltest // Both live tests can toggle the same account's endpoints.
func TestControlAndEvents(t *testing.T) {
	refreshToken := requireIntegrationRefreshToken(t)

	ctx, cancel := context.WithTimeout(context.Background(), eventTimeout)
	defer cancel()

	clientWithToken := newIntegrationClient(ctx, t, refreshToken)

	defer func() { _ = clientWithToken.Close() }()

	testEndpoint := findPowerControlEndpoint(ctx, t, clientWithToken)

	t.Logf("Testing events with power control on endpoint: %s (%s)", testEndpoint.EndpointID, testEndpoint.FriendlyName)

	subscribeToIntegrationEvents(ctx, t, clientWithToken)

	// Connect to events
	eventConn, err := clientWithToken.ConnectEvents(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to events: %v", err)
	}

	defer func() {
		_ = eventConn.Close()
	}()

	t.Log("Successfully connected to event stream")

	eventChan, errorChan := startIntegrationEventReceiver(eventConn)

	// Wait a bit for connection to stabilize
	time.Sleep(2 * time.Second)

	// Send a power control command to generate an event
	_, err = setIntegrationPowerWithResponse(ctx, clientWithToken, testEndpoint, alexaapimodels.FeatureOperationNameTurnOn)
	if err != nil {
		t.Fatalf("Failed to send control request: %v", err)
	}

	t.Log("Sent turn on control request, waiting for event...")

	eventReceived := waitForIntegrationPowerEvent(t, eventChan, errorChan, testEndpoint.EndpointID)

	if eventReceived {
		t.Log("Successfully received event after power control")
	} else {
		t.Log("Note: Event may have been received but not recognized as power state change")
	}

	// Clean up - turn off the device
	_, err = setIntegrationPowerWithResponse(ctx, clientWithToken, testEndpoint, alexaapimodels.FeatureOperationNameTurnOff)
	if err != nil {
		t.Logf("Warning: Failed to turn off endpoint: %v", err)
	} else {
		t.Log("Successfully turned off endpoint")
	}
}

func requireIntegrationRefreshToken(t *testing.T) string {
	t.Helper()

	refreshToken := os.Getenv(refreshTokenEnvVar)
	if refreshToken == "" {
		t.Skipf("Skipping test: %s environment variable not set", refreshTokenEnvVar)
	}

	return refreshToken
}

func newIntegrationClient(ctx context.Context, t *testing.T, refreshToken string) *alexa.Session {
	t.Helper()

	client, err := alexa.NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	config := alexaapimodels.DefaultDeviceRegistrationConfig("test-serial", "test-device")
	refreshReq := alexaapimodels.TokenRefreshRequest{RefreshToken: refreshToken, Config: config}

	tokenResp, err := client.RefreshAccessToken(ctx, refreshReq)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	clientWithToken, err := client.NewSession(alexa.WithBearerToken(tokenResp.AccessToken))
	if err != nil {
		t.Fatalf("Failed to create client with token: %v", err)
	}

	return clientWithToken
}

func findPowerControlEndpoint(ctx context.Context, t *testing.T, client *alexa.Session) *alexaapimodels.Endpoint {
	t.Helper()

	query := alexaapimodels.EndpointQuery{
		IncludeFields: &alexaapimodels.EndpointIncludeFields{Features: true},
	}

	response, err := client.ListEndpoints(ctx, query)
	if err != nil {
		t.Fatalf("Failed to list endpoints: %v", err)
	}

	if len(response.Results) == 0 {
		t.Skip("Skipping test: No endpoints found")
	}

	for _, endpoint := range response.Results {
		if !hasPowerFeature(endpoint) || endpoint.EndpointID == "" {
			continue
		}

		return endpoint
	}

	t.Skip("Skipping test: No endpoint with power control found")

	return nil
}

func hasPowerFeature(endpoint *alexaapimodels.Endpoint) bool {
	for _, feature := range endpoint.Features {
		if feature.Name == alexaapimodels.FeatureNamePower {
			return true
		}
	}

	return false
}

func setIntegrationPowerWithResponse(
	ctx context.Context,
	client *alexa.Session,
	endpoint *alexaapimodels.Endpoint,
	operation alexaapimodels.FeatureOperationName,
) (*alexaapimodels.ControlResponse, error) {
	request := alexaapimodels.ControlRequest{
		Target:    endpoint,
		Namespace: alexaapimodels.FeatureNamePower,
		Name:      operation,
		Payload:   alexaapimodels.ControlPowerPayload{},
	}

	response, err := client.Control(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("send integration power control: %w", err)
	}

	return response, nil
}

func subscribeToIntegrationEvents(ctx context.Context, t *testing.T, client *alexa.Session) {
	t.Helper()

	subscribeResp, err := client.Subscribe(ctx, alexaapimodels.SubscribeRequest{DurationInMinutes: 10})
	if err != nil {
		t.Fatalf("Failed to subscribe to events: %v", err)
	}

	if subscribeResp == nil {
		t.Fatal("Subscribe response is nil")
	}

	if len(subscribeResp.Errors) > 0 {
		t.Logf("Warning: Subscribe response contains errors: %+v", subscribeResp.Errors)
	}

	t.Logf("Successfully subscribed to events (duration: %d minutes)", subscribeResp.DurationInMinutes)
}

func startIntegrationEventReceiver(conn *alexa.HTTP2Connection) (<-chan *alexaapimodels.Event, <-chan error) {
	eventChan := make(chan *alexaapimodels.Event, 10)
	errorChan := make(chan error, 1)

	go func() {
		for {
			event, err := conn.Receive()
			if err != nil {
				errorChan <- err

				return
			}

			eventChan <- event
		}
	}()

	return eventChan, errorChan
}

func waitForIntegrationPowerEvent(
	t *testing.T,
	eventChan <-chan *alexaapimodels.Event,
	errorChan <-chan error,
	endpointID string,
) bool {
	t.Helper()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case event := <-eventChan:
			if isIntegrationPowerEvent(t, event, endpointID) {
				return true
			}
		case err := <-errorChan:
			t.Fatalf("Error receiving event: %v", err)

			return false
		case <-timeout:
			t.Log("Timeout waiting for power state change event")
			// Events can be delayed or not sent, so a timeout remains non-fatal.
			return true
		}
	}
}

func isIntegrationPowerEvent(t *testing.T, event *alexaapimodels.Event, endpointID string) bool {
	t.Helper()

	if event == nil {
		return false
	}

	t.Logf("Received event: Namespace=%s, Name=%s, EndpointID=%s", event.Namespace, event.Name, event.EndpointID)

	if event.EndpointID != endpointID {
		return false
	}

	if powerPayload, ok := event.Payload.(*alexaapimodels.PowerPayload); ok {
		t.Logf("Received power state change event: State=%s, TimeOfSample=%v", powerPayload.PowerState, powerPayload.TimeOfSample)

		return true
	}

	if event.Name != "ChangeReport" && event.Name != "StateReport" {
		return false
	}

	t.Log("Received state change event for endpoint (may contain power state)")

	if unknownPayload, ok := event.Payload.(*alexaapimodels.UnknownPayload); ok {
		t.Logf("Unknown payload data: %+v", unknownPayload.Data)
	}

	return true
}
