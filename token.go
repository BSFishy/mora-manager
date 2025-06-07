package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/BSFishy/mora-manager/templates"
)

func (a *App) getTokenIds(ctx context.Context, userId string) ([]string, error) {
	tokens, err := a.db.GetTokens(ctx, userId)
	if err != nil {
		return nil, err
	}

	tokenIds := make([]string, len(tokens))
	for i, token := range tokens {
		tokenIds[i] = token.Id
	}

	return tokenIds, nil
}

func (a *App) tokenHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := GetUser(ctx)

	_, err := a.db.NewToken(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("creating token: %w", err)
	}

	tokens, err := a.getTokenIds(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("getting tokens: %w", err)
	}

	return templates.TokenList(tokens).Render(ctx, w)
}

func (a *App) revokeTokenHtmxRoute(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user, _ := GetUser(ctx)

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	id := r.Form.Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	token, err := a.db.GetToken(ctx, id)
	if err != nil {
		return err
	}

	if token == nil {
		// return bad request here to prevent brute forcing tokens
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	if token.UserID != user.Id {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	if err := token.Delete(ctx, a.db); err != nil {
		return fmt.Errorf("deleting token: %w", err)
	}

	tokens, err := a.getTokenIds(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("getting tokens: %w", err)
	}

	return templates.TokenList(tokens).Render(ctx, w)
}
