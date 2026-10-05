package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
)

const syntheticReplayFilePermissions = 0o600

// SyntheticExchange is a hand-authored request/response pair, never account traffic.
// Request fields are exact; stable synthetic IDs and credentials need no wildcard.
//
//modelinventory:domain SyntheticExchange: ordered local synthetic request and response pair used by exact-match replay.
type SyntheticExchange struct {
	Source    string            `json:"source"`
	Operation string            `json:"operation"`
	Request   SyntheticRequest  `json:"request"`
	Response  SyntheticResponse `json:"response"`
}

// SyntheticRequest is the exact local HTTP request recorded for synthetic replay,
// including a distinct HTTP authority when Request.Host overrides URL.Host.
//
//modelinventory:domain SyntheticRequest: hand-authored local request fields used to match method, origin, path, query, headers, and body.
type SyntheticRequest struct {
	Method string `json:"method"`
	Origin string `json:"origin"`
	// Host is the optional Request.Host authority override.
	Host        string      `json:"host,omitempty"`
	EscapedPath string      `json:"escaped_path"`
	Query       url.Values  `json:"query"`
	Headers     http.Header `json:"headers"`
	Body        string      `json:"body"`
}

// SyntheticResponse is the hand-authored local response returned after a replay match.
//
//modelinventory:domain SyntheticResponse: hand-authored local status, headers, and body returned after a synthetic request matches.
type SyntheticResponse struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"`
}

// SyntheticReplay returns responses only after the next exact request matches.
type SyntheticReplay struct {
	mu    sync.Mutex
	pairs []SyntheticExchange
	next  int
}

var _ http.RoundTripper = (*SyntheticReplay)(nil)

// LoadSyntheticReplay reads an ordered set of exchanges from path.
func LoadSyntheticReplay(path string) (*SyntheticReplay, error) {
	//nolint:gosec // Callers provide an explicit local replay fixture path.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read synthetic replay %s: %w", path, err)
	}

	var pairs []SyntheticExchange
	{
		err := json.Unmarshal(data, &pairs)
		if err != nil {
			return nil, fmt.Errorf("decode synthetic replay %s: %w", path, err)
		}
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("%w %s has no exchanges", errSyntheticReplayInvalid, path)
	}

	for i, pair := range pairs {
		incomplete := pair.Source != "synthetic" || pair.Operation == "" ||
			pair.Request.Method == "" || pair.Request.Origin == "" ||
			pair.Request.EscapedPath == "" || pair.Response.Status == 0
		if incomplete {
			return nil, fmt.Errorf("%w %s pair %d is incomplete or lacks provenance", errSyntheticReplayInvalid, path, i+1)
		}
	}

	return &SyntheticReplay{
		mu:    sync.Mutex{},
		pairs: pairs,
		next:  0,
	}, nil
}

// RecordSyntheticExchange copies a fixed test request and its hand-authored
// response into a pair. It is used only to update checked-in synthetic cases.
func RecordSyntheticExchange(operation string, req *http.Request, status int, headers http.Header, responseBody string) (SyntheticExchange, error) {
	var pair SyntheticExchange

	identity, identityIssues := syntheticRequestIdentityFor(req)

	if len(identityIssues) > 0 {
		return pair, fmt.Errorf("%w request identity mismatch: %s", errSyntheticReplayInvalid, strings.Join(identityIssues, ", "))
	}

	hostOverride := ""
	if req.Host != "" && req.Host != req.URL.Host {
		hostOverride = req.Host
	}

	pair.Source = "synthetic"
	pair.Operation = operation
	pair.Request = SyntheticRequest{
		Method:      req.Method,
		Origin:      identity.origin,
		Host:        hostOverride,
		EscapedPath: identity.escapedPath,
		Query:       identity.query,
		Headers:     req.Header.Clone(),
		Body:        "",
	}

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return pair, fmt.Errorf("read synthetic request body: %w", err)
		}

		pair.Request.Body = string(body)
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	pair.Response = SyntheticResponse{
		Status:  status,
		Headers: headers.Clone(),
		Body:    responseBody,
	}

	return pair, nil
}

// WriteSyntheticReplay writes exchanges to path as indented JSON.
func WriteSyntheticReplay(path string, pairs []SyntheticExchange) error {
	data, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode synthetic replay: %w", err)
	}

	err = os.WriteFile(path, append(data, '\n'), syntheticReplayFilePermissions)
	if err != nil {
		return fmt.Errorf("write synthetic replay %s: %w", path, err)
	}

	return nil
}

