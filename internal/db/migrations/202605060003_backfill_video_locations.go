package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060003_backfill_video_locations.go", migrateBackfillVideoLocations, irreversibleMigration)
}

func migrateBackfillVideoLocations(ctx context.Context, tx *sql.Tx) error {
	if err := execStatements(ctx, tx,
		`UPDATE video SET play_count = 0 WHERE play_count IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_location_directory_path ON video_location(directory_id, relative_path)`,
	); err != nil {
		return err
	}

	done, err := configValueEqualsSQL(ctx, tx, videoLocationBackfillMarkerKey, "1")
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	hasLocations, err := rowsExist(ctx, tx, `SELECT 1 FROM video_location LIMIT 1`)
	if err != nil {
		return err
	}
	if !hasLocations {
		if err := execStatements(ctx, tx,
			`INSERT OR IGNORE INTO video_location (
				video_id,
				directory_id,
				relative_path,
				filename,
				modified_at,
				jav_id,
				is_delete,
				created_at,
				updated_at
			)
			SELECT
				id,
				directory_id,
				path,
				filename,
				modified_at,
				jav_id,
				COALESCE(hidden, 0),
				created_at,
				updated_at
			FROM video
			WHERE directory_id > 0 AND COALESCE(path, '') <> ''`,
			`UPDATE video_location
			 SET jav_id = (
				SELECT video.jav_id
				FROM video
				WHERE video.id = video_location.video_id
			 )
			 WHERE jav_id IS NULL
			   AND EXISTS (
				SELECT 1
				FROM video
				WHERE video.id = video_location.video_id
				  AND video.jav_id IS NOT NULL
			   )`,
		); err != nil {
			return err
		}
	}
	if err := backfillVideoLocationFilenamesSQL(ctx, tx); err != nil {
		return err
	}
	return setConfigValueSQL(ctx, tx, videoLocationBackfillMarkerKey, "1")
}

func backfillVideoLocationFilenamesSQL(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, relative_path
		FROM video_location
		WHERE COALESCE(filename, '') = '' AND COALESCE(relative_path, '') <> ''
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type location struct {
		id           int64
		relativePath string
	}
	var locations []location
	for rows.Next() {
		var loc location
		if err := rows.Scan(&loc.id, &loc.relativePath); err != nil {
			return err
		}
		locations = append(locations, loc)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, loc := range locations {
		filename := baseNameFromSlashPath(loc.relativePath)
		if filename == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE video_location SET filename = ? WHERE id = ?`, filename, loc.id); err != nil {
			return err
		}
	}
	return nil
}
