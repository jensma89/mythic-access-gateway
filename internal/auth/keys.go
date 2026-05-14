// keys.go
// Generate API key, hashing, validation

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// keyPrefix makes gateway key recognizable at a glance
	keyPrefix = "mag_" // Mythic Access Gateway

	// bcryptCost - 12 is a good balance between security and speed
	bcryptCost = 12

	// rawKeyBytes - 32 random bytes >> 64 hex chars >> total key ~68 chars
	rawKeyBytes = 32
)

// GenerateAPIKey creates a new random API key in the format:
// mag_<64 hex characters>
// The raw key is returned once and never stored - only its hash is saved.
func GenerateAPIKey() (string, error) {
	b := make([]byte, rawKeyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return keyPrefix + hex.EncodeToString(b), nil
}

// HashAPIKey returns a bcrypt hash of the given key.
// Store this hash in the database, never the raw key.
func HashAPIKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash api key: %w", err)
	}
	return string(hash), nil
}

// ValidateAPIKey checks whether rawKey matches the stored bcrypt hash.
func ValidateAPIKey(rawKey, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawKey))
	return err == nil
}

// GenerateVerificationToken creates a random hex token for email verification.
func GenerateVerificationToken() (string, error) {
	b := make([]byte, 16) // 16 bytes >> 32 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate verification token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
