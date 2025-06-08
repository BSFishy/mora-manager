package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Environment struct {
	Id     string
	UserId string
	Name   string
	Slug   string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (d *DB) NewEnvironment(ctx context.Context, userId, name, slug string) (*Environment, error) {
	environment := Environment{
		UserId: userId,
		Name:   name,
		Slug:   slug,
	}

	err := d.db.QueryRowContext(ctx, "INSERT INTO environments (user_id, name, slug) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at", userId, name, slug).Scan(&environment.Id, &environment.CreatedAt, &environment.UpdatedAt)
	if err == nil {
		return &environment, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

func (d *DB) GetUserEnvironments(ctx context.Context, userId string) ([]Environment, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, name, slug, created_at, updated_at FROM environments WHERE user_id = $1 AND deleted_at IS NULL", userId)
	if err != nil {
		return nil, fmt.Errorf("getting environments: %w", err)
	}

	environments := []Environment{}
	for rows.Next() {
		environment := Environment{
			UserId: userId,
		}

		err = rows.Scan(&environment.Id, &environment.Name, &environment.Slug, &environment.CreatedAt, &environment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning environment: %w", err)
		}

		environments = append(environments, environment)
	}

	return environments, nil
}

func (d *DB) GetEnvironment(ctx context.Context, id string) (*Environment, error) {
	environment := Environment{
		Id: id,
	}

	err := d.db.QueryRowContext(ctx, "SELECT user_id, name, slug, created_at, updated_at FROM environments WHERE id = $1 AND deleted_at IS NULL", id).Scan(&environment.UserId, &environment.Name, &environment.Slug, &environment.CreatedAt, &environment.UpdatedAt)
	if err == nil {
		return &environment, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

func (d *DB) GetEnvironmentBySlug(ctx context.Context, userId, slug string) (*Environment, error) {
	environment := Environment{
		UserId: userId,
		Slug:   slug,
	}

	err := d.db.QueryRowContext(ctx, "SELECT id, name, created_at, updated_at FROM environments WHERE user_id = $1 AND slug = $2 AND deleted_at IS NULL", userId, slug).Scan(&environment.Id, &environment.Name, &environment.CreatedAt, &environment.UpdatedAt)
	if err == nil {
		return &environment, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

func (e *Environment) Delete(ctx context.Context, d *DB) error {
	_, err := d.db.ExecContext(ctx, "UPDATE environments SET deleted_at = now() WHERE id = $1", e.Id)
	return err
}
