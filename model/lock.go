package model

import (
	"context"
	"fmt"
	"hash/fnv"
)

func hashStringToInt64(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64()) // truncate to signed int64
}

func (d *DB) sessionLock(ctx context.Context, name string) (bool, error) {
	lockKey := hashStringToInt64(name)

	var ok bool
	err := d.db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("taking session lock: %w", err)
	}

	return ok, nil
}
