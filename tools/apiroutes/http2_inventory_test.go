package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckExternalHTTP2InventoryAcceptsCurrentManifest(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)

	openAPI, asyncAPI := readCurrentAPIDocuments(t, root)

	err := checkExternalHTTP2Inventory(root, openAPI, asyncAPI)
	if err != nil {
		t.Fatalf("current external HTTP/2 inventory rejected: %v", err)
	}
}

func TestCheckExternalHTTP2InventoryRejectsWrongVersion(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	openAPI, asyncAPI := readCurrentAPIDocuments(t, root)
	manifest := readTestFile(t, filepath.Join(root, "api", "external", "http2.yaml"))
	goMod := readTestFile(t, filepath.Join(root, "go.mod"))
	currentVersion := moduleVersion(t, goMod)

	manifest = strings.Replace(manifest, "version: "+currentVersion, "version: v0.0.0", 1)
	if !strings.Contains(manifest, "version: v0.0.0") {
		t.Fatal("could not change the manifest version for the negative test")
	}

	testRoot := writeHTTP2InventoryRoot(t, root, manifest)

	err := checkExternalHTTP2Inventory(testRoot, openAPI, asyncAPI)
	if err == nil || !strings.Contains(err.Error(), "does not match golang.org/x/net in go.mod") {
		t.Fatalf("got error %v, want wrong dependency version error", err)
	}
}

func TestCheckExternalHTTP2InventoryRejectsExtraOperation(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	openAPI, asyncAPI := readCurrentAPIDocuments(t, root)
	manifest := readTestFile(t, filepath.Join(root, "api", "external", "http2.yaml"))
	manifest = strings.Replace(
		manifest,
		"    callsite: pkg/alexa/http2.go#ping\n",
		"    callsite: pkg/alexa/http2.go#ping\n  - operation: unregisteredOperation\n    schema: ../openapi.yaml#/paths/~1unknown/get\n",
		1,
	)
	testRoot := writeHTTP2InventoryRoot(t, root, manifest)

	err := checkExternalHTTP2Inventory(testRoot, openAPI, asyncAPI)
	if err == nil || !strings.Contains(err.Error(), "exactly 2 schema-backed exchanges") {
		t.Fatalf("got error %v, want extra operation error", err)
	}
}

func TestCheckExternalHTTP2InventoryRejectsMismatchedMethod(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	openAPI, asyncAPI := readCurrentAPIDocuments(t, root)
	operations := openAPI.Paths["/v20160207/directives"]
	streamOperation := operations["get"]
	delete(operations, "get")
	operations["post"] = streamOperation

	testRoot := writeHTTP2InventoryRoot(t, root, readTestFile(t, filepath.Join(root, "api", "external", "http2.yaml")))

	err := checkExternalHTTP2Inventory(testRoot, openAPI, asyncAPI)
	if err == nil || !strings.Contains(err.Error(), "does not match an inventoried OpenAPI route") {
		t.Fatalf("got error %v, want HTTP method mismatch error", err)
	}
}

func readCurrentAPIDocuments(t *testing.T, root string) (openAPIDocument, asyncAPIDocument) {
	t.Helper()

	var openAPI openAPIDocument

	err := readYAML(filepath.Join(root, "api", "openapi.yaml"), &openAPI)
	if err != nil {
		t.Fatal(err)
	}

	var asyncAPI asyncAPIDocument

	err = readYAML(filepath.Join(root, "api", "asyncapi.yaml"), &asyncAPI)
	if err != nil {
		t.Fatal(err)
	}

	return openAPI, asyncAPI
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate HTTP/2 inventory test source")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", ".."))
}

func moduleVersion(t *testing.T, goMod string) string {
	t.Helper()

	for _, line := range strings.Split(goMod, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "golang.org/x/net" {
			return fields[1]
		}
	}

	t.Fatal("go.mod does not declare golang.org/x/net")

	return ""
}

func writeHTTP2InventoryRoot(t *testing.T, repository, manifest string) string {
	t.Helper()

	root := t.TempDir()

	err := os.MkdirAll(filepath.Join(root, "api", "external"), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	goMod := readTestFile(t, filepath.Join(repository, "go.mod"))

	err = os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(root, "api", "external", "http2.yaml"), []byte(manifest), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return root
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	//nolint:gosec // Tests read fixed repository paths or files under t.TempDir.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
