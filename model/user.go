package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       string
	Username string
	Password string
	Admin    bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (d *DB) UsersExist(ctx context.Context) (bool, error) {
	var exists bool
	err := d.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users LIMIT 1)").Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *DB) NewAdminUser(ctx context.Context, username, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := User{
		Username: username,
		Password: string(hashedPassword),
		Admin:    true,
	}

	err = d.db.QueryRowContext(ctx, "INSERT INTO users (username, password, admin) VALUES ($1, $2, true) RETURNING id, created_at, updated_at", username, string(hashedPassword)).Scan(&user.Id, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return &user, nil
}

func (d *DB) GetUserById(ctx context.Context, id string) (*User, error) {
	user := User{
		Id: id,
	}

	err := d.db.QueryRowContext(ctx, "SELECT username, password, admin, created_at, updated_at, deleted_at FROM users WHERE id = $1", id).Scan(&user.Username, &user.Password, &user.Admin, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *DB) GetUserByCredentials(ctx context.Context, username string, password string) (*User, error) {
	user := User{
		Username: username,
	}

	err := d.db.QueryRowContext(ctx, "SELECT id, password, admin, created_at, updated_at, deleted_at FROM users WHERE username = $1 AND deleted_at IS NULL", username).Scan(&user.Id, &user.Password, &user.Admin, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	var deferedErr error
	if err != nil {
		deferedErr = err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if deferedErr != nil {
		if deferedErr == sql.ErrNoRows {
			return nil, nil
		}

		return nil, deferedErr
	}

	if err != nil {
		return nil, nil
	}

	return &user, nil
}

// return nil on success, error on failure
func (u *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
