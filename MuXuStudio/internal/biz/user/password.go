package user

import (
	"golang.org/x/crypto/bcrypt"
)

const maxPasswordBytes = 72

type Password struct {
	Value string
}

// Hash converts a plaintext password into a salted hash for storage.
func (p *Password) Hash() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Value), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.Value = string(hash)
	return nil
}

// Matches reports whether a plaintext password matches the stored hash p.
func (p *Password) Matches(plaintext string) bool {
	if len(plaintext) == 0 || len(plaintext) > maxPasswordBytes {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(p.Value), []byte(plaintext)) == nil
}
