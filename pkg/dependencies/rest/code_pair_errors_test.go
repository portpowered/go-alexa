//nolint:testpackage // Verifies private error mapping without retry delays or provider traffic.
package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func TestCodePairErrorMappingPreservesCancellationAndClassifiesTransport(t *testing.T) {
	t.Parallel()

	for _, mapError := range []func(error) error{mapCodePairRegistrationError, mapCodePairGenerationError} {
		for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
			err := mapError(fmt.Errorf("wrapped context: %w", cause))
			if !errors.Is(err, cause) {
				t.Fatalf("lost cancellation: %v", err)
			}
		}

		err := mapError(&alexaapimodels.NetworkError{Message: "synthetic-private-network-details"})
		if !alexaapimodels.IsNetworkError(err) || errors.Unwrap(err) != nil {
			t.Fatal("transport error lost safe network classification")
		}
	}
}

func TestCodePairServerErrorsCannotMasqueradeAsDeviceFailures(t *testing.T) {
	t.Parallel()

	rejection := &UnauthenticatedRequestError{StatusCode: http.StatusServiceUnavailable, Body: `{"response":{"error":{"code":"InvalidDevice"}}}`}
	registration := mapCodePairRegistrationError(rejection)

	var registerError *alexaapimodels.CodePairRegistrationError

	if !errors.As(registration, &registerError) || registerError.StatusCode != http.StatusServiceUnavailable {
		t.Fatal("server failure misclassified as invalid metadata")
	}

	generation := mapCodePairGenerationError(rejection)

	var generateError *alexaapimodels.CodePairGenerationError

	if !errors.As(generation, &generateError) || generateError.StatusCode != http.StatusServiceUnavailable {
		t.Fatal("server failure lost generation status")
	}
}

func TestCodePairDuplicateNameServerResponseRetainsServerClassification(t *testing.T) {
	t.Parallel()

	err := mapCodePairRegistrationError(&UnauthenticatedRequestError{
		StatusCode: http.StatusServiceUnavailable,
		Body:       `{"response":{"error":{"code":"DuplicateDeviceName"}}}`,
	})

	var failure *alexaapimodels.CodePairRegistrationError

	if !errors.As(err, &failure) || failure.StatusCode != http.StatusServiceUnavailable {
		t.Fatal("server failure misclassified as duplicate device name")
	}
}
