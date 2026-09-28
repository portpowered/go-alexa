package alexa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// captureTransport is a test RoundTripper that captures the request body and returns 200 OK.
type captureTransport struct {
	lastRequestBody []byte
	lastRequestURL  string
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		t.lastRequestBody, _ = io.ReadAll(req.Body)
	}
	t.lastRequestURL = req.URL.String()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
		Header:     make(http.Header),
	}, nil
}

func TestControlAudioPlayerURI(t *testing.T) {
	transport := &captureTransport{}
	httpClient := &http.Client{Transport: transport}

	client, _ := NewClient(WithHttpClient(httpClient), WithBearerToken("test-token"))

	endpoint := &alexaapimodels.Endpoint{
		DeviceType:         "synthetic-device-type",
		DeviceSerialNumber: "synthetic-device-serial",
		EndpointID:         "amzn1.alexa.endpoint.test-123",
	}

	testURI := "https://example.com/audio/test-tts.mp3"

	resp, err := client.Control(context.Background(), alexaapimodels.ControlRequest{
		Target:    endpoint,
		Namespace: alexaapimodels.FeatureNameAudioPlayer,
		Name:      alexaapimodels.FeatureOperationNamePlayURI,
		Payload: alexaapimodels.ControlAudioPlayerURIPayload{
			URI:   testURI,
			Title: "Test Audio",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", resp.Errors)
	}

	// Verify the request was sent to the behaviors preview endpoint
	if transport.lastRequestURL == "" {
		t.Fatal("no request was captured")
	}

	// Parse the behavior preview request body
	var behaviorReq alexamodels.BehaviorPreviewRequest
	if err := json.Unmarshal(transport.lastRequestBody, &behaviorReq); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	if behaviorReq.BehaviorID != alexamodels.DefaultBehaviorID {
		t.Errorf("expected behaviorId %q, got %q", alexamodels.DefaultBehaviorID, behaviorReq.BehaviorID)
	}
	if behaviorReq.Status != alexamodels.DefaultBehaviorStatus {
		t.Errorf("expected status %q, got %q", alexamodels.DefaultBehaviorStatus, behaviorReq.Status)
	}

	// Parse the sequence JSON
	var sequence alexamodels.Sequence
	if err := json.Unmarshal([]byte(behaviorReq.SequenceJSON), &sequence); err != nil {
		t.Fatalf("failed to parse sequence JSON: %v", err)
	}

	if sequence.Type != alexamodels.ModelTypeSequence {
		t.Errorf("expected sequence type %q, got %q", alexamodels.ModelTypeSequence, sequence.Type)
	}

	// Parse the start node
	startNodeBytes, err := json.Marshal(sequence.StartNode)
	if err != nil {
		t.Fatalf("failed to marshal startNode: %v", err)
	}

	var node alexamodels.OpaquePayloadOperationNode
	if err := json.Unmarshal(startNodeBytes, &node); err != nil {
		t.Fatalf("failed to parse start node: %v", err)
	}

	if node.Type != alexamodels.ModelTypeOpaquePayloadOperationNode {
		t.Errorf("expected node type %q, got %q", alexamodels.ModelTypeOpaquePayloadOperationNode, node.Type)
	}
	if node.OperationType != alexamodels.OperationTypeSpeak {
		t.Errorf("expected operation type %q, got %q", alexamodels.OperationTypeSpeak, node.OperationType)
	}

	// Verify payload contents
	payload := node.OperationPayload
	if payload[alexamodels.PayloadKeyDeviceType] != "synthetic-device-type" {
		t.Errorf("expected deviceType %q, got %q", "synthetic-device-type", payload[alexamodels.PayloadKeyDeviceType])
	}
	if payload[alexamodels.PayloadKeyDeviceSerialNumber] != "synthetic-device-serial" {
		t.Errorf("expected deviceSerialNumber %q, got %q", "synthetic-device-serial", payload[alexamodels.PayloadKeyDeviceSerialNumber])
	}
	textOut := fmt.Sprintf("<audio src='%s'/>", testURI)
	if payload[alexamodels.PayloadKeyTextToSpeak] != textOut {
		t.Errorf("expected textToSpeak %q, got %q", textOut, payload[alexamodels.PayloadKeyTextToSpeak])
	}
}

func TestControlAudioPlayerURI_InvalidPayload(t *testing.T) {
	transport := &captureTransport{}
	httpClient := &http.Client{Transport: transport}

	client, _ := NewClient(WithHttpClient(httpClient))

	endpoint := &alexaapimodels.Endpoint{
		DeviceType:         "synthetic-device-type",
		DeviceSerialNumber: "synthetic-device-serial",
	}

	// Wrong payload type should return error
	_, err := client.Control(context.Background(), alexaapimodels.ControlRequest{
		Target:    endpoint,
		Namespace: alexaapimodels.FeatureNameAudioPlayer,
		Name:      alexaapimodels.FeatureOperationNamePlayURI,
		Payload:   "not-a-valid-payload",
	})
	if err == nil {
		t.Fatal("expected error for invalid payload type")
	}
}

func TestControlAudioPlayerURI_NilTarget(t *testing.T) {
	transport := &captureTransport{}
	httpClient := &http.Client{Transport: transport}

	client, _ := NewClient(WithHttpClient(httpClient))

	_, err := client.Control(context.Background(), alexaapimodels.ControlRequest{
		Target:    nil,
		Namespace: alexaapimodels.FeatureNameAudioPlayer,
		Name:      alexaapimodels.FeatureOperationNamePlayURI,
		Payload: alexaapimodels.ControlAudioPlayerURIPayload{
			URI: "https://example.com/audio.mp3",
		},
	})
	if err == nil {
		t.Fatal("expected error for nil target")
	}
}
