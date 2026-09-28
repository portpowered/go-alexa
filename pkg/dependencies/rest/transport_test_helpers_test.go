package rest

import (
	"io"
	"net/http"
	"strings"
)

type syntheticRoundTripper func(*http.Request) (*http.Response, error)

func (roundTrip syntheticRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func syntheticResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}
