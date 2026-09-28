package alexa

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type responseTransport func(*http.Request) (*http.Response, error)

func (transport responseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

type labeledTransport struct{ label string }

func (transport *labeledTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return syntheticResponse(request, http.StatusOK, `{}`, nil), nil
}

func syntheticResponse(request *http.Request, status int, body string, headers http.Header) *http.Response {
	if headers == nil {
		headers = make(http.Header)
	}
	return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

func TestNewClientAppliesRegionTransportAndTimeoutOptions(t *testing.T) {
	commonTransport := &labeledTransport{label: "shared"}
	restTransport := &labeledTransport{label: "REST"}
	graphqlTransport := &labeledTransport{label: "GraphQL"}
	eventTransport := &labeledTransport{label: "event"}
	shared := &http.Client{Transport: commonTransport, Timeout: 9 * time.Second}
	client, err := NewClient(
		WithRegion(alexaapimodels.RegionEU),
		WithAlexaAPIBaseURL("https://rest.example.test"),
		WithAmazonAPIBaseURL("https://auth.example.test"),
		WithAlexaWebBaseURL("https://web.example.test"),
		WithGraphQLBaseURL("https://graphql.example.test"),
		WithEventAuthority("events.example.test:8443"),
		WithHTTPClient(shared),
		WithRESTHTTPClient(&http.Client{Transport: restTransport, Timeout: 8 * time.Second}),
		WithGraphQLHTTPClient(&http.Client{Transport: graphqlTransport, Timeout: 7 * time.Second}),
		WithEventTransport(eventTransport),
		WithTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.region != alexaapimodels.RegionEU || client.apiBaseURL != "https://rest.example.test" || client.amazonBaseURL != "https://auth.example.test" || client.webBaseURL != "https://web.example.test" || client.graphqlBaseURL != "https://graphql.example.test" || client.eventAuthority != "events.example.test:8443" {
		t.Fatalf("client endpoints not configured: %#v", client)
	}
	if client.restHTTPClient.Transport != restTransport || client.restHTTPClient.Timeout != 8*time.Second {
		t.Fatalf("REST client = %#v", client.restHTTPClient)
	}
	if client.graphqlHTTPClient.Transport != graphqlTransport || client.graphqlHTTPClient.Timeout != 7*time.Second {
		t.Fatalf("GraphQL client = %#v", client.graphqlHTTPClient)
	}
	if client.eventHTTPClient == nil || client.eventHTTPClient.Transport != eventTransport || client.eventHTTPClient.Timeout != 0 {
		t.Fatalf("event client = %#v", client.eventHTTPClient)
	}
	if client.restHTTPClient == shared || client.graphqlHTTPClient == shared {
		t.Fatal("NewClient retained the caller's mutable HTTP client pointer")
	}

	defaultClient, err := NewClient(WithRegion(alexaapimodels.RegionJP), WithTimeout(2*time.Second))
	if err != nil {
		t.Fatalf("create region client: %v", err)
	}
	if defaultClient.region != alexaapimodels.RegionJP || defaultClient.restHTTPClient.Timeout != 2*time.Second || defaultClient.graphqlHTTPClient.Timeout != 2*time.Second {
		t.Fatalf("region defaults or timeout not applied: %#v", defaultClient)
	}
}

func TestNewClientRejectsInvalidAndConflictingOptions(t *testing.T) {
	transport := responseTransport(func(request *http.Request) (*http.Response, error) {
		return syntheticResponse(request, http.StatusOK, `{}`, nil), nil
	})
	cases := []struct {
		name    string
		options []Option
	}{
		{name: "nil option", options: []Option{nil}},
		{name: "unsupported region", options: []Option{WithRegion(alexaapimodels.Region("AU"))}},
		{name: "conflicting regions", options: []Option{WithRegion(alexaapimodels.RegionEU), WithRegion(alexaapimodels.RegionJP)}},
		{name: "invalid URL", options: []Option{WithGraphQLBaseURL("ftp://example.test")}},
		{name: "URL credentials", options: []Option{WithAlexaAPIBaseURL("https://user:pass@example.test")}},
		{name: "URL query", options: []Option{WithAmazonAPIBaseURL("https://example.test?x=y")}},
		{name: "conflicting URLs", options: []Option{WithAlexaWebBaseURL("https://one.example"), WithAlexaWebBaseURL("https://two.example")}},
		{name: "invalid event authority", options: []Option{WithEventAuthority("https://events.example.test/path")}},
		{name: "conflicting event authorities", options: []Option{WithEventAuthority("one.example"), WithEventAuthority("two.example")}},
		{name: "nil shared client", options: []Option{WithHTTPClient(nil)}},
		{name: "nil REST client", options: []Option{WithRESTHTTPClient(nil)}},
		{name: "nil GraphQL client", options: []Option{WithGraphQLHTTPClient(nil)}},
		{name: "nil event client", options: []Option{WithEventHTTPClient(nil)}},
		{name: "nil event transport", options: []Option{WithEventTransport(nil)}},
		{name: "conflicting event injection", options: []Option{WithEventHTTPClient(&http.Client{Transport: transport}), WithEventTransport(transport)}},
		{name: "nonpositive timeout", options: []Option{WithTimeout(0)}},
		{name: "conflicting timeouts", options: []Option{WithTimeout(time.Second), WithTimeout(2 * time.Second)}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if client, err := NewClient(test.options...); err == nil || client != nil {
				t.Fatalf("NewClient = (%v, %v), want an error and nil client", client, err)
			}
		})
	}
}

func TestSessionOptionsKeepCredentialsIsolatedAndCallerVisible(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	inputCookie := &http.Cookie{Name: "session-id", Value: "synthetic-cookie"}
	first, err := client.NewSession(
		WithBearerToken("synthetic-access-one"),
		WithRefreshToken("synthetic-refresh-one"),
		WithCSRFToken("synthetic-csrf-one"),
		WithCustomerID("synthetic-customer-one"),
		WithCookies(map[string]*http.Cookie{"session-id": inputCookie}),
	)
	if err != nil {
		t.Fatalf("create first session: %v", err)
	}
	defer first.Close()
	second, err := client.NewSession(WithBearerToken("synthetic-access-two"))
	if err != nil {
		t.Fatalf("create second session: %v", err)
	}
	defer second.Close()

	inputCookie.Value = "mutated-after-construction"
	credentials := first.Credentials()
	if credentials.AccessToken != "synthetic-access-one" || credentials.RefreshToken != "synthetic-refresh-one" || credentials.CSRFToken != "synthetic-csrf-one" || credentials.Cookies["session-id"].Value != "synthetic-cookie" {
		t.Fatalf("first session credentials = %#v", credentials)
	}
	credentials.Cookies["session-id"].Value = "mutated-copy"
	if first.Credentials().Cookies["session-id"].Value != "synthetic-cookie" {
		t.Fatal("Credentials exposed the session's mutable cookie pointer")
	}
	if token, err := first.accessTokenForRequest(context.Background()); err != nil || token != "synthetic-access-one" {
		t.Fatalf("first session token = %q, %v", token, err)
	}
	if err := first.SetAccessToken("synthetic-access-rotated"); err != nil {
		t.Fatalf("rotate first session token: %v", err)
	}
	if first.Credentials().AccessToken != "synthetic-access-rotated" || second.Credentials().AccessToken != "synthetic-access-two" {
		t.Fatalf("token rotation leaked across sessions: first=%#v second=%#v", first.Credentials(), second.Credentials())
	}
	if err := first.SetAccessToken("  "); err == nil {
		t.Fatal("SetAccessToken accepted an empty token")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close session: %v", err)
	}
	if _, err := first.accessTokenForRequest(context.Background()); err == nil {
		t.Fatal("closed session returned an access token")
	}
}

func TestSessionOptionsRejectMissingAndConflictingCredentials(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		options []SessionOption
	}{
		{name: "nil option", options: []SessionOption{nil}},
		{name: "empty access token", options: []SessionOption{WithBearerToken(" ")}},
		{name: "conflicting access tokens", options: []SessionOption{WithBearerToken("one"), WithBearerToken("two")}},
		{name: "empty refresh token", options: []SessionOption{WithRefreshToken("")}},
		{name: "conflicting refresh tokens", options: []SessionOption{WithRefreshToken("one"), WithRefreshToken("two")}},
		{name: "empty customer ID", options: []SessionOption{WithCustomerID(" ")}},
		{name: "conflicting customer IDs", options: []SessionOption{WithCustomerID("one"), WithCustomerID("two")}},
		{name: "nil cookie map", options: []SessionOption{WithCookies(nil)}},
		{name: "nil cookie value", options: []SessionOption{WithCookies(map[string]*http.Cookie{"broken": nil})}},
		{name: "empty CSRF token", options: []SessionOption{WithCSRFToken(" ")}},
		{name: "conflicting CSRF tokens", options: []SessionOption{WithCSRFToken("one"), WithCSRFToken("two")}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if session, err := client.NewSession(test.options...); err == nil || session != nil {
				t.Fatalf("NewSession = (%v, %v), want an error and nil session", session, err)
			}
		})
	}
}

