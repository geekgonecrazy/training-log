package router

import (
	"context"
	"strconv"

	"github.com/geekgonecrazy/training-log/core/auth"
)

// ctxUserID is a small adapter so router.go doesn't import core/auth's key directly.
func ctxUserID(ctx context.Context) (int64, bool) {
	uid, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return 0, false
	}
	return uid, true
}

func itoa(i int64) string { return strconv.FormatInt(i, 10) }
