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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bearerToken := os.Getenv("ALEXA_BEARER_TOKEN")
	refreshToken := os.Getenv("ALEXA_REFRESH_TOKEN")
	if bearerToken == "" && refreshToken == "" {
		log.Fatal("Set ALEXA_BEARER_TOKEN or ALEXA_REFRESH_TOKEN")
	}

	client, err := alexa.NewClient()
	if err != nil {
		log.Fatalf("Create Alexa client: %v", err)
	}

	sessionOptions := make([]alexa.SessionOption, 0, 3)
	if bearerToken != "" {
		sessionOptions = append(sessionOptions, alexa.WithBearerToken(bearerToken))
	}
	if refreshToken != "" {
		sessionOptions = append(sessionOptions, alexa.WithRefreshToken(refreshToken))
	}
	if customerID := os.Getenv("ALEXA_CUSTOMER_ID"); customerID != "" {
		sessionOptions = append(sessionOptions, alexa.WithCustomerID(customerID))
	}

	session, err := client.NewSession(sessionOptions...)
	if err != nil {
		log.Fatalf("Create Alexa session: %v", err)
	}
	if bearerToken == "" {
		if refreshToken == "" {
			log.Fatal("Set ALEXA_BEARER_TOKEN or ALEXA_REFRESH_TOKEN")
		}
		cookies, err := session.ExchangeRefreshTokenForCookies(ctx, "amazon.com")
		if err != nil {
			log.Fatalf("Exchange refresh token for cookies: %v", err)
		}
		cookieOptions := []alexa.SessionOption{alexa.WithCookies(cookies), alexa.WithRefreshToken(refreshToken)}
		if customerID := os.Getenv("ALEXA_CUSTOMER_ID"); customerID != "" {
			cookieOptions = append(cookieOptions, alexa.WithCustomerID(customerID))
		}
		_ = session.Close()
		session, err = client.NewSession(cookieOptions...)
		if err != nil {
			log.Fatalf("Create cookie session: %v", err)
		}
		if _, err := session.GetCSRFToken(ctx); err != nil {
			log.Fatalf("Retrieve CSRF token: %v", err)
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
		log.Fatalf("List endpoints: %v", err)
	}
	fmt.Printf("Found %d endpoints\n", len(response.Results))

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
		return
	}

	_, err = session.Control(ctx, alexaapimodels.ControlRequest{
		Target:    endpoint,
		Namespace: alexaapimodels.FeatureNamePower,
		Name:      alexaapimodels.FeatureOperationNameTurnOff,
		Payload:   alexaapimodels.ControlPowerPayload{},
	})
	if err != nil {
		log.Fatalf("Send power-off request: %v", err)
	}
	fmt.Println("Power-off request sent")
}
