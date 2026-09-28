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
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints, nil, nil)
			}`,
		},
		{
			name: "mismatched method and path",
			source: `package rest
			func send() { c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathGetRestEndpoint, nil, nil) }`,
			want: "not paired with its generated route",
		},
		{
			name: "handwritten query key",
			source: `package rest
			func send() { query := url.Values{}; query.Set("owner", "caller") }`,
			want: "schema-generated QueryParam constant",
		},
		{
			name: "unschematized endpoint",
			source: `package rest
			func send() { c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, "/unlisted", nil, nil) }`,
			want: "not paired with its generated route",
		},
		{
			name: "handwritten method",
			source: `package rest
			func send() { c.doJSONRequest(ctx, "POST", apiroutes.PathListRestEndpoints, nil, nil) }`,
			want: "method must be a generated OpenAPI operation",
		},
		{
			name: "decoy matching path",
			source: `package rest
			func send() { _ = apiroutes.PathListRestEndpoints; c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathGetRestEndpoint, nil, nil) }`,
			want: "not paired with its generated route",
		},
		{
			name: "reassigned path after matching declaration",
			source: `package rest
			func send() { path := apiroutes.PathListRestEndpoints; path = apiroutes.PathGetRestEndpoint; c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil) }`,
			want: "not paired with its generated route",
		},
		{
			name: "generated directive channel",
			source: `package alexa
			func send() { http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream, base+apiroutes.ChannelDirectivesAddress, nil) }`,
		},
		{
			name: "unschematized channel",
			source: `package alexa
			func send() { http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream, base+"/other-channel", nil) }`,
			want: "not paired with its generated route or channel",
		},
		{
			name: "OpenAPI path cannot stand in for AsyncAPI channel",
			source: `package alexa
			func send() { http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream, base+apiroutes.PathOpenDirectiveStream, nil) }`,
			want: "not paired with its generated route or channel",
		},
		{
			name: "handwritten request header",
			source: `package rest
			func send() { req.Header.Set("X-Undeclared", "value") }`,
			want: "request header name must use a schema-generated Header constant",
		},
		{
			name: "handwritten added request header",
			source: `package rest
			func send() { req.Header.Add("X-Undeclared", "value") }`,
			want: "request header name must use a schema-generated Header constant",
		},
		{
			name: "handwritten custom request header",
			source: `package rest
			func send() { customHeaders["X-Undeclared"] = "value" }`,
			want: "custom request header name must use a schema-generated Header constant",
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
			checkWireCallsites(file, fset, nil, map[string]bool{"ListRestEndpoints": true, "OpenDirectiveStream": true}, &violations)
			got := strings.Join(violations, "\n")
			if test.want == "" && got != "" || test.want != "" && !strings.Contains(got, test.want) {
				t.Fatalf("gate result %q, want %q", got, test.want)
			}
		})
	}
}
