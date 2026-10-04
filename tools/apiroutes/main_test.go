package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
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
			func (c *Client) send() { http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream, c.baseURL+apiroutes.ChannelDirectivesAddress, nil) }`,
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
	tests = append(tests, routeMutationCases()...)
	tests = append(tests, routeShadowCases()...)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, test)
		})
	}
}

func routeMutationCases() []wireCallsiteTestCase {
	return []wireCallsiteTestCase{
		{
			name: "compound path suffix",
			source: `package rest
			func send() {
				path := apiroutes.PathListRestEndpoints
				path += suffix
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "binary path suffix",
			source: `package rest
			func send() {
				path := apiroutes.PathListRestEndpoints
				path = path + suffix
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "direct path suffix",
			source: `package rest
			func send() {
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints + suffix, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "untrusted route prefix",
			source: `package rest
			func send() {
				basePath := configuredPath
				path := basePath + apiroutes.PathListRestEndpoints
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "wrapped path",
			source: `package rest
			func send() {
				path := apiroutes.PathListRestEndpoints
				path = normalize(path)
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "formatted route with extra argument",
			source: `package rest
			func send() {
				path := fmt.Sprintf(apiroutes.PathListRestEndpoints, suffix)
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "formatted template with extra argument",
			source: `package rest
			func send() {
				path := fmt.Sprintf(apiroutes.PathGetRestEndpoint, endpointID, suffix)
				c.doJSONRequest(ctx, apiroutes.MethodGetRestEndpoint, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "formatted channel with extra argument",
			source: `package alexa
			func send() {
				http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream,
					fmt.Sprintf("https://%s%s", prefix, authority, apiroutes.ChannelDirectivesAddress), nil)
			}`,
			want: "not paired with its generated route or channel",
		},
		{
			name: "formatted server with extra argument",
			source: `package rest
			func send() {
				path := fmt.Sprintf(apiroutes.ServerExchangeRefreshTokenForCookies, domain, suffix) +
					apiroutes.PathExchangeRefreshTokenForCookies
				c.doJSONRequest(ctx, apiroutes.MethodExchangeRefreshTokenForCookies, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "generated route with encoded query", want: "",
			source: `package rest
			func send() {
				params := url.Values{}
				params.Set(alexamodels.QueryParamExpand, "all")
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
		},
	}
}

func routeShadowCases() []wireCallsiteTestCase {
	return []wireCallsiteTestCase{
		{
			name: "conditional route replacement",
			source: `package rest
			func send() {
				path := dynamicPath
				if useList { path = apiroutes.PathListRestEndpoints }
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "different generated routes across branches",
			source: `package rest
			func send(useList bool) {
				var path string
				if useList { path = apiroutes.PathListRestEndpoints } else { path = apiroutes.PathGetRestEndpoint }
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "route pointer replacement",
			source: `package rest
			func send() {
				path := apiroutes.PathListRestEndpoints
				pointer := &path
				*pointer = dynamicPath
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "route pointer helper escape",
			source: `package rest
			func send() {
				path := apiroutes.PathListRestEndpoints
				mutatePath(&path, dynamicPath)
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "untrusted directive authority",
			source: `package alexa
			func send() {
				http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream,
					fmt.Sprintf("https://%s%s", untrustedAuthority, apiroutes.ChannelDirectivesAddress), nil)
			}`,
			want: "not paired with its generated route or channel",
		},
		{
			name: "untrusted REST authority",
			source: `package rest
			func send() {
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints,
					fmt.Sprintf("https://%s%s", untrustedAuthority, apiroutes.PathListRestEndpoints), nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "shadowed route variable",
			source: `package rest
			func send() {
				path := apiroutes.PathGetRestEndpoint
				{ path := apiroutes.PathListRestEndpoints; _ = path }
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
	}
}

func TestWireCallsiteGateRejectsQueryMapKeys(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "direct map write",
			source: `package rest
			func send() {
				params := url.Values{}
				params["handwritten"] = []string{"x"}
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map key must use a schema-generated QueryParam constant",
		},
		{
			name: "aliased map write",
			source: `package rest
			func send() {
				params := url.Values{}
				alias := params
				alias["handwritten"] = []string{"x"}
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map key must use a schema-generated QueryParam constant",
		},
		{
			name: "composite map key",
			source: `package rest
			func send() {
				params := url.Values{"handwritten": {"x"}}
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map key must use a schema-generated QueryParam constant",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsQueryMapHelperEscape(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "query map helper escape",
		source: `package rest
		func fill(params url.Values) { params.Set("unknown", "x") }
		func send() {
			params := url.Values{}
			fill(params)
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "query parameter map must not escape to an unverified helper",
	})
}

func TestWireCallsiteGateRejectsQueryMapMethodValues(t *testing.T) {
	t.Parallel()

	for _, method := range []string{"Set", "Add"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, wireCallsiteTestCase{
				name: "query map method value",
				source: `package rest
				func send() {
					params := url.Values{}
					set := params.` + method + `
					set("raw", "unmodeled")
					path := apiroutes.PathListRestEndpoints
					path += "?" + params.Encode()
					c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
				}`,
				want: "schema-keyed map method value must not be aliased",
			})
		})
	}
}

func TestWireCallsiteGateRejectsParenthesizedQueryKey(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "parenthesized query receiver",
		source: `package rest
		func send() {
			params := url.Values{}
			(params).Set("raw", "x")
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "query parameter key must use a schema-generated QueryParam constant",
	})
}

func TestWireCallsiteGateAllowsQueryMapLength(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "query map length",
		source: `package rest
		func send() {
			params := url.Values{}
			params.Set(alexamodels.QueryParamExpand, "all")
			if len(params) > 0 {
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}
		}`,
		want: "",
	})
}

func TestWireCallsiteGateRejectsShadowedLengthHelper(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "shadowed length helper",
		source: `package rest
		func send() {
			len := func(values url.Values) { values.Set("raw", "x") }
			params := url.Values{}
			len(params)
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "query parameter map must not escape to an unverified helper",
	})
}

func TestWireCallsiteGateRejectsShadowedUntrustedQueryMap(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "shadowed untrusted query map",
		source: `package rest
		func parseQuery(raw string) url.Values { values, _ := url.ParseQuery(raw); return values }
		func send(raw string) {
			params := parseQuery(raw)
			{ params := url.Values{}; params.Set(alexamodels.QueryParamExpand, "all"); _ = params }
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "not paired with its generated route",
	})
}

func TestWireCallsiteGateRejectsParsedQueryMap(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "parsed untrusted query map",
		source: `package rest
		func send(raw string) {
			params, _ := url.ParseQuery(raw)
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "not paired with its generated route",
	})
}

func TestWireCallsiteGateRejectsReassignedQueryMap(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "reassigned untrusted query map",
		source: `package rest
		func parseQuery(raw string) url.Values { values, _ := url.ParseQuery(raw); return values }
		func send(raw string) {
			params := url.Values{}
			params = parseQuery(raw)
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "not paired with its generated route",
	})
}

func TestWireCallsiteGateRejectsQueryMapPointerMutation(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "query map pointer mutation",
		source: `package rest
		func send(raw string) {
			params := url.Values{}
			pointer := &params
			*pointer, _ = url.ParseQuery(raw)
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "not paired with its generated route",
	})
}

func TestWireCallsiteGateRejectsDirectTransportHelperBypass(t *testing.T) {
	t.Parallel()

	for _, helper := range []string{"doRequest", "doRequestWithFullURL", "doUnauthenticatedRequest"} {
		t.Run(helper, func(t *testing.T) {
			t.Parallel()

			source := `package rest
			func send() {
				method := selectedMethod()
				path := selectedPath()
				c.` + helper + `(ctx, method, path, nil)
			}`
			assertWireCallsiteGateCase(t, wireCallsiteTestCase{
				name: helper, source: source, want: "method must be a generated OpenAPI operation",
			})
		})
	}
}

func TestWireCallsiteGateRejectsTransportMethodValues(t *testing.T) {
	t.Parallel()

	for _, helper := range []string{"doRequest", "doRequestWithFullURL", "doUnauthenticatedRequest", "doJSONRequest"} {
		t.Run(helper, func(t *testing.T) {
			t.Parallel()

			source := `package rest
			func send() {
				transport := c.` + helper + `
				transport(ctx, selectedMethod(), selectedPath(), nil)
			}`
			assertWireCallsiteGateCase(t, wireCallsiteTestCase{
				name: helper, source: source, want: "transport helper method value must not be aliased",
			})
		})
	}
}

func TestWireCallsiteGateRejectsPackageTransportMethodValue(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "package transport method value",
		source: `package rest
		var invokeRequest = (*Client).doRequest
		func send(c *Client) { invokeRequest(c, ctx, selectedMethod(), selectedPath(), nil) }`,
		want: "transport helper method value must not be aliased",
	})
}

