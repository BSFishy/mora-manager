package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type DeploymentStatus string

const (
	NotStarted DeploymentStatus = "not_started"
	InProgress DeploymentStatus = "in_progress"
	Waiting    DeploymentStatus = "waiting"
	Cancelled  DeploymentStatus = "cancelled"
	Errored    DeploymentStatus = "errored"
	Success    DeploymentStatus = "success"
)

type Deployment struct {
	Id            string
	EnvironmentId string
	Status        DeploymentStatus
	Config        json.RawMessage
	State         *json.RawMessage

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *Environment) NewDeployment(ctx context.Context, d *DB, config any) (*Deployment, error) {
	rawConfig, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshalling config: %w", err)
	}

	deployment := Deployment{
		EnvironmentId: e.Id,
		Status:        NotStarted,
		Config:        rawConfig,
	}
	err = d.db.QueryRowContext(ctx, "INSERT INTO deployments (environment_id, status, config) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at", e.Id, NotStarted, rawConfig).Scan(&deployment.Id, &deployment.CreatedAt, &deployment.UpdatedAt)
	if err == nil {
		return &deployment, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

// TODO: will we ever want to like undo a cancelled deployment or something like
// that?
func (e *Environment) CancelInProgressDeployments(ctx context.Context, d *DB) error {
	_, err := d.db.ExecContext(ctx, "UPDATE deployments SET status = $1, updated_at = now() WHERE environment_id = $2 AND status IN ($3, $4, $5)", Cancelled, e.Id, NotStarted, InProgress, Waiting)
	return err
}

func (d *Deployment) Lock(ctx context.Context, db *DB) (*sql.Tx, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", hashStringToInt64(fmt.Sprintf("%s/%s", d.EnvironmentId, d.Id)))
	if err != nil {
		return nil, fmt.Errorf("obtaining lock: %w", err)
	}

	return tx, nil
}

func (d *Deployment) Unlock(tx *sql.Tx) error {
	return tx.Commit()
}

func (d *Deployment) Refresh(ctx context.Context, tx *sql.Tx) error {
	err := tx.QueryRowContext(ctx, "SELECT status, config, state FROM deployments WHERE id = $1", d.Id).Scan(&d.Status, &d.Config, &d.State)
	return err
}

func (d *Deployment) UpdateStatus(ctx context.Context, tx *sql.Tx, status DeploymentStatus) error {
	_, err := tx.ExecContext(ctx, "UPDATE deployments SET status = $1, updated_at = now() WHERE id = $2", status, d.Id)
	return err
}

func (d *Deployment) UpdateState(ctx context.Context, tx *sql.Tx, state map[string]any) error {
	stateBlob, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE deployments SET state = $1, updated_at = now() WHERE id = $2", stateBlob, d.Id)
	if err != nil {
		return fmt.Errorf("updating database: %w", err)
	}

	return nil
}

func (d *Deployment) UpdateStateAndStatus(ctx context.Context, tx *sql.Tx, status DeploymentStatus, state any) error {
	stateBlob, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE deployments SET status = $1, state = $2, updated_at = now() WHERE id = $3", status, stateBlob, d.Id)
	if err != nil {
		return fmt.Errorf("updating database: %w", err)
	}

	return nil
}
