package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Token struct {
	Id     string
	UserID string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (d *DB) NewToken(ctx context.Context, userId string) (*Token, error) {
	token := Token{
		UserID: userId,
	}

	err := d.db.QueryRowContext(ctx, "INSERT INTO tokens (user_id) VALUES ($1) RETURNING id, created_at, updated_at", userId).Scan(&token.Id, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (d *DB) GetTokens(ctx context.Context, userId string) ([]Token, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, created_at, updated_at FROM tokens WHERE user_id = $1 AND deleted_at IS NULL", userId)
	if err != nil {
		return nil, fmt.Errorf("getting tokens: %w", err)
	}

	defer rows.Close()

	tokens := []Token{}
	for rows.Next() {
		token := Token{
			UserID: userId,
		}

		err = rows.Scan(&token.Id, &token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (d *DB) GetToken(ctx context.Context, id string) (*Token, error) {
	token := Token{
		Id: id,
	}

	err := d.db.QueryRowContext(ctx, "SELECT user_id, created_at, updated_at, deleted_at FROM tokens WHERE id = $1", id).Scan(&token.UserID, &token.CreatedAt, &token.UpdatedAt, &token.DeletedAt)
	if err == nil {
		return &token, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

func (d *DB) GetTokenAndUser(ctx context.Context, id string) (*Token, *User, error) {
	token := Token{
		Id: id,
	}

	user := User{}

	err := d.db.QueryRowContext(ctx, `
		SELECT
			tokens.user_id, tokens.created_at, tokens.updated_at,
			users.id, users.username, users.admin, users.created_at, users.updated_at
		FROM tokens
		INNER JOIN users
			ON tokens.user_id = users.id
		WHERE tokens.id = $1
			AND tokens.deleted_at IS NULL
			AND users.deleted_at IS NULL
		`, id).Scan(
		&token.UserID, &token.CreatedAt, &token.UpdatedAt,
		&user.Id, &user.Username, &user.Admin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == nil {
		return &token, &user, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil, nil
	}

	return nil, nil, err
}

func (t *Token) Delete(ctx context.Context, d *DB) error {
	_, err := d.db.ExecContext(ctx, "UPDATE tokens SET deleted_at = now() WHERE id = $1", t.Id)
	return err
}
