package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func (a *App) runAuth(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errInvalidOptions
	}

	switch args[0] {
	case "login":
		return a.runCodeLink(ctx, args[1:], true)
	case "link":
		return a.runCodeLink(ctx, args[1:], false)
	case "refresh":
		return a.runAuthRefresh(ctx, args[1:])
	case "export":
		return a.runAuthExport(args[1:])
	case "logout":
		return a.runAuthLogout(args[1:])
	default:
		return errInvalidOptions
	}
}

func (a *App) runCodeLink(ctx context.Context, args []string, login bool) error {
	usage := "go-alexa auth link --credentials-out FILE [--device-serial ID] [--device-name NAME] [--region US|EU|JP]"
	if login {
		usage = "alexa auth login [--credentials-out FILE] [--device-serial ID] [--device-name NAME] [--region US|EU|JP]"
	}

	flags := a.flagSet(usage)
	outputPath := flags.String("credentials-out", "", "new owner-only credential file (must not already exist)")
	serial := flags.String("device-serial", "", "device serial identifier; generated when omitted")
	name := flags.String("device-name", "", "device name shown in account settings; generated uniquely when omitted")

	regionName := flags.String("region", "US", "service region: US, EU, or JP")
	recordDir := flags.String("record-dir", "", "directory for private request/response captures")

	err := a.parseFlags(flags, args)
	if err != nil {
		return err
	}

	if login && *outputPath == "" {
		*outputPath, err = a.defaultCredentialPath()
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(*outputPath) == "" {
		return errCredentialOutput
	}

	if *serial == "" {
		generated, err := randomDeviceSerial()
		if err != nil {
			return fmt.Errorf("generate device identifier: %w", err)
		}

		*serial = generated
	}

	client, err := a.newRecordedClient(*regionName, *recordDir)
	if err != nil {
		return err
	}

	config := alexaapimodels.DefaultDeviceRegistrationConfig(*serial, codeLinkDeviceName(*name, *serial))

	pair, err := client.GenerateCodePair(ctx, config)
	if err != nil {
		return fmt.Errorf("generate login code: %w", err)
	}

	a.addSecret(pair.PrivateCode)

	err = writeJSON(a.output, struct {
		PublicCode       string `json:"publicCode"`
		AuthorizationURL string `json:"authorizationUrl"`
		Next             string `json:"next"`
	}{
		PublicCode:       pair.PublicCode,
		AuthorizationURL: "https://www.amazon.com/code",
		Next:             "approve this code in your browser, then press Enter here",
	})
	if err != nil {
		return err
	}

	err = awaitApproval(ctx, a.input)
	if err != nil {
		return err
	}

	registration, err := client.RegisterWithCodePair(ctx, pair.PublicCode, pair.PrivateCode, config)
	if err != nil {
		return fmt.Errorf("register linked device: %w", err)
	}

	a.addSecret(registration.AccessToken)
	a.addSecret(registration.RefreshToken)

	if registration.AccessToken == "" || registration.RefreshToken == "" {
		return errIncompleteRegistration
	}

	credentials := credentialFile{
		AccessToken:        registration.AccessToken,
		RefreshToken:       registration.RefreshToken,
		DeviceRegistration: config,
	}

	err = a.saveLinkedCredentials(ctx, client, *outputPath, credentials, login)
	if err != nil {
		return err
	}

	return writeJSON(a.output, struct {
		Linked          bool   `json:"linked"`
		CredentialsFile string `json:"credentialsFile"`
	}{Linked: true, CredentialsFile: *outputPath})
}

func (a *App) runAuthRefresh(ctx context.Context, args []string) error {
	flags := a.flagSet("alexa auth refresh [--credentials-file FILE] [--credentials-out FILE] [--region US|EU|JP]")
	credentialsSource := addCredentialFlags(flags)
	outputPath := flags.String("credentials-out", "", "new owner-only credential file (must not already exist)")

	regionName := flags.String("region", "US", "service region: US, EU, or JP")
	recordDir := flags.String("record-dir", "", "directory for private request/response captures")

	err := a.parseCredentialFlags(flags, credentialsSource, args)
	if err != nil {
		return err
	}

	credentials, err := a.readCredentials(credentialsSource)
	if err != nil {
		return err
	}

	if *outputPath == "" {
		if credentialsSource.file == "" {
			return errCredentialOutput
		}

		*outputPath = credentialsSource.file
	}

	if credentials.RefreshToken == "" {
		return errRefreshTokenRequired
	}

	if credentials.DeviceRegistration.DeviceSerial == "" {
		return errDeviceRegistrationMissing
	}

	client, err := a.newRecordedClient(*regionName, *recordDir)
	if err != nil {
		return err
	}

	refreshed, err := client.RefreshAccessToken(ctx, alexaapimodels.TokenRefreshRequest{
		RefreshToken: credentials.RefreshToken,
		Config:       credentials.DeviceRegistration,
	})
	if err != nil {
		return fmt.Errorf("refresh access token: %w", err)
	}

	if refreshed.AccessToken == "" || refreshed.ExpiresInSeconds <= 0 {
		return errIncompleteRefresh
	}

	a.addSecret(refreshed.AccessToken)

	credentials.AccessToken = refreshed.AccessToken

	if sameCredentialPath(*outputPath, credentialsSource.file) {
		err = replaceCredentials(*outputPath, credentials)
	} else {
		err = saveCredentials(*outputPath, credentials)
	}

	if err != nil {
		return err
	}

	return writeJSON(a.output, struct {
		Refreshed       bool   `json:"refreshed"`
		CredentialsFile string `json:"credentialsFile"`
	}{Refreshed: true, CredentialsFile: *outputPath})
}

func (a *App) runAuthExport(args []string) error {
	flags := a.flagSet(
		"go-alexa auth export (--credentials-file FILE|--credentials-stdin|environment) --credentials-out FILE",
	)
	credentialsSource := addCredentialFlags(flags)

	outputPath := flags.String("credentials-out", "", "new owner-only credential file (must not already exist)")

	err := a.parseCredentialFlags(flags, credentialsSource, args)
	if err != nil {
		return err
	}

	if strings.TrimSpace(*outputPath) == "" {
		return errCredentialOutput
	}

	credentials, err := a.readCredentials(credentialsSource)
	if err != nil {
		return err
	}

	err = saveCredentials(*outputPath, credentials)
	if err != nil {
		return err
	}

	return writeJSON(a.output, struct {
		Exported        bool   `json:"exported"`
		CredentialsFile string `json:"credentialsFile"`
	}{Exported: true, CredentialsFile: *outputPath})
}

func (a *App) runAuthLogout(args []string) error {
	flags := a.flagSet("go-alexa auth logout --credentials-file FILE")
	credentialsPath := flags.String("credentials-file", "", "local credential file to remove")

	err := a.parseFlags(flags, args)
	if err != nil {
		return err
	}

	if strings.TrimSpace(*credentialsPath) == "" {
		if a.hasEnvironmentCredentials() {
			return errEnvironmentCredentials
		}

		*credentialsPath, err = a.defaultCredentialPath()
		if err != nil {
			return err
		}
	}

	file, err := os.Lstat(*credentialsPath)
	if errors.Is(err, os.ErrNotExist) {
		return writeJSON(a.output, authLogoutResult{
			CredentialsFile:                  *credentialsPath,
			Removed:                          false,
			EnvironmentCredentialsConfigured: a.hasEnvironmentCredentials(),
		})
	}

	if err != nil {
		return fmt.Errorf("inspect credential file: %w", err)
	}

	if file.IsDir() {
		return errCredentialPathDirectory
	}

	err = os.Remove(*credentialsPath)
	if err != nil {
		return fmt.Errorf("remove credential file: %w", err)
	}

	return writeJSON(a.output, authLogoutResult{
		CredentialsFile:                  *credentialsPath,
		Removed:                          true,
		EnvironmentCredentialsConfigured: a.hasEnvironmentCredentials(),
	})
}

type authLogoutResult struct {
	CredentialsFile                  string `json:"credentialsFile"`
	Removed                          bool   `json:"removed"`
	EnvironmentCredentialsConfigured bool   `json:"environmentCredentialsConfigured"`
}

func (a *App) hasEnvironmentCredentials() bool {
	accessToken, _ := a.lookupEnv("ALEXA_ACCESS_TOKEN")
	refreshToken, _ := a.lookupEnv("ALEXA_REFRESH_TOKEN")

	return accessToken != "" || refreshToken != ""
}

func (a *App) runEndpoints(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "list" {
		return errInvalidOptions
	}

	flags := a.flagSet("go-alexa endpoints list [--include-states] [credential flags]")
	credentialsSource := addCredentialFlags(flags)
	regionName := flags.String("region", "US", "service region: US, EU, or JP")

	includeStates := flags.Bool("include-states", false, "include current feature property values")

	err := a.parseCredentialFlags(flags, credentialsSource, args[1:])
	if err != nil {
		return err
	}

	credentials, err := a.readCredentials(credentialsSource)
	if err != nil {
		return err
	}

	client, err := a.newClient(*regionName)
	if err != nil {
		return err
	}

	session, err := a.newSession(client, credentials)
	if err != nil {
		return err
	}

	defer func() { _ = session.Close() }()

	response, err := session.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
		EndpointIDs:   nil,
		IncludeFields: &alexaapimodels.EndpointIncludeFields{Features: true, Properties: *includeStates},
	})
	if err != nil {
		return fmt.Errorf("list endpoints: %w", err)
	}

	if response.Results == nil {
		response.Results = []*alexaapimodels.Endpoint{}
	}

	return writeJSON(a.output, response)
}

