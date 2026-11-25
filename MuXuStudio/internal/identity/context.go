// Package identity carries authenticated user identity through a request context.
package identity

import "context"

type userIDKey struct{}

// withUserID is only called after successful authentication within this package.
func withUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserID returns the authenticated user ID, including from derived contexts.
// An absent or nonpositive ID is treated as unauthenticated.
func UserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey{}).(int)
	return id, ok && id > 0
}
