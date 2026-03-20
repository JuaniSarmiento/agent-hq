package db

import (
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

// InitSchema creates all tables and indexes if they don't exist.
func (db *DB) InitSchema() error {
	if _, err := db.conn.Exec(schemaSQL); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}
