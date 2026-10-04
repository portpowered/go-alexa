package main

import (
	"go/token"
	"strings"
	"testing"
)

func TestWirePackageStorageRejectsCrossFunctionMutation(t *testing.T) {
	t.Parallel()

	for _, split := range []bool{false, true} {
		set := token.NewFileSet()
		storage := `package probe
var saved map[string]any
func mutate(caller string) { saved["brandNewCrossFunctionKey"] = caller }
`
		consumer := `import (
 "encoding/json"
 wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
)
func retain(input *wire.Payload) { saved = input.Extra }
func send(input *wire.Payload, caller string) { retain(input); mutate(caller); _, _ = json.Marshal(input) }
`

		var files []wireSourceFile

		if split {
			files = []wireSourceFile{
				parseWireSourceTestFile(t, set, "pkg/probe/storage.go", storage),
				parseWireSourceTestFile(t, set, "pkg/probe/consumer.go", "package probe\n"+consumer),
			}
		} else {
			imports, functions, _ := strings.Cut(consumer, "func retain")
			files = []wireSourceFile{parseWireSourceTestFile(t, set, "pkg/probe/one.go",
				"package probe\n"+imports+storage[len("package probe\n"):]+"func retain"+functions)}
		}

		err := checkWireSourcePackage(files, set, wireConstructionTestModels())
		if err == nil {
			t.Fatalf("cross-function generated map mutation accepted (split=%v)", split)
		}
	}
}

func TestWireConstructionRejectsHelperReturnedFixedValues(t *testing.T) {
	t.Parallel()

	for _, helper := range []string{
		`func inventedValue() string { return "brandNewUnregisteredValue" }`,
		`func inventedValue() string { return intermediate() }; func intermediate() string { return "brandNewUnregisteredValue" }`,
		`func inventedValue() string { value := "brandNewUnregisteredValue"; return value }`,
	} {
		set := token.NewFileSet()

		file := parseWireSourceTestFile(t, set, "pkg/probe.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
`+helper+`
func request() { _ = wire.Payload{Value: inventedValue()} }
`)

		err := checkWireSourcePackage([]wireSourceFile{file}, set, wireConstructionTestModels())
		if err == nil {
			t.Fatalf("helper-returned fixed wire value accepted: %s", helper)
		}
	}
}

func TestWirePackageStorageRejectsExternallySuppliedGeneratedPayload(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()
	files := []wireSourceFile{
		parseWireSourceTestFile(t, set, "pkg/probe/storage.go", `package probe
var saved map[string]any
func Mutate(caller string) { saved["brandNewCrossFunctionKey"] = caller }
`),
		parseWireSourceTestFile(t, set, "pkg/probe/consumer.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func Retain(input *wire.Payload) { saved = input.Extra }
`),
	}

	err := checkWireSourcePackage(files, set, wireConstructionTestModels())
	if err == nil {
		t.Fatal("cross-file generated map mutation accepted without a local helper call")
	}
}
