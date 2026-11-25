// Package token issues and verifies AES-256-GCM encrypted access tokens.
package token

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"jieyuc.cn/muxu/studio/internal/identity"
)

const maxTokenLength = 256

var decimalPattern = regexp.MustCompile(`^[0-9]+$`)

// ErrInvalidToken deliberately does not disclose why verification failed.
var ErrInvalidToken = errors.New("invalid or expired token")

// Manager is safe for concurrent use. Construct it with New.
// The wire format is RawURLEncoding(nonce || GCM ciphertext and tag).
// Plaintext is userID|expiration, where expiration is Unix seconds.
type Manager struct {
	aead cipher.AEAD
	ttl  time.Duration
	now  func() time.Time
}

// New requires a 32-byte random secret and a TTL of at least one second.
// Load the key from secure configuration; never embed it in source code.
func New(key []byte, ttl time.Duration) (*Manager, error) {
	if len(key) != 32 {
		return nil, errors.New("token key must contain exactly 32 bytes")
	}
	if ttl < time.Second {
		return nil, errors.New("token TTL must be at least one second")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Manager{aead: aead, ttl: ttl, now: time.Now}, nil
}

// Issue creates a token for a positive user ID. The returned expiration matches
// the token's second-precision expiration. Each issue uses a fresh random nonce.
func (m *Manager) Issue(ctx context.Context, userID int) (string, time.Time, error) {
	if err := ctx.Err(); err != nil {
		return "", time.Time{}, err
	}
	if userID <= 0 {
		return "", time.Time{}, errors.New("user ID must be positive")
	}
	expiresAt := time.Unix(m.now().Add(m.ttl).Unix(), 0).UTC()
	plaintext := strconv.Itoa(userID) + "|" + strconv.FormatInt(expiresAt.Unix(), 10)
	nonce := make([]byte, m.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", time.Time{}, err
	}
	sealed := m.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), expiresAt, nil
}

// Implements TokenVerifier
var _ identity.TokenVerifier = (*Manager)(nil)

// Verify implements identity.TokenVerifier structurally. It authenticates the
// ciphertext before parsing it and rejects IDs outside the platform int range.
func (m *Manager) Verify(ctx context.Context, token string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(token) == 0 || len(token) > maxTokenLength {
		return "", ErrInvalidToken
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || base64.RawURLEncoding.EncodeToString(raw) != token {
		return "", ErrInvalidToken
	}
	nonceSize := m.aead.NonceSize()
	if len(raw) < nonceSize+m.aead.Overhead() {
		return "", ErrInvalidToken
	}
	plaintext, err := m.aead.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", ErrInvalidToken
	}
	idText, expiryText, ok := strings.Cut(string(plaintext), "|")
	if !ok || !decimalPattern.MatchString(idText) || !decimalPattern.MatchString(expiryText) {
		return "", ErrInvalidToken
	}
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		return "", ErrInvalidToken
	}
	expiry, err := strconv.ParseInt(expiryText, 10, 64)
	if err != nil || expiry <= 0 || m.now().Unix() >= expiry {
		return "", ErrInvalidToken
	}
	return strconv.Itoa(id), nil
}
