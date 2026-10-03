package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestSplitMovesEveryGraphQLWireTypeAndPreservesCompatibilityAliases(t *testing.T) {
	t.Parallel()

	input := []byte(`package graphql

import (
	"context"
	"encoding/json"
	"github.com/Khan/genqlient/graphql"
)

type State string
const (
	StateOn State = "ON"
	List_Operation = "query List { item { state } }"
)
var AllState = []State{StateOn}

// This exported type is deliberately unused by the operation below.
type LegacyWire struct {
	Payload struct { Name string ` + "`json:\"name\"`" + ` } ` + "`json:\"payload\"`" + `
}

type __QueryInput struct { ID string ` + "`json:\"id\"`" + ` }
type QueryResponse struct { Item LegacyWire ` + "`json:\"item\"`" + ` }
func (v *QueryResponse) UnmarshalJSON(data []byte) error { type alias QueryResponse; return json.Unmarshal(data, (*alias)(v)) }
func __marshalQueryResponse(v *QueryResponse) ([]byte, error) { return json.Marshal(v) }
func List(ctx context.Context, client graphql.Client) (*QueryResponse, error) {
	return nil, nil
}
`)

	result, err := split(input)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"State", "LegacyWire", "QueryResponse", "GraphQLQueryInput"} {
		if !strings.Contains(string(result.models), "type "+name+" ") {
			t.Errorf("models output is missing generated type %s", name)
		}
	}

	if !strings.Contains(string(result.models), "UnmarshalJSON") || !strings.Contains(string(result.models), "__marshalQueryResponse") {
		t.Error("models output must retain custom JSON methods and generated marshal helpers")
	}

	if strings.Contains(string(result.operations), "type LegacyWire") || strings.Contains(string(result.operations), "type QueryResponse") {
		t.Error("operation output retained generated wire model declarations")
	}

	for _, name := range []string{"List_Operation", "func List("} {
		if !strings.Contains(string(result.operations), name) {
			t.Errorf("operation output is missing %q", name)
		}
	}

	aliasNames := []string{
		"type LegacyWire = alexamodels.LegacyWire",
		"type __QueryInput = alexamodels.GraphQLQueryInput",
		"const StateOn = alexamodels.StateOn",
		"var AllState",
	}
	for _, name := range aliasNames {
		if !strings.Contains(string(result.aliases), name) {
			t.Errorf("compatibility aliases are missing %q", name)
		}
	}

	_, err = parser.ParseFile(token.NewFileSet(), "operations.go", result.operations, parser.AllErrors)
	if err != nil {
		t.Errorf("operation output is not valid Go: %v", err)
	}

	_, err = parser.ParseFile(token.NewFileSet(), "models.go", result.models, parser.AllErrors)
	if err != nil {
		t.Errorf("model output is not valid Go: %v", err)
	}

	_, err = parser.ParseFile(token.NewFileSet(), "aliases.go", result.aliases, parser.AllErrors)
	if err != nil {
		t.Errorf("alias output is not valid Go: %v", err)
	}
}
