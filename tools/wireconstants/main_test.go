package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"testing"
)

func TestCompatibilityConstantsPreserveAliasAndUntypedStringSemantics(t *testing.T) {
	t.Parallel()

	schema := []byte(`components:
  schemas:
    CapabilityInterface:
      type: string
      x-go-compat-constant-type: CapabilityInterface
      x-go-compat-constant-values:
        TemperatureSensorCapability: TEMPERATURE_SENSOR
    Operation:
      type: string
      enum: [wire.operation]
      x-go-compat-constant-names: [OperationType]
`)

	generated, err := renderCompatibilityConstants([][]byte{schema})
	if err != nil {
		t.Fatalf("render compatibility constants: %v", err)
	}

	source := string(generated) + "\ntype CapabilityInterface = string\n"
	files := token.NewFileSet()

	parsed, err := parser.ParseFile(files, "generated.go", source, 0)
	if err != nil {
		t.Fatalf("parse generated constants: %v", err)
	}

	var checker types.Config

	checked, err := checker.Check("example.com/generated", files, []*ast.File{parsed}, nil)
	if err != nil {
		t.Fatalf("type-check generated constants: %v", err)
	}

	for name, expected := range map[string]*types.Basic{
		"TemperatureSensorCapability": types.Typ[types.String],
		"OperationType":               types.Typ[types.UntypedString],
	} {
		constant := checked.Scope().Lookup(name)
		if constant == nil || !types.Identical(constant.Type(), expected) {
			t.Fatalf("constant %s = %v; want %s semantics", name, constant, expected)
		}
	}
}

func TestRenderCompatibilityConstantsKeepsPublicConstantsUntyped(t *testing.T) {
	t.Parallel()

	schema := []byte(`openapi: 3.0.3
components:
  schemas:
    Operation:
      type: string
      enum: [wire.operation]
      x-go-compat-constant-names: [OperationType]
`)

	generated, err := renderCompatibilityConstants([][]byte{schema})
	if err != nil {
		t.Fatalf("render compatibility constants: %v", err)
	}

	if !strings.Contains(string(generated), "OperationType = \"wire.operation\"") {
		t.Fatalf("generated output does not preserve an untyped literal constant:\n%s", generated)
	}
}

func TestRenderCompatibilityConstantsRejectsMissingEnumValues(t *testing.T) {
	t.Parallel()

	schema := []byte(`openapi: 3.0.3
components:
  schemas:
    Operation:
      type: string
      enum: [wire.operation]
      x-go-compat-constant-names: [First, Second]
`)

	_, err := renderCompatibilityConstants([][]byte{schema})
	if err == nil || !strings.Contains(err.Error(), "enum values") {
		t.Fatalf("expected mismatched schema metadata to fail, got %v", err)
	}
}

func TestWireKeysFollowSchemaPropertyNames(t *testing.T) {
	t.Parallel()

	schema := []byte(`components:
  schemas:
    Directive:
      type: object
      properties:
        renamedHeader:
          type: object
          x-go-wire-key-constant-name: EventHeaderKey
          properties:
            renamedMessageId:
              type: string
              x-go-wire-key-constant-name: EventMessageIDKey
`)

	generated, err := renderCompatibilityConstants([][]byte{schema})
	if err != nil {
		t.Fatalf("render property keys: %v", err)
	}

	files := token.NewFileSet()

	parsed, err := parser.ParseFile(files, "generated.go", generated, 0)
	if err != nil {
		t.Fatalf("parse generated keys: %v", err)
	}

	var checker types.Config

	checked, err := checker.Check("example.com/generated", files, []*ast.File{parsed}, nil)
	if err != nil {
		t.Fatalf("type-check generated keys: %v", err)
	}

	for name, value := range map[string]string{
		"EventHeaderKey":    "renamedHeader",
		"EventMessageIDKey": "renamedMessageId",
	} {
		constant, valid := checked.Scope().Lookup(name).(*types.Const)
		if !valid || constant.Val().ExactString() != strconv.Quote(value) {
			t.Fatalf("constant %s = %v; want schema property %s", name, constant, value)
		}
	}
}
