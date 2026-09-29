package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

type wireCallsiteTestCase struct {
	name, source, want string
}

func TestWireCallsiteGate(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "generated route and key", want: "",
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
			func send() {
				path := apiroutes.PathListRestEndpoints
				path = apiroutes.PathGetRestEndpoint
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "generated directive channel", want: "",
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsHandwrittenHeaders(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
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
		{
			name: "direct request header map write",
			source: `package rest
			func send() { req.Header["X-Undeclared"] = []string{"value"} }`,
			want: "request header map key must use a schema-generated Header constant",
		},
		{
			name: "aliased request header map write",
			source: `package rest
			func send() { headers := req.Header; headers["X-Undeclared"] = []string{"value"} }`,
			want: "request header map key must use a schema-generated Header constant",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, test)
		})
	}
}

func assertWireCallsiteGateCase(t *testing.T, test wireCallsiteTestCase) {
	t.Helper()

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
}

func TestNetworkInventoryGate(t *testing.T) {
	t.Parallel()

	tests := []struct{ name, source, want string }{
		{"new Client.Post", `package rest
		func send() { client.Post("https://example.com", "application/json", nil) }`, "unregistered outbound network primitive Post"},
		{"new Client.Do", `package rest
		func send() { client.Do(req) }`, "unregistered outbound network primitive Do"},
		{"new request constructor", `package rest
		func send() { http.NewRequest("POST", "https://example.com", nil) }`, "unregistered outbound network primitive NewRequest"},
		{"new network import", `package rest
		import "net/http"
		func send() {}`, "unregistered network import"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()

			file, err := parser.ParseFile(fset, "pkg/dependencies/rest/new.go", test.source, 0)
			if err != nil {
				t.Fatal(err)
			}

			var violations []string

			checkNetworkInventory(file, fset, "pkg/dependencies/rest/new.go", &violations)

			if got := strings.Join(violations, "\n"); !strings.Contains(got, test.want) {
				t.Fatalf("gate result %q, want %q", got, test.want)
			}
		})
	}
}

func TestNetworkInventoryRejectsReassignedRequest(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package alexa
	func (c *Client) Connect() {
		req, err := http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream, apiroutes.ChannelDirectivesAddress, nil)
		_ = err
		req = otherRequest
		c.client.Do(req)
	}`

	file, err := parser.ParseFile(fset, "pkg/alexa/http2.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

	var violations []string

	checkNetworkInventory(file, fset, "pkg/alexa/http2.go", &violations)

	if got := strings.Join(violations, "\n"); !strings.Contains(got, "request was reassigned") {
		t.Fatalf("gate result %q did not reject reassignment", got)
	}
}

func TestNetworkInventoryRejectsRequestRouteMutation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mutation string
		want     string
	}{
		{name: "path", mutation: "req.URL.Path = apiroutes.PathGetRestEndpoint", want: "request was reassigned or mutated"},
		{name: "parenthesized path", mutation: "(req).URL.Path = apiroutes.PathGetRestEndpoint", want: "request was reassigned or mutated"},
		{name: "method", mutation: "req.Method = apiroutes.MethodGetRestEndpoint", want: "request was reassigned or mutated"},
		{name: "url alias", mutation: "route := req.URL; route.Path = apiroutes.PathGetRestEndpoint", want: "request was reassigned or mutated"},
		{name: "request alias", mutation: "other := req; other.Method = apiroutes.MethodGetRestEndpoint", want: "request was reassigned or mutated"},
		{
			name: "reassigned alias", mutation: "other := req; other.URL.Path = apiroutes.PathGetRestEndpoint; other = unrelated",
			want: "request was reassigned or mutated",
		},
		{name: "var alias", mutation: "var other = req; other.URL.RawPath = apiroutes.PathGetRestEndpoint", want: "request was reassigned or mutated"},
		{
			name: "composite alias", mutation: "holder := struct { request *http.Request }{request: req}; holder.request.URL.Path = apiroutes.PathGetRestEndpoint",
			want: "request was reassigned or mutated",
		},
		{
			name: "field store", mutation: "holder := &routeHolder{}; holder.request = req; holder.request.URL.Path = apiroutes.PathGetRestEndpoint",
			want: "request was reassigned or mutated",
		},
		{
			name: "indexed store", mutation: "holders := make([]*http.Request, 1); holders[0] = req; holders[0].URL.Path = apiroutes.PathGetRestEndpoint",
			want: "request was reassigned or mutated",
		},
		{name: "request helper", mutation: "mutateRoute(req)", want: "request or URL escaped"},
		{name: "parenthesized helper", mutation: "mutateRoute((req))", want: "request or URL escaped"},
		{
			name: "composite helper", mutation: "mutateRoute(struct { request *http.Request }{request: req})",
			want: "request or URL escaped",
		},
		{
			name: "composite receiver", mutation: "routeHolder{request: req}.mutate()",
			want: "request or URL escaped",
		},
		{name: "url helper", mutation: "mutateRoute(req.URL)", want: "request or URL escaped"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			source := `package rest
			func (c *Client) doRequest() {
				req, err := http.NewRequestWithContext(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints, nil)
				_ = err
				` + test.mutation + `
				c.httpClient.Do(req)
			}`

			file, err := parser.ParseFile(fset, "pkg/dependencies/rest/client.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}

			var violations []string

			checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

			if got := strings.Join(violations, "\n"); !strings.Contains(got, test.want) {
				t.Fatalf("gate result %q did not reject route mutation", got)
			}
		})
	}
}
