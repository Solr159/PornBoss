package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060002_ensure_columns.go", migrateEnsureColumns, irreversibleMigration)
}

func migrateEnsureColumns(ctx context.Context, tx *sql.Tx) error {
	columns := []struct {
		table  string
		column string
		decl   string
	}{
		{"directory", "path", "text"},
		{"directory", "missing", "numeric"},
		{"directory", "is_delete", "numeric"},
		{"directory", "created_at", "datetime"},
		{"directory", "updated_at", "datetime"},
		{"jav", "code", "text"},
		{"jav", "title", "text"},
		{"jav", "release_unix", "integer"},
		{"jav", "duration_min", "integer"},
		{"jav", "provider", "integer NOT NULL DEFAULT 0"},
		{"jav", "fetched_at", "datetime"},
		{"jav", "created_at", "datetime"},
		{"jav", "updated_at", "datetime"},
		{"video", "directory_id", "integer NOT NULL DEFAULT 0"},
		{"video", "path", "text"},
		{"video", "filename", "text"},
		{"video", "size", "integer"},
		{"video", "modified_at", "datetime"},
		{"video", "fingerprint", "text"},
		{"video", "duration_sec", "integer"},
		{"video", "play_count", "integer NOT NULL DEFAULT 0"},
		{"video", "created_at", "datetime"},
		{"video", "updated_at", "datetime"},
		{"video", "jav_id", "integer"},
		{"video", "hidden", "numeric"},
		{"video_location", "video_id", "integer NOT NULL DEFAULT 0"},
		{"video_location", "directory_id", "integer NOT NULL DEFAULT 0"},
		{"video_location", "relative_path", "text NOT NULL DEFAULT ''"},
		{"video_location", "filename", "text"},
		{"video_location", "modified_at", "datetime"},
		{"video_location", "jav_id", "integer"},
		{"video_location", "is_delete", "numeric"},
		{"video_location", "created_at", "datetime"},
		{"video_location", "updated_at", "datetime"},
		{"tag", "name", "text"},
		{"tag", "created_at", "datetime"},
		{"tag", "updated_at", "datetime"},
		{"video_tag", "created_at", "datetime"},
		{"config", "value", "text"},
		{"config", "created_at", "datetime"},
		{"config", "updated_at", "datetime"},
		{"jav_tag", "name", "text"},
		{"jav_tag", "provider", "integer NOT NULL DEFAULT 0"},
		{"jav_tag", "created_at", "datetime"},
		{"jav_tag", "updated_at", "datetime"},
		{"jav_idol", "name", "text"},
		{"jav_idol", "is_english", "numeric NOT NULL DEFAULT 0"},
		{"jav_idol", "roman_name", "text"},
		{"jav_idol", "japanese_name", "text"},
		{"jav_idol", "chinese_name", "text"},
		{"jav_idol", "height_cm", "integer"},
		{"jav_idol", "birth_date", "datetime"},
		{"jav_idol", "bust", "integer"},
		{"jav_idol", "waist", "integer"},
		{"jav_idol", "hips", "integer"},
		{"jav_idol", "cup", "integer"},
		{"jav_idol", "created_at", "datetime"},
		{"jav_idol", "updated_at", "datetime"},
		{"jav_tag_map", "created_at", "datetime"},
		{"jav_idol_map", "created_at", "datetime"},
	}
	for _, col := range columns {
		if err := addColumnIfMissing(ctx, tx, col.table, col.column, col.decl); err != nil {
			return err
		}
	}
	return nil
}
