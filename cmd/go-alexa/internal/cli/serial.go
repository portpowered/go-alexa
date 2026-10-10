package cli

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func randomDeviceSerial() (string, error) {
	// Match the CBL example from portos-backend before SDK extraction.
	const (
		alphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		serialLength = 13
	)

	identifier := make([]byte, serialLength)

	limit := big.NewInt(int64(len(alphabet)))
	for index := range identifier {
		value, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("read random source: %w", err)
		}

		identifier[index] = alphabet[value.Int64()]
	}

	return string(identifier), nil
}
