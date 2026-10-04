package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func randomDeviceSerial() (string, error) {
	var identifier [16]byte

	_, err := rand.Read(identifier[:])
	if err != nil {
		return "", fmt.Errorf("read random source: %w", err)
	}

	return "go-alexa-" + hex.EncodeToString(identifier[:]), nil
}
