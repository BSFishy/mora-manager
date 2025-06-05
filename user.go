package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BSFishy/mora-manager/model"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
)

const userKey contextKey = "user"

func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(userKey).(*model.User)
	return user, ok
}

func (a *App) userProtected(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session cookie: %w", err)
		}

		if sessionId == nil {
			util.Redirect(w, "/login")
			return nil
		}

		session, err := a.db.GetSession(*sessionId)
		if err != nil {
			return fmt.Errorf("getting session: %w", err)
		}

		if session == nil || session.UserID == nil {
			// invalid session cookie, delete
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				MaxAge:   -1,
				HttpOnly: false,
			})

			util.Redirect(w, "/login")
			return nil
		}

		user, err := a.db.GetUserById(*session.UserID)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}

		r = r.WithContext(WithUser(r.Context(), user))

		handler.ServeHTTP(w, r)
		return nil
	})
}

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
				HttpOnly: false,
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
				HttpOnly: false,
			})

			util.Redirect(w, "/setup/secret")
			return nil
		}

		r = r.WithContext(WithSession(r.Context(), session))

		handler.ServeHTTP(w, r)
		return nil
	})
}

func (a *App) userHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	username := r.Form.Get("username")
	if username == "" {
		return templates.UserForm(true).Render(ctx, w)
	}

	password := r.Form.Get("password")
	if password == "" {
		return templates.UserForm(true).Render(ctx, w)
	}

	user, err := a.db.NewAdminUser(username, password)
	if err != nil {
		return fmt.Errorf("creating new user: %w", err)
	}

	session, _ := GetSession(ctx)
	err = session.UpdateUserId(a.db, user.Id)
	if err != nil {
		return fmt.Errorf("updating session user id: %w", err)
	}

	w.Header().Set("Hx-Location", "/dashboard")
	return nil
}
