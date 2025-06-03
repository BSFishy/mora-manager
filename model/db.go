package model

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

func NewDB() (*DB, error) {
	// TODO: make this configurable for production
	connStr := "user=mora password=mora dbname=mora sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
	}, nil
}
