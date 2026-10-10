package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
)

func (a *App) refreshLoginToken(ctx context.Context, client *alexa.Client, credentials credentialFile) (string, error) {
	refreshed, err := client.RefreshAccessToken(ctx, alexaapimodels.TokenRefreshRequest{
		RefreshToken: credentials.RefreshToken, Config: credentials.DeviceRegistration,
	})
	if err != nil {
		return "", fmt.Errorf("refresh linked access token: %w", err)
	}

	if refreshed.AccessToken == "" || refreshed.ExpiresInSeconds <= 0 {
		return "", errIncompleteRefresh
	}

	a.addSecret(refreshed.AccessToken)

	return refreshed.AccessToken, nil
}

func (a *App) saveLinkedCredentials(ctx context.Context, client *alexa.Client, path string, credentials credentialFile, login bool) error {
	if !login {
		return saveCredentials(path, credentials)
	}

	accessToken, err := a.refreshLoginToken(ctx, client, credentials)
	if err != nil {
		return err
	}

	credentials.AccessToken = accessToken

	return replaceCredentials(path, credentials)
}

var (
	errIncompleteRegistration = errors.New("amazon returned incomplete credentials; start login again")
	errIncompleteRefresh      = errors.New("amazon returned an incomplete access token response")
	errCredentialNotRegular   = errors.New("saved credential path must be a regular file")
)

func (a *App) defaultCredentialPath() (string, error) {
	if path, ok := a.lookupEnv("GO_ALEXA_CREDENTIALS"); ok && path != "" {
		return path, nil
	}

	directory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user configuration directory: %w", err)
	}

	return filepath.Join(directory, "go-alexa", "credentials.json"), nil
}

// replaceCredentials writes a private sibling file before replacing the saved
// login, so a failed write leaves the previous credentials usable.
func replaceCredentials(path string, credentials credentialFile) error {
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect saved credential file: %w", err)
	}

	if info != nil && !info.Mode().IsRegular() {
		return errCredentialNotRegular
	}

	suffix, err := randomDeviceSerial()
	if err != nil {
		return fmt.Errorf("create credential temporary name: %w", err)
	}

	temporaryPath := path + "." + suffix + ".tmp"

	err = saveCredentials(temporaryPath, credentials)
	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(temporaryPath) }()

	err = os.Rename(temporaryPath, path)
	if err != nil {
		return fmt.Errorf("replace saved credential file: %w", err)
	}

	return nil
}

func sameCredentialPath(first, second string) bool {
	if second == "" {
		return false
	}

	firstPath, firstErr := filepath.Abs(first)
	secondPath, secondErr := filepath.Abs(second)

	return firstErr == nil && secondErr == nil && firstPath == secondPath
}
