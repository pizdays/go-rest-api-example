package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRandomString generates token based on crypto/rand package.
func GenerateRandomString(byteSize uint) (string, error) {
	b := make([]byte, byteSize)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("util.GenerateRandomString: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
