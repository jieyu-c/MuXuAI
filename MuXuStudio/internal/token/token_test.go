package token

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"jieyuc.cn/muxu/studio/internal/identity"
)

var _ identity.TokenVerifier = (*Manager)(nil)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m, err := New([]byte("0123456789abcdef0123456789abcdef"), 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return time.Unix(1800000000, 123456789) }
	return m
}

func TestIssueAndVerify(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	token, expires, err := m.Issue(ctx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if expires.Unix() != 1800007200 || expires.Nanosecond() != 0 {
		t.Fatalf("expiration=%v", expires)
	}
	if id, err := m.Verify(ctx, token); err != nil || id != "42" {
		t.Fatalf("id=%q err=%v", id, err)
	}
	second, _, err := m.Issue(ctx, 42)
	if err != nil || second == token {
		t.Fatal("expected distinct tokens")
	}
	raw, _ := base64.RawURLEncoding.DecodeString(token)
	plain, err := m.aead.Open(nil, raw[:m.aead.NonceSize()], raw[m.aead.NonceSize():], nil)
	if err != nil || string(plain) != "42|1800007200" {
		t.Fatalf("plaintext=%q err=%v", plain, err)
	}
	m.now = func() time.Time { return expires.Add(-time.Nanosecond) }
	if _, err := m.Verify(ctx, token); err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return expires }
	if _, err := m.Verify(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expiration boundary: %v", err)
	}
}

func TestRejectInvalidTokens(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	valid, _, _ := m.Issue(ctx, 42)
	raw, _ := base64.RawURLEncoding.DecodeString(valid)
	raw[len(raw)-1] ^= 1
	cases := []string{"", strings.Repeat("a", 257), "%%%", valid + "=", valid + "\n", base64.RawURLEncoding.EncodeToString([]byte("short")), base64.RawURLEncoding.EncodeToString(raw)}
	for _, token := range cases {
		if id, err := m.Verify(ctx, token); id != "" || !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("invalid accepted: id=%q err=%v", id, err)
		}
	}
	other, err := New([]byte("abcdef0123456789abcdef0123456789"), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Verify(ctx, valid); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("wrong key: %v", err)
	}
}

func TestRejectMalformedPlaintext(t *testing.T) {
	m := newTestManager(t)
	for _, plain := range []string{
		"42", "42|1800007200|extra", "|1800007200", "0|1800007200", "-1|1800007200", "+1|1800007200", " 42|1800007200", "abc|1800007200", "9999999999999999999999|1800007200", "42|", "42|abc", "42|-1", "42|+1800007200", "42|0", "42|1800000000", "42|1799999999", "42|999999999999999999999",
	} {
		t.Run(plain, func(t *testing.T) {
			nonce := make([]byte, m.aead.NonceSize())
			if _, err := rand.Read(nonce); err != nil {
				t.Fatal(err)
			}
			token := base64.RawURLEncoding.EncodeToString(m.aead.Seal(nonce, nonce, []byte(plain), nil))
			if id, err := m.Verify(context.Background(), token); id != "" || !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("id=%q err=%v", id, err)
			}
		})
	}
}

func TestConfigurationAndIDBoundaries(t *testing.T) {
	for _, size := range []int{0, 16, 24, 31, 33} {
		if _, err := New(make([]byte, size), time.Hour); err == nil {
			t.Fatalf("key size %d accepted", size)
		}
	}
	for _, ttl := range []time.Duration{-time.Second, 0, time.Nanosecond} {
		if _, err := New(make([]byte, 32), ttl); err == nil {
			t.Fatalf("TTL %v accepted", ttl)
		}
	}
	m := newTestManager(t)
	for _, id := range []int{0, -1} {
		if _, _, err := m.Issue(context.Background(), id); err == nil {
			t.Fatalf("ID %d accepted", id)
		}
	}
	maxInt := int(^uint(0) >> 1)
	token, _, err := m.Issue(context.Background(), maxInt)
	if err != nil {
		t.Fatal(err)
	}
	if id, err := m.Verify(context.Background(), token); err != nil || id != strconv.Itoa(maxInt) {
		t.Fatalf("max int: id=%q err=%v", id, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := m.Issue(ctx, 42); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := m.Verify(ctx, token); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
