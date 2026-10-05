package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestWireConstructionRejectsSiblingFileKeysAndHelperMutations(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		`input.Extra[rawKey] = caller`,
		`mutate(input.Extra, caller)`,
	} {
		set := token.NewFileSet()
		files := []wireSourceFile{
			parseWireSourceTestFile(t, set, "pkg/definitions.go", `package probe
const rawKey = "brandNewSiblingKey"
func mutate(extra map[string]any, caller string) { extra[rawKey] = caller }
`),
			parseWireSourceTestFile(t, set, "cmd/example/use.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func request(input *wire.Payload, caller string) { `+source+` }
`),
		}

		err := checkWireSourcePackage(files, set, wireConstructionTestModels())
		if err == nil {
			t.Fatalf("accepted sibling-file wire mutation: %s", source)
		}
	}
}

func TestWireValueProvenanceRejectsImportedLibraryResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, helper, construction, use string
	}{
		{
			name: "imported UUID value",
			use:  "",
			helper: `package probe
import "github.com/google/uuid"
func generatedValue() string { return uuid.NewString() }
`,
			construction: `wire.Payload{Value: generatedValue()}`,
		},
		{
			name: "sibling helper returning closed enum conversion",
			use:  "",
			helper: `package probe
import (
 "time"
 wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
)
func generatedSequenceType() wire.SequenceType {
 return wire.SequenceType(time.Now().Format(time.RFC3339Nano))
}
`,
			construction: `wire.Sequence{Type: generatedSequenceType()}`,
		},
		{
			name: "global imported output",
			use:  "",
			helper: `package probe
import "github.com/google/uuid"
var generatedValue = uuid.NewString()
`,
			construction: `wire.Payload{Value: generatedValue}`,
		},
		{
			name: "nested imported output",
			use:  "",
			helper: `package probe
import "github.com/google/uuid"
func generatedValue() string { return uuid.NewString() }
`,
			construction: `wire.Payload{Extra: map[string]any{"known": []string{generatedValue()}}}`,
		},
		{
			name:         "library result passed into a helper-owned wire sink",
			construction: "",
			helper: `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func fillWireValue(value string) { _ = wire.Payload{Value: value} }
`,
			use: `package probe
import "github.com/google/uuid"
func request() { fillWireValue(uuid.NewString()) }
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			set := token.NewFileSet()

			use := test.use
			if use == "" {
				use = `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func request() { _ = ` + test.construction + ` }
`
			}

			files := []wireSourceFile{
				parseWireSourceTestFile(t, set, "pkg/probe/helpers.go", test.helper),
				parseWireSourceTestFile(t, set, "pkg/probe/use.go", use),
			}

			err := checkWireSourcePackage(files, set, wireConstructionTestModels())
			if err == nil || !strings.Contains(err.Error(), "unresolved call result") &&
				!strings.Contains(err.Error(), "unresolved library helper") {
				t.Fatalf("unverified library value was accepted or lacked a provenance diagnostic: %v", err)
			}
		})
	}
}

func TestWireValueProvenanceAllowsCallerForwardingIntoHelperSinks(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()

	files := []wireSourceFile{
		parseWireSourceTestFile(t, set, "pkg/probe/helpers.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func fillWireValue(value string) { _ = wire.Payload{Value: value} }
`),
		parseWireSourceTestFile(t, set, "pkg/probe/use.go", `package probe
func Request(caller string) { fillWireValue(caller) }
`),
	}

	err := checkWireSourcePackage(files, set, wireConstructionTestModels())
	if err != nil {
		t.Fatalf("caller-owned value forwarded into a helper sink was rejected: %v", err)
	}
}

func TestWireValueProvenanceRejectsHardcodedHelperSinkArguments(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()

	files := []wireSourceFile{
		parseWireSourceTestFile(t, set, "pkg/probe/helpers.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func fillWireValue(value string) { _ = wire.Payload{Value: value} }
`),
		parseWireSourceTestFile(t, set, "pkg/probe/use.go", `package probe
func request() { fillWireValue("schema-unknown-hardcoded-value") }
`),
	}

	err := checkWireSourcePackage(files, set, wireConstructionTestModels())

	want := `fixed wire value "schema-unknown-hardcoded-value" must be generated from schema`
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("hardcoded value forwarded into a generated wire sink was accepted or lacked a schema diagnostic: %v", err)
	}
}

func TestWireValueProvenancePreservesCallerOwnedPaths(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()

	files := []wireSourceFile{
		parseWireSourceTestFile(t, set, "pkg/probe/helpers.go", `package probe
func pass[T any](value T) T { return value }
func invoke(value func() string) string { return value() }
type values struct{}
func (values) pass(value string) string { return value }
`),
		parseWireSourceTestFile(t, set, "pkg/probe/use.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func Request(caller string, callback func() string) {
 _ = wire.Payload{Value: pass(caller)}
 _ = wire.Payload{Value: values{}.pass(caller)}
 _ = wire.Payload{Value: invoke(callback)}
 _ = wire.Payload{Value: callback()}
}
`),
	}

	err := checkWireSourcePackage(files, set, wireConstructionTestModels())
	if err != nil {
		t.Fatalf("caller-owned helper and callback paths were rejected: %v", err)
	}
}

func TestWireValueProvenanceRejectsRecursiveAndDeepHelpers(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, helpers, value string
	}{
		{
			name: "recursive helper",
			helpers: `func recursive(value string) string { return recursive(value) }
`,
			value: `recursive(caller)`,
		},
		{
			name:    "deep helper chain",
			helpers: deepWireHelperChain(maxWireValueResolutionDepth + 4),
			value:   `wireHelper0(caller)`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			set := token.NewFileSet()
			file := parseWireSourceTestFile(t, set, "pkg/probe/probe.go", `package probe
import (
 wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
 "github.com/google/uuid"
)
`+test.helpers+`func request(caller string) { _ = wire.Payload{Value: `+test.value+`} }
`)

			err := checkWireSourcePackage([]wireSourceFile{file}, set, wireConstructionTestModels())
			if err == nil || !strings.Contains(err.Error(), "wire value helper") &&
				!strings.Contains(err.Error(), "traversal limit") &&
				!strings.Contains(err.Error(), "recursive or unresolved provenance") {
				t.Fatalf("recursive/deep helper provenance was accepted or lacked a diagnostic: %v", err)
			}
		})
	}
}

func deepWireHelperChain(length int) string {
	var helpers strings.Builder

	for index := length - 1; index >= 0; index-- {
		if index == length-1 {
			fmt.Fprintf(&helpers, "func wireHelper%d(value string) string { return uuid.NewString() }\n", index)

			continue
		}

		fmt.Fprintf(&helpers, "func wireHelper%d(value string) string { return wireHelper%d(value) }\n", index, index+1)
	}

	return helpers.String()
}

func parseWireSourceTestFile(t *testing.T, set *token.FileSet, path, content string) wireSourceFile {
	t.Helper()

	file, err := parser.ParseFile(set, path, content, 0)
	if err != nil {
		t.Fatal(err)
	}

	return wireSourceFile{path: path, file: file}
}
