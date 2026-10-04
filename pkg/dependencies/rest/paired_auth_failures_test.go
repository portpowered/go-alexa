//nolint:testpackage // Exercises private registration errors through complete paired synthetic exchanges.
package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func restAuthenticationFailureReplayCases(ctx context.Context, config *DeviceRegistrationConfig) []restReplayResponse {
	return []restReplayResponse{
		{
			operation: "registerDevice_email_otp",
			body:      `{"response":{"challenge":{"challenge_reason":"MissingRequiredAuthenticationData","required_authentication_method":"OTP"}}}`,
			status:    http.StatusBadRequest,
			headers:   nil,
			call: func(client *Client) error {
				_, err := client.RegisterWithEmailPassword(ctx, "synthetic@example.invalid", "synthetic-password", config)

				var challenge *RegistrationChallengeError

				if !errors.As(err, &challenge) || !challenge.IsOTPRequired() || challenge.IsCBLRequired() || challenge.IsAuthenticationFailed() {
					return syntheticFailureError("expected explicit OTP registration challenge")
				}

				return nil
			},
		},
		{
			operation: "registerDevice_codePair_rejected",
			body:      `{"message":"synthetic code-pair rejection"}`,
			status:    http.StatusUnauthorized,
			headers:   nil,
			call: func(client *Client) error {
				_, err := client.RegisterWithCodePair(ctx, "synthetic-public", "synthetic-private", config)
				if !alexaapimodels.IsNetworkError(err) {
					return syntheticFailureError("expected rejected code-pair network error")
				}

				return nil
			},
		},
		{
			operation: "registerDevice_email_rejected",
			body:      `{"message":"not a registration challenge"}`,
			status:    http.StatusBadRequest,
			headers:   nil,
			call: func(client *Client) error {
				_, err := client.RegisterWithEmailPassword(ctx, "synthetic@example.invalid", "synthetic-password", config)
				if !alexaapimodels.IsNetworkError(err) {
					return syntheticFailureError("expected rejected email registration network error")
				}

				return nil
			},
		},
	}
}
