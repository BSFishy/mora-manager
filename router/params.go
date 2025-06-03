package router

import (
	"context"
	"maps"
	"net/http"
)

type contextKey string

const paramsKey contextKey = "params"

func WithParams(r *http.Request, newParams map[string]string) *http.Request {
	params := Params(r)
	maps.Copy(params, newParams)

	ctx := context.WithValue(r.Context(), paramsKey, params)
	return r.WithContext(ctx)
}

func Params(r *http.Request) map[string]string {
	val := r.Context().Value(paramsKey)
	if p, ok := val.(map[string]string); ok {
		return p
	}

	return map[string]string{}
}
