package auth

import (
	"context"
	"errors"
)

type ctxKey struct{}

var ErrNoUser = errors.New("no user in context")

// ContextWithUserID returns a derived context tagged with the user_id.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ctxKey{}, userID)
}

// UserIDFromContext returns the authenticated user_id, or an error if none.
func UserIDFromContext(ctx context.Context) (int64, error) {
	v, ok := ctx.Value(ctxKey{}).(int64)
	if !ok || v <= 0 {
		return 0, ErrNoUser
	}
	return v, nil
}
