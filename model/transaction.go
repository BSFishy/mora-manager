package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
)

func (d *DB) Transact(ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			util.LogFromCtx(ctx).Error("transaction panicked", "err", r)
		}
	}()

	if err = f(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
