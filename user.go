package main

import (
	"net/http"
	"time"
)

func (a *App) userMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := LogFromCtx(ctx)

		usersExist, err := a.db.UsersExist()
		if err != nil {
			logger.Error("failed to check if users exist", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if usersExist {
			w.Header().Set("location", "/login")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			logger.Error("failed to get session cookie", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if sessionId == nil {
			w.Header().Set("location", "/setup/secret")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		session, err := a.db.GetSession(*sessionId)
		if err != nil {
			logger.Error("failed to get session", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if session == nil {
			// invalid session cookie, delete
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})

			w.Header().Set("location", "/setup/secret")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		if session.UserID != nil {
			// session is associated with an actual user. send them directly to the
			// dashboard
			w.Header().Set("location", "/dashboard")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		if !session.Admin {
			// not an admin, restart session
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: true,
			})

			w.Header().Set("location", "/setup/secret")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
