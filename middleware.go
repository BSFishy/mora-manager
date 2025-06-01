package main

import (
	"log/slog"
	"net/http"
)

func log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := LogFromCtx(ctx)
		logger = logger.With(slog.Group("request", "method", r.Method, "url", r.URL))

		logger.Info("handling request")

		r = r.WithContext(WithLogger(ctx, logger))
		next.ServeHTTP(w, r)
	})
}
