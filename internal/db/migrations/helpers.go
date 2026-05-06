package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	providerJavBus      = 1
	providerJavDatabase = 2
	providerUser        = 3
)

const videoLocationBackfillMarkerKey = "migration.video_locations.backfilled"
const javIdolEnglishFlagsBackfillMarkerKey = "migration.jav_idol_english_flags.v3.backfilled"

type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func irreversibleMigration(context.Context, *sql.Tx) error {
	return fmt.Errorf("database migrations are irreversible")
}

func execStatements(ctx context.Context, execer sqlExecer, statements ...string) error {
	for _, stmt := range statements {
		if err := execDB(ctx, execer, stmt); err != nil {
			return err
		}
	}
	return nil
}

func execDB(ctx context.Context, execer sqlExecer, stmt string, args ...any) error {
	_, err := execer.ExecContext(ctx, stmt, args...)
	return err
}

func addColumnIfMissing(ctx context.Context, tx *sql.Tx, table, column, decl string) error {
	exists, err := columnExists(ctx, tx, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return execDB(ctx, tx, fmt.Sprintf(
		`ALTER TABLE %s ADD COLUMN %s %s`,
		quoteIdentifier(table),
		quoteIdentifier(column),
		decl,
	))
}

func dropColumnIfExists(ctx context.Context, tx *sql.Tx, table, column string) error {
	exists, err := columnExists(ctx, tx, table, column)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	err = execDB(ctx, tx, fmt.Sprintf(
		`ALTER TABLE %s DROP COLUMN %s`,
		quoteIdentifier(table),
		quoteIdentifier(column),
	))
	if ignorableSQLiteDropColumnErr(err) {
		return nil
	}
	return err
}

func columnExists(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, quoteIdentifier(table)))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			colType    string
			notNull    int
			defaultV   any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &primaryKey); err != nil {
			return false, err
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func rowsExist(ctx context.Context, tx *sql.Tx, query string, args ...any) (bool, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), rows.Err()
}

func configValueEqualsSQL(ctx context.Context, tx *sql.Tx, key, value string) (bool, error) {
	var got string
	err := tx.QueryRowContext(ctx, `SELECT value FROM config WHERE key = ?`, key).Scan(&got)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return got == value, nil
}

func setConfigValueSQL(ctx context.Context, tx *sql.Tx, key, value string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO config (key, value, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}

func baseNameFromSlashPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	return filepath.Base(filepath.FromSlash(p))
}

func ignorableSQLiteDropColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such column") || strings.Contains(msg, "duplicate column name")
}

func looksLikeJapaneseName(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 0x3040 && r <= 0x30ff:
			return true
		case r >= 0x31f0 && r <= 0x31ff:
			return true
		case r >= 0x4e00 && r <= 0x9fff:
			return true
		case r >= 0xff66 && r <= 0xff9d:
			return true
		}
	}
	return false
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
