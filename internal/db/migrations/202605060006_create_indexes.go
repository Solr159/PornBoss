package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060006_create_indexes.go", migrateCreateIndexes, irreversibleMigration)
}

func migrateCreateIndexes(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_directory_path ON directory(path)`,
		`CREATE INDEX IF NOT EXISTS idx_directory_missing ON directory(missing)`,
		`CREATE INDEX IF NOT EXISTS idx_directory_is_delete ON directory(is_delete)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_code ON jav(code)`,
		`CREATE INDEX IF NOT EXISTS idx_video_directory_id ON video(directory_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_path ON video(path)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_fingerprint ON video(fingerprint)`,
		`DROP INDEX IF EXISTS idx_video_jav_id_visible`,
		`CREATE INDEX IF NOT EXISTS idx_video_jav_id ON video(jav_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_hidden ON video(hidden)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_video_id ON video_location(video_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_directory_id ON video_location(directory_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_video_location_directory_path ON video_location(directory_id, relative_path)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_filename ON video_location(filename)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_jav_id ON video_location(jav_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_is_delete ON video_location(is_delete)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_jav_id_is_delete ON video_location(jav_id, is_delete)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_video_id_jav_id ON video_location(video_id, jav_id)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_visible_path ON video_location(jav_id, is_delete, relative_path)`,
		`CREATE INDEX IF NOT EXISTS idx_video_location_visible_filename ON video_location(jav_id, is_delete, filename COLLATE NOCASE)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_name ON tag(name)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_tag_name_source ON jav_tag(name, provider)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_tag_provider ON jav_tag(provider)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_idol_name_language ON jav_idol(name, is_english)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_idol_map_jav_idol_id_jav_id ON jav_idol_map(jav_idol_id, jav_id)`,
	)
}
