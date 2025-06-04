package util

import (
	"net/http"
)

func Redirect(w http.ResponseWriter, location string) {
	w.Header().Set("location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func ErrorHandle(handler func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := LogFromCtx(ctx)

		err := handler(w, r)
		if err != nil {
			logger.Error("failed to handle route", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
