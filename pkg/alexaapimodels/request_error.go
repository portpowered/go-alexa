package alexaapimodels

import "fmt"

// RequestError gives safe operational context for an Alexa API failure.
// Path contains route literals and {id} placeholders, never customer identifiers.
//
//modelinventory:semantic Safe public operational diagnostics; no provider payload or account identity.
type RequestError struct {
	Method         string
	Path           string
	StatusCode     int
	ProviderReason string
	Stage          string
	Cause          error
}

func (e *RequestError) Error() string {
	message := fmt.Sprintf("Alexa %s %s", e.Method, e.Path)
	if e.StatusCode > 0 {
		message += fmt.Sprintf(" returned HTTP %d", e.StatusCode)
	}

	if e.ProviderReason != "" {
		message += ": " + e.ProviderReason
	}

	if e.Stage != "" {
		message += " (" + e.Stage + ")"
	}

	return message
}

func (e *RequestError) Unwrap() error { return e.Cause }
