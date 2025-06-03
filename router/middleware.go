package router

import "net/http"

type Middleware func(http.Handler) http.Handler

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
