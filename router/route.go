package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BSFishy/mora-manager/util"
)

func (r *Router) register(method, path string, handler http.HandlerFunc) {
	util.Assert(len(path) >= 1, "path must not be empty")
	util.Assert(strings.HasPrefix(path, "/"), "path must start with /")

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
	util.Assert(len(path) >= 1, "path must not be empty")
	util.Assert(strings.HasPrefix(path, "/"), "path must start with /")

	routePath := fmt.Sprintf("%s/%s", r.path, path[1:])
	route := route{
		patternType: prefix,
		method:      nil,
		pattern:     routePath,
		handler:     r.middleware(handler),
	}

	*r.routes = append(*r.routes, route)
}

func (r *Router) handle(method, path string, handle http.Handler) {
	r.register(method, path, func(w http.ResponseWriter, r *http.Request) {
		handle.ServeHTTP(w, r)
	})
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.register(http.MethodGet, path, handler)
}

func (r *Router) HandleGet(path string, handle http.Handler) {
	r.handle(http.MethodGet, path, handle)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.register(http.MethodPost, path, handler)
}

func (r *Router) Route(path string) *Router {
	util.Assert(len(path) >= 2, "path must not be empty")
	util.Assert(strings.HasPrefix(path, "/"), "path must start with /")

	return &Router{
		routes:      r.routes,
		middlewares: r.middlewares,
		path:        fmt.Sprintf("%s/%s", r.path, path[1:]),
	}
}

func (r *Router) RouteFunc(path string, f func(*Router)) {
	f(r.Route(path))
}
