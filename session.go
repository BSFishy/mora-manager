package main

import (
	"fmt"
	"net/http"
)

func (a *App) getSessionCookie(r *http.Request) (*string, error) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		return &cookie.Value, nil
	}

	if err == http.ErrNoCookie {
		return nil, nil
	}

	return nil, fmt.Errorf("getting session cookie: %w", err)
}