func TestSessionCSRFRetrievalIsExplicitAndReturned(t *testing.T) {
	requests := 0
	client, err := NewClient(WithHTTPClient(&http.Client{Transport: responseTransport(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Method != http.MethodGet || request.URL.Path != "/api/language" {
			t.Fatalf("unexpected CSRF request: %s %s", request.Method, request.URL)
		}
		if got := request.Header.Get("Cookie"); !strings.Contains(got, "session-id=synthetic-cookie") {
			t.Fatalf("cookie header = %q", got)
		}
		headers := make(http.Header)
		headers.Add("Set-Cookie", "csrf=synthetic-csrf; Path=/")
		return syntheticResponse(request, http.StatusOK, `{}`, headers), nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.NewSession(WithCookies(map[string]*http.Cookie{"session-id": {Name: "session-id", Value: "synthetic-cookie"}}))
	if err != nil {
		t.Fatalf("create cookie session: %v", err)
	}
	defer session.Close()
	if token, err := session.GetCSRFToken(context.Background()); err != nil || token != "synthetic-csrf" {
		t.Fatalf("GetCSRFToken = %q, %v", token, err)
	}
	if requests != 1 || session.Credentials().CSRFToken != "synthetic-csrf" {
		t.Fatalf("request count = %d, credentials = %#v", requests, session.Credentials())
	}
	if _, err := session.GetCSRFToken(context.Background()); err != nil || requests != 1 {
		t.Fatalf("cached GetCSRFToken made another request: requests=%d err=%v", requests, err)
	}
}

func TestSessionRefreshReturnsCredentialsWithoutImplicitlyReplacingThem(t *testing.T) {
	var refreshRequest struct {
		SourceToken string `json:"source_token"`
	}
	client, err := NewClient(WithHTTPClient(&http.Client{Transport: responseTransport(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read refresh request: %v", err)
		}
		if err := json.Unmarshal(body, &refreshRequest); err != nil {
			t.Fatalf("decode refresh request: %v", err)
		}
		return syntheticResponse(request, http.StatusOK, `{"access_token":"synthetic-access-refreshed","expires_in":3600}`, nil), nil
	})}))
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.NewSession(WithBearerToken("synthetic-access-before"), WithRefreshToken("synthetic-refresh"))
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	defer session.Close()
	result, err := session.RefreshAccessToken(context.Background(), alexaapimodels.DefaultDeviceRegistrationConfig("synthetic-serial", "synthetic-device"))
	if err != nil {
		t.Fatalf("refresh access token: %v", err)
	}
	if refreshRequest.SourceToken != "synthetic-refresh" || result.AccessToken != "synthetic-access-refreshed" || result.ExpiresInSeconds != 3600 {
		t.Fatalf("refresh request=%#v result=%#v", refreshRequest, result)
	}
	if got := session.Credentials().AccessToken; got != "synthetic-access-before" {
		t.Fatalf("refresh operation silently replaced the session token with %q", got)
	}
	if err := session.SetAccessToken(result.AccessToken); err != nil {
		t.Fatalf("install caller-managed refreshed token: %v", err)
	}
	if got := session.Credentials().AccessToken; got != "synthetic-access-refreshed" {
		t.Fatalf("session token after explicit SetAccessToken = %q", got)
	}
}
