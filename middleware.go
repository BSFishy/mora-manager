package main

import (
	"fmt"
	"net/http"
)

func log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Handling %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
