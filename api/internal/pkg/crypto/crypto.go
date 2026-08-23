// Package crypto provides helpers for hashing, token generation, and
// encryption of sensitive data at rest.
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12
)

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(b), nil
}

// CheckPassword returns nil if plain matches the hashed password.
func CheckPassword(plain, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// GenerateOTP returns a numeric OTP of the given length.
func GenerateOTP(length int) (string, error) {
	otp := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generating OTP digit: %w", err)
		}
		otp += n.String()
	}
	return otp, nil
}

// GenerateToken returns a cryptographically random URL-safe token of the given byte length.
func GenerateToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// SHA256Hex returns the hex-encoded SHA-256 hash of data.
// Used for document hashes in agreements.
func SHA256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// GenerateIdempotencyKey returns a random idempotency key suitable
// for financial operations.
func GenerateIdempotencyKey() (string, error) {
	return GenerateToken(24)
}
