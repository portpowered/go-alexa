// Package cli implements the standalone customer command line interface.
package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

const (
	defaultRequestTimeout       = 30 * time.Second
	defaultEventDurationMinutes = 30
	maximumEventDurationMinutes = 1440
	credentialFileMode          = 0o600
	credentialDirectoryMode     = 0o700
	clientOptionCapacity        = 2
	maximumCredentialSize       = 1 << 20
)

var (
	errInvalidOptions            = errors.New("invalid command options")
	errCredentialsRequired       = errors.New("provide credentials with --credentials-file, --credentials-stdin, or ALEXA_ACCESS_TOKEN")
	errAccessTokenRequired       = errors.New("credentials do not contain an access token; run auth refresh explicitly")
	errRefreshTokenRequired      = errors.New("credentials do not contain a refresh token")
	errCommandFailed             = errors.New("go-alexa command failed")
	errUnknownCommand            = errors.New("unknown command; run go-alexa help")
	errUnsupportedRegion         = errors.New("unsupported region; use US, EU, or JP")
	errCredentialSource          = errors.New("choose either --credentials-file or --credentials-stdin")
	errCredentialSize            = errors.New("credential JSON could not be read or exceeds the size limit")
	errCredentialDecode          = errors.New("could not decode credential JSON")
	errCredentialTrailing        = errors.New("credential JSON contains trailing data")
	errCredentialOutput          = errors.New("--credentials-out is required")
	errApprovalInputEnded        = errors.New("approval input ended before confirmation")
	errDeviceRegistrationMissing = errors.New("credential data must include deviceRegistration.deviceSerial for refresh")
	errPowerTargetRequired       = errors.New("--id is required for device control")
	errPowerStateRequired        = errors.New("--state must be on or off")
	errEndpointNotFound          = errors.New("endpoint was not found; run endpoints list to check its ID")
	errPowerFeatureMissing       = errors.New("endpoint does not advertise the power feature")
	errControlFeatureErrors      = errors.New("endpoint control returned feature errors")
	errPlayerTargetRequired      = errors.New("--id is required")
	errPlayerIdentityMissing     = errors.New("endpoint is missing player-state device identity; refresh endpoint data and try again")
	errEventDurationRange        = errors.New("--duration-minutes must be between 1 and 1440")
	errSubscriptionRejected      = errors.New("subscription rejected")
)

var (
	errCredentialFileRequired = errors.New("--credentials-file is required; auth logout removes a local credential file only")
	errEnvironmentCredentials = errors.New(
		"environment credentials cannot be removed by auth logout; " +
			"unset ALEXA_ACCESS_TOKEN and ALEXA_REFRESH_TOKEN in your shell",
	)
	errCredentialPathDirectory = errors.New("auth logout path must name a file, not a directory")
)

// App connects command input and output to the public Alexa SDK.
type App struct {
	input         io.Reader
	output        io.Writer
	errorOutput   io.Writer
	lookupEnv     func(string) (string, bool)
	clientOptions []alexa.Option
	secrets       []string
}

// New creates a CLI application. clientOptions are public SDK options and are
// primarily useful for callers that provide their own transports.
func New(input io.Reader, output, errorOutput io.Writer, clientOptions ...alexa.Option) *App {
	return &App{
		input:         input,
		output:        output,
		errorOutput:   errorOutput,
		lookupEnv:     os.LookupEnv,
		clientOptions: clientOptions,
		secrets:       nil,
	}
}

// Run executes one CLI command. Errors are redacted before they are returned so
// the main program can report them without printing credentials.
func (a *App) Run(ctx context.Context, args []string) (runErr error) {
	a.secrets = nil

	defer func() {
		if errors.Is(runErr, flag.ErrHelp) {
			runErr = nil
		}

		if errors.Is(runErr, context.Canceled) {
			runErr = context.Canceled

			return
		}

		if runErr != nil {
			runErr = fmt.Errorf("%w: %s", errCommandFailed, a.redact(runErr.Error()))
		}
	}()

	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return a.writeRootHelp()
	}

	if len(args) > 1 && (args[1] == "help" || args[1] == "--help" || args[1] == "-h") {
		return a.writeGroupHelp(args[0])
	}

	switch args[0] {
	case "auth":
		return a.runAuth(ctx, args[1:])
	case "endpoints":
		return a.runEndpoints(ctx, args[1:])
	case "endpoint":
		return a.runEndpoint(ctx, args[1:])
	case "player":
		return a.runPlayer(ctx, args[1:])
	case "events":
		return a.runEvents(ctx, args[1:])
	default:
		return fmt.Errorf("%w: %q", errUnknownCommand, args[0])
	}
}

