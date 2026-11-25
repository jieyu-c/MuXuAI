package identity

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
)

// TokenVerifier must validate token authenticity and expiration before returning
// a positive decimal user ID. Invalid tokens must return an error.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (userID string, err error)
}

// Authenticate validates a Bearer token and privately injects its verified user ID.
// A missing transport, nil verifier, malformed header or invalid identity is rejected.
func Authenticate(verifier TokenVerifier) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			unauthorized := func() (any, error) {
				return nil, errors.Unauthorized("UNAUTHORIZED", "Token 缺失、无效或已过期")
			}
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return unauthorized()
			}
			headers := tr.RequestHeader()
			if headers == nil {
				return unauthorized()
			}
			values := headers.Values("Authorization")
			if len(values) != 1 {
				return unauthorized()
			}
			parts := strings.Fields(values[0])
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || verifier == nil {
				return unauthorized()
			}
			userID, err := verifier.Verify(ctx, parts[1])
			if err != nil || userID == "" {
				return unauthorized()
			}
			id, err := strconv.Atoi(userID)
			if err != nil || id <= 0 {
				return unauthorized()
			}
			return next(withUserID(ctx, id), req)
		}
	}
}
