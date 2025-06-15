package model

import (
	"database/sql"
	"fmt"

	"github.com/BSFishy/mora-manager/util"
	_ "github.com/lib/pq"
)

var (
	HOST = util.GetenvDefault("MORA_DB_HOST", "[::1]:5432")
	USER = util.GetenvDefault("MORA_DB_USER", "mora")
	NAME = util.GetenvDefault("MORA_DB_NAME", "mora")
	PASS = util.GetenvDefault("MORA_DB_PASS", "mora")
)

type DB struct {
	db *sql.DB
}

func NewDB() (*DB, error) {
	// TODO: disable ssl mode?
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", USER, PASS, HOST, NAME)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
	}, nil
}
