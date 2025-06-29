package model

import (
	"context"
	"fmt"
)

var migrations = map[string]string{
	"000-initial": `CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE kv (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE TABLE users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		admin BOOLEAN NOT NULL DEFAULT false,

		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at TIMESTAMPTZ
	);

	CREATE TABLE tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,

		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at TIMESTAMPTZ,

	  FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID,

		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at TIMESTAMPTZ,

		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE environments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,

		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		deleted_at TIMESTAMPTZ,

		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE deployments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		environment_id UUID NOT NULL,
		previous_deployment_id UUID,
		status TEXT NOT NULL,
		config JSONB NOT NULL,
		state JSONB,

		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

		FOREIGN KEY (environment_id) REFERENCES environments(id),
		FOREIGN KEY (previous_deployment_id) REFERENCES deployments(id)
	);`,
}

func (d *DB) SetupMigrations(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);`)
	if err != nil {
		return fmt.Errorf("creating the migrations table: %w", err)
	}

	dbMigrations, err := d.getMigrations(ctx)
	if err != nil {
		return fmt.Errorf("getting migrations: %w", err)
	}

	for version, script := range migrations {
		if !includesMigration(version, dbMigrations) {
			_, err = d.db.ExecContext(ctx, script)
			if err != nil {
				return fmt.Errorf("running migration %s: %w", version, err)
			}

			_, err := d.db.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
			if err != nil {
				return fmt.Errorf("inserting version %s into migrations table: %w", version, err)
			}
		}
	}

	return nil
}

type migration struct {
	version string
}

// we could probably do this better but honestly this probably wont be too big
func includesMigration(version string, migrations []migration) bool {
	for _, migration := range migrations {
		if migration.version == version {
			return true
		}
	}

	return false
}

func (d *DB) getMigrations(ctx context.Context) ([]migration, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("selecting versions: %w", err)
	}

	defer rows.Close()

	migrations := []migration{}
	for rows.Next() {
		var version string
		if err = rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		migrations = append(migrations, migration{
			version: version,
		})
	}

	return migrations, nil
}
