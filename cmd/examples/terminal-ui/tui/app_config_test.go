package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigFromFilePreservesCustomerIDKey(t *testing.T) {
	configDir := t.TempDir()
	contents := []byte(`{"accessToken":"synthetic-access","refreshToken":"synthetic-refresh","customerID":"synthetic-customer"}`)

	err := os.WriteFile(filepath.Join(configDir, "config.json"), contents, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(configDir)

	got := getConfigFromFile()
	if got == nil || got.CustomerID != "synthetic-customer" {
		t.Fatalf("legacy config key was not loaded: %+v", got)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(encoded), `"customerID":"synthetic-customer"`) {
		t.Fatalf("config changed its persisted customer ID key: %s", encoded)
	}
}