func (a *App) runEndpoint(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "power" {
		return errInvalidOptions
	}

	flags := a.flagSet("go-alexa endpoint power --id ID --state on|off [credential flags]")
	credentialsSource := addCredentialFlags(flags)
	regionName := flags.String("region", "US", "service region: US, EU, or JP")
	endpointID := flags.String("id", "", "endpoint ID from endpoints list")

	state := flags.String("state", "", "requested state: on or off")

	err := a.parseCredentialFlags(flags, credentialsSource, args[1:])
	if err != nil {
		return err
	}

	if strings.TrimSpace(*endpointID) == "" {
		return errPowerTargetRequired
	}

	var operation alexaapimodels.FeatureOperationName

	switch strings.ToLower(*state) {
	case "on":
		operation = alexaapimodels.FeatureOperationNameTurnOn
	case "off":
		operation = alexaapimodels.FeatureOperationNameTurnOff
	default:
		return errPowerStateRequired
	}

	credentials, err := a.readCredentials(credentialsSource)
	if err != nil {
		return err
	}

	client, err := a.newClient(*regionName)
	if err != nil {
		return err
	}

	session, err := a.newSession(client, credentials)
	if err != nil {
		return err
	}

	defer func() { _ = session.Close() }()

	listing, err := session.ListEndpoints(ctx, alexaapimodels.EndpointQuery{
		EndpointIDs:   nil,
		IncludeFields: &alexaapimodels.EndpointIncludeFields{Features: true, Properties: false},
	})
	if err != nil {
		return fmt.Errorf("find endpoint: %w", err)
	}

	endpoint := findEndpoint(listing.Results, *endpointID)
	if endpoint == nil {
		return errEndpointNotFound
	}

	if !endpoint.HasFeature(alexaapimodels.FeatureNamePower) {
		return errPowerFeatureMissing
	}

	response, err := session.Control(ctx, alexaapimodels.ControlRequest{
		Target:    endpoint,
		Namespace: alexaapimodels.FeatureNamePower,
		Name:      operation,
		Instance:  "",
		Payload:   alexaapimodels.ControlPowerPayload{},
	})
	if err != nil {
		return fmt.Errorf("control endpoint: %w", err)
	}

	if len(response.Errors) > 0 {
		err := writeJSON(a.output, response)
		if err != nil {
			return err
		}

		return fmt.Errorf("%w: %d", errControlFeatureErrors, len(response.Errors))
	}

	return writeJSON(a.output, response)
}