func TestWireCallsiteGateRejectsCustomHeaderHelperEscapes(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "returned map",
			source: `package rest
			func rawHeaders() map[string]string { return map[string]string{"Cookie": "x"} }
			func send() {
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, rawHeaders(), nil, true)
			}`,
			want: "custom header map must be constructed from generated Header keys at this call site",
		},
		{
			name: "mutating helper",
			source: `package rest
			func fill(headers map[string]string) { headers["Cookie"] = "x" }
			func send() {
				headers := make(map[string]string)
				fill(headers)
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "custom header map must not escape to an unverified helper",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsAggregatedHeaderMapEscape(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "aggregated header map helper escape",
		source: `package rest
		type headerHolder struct{ values map[string]string }
		func fill(h headerHolder) { h.values["Cookie"] = "raw" }
		func send() {
			headers := map[string]string{apiroutes.HeaderCookie: "safe"}
			fill(headerHolder{values: headers})
			c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
				apiroutes.PathListRestEndpoints, nil, headers, nil, true)
		}`,
		want: "custom header map must not escape to an unverified helper",
	})
}

func TestWireCallsiteGateRejectsHeaderMapAggregateStorage(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "struct field storage",
			source: `package rest
			func send() {
				headers := map[string]string{apiroutes.HeaderCookie: "safe"}
				holder := struct{ values map[string]string }{headers}
				holder.values["Cookie"] = "raw"
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "custom header map must not escape into aggregate storage",
		},
		{
			name: "indexed storage",
			source: `package rest
			func send() {
				headers := map[string]string{apiroutes.HeaderCookie: "safe"}
				holders := []map[string]string{headers}
				holders[0]["Cookie"] = "raw"
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "custom header map must not escape into aggregate storage",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsQueryMapAggregateStorage(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "query struct field storage",
			source: `package rest
			func send() {
				params := url.Values{}
				holder := struct{ values url.Values }{params}
				holder.values["raw"] = []string{"x"}
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map must not escape into aggregate storage",
		},
		{
			name: "query indexed storage",
			source: `package rest
			func send() {
				params := url.Values{}
				holders := []url.Values{params}
				holders[0]["raw"] = []string{"x"}
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map must not escape into aggregate storage",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsGlobalSchemaMapStorage(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "global header map",
			source: `package rest
			var shared map[string]string
			func send() {
				headers := map[string]string{apiroutes.HeaderCookie: "safe"}
				shared = headers
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "custom header map must not escape into aggregate storage",
		},
		{
			name: "global query map",
			source: `package rest
			var shared url.Values
			func send() {
				params := url.Values{}
				shared = params
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map must not escape into aggregate storage",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsPackageHeaderMap(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "package header map",
		source: `package rest
		var headers = map[string]string{"Cookie": "x"}
		func send() {
			c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
				apiroutes.PathListRestEndpoints, nil, headers, nil, true)
		}`,
		want: "custom header map must be constructed from generated Header keys at this call site",
	})
}

func TestWireCallsiteGateRejectsShadowedHeaderMap(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "shadowed header map",
		source: `package rest
		func send() {
			headers := map[string]string{"Cookie": "raw"}
			{ headers := map[string]string{apiroutes.HeaderCookie: "generated"}; _ = headers }
			c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
				apiroutes.PathListRestEndpoints, nil, headers, nil, true)
		}`,
		want: "custom request header map key must use a schema-generated Header constant",
	})
}

func TestWireCallsiteGateRejectsHeaderMapPointerMutation(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "header map pointer mutation",
		source: `package rest
		func send() {
			headers := map[string]string{apiroutes.HeaderCookie: "generated"}
			pointer := &headers
			*pointer = map[string]string{"Cookie": "raw"}
			c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
				apiroutes.PathListRestEndpoints, nil, headers, nil, true)
		}`,
		want: "custom header map must be constructed from generated Header keys at this call site",
	})
}

func TestWireCallsiteGateRejectsParenthesizedHeaderKeys(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "parenthesized custom header map",
			source: `package rest
			func send() {
				headers := map[string]string{apiroutes.HeaderCookie: "generated"}
				(headers)["Cookie"] = "raw"
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "request header map key must use a schema-generated Header constant",
		},
		{
			name: "parenthesized request header index",
			source: `package rest
			func send() { (req.Header)["Cookie"] = []string{"raw"} }`,
			want: "request header map key must use a schema-generated Header constant",
		},
		{
			name: "parenthesized request header method",
			source: `package rest
			func send() { (req.Header).Set("X-Undeclared", "raw") }`,
			want: "request header name must use a schema-generated Header constant",
		},
		{
			name: "converted request header method",
			source: `package rest
			func send() { http.Header(req.Header).Set("X-Undeclared", "raw") }`,
			want: "request header name must use a schema-generated Header constant",
		},
		{
			name: "converted request header alias",
			source: `package rest
			func send() { headers := http.Header(req.Header); headers.Set("X-Undeclared", "raw") }`,
			want: "request header name must use a schema-generated Header constant",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsShadowedPackages(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "generated package qualifier",
			source: `package rest
			type routeValues struct{ MethodListRestEndpoints, PathListRestEndpoints string }
			func send(apiroutes routeValues) {
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints, nil, nil)
			}`,
			want: "method must be a generated OpenAPI operation",
		},
		{
			name: "formatter qualifier",
			source: `package rest
			type routeFmt struct{}
			func (routeFmt) Sprintf(_ string, _ ...any) string { return "/v2/evil" }
			func send(fmt routeFmt) {
				path := fmt.Sprintf(apiroutes.PathGetRestEndpoint, "id")
				c.doJSONRequest(ctx, apiroutes.MethodGetRestEndpoint, path, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
		{
			name: "query constant qualifier",
			source: `package rest
			func send(alexamodels struct{ QueryParamExpand string }) {
				params := url.Values{}
				params.Set(alexamodels.QueryParamExpand, "raw")
			}`,
			want: "query parameter key must use a schema-generated QueryParam constant",
		},
		{
			name: "header constant qualifier",
			source: `package rest
			func send(apiroutes struct{ HeaderCookie string }) { req.Header.Set(apiroutes.HeaderCookie, "raw") }`,
			want: "request header name must use a schema-generated Header constant",
		},
		{
			name: "shadowed event authority",
			source: `package alexa
			type fakeClient struct{ authority string }
			func send(c fakeClient) {
				http.NewRequestWithContext(ctx, apiroutes.MethodOpenDirectiveStream,
					fmt.Sprintf("https://%s%s", c.authority, apiroutes.ChannelDirectivesAddress), nil)
			}`,
			want: "not paired with its generated route or channel",
		},
		{
			name: "shadowed REST base URL",
			source: `package rest
			type fakeClient struct{ baseURL string }
			func send(c fakeClient) {
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints,
					c.baseURL+apiroutes.PathListRestEndpoints, nil, nil)
			}`,
			want: "not paired with its generated route",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsCounterfeitImports(t *testing.T) {
	t.Parallel()

	for _, alias := range []string{"apiroutes", "alexamodels", "fmt", "http", "url"} {
		t.Run(alias, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, wireCallsiteTestCase{
				name: "counterfeit import",
				source: `package rest
				import ` + alias + ` "example.com/unverified/routes"
				func send() {}`,
				want: "protected package alias",
			})
		})
	}
}

func TestWireCallsiteGateAcceptsExactPathImportAliases(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package rest
	import (
		format "fmt"
		values "net/url"
		web "net/http"
		models "github.com/portpowered/go-alexa/pkg/dependencymodels"
		routes "github.com/portpowered/go-alexa/pkg/internal/apiroutes"
	)
	func (c *Client) ExchangeRefreshTokenForCookies(ctx context.Context, domain string) {
		query := values.Values{}
		query.Set(models.QueryParamOwner, "~caller")
		headers := web.Header{}
		headers.Set(routes.HeaderCsrf, "token")
		target := format.Sprintf(routes.ServerExchangeRefreshTokenForCookies, domain) + routes.PathExchangeRefreshTokenForCookies
		req, err := web.NewRequestWithContext(ctx, routes.MethodExchangeRefreshTokenForCookies, target, nil)
		_ = err
		c.httpClient.Do(req)
	}`
	file := parseTestFile(t, fset, "pkg/dependencies/rest/auth.go", source)

	var violations []string

	checkWireCallsites(file, fset, nil, map[string]string{
		"ExchangeRefreshTokenForCookies": "/ap/exchangetoken/cookies",
	}, 1, &violations)
	checkNetworkInventory(file, fset, "pkg/dependencies/rest/auth.go", &violations)

	if got := strings.Join(violations, "\n"); got != "" {
		t.Fatalf("exact import aliases were rejected: %s", got)
	}
}

