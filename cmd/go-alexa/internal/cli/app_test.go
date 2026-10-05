package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/portpowered/go-alexa/cmd/go-alexa/internal/cli"
	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

const (
	syntheticAccessToken  = "synthetic-access-token"
	syntheticRefreshToken = "synthetic-refresh-token"
	syntheticPrivateCode  = "synthetic-private-code"
)

type authLogoutJSON struct {
	CredentialsFile                  string `json:"credentialsFile"`
	Removed                          bool   `json:"removed"`
	EnvironmentCredentialsConfigured bool   `json:"environmentCredentialsConfigured"`
}

func TestAuthLinkUsesPairedRequestsAndWritesPrivateCredentials(t *testing.T) {
	t.Parallel()

	fixture := loadExchanges(t, "alexa-rest.json")
	codePair := exchangeByOperation(t, fixture, "createCodePair")
	registration := exchangeByOperation(t, fixture, "registerDevice_codePair")

	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"app_name":"synthetic-app"`, `"app_name":"Client SDK"`)
	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"app_version":"1"`, `"app_version":"1.0"`)
	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"device_model":""`, `"device_model":"Client SDK"`)
	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"device_name":""`, `"device_name":"Synthetic device"`)
	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"device_type":"synthetic-device"`, `"device_type":"A2IVLV5VM2W81"`)
	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"domain":""`, `"domain":"Device"`)
	codePair.Request.Body = replaceExpected(t, codePair.Request.Body,
		`"os_version":""`, `"os_version":"0"`)
	codePair.Response.Body = `{"private_code":"synthetic-private-code","public_code":"synthetic-public-code"}`

	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"private_code":"synthetic-private"`, `"private_code":"synthetic-private-code"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"public_code":"synthetic-public"`, `"public_code":"synthetic-public-code"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"app_name":"synthetic-app"`, `"app_name":"Client SDK"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"app_version":"1"`, `"app_version":"1.0"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"device_model":""`, `"device_model":"Client SDK"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"device_name":""`, `"device_name":"Synthetic device"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"device_type":"synthetic-device"`, `"device_type":"A2IVLV5VM2W81"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"domain":""`, `"domain":"Device"`)
	registration.Request.Body = replaceExpected(t, registration.Request.Body,
		`"os_version":""`, `"os_version":"0"`)
	registration.Response.Body = `{"response":{"success":{"tokens":{"bearer":{"access_token":"synthetic-access-token",` +
		`"refresh_token":"synthetic-refresh-token"}}}}}`

	player := writeAndLoadReplay(t, []replay.SyntheticExchange{codePair, registration})
	credentialPath := filepath.Join(t.TempDir(), "linked-credentials.json")

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	application := cli.New(strings.NewReader("\n"), &stdout, &stderr,
		alexa.WithAmazonAPIBaseURL("https://auth.synthetic.test"),
		alexa.WithHTTPClient(testHTTPClient(player)),
	)

	err := application.Run(context.Background(), []string{
		"auth", "link", "--device-serial", "synthetic-serial", "--device-name", "Synthetic device",
		"--credentials-out", credentialPath,
	})
	if err != nil {
		t.Fatalf("auth link returned error: %v; stderr=%q", err, stderr.String())
	}

	if strings.Contains(stdout.String(), syntheticPrivateCode) || strings.Contains(stdout.String(), syntheticAccessToken) ||
		strings.Contains(stdout.String(), syntheticRefreshToken) {
		t.Fatal("auth link printed a private code or token")
	}

	if !strings.Contains(stdout.String(), "synthetic-public-code") || !strings.Contains(stdout.String(), alexaapimodels.AuthorizationUriNa) {
		t.Fatalf("auth link did not print the user approval code and URL: %s", stdout.String())
	}

	credentials := readCredentialFile(t, credentialPath)
	if credentials.AccessToken != syntheticAccessToken || credentials.RefreshToken != syntheticRefreshToken {
		t.Fatalf("saved credentials = %#v", credentials)
	}

	info, statErr := os.Stat(credentialPath)
	if statErr != nil {
		t.Fatalf("stat credential file: %v", statErr)
	}

	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("credential file permissions = %v; want 0600", info.Mode().Perm())
	}

	assertConsumed(t, player)
}

func TestAuthRefreshReplaysExactRequestAndRedactsFailure(t *testing.T) {
	t.Parallel()

	fixture := loadExchanges(t, "alexa-rest.json")
	refresh := exchangeByOperation(t, fixture, "refreshAccessToken")
	refresh.Response.Body = `{"access_token":"synthetic-access-token","expires_in":3600}`
	credentialPath := filepath.Join(t.TempDir(), "input.json")
	writeCredentials(t, credentialPath, credentialDocument{
		AccessToken:  "old-synthetic-access",
		RefreshToken: "synthetic-refresh",
		DeviceRegistration: alexaapimodels.DeviceRegistrationConfig{
			AppName: "synthetic-app", AppVersion: "1", DeviceType: "synthetic-device",
			Domain: "", DeviceModel: "", OSVersion: "", DeviceSerial: "synthetic-serial",
			DeviceName: "", Manufacturer: "Synthetic",
		},
	})

	outputPath := filepath.Join(t.TempDir(), "refreshed.json")

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		player := writeAndLoadReplay(t, []replay.SyntheticExchange{refresh})

		var stdout bytes.Buffer

		application := cli.New(strings.NewReader(""), &stdout, io.Discard,
			alexa.WithAmazonAPIBaseURL("https://auth.synthetic.test"),
			alexa.WithHTTPClient(testHTTPClient(player)),
		)

		err := application.Run(context.Background(), []string{
			"auth", "refresh", "--credentials-file", credentialPath, "--credentials-out", outputPath,
		})
		if err != nil {
			t.Fatalf("auth refresh returned error: %v", err)
		}

		credentials := readCredentialFile(t, outputPath)
		if credentials.AccessToken != syntheticAccessToken || credentials.RefreshToken != "synthetic-refresh" {
			t.Fatalf("refreshed credentials = %#v", credentials)
		}

		if strings.Contains(stdout.String(), syntheticAccessToken) || strings.Contains(stdout.String(), "synthetic-refresh") {
			t.Fatal("auth refresh printed a token")
		}

		assertConsumed(t, player)
	})

	t.Run("provider rejects refresh", func(t *testing.T) {
		t.Parallel()

		failure := refresh
		failure.Response.Status = http.StatusUnauthorized
		failure.Response.Body = `{"message":"synthetic-refresh"}`
		player := writeAndLoadReplay(t, []replay.SyntheticExchange{failure})
		output := filepath.Join(t.TempDir(), "failed-refresh.json")

		var stdout bytes.Buffer

		application := cli.New(strings.NewReader(""), &stdout, io.Discard,
			alexa.WithAmazonAPIBaseURL("https://auth.synthetic.test"),
			alexa.WithHTTPClient(testHTTPClient(player)),
		)

		err := application.Run(context.Background(), []string{
			"auth", "refresh", "--credentials-file", credentialPath, "--credentials-out", output,
		})
		if err == nil || strings.Contains(err.Error(), "synthetic-refresh") || strings.Contains(stdout.String(), "synthetic-refresh") {
			t.Fatalf("auth error leaked credentials or succeeded: err=%v stdout=%q", err, stdout.String())
		}

		_, statErr := os.Stat(output)
		if !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("failed refresh created credential output: %v", statErr)
		}

		assertConsumed(t, player)
	})
}

func TestAuthLogoutRemovesCredentialFileAndIsIdempotent(t *testing.T) {
	//nolint:gosec // This is a synthetic environment value used to verify logout does not alter the caller's environment.
	const environmentAccessToken = "synthetic-environment-access-token"

	t.Setenv("ALEXA_ACCESS_TOKEN", environmentAccessToken)
	t.Setenv("ALEXA_REFRESH_TOKEN", "")

	credentialPath := filepath.Join(t.TempDir(), "credentials.json")

	var stdout bytes.Buffer

	application := cli.New(
		strings.NewReader(`{"accessToken":"synthetic-access-token","refreshToken":"synthetic-refresh-token"}`),
		&stdout,
		io.Discard,
	)

	err := application.Run(context.Background(), []string{
		"auth", "export", "--credentials-stdin", "--credentials-out", credentialPath,
	})
	if err != nil {
		t.Fatalf("create protected credential file: %v", err)
	}

	info, err := os.Stat(credentialPath)
	if err != nil {
		t.Fatalf("stat exported credential file: %v", err)
	}

	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("credential file permissions = %v; want 0600", info.Mode().Perm())
	}

	firstResult := runAuthLogout(t, application, &stdout, credentialPath)

	if firstResult.CredentialsFile != credentialPath || !firstResult.Removed || !firstResult.EnvironmentCredentialsConfigured {
		t.Fatalf("first logout result = %#v", firstResult)
	}

	if os.Getenv("ALEXA_ACCESS_TOKEN") != environmentAccessToken {
		t.Fatal("auth logout changed the caller's environment credentials")
	}

	_, err = os.Stat(credentialPath)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("credential file remains after logout: %v", err)
	}

	repeatedResult := runAuthLogout(t, application, &stdout, credentialPath)

	if repeatedResult.CredentialsFile != credentialPath || repeatedResult.Removed || !repeatedResult.EnvironmentCredentialsConfigured {
		t.Fatalf("repeated logout result = %#v", repeatedResult)
	}
}

func runAuthLogout(t *testing.T, application *cli.App, output *bytes.Buffer, credentialPath string) authLogoutJSON {
	t.Helper()
	output.Reset()

	err := application.Run(context.Background(), []string{"auth", "logout", "--credentials-file", credentialPath})
	if err != nil {
		t.Fatalf("auth logout returned error: %v", err)
	}

	var result authLogoutJSON

	err = json.Unmarshal(output.Bytes(), &result)
	if err != nil {
		t.Fatalf("decode auth logout result: %v", err)
	}

	return result
}

func TestAuthLogoutExplainsEnvironmentOnlyCredentials(t *testing.T) {
	t.Setenv("ALEXA_ACCESS_TOKEN", syntheticAccessToken)
	t.Setenv("ALEXA_REFRESH_TOKEN", "")

	var stdout bytes.Buffer

	application := cli.New(strings.NewReader(""), &stdout, io.Discard)
	err := application.Run(context.Background(), []string{"auth", "logout"})

	if err == nil || !strings.Contains(err.Error(), "unset ALEXA_ACCESS_TOKEN") {
		t.Fatalf("environment-only logout error = %v, want instructions to unset environment credentials", err)
	}

	if stdout.Len() != 0 || os.Getenv("ALEXA_ACCESS_TOKEN") != syntheticAccessToken {
		t.Fatalf("environment-only logout changed output or the caller's environment: stdout=%q token=%q",
			stdout.String(), os.Getenv("ALEXA_ACCESS_TOKEN"))
	}

	if strings.Contains(err.Error(), syntheticAccessToken) {
		t.Fatal("environment-only logout error disclosed the access token")
	}
}

func TestAuthLogoutIsShownInHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer

	application := cli.New(strings.NewReader(""), &stdout, io.Discard)

	for _, args := range [][]string{{"help"}, {"auth", "help"}} {
		stdout.Reset()

		err := application.Run(context.Background(), args)
		if err != nil {
			t.Fatalf("help command %v returned error: %v", args, err)
		}

		want := "go-alexa auth logout --credentials-file FILE"
		if args[0] == "auth" {
			want = "logout"
		}

		if !strings.Contains(stdout.String(), want) {
			t.Errorf("help command %v did not describe auth logout: %s", args, stdout.String())
		}
	}
}

func TestEndpointListUsesPairedGraphQLAndRESTExchanges(t *testing.T) {
	t.Parallel()

	graphqlFixture := loadExchanges(t, "alexa-graphql.json")
	endpoints := exchangeByOperation(t, graphqlFixture, "Endpoints")
	endpoints.Request.Body = replaceExpected(t, endpoints.Request.Body,
		`"featureLatencyTolerance":""`, `"featureLatencyTolerance":"LOW"`)
	endpoints.Request.Body = replaceExpected(t, endpoints.Request.Body,
		`"disablePagination":false,"nextToken":"synthetic-next","pageSize":0`,
		`"disablePagination":true,"nextToken":"","pageSize":50`)
	list := exchangeByOperation(t, graphqlFixture, "ListEndpoints")
	list.Request.Body = replaceExpected(t, list.Request.Body,
		`"variables":{"input":{"endpointIds":["synthetic-endpoint"],"filterExpressions":null,`+
			`"paginationParams":{"disablePagination":false,"nextToken":"","pageSize":0}}}`,
		`"variables":{"input":{"displayCategory":"AIR_QUALITY_MONITOR","filterExpressions":null,`+
			`"paginationParams":{"disablePagination":true,"nextToken":"","pageSize":50}}}`)
	graphqlReplay := writeAndLoadReplay(t, []replay.SyntheticExchange{endpoints, list})

	restFixture := loadExchanges(t, "alexa-rest.json")
	devices := exchangeByOperation(t, restFixture, "listFirstPartyDevices")
	devices.Request.Headers.Set("Authorization", "Bearer synthetic-graphql-token")
	devices.Request.Headers.Del("Cookie")
	restReplay := writeAndLoadReplay(t, []replay.SyntheticExchange{devices})
	credentialPath := filepath.Join(t.TempDir(), "credentials.json")
	writeCredentials(t, credentialPath, credentialDocument{
		AccessToken: "synthetic-graphql-token", RefreshToken: "",
		DeviceRegistration: emptyDeviceRegistration(),
	})

	var stdout bytes.Buffer

	application := cli.New(strings.NewReader(""), &stdout, io.Discard,
		alexa.WithGraphQLBaseURL("https://graphql.synthetic.test"),
		alexa.WithAlexaWebBaseURL("https://web.synthetic.test"),
		alexa.WithGraphQLHTTPClient(testHTTPClient(graphqlReplay)),
		alexa.WithRESTHTTPClient(testHTTPClient(restReplay)),
	)

	err := application.Run(context.Background(), []string{
		"endpoints", "list", "--credentials-file", credentialPath,
	})
	if err != nil {
		t.Fatalf("endpoints list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"results":[]`) {
		t.Fatalf("endpoint output = %s", stdout.String())
	}

	assertConsumed(t, graphqlReplay)
	assertConsumed(t, restReplay)
}

