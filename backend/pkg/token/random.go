package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

func New() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("sinh token ngẫu nhiên: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Hash(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func NumericCode(digits int) (string, error) {
	if digits < 4 {
		digits = 4
	}
	const numbers = "0123456789"
	max := big.NewInt(int64(len(numbers)))

	b := make([]byte, digits)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("sinh mã OTP: %w", err)
		}
		b[i] = numbers[n.Int64()]
	}
	return string(b), nil
}

const passwordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789@#$%"

func RandomPassword(length int) (string, error) {
	if length < 12 {
		length = 12
	}
	max := big.NewInt(int64(len(passwordAlphabet)))

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("sinh mật khẩu ngẫu nhiên: %w", err)
		}
		b[i] = passwordAlphabet[n.Int64()]
	}
	return string(b), nil
}