func TestWireCallsiteGateRejectsAliasedImportCounterfeitsAndShadows(t *testing.T) {
	t.Parallel()

	tests := []struct{ name, source, want string }{
		{
			name: "wrong path route alias",
			source: `package rest
			import routes "example.com/fake/routes"
			func send() { c.doJSONRequest(ctx, routes.MethodListRestEndpoints, routes.PathListRestEndpoints, nil, nil) }`,
			want: "method must be a generated OpenAPI operation",
		},
		{
			name: "local route alias shadow",
			source: `package rest
			import routes "github.com/portpowered/go-alexa/pkg/internal/apiroutes"
			type routeValues struct{ MethodListRestEndpoints, PathListRestEndpoints string }
			func send() { { routes := routeValues{}; c.doJSONRequest(ctx, routes.MethodListRestEndpoints, routes.PathListRestEndpoints, nil, nil) } }`,
			want: "method must be a generated OpenAPI operation",
		},
		{
			name: "file scope route name shadow",
			source: `package rest
			import routepkg "github.com/portpowered/go-alexa/pkg/internal/apiroutes"
			type routeValues struct{ MethodListRestEndpoints, PathListRestEndpoints string }
			var routes routeValues
			func send() { c.doJSONRequest(ctx, routes.MethodListRestEndpoints, routes.PathListRestEndpoints, nil, nil) }`,
			want: "method must be a generated OpenAPI operation",
		},
		{
			name: "local HTTP alias shadow",
			source: `package rest
			import (
				web "net/http"
				routes "github.com/portpowered/go-alexa/pkg/internal/apiroutes"
			)
			type fakeHTTP struct{}
			type methodName string
			func (fakeHTTP) NewRequestWithContext(context.Context, methodName, string, io.Reader) (*http.Request, error) { return nil, nil }
			func (c *Client) ExchangeRefreshTokenForCookies(web fakeHTTP) {
				req, _ := web.NewRequestWithContext(ctx, routes.MethodExchangeRefreshTokenForCookies, "/ap/exchangetoken/cookies", nil)
				c.httpClient.Do(req)
			}`,
			want: "request constructor must be http.NewRequestWithContext",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			file := parseTestFile(t, fset, "pkg/dependencies/rest/auth.go", test.source)

			var violations []string

			checkWireCallsites(file, fset, nil, map[string]string{
				"ListRestEndpoints": "/v2/endpoints", "ExchangeRefreshTokenForCookies": "/ap/exchangetoken/cookies",
			}, 1, &violations)
			checkNetworkInventory(file, fset, "pkg/dependencies/rest/auth.go", &violations)

			if got := strings.Join(violations, "\n"); !strings.Contains(got, test.want) {
				t.Fatalf("gate result %q, want %q", got, test.want)
			}
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
		{
			name: "custom header map literal",
			source: `package rest
			func send() {
				headers := map[string]string{"X-Undeclared": "value"}
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "custom request header map key must use a schema-generated Header constant",
		},
		{
			name: "aliased custom header map literal",
			source: `package rest
			func send() {
				headers := map[string]string{"X-Undeclared": "value"}
				alias := headers
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, alias, nil, true)
			}`,
			want: "custom request header map key must use a schema-generated Header constant",
		},
		{
			name: "custom header map write with another name",
			source: `package rest
			func send() {
				headers := make(map[string]string)
				headers["Cookie"] = "value"
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "request header map key must use a schema-generated Header constant",
		},
		{
			name: "aliased custom header map write",
			source: `package rest
			func send() {
				headers := make(map[string]string)
				alias := headers
				alias["Cookie"] = "value"
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
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

	file := parseTestFile(t, fset, test.name+".go", test.source)

	var violations []string

	checkWireCallsites(file, fset, nil, map[string]string{
		"ListRestEndpoints": "/v2/endpoints", "GetRestEndpoint": "/v2/endpoints/%s",
		"OpenDirectiveStream": "/directives", "ExchangeRefreshTokenForCookies": "/ap/exchangetoken/cookies",
	}, 1, &violations)

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
		{"aliased request constructor", `package rest
		func (c *Client) send() { newRequest := http.NewRequestWithContext; _, _ = newRequest(ctx, "GET", c.baseURL+suffix, nil) }`,
			"outbound network method value must be called directly"},
		{"aliased Client.Do", `package rest
		func (c *Client) send() { send := c.httpClient.Do; _, _ = send(req) }`,
			"outbound network method value must be called directly"},
		{"package-level network aliases", `package rest
		var makeRequest = http.NewRequestWithContext
		var sendRequest = http.DefaultClient.Do
		func send() { req, _ := makeRequest(ctx, "GET", base+suffix, nil); _, _ = sendRequest(req) }`,
			"outbound network method value must be called directly"},
		{"new network import", `package rest
		import "net/http"
		func send() {}`, "unregistered network import"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()

			file := parseTestFile(t, fset, "pkg/dependencies/rest/new.go", test.source)

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

	file := parseTestFile(t, fset, "pkg/alexa/http2.go", source)

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

			file := parseTestFile(t, fset, "pkg/dependencies/rest/client.go", source)

			var violations []string

			checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

			if got := strings.Join(violations, "\n"); !strings.Contains(got, test.want) {
				t.Fatalf("gate result %q did not reject route mutation", got)
			}
		})
	}
}

func TestNetworkInventoryRejectsHeaderHelperEscape(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package rest
	func (c *Client) doRequest() {
		req, _ := http.NewRequestWithContext(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints, nil)
		mutateHeader(req.Header)
		c.httpClient.Do(req)
	}`

	file := parseTestFile(t, fset, "pkg/dependencies/rest/client.go", source)

	var violations []string

	checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

	if got := strings.Join(violations, "\n"); !strings.Contains(got, "request or URL escaped") {
		t.Fatalf("gate result %q did not reject header escape", got)
	}
}

func TestNetworkInventoryRejectsShadowedInjectedClient(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package rest
	func (c *Client) doRequest() {
		req, _ := http.NewRequestWithContext(ctx, apiroutes.MethodListRestEndpoints,
			c.baseURL+apiroutes.PathListRestEndpoints, nil)
		{ c := fakeClient{httpClient: http.DefaultClient}; c.httpClient.Do(req) }
	}`

	file := parseTestFile(t, fset, "pkg/dependencies/rest/client.go", source)

	var violations []string

	checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

	if got := strings.Join(violations, "\n"); !strings.Contains(got, "Client.Do must use the inventoried injected client") {
		t.Fatalf("gate result %q did not reject shadowed injected client", got)
	}
}

func TestNetworkInventoryRejectsShadowedHTTPPackage(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package rest
	func (c *Client) doRequest(http fakeHTTP) {
		req, _ := http.NewRequestWithContext(ctx, apiroutes.MethodListRestEndpoints,
			c.baseURL+apiroutes.PathListRestEndpoints, nil)
		c.httpClient.Do(req)
	}`

	file := parseTestFile(t, fset, "pkg/dependencies/rest/client.go", source)

	var violations []string

	checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

	if got := strings.Join(violations, "\n"); !strings.Contains(got, "request constructor must be http.NewRequestWithContext") {
		t.Fatalf("gate result %q did not reject shadowed http package", got)
	}
}

func parseTestFile(t *testing.T, fset *token.FileSet, filename, source string) *ast.File {
	t.Helper()

	file, err := parser.ParseFile(fset, filename, source, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Imports) != 0 {
		return file
	}

	packageEnd := strings.Index(source, "\n")
	if packageEnd < 0 {
		t.Fatal("test source has no package clause")
	}

	imports := `import (
		"fmt"
		http "net/http"
		url "net/url"
		alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
		apiroutes "github.com/portpowered/go-alexa/pkg/internal/apiroutes"
	)`
	source = source[:packageEnd+1] + imports + "\n" + source[packageEnd+1:]

	file, err = parser.ParseFile(fset, filename, source, 0)
	if err != nil {
		t.Fatal(err)
	}

	return file
}

func TestWireCallsiteGateChecksVerifiedHeaderHelper(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "verified header helper raw key",
		source: `package rest
		func (c *Client) setFullURLAuthentication(headers http.Header) {
			headers.Set("X-Undeclared", "raw")
		}`,
		want: "request header name must use a schema-generated Header constant",
	})
}

func TestWireCallsiteGateRejectsOutboundURLQueryProvenance(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "query values copied from outbound URL",
		source: `package rest
		func send(raw string) {
			parsedURL, _ := url.Parse(raw)
			params := parsedURL.Query()
			params.Set(alexamodels.QueryParamExpand, "all")
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "not paired with its generated route",
	})
}

func TestWireCallsiteGateRejectsRequestHeaderMethodValueAliases(t *testing.T) {
	t.Parallel()

	for _, method := range []string{"Set", "Add"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			assertWireCallsiteGateCase(t, wireCallsiteTestCase{
				name: "request Header." + method + " method value",
				source: `package rest
				func send() {
					request := req
					writeHeader := request.Header.` + method + `
					writeHeader("X-Unmodeled", "value")
				}`,
				want: "schema-keyed map method value must not be aliased",
			})
		})
	}
}

func TestWireCallsiteGateRejectsNestedAggregateHelperArguments(t *testing.T) {
	t.Parallel()

	tests := []wireCallsiteTestCase{
		{
			name: "nested header aggregate argument",
			source: `package rest
			func wrap(value any) any { return value }
			func send() {
				headers := map[string]string{apiroutes.HeaderCookie: "safe"}
				wrap(struct {
					rows [1][]map[string]string
				}{rows: [1][]map[string]string{{headers}}})
				c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListRestEndpoints,
					apiroutes.PathListRestEndpoints, nil, headers, nil, true)
			}`,
			want: "custom header map must not escape to an unverified helper",
		},
		{
			name: "nested query aggregate argument",
			source: `package rest
			func wrap(value any) any { return value }
			func send() {
				params := url.Values{}
				params.Set(alexamodels.QueryParamExpand, "all")
				wrap(struct {
					rows [1][]url.Values
				}{rows: [1][]url.Values{{params}}})
				path := apiroutes.PathListRestEndpoints
				path += "?" + params.Encode()
				c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
			}`,
			want: "query parameter map must not escape to an unverified helper",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertWireCallsiteGateCase(t, test)
		})
	}
}

func TestWireCallsiteGateRejectsUnverifiedQueryMapMethodReceiver(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "unverified query method receiver",
		source: `package rest
		func send() {
			params := url.Values{}
			params.mutate()
			path := apiroutes.PathListRestEndpoints
			path += "?" + params.Encode()
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "query parameter map must not escape to an unverified helper",
	})
}

