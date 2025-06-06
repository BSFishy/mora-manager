package main

import (
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
)

func (a *App) secretMiddleware(handler http.Handler) http.Handler {
	return util.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		usersExist, err := a.db.UsersExist(ctx)
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

		if sessionId != nil {
			util.Redirect(w, "/setup/user")
			return nil
		}

		handler.ServeHTTP(w, r)
		return nil
	})
}

func (a *App) secretHtmxRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.LogFromCtx(ctx)

	err := r.ParseForm()
	if err != nil {
		logger.Error("failed to parse secret form", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !r.Form.Has("secret") {
		templates.SecretForm(true).Render(ctx, w)
		return
	}

	secret := r.Form.Get("secret")
	if secret != a.secret {
		templates.SecretForm(true).Render(ctx, w)
		return
	}

	session, err := a.db.NewSetupSession(ctx)
	if err != nil {
		logger.Error("failed to create new admin session", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	CreateSessionCookie(w, session.Id)

	w.Header().Set("HX-Location", "/setup/user")
}
