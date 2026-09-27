package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// MigrateUp applies the media schema for dialect (mysql|postgres|sqlite3).
// Empty tableName defaults to DefaultTableName.
func MigrateUp(ctx context.Context, db *sql.DB, dialect, idStrategy, tableName string) error {
	if idStrategy == "" {
		idStrategy = IDStrategyFromEnv()
	}
	q, err := CreateMediaSQL(dialect, idStrategy, tableName)
	if err != nil {
		return err
	}
	for _, stmt := range splitSQL(q) {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
	}
	return nil
}

// MigrateDown drops the media table. Empty tableName defaults to DefaultTableName.
func MigrateDown(ctx context.Context, db *sql.DB, tableName string) error {
	table, err := NormalizeTableName(tableName)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, table))
	return err
}

func splitSQL(q string) []string {
	parts := strings.Split(q, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}
