package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func mapCodePairRegistrationError(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}

	var requestErr *UnauthenticatedRequestError
	if !errors.As(err, &requestErr) {
		return &alexaapimodels.NetworkError{Message: "failed to register with code pair"}
	}

	if requestErr.StatusCode < http.StatusBadRequest || requestErr.StatusCode >= http.StatusInternalServerError {
		return &alexaapimodels.CodePairRegistrationError{StatusCode: requestErr.StatusCode}
	}

	switch code := codePairErrorCode(requestErr.Body); code {
	case "authorization_pending":
		return &alexaapimodels.CodePairPendingError{}
	case "expired_token", "invalid_code_pair":
		return &alexaapimodels.CodePairExpiredError{}
	case "access_denied":
		return &alexaapimodels.CodePairDeniedError{}
	case "InvalidDevice":
		return &alexaapimodels.CodePairInvalidDeviceError{StatusCode: requestErr.StatusCode}
	default:
		return &alexaapimodels.CodePairRegistrationError{StatusCode: requestErr.StatusCode}
	}
}

func codePairErrorCode(body string) string {
	var response alexamodels.WireCodePairErrorResponse

	err := json.Unmarshal([]byte(body), &response)
	if err != nil {
		return ""
	}

	if response.Response != nil && response.Response.Error != nil && response.Response.Error.Code != nil {
		return *response.Response.Error.Code
	}

	if response.Error != nil {
		return *response.Error
	}

	return ""
}

func mapCodePairGenerationError(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}

	var requestErr *UnauthenticatedRequestError
	if errors.As(err, &requestErr) {
		if requestErr.StatusCode >= http.StatusBadRequest && requestErr.StatusCode < http.StatusInternalServerError &&
			codePairErrorCode(requestErr.Body) == "InvalidDevice" {
			return &alexaapimodels.CodePairInvalidDeviceError{StatusCode: requestErr.StatusCode}
		}

		return &alexaapimodels.CodePairGenerationError{StatusCode: requestErr.StatusCode}
	}

	return &alexaapimodels.NetworkError{Message: "failed to generate code pair"}
}
