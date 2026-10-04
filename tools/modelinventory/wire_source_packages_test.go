package main

import (
	"go/parser"
	"go/token"
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

func parseWireSourceTestFile(t *testing.T, set *token.FileSet, path, content string) wireSourceFile {
	t.Helper()

	file, err := parser.ParseFile(set, path, content, 0)
	if err != nil {
		t.Fatal(err)
	}

	return wireSourceFile{path: path, file: file}
}
