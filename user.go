package main

import (
	"context"
	"fmt"
	"net/http"

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

func (a *App) loginMiddleware(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session id: %w", err)
		}

		if sessionId != nil {
			usersExist, err := a.db.UsersExist()
			if err != nil {
				return fmt.Errorf("getting if users exist: %w", err)
			}

			if !usersExist {
				util.Redirect(w, "/setup/user")
				return nil
			}

			util.Redirect(w, "/dashboard")
			return nil
		}

		handler.ServeHTTP(w, r)
		return nil
	})
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
			// invalid session cookie probably. let's be safe
			DeleteSessionCookie(w)
			util.Redirect(w, "/login")
			return fmt.Errorf("getting session: %w", err)
		}

		if session == nil || session.UserID == nil {
			// invalid session cookie, delete
			DeleteSessionCookie(w)
			util.Redirect(w, "/login")
			return nil
		}

		user, err := a.db.GetUserById(*session.UserID)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}

		r = r.WithContext(WithSession(r.Context(), session))
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
			DeleteSessionCookie(w)
			util.Redirect(w, "/setup/secret")
			return nil
		}

		if session.UserID != nil {
			// session is associated with an actual user. send them directly to the
			// dashboard
			util.Redirect(w, "/dashboard")
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

func (a *App) loginHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	username := r.Form.Get("username")
	if username == "" {
		return templates.LoginForm(true).Render(ctx, w)
	}

	password := r.Form.Get("password")
	if password == "" {
		return templates.LoginForm(true).Render(ctx, w)
	}

	user, err := a.db.GetUserByCredentials(username, password)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	if user == nil {
		return templates.LoginForm(true).Render(ctx, w)
	}

	session, err := a.db.NewSessionForUser(user.Id)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	CreateSessionCookie(w, session.Id)
	w.Header().Set("Hx-Location", "/dashboard")

	return nil
}

func (a *App) signOut(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	session, ok := GetSession(ctx)
	if !ok {
		w.Header().Set("Hx-Location", "/login")
		return nil
	}

	err := session.Delete(a.db)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	DeleteSessionCookie(w)
	w.Header().Set("Hx-Location", "/login")

	return nil
}
