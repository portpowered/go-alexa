//go:build !windows

package cli

import "os"

func openCredentialFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, credentialFileMode)
}

func restrictCredentialFile(*os.File) error { return nil }
