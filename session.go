package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

func CreateSessionCookie(w http.ResponseWriter, sessionId string) {
	// TODO: make this secure for non dev?
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})
}

func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})
}

func WithSession(ctx context.Context, session *model.Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func GetSession(ctx context.Context) (*model.Session, bool) {
	session, ok := ctx.Value(sessionKey).(*model.Session)
	return session, ok
}
