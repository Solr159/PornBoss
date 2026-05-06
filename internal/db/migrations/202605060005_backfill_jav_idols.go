package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060005_backfill_jav_idols.go", migrateBackfillJavIdols, irreversibleMigration)
}

func migrateBackfillJavIdols(ctx context.Context, tx *sql.Tx) error {
	done, err := configValueEqualsSQL(ctx, tx, javIdolEnglishFlagsBackfillMarkerKey, "1")
	if err != nil {
		return err
	}
	if !done {
		if _, err := tx.ExecContext(ctx, `
			UPDATE jav_idol
			SET is_english = 1
			WHERE COALESCE(is_english, 0) = 0
			  AND id IN (
				SELECT jim.jav_idol_id
				FROM jav_idol_map jim
				JOIN jav j ON j.id = jim.jav_id
				WHERE j.provider = ?
			  )
		`, providerJavDatabase); err != nil {
			return err
		}
		if err := setConfigValueSQL(ctx, tx, javIdolEnglishFlagsBackfillMarkerKey, "1"); err != nil {
			return err
		}
	}
	if err := execStatements(ctx, tx,
		`DROP INDEX IF EXISTS idx_jav_idol_name`,
		`DROP INDEX IF EXISTS uni_jav_idol_name`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_idol_name_language ON jav_idol(name, is_english)`,
	); err != nil {
		return err
	}
	return backfillJavIdolJapaneseNamesSQL(ctx, tx)
}

func backfillJavIdolJapaneseNamesSQL(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name
		FROM jav_idol
		WHERE (japanese_name IS NULL OR japanese_name = '') AND name IS NOT NULL AND name <> ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type idol struct {
		id   int64
		name string
	}
	var idols []idol
	for rows.Next() {
		var item idol
		if err := rows.Scan(&item.id, &item.name); err != nil {
			return err
		}
		idols = append(idols, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range idols {
		if !looksLikeJapaneseName(item.name) {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE jav_idol
			SET japanese_name = name
			WHERE id = ? AND (japanese_name IS NULL OR japanese_name = '')
		`, item.id); err != nil {
			return err
		}
	}
	return nil
}
