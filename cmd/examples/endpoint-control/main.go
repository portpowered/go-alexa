// Package main demonstrates controlling an Alexa endpoint.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	clientTimeout                     = 30 * time.Second
	sessionOptionCapacity             = 3
	errMissingCredentials staticError = "Set ALEXA_BEARER_TOKEN or ALEXA_REFRESH_TOKEN"
)

type staticError string

func (err staticError) Error() string { return string(err) }

func main() {
	runErr := run()
	if runErr != nil {
		log.Fatal(runErr)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	bearerToken := os.Getenv("ALEXA_BEARER_TOKEN")

	refreshToken := os.Getenv("ALEXA_REFRESH_TOKEN")

	if bearerToken == "" && refreshToken == "" {
		return errMissingCredentials
	}

	client, err := alexa.NewClient()
	if err != nil {
		return fmt.Errorf("create Alexa client: %w", err)
	}

	sessionOptions := sessionOptionsForCredentials(bearerToken, refreshToken, os.Getenv("ALEXA_CUSTOMER_ID"))

	session, err := client.NewSession(sessionOptions...)
	if err != nil {
		return fmt.Errorf("create Alexa session: %w", err)
	}

	if bearerToken == "" {
		cookies, err := session.ExchangeRefreshTokenForCookies(ctx, "amazon.com")
		if err != nil {
			return fmt.Errorf("exchange refresh token for cookies: %w", err)
		}

		cookieOptions := []alexa.SessionOption{
			alexa.WithCookies(cookies),
			alexa.WithRefreshToken(refreshToken),
		}
		if customerID := os.Getenv("ALEXA_CUSTOMER_ID"); customerID != "" {
			cookieOptions = append(cookieOptions, alexa.WithCustomerID(customerID))
		}

		_ = session.Close()

		session, err = client.NewSession(cookieOptions...)
		if err != nil {
			return fmt.Errorf("create cookie session: %w", err)
		}

		_, err = session.GetCSRFToken(ctx)
		if err != nil {
			return fmt.Errorf("retrieve CSRF token: %w", err)
		}
	}

	defer func() { _ = session.Close() }()

	response, err := session.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
		IncludeFields: &alexaapimodels.EndpointIncludeFields{
			Properties: true,
			Features:   true,
		},
	})
	if err != nil {
		return fmt.Errorf("list endpoints: %w", err)
	}

	outputf("Found %d endpoints\n", len(response.Results))

	var endpoint *alexaapimodels.Endpoint

	for _, candidate := range response.Results {
		for _, feature := range candidate.Features {
			if feature.IsType(alexaapimodels.FeatureNamePower) {
				endpoint = candidate

				break
			}
		}

		if endpoint != nil {
			break
		}
	}

	if endpoint == nil {
		log.Print("No power-capable endpoint found")

		return nil
	}

	_, err = session.Control(ctx, alexaapimodels.ControlRequest{
		Target:    endpoint,
		Namespace: alexaapimodels.FeatureNamePower,
		Name:      alexaapimodels.FeatureOperationNameTurnOff,
		Payload:   alexaapimodels.ControlPowerPayload{},
	})
	if err != nil {
		return fmt.Errorf("send power-off request: %w", err)
	}

	outputLine("Power-off request sent")

	return nil
}

func sessionOptionsForCredentials(bearerToken, refreshToken, customerID string) []alexa.SessionOption {
	sessionOptions := make([]alexa.SessionOption, 0, sessionOptionCapacity)
	if bearerToken != "" {
		sessionOptions = append(sessionOptions, alexa.WithBearerToken(bearerToken))
	}

	if refreshToken != "" {
		sessionOptions = append(sessionOptions, alexa.WithRefreshToken(refreshToken))
	}

	if customerID != "" {
		sessionOptions = append(sessionOptions, alexa.WithCustomerID(customerID))
	}

	return sessionOptions
}

func outputLine(values ...any) {
	_, err := fmt.Fprintln(os.Stdout, values...)
	if err != nil {
		panic(err)
	}
}

func outputf(format string, values ...any) {
	_, err := fmt.Fprintf(os.Stdout, format, values...)
	if err != nil {
		panic(err)
	}
}
