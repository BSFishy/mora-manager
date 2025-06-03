package main

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strings"
)

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

type Middleware func(http.Handler) http.Handler

type patternType int

const (
	prefix  patternType = iota
	pattern             = iota
)

type route struct {
	patternType patternType
	method      *string
	pattern     string
	handler     http.Handler
}

type Router struct {
	routes      *[]route
	middlewares []Middleware

	// something something the root router will be empty so paths will always be
	// prefixed with /
	path string
}

func NewRouter() Router {
	return Router{
		routes: &[]route{},
	}
}

func match(patternType patternType, pattern, path string) (bool, map[string]string) {
	if patternType == prefix {
		return strings.HasPrefix(path, pattern), map[string]string{}
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false, nil
	}

	params := map[string]string{}
	for i := 1; i < len(patternParts); i++ {
		pp := patternParts[i]
		cp := pathParts[i]

		if strings.HasPrefix(pp, ":") {
			params[pp[1:]] = cp
		} else if pp != cp {
			return false, nil
		}
	}

	return true, params
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range *r.routes {
		if route.method != nil && *route.method != req.Method {
			continue
		}

		if ok, params := match(route.patternType, route.pattern, req.URL.Path); ok {
			req = WithParams(req, params)
			route.handler.ServeHTTP(w, req)
			return
		}
	}

	r.middlewareFunc(http.NotFound).ServeHTTP(w, req)
}

func (r *Router) ListenAndServe(addr string) error {
	slog.Info("listening for requests", "addr", addr)

	return http.ListenAndServe(addr, r)
}

func (r *Router) Use(m ...Middleware) *Router {
	return &Router{
		routes:      r.routes,
		middlewares: append(r.middlewares, m...),
		path:        r.path,
	}
}

func (r *Router) middlewareFunc(handler http.HandlerFunc) http.Handler {
	return r.middleware(handler)
}

func (r *Router) middleware(handler http.Handler) http.Handler {
	wrapped := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		wrapped = r.middlewares[i](wrapped)
	}

	return wrapped
}

func (r *Router) Register(method, path string, handler http.HandlerFunc) {
	Assert(len(path) >= 1, "path must not be empty")
	Assert(strings.HasPrefix(path, "/"), "path must start with /")

	routePath := fmt.Sprintf("%s/%s", r.path, path[1:])
	route := route{
		patternType: pattern,
		method:      &method,
		pattern:     routePath,
		handler:     r.middleware(handler),
	}

	*r.routes = append(*r.routes, route)
}

// does not support path parameters
func (r *Router) Prefix(path string, handler http.Handler) {
	Assert(len(path) >= 1, "path must not be empty")
	Assert(strings.HasPrefix(path, "/"), "path must start with /")

	routePath := fmt.Sprintf("%s/%s", r.path, path[1:])
	route := route{
		patternType: prefix,
		method:      nil,
		pattern:     routePath,
		handler:     r.middleware(handler),
	}

	*r.routes = append(*r.routes, route)
}

func (r *Router) Handle(method, path string, handle http.Handler) {
	r.Register(method, path, func(w http.ResponseWriter, r *http.Request) {
		handle.ServeHTTP(w, r)
	})
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.Register(http.MethodGet, path, handler)
}

func (r *Router) HandleGet(path string, handle http.Handler) {
	r.Handle(http.MethodGet, path, handle)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.Register(http.MethodPost, path, handler)
}

func (r *Router) Route(path string) *Router {
	Assert(len(path) >= 2, "path must not be empty")
	Assert(strings.HasPrefix(path, "/"), "path must start with /")

	return &Router{
		routes:      r.routes,
		middlewares: r.middlewares,
		path:        fmt.Sprintf("%s/%s", r.path, path[1:]),
	}
}

func (r *Router) RouteFunc(path string, f func(*Router)) {
	f(r.Route(path))
}
