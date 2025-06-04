package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BSFishy/mora-manager/util"
)

func (a *App) userMiddleware(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		usersExist, err := a.db.UsersExist()
		if err != nil {
			return fmt.Errorf("checking if users exist: %w", err)
		}

		if usersExist {
			util.Redirect(w, "/login")
			return nil
		}

		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session cookie: %w", err)
		}

		if sessionId == nil {
			util.Redirect(w, "/setup/secret")
			return nil
		}

		session, err := a.db.GetSession(*sessionId)
		if err != nil {
			return fmt.Errorf("getting session: %w", err)
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

			util.Redirect(w, "/setup/secret")
			return nil
		}

		if session.UserID != nil {
			// session is associated with an actual user. send them directly to the
			// dashboard
			util.Redirect(w, "/dashboard")
			return nil
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

			util.Redirect(w, "/setup/secret")
			return nil
		}

		handler.ServeHTTP(w, r)
		return nil
	})
}
