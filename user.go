package main

import (
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
)

func (a *App) loginMiddleware(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session id: %w", err)
		}

		if sessionId != nil {
			usersExist, err := a.db.UsersExist(ctx)
			if err != nil {
				return fmt.Errorf("getting if users exist: %w", err)
			}

			if !usersExist {
				http.Redirect(w, r, "/setup/user", http.StatusFound)
				return nil
			}

			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return nil
		}

		usersExist, err := a.db.UsersExist(ctx)
		if err != nil {
			return fmt.Errorf("getting if users exist: %w", err)
		}

		if !usersExist {
			http.Redirect(w, r, "/setup/secret", http.StatusFound)
			return nil
		}

		handler.ServeHTTP(w, r)
		return nil
	})
}

func (a *App) userProtected(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session cookie: %w", err)
		}

		if sessionId == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return nil
		}

		session, err := a.db.GetSession(ctx, *sessionId)
		if err != nil {
			// invalid session cookie probably. let's be safe
			DeleteSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return fmt.Errorf("getting session: %w", err)
		}

		if session == nil || session.UserID == nil {
			// invalid session cookie, delete
			DeleteSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return nil
		}

		user, err := a.db.GetUserById(ctx, *session.UserID)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}

		r = r.WithContext(WithUser(WithSession(ctx, session), user))

		handler.ServeHTTP(w, r)
		return nil
	})
}

func (a *App) userMiddleware(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		usersExist, err := a.db.UsersExist(ctx)
		if err != nil {
			return fmt.Errorf("checking if users exist: %w", err)
		}

		if usersExist {
			http.Redirect(w, r, "/login", http.StatusFound)
			return nil
		}

		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session cookie: %w", err)
		}

		if sessionId == nil {
			http.Redirect(w, r, "/setup/secret", http.StatusFound)
			return nil
		}

		session, err := a.db.GetSession(ctx, *sessionId)
		if err != nil {
			return fmt.Errorf("getting session: %w", err)
		}

		if session == nil {
			// invalid session cookie, delete
			DeleteSessionCookie(w)
			http.Redirect(w, r, "/setup/secret", http.StatusFound)
			return nil
		}

		if session.UserID != nil {
			// session is associated with an actual user. send them directly to the
			// dashboard
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return nil
		}

		r = r.WithContext(WithSession(ctx, session))

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

	user, err := a.db.NewAdminUser(ctx, username, password)
	if err != nil {
		return fmt.Errorf("creating new user: %w", err)
	}

	session, _ := GetSession(ctx)
	err = session.UpdateUserId(ctx, a.db, user.Id)
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

	user, err := a.db.GetUserByCredentials(ctx, username, password)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	if user == nil {
		return templates.LoginForm(true).Render(ctx, w)
	}

	session, err := a.db.NewSessionForUser(ctx, user.Id)
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

	err := session.Delete(ctx, a.db)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	DeleteSessionCookie(w)
	w.Header().Set("Hx-Location", "/login")

	return nil
}