func (a *App) runPlayer(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "state" {
		return errInvalidOptions
	}

	flags := a.flagSet("go-alexa player state --id ID [credential flags]")
	credentialsSource := addCredentialFlags(flags)
	regionName := flags.String("region", "US", "service region: US, EU, or JP")

	endpointID := flags.String("id", "", "endpoint ID from endpoints list")

	err := a.parseCredentialFlags(flags, credentialsSource, args[1:])
	if err != nil {
		return err
	}

	if strings.TrimSpace(*endpointID) == "" {
		return errPlayerTargetRequired
	}

	credentials, err := a.readCredentials(credentialsSource)
	if err != nil {
		return err
	}

	client, err := a.newClient(*regionName)
	if err != nil {
		return err
	}

	session, err := a.newSession(client, credentials)
	if err != nil {
		return err
	}

	defer func() { _ = session.Close() }()

	listing, err := session.ListEndpoints(ctx, alexaapimodels.EndpointQuery{EndpointIDs: nil, IncludeFields: nil})
	if err != nil {
		return fmt.Errorf("find endpoint: %w", err)
	}

	endpoint := findEndpoint(listing.Results, *endpointID)
	if endpoint == nil {
		return errEndpointNotFound
	}

	if endpoint.DeviceSerialNumber == "" || endpoint.DeviceType == "" {
		return errPlayerIdentityMissing
	}

	response, err := session.GetPlayerState(ctx, alexaapimodels.PlayerStateRequest{Target: endpoint})
	if err != nil {
		return fmt.Errorf("get player state: %w", err)
	}

	return writeJSON(a.output, response)
}

