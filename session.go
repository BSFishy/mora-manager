package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/model"
)

type contextKey string

const sessionKey contextKey = "session"

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

func WithSession(ctx context.Context, session *model.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func GetSession(ctx context.Context) (*model.Session, bool) {
	session, ok := ctx.Value(sessionKey).(*model.Session)
	return session, ok
}
