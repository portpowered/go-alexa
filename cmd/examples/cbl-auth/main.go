// Package main demonstrates cookie-based Alexa authentication.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	deviceSerialLength      = 13
	deviceNameSuffixLength  = 3
	maximumEndpointsToPrint = 5
)

// generateRandomString generates a random string of specified length.
func generateRandomString(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	limit := big.NewInt(int64(len(charset)))
	value := make([]byte, length)

	for index := range value {
		randomIndex, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate random device identifier: %w", err)
		}

		value[index] = charset[randomIndex.Int64()]
	}

	return string(value), nil
}

func main() {
	ctx := context.Background()

	// Create Alexa client (no token needed for authentication endpoints)
	client, err := alexa.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Generate device identifiers
	deviceSerial, err := generateRandomString(deviceSerialLength)
	if err != nil {
		log.Fatalf("Generate device serial: %v", err)
	}

	deviceNameSuffix, err := generateRandomString(deviceNameSuffixLength)
	if err != nil {
		log.Fatalf("Generate device name suffix: %v", err)
	}

	deviceName := "Client_Ad_Hoc_" + deviceNameSuffix

	// Create device registration config
	config := alexaapimodels.DefaultDeviceRegistrationConfig(deviceSerial, deviceName)

	outputLine("=== Code-Based Linking (CBL) Authentication Example ===")
	outputLine()

	// Step 1: Generate code pair
	outputLine("Step 1: Generating code pair for CBL...")

	codePair, err := client.GenerateCodePair(ctx, config)
	if err != nil {
		log.Fatalf("Failed to generate code pair: %v", err)
	}

	outputf("✓ Code pair generated successfully!\n")
	outputf("  Public Code:  %s\n", codePair.PublicCode)
	outputLine()

	// Step 2: Prompt user to enter code on Amazon website
	outputLine("Step 2: Please visit https://amazon.com/code")
	outputf("   Enter the following code: %s\n", codePair.PublicCode)
	outputLine()
	outputText("Press Enter after you have entered the code on Amazon's website...")

	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	outputLine()

	// Step 3: Register using code pair
	outputLine("Step 3: Registering device with code pair...")

	registrationResp, err := client.RegisterWithCodePair(ctx, codePair.PublicCode, codePair.PrivateCode, config)
	if err != nil {
		log.Fatalf("Failed to register with code pair: %v", err)
	}

	refreshToken := registrationResp.RefreshToken

	outputLine("Device registered successfully; credential values are kept out of console output.")
	outputLine()

	// Step 4: Demonstrate token refresh
	outputLine("Step 4: Refreshing access token...")

	refreshReq := alexaapimodels.TokenRefreshRequest{
		RefreshToken: refreshToken,
		Config:       config,
	}

	refreshResp, err := client.RefreshAccessToken(ctx, refreshReq)
	if err != nil {
		log.Fatalf("Failed to refresh access token: %v", err)
	}

	newAccessToken := refreshResp.AccessToken

	outputLine("Access token refreshed successfully.")
	outputLine()

	// Step 5: Demonstrate using the token with the Alexa client
	outputLine("Step 5: Using the access token to list endpoints...")

	authenticatedClient, err := client.NewSession(
		alexa.WithBearerToken(newAccessToken),
	)
	if err != nil {
		log.Fatalf("Failed to create authenticated session: %v", err)
	}

	defer func() {
		_ = authenticatedClient.Close()
	}()

	endpoints, err := authenticatedClient.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
		IncludeFields: &alexaapimodels.EndpointIncludeFields{
			Properties: true,
			Features:   true,
		},
	})
	if err != nil {
		log.Printf("Warning: Failed to list endpoints (this is expected if you don't have any devices): %v", err)
	} else {
		printEndpoints(endpoints.Results)
	}

	printCompletion()
}

func printCompletion() {
	outputLine()
	outputLine("=== CBL Authentication Complete ===")
}

func printEndpoints(endpoints []*alexaapimodels.Endpoint) {
	outputf("✓ Found %d endpoints\n", len(endpoints))

	for index, endpoint := range endpoints {
		if index >= maximumEndpointsToPrint {
			outputf("  ... and %d more\n", len(endpoints)-maximumEndpointsToPrint)

			break
		}

		name := endpoint.EndpointID
		if endpoint.FriendlyName != nil {
			name = endpoint.FriendlyName.Value
		}

		outputf("  - %s (%s)\n", name, endpoint.ID)
	}
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

func outputText(values ...any) {
	_, err := fmt.Fprint(os.Stdout, values...)
	if err != nil {
		panic(err)
	}
}
