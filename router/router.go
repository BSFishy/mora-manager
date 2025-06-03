package router

import (
	"log/slog"
	"net/http"
	"strings"
)

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
