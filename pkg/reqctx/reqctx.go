// Package reqctx stores and retrieves per-request values from context.
package reqctx

import "context"

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserID returns a new context with the authenticated user's ID attached.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserID returns the authenticated user's ID from the context.
func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}