func TestEndpointPowerUsesPairedDiscoveryAndControlExchanges(t *testing.T) {
	t.Parallel()

	graphqlPairs := endpointFlowGraphQLPairs(t, true)
	graphqlReplay := writeAndLoadReplay(t, graphqlPairs)
	devices := endpointFlowDeviceExchange(t)
	restReplay := writeAndLoadReplay(t, []replay.SyntheticExchange{devices})
	credentialPath := filepath.Join(t.TempDir(), "credentials.json")
	writeCredentials(t, credentialPath, credentialDocument{
		AccessToken: "synthetic-graphql-token", RefreshToken: "",
		DeviceRegistration: emptyDeviceRegistration(),
	})

	var stdout bytes.Buffer

	application := cli.New(strings.NewReader(""), &stdout, io.Discard,
		alexa.WithGraphQLBaseURL("https://graphql.synthetic.test"),
		alexa.WithAlexaWebBaseURL("https://web.synthetic.test"),
		alexa.WithGraphQLHTTPClient(testHTTPClient(graphqlReplay)),
		alexa.WithRESTHTTPClient(testHTTPClient(restReplay)),
	)

	err := application.Run(context.Background(), []string{
		"endpoint", "power", "--id", "synthetic-endpoint", "--state", "on", "--credentials-file", credentialPath,
	})
	if err != nil {
		t.Fatalf("endpoint power returned error: %v", err)
	}

	if stdout.String() != "{}\n" {
		t.Fatalf("endpoint power output = %s", stdout.String())
	}

	assertConsumed(t, graphqlReplay)
	assertConsumed(t, restReplay)
}

