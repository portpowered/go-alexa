package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSDKConstantsPreserveTypedAndUntypedPublicValues(t *testing.T) {
	t.Parallel()

	schema := []byte(`components:
  schemas:
    State:
      x-go-sdk-constant-type: PowerState
      x-go-compat-constant-values:
        SDKPowerStateOn: ON
        SDKPowerStateOff: OFF
    Default:
      x-go-compat-constant-values:
        SDKDefaultLocale: en-US
        WireOnly: wire
`)

	generated, err := renderSDKConstants([][]byte{schema})
	if err != nil {
		t.Fatalf("render SDK constants: %v", err)
	}

	for _, expected := range []string{"package alexaapimodels", `PowerState = "ON"`, `PowerState = "OFF"`, `= "en-US"`} {
		if !strings.Contains(string(generated), expected) {
			t.Fatalf("missing expected SDK semantics %s: %s", expected, generated)
		}
	}

	if bytes.Contains(generated, []byte("WireOnly")) {
		t.Fatal("wire-only constant leaked into SDK projection")
	}

	reordered := bytes.ReplaceAll(schema, []byte("        SDKPowerStateOn: ON\n        SDKPowerStateOff: OFF"),
		[]byte("        SDKPowerStateOff: OFF\n        SDKPowerStateOn: ON"))

	repeated, err := renderSDKConstants([][]byte{reordered})
	if err != nil || !bytes.Equal(generated, repeated) {
		t.Fatalf("SDK output depends on schema map order: %v", err)
	}
}

func TestSDKConstantsRejectEmptyInvalidAndDuplicateMetadata(t *testing.T) {
	t.Parallel()

	for _, schema := range []string{
		"components: {schemas: {}}",
		"components: {schemas: {State: {x-go-compat-constant-values: {SDKbad-name: ON}}}}",
		"components: {schemas: {State: {x-go-compat-constant-values: {SDKState: ON}}, Other: {x-go-compat-constant-values: {SDKState: OFF}}}}}",
	} {
		_, err := renderSDKConstants([][]byte{[]byte(schema)})
		if err == nil {
			t.Fatalf("invalid SDK metadata accepted: %s", schema)
		}
	}
}
