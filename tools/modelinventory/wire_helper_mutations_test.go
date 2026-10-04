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
		`unknown(input.Extra, caller)`,
		`unknownAny([]any{input.Extra})`,
		`unknownAny(struct{ Extra map[string]any }{input.Extra})`,
		`unknownAny(expose(input))`,
		`unknownAny(func() map[string]any { return input.Extra })`,
		`genericHelper[int].mutate(genericHelper[int]{}, input.Extra)`,
		`external.Mutate(input.Extra)`,
		`maps.Copy(input.Extra, map[string]any{caller: "brandNewValue"})`,
		`clone := maps.Copy[map[string]any, map[string]any]; clone(input.Extra, map[string]any{"brandNewKey": caller})`,
		`copy(input.Values, []string{"brandNewValue"})`,
		`delete(input.Extra, "brandNewKey")`,
		`append(input.Values, "brandNewValue")`,
		`alias := mapAlias(input.Extra); alias["brandNewKey"] = caller`,
		`mutate(input.Extra, caller)`,
		`indirect(input.Extra, caller)`,
		`unnamed(nil, input.Extra)`,
		`helper{}.mutate(input.Extra, caller)`,
		`helper.mutate(helper{}, input.Extra, caller)`,
		`variadic(caller, nil, input.Extra)`,
		`fn := helper{}.mutate; fn(input.Extra, caller)`,
		`func(extra map[string]any) { extra["brandNewKey"] = caller }(input.Extra)`,
		`fn := func(extra map[string]any) { extra["brandNewKey"] = caller }; fn(input.Extra)`,
		`fn := mutate; fn(input.Extra, caller)`,
		`fn := noop; fn = mutate; fn(input.Extra, caller)`,
		`fn := noop; fn = unknown; fn(input.Extra, caller)`,
		`fn := unknown; fn = noop; fn(input.Extra, caller)`,
	} {
		set := token.NewFileSet()

		file, err := parser.ParseFile(set, "probe.go", `package probe
import "maps"
import external "example.com/unverified"
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
type mapAlias map[string]any
func mutate(extra map[string]any, caller string) { extra["brandNewKey"] = caller }
func noop(extra map[string]any, caller string) {}
type helper struct{}
func (helper) Mutate(map[string]any) {}
type genericHelper[T any] struct{}
func (genericHelper[T]) mutate(extra map[string]any) { extra["brandNewKey"] = "brandNewValue" }
func expose(input *wire.Payload) map[string]any { return input.Extra }
func (helper) mutate(extra map[string]any, caller string) { extra["brandNewKey"] = caller }
func variadic(caller string, extras ...map[string]any) { extras[0]["brandNewKey"] = caller }
func unnamed(_ map[string]any, extra map[string]any) { extra["brandNewKey"] = "brandNewValue" }
func indirect(extra map[string]any, caller string) { mutate(extra, caller) }
func request(input *wire.Payload, caller string, unknown func(map[string]any, string), unknownAny func(any)) { `+mutation+` }
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
