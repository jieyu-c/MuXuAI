package server

import (
	"encoding/base64"
	"errors"

	"jieyuc.cn/muxu/studio/internal/conf"
	"jieyuc.cn/muxu/studio/internal/token"
)

// NewTokenManager constructs the shared issuer/verifier from server configuration.
// Configuration errors stop startup; there is no fallback signing key.
func NewTokenManager(c *conf.Server) (*token.Manager, error) {
	settings := c.GetToken()
	if settings == nil || settings.GetTtl() == nil {
		return nil, errors.New("server.token and its ttl are required")
	}
	if err := settings.GetTtl().CheckValid(); err != nil {
		return nil, errors.New("server.token.ttl is invalid")
	}
	key, err := base64.StdEncoding.Strict().DecodeString(settings.GetKey())
	if err != nil || len(key) != 32 {
		return nil, errors.New("server.token.key must be a Base64-encoded 32-byte key")
	}
	return token.New(key, settings.GetTtl().AsDuration())
}