func (a *App) writeRootHelp() error {
	_, err := fmt.Fprint(a.output, `go-alexa — use the public go-alexa SDK from your terminal

Usage:
  go-alexa auth link --credentials-out FILE
  go-alexa auth refresh --credentials-file FILE --credentials-out FILE
  go-alexa auth export --credentials-stdin --credentials-out FILE
  go-alexa auth logout --credentials-file FILE
  go-alexa endpoints list [credential flags]
  go-alexa endpoint power --id ID --state on|off [credential flags]
  go-alexa player state --id ID [credential flags]
  go-alexa events listen [credential flags]

Credential flags for authenticated commands:
  --credentials-file FILE   Read access and refresh tokens from a file.
  --credentials-stdin       Read a JSON credential object from stdin.
  Otherwise use ALEXA_ACCESS_TOKEN and ALEXA_REFRESH_TOKEN.

Secrets are never accepted as command arguments or printed in normal output.
Credential files are created with owner-only permissions and are never overwritten.
Logout removes only the named local file; unset environment credentials in your shell.
`)
	if err != nil {
		return fmt.Errorf("write help: %w", err)
	}

	return nil
}

func (a *App) writeGroupHelp(command string) error {
	help := map[string]string{
		"auth": "Usage: go-alexa auth link|refresh|export|logout\n" +
			"Link accounts, explicitly refresh access tokens, export credentials, or remove a local credential file.\n",
		"endpoints": "Usage: go-alexa endpoints list [flags]\n" +
			"List account endpoints and optionally include current feature states.\n",
		"endpoint": "Usage: go-alexa endpoint power [flags]\n" +
			"Send an on or off command to an endpoint with the power feature.\n",
		"player": "Usage: go-alexa player state [flags]\n" +
			"Read player state for a listed endpoint.\n",
		"events": "Usage: go-alexa events listen [flags]\n" +
			"Subscribe to endpoint events and write event JSON lines until cancellation or timeout.\n",
	}
	if commandHelp, ok := help[command]; ok {
		_, err := fmt.Fprint(a.output, commandHelp)
		if err != nil {
			return fmt.Errorf("write command help: %w", err)
		}

		return nil
	}

	return a.writeRootHelp()
}

func (a *App) flagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {
		_, _ = fmt.Fprintf(a.errorOutput, "Usage: %s\n", name)

		flags.PrintDefaults()
	}

	return flags
}

func (a *App) parseFlags(flags *flag.FlagSet, args []string) error {
	err := flags.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		return flag.ErrHelp
	}

	if err != nil || len(flags.Args()) > 0 {
		return errInvalidOptions
	}

	return nil
}

func writeJSON(output io.Writer, value any) error {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(value)
	if err != nil {
		return fmt.Errorf("write JSON output: %w", err)
	}

	return nil
}

func (a *App) newClient(regionName string) (*alexa.Client, error) {
	region, err := parseRegion(regionName)
	if err != nil {
		return nil, err
	}

	options := make([]alexa.Option, 0, len(a.clientOptions)+clientOptionCapacity)
	options = append(options, alexa.WithRegion(region), alexa.WithTimeout(defaultRequestTimeout))
	options = append(options, a.clientOptions...)

	client, err := alexa.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("create Alexa client: %w", err)
	}

	return client, nil
}

func parseRegion(value string) (alexaapimodels.Region, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "US":
		return alexaapimodels.RegionUS, nil
	case "EU":
		return alexaapimodels.RegionEU, nil
	case "JP":
		return alexaapimodels.RegionJP, nil
	default:
		return "", fmt.Errorf("%w: %q", errUnsupportedRegion, value)
	}
}

type credentialFlags struct {
	file  string
	stdin bool
}

func addCredentialFlags(flags *flag.FlagSet) *credentialFlags {
	credentials := &credentialFlags{file: "", stdin: false}
	flags.StringVar(&credentials.file, "credentials-file", "", "read credentials from this file")
	flags.BoolVar(&credentials.stdin, "credentials-stdin", false, "read a JSON credential object from stdin")

	return credentials
}

type credentialFile struct {
	AccessToken        string                                  `json:"accessToken,omitempty"`
	RefreshToken       string                                  `json:"refreshToken,omitempty"`
	DeviceRegistration alexaapimodels.DeviceRegistrationConfig `json:"deviceRegistration,omitempty"`
}

