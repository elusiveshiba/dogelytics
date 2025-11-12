package web

import (
	"context"

	"github.com/dogeorg/dogelytics/store"
)

type ctxKey string

const (
	ctxUserIDKey ctxKey = "user_id"
	ctxAPIKeyKey ctxKey = "api_key"
	ctxAPIKIDKey ctxKey = "api_kid"
)

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserIDKey, userID)
}

func getUserID(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserIDKey)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}

func withAPIKey(ctx context.Context, key store.APIKey) context.Context {
	return context.WithValue(ctx, ctxAPIKeyKey, key)
}

func getAPIKey(ctx context.Context) (store.APIKey, bool) {
	v := ctx.Value(ctxAPIKeyKey)
	if v == nil {
		return store.APIKey{}, false
	}
	k, ok := v.(store.APIKey)
	return k, ok
}
