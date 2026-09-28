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

	options := make([]alexa.Option, 0, 3)
	if bearerToken != "" {
		options = append(options, alexa.WithBearerToken(bearerToken))
	}
	if refreshToken != "" {
		options = append(options, alexa.WithRefreshToken(refreshToken))
	}
	if customerID := os.Getenv("ALEXA_CUSTOMER_ID"); customerID != "" {
		options = append(options, alexa.WithCustomerID(customerID))
	}

	client, err := alexa.NewClient(options...)
	if err != nil {
		log.Fatalf("Create Alexa client: %v", err)
	}
	defer func() { _ = client.Close() }()

	response, err := client.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
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

	_, err = client.Control(ctx, alexaapimodels.ControlRequest{
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
