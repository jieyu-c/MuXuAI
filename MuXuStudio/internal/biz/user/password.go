package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const maxPasswordBytes = 72

type Password string

// Hash converts a plaintext password into a salted hash for storage.
func (p Password) Hash() (Password, error) {
	if p == "" {
		return "", errors.New("password must not be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return Password(hash), nil
}

// Matches reports whether a plaintext password matches the stored hash p.
func (p Password) Matches(plaintext string) bool {
	if len(plaintext) == 0 || len(plaintext) > maxPasswordBytes {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(p), []byte(plaintext)) == nil
}
