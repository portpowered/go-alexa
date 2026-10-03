//go:build windows

package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/portpowered/go-alexa/cmd/go-alexa/internal/cli"
	"golang.org/x/sys/windows"
)

const windowsFileAllAccessMask = 0x001f01ff

func TestSavedCredentialFileHasProtectedOwnerOnlyDACL(t *testing.T) {
	t.Parallel()

	tempDirectory := t.TempDir()
	path := filepath.Join(tempDirectory, "credentials.json")

	var output bytes.Buffer

	app := cli.New(strings.NewReader(`{"accessToken":"synthetic-access"}`), &output, &output)

	err := app.Run(context.Background(), []string{
		"auth", "export", "--credentials-stdin", "--credentials-out", path,
	})
	if err != nil {
		t.Fatalf("export credential file: %v", err)
	}

	descriptor, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatalf("read credential file DACL: %v", err)
	}

	control, _, err := descriptor.Control()
	if err != nil {
		t.Fatalf("read credential file DACL control: %v", err)
	}

	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatal("credential file DACL inherits permissions from its directory")
	}

	acl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatalf("read credential file ACL: %v", err)
	}

	if acl == nil || acl.AceCount != 1 {
		t.Fatalf("credential file has %d ACL entries; want one owner entry", aceCount(acl))
	}

	var entry *windows.ACCESS_ALLOWED_ACE

	err = windows.GetAce(acl, 0, &entry)
	if err != nil {
		t.Fatalf("read credential file ACL entry: %v", err)
	}

	if entry.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || uint32(entry.Mask) != windowsFileAllAccessMask {
		t.Fatalf("credential file ACL entry = type %d mask %#x; want owner full access", entry.Header.AceType, entry.Mask)
	}

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatalf("get current Windows user: %v", err)
	}

	if !strings.Contains(descriptor.String(), user.User.Sid.String()) {
		t.Fatalf("credential file ACL does not grant access to current user: %s", descriptor)
	}

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("stat saved credential file: %v", err)
	}
}

func aceCount(acl *windows.ACL) uint16 {
	if acl == nil {
		return 0
	}

	return acl.AceCount
}
