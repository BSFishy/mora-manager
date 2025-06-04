package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Session struct {
	Id     string
	UserID *string
	Admin  bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (d *DB) NewSetupSession() (*Session, error) {
	session := Session{
		Admin: true,
	}

	err := d.db.QueryRow("INSERT INTO sessions (admin) VALUES (true) RETURNING id, created_at, updated_at").Scan(&session.Id, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting session into database: %w", err)
	}

	return &session, nil
}

func (d *DB) ValidateSession(id string) (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS (SELECT 1 FROM sessions WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking if session %s exists: %w", id, err)
	}

	return exists, nil
}

func (d *DB) GetSession(id string) (*Session, error) {
	session := Session{
		Id: id,
	}

	err := d.db.QueryRow("SELECT user_id, admin, created_at, updated_at, deleted_at FROM sessions WHERE id = $1", id).Scan(&session.UserID, &session.Admin, &session.CreatedAt, &session.UpdatedAt, &session.DeletedAt)
	if err == nil {
		return &session, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, fmt.Errorf("getting session %s: %w", id, err)
}
