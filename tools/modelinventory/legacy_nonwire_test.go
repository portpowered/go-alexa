package main

import (
	"go/token"
	"strings"
	"testing"
)

func TestLegacyNonwireDispositionRejectsWireFieldsAndCodecs(t *testing.T) {
	t.Parallel()

	const (
		path   = "pkg/dependencymodels/auth.go"
		name   = "AuthConfig"
		reason = "AuthConfig: unused caller configuration for source compatibility; no SDK consumer, JSON tags, or codec."
	)

	allowances := map[string]handwrittenAllowance{
		path + "::" + name: {
			File: path, Name: name, Classification: legacyNonwireClassification, Reason: reason,
			Evidence: []string{"../../pkg/dependencymodels/legacy_auth_test.go#TestLegacyOAuthConfigurationHasNoWireTags"},
		},
	}

	for _, change := range []string{
		"type AuthConfig struct { ClientID string }",
		"type AuthConfig struct { ClientID string `json:\"client_id\"` }",
		"type AuthConfig struct { ClientID string }; func encode(value AuthConfig) { json.Marshal(value) }",
		"type AuthConfig struct { ClientID string }; func (AuthConfig) MarshalJSON() ([]byte, error) { return nil, nil }",
	} {
		data := []byte("package alexamodels\nimport \"encoding/json\"\n//modelinventory:legacy-nonwire " + reason + "\n" + change)
		_, problems, seen := scanHandwrittenSource(path, data, token.NewFileSet(), false, allowances)

		wantRejected := strings.Contains(change, "json") || strings.Contains(change, "MarshalJSON")
		if (len(problems) != 0) != wantRejected || !seen[path+"::"+name] {
			t.Fatalf("legacy disposition for %q: problems=%v, seen=%v", change, problems, seen)
		}
	}
}

func TestUnlistedDependencyStructWithoutTagsIsRejected(t *testing.T) {
	t.Parallel()

	const path = "pkg/dependencymodels/unused.go"

	data := []byte("package alexamodels; type UnusedResponse struct { Status string }; type privateWire struct { Status string }")

	_, problems, seen := scanHandwrittenSource(path, data, token.NewFileSet(), false, nil)
	if len(problems) == 0 || !seen[path+"::UnusedResponse"] || !seen[path+"::privateWire"] {
		t.Fatalf("untagged exported dependency model escaped inventory: problems=%v, seen=%v", problems, seen)
	}
}
