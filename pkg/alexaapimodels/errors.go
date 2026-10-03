package alexaapimodels

import (
	"errors"
	"fmt"
	"net/http"
)

// AuthenticationError represents an authentication failure.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type AuthenticationError struct {
	Message string
	Status  int
}

// NewAuthenticationError creates a new AuthenticationError.
func NewAuthenticationError(message string, status int) *AuthenticationError {
	return &AuthenticationError{
		Message: message,
		Status:  status,
	}
}

func (e *AuthenticationError) Error() string {
	if e.Message != "" {
		return "authentication error: " + e.Message
	}

	return fmt.Sprintf("authentication error (status: %d)", e.Status)
}

// IsAuthenticationError checks if an error is an AuthenticationError.
func IsAuthenticationError(err error) bool {
	authenticationError := &AuthenticationError{}
	ok := errors.As(err, &authenticationError)

	return ok
}

// ConnectionError represents a connection failure.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type ConnectionError struct {
	Message string
	Err     error
}

// NewConnectionError creates a new ConnectionError.
func NewConnectionError(message string, err error) *ConnectionError {
	return &ConnectionError{
		Message: message,
		Err:     err,
	}
}

func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("connection error: %s: %v", e.Message, e.Err)
	}

	return "connection error: " + e.Message
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// IsConnectionError checks if an error is a ConnectionError.
func IsConnectionError(err error) bool {
	connectionError := &ConnectionError{}
	ok := errors.As(err, &connectionError)

	return ok
}

// NetworkError represents a network-level error.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type NetworkError struct {
	Message string
	Err     error
}

// NewNetworkError creates a new NetworkError.
func NewNetworkError(message string, err error) *NetworkError {
	return &NetworkError{
		Message: message,
		Err:     err,
	}
}

func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Err)
	}

	return "network error: " + e.Message
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// IsNetworkError checks if an error is a NetworkError.
func IsNetworkError(err error) bool {
	networkError := &NetworkError{}
	ok := errors.As(err, &networkError)

	return ok
}

// TokenError represents a token retrieval or validation error.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type TokenError struct {
	Message string
	Err     error
}

// NewTokenError creates a new TokenError.
func NewTokenError(message string, err error) *TokenError {
	return &TokenError{
		Message: message,
		Err:     err,
	}
}

func (e *TokenError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("token error: %s: %v", e.Message, e.Err)
	}

	return "token error: " + e.Message
}

func (e *TokenError) Unwrap() error {
	return e.Err
}

// IsTokenError checks if an error is a TokenError.
func IsTokenError(err error) bool {
	tokenError := &TokenError{}
	ok := errors.As(err, &tokenError)

	return ok
}

// SdkError represents an error reported by the Alexa SDK layer.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type SdkError struct {
	Message string
	Err     error
}

func (e *SdkError) Error() string {
	return fmt.Sprintf("sdk error: %s: %v", e.Message, e.Err)
}

func (e *SdkError) Unwrap() error {
	return e.Err
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error: %s: %v", e.Message, e.Err)
}

func (e *ServerError) Unwrap() error {
	return e.Err
}

// ServerError represents an error returned by an Alexa service.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type ServerError struct {
	Message string
	Err     error
}

// BadRequestError represents a request rejected as invalid by the service.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type BadRequestError struct {
	Message string
	Err     error
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("bad request error: %s: %v", e.Message, e.Err)
}

func (e *BadRequestError) Unwrap() error {
	return e.Err
}

// IsBadRequestError checks if an error is a BadRequestError.
func IsBadRequestError(err error) bool {
	badRequestError := &BadRequestError{}
	ok := errors.As(err, &badRequestError)

	return ok
}

// UnauthorizedError represents a request rejected for missing or invalid authorization.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type UnauthorizedError struct {
	Message string
	Err     error
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized error: %s: %v", e.Message, e.Err)
}

func (e *UnauthorizedError) Unwrap() error {
	return e.Err
}

// IsUnauthorizedError checks if an error is a UnauthorizedError.
func IsUnauthorizedError(err error) bool {
	unauthorizedError := &UnauthorizedError{}
	ok := errors.As(err, &unauthorizedError)

	return ok
}

// NotFoundError represents a resource that the service could not find.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type NotFoundError struct {
	Message string
	Err     error
}

func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return "not found error: " + e.Message
	}

	return "not found"
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

// InternalServerError represents an unexpected failure in the service.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type InternalServerError struct {
	Message string
	Err     error
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("internal server error: %s: %v", e.Message, e.Err)
}

func (e *InternalServerError) Unwrap() error {
	return e.Err
}

// IsInternalServerError checks if an error is a InternalServerError.
func IsInternalServerError(err error) bool {
	internalServerError := &InternalServerError{}
	ok := errors.As(err, &internalServerError)

	return ok
}

// HTTPError represents an HTTP-level error with status code.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
	Message    string
}

// NewHTTPError creates a new HTTPError from an HTTP response.
func NewHTTPError(resp *http.Response, body string) *HTTPError {
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       body,
	}
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP error %d (%s): %s", e.StatusCode, e.Status, e.Message)
	}

	if e.Body != "" {
		return fmt.Sprintf("HTTP error %d (%s): %s", e.StatusCode, e.Status, e.Body)
	}

	return fmt.Sprintf("HTTP error %d (%s)", e.StatusCode, e.Status)
}

// IsHTTPError checks if an error is an HTTPError.
func IsHTTPError(err error) bool {
	hTTPError := &HTTPError{}
	ok := errors.As(err, &hTTPError)

	return ok
}

// IsHTTPStatusCode checks if an error is an HTTPError with a specific status code.
func IsHTTPStatusCode(err error, code int) bool {
	httpErr := &HTTPError{}

	ok := errors.As(err, &httpErr)
	if !ok {
		return false
	}

	return httpErr.StatusCode == code
}

// ClosedError represents an error when trying to use a closed connection.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type ClosedError struct {
	Message string
}

// NewClosedError creates a new ClosedError.
func NewClosedError(message string) *ClosedError {
	return &ClosedError{
		Message: message,
	}
}

func (e *ClosedError) Error() string {
	if e.Message != "" {
		return "connection closed: " + e.Message
	}

	return "connection closed"
}

// IsClosedError checks if an error is a ClosedError.
func IsClosedError(err error) bool {
	closedError := &ClosedError{}
	ok := errors.As(err, &closedError)

	return ok
}

// PingError represents an error during ping operations.
//modelinventory:semantic Public SDK input or output projection; provider transport types are generated under pkg/dependencymodels.
type PingError struct {
	StatusCode int
	Message    string
	Err        error
}

// NewPingError creates a new PingError.
func NewPingError(statusCode int, message string, err error) *PingError {
	return &PingError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

func (e *PingError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("ping error (status %d): %s", e.StatusCode, e.Message)
	}

	if e.Err != nil {
		return fmt.Sprintf("ping error: %s: %v", e.Message, e.Err)
	}

	return "ping error: " + e.Message
}

func (e *PingError) Unwrap() error {
	return e.Err
}

// IsPingError checks if an error is a PingError.
func IsPingError(err error) bool {
	pingError := &PingError{}
	ok := errors.As(err, &pingError)

	return ok
}
