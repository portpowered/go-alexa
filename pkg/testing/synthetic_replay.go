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

// SyntheticRequest is the exact local HTTP request recorded for synthetic replay.
//
//modelinventory:domain SyntheticRequest: hand-authored local request fields used to match method, origin, path, query, headers, and body.
type SyntheticRequest struct {
	Method      string      `json:"method"`
	Origin      string      `json:"origin"`
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

	pair.Source = "synthetic"
	pair.Operation = operation
	pair.Request = SyntheticRequest{
		Method:      req.Method,
		Origin:      req.URL.Scheme + "://" + req.URL.Host,
		EscapedPath: req.URL.EscapedPath(),
		Query:       req.URL.Query(),
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

	if req.Body != nil {
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

	origin := req.URL.Scheme + "://" + req.URL.Host

	requestMatches := req.Method == pair.Request.Method && origin == pair.Request.Origin &&
		req.URL.EscapedPath() == pair.Request.EscapedPath &&
		reflect.DeepEqual(req.URL.Query(), pair.Request.Query) &&
		reflect.DeepEqual(req.Header, pair.Request.Headers) &&
		bytes.Equal(body, []byte(pair.Request.Body))
	if !requestMatches {
		return nil, fmt.Errorf(
			"%w %d (%s) mismatch: %s %s, query=%v, headers=%v, body=%q",
			errSyntheticRequestMismatch,
			r.next+1,
			pair.Operation,
			req.Method,
			req.URL,
			req.URL.Query(),
			req.Header,
			body,
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

// AssertConsumed reports whether every configured exchange was used.
func (r *SyntheticReplay) AssertConsumed() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.next != len(r.pairs) {
		return fmt.Errorf("%w %d of %d exchanges", errSyntheticPairsUnconsumed, r.next, len(r.pairs))
	}

	return nil
}