func TestPlayerStateUsesPairedDiscoveryAndPlayerExchanges(t *testing.T) {
	t.Parallel()

	graphqlReplay := writeAndLoadReplay(t, endpointFlowGraphQLPairs(t, false))
	devices := endpointFlowDeviceExchange(t)
	player := exchangeByOperation(t, loadExchanges(t, "alexa-rest.json"), "getMediaPlayerState")
	player.Request.Headers = http.Header{
		"Accept":        {"application/json"},
		"Authorization": {"Bearer synthetic-graphql-token"},
	}
	player.Response.Body = `{"playerInfo":{"state":"PLAYING","provider":{"providerName":"Synthetic source"}}}`
	restReplay := writeAndLoadReplay(t, []replay.SyntheticExchange{devices, player})
	credentialPath := filepath.Join(t.TempDir(), "credentials.json")
	writeCredentials(t, credentialPath, credentialDocument{
		AccessToken: "synthetic-graphql-token", RefreshToken: "",
		DeviceRegistration: emptyDeviceRegistration(),
	})

	var stdout bytes.Buffer

	application := cli.New(strings.NewReader(""), &stdout, io.Discard,
		alexa.WithGraphQLBaseURL("https://graphql.synthetic.test"),
		alexa.WithAlexaWebBaseURL("https://web.synthetic.test"),
		alexa.WithGraphQLHTTPClient(testHTTPClient(graphqlReplay)),
		alexa.WithRESTHTTPClient(testHTTPClient(restReplay)),
	)

	err := application.Run(context.Background(), []string{
		"player", "state", "--id", "synthetic-endpoint", "--credentials-file", credentialPath,
	})
	if err != nil {
		t.Fatalf("player state returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"state":"PLAYING"`) ||
		!strings.Contains(stdout.String(), `"providerName":"Synthetic source"`) {
		t.Fatalf("player state output = %s", stdout.String())
	}

	assertConsumed(t, graphqlReplay)
	assertConsumed(t, restReplay)
}

func TestEventsListenCancelsAndClosesPairedEventStream(t *testing.T) {
	t.Parallel()

	graphqlFixture := loadExchanges(t, "alexa-graphql.json")
	subscribe := exchangeByOperation(t, graphqlFixture, "Subscribe")
	subscribe.Request.Body = replaceExpected(t, subscribe.Request.Body,
		`"entityType":"endpoint","ids":["synthetic-endpoint"]`,
		`"entityType":"Endpoint","ids":null`)
	subscribe.Request.Body = replaceExpected(t, subscribe.Request.Body,
		`"format":"","reset":false,"version":""`,
		`"format":"GraphQL","reset":true,"version":"2"`)
	graphqlReplay := writeAndLoadReplay(t, []replay.SyntheticExchange{subscribe})

	eventPairs := loadExchanges(t, "alexa-events.json")
	for index := range eventPairs {
		eventPairs[index].Request.Headers.Set("Authorization", "Bearer synthetic-graphql-token")
	}

	eventReplay := writeAndLoadReplay(t, eventPairs)
	trackedEventReplay := &trackedReplayTransport{
		replay: eventReplay, pinged: make(chan struct{}, 1), bodyClosed: make(chan struct{}),
	}
	credentialPath := filepath.Join(t.TempDir(), "credentials.json")
	writeCredentials(t, credentialPath, credentialDocument{
		AccessToken: "synthetic-graphql-token", RefreshToken: "",
		DeviceRegistration: emptyDeviceRegistration(),
	})

	output := &eventOutput{mu: sync.Mutex{}, buffer: bytes.Buffer{}, eventSeen: make(chan struct{}, 1), once: sync.Once{}}
	application := cli.New(strings.NewReader(""), output, io.Discard,
		alexa.WithGraphQLBaseURL("https://graphql.synthetic.test"),
		alexa.WithEventAuthority("events.synthetic.test"),
		alexa.WithGraphQLHTTPClient(testHTTPClient(graphqlReplay)),
		alexa.WithEventTransport(trackedEventReplay),
	)
	ctx, cancel := context.WithCancel(context.Background())
	completed := make(chan error, 1)

	go func() {
		completed <- application.Run(ctx, []string{
			"events", "listen", "--duration-minutes", "5", "--credentials-file", credentialPath,
		})
	}()

	select {
	case <-output.eventSeen:
	case <-time.After(3 * time.Second):
		cancel()
		t.Fatal("event stream did not emit the synthetic event")
	}

	select {
	case <-trackedEventReplay.pinged:
	case <-time.After(3 * time.Second):
		cancel()
		t.Fatal("event stream did not send its paired keepalive request")
	}

	if !strings.Contains(output.String(), "synthetic-message") {
		t.Fatalf("event output did not contain the paired message: %s", output.String())
	}

	cancel()

	select {
	case err := <-completed:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("events listen error = %v, want canceled context", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("events listen did not stop after cancellation")
	}

	select {
	case <-trackedEventReplay.bodyClosed:
	case <-time.After(3 * time.Second):
		t.Fatal("event response body was not closed")
	}

	assertConsumed(t, graphqlReplay)
	assertConsumed(t, eventReplay)
}

func loadExchanges(t *testing.T, name string) []replay.SyntheticExchange {
	t.Helper()

	path := filepath.Join("..", "..", "..", "..", "tests", "replay", "fixtures", "synthetic", name)
	//nolint:gosec // Test callers pass only the checked-in synthetic fixture names.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read paired synthetic fixture %s: %v", name, err)
	}

	var exchanges []replay.SyntheticExchange

	err = json.Unmarshal(data, &exchanges)
	if err != nil {
		t.Fatalf("decode paired synthetic fixture %s: %v", name, err)
	}

	return exchanges
}

func endpointFlowGraphQLPairs(t *testing.T, includeControl bool) []replay.SyntheticExchange {
	t.Helper()

	fixture := loadExchanges(t, "alexa-graphql.json")
	endpoints := exchangeByOperation(t, fixture, "Endpoints")
	endpoints.Request.Body = replaceExpected(t, endpoints.Request.Body,
		`"featureLatencyTolerance":""`, `"featureLatencyTolerance":"LOW"`)
	endpoints.Request.Body = replaceExpected(t, endpoints.Request.Body,
		`"disablePagination":false,"nextToken":"synthetic-next","pageSize":0`,
		`"disablePagination":true,"nextToken":"","pageSize":50`)
	endpoints.Response.Body = `{"data":{"endpoints":{"items":[` +
		`{"id":"synthetic-endpoint","endpointId":"synthetic-endpoint",` +
		`"dmsIdentifier":{"deviceType":"synthetic-device","dsn":"synthetic-serial"},` +
		`"features":[{"name":"power","instance":"","operations":[{"name":"turnOn"}]}]}` +
		`]}}}`

	list := exchangeByOperation(t, fixture, "ListEndpoints")
	list.Request.Body = replaceExpected(t, list.Request.Body,
		`"variables":{"input":{"endpointIds":["synthetic-endpoint"],"filterExpressions":null,`+
			`"paginationParams":{"disablePagination":false,"nextToken":"","pageSize":0}}}`,
		`"variables":{"input":{"displayCategory":"AIR_QUALITY_MONITOR","filterExpressions":null,`+
			`"paginationParams":{"disablePagination":true,"nextToken":"","pageSize":50}}}`)

	pairs := []replay.SyntheticExchange{endpoints, list}
	if includeControl {
		pairs = append(pairs, exchangeByOperation(t, fixture, "SetEndpointFeatures"))
	}

	return pairs
}

func endpointFlowDeviceExchange(t *testing.T) replay.SyntheticExchange {
	t.Helper()

	devices := exchangeByOperation(t, loadExchanges(t, "alexa-rest.json"), "listFirstPartyDevices")
	devices.Request.Headers.Set("Authorization", "Bearer synthetic-graphql-token")
	devices.Request.Headers.Del("Cookie")

	return devices
}

func exchangeByOperation(t *testing.T, exchanges []replay.SyntheticExchange, operation string) replay.SyntheticExchange {
	t.Helper()

	for _, exchange := range exchanges {
		if exchange.Operation == operation {
			return exchange
		}
	}

	t.Fatalf("paired fixture does not contain operation %q", operation)

	return replay.SyntheticExchange{
		Source: "", Operation: "",
		Request: replay.SyntheticRequest{
			Method: "", Origin: "", Host: "", EscapedPath: "", Query: nil, Headers: nil, Body: "",
		},
		Response: replay.SyntheticResponse{Status: 0, Headers: nil, Body: ""},
	}
}

func testHTTPClient(transport http.RoundTripper) *http.Client {
	return &http.Client{CheckRedirect: nil, Jar: nil, Timeout: 0, Transport: transport}
}

func emptyDeviceRegistration() alexaapimodels.DeviceRegistrationConfig {
	return alexaapimodels.DeviceRegistrationConfig{
		AppName: "", AppVersion: "", DeviceType: "", Domain: "", DeviceModel: "",
		OSVersion: "", DeviceSerial: "", DeviceName: "", Manufacturer: "",
	}
}

func replaceExpected(t *testing.T, value, old, replacement string) string {
	t.Helper()

	if !strings.Contains(value, old) {
		t.Fatalf("paired fixture does not contain %q", old)
	}

	return strings.Replace(value, old, replacement, 1)
}

func writeAndLoadReplay(t *testing.T, exchanges []replay.SyntheticExchange) *replay.SyntheticReplay {
	t.Helper()

	path := filepath.Join(t.TempDir(), "pairs.json")

	err := replay.WriteSyntheticReplay(path, exchanges)
	if err != nil {
		t.Fatalf("write paired synthetic replay: %v", err)
	}

	player, err := replay.LoadSyntheticReplay(path)
	if err != nil {
		t.Fatalf("load paired synthetic replay: %v", err)
	}

	return player
}

func assertConsumed(t *testing.T, player *replay.SyntheticReplay) {
	t.Helper()

	err := player.AssertConsumed()
	if err != nil {
		t.Fatalf("paired request/response exchanges were not consumed: %v", err)
	}
}

type credentialDocument struct {
	AccessToken        string                                  `json:"accessToken,omitempty"`
	RefreshToken       string                                  `json:"refreshToken,omitempty"`
	DeviceRegistration alexaapimodels.DeviceRegistrationConfig `json:"deviceRegistration,omitempty"`
}

func writeCredentials(t *testing.T, path string, credentials credentialDocument) {
	t.Helper()

	data, err := json.Marshal(credentials)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		t.Fatalf("write synthetic credential fixture: %v", err)
	}
}

func readCredentialFile(t *testing.T, path string) credentialDocument {
	t.Helper()

	//nolint:gosec // Test paths point into t.TempDir and contain synthetic credentials only.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read credential file: %v", err)
	}

	var credentials credentialDocument

	err = json.Unmarshal(data, &credentials)
	if err != nil {
		t.Fatalf("decode credential file: %v", err)
	}

	return credentials
}

