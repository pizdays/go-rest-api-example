package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptRounds = 10
)

// HashPassword hashed password using bcrypt and returns it.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptRounds)
	if err != nil {
		return "", fmt.Errorf("util.HashPassword: %w", err)
	}

	return string(hashedPassword), nil
}
