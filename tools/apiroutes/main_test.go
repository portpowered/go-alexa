package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestWireCallsiteGate(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{
			name: "generated route and key",
			source: `package rest
			func send() {
				query := url.Values{}
				query.Set(alexamodels.QueryParamOwner, "caller")
				request(apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints)
			}`,
		},
		{
			name: "mismatched method and path",
			source: `package rest
			func send() { request(apiroutes.MethodListRestEndpoints, apiroutes.PathGetRestEndpoint) }`,
			want: "not paired with its generated path",
		},
		{
			name: "handwritten query key",
			source: `package rest
			func send() { query := url.Values{}; query.Set("owner", "caller") }`,
			want: "schema-generated QueryParam constant",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, test.name+".go", test.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			var violations []string
			checkWireCallsites(file, fset, nil, map[string]bool{"ListRestEndpoints": true}, &violations)
			got := strings.Join(violations, "\n")
			if test.want == "" && got != "" || test.want != "" && !strings.Contains(got, test.want) {
				t.Fatalf("gate result %q, want %q", got, test.want)
			}
		})
	}
}