type trackedReplayTransport struct {
	replay     *replay.SyntheticReplay
	pinged     chan struct{}
	bodyClosed chan struct{}
}

func (transport *trackedReplayTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.replay.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("synthetic replay request: %w", err)
	}

	if request.URL.Path == "/ping" {
		select {
		case transport.pinged <- struct{}{}:
		default:
		}
	}

	if request.URL.Path == "/v20160207/directives" {
		response.Body = &heldReplayBody{ReadCloser: response.Body, closed: transport.bodyClosed, once: sync.Once{}}
	}

	return response, nil
}

type heldReplayBody struct {
	io.ReadCloser

	closed chan struct{}
	once   sync.Once
}

func (body *heldReplayBody) Read(buffer []byte) (int, error) {
	count, err := body.ReadCloser.Read(buffer)
	if errors.Is(err, io.EOF) {
		<-body.closed

		return 0, fmt.Errorf("read held replay body: %w", io.EOF)
	}

	if err != nil {
		return count, fmt.Errorf("read held replay body: %w", err)
	}

	return count, nil
}

func (body *heldReplayBody) Close() error {
	body.once.Do(func() { close(body.closed) })

	err := body.ReadCloser.Close()
	if err != nil {
		return fmt.Errorf("close held replay body: %w", err)
	}

	return nil
}

type eventOutput struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	eventSeen chan struct{}
	once      sync.Once
}

func (output *eventOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()

	count, err := output.buffer.Write(data)

	if bytes.Contains(data, []byte("synthetic-message")) {
		output.once.Do(func() { close(output.eventSeen) })
	}

	if err != nil {
		return count, fmt.Errorf("write synthetic event output: %w", err)
	}

	return count, nil
}

func (output *eventOutput) String() string {
	output.mu.Lock()
	defer output.mu.Unlock()

	return output.buffer.String()
}
