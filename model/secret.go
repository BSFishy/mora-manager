package model

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"
)

func generateSecret(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func (d *DB) GetOrCreateSecret(ctx context.Context) (string, error) {
	var secret string
	err := d.db.QueryRowContext(ctx, "SELECT value FROM kv WHERE key = 'secret'").Scan(&secret)

	if err == sql.ErrNoRows {
		hasLock, err := d.sessionLock(ctx, "create_secret")
		if err != nil {
			return "", err
		}

		if hasLock {
			secret := generateSecret(16)

			_, err = d.db.ExecContext(ctx, "INSERT INTO kv (key, value) VALUES ('secret', $1)", secret)
			if err != nil {
				return "", fmt.Errorf("inserting new secret: %w", err)
			}

			return secret, nil
		} else {
			for {
				err = d.db.QueryRowContext(ctx, "SELECT value FROM kv WHERE key = 'secret'").Scan(&secret)
				if err == nil {
					return secret, nil
				}

				if err != sql.ErrNoRows {
					return "", fmt.Errorf("polling for secret: %w", err)
				}

				time.Sleep(time.Millisecond * 500)
			}
		}
	} else if err != nil {
		return "", fmt.Errorf("getting secret: %w", err)
	}

	return secret, nil
}
