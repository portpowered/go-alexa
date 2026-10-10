package main

import (
	"go/token"
	"strings"
	"testing"
)

func TestRecordingTransportInventory(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, body string
		rejected   bool
	}{
		{"forward original", `_, _ = r.transport.RoundTrip(request)`, false},
		{"wrong transport", `_, _ = other.RoundTrip(request)`, true},
		{"wrong request", `_, _ = r.transport.RoundTrip(other)`, true},
		{"second send", `_, _ = r.transport.RoundTrip(request); _, _ = r.transport.RoundTrip(request)`, true},
		{"mutated route", `request.URL = other; _, _ = r.transport.RoundTrip(request)`, true},
		{"mutated name via body is replay constrained", `request.Body = other; _, _ = r.transport.RoundTrip(request)`, false},
		{"header mutation", `request.Header.Set("X-Test", "value"); _, _ = r.transport.RoundTrip(request)`, true},
		{"request alias", `alias := request; _, _ = r.transport.RoundTrip(alias)`, true},
		{"declared header alias", `var alias = request.Header; alias.Set("X-Test", "value"); _, _ = r.transport.RoundTrip(request)`, true},
		{"request escape", `mutate(request); _, _ = r.transport.RoundTrip(request)`, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			path := strings.Split(wireRecordingDelegate, "#")[0]
			file := parseTestFile(t, fset, path, "package cli\nfunc (r *wireRecorder) RoundTrip(request *http.Request) {"+test.body+"}")

			var violations []string

			checkNetworkInventory(file, fset, path, &violations)

			if (len(violations) > 0) != test.rejected {
				t.Fatalf("unexpected inventory result: %v", violations)
			}
		})
	}
}
