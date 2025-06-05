package model

import (
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

func (d *DB) NewAdminUser(username, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := User{
		Username: username,
		Password: string(hashedPassword),
		Admin:    true,
	}

	err = d.db.QueryRow("INSERT INTO users (username, password, admin) VALUES ($1, $2, true) RETURNING id, created_at, updated_at, deleted_at", username, string(hashedPassword)).Scan(&user.Id, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return &user, nil
}

func (d *DB) GetUserById(id string) (*User, error) {
	user := User{
		Id: id,
	}

	err := d.db.QueryRow("SELECT username, password, admin, created_at, updated_at, deleted_at FROM users WHERE id = $1", id).Scan(&user.Username, &user.Password, &user.Admin, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("getting user %s: %w", id, err)
	}

	return &user, nil
}

func (d *DB) GetUserByCredentials(username string, password string) (*User, error) {
	user := User{
		Username: username,
	}

	err := d.db.QueryRow("SELECT id, password, admin, created_at, updated_at, deleted_at FROM users WHERE username = $1 AND deleted_at IS NULL", username).Scan(&user.Id, &user.Password, &user.Admin, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	var deferedErr error
	if err != nil {
		deferedErr = err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if deferedErr != nil {
		return nil, fmt.Errorf("getting user: %w", deferedErr)
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
