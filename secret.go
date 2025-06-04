package main

import (
	"net/http"

	"github.com/BSFishy/mora-manager/templates"
)

func (a *App) secretMiddleware(handler http.Handler) http.Handler {
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

		if sessionId != nil {
			w.Header().Set("location", "/setup/user")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (a *App) secretHtmxRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := LogFromCtx(ctx)

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