func (a *App) runEvents(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "listen" {
		return errInvalidOptions
	}

	flags := a.flagSet("go-alexa events listen [--duration-minutes N] [credential flags]")
	credentialsSource := addCredentialFlags(flags)
	regionName := flags.String("region", "US", "service region: US, EU, or JP")

	duration := flags.Int("duration-minutes", defaultEventDurationMinutes, "subscription duration in minutes (1 to 1440)")

	err := a.parseCredentialFlags(flags, credentialsSource, args[1:])
	if err != nil {
		return err
	}

	if *duration < 1 || *duration > maximumEventDurationMinutes {
		return errEventDurationRange
	}

	credentials, err := a.readCredentials(credentialsSource)
	if err != nil {
		return err
	}

	client, err := a.newClient(*regionName)
	if err != nil {
		return err
	}

	session, err := a.newSession(client, credentials)
	if err != nil {
		return err
	}

	defer func() { _ = session.Close() }()

	subscription, err := session.Subscribe(ctx, alexaapimodels.SubscribeRequest{
		Entities:          []alexaapimodels.SubscribeEntity{{EntityType: alexaapimodels.EntityTypeEndpoint}},
		DurationInMinutes: *duration,
	})
	if err != nil {
		return fmt.Errorf("subscribe to endpoint events: %w", err)
	}

	if len(subscription.Errors) > 0 {
		return fmt.Errorf("%w: %s", errSubscriptionRejected, subscription.Errors[0].Message)
	}

	err = writeJSON(a.output, struct {
		Subscribed bool `json:"subscribed"`
	}{Subscribed: true})
	if err != nil {
		return err
	}

	listenCtx, cancel := context.WithTimeout(ctx, time.Duration(*duration)*time.Minute)
	defer cancel()

	connection, err := session.ConnectEvents(listenCtx)
	if err != nil {
		return fmt.Errorf("connect event stream: %w", err)
	}

	defer func() { _ = connection.Close() }()

	for {
		type receiveResult struct {
			event *alexaapimodels.Event
			err   error
		}

		received := make(chan receiveResult, 1)

		go func() {
			event, receiveErr := connection.Receive()
			received <- receiveResult{event: event, err: receiveErr}
		}()

		select {
		case <-listenCtx.Done():
			if ctx.Err() != nil {
				return fmt.Errorf("listen for events: %w", ctx.Err())
			}

			if errors.Is(listenCtx.Err(), context.DeadlineExceeded) {
				return nil
			}

			return fmt.Errorf("listen for events: %w", listenCtx.Err())
		case result := <-received:
			if result.err != nil {
				return fmt.Errorf("receive event: %w", result.err)
			}

			err = writeJSON(a.output, result.event)
			if err != nil {
				return err
			}
		}
	}
}

func codeLinkDeviceName(name, serial string) string {
	if name != "" {
		return name
	}

	return "go-alexa CLI " + serial
}

func findEndpoint(endpoints []*alexaapimodels.Endpoint, id string) *alexaapimodels.Endpoint {
	for _, endpoint := range endpoints {
		if endpoint != nil && (endpoint.EndpointID == id || endpoint.ID == id) {
			return endpoint
		}
	}

	return nil
}
