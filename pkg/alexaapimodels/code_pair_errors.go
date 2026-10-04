package alexaapimodels

import "fmt"

// CodePairPendingError indicates that the user has not completed Amazon's
// code-based linking step yet. Its message intentionally omits vendor details.
//
//modelinventory:semantic Public code-link failure classification with no provider payload.
type CodePairPendingError struct{}

func (*CodePairPendingError) Error() string {
	return "code-based linking is still pending"
}

// CodePairExpiredError indicates that a code pair is expired or was rejected
// by Amazon. Its message intentionally omits vendor details.
//
//modelinventory:semantic Public code-link failure classification with no provider payload.
type CodePairExpiredError struct{}

func (*CodePairExpiredError) Error() string {
	return "code-based linking code expired or was rejected"
}

// CodePairDeniedError indicates that the user denied code-based linking.
//
//modelinventory:semantic Public code-link failure classification with no provider payload.
type CodePairDeniedError struct{}

func (*CodePairDeniedError) Error() string {
	return "code-based linking was denied"
}

// CodePairRegistrationError represents any other HTTP response from the
// code-pair registration endpoint without retaining its response body.
//
//modelinventory:semantic Public code-link failure classification with no provider payload.
type CodePairRegistrationError struct {
	StatusCode int
}

func (e *CodePairRegistrationError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("code-pair registration failed with status %d", e.StatusCode)
	}

	return "code-pair registration failed"
}

// CodePairGenerationError represents an HTTP rejection while requesting a new
// code pair. The provider response body is intentionally omitted.
//
//modelinventory:semantic Public code-link failure classification with no provider payload.
type CodePairGenerationError struct {
	StatusCode int
}

func (e *CodePairGenerationError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("code-pair generation failed with status %d", e.StatusCode)
	}

	return "code-pair generation failed"
}
