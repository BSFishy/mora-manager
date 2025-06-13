package util

import "context"

func Protect(ctx context.Context, fn func() error) {
	logger := LogFromCtx(ctx)

	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic", "r", r)
		}
	}()

	if err := fn(); err != nil {
		logger.Error("error", "err", err)
	}
}
