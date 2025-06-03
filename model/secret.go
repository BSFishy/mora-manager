package model

import (
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

func (d *DB) UsersExist() (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS (SELECT 1 FROM users LIMIT 1)").Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (d *DB) GetOrCreateSecret() (string, error) {
	var secret string
	err := d.db.QueryRow("SELECT value FROM kv WHERE key = 'secret'").Scan(&secret)

	if err == sql.ErrNoRows {
		hasLock, err := d.sessionLock("create_secret")
		if err != nil {
			return "", err
		}

		if hasLock {
			secret := generateSecret(16)

			_, err = d.db.Exec("INSERT INTO kv (key, value) VALUES ('secret', $1)", secret)
			if err != nil {
				return "", fmt.Errorf("inserting new secret: %w", err)
			}

			return secret, nil
		} else {
			for {
				err = d.db.QueryRow("SELECT value FROM kv WHERE key = 'secret'").Scan(&secret)
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