func TestWireCallsiteGateAcceptsSameGeneratedRouteOnEveryBranch(t *testing.T) {
	t.Parallel()

	assertWireCallsiteGateCase(t, wireCallsiteTestCase{
		name: "same generated route assigned in both branches",
		source: `package rest
		func send(useFirst bool) {
			var path string
			if useFirst {
				path = apiroutes.PathListRestEndpoints
			} else {
				path = apiroutes.PathListRestEndpoints
			}
			c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, nil)
		}`,
		want: "",
	})
}

func TestNetworkInventoryRejectsOutboundPrimitives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "http Get", source: `package rest
			import "net/http"
			func send() { _, _ = http.Get("https://example.com") }`, want: "unregistered outbound network primitive Get"},
		{name: "http PostForm", source: `package rest
			import "net/http"
			func send() { _, _ = http.PostForm("https://example.com", nil) }`, want: "unregistered outbound network primitive PostForm"},
		{name: "http Head", source: `package rest
			import "net/http"
			func send() { _, _ = http.Head("https://example.com") }`, want: "unregistered outbound network primitive Head"},
		{name: "RoundTrip", source: `package rest
			func send() { _, _ = transport.RoundTrip(req) }`, want: "unregistered outbound network primitive RoundTrip"},
		{name: "net Dial", source: `package rest
			import "net"
			func send() { _, _ = net.Dial("tcp", "example.com:443") }`, want: "unregistered outbound network primitive Dial"},
		{name: "net DialContext", source: `package rest
			import "net"
			func send() { _, _ = net.DialContext(ctx, "tcp", "example.com:443") }`, want: "unregistered outbound network primitive DialContext"},
		{name: "transport DialTLS", source: `package rest
			func send() { _ = transport.DialTLS(ctx, "tcp", "example.com:443") }`, want: "unregistered outbound network primitive DialTLS"},
		{name: "transport DialTLSContext", source: `package rest
			func send() { _ = transport.DialTLSContext(ctx, "tcp", "example.com:443") }`, want: "unregistered outbound network primitive DialTLSContext"},
		{name: "connection Upgrade", source: `package rest
			func send() { _ = connection.Upgrade(req, response) }`, want: "unregistered outbound network primitive Upgrade"},
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

func TestNetworkInventoryRejectsDotImportedProtectedPackages(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package rest
	import . "net/http"
	func send() { _, _ = Get("https://example.com") }`

	file, err := parser.ParseFile(fset, "pkg/dependencies/rest/client.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

	var violations []string

	checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

	if got := strings.Join(violations, "\n"); !strings.Contains(got, "protected package import must not be dot-imported") {
		t.Fatalf("gate result %q did not reject a dot-imported protected package", got)
	}
}

func TestNetworkInventoryRejectsUnregisteredNetworkImports(t *testing.T) {
	t.Parallel()

	for _, importPath := range []string{"net", "golang.org/x/net/websocket", "github.com/gorilla/websocket"} {
		t.Run(importPath, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			source := "package rest\nimport _ " + strconv.Quote(importPath) + "\nfunc send() {}"

			file, err := parser.ParseFile(fset, "pkg/dependencies/rest/new.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}

			var violations []string

			checkNetworkInventory(file, fset, "pkg/dependencies/rest/new.go", &violations)

			if got := strings.Join(violations, "\n"); !strings.Contains(got, "unregistered network import") {
				t.Fatalf("gate result %q, want an unregistered network import error", got)
			}
		})
	}
}
