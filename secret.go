package main

import (
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/router"
	"github.com/BSFishy/mora-manager/templates"
	"github.com/BSFishy/mora-manager/util"
)

func (a *App) secretMiddleware(handler http.Handler) http.Handler {
	return router.ErrorHandle(func(w http.ResponseWriter, r *http.Request) error {
		usersExist, err := a.db.UsersExist()
		if err != nil {
			return fmt.Errorf("checking if users exist: %w", err)
		}

		if usersExist {
			router.Redirect(w, "/login")
			return nil
		}

		sessionId, err := a.getSessionCookie(r)
		if err != nil {
			return fmt.Errorf("getting session cookie: %w", err)
		}

		if sessionId != nil {
			router.Redirect(w, "/setup/user")
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

	session, err := a.db.NewSetupSession()
	if err != nil {
		logger.Error("failed to create new admin session", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: make this secure for non dev?
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.Id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Location", "/setup/user")
}
