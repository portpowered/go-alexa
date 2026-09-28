package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	ctx := context.Background()

	// Create Alexa client (no token needed for authentication endpoints)
	client, err := alexa.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Generate device identifiers
	deviceSerial := generateRandomString(13)
	deviceName := fmt.Sprintf("Client_Ad_Hoc_%d", rand.Intn(1000)+1)

	// Create device registration config
	config := alexaapimodels.DefaultDeviceRegistrationConfig(deviceSerial, deviceName)

	fmt.Println("=== Code-Based Linking (CBL) Authentication Example ===")
	fmt.Println()

	// Step 1: Generate code pair
	fmt.Println("Step 1: Generating code pair for CBL...")
	codePair, err := client.GenerateCodePair(ctx, config)
	if err != nil {
		log.Fatalf("Failed to generate code pair: %v", err)
	}

	fmt.Printf("✓ Code pair generated successfully!\n")
	fmt.Printf("  Public Code:  %s\n", codePair.PublicCode)
	fmt.Println()

	// Step 2: Prompt user to enter code on Amazon website
	fmt.Println("Step 2: Please visit https://amazon.com/code")
	fmt.Printf("   Enter the following code: %s\n", codePair.PublicCode)
	fmt.Println()
	fmt.Print("Press Enter after you have entered the code on Amazon's website...")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	fmt.Println()

	// Step 3: Register using code pair
	fmt.Println("Step 3: Registering device with code pair...")
	registrationResp, err := client.RegisterWithCodePair(ctx, codePair.PublicCode, codePair.PrivateCode, config)
	if err != nil {
		log.Fatalf("Failed to register with code pair: %v", err)
	}

	refreshToken := registrationResp.RefreshToken

	fmt.Println("Device registered successfully; credential values are kept out of console output.")
	fmt.Println()

	// Step 4: Demonstrate token refresh
	fmt.Println("Step 4: Refreshing access token...")
	refreshReq := alexaapimodels.TokenRefreshRequest{
		RefreshToken: refreshToken,
		Config:       config,
	}
	refreshResp, err := client.RefreshAccessToken(ctx, refreshReq)
	if err != nil {
		log.Fatalf("Failed to refresh access token: %v", err)
	}

	newAccessToken := refreshResp.AccessToken
	fmt.Println("Access token refreshed successfully.")
	fmt.Println()

	// Step 5: Demonstrate using the token with the Alexa client
	fmt.Println("Step 5: Using the access token to list endpoints...")
	authenticatedClient, err := alexa.NewClient(
		alexa.WithBearerToken(newAccessToken),
		alexa.WithRegion(alexaapimodels.RegionUS),
	)
	if err != nil {
		log.Fatalf("Failed to create authenticated client: %v", err)
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
		fmt.Printf("✓ Found %d endpoints\n", len(endpoints.Results))
		for i, endpoint := range endpoints.Results {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(endpoints.Results)-5)
				break
			}
			name := endpoint.EndpointID
			if endpoint.FriendlyName != nil {
				name = endpoint.FriendlyName.Value
			}
			fmt.Printf("  - %s (%s)\n", name, endpoint.ID)
		}
	}

	fmt.Println()
	fmt.Println("=== CBL Authentication Complete ===")
}
