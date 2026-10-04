package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

func safePathSegment(segment string) bool {
	switch segment {
	case "api", "v2", "endpoints", "endpoint-query",
		"features", "interfaces", alexamodels.SDKFeatureNamePower, alexamodels.SDKFeatureNameSpeaker,
		alexamodels.FeatureDefaultPropertyNameVolume, alexamodels.SDKFeatureNamePlayback,
		alexamodels.SDKFeatureOperationNamePause, alexamodels.SDKFeatureOperationNamePlay,
		"behaviors", "preview", "np", "command",
		"player", "devices-v2", "device", "actions",
		"routines", "state", "control", "forget",
		"deregister", "friendlyName":
		return true
	default:
		return false
	}
}

func safeRequestPath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Path == "" {
		return "/{path}"
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i, segment := range segments {
		if !safePathSegment(segment) {
			segments[i] = "{id}"
		}
	}

	return "/" + strings.Join(segments, "/")
}

func providerReason(body []byte) string {
	var response map[string]json.RawMessage
	if json.Unmarshal(body, &response) != nil {
		return ""
	}

	for _, key := range []string{"code", "errorCode", "reason", alexamodels.PayloadKeyType, "error"} {
		value, ok := response[key]
		if !ok {
			continue
		}

		var code string
		if json.Unmarshal(value, &code) == nil && safeProviderCode(code) {
			return code
		}

		var nested map[string]json.RawMessage
		if json.Unmarshal(value, &nested) == nil {
			for _, nestedKey := range []string{"code", "errorCode", "reason", alexamodels.PayloadKeyType} {
				if json.Unmarshal(nested[nestedKey], &code) == nil && safeProviderCode(code) {
					return code
				}
			}
		}
	}

	return ""
}

// Only fixed diagnostic codes can cross into logs. An arbitrary provider code
// could itself be a device ID or contain an account-specific value.
func safeProviderCode(code string) bool {
	switch code {
	case "ACCESS_DENIED", "BAD_REQUEST", "DEVICE_NOT_FOUND", "DEVICE_OFFLINE",
		"ENDPOINT_NOT_FOUND", "ENDPOINT_UNREACHABLE", "FORBIDDEN", "INTERNAL_ERROR",
		"INVALID_COMMAND", "INVALID_DEVICE_TYPE", "INVALID_PARAMETER", "INVALID_REQUEST",
		"NOT_FOUND", "RATE_LIMITED", "SERVICE_UNAVAILABLE", "THROTTLED",
		"TIMEOUT", "UNAUTHORIZED", "UNSUPPORTED_OPERATION":
		return true
	default:
		return false
	}
}

func requestFailure(method, rawURL string, status int, reason, stage string, cause error) error {
	return &alexaapimodels.RequestError{
		Method: method, Path: safeRequestPath(rawURL), StatusCode: status,
		ProviderReason: reason, Stage: stage, Cause: cause,
	}
}

func statusCause(status int) error {
	switch status {
	case http.StatusUnauthorized:
		return &alexaapimodels.UnauthorizedError{Message: "provider rejected authorization"}
	case http.StatusNotFound:
		return &alexaapimodels.NotFoundError{Message: "provider resource not found", Err: nil}
	}

	if status >= 400 && status < 500 {
		return &alexaapimodels.BadRequestError{Message: "provider rejected request"}
	}

	return &alexaapimodels.InternalServerError{Message: "provider unavailable"}
}

func networkFailureCause(ctx context.Context, cause error, message string) error {
	if ctx.Err() != nil {
		cause = ctx.Err()
	}

	return &alexaapimodels.NetworkError{Message: message, Err: safeTransportCause(cause)}
}

func safeTransportCause(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}

	return nil
}

func tokenFailureCause(ctx context.Context, err error) error {
	var authErr *alexaapimodels.AuthenticationError
	if errors.As(err, &authErr) {
		return &alexaapimodels.TokenError{Message: "failed to get token", Err: &alexaapimodels.AuthenticationError{
			Status: authErr.Status, Message: "provider authentication failed",
		}}
	}

	return &alexaapimodels.TokenError{Message: "failed to get token", Err: ctx.Err()}
}
