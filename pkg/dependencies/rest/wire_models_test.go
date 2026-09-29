//nolint:testpackage // Exercises private request paths while checking generated wire shapes.
package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/portpowered/go-alexa/pkg/dependencies/internal/wire"
)

const expandAll = "all"

func TestGenerateCodePairUsesGeneratedWireModels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, httpRequest *http.Request) {
		t.Helper()

		if httpRequest.Method != http.MethodPost || httpRequest.URL.Path != "/auth/create/codepair" {
			t.Errorf("unexpected request: %s %s", httpRequest.Method, httpRequest.URL.Path)
		}

		var codePairRequest wire.WireCodePairRequest

		err := json.NewDecoder(httpRequest.Body).Decode(&codePairRequest)
		if err != nil {
			t.Errorf("decode synthetic request: %v", err)
		}

		if codePairRequest.CodeData.AppName != "test-client" || codePairRequest.CodeData.DeviceSerial != "synthetic-device" {
			t.Errorf("unexpected registration data: %#v", codePairRequest.CodeData)
		}

		if codePairRequest.CodeData.SecondaryRegistration != "False" || len(codePairRequest.Scopes) != 0 {
			t.Errorf("unexpected code-pair options: %#v", codePairRequest)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"public_code":"synthetic-public","private_code":"synthetic-private","futureField":"preserved-by-wire"}`))
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
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Helper()

		if request.Method != http.MethodGet || request.URL.Path != "/v2/endpoints" {
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
		}

		query := request.URL.Query()
		if query.Get("owner") != "~caller" || query.Get("maxResults") != "5" || query.Get("expand") != expandAll {
			t.Errorf("unexpected query parameters: %s", request.URL.RawQuery)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(
			`{"results":[{"id":"synthetic-endpoint","name":"Lamp","type":"LIGHT",` +
				`"capabilities":["POWER"],"state":{"on":true},"metadata":{"room":"test"},` +
				`"futureDeviceField":"ignored-by-public-model"}],"nextToken":"synthetic-next",` +
				`"futureResponseField":"ignored-by-public-model"}`,
		))
	}))
	defer server.Close()

	client := NewClient(WithHTTPClient(server.Client()), WithAmazonalexaAPIBaseURI(server.URL), WithBearerToken("synthetic-token"))

	response, err := client.GetEndpoints(context.Background(), &ListEndpointsOptions{
		Owner:      "~caller",
		Expand:     []string{expandAll},
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
	t.Parallel()

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
	{
		err := json.Unmarshal(encoded, &actual)
		if err != nil {
			t.Fatalf("unmarshal generated model: %v", err)
		}
	}

	if actual["futureResponseField"] != "retained" {
		t.Errorf("top-level additional property was lost: %s", encoded)
	}

	results, isResultsList := actual["results"].([]interface{})
	if !isResultsList || len(results) != 1 {
		t.Fatalf("unexpected generated model results: %#v", actual["results"])
	}

	device, isDeviceMap := results[0].(map[string]interface{})
	if !isDeviceMap || device["futureDeviceField"] != "retained" {
		t.Errorf("nested additional property was lost: %#v", results[0])
	}
}
