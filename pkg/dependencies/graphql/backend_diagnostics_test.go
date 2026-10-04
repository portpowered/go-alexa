//nolint:testpackage // Verifies complete error chains at the private request boundary.
package graphql

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

type privateFailureEncoder struct{}

func (privateFailureEncoder) MarshalJSON() ([]byte, error) {
	return nil, syntheticFailureError("private-encoder-detail")
}

func TestGraphQLDiagnosticsDiscardPrivateCauses(t *testing.T) {
	t.Parallel()

	for _, encodingFailure := range []bool{true, false} {
		t.Run(map[bool]string{true: "encoding", false: "transport"}[encodingFailure], func(t *testing.T) {
			t.Parallel()

			client := NewClient(WithBearerToken("synthetic-token"), WithHTTPClient(&http.Client{
				Transport: graphqlRoundTripper(func(*http.Request) (*http.Response, error) {
					if encodingFailure {
						t.Fatal("encoding failure reached HTTP")
					}

					return nil, syntheticFailureError("private-transport-detail")
				}),
			}))

			variables := map[string]interface{}{}
			if encodingFailure {
				variables["input"] = privateFailureEncoder{}
			}

			err := client.Execute(context.Background(), GetEndpoint_Operation, variables, nil)

			var diagnostic *alexaapimodels.RequestError

			if !errors.As(err, &diagnostic) {
				t.Fatalf("expected typed request diagnostics, got %T", err)
			}

			for cause := err; cause != nil; cause = errors.Unwrap(cause) {
				if strings.Contains(cause.Error(), "private-") {
					t.Fatalf("private detail survived in %T", cause)
				}
			}
		})
	}
}
