package utils

import (
	"context"

	"github.com/ZiplEix/scrabble/api/middleware"
)

func GetUserID(ctx context.Context) (int64, bool) {
	val := ctx.Value(middleware.UserIDKey)
	id, ok := val.(int64)
	return id, ok
}
