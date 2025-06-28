package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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

func (d *DB) GetDeployments(ctx context.Context, environments []Environment) ([]Deployment, error) {
	if len(environments) < 1 {
		return []Deployment{}, nil
	}

	environmentIds := make([]string, len(environments))
	for i, env := range environments {
		environmentIds[i] = fmt.Sprintf("'%s'", env.Id)
	}

	query := fmt.Sprintf("SELECT id, environment_id, status, created_at, updated_at FROM deployments WHERE environment_id IN (%s) ORDER BY created_at DESC", strings.Join(environmentIds, ", "))
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("selecting rows: %w", err)
	}

	defer rows.Close()

	result := []Deployment{}
	for rows.Next() {
		deployment := Deployment{}
		err := rows.Scan(&deployment.Id, &deployment.EnvironmentId, &deployment.Status, &deployment.CreatedAt, &deployment.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		result = append(result, deployment)
	}

	return result, nil
}

func (d *DB) GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	deployment := Deployment{
		Id: id,
	}

	err := d.db.QueryRowContext(ctx, "SELECT environment_id, status, config, state, created_at, updated_at FROM deployments WHERE id = $1", id).Scan(&deployment.EnvironmentId, &deployment.Status, &deployment.Config, &deployment.State, &deployment.CreatedAt, &deployment.UpdatedAt)
	if err == nil {
		return &deployment, nil
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return nil, err
}

func (d *Deployment) Lock(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", hashStringToInt64(fmt.Sprintf("%s/%s", d.EnvironmentId, d.Id)))
	if err != nil {
		return fmt.Errorf("obtaining lock: %w", err)
	}

	return nil
}

func (d *Deployment) Refresh(ctx context.Context, tx *sql.Tx) error {
	err := tx.QueryRowContext(ctx, "SELECT status, config, state FROM deployments WHERE id = $1", d.Id).Scan(&d.Status, &d.Config, &d.State)
	return err
}

func (d *Deployment) UpdateStatus(ctx context.Context, tx *sql.Tx, status DeploymentStatus) error {
	_, err := tx.ExecContext(ctx, "UPDATE deployments SET status = $1, updated_at = now() WHERE id = $2", status, d.Id)
	return err
}

// need to be careful about where this is called. it SHOULD primarily be in a
// transaction or to update error status.
func (d *Deployment) UpdateStatusDb(ctx context.Context, db *DB, status DeploymentStatus) error {
	_, err := db.db.ExecContext(ctx, "UPDATE deployments SET status = $1, updated_at = now() WHERE id = $2", status, d.Id)
	return err
}

func (d *Deployment) UpdateState(ctx context.Context, tx *sql.Tx, state any) error {
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

func (d *Deployment) UpdateConfig(ctx context.Context, tx *sql.Tx, config any) error {
	configBlob, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE deployments SET config = $1, updated_at = now() WHERE id = $2", configBlob, d.Id)
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
