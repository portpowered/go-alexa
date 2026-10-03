package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestWireConstructionRejectsUnregisteredFixedValuesAndAliases(t *testing.T) {
	t.Parallel()

	for _, construction := range []string{
		`wire.Payload{Value: "brandNewUnregisteredValue"}`,
		`wire.Payload{Value: "brandNew" + "UnregisteredValue"}`,
		`wire.Payload{Value: string("brandNewUnregisteredValue")}`,
		`wire.Payload{Value: raw}`,
		`wire.Payload{Value: indirect}`,
		`wire.Payload{Extra: map[string]any{"brandNewUnregisteredKey": caller}}`,
	} {
		set := token.NewFileSet()

		file, err := parser.ParseFile(set, "probe.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
const raw = "brandNewUnregisteredValue"
func request(caller string) { indirect := raw; _ = `+construction+` }
`, 0)
		if err != nil {
			t.Fatalf("parse construction probe: %v", err)
		}

		err = rejectRawGeneratedWireConstructions(file, set, "pkg/probe.go", wireConstructionTestModels())
		if err == nil {
			t.Fatalf("unregistered fixed value/key accepted: %s", construction)
		}
	}
}

func TestWireConstructionAllowsCallerInputsAndGeneratedConstants(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()

	file, err := parser.ParseFile(set, "probe.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func request(caller string) { _ = wire.Payload{Value: caller}; _ = wire.Payload{Value: wire.RegisteredValue} }
`, 0)
	if err != nil {
		t.Fatalf("parse caller probe: %v", err)
	}

	err = rejectRawGeneratedWireConstructions(file, set, "pkg/probe.go", wireConstructionTestModels())
	if err != nil {
		t.Fatalf("caller or generated value rejected: %v", err)
	}
}

func wireConstructionTestModels() map[string]generatedModel {
	var payload generatedModel

	payload.Name = "Payload"
	payload.File = "pkg/dependencymodels/payload.gen.go"
	payload.Generated = true

	return map[string]generatedModel{"Payload": payload}
}

func TestWireMutationRejectsNewFixedValuesAfterInitialization(t *testing.T) {
	t.Parallel()

	for _, mutation := range []string{
		`var payload wire.Payload; payload.Value = "unregistered"`,
		`payload := wire.Payload{}; payload.Value = "unregistered"`,
		`payload := &wire.Payload{}; alias := payload; alias.Value = "unregistered"`,
		`input.Value = "unregistered"`,
	} {
		set := token.NewFileSet()

		file, err := parser.ParseFile(set, "probe.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func request(input *wire.Payload) { `+mutation+` }
`, 0)
		if err != nil {
			t.Fatalf("parse mutation probe: %v", err)
		}

		err = rejectRawGeneratedWireConstructions(file, set, "pkg/probe.go", wireConstructionTestModels())
		if err == nil {
			t.Fatalf("unregistered mutation accepted: %s", mutation)
		}
	}
}

func TestWireConstructionDoesNotTreatFieldNamesOrSiblingAssignmentsAsValues(t *testing.T) {
	t.Parallel()

	set := token.NewFileSet()

	file, err := parser.ParseFile(set, "probe.go", `package probe
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
const Value = "not a wire value"
func request(caller string) {
 local, diagnostic := caller, "diagnostic"
 _ = diagnostic
 _ = wire.Payload{Nested: wire.Payload{Value: local}}
}
`, 0)
	if err != nil {
		t.Fatalf("parse scope probe: %v", err)
	}

	err = rejectRawGeneratedWireConstructions(file, set, "pkg/probe.go", wireConstructionTestModels())
	if err != nil {
		t.Fatalf("non-wire strings were treated as wire values: %v", err)
	}
}
