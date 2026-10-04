package graphql

import (
	"context"
	"errors"
	"net/http"

	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// The GraphQL route is fixed. Never include its host, query or variables in diagnostics.
func graphQLRequestFailure(status int, stage string, cause error) error {
	return &alexaapimodels.RequestError{
		Method: http.MethodPost, Path: apiroutes.PathExecuteNexusGraphQL,
		StatusCode: status, ProviderReason: "", Stage: stage, Cause: cause,
	}
}

func graphQLTransportCause(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}

	return nil
}
