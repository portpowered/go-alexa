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

	file, err := parser.ParseFile(fset, test.name+".go", test.source, 0)
	if err != nil {
		t.Fatal(err)
	}

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

func TestNetworkInventoryRejectsHeaderHelperEscape(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	source := `package rest
	func (c *Client) doRequest() {
		req, _ := http.NewRequestWithContext(ctx, apiroutes.MethodListRestEndpoints, apiroutes.PathListRestEndpoints, nil)
		mutateHeader(req.Header)
		c.httpClient.Do(req)
	}`

	file, err := parser.ParseFile(fset, "pkg/dependencies/rest/client.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

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

	file, err := parser.ParseFile(fset, "pkg/dependencies/rest/client.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

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

	file, err := parser.ParseFile(fset, "pkg/dependencies/rest/client.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}

	var violations []string

	checkNetworkInventory(file, fset, "pkg/dependencies/rest/client.go", &violations)

	if got := strings.Join(violations, "\n"); !strings.Contains(got, "request constructor must be http.NewRequestWithContext") {
		t.Fatalf("gate result %q did not reject shadowed http package", got)
	}
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
