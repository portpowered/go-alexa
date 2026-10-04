package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestBuildBundleMergesResponsibilitiesDeterministically(t *testing.T) {
	t.Parallel()
	root := prepareTestBundle(t)

	first, outputPath, err := buildBundle(root)
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}

	second, secondOutputPath, err := buildBundle(root)
	if err != nil {
		t.Fatalf("rebuild bundle: %v", err)
	}

	if outputPath != "api/openapi.yaml" || secondOutputPath != outputPath {
		t.Fatalf("unexpected output paths %q and %q", outputPath, secondOutputPath)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatal("bundle output changed between identical builds")
	}

	var decoded interface{}

	err = yaml.Unmarshal(first, &decoded)
	if err != nil {
		t.Fatalf("parse bundle output: %v", err)
	}

	bundle, ok := normalize(decoded).(map[string]interface{})
	if !ok {
		t.Fatalf("bundle root has type %T", decoded)
	}

	paths, err := stringMap(bundle["paths"], "paths")
	if err != nil {
		t.Fatal(err)
	}

	if len(paths) != 2 {
		t.Fatalf("got %d bundled paths, want 2", len(paths))
	}

	components, err := stringMap(bundle["components"], "components")
	if err != nil {
		t.Fatal(err)
	}

	schemas, err := stringMap(components["schemas"], "components.schemas")
	if err != nil {
		t.Fatal(err)
	}

	if len(schemas) != 2 {
		t.Fatalf("got %d bundled schemas, want 2", len(schemas))
	}

	securitySchemes, err := stringMap(components["securitySchemes"], "components.securitySchemes")
	if err != nil {
		t.Fatal(err)
	}

	if len(securitySchemes) != 1 {
		t.Fatalf("got %d bundled security schemes, want 1", len(securitySchemes))
	}
}

func TestValidateSourceSpecRejectsCrossResponsibilityOperation(t *testing.T) {
	t.Parallel()

	spec := map[string]interface{}{
		"paths": map[string]interface{}{
			"/accounts": map[string]interface{}{
				"get": map[string]interface{}{"tags": []interface{}{"Accounts", "Devices"}},
			},
		},
		"tags": []interface{}{map[string]interface{}{"name": "Accounts"}},
	}

	err := validateSourceSpec(sourceSpec{File: "accounts.yaml", Responsibility: "Accounts", Tag: "Accounts"}, spec)
	if err == nil {
		t.Fatal("expected cross-responsibility tag to fail validation")
	}
}

func TestValidateComponentReferencesRejectsMissingDependency(t *testing.T) {
	t.Parallel()

	spec := map[string]interface{}{
		"paths": map[string]interface{}{
			"/accounts": map[string]interface{}{
				"get": map[string]interface{}{"schema": map[string]interface{}{"$ref": "#/components/schemas/Missing"}},
			},
		},
		"components": map[string]interface{}{"schemas": map[string]interface{}{}},
	}

	err := validateComponentReferences(spec)
	if err == nil || !strings.Contains(err.Error(), "unresolved component reference") {
		t.Fatalf("got error %v, want unresolved component error", err)
	}
}

func writeTestFile(t *testing.T, root, path, content string) {
	t.Helper()

	fullPath := filepath.Join(root, filepath.FromSlash(path))

	err := os.MkdirAll(filepath.Dir(fullPath), 0o750)
	if err != nil {
		t.Fatalf("create test directory: %v", err)
	}

	err = os.WriteFile(fullPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func prepareTestBundle(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, sourceManifestPath, `bundle: api/openapi.yaml
sources:
  - file: api/openapi/sources/accounts.yaml
    responsibility: Accounts
    tag: Accounts
  - file: api/openapi/sources/devices.yaml
    responsibility: Devices
    tag: Devices
standalone:
  - file: api/behaviors.yaml
    responsibility: Behavior models
`)

	sharedSecurity := "security:\n  - BearerToken: []\n"
	writeTestFile(t, root, "api/openapi/sources/accounts.yaml", `openapi: 3.0.3
info: {title: Test API, version: 1.0.0}
servers: [{url: https://example.test}]
`+sharedSecurity+`tags: [{name: Accounts}]
paths:
  /accounts:
    get:
      tags: [Accounts]
      responses:
        '200':
          description: account
          content:
            application/json:
              schema: {$ref: '#/components/schemas/Account'}
components:
  securitySchemes:
    BearerToken: {type: http, scheme: bearer}
  schemas:
    Account:
      type: object
      properties:
        id: {type: string}
`)
	writeTestFile(t, root, "api/openapi/sources/devices.yaml", `openapi: 3.0.3
info: {title: Test API, version: 1.0.0}
servers: [{url: https://example.test}]
`+sharedSecurity+`tags: [{name: Devices}]
paths:
  /devices:
    get:
      tags: [Devices]
      responses:
        '200':
          description: device
          content:
            application/json:
              schema: {$ref: '#/components/schemas/Device'}
components:
  securitySchemes:
    BearerToken: {type: http, scheme: bearer}
  schemas:
    Account:
      type: object
      properties:
        id: {type: string}
    Device:
      type: object
      properties:
        account: {$ref: '#/components/schemas/Account'}
`)
	writeTestFile(t, root, "api/behaviors.yaml", `openapi: 3.0.3
info: {title: Behavior models, version: 1.0.0}
paths: {}
`)

	return root
}
