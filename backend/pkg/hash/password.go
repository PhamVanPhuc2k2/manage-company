package hash

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var ErrMismatch = errors.New("mật khẩu không đúng")

func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("băm mật khẩu: %w", err)
	}
	return string(b), nil
}

func Verify(hashed, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrMismatch
	}
	return err
}

func NeedsRehash(hashed string) bool {
	c, err := bcrypt.Cost([]byte(hashed))
	return err == nil && c < bcryptCost
}

func DummyVerify(plain string) {
	const dummyHash = "$2a$12$C6UzMDM.H6dfI/f/IKcEe.7Zfi2H8wHy9rLMkhuGCYYzjKdWGwSxG"
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(plain))
}
