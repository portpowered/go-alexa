package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/portpowered/go-alexa/pkg/dependencies/rest/internal/wire"
)

func TestGenerateCodePairUsesGeneratedWireModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		if r.Method != http.MethodPost || r.URL.Path != "/auth/create/codepair" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var request wire.WireCodePairRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode synthetic request: %v", err)
		}
		if request.CodeData.AppName != "test-client" || request.CodeData.DeviceSerial != "synthetic-device" {
			t.Errorf("unexpected registration data: %#v", request.CodeData)
		}
		if request.CodeData.SecondaryRegistration != "False" || len(request.Scopes) != 0 {
			t.Errorf("unexpected code-pair options: %#v", request)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"public_code":"synthetic-public","private_code":"synthetic-private","futureField":"preserved-by-wire"}`))
	}))
	defer server.Close()

	client := NewClient(WithHTTPClient(server.Client()), WithAmazonapiBaseURI(server.URL))
	response, err := client.GenerateCodePair(context.Background(), &DeviceRegistrationConfig{
		AppName:      "test-client",
		AppVersion:   "1.0",
		DeviceType:   "test-device-type",
		Domain:       "Device",
		DeviceModel:  "test-model",
		OSVersion:    "test-os",
		DeviceSerial: "synthetic-device",
		DeviceName:   "Test device",
	})
	if err != nil {
		t.Fatalf("GenerateCodePair failed: %v", err)
	}
	if response.PublicCode != "synthetic-public" || response.PrivateCode != "synthetic-private" {
		t.Fatalf("unexpected public response: %#v", response)
	}
}

func TestGetEndpointsUsesGeneratedWireModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		if r.Method != http.MethodGet || r.URL.Path != "/v2/endpoints" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("owner") != "~caller" || query.Get("maxResults") != "5" || query.Get("expand") != "all" {
			t.Errorf("unexpected query parameters: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"synthetic-endpoint","name":"Lamp","type":"LIGHT","capabilities":["POWER"],"state":{"on":true},"metadata":{"room":"test"},"futureDeviceField":"ignored-by-public-model"}],"nextToken":"synthetic-next","futureResponseField":"ignored-by-public-model"}`))
	}))
	defer server.Close()

	client := NewClient(WithHTTPClient(server.Client()), WithAmazonalexaAPIBaseURI(server.URL), WithBearerToken("synthetic-token"))
	response, err := client.GetEndpoints(context.Background(), &ListEndpointsOptions{
		Owner:      "~caller",
		Expand:     []string{"all"},
		MaxResults: 5,
	})
	if err != nil {
		t.Fatalf("GetEndpoints failed: %v", err)
	}
	if response.NextToken != "synthetic-next" || len(response.Results) != 1 {
		t.Fatalf("unexpected endpoint-list response: %#v", response)
	}
	device := response.Results[0]
	if device.ID != "synthetic-endpoint" || device.Name != "Lamp" || device.Type != "LIGHT" {
		t.Fatalf("unexpected device: %#v", device)
	}
	if device.State["on"] != true || device.Metadata["room"] != "test" {
		t.Fatalf("unexpected device state or metadata: %#v", device)
	}
}

func TestConvertWireModelRetainsAdditionalProperties(t *testing.T) {
	input := map[string]interface{}{
		"futureResponseField": "retained",
		"results": []interface{}{map[string]interface{}{
			"id":                "synthetic-endpoint",
			"futureDeviceField": "retained",
		}},
	}

	converted, err := convertWireModel[wire.WireEndpointListResponse](input)
	if err != nil {
		t.Fatalf("convertWireModel failed: %v", err)
	}
	encoded, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf("marshal generated model: %v", err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(encoded, &actual); err != nil {
		t.Fatalf("unmarshal generated model: %v", err)
	}
	if actual["futureResponseField"] != "retained" {
		t.Errorf("top-level additional property was lost: %s", encoded)
	}
	results, ok := actual["results"].([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("unexpected generated model results: %#v", actual["results"])
	}
	device, ok := results[0].(map[string]interface{})
	if !ok || device["futureDeviceField"] != "retained" {
		t.Errorf("nested additional property was lost: %#v", results[0])
	}
}
