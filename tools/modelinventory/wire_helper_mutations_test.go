package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestWireMutationRejectsCopyAndLocalHelperEscapes(t *testing.T) {
	t.Parallel()

	for _, mutation := range []string{
		`maps.Copy(input.Extra, map[string]any{"brandNewKey": caller})`,
		`maps.Copy(input.Extra, map[string]any{caller: "brandNewValue"})`,
		`clone := maps.Copy[map[string]any, map[string]any]; clone(input.Extra, map[string]any{"brandNewKey": caller})`,
		`copy(input.Values, []string{"brandNewValue"})`,
		`mutate(input.Extra, caller)`,
		`indirect(input.Extra, caller)`,
		`fn := mutate; fn(input.Extra, caller)`,
		`fn := noop; fn = mutate; fn(input.Extra, caller)`,
	} {
		set := token.NewFileSet()

		file, err := parser.ParseFile(set, "probe.go", `package probe
import "maps"
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func mutate(extra map[string]any, caller string) { extra["brandNewKey"] = caller }
func noop(extra map[string]any, caller string) {}
func indirect(extra map[string]any, caller string) { mutate(extra, caller) }
func request(input *wire.Payload, caller string) { `+mutation+` }
`, 0)
		if err != nil {
			t.Fatalf("parse helper mutation: %v", err)
		}

		err = rejectRawGeneratedWireConstructions(file, set, wireConstructionTestModels())
		if err == nil {
			t.Fatalf("unregistered helper mutation accepted: %s", mutation)
		}
	}
}

func TestWireMutationAllowsOpenInputsThroughHelpers(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()

	file, err := parser.ParseFile(set, "probe.go", `package probe
import "maps"
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func mutate(extra map[string]any, caller string) { extra[caller] = caller }
func request(input *wire.Payload, caller string) {
 maps.Copy(input.Extra, map[string]any{wire.RegisteredValue: caller})
 mutate(input.Extra, caller)
}
`, 0)
	if err != nil {
		t.Fatalf("parse open helper mutation: %v", err)
	}

	err = rejectRawGeneratedWireConstructions(file, set, wireConstructionTestModels())
	if err != nil {
		t.Fatalf("caller/generated helper keys rejected: %v", err)
	}
}
