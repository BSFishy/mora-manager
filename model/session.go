package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Session struct {
	Id     string
	UserID *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (d *DB) NewSetupSession(ctx context.Context) (*Session, error) {
	session := Session{}

	err := d.db.QueryRowContext(ctx, "INSERT INTO sessions DEFAULT VALUES RETURNING id, created_at, updated_at").Scan(&session.Id, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (d *DB) NewSessionForUser(ctx context.Context, userId string) (*Session, error) {
	session := Session{
		UserID: &userId,
	}

	err := d.db.QueryRowContext(ctx, "INSERT INTO sessions (user_id) VALUES ($1) RETURNING id, created_at, updated_at", userId).Scan(&session.Id, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (d *DB) ValidateSession(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := d.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM sessions WHERE id = $1 AND deleted_at IS null)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking if session %s exists: %w", id, err)
	}

	return exists, nil
}

func (d *DB) GetSession(ctx context.Context, id string) (*Session, error) {
	session := Session{
		Id: id,
	}

	err := d.db.QueryRowContext(ctx, "SELECT user_id, created_at, updated_at, deleted_at FROM sessions WHERE id = $1 AND deleted_at IS null", id).Scan(&session.UserID, &session.CreatedAt, &session.UpdatedAt, &session.DeletedAt)
	if err == nil {
		return &session, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

func (s *Session) UpdateUserId(ctx context.Context, d *DB, userId string) error {
	_, err := d.db.ExecContext(ctx, "UPDATE sessions SET user_id = $1, updated_at = now() WHERE id = $2", userId, s.Id)
	if err != nil {
		return err
	}

	s.UserID = &userId
	return nil
}

func (s *Session) Delete(ctx context.Context, d *DB) error {
	_, err := d.db.ExecContext(ctx, "UPDATE sessions SET deleted_at = now() WHERE id = $1", s.Id)
	if err != nil {
		return err
	}

	return nil
}
