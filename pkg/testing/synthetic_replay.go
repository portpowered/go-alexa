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

// SyntheticExchange is a hand-authored request/response pair, never account traffic.
// Request fields are exact; stable synthetic IDs and credentials need no wildcard.
type SyntheticExchange struct {
	Source    string `json:"source"`
	Operation string `json:"operation"`
	Request   struct {
		Method      string      `json:"method"`
		Origin      string      `json:"origin"`
		EscapedPath string      `json:"escaped_path"`
		Query       url.Values  `json:"query"`
		Headers     http.Header `json:"headers"`
		Body        string      `json:"body"`
	} `json:"request"`
	Response struct {
		Status  int         `json:"status"`
		Headers http.Header `json:"headers"`
		Body    string      `json:"body"`
	} `json:"response"`
}

// SyntheticReplay returns responses only after the next exact request matches.
type SyntheticReplay struct {
	mu    sync.Mutex
	pairs []SyntheticExchange
	next  int
}

var _ http.RoundTripper = (*SyntheticReplay)(nil)

func LoadSyntheticReplay(path string) (*SyntheticReplay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pairs []SyntheticExchange
	if err := json.Unmarshal(data, &pairs); err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, fmt.Errorf("synthetic replay %s has no exchanges", path)
	}
	for i, pair := range pairs {
		if pair.Source != "synthetic" || pair.Operation == "" || pair.Request.Method == "" || pair.Request.Origin == "" || pair.Request.EscapedPath == "" || pair.Response.Status == 0 {
			return nil, fmt.Errorf("synthetic replay %s pair %d is incomplete or lacks provenance", path, i+1)
		}
	}
	return &SyntheticReplay{pairs: pairs}, nil
}

// RecordSyntheticExchange copies a fixed test request and its hand-authored
// response into a pair. It is used only to update checked-in synthetic cases.
func RecordSyntheticExchange(operation string, req *http.Request, status int, headers http.Header, responseBody string) (SyntheticExchange, error) {
	var pair SyntheticExchange
	pair.Source = "synthetic"
	pair.Operation = operation
	pair.Request.Method = req.Method
	pair.Request.Origin = req.URL.Scheme + "://" + req.URL.Host
	pair.Request.EscapedPath = req.URL.EscapedPath()
	pair.Request.Query = req.URL.Query()
	pair.Request.Headers = req.Header.Clone()
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return pair, err
		}
		pair.Request.Body = string(body)
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	pair.Response.Status = status
	pair.Response.Headers = headers.Clone()
	pair.Response.Body = responseBody
	return pair, nil
}

func WriteSyntheticReplay(path string, pairs []SyntheticExchange) error {
	data, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

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
		return nil, fmt.Errorf("unexpected synthetic replay request after %d exchanges", len(r.pairs))
	}
	pair := &r.pairs[r.next]
	origin := req.URL.Scheme + "://" + req.URL.Host
	if req.Method != pair.Request.Method || origin != pair.Request.Origin || req.URL.EscapedPath() != pair.Request.EscapedPath ||
		!reflect.DeepEqual(req.URL.Query(), pair.Request.Query) || !reflect.DeepEqual(req.Header, pair.Request.Headers) ||
		!bytes.Equal(body, []byte(pair.Request.Body)) {
		return nil, fmt.Errorf("synthetic replay request %d (%s) mismatch: %s %s, query=%v, headers=%v, body=%q", r.next+1, pair.Operation, req.Method, req.URL, req.URL.Query(), req.Header, body)
	}
	r.next++
	return &http.Response{StatusCode: pair.Response.Status, Status: fmt.Sprintf("%d %s", pair.Response.Status, http.StatusText(pair.Response.Status)), Header: pair.Response.Headers.Clone(), Body: io.NopCloser(strings.NewReader(pair.Response.Body)), Request: req}, nil
}

func (r *SyntheticReplay) AssertConsumed() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.next != len(r.pairs) {
		return fmt.Errorf("synthetic replay consumed %d of %d exchanges", r.next, len(r.pairs))
	}
	return nil
}
