package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware

	// root path will be empty, so handler paths will start with / no matter what
	path string
}

func NewRouter() Router {
	return Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) ListenAndServe(addr string) error {
	fmt.Printf("Listening on %s\n", addr)

	return http.ListenAndServe(addr, r.mux)
}

func (r *Router) Use(m ...Middleware) *Router {
	return &Router{
		mux:         r.mux,
		middlewares: append(r.middlewares, m...),
		path:        r.path,
	}
}

func (r *Router) Register(method string, path string, handler http.HandlerFunc) {
	Assert(len(path) >= 2, "path must not be empty")
	Assert(strings.HasPrefix(path, "/"), "path must start with /")

	wrapped := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		wrapped = r.middlewares[i](wrapped).(http.HandlerFunc)
	}

	handlerPath := fmt.Sprintf("%s/%s", r.path, path[1:])
	r.mux.HandleFunc(fmt.Sprintf("%s %s", method, handlerPath), wrapped)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.Register("GET", path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.Register("POST", path, handler)
}

func (r *Router) Route(path string) *Router {
	Assert(len(path) >= 2, "path must not be empty")
	Assert(strings.HasPrefix(path, "/"), "path must start with /")

	return &Router{
		mux:         r.mux,
		middlewares: r.middlewares,
		path:        fmt.Sprintf("%s/%s", r.path, path[1:]),
	}
}

func (r *Router) RouteFunc(path string, f func(*Router)) {
	f(r.Route(path))
}