// RoundTrip matches req against the next exchange and returns its captured response.
func (r *SyntheticReplay) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte

	identity, identityIssues := syntheticRequestIdentityFor(req)

	if req != nil && req.Body != nil {
		var err error

		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read synthetic replay request: %w", err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.next >= len(r.pairs) {
		return nil, fmt.Errorf("%w %d exchanges", errSyntheticRequestsExhausted, len(r.pairs))
	}

	pair := &r.pairs[r.next]

	mismatchFields := syntheticRequestMismatchFields(req, pair.Request, identity, identityIssues, body)

	if len(mismatchFields) > 0 {
		return nil, fmt.Errorf(
			"%w %d (%s) mismatch: %s",
			errSyntheticRequestMismatch,
			r.next+1,
			pair.Operation,
			strings.Join(mismatchFields, ", "),
		)
	}

	r.next++

	return &http.Response{
		StatusCode: pair.Response.Status,
		Status:     fmt.Sprintf("%d %s", pair.Response.Status, http.StatusText(pair.Response.Status)),
		Header:     pair.Response.Headers.Clone(),
		Body:       io.NopCloser(strings.NewReader(pair.Response.Body)),
		Request:    req,
	}, nil
}

func syntheticRequestMismatchFields(
	req *http.Request,
	expected SyntheticRequest,
	identity syntheticRequestIdentity,
	identityIssues []string,
	body []byte,
) []string {
	fields := append([]string{}, identityIssues...)
	if req == nil {
		fields = append(fields, "request")
	} else {
		if req.Method != expected.Method {
			fields = append(fields, "method")
		}

		fields = append(fields, syntheticURLMismatchFields(req, expected, identity, identityIssues)...)
		if !reflect.DeepEqual(req.Header, expected.Headers) {
			fields = append(fields, "headers")
		}
	}

	if !bytes.Equal(body, []byte(expected.Body)) {
		fields = append(fields, "body")
	}

	return fields
}

func syntheticURLMismatchFields(
	req *http.Request,
	expected SyntheticRequest,
	identity syntheticRequestIdentity,
	identityIssues []string,
) []string {
	if req.URL == nil {
		return nil
	}

	var fields []string
	if identity.origin != expected.Origin {
		fields = append(fields, "origin")
	}

	expectedAuthority, validOrigin := syntheticOriginAuthority(expected.Origin)
	if !validOrigin {
		fields = append(fields, "paired origin")
	} else {
		if expected.Host != "" {
			expectedAuthority = expected.Host
		}

		if identity.authority != expectedAuthority {
			fields = append(fields, "effective authority")
		}
	}

	if identity.escapedPath != expected.EscapedPath {
		fields = append(fields, "escaped path")
	}

	if !containsSyntheticIdentityIssue(identityIssues, "malformed query") &&
		!reflect.DeepEqual(identity.query, expected.Query) {
		fields = append(fields, "query")
	}

	return fields
}

type syntheticRequestIdentity struct {
	origin      string
	authority   string
	escapedPath string
	query       url.Values
}

func syntheticRequestIdentityFor(req *http.Request) (syntheticRequestIdentity, []string) {
	if req == nil {
		return syntheticRequestIdentity{
			origin:      "",
			authority:   "",
			escapedPath: "",
			query:       nil,
		}, []string{"request"}
	}

	if req.URL == nil {
		return syntheticRequestIdentity{
			origin:      "",
			authority:   "",
			escapedPath: "",
			query:       nil,
		}, []string{"URL"}
	}

	var issues []string
	if req.RequestURI != "" {
		issues = append(issues, "RequestURI")
	}

	if req.URL.User != nil {
		issues = append(issues, "URL user information")
	}

	if req.URL.Opaque != "" {
		issues = append(issues, "opaque URL")
	}

	if req.URL.Fragment != "" {
		issues = append(issues, "URL fragment")
	}

	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		issues = append(issues, "malformed query")
	}

	authority := req.Host
	if authority == "" {
		authority = req.URL.Host
	}

	return syntheticRequestIdentity{
		origin:      req.URL.Scheme + "://" + req.URL.Host,
		authority:   authority,
		escapedPath: req.URL.EscapedPath(),
		query:       query,
	}, issues
}

func syntheticOriginAuthority(origin string) (string, bool) {
	parsedOrigin, err := url.Parse(origin)
	if err != nil || parsedOrigin.Scheme == "" || parsedOrigin.Host == "" ||
		parsedOrigin.User != nil || parsedOrigin.Opaque != "" || parsedOrigin.Path != "" ||
		parsedOrigin.RawQuery != "" || parsedOrigin.Fragment != "" {
		return "", false
	}

	return parsedOrigin.Host, true
}

func containsSyntheticIdentityIssue(issues []string, issue string) bool {
	for _, current := range issues {
		if current == issue {
			return true
		}
	}

	return false
}

// AssertConsumed reports whether every configured exchange was used.
func (r *SyntheticReplay) AssertConsumed() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.next != len(r.pairs) {
		return fmt.Errorf("%w %d of %d exchanges", errSyntheticPairsUnconsumed, r.next, len(r.pairs))
	}

	return nil
}
