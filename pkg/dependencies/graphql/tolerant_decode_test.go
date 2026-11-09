//nolint:testpackage // Verifies private transport decoding through the public client.
package graphql

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	goodBrightness = `{"__typename":"Brightness","name":"brightness","brightnessStateValue":40}`
	goodPower      = `{"__typename":"Power","name":"powerState"}`
)

func decodeStates(t *testing.T, body string) (*ListEndpointsWithStatesResponse, error) {
	t.Helper()

	client := NewClient(
		WithBaseURL("https://alexa.synthetic.test"),
		WithBearerToken("synthetic-token"),
		WithHTTPClient(&http.Client{Transport: graphqlRoundTripper(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		})}),
	)

	return client.ListEndpointsWithStates(context.Background(), testListEndpointsInput([]string{"synthetic-endpoint"}, nil))
}

func statesBody(features string) string {
	return `{"data":{"listEndpoints":{"completeResult":true,"endpoints":[` +
		`{"id":"e1","endpointId":"synthetic-1","dmsIdentifier":{"deviceType":"T","dsn":"s1"},"features":` + features + `},` +
		`{"id":"e2","endpointId":"synthetic-2","dmsIdentifier":{"deviceType":"T","dsn":"s2"},` +
		`"features":[{"name":"brightness","instance":"","properties":[` + goodBrightness + `]}]}` +
		`]}}}`
}

func propertyNames(response *ListEndpointsWithStatesResponse, endpoint int) []string {
	var names []string

	for _, feature := range response.ListEndpoints.Endpoints[endpoint].Features {
		for _, property := range feature.Properties {
			names = append(names, property.GetName())
		}
	}

	return names
}

func TestListEndpointsWithStatesSkipsUndecodableProperties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		properties string
		want       []string
	}{
		{
			"unknown typename among good ones",
			goodBrightness + `,{"__typename":"SyntheticFutureKind","name":"future"},` + goodPower,
			[]string{"brightness", "powerState"},
		},
		{"missing typename", `{"name":"untyped"},` + goodBrightness, []string{"brightness"}},
		{"null property", `null,` + goodBrightness + `,null`, []string{"brightness"}},
		{"known kind with wrong field type", `{"__typename":"Brightness","name":"brightness","brightnessStateValue":"loud"},` + goodPower, []string{"powerState"}},
		{"only bad properties", `{"__typename":"SyntheticFutureKind","name":"future"}`, nil},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			response, err := decodeStates(t, statesBody(`[{"name":"brightness","instance":"","properties":[`+testCase.properties+`]}]`))
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if len(response.ListEndpoints.Endpoints) != 2 {
				t.Fatalf("every endpoint must survive, got %d", len(response.ListEndpoints.Endpoints))
			}

			if got := strings.Join(propertyNames(response, 0), ","); got != strings.Join(testCase.want, ",") {
				t.Fatalf("properties = %q, want %q", got, testCase.want)
			}

			if got := propertyNames(response, 1); len(got) != 1 || got[0] != "brightness" {
				t.Fatalf("unrelated endpoint lost its property: %v", got)
			}
		})
	}
}

func TestListEndpointsWithStatesToleratesUnknownConfigurationAndNullFeatures(t *testing.T) {
	t.Parallel()

	features := `[null,{"name":"range","instance":"i",` +
		`"configuration":{"__typename":"SyntheticFutureConfig"},` +
		`"semantics":{"__typename":"SyntheticFutureSemantics"},` +
		`"properties":[` + goodBrightness + `]}]`

	response, err := decodeStates(t, statesBody(features))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	endpoint := response.ListEndpoints.Endpoints[0]
	if len(endpoint.Features) != 2 || endpoint.Features[1].Configuration != nil || len(endpoint.Features[1].Properties) != 1 {
		t.Fatalf("unexpected features: %#v", endpoint.Features)
	}
}

func TestListEndpointsWithStatesMalformedBodiesStillFail(t *testing.T) {
	t.Parallel()

	valid := statesBody(`[{"name":"brightness","instance":"","properties":[` + goodBrightness + `]}]`)

	bodies := map[string]string{
		"malformed": `{"data":{"listEndpoints":`,
		"truncated": valid[:len(valid)/2],
		"garbage":   `<html>not json</html>`,
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := decodeStates(t, body)
			if err == nil {
				t.Fatal("expected a decode error")
			}

			var requestErr *alexaapimodels.RequestError
			if !errors.As(err, &requestErr) || requestErr.Stage != "response_decode" {
				t.Fatalf("expected response_decode request error, got %v", err)
			}
		})
	}
}

func TestResponseDecodeKeepsCauseAndTypeName(t *testing.T) {
	t.Parallel()

	// A wrongly typed scalar outside the tolerant paths still fails and keeps its cause.
	_, err := decodeStates(t, `{"data":{"listEndpoints":{"completeResult":"synthetic-secret","endpoints":[]}}}`)
	if err == nil {
		t.Fatal("expected a decode error")
	}

	var bad *alexaapimodels.BadRequestError
	if !errors.As(err, &bad) || bad.Err == nil {
		t.Fatalf("underlying cause was dropped: %v", err)
	}

	if !strings.Contains(err.Error(), "ListEndpointsWithStates") {
		t.Fatalf("message should name the type, got %q", err.Error())
	}

	if strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatalf("message leaks payload: %q", err.Error())
	}
}
