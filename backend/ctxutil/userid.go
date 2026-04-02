package ctxutil

import (
	"context"
	"errors"
)

type contextKey string

const CtxKeyUserID contextKey = "userID"

// UserIDFromCtx extracts the userID injected by AuthMiddleware.
// JWT MapClaims stores numbers as float64.
func UserIDFromCtx(ctx context.Context) (int64, error) {
	v := ctx.Value(CtxKeyUserID)
	if v == nil {
		return 0, errors.New("userID not found in context")
	}
	switch id := v.(type) {
	case float64:
		return int64(id), nil
	case int64:
		return id, nil
	}
	return 0, errors.New("userID has unexpected type in context")
}
