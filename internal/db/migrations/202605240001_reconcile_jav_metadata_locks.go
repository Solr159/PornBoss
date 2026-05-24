package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605240001_reconcile_jav_metadata_locks.go", reconcileJavMetadataLocks, irreversibleMigration)
}

var javModelColumnOrder = []string{
	"id",
	"code",
	"title",
	"title_en",
	"studio_id",
	"series_id",
	"series_en_id",
	"release_unix",
	"duration_min",
	"fetched_at",
	"title_locked",
	"title_en_locked",
	"studio_locked",
	"series_locked",
	"series_en_locked",
	"tags_locked",
	"created_at",
	"updated_at",
}

func reconcileJavMetadataLocks(ctx context.Context, tx *sql.Tx) error {
	matches, err := javTableMatchesModel(ctx, tx)
	if err != nil || matches {
		return err
	}
	return rebuildJavTableWithoutProviderAndWithLocks(ctx, tx)
}

func javTableMatchesModel(ctx context.Context, tx *sql.Tx) (bool, error) {
	columns, err := tableColumnNames(ctx, tx, "jav")
	if err != nil {
		return false, err
	}
	if len(columns) != len(javModelColumnOrder) {
		return false, nil
	}
	for i, name := range javModelColumnOrder {
		if !strings.EqualFold(columns[i], name) {
			return false, nil
		}
	}
	return true, nil
}

func tableColumnNames(ctx context.Context, tx *sql.Tx, table string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, quoteIdentifier(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
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
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, rows.Err()
}

func rebuildJavTableWithoutProviderAndWithLocks(ctx context.Context, tx *sql.Tx) error {
	hasTitleLocked, err := columnExists(ctx, tx, "jav", "title_locked")
	if err != nil {
		return err
	}
	lockSelect := "0, 0, 0, 0, 0, 0"
	if hasTitleLocked {
		lockSelect = `COALESCE(title_locked, 0),
			COALESCE(title_en_locked, 0),
			COALESCE(studio_locked, 0),
			COALESCE(series_locked, 0),
			COALESCE(series_en_locked, 0),
			COALESCE(tags_locked, 0)`
	}
	const columns = `"id", "code", "title", "title_en", "studio_id", "series_id", "series_en_id", "release_unix", "duration_min", "fetched_at", "title_locked", "title_en_locked", "studio_locked", "series_locked", "series_en_locked", "tags_locked", "created_at", "updated_at"`
	return execStatements(ctx, tx,
		`DROP TABLE IF EXISTS "__new_jav"`,
		`CREATE TABLE "__new_jav" (
			id integer PRIMARY KEY AUTOINCREMENT,
			code text,
			title text,
			title_en text,
			studio_id integer,
			series_id integer,
			series_en_id integer,
			release_unix integer,
			duration_min integer,
			fetched_at datetime,
			title_locked numeric NOT NULL DEFAULT 0,
			title_en_locked numeric NOT NULL DEFAULT 0,
			studio_locked numeric NOT NULL DEFAULT 0,
			series_locked numeric NOT NULL DEFAULT 0,
			series_en_locked numeric NOT NULL DEFAULT 0,
			tags_locked numeric NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime,
			CONSTRAINT fk_jav_studio FOREIGN KEY (studio_id) REFERENCES jav_studio(id) ON UPDATE CASCADE ON DELETE SET NULL,
			CONSTRAINT fk_jav_series FOREIGN KEY (series_id) REFERENCES jav_series(id) ON UPDATE CASCADE ON DELETE SET NULL,
			CONSTRAINT fk_jav_series_en FOREIGN KEY (series_en_id) REFERENCES jav_series(id) ON UPDATE CASCADE ON DELETE SET NULL
		)`,
		fmt.Sprintf(`INSERT INTO "__new_jav" (`+columns+`)
		 SELECT
			id, code, title, title_en, studio_id, series_id, series_en_id, release_unix, duration_min, fetched_at,
			%s,
			created_at, updated_at
		 FROM "jav"`, lockSelect),
		`DROP TABLE "jav"`,
		`ALTER TABLE "__new_jav" RENAME TO "jav"`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_code ON jav(code)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_studio_id ON jav(studio_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_series_id ON jav(series_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_series_en_id ON jav(series_en_id)`,
	)
}
