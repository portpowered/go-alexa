package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestExampleWireInputsRejectRawKnownStatesOnly(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		body  string
		valid bool
	}{
		{body: `_=sdk.ControlLockPayload{State:"LOCKED"}`, valid: false},
		{body: `state:="LOCKED"; _=sdk.ControlLockPayload{State:state}`, valid: false},
		{body: `_=sdk.ControlLockPayload{State:caller}`, valid: true},
		{body: `label:="LOCKED"; _=label`, valid: true},
		{body: `_=sdk.ControlLockPayload{State:wire.LockStateValueLocked}`, valid: true},
	} {
		set := token.NewFileSet()

		file, err := parser.ParseFile(set, "probe.go", `package probe
import sdk "github.com/portpowered/go-alexa/pkg/alexaapimodels"
import wire "github.com/portpowered/go-alexa/pkg/dependencymodels"
func example(caller string) { `+test.body+` }
`, 0)
		if err != nil {
			t.Fatalf("parse example: %v", err)
		}

		owner := primitivePackage{ImportPath: sdkImportPath, Name: "alexaapimodels", Directory: "pkg/alexaapimodels"}

		err = checkExampleInputPrimitives(file, set, "cmd/example.go", owner, map[string]bool{"LOCKED": true})
		if (err == nil) != test.valid {
			t.Fatalf("example validity = %v, want %v: %s", err, test.valid, test.body)
		}
	}
}
