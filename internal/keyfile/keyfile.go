// Package keyfile initializes the deployment's persistent encryption key.
package keyfile

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

// Initialize preserves a valid existing key and creates a missing key exclusively.
// A failed or interrupted write is reported on the next initialization.
func Initialize(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if errors.Is(err, os.ErrExist) {
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		key, decodeErr := hex.DecodeString(strings.TrimSpace(string(b)))
		if decodeErr != nil || len(key) != 32 {
			return errors.New("existing key must contain 64 hexadecimal characters")
		}
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return err
	}
	if _, err = f.WriteString(hex.EncodeToString(key) + "\n"); err != nil {
		return err
	}
	return f.Sync()
}
