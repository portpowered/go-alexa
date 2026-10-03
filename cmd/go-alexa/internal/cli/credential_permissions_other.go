//go:build !windows

package cli

import (
	"fmt"
	"os"
)

func openCredentialFile(path string) (*os.File, error) {
	//nolint:gosec // Caller-selected output; exclusive creation with owner-only mode.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, credentialFileMode)
	if err != nil {
		return nil, fmt.Errorf("create credential file: %w", err)
	}

	return file, nil
}

func restrictCredentialFile(*os.File) error { return nil }
