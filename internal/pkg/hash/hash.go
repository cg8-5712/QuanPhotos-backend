package hash

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = 10
)

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashToken generates a SHA-256 hash for tokens (refresh tokens, etc.)
// Use this instead of bcrypt for tokens because:
// 1. bcrypt has a 72-byte limit, JWT tokens are longer
// 2. Tokens are already random, don't need bcrypt's slow hashing
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CheckToken compares a token with its SHA-256 hash
func CheckToken(token, hash string) bool {
	return HashToken(token) == hash
}
