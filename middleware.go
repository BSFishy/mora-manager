package main

import (
	"log/slog"
	"net/http"

	"github.com/BSFishy/mora-manager/util"
)

type statusCodeResponseWriter struct {
	inner      http.ResponseWriter
	statusCode int
}

func (s *statusCodeResponseWriter) Header() http.Header {
	return s.inner.Header()
}

func (s *statusCodeResponseWriter) Write(data []byte) (int, error) {
	return s.inner.Write(data)
}

func (s *statusCodeResponseWriter) WriteHeader(statusCode int) {
	s.statusCode = statusCode
	s.inner.WriteHeader(statusCode)
}

func log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := util.LogFromCtx(ctx)
		logger = logger.With(slog.Group("request", "method", r.Method, "url", r.URL.String()))

		logger.Info("handling request")

		r = r.WithContext(util.WithLogger(ctx, logger))

		writer := &statusCodeResponseWriter{
			inner:      w,
			statusCode: 200,
		}

		next.ServeHTTP(writer, r)

		logger.Info("completed request", slog.Group("response", "statusCode", writer.statusCode))
	})
}
