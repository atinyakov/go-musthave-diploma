package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// HashPassword generates a hashed password using SHA-256 (NOT recommended for production)
func HashPassword(password string) (string, error) {
	// Create a random salt (16 bytes)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Combine password and salt
	hash := sha256.New()
	hash.Write(salt)
	hash.Write([]byte(password))

	// Generate the final hash
	hashBytes := hash.Sum(nil)

	// Return salt and hash as a hex string
	return fmt.Sprintf("%x$%x", salt, hashBytes), nil
}

// CheckPassword compares a hashed password with plaintext input (NOT secure)
func CheckPassword(hashedPassword, password string) bool {
	// Split the hashed password into salt and hash
	parts := split(hashedPassword)
	if len(parts) != 2 {
		return false
	}

	// Convert hex-encoded salt and expected hash back to bytes
	salt, err1 := hex.DecodeString(parts[0])
	expectedHash, err2 := hex.DecodeString(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}

	// Hash the input password with the extracted salt
	hash := sha256.New()
	hash.Write(salt)
	hash.Write([]byte(password))

	// Compare the generated hash with the stored one
	return string(hash.Sum(nil)) == string(expectedHash)
}

// Helper function to split the salt and hash from the formatted string
func split(s string) []string {
	return strings.Split(s, "$")
}