func (a *App) readCredentials(source *credentialFlags) (credentialFile, error) {
	if source.file != "" && source.stdin {
		return credentialFile{}, errCredentialSource
	}

	var (
		credentials credentialFile
		err         error
	)

	switch {
	case source.file != "":
		input, err := os.Open(source.file)
		if err != nil {
			return credentialFile{}, fmt.Errorf("open credential file: %w", err)
		}

		defer func() { _ = input.Close() }()

		err = decodeCredentials(io.LimitReader(input, maximumCredentialSize), &credentials)
		if err != nil {
			return credentialFile{}, err
		}
	case source.stdin:
		err = decodeCredentials(io.LimitReader(a.input, maximumCredentialSize), &credentials)
		if err != nil {
			return credentialFile{}, err
		}
	default:
		credentials.AccessToken, _ = a.lookupEnv("ALEXA_ACCESS_TOKEN")
		credentials.RefreshToken, _ = a.lookupEnv("ALEXA_REFRESH_TOKEN")
		serial, _ := a.lookupEnv("ALEXA_DEVICE_SERIAL")

		name, _ := a.lookupEnv("ALEXA_DEVICE_NAME")
		if serial != "" {
			if name == "" {
				name = "go-alexa CLI"
			}

			credentials.DeviceRegistration = alexaapimodels.DefaultDeviceRegistrationConfig(serial, name)
		}
	}

	a.addSecret(credentials.AccessToken)
	a.addSecret(credentials.RefreshToken)

	if credentials.AccessToken == "" && credentials.RefreshToken == "" {
		return credentialFile{}, errCredentialsRequired
	}

	return credentials, nil
}

func decodeCredentials(input io.Reader, credentials *credentialFile) error {
	data, err := io.ReadAll(io.LimitReader(input, maximumCredentialSize+1))
	if err != nil || len(data) > maximumCredentialSize {
		return errCredentialSize
	}

	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(credentials)
	if err != nil {
		return errCredentialDecode
	}

	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errCredentialTrailing
	}

	return nil
}

func (a *App) newSession(client *alexa.Client, credentials credentialFile) (*alexa.Session, error) {
	if credentials.AccessToken == "" {
		return nil, errAccessTokenRequired
	}

	options := []alexa.SessionOption{alexa.WithBearerToken(credentials.AccessToken)}
	if credentials.RefreshToken != "" {
		options = append(options, alexa.WithRefreshToken(credentials.RefreshToken))
	}

	session, err := client.NewSession(options...)
	if err != nil {
		return nil, fmt.Errorf("create Alexa session: %w", err)
	}

	return session, nil
}

func (a *App) addSecret(secret string) {
	if secret != "" {
		a.secrets = append(a.secrets, secret)
	}
}

func (a *App) redact(message string) string {
	for _, secret := range a.secrets {
		message = strings.ReplaceAll(message, secret, "[REDACTED]")
	}

	return message
}

func saveCredentials(path string, credentials credentialFile) error {
	if strings.TrimSpace(path) == "" {
		return errCredentialOutput
	}

	directory := filepath.Dir(path)
	if directory != "." {
		err := os.MkdirAll(directory, credentialDirectoryMode)
		if err != nil {
			return fmt.Errorf("create credential directory: %w", err)
		}
	}

	file, err := openCredentialFile(path)
	if err != nil {
		return fmt.Errorf("create credential file: %w", err)
	}

	err = restrictCredentialFile(file)
	if err != nil {
		_ = file.Close()
		_ = os.Remove(path)

		return fmt.Errorf("restrict credential file permissions: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	writeErr := encoder.Encode(credentials)
	closeErr := file.Close()

	if writeErr != nil {
		_ = os.Remove(path)

		return fmt.Errorf("write credential file: %w", writeErr)
	}

	if closeErr != nil {
		_ = os.Remove(path)

		return fmt.Errorf("close credential file: %w", closeErr)
	}

	return nil
}

func awaitApproval(ctx context.Context, input io.Reader) error {
	completed := make(chan error, 1)

	go func() {
		_, err := bufio.NewReader(input).ReadString('\n')
		completed <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("wait for approval: %w", ctx.Err())
	case err := <-completed:
		if errors.Is(err, io.EOF) {
			return errApprovalInputEnded
		}

		if err != nil {
			return fmt.Errorf("read approval confirmation: %w", err)
		}
	}

	return nil
}

func (a *App) parseCredentialFlags(flags *flag.FlagSet, source *credentialFlags, args []string) error {
	err := a.parseFlags(flags, args)
	if errors.Is(err, flag.ErrHelp) {
		return err
	}

	if err != nil {
		return err
	}

	if source.file != "" && source.stdin {
		return errCredentialSource
	}

	return nil
}
