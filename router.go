package main

import (
	"context"
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

type route struct {
	method  *string
	pattern string
	handler http.Handler
}

type Router struct {
	routes      *[]route
	middlewares []Middleware
	pathShift   int
}

func NewRouter() Router {
	return Router{
		routes: &[]route{},
	}
}

func match(pattern, path string, shift int) (bool, map[string]string) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	params := map[string]string{}
	for i := 1; i < len(patternParts); i++ {
		pp := patternParts[i]
		if pp == "" && i == len(patternParts)-1 && len(pathParts) == i+shift {
			break
		}

		cp := pathParts[i+shift]

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

		if ok, params := match(route.pattern, req.URL.Path, r.pathShift); ok {
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
		pathShift:   r.pathShift,
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

	route := route{
		method:  &method,
		pattern: path,
		handler: r.middleware(handler),
	}

	*r.routes = append(*r.routes, route)
}

func (r *Router) Handle(method, path string, handle http.Handler) {
	r.Register(method, path, func(w http.ResponseWriter, r *http.Request) {
		handle.ServeHTTP(w, r)
	})
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.Register("GET", path, handler)
}

func (r *Router) HandleGet(path string, handle http.Handler) {
	r.Handle("GET", path, handle)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.Register("POST", path, handler)
}

func (r *Router) Route(path string) *Router {
	Assert(len(path) >= 2, "path must not be empty")
	Assert(strings.HasPrefix(path, "/"), "path must start with /")

	router := &Router{
		routes:      &[]route{},
		middlewares: r.middlewares,
		pathShift:   r.pathShift + 1,
	}

	*r.routes = append(*r.routes, route{
		method:  nil,
		pattern: path,
		handler: router,
	})

	return router
}

func (r *Router) RouteFunc(path string, f func(*Router)) {
	f(r.Route(path))
}
