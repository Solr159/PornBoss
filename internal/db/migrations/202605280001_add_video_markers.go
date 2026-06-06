package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605280001_add_video_markers.go", addVideoMarkers, irreversibleMigration)
}

func addVideoMarkers(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE TABLE IF NOT EXISTS "video_marker" (
			id integer PRIMARY KEY AUTOINCREMENT,
			video_id integer NOT NULL,
			time_sec real NOT NULL,
			note text NOT NULL,
			created_at datetime,
			updated_at datetime,
			CONSTRAINT fk_video_marker_video FOREIGN KEY (video_id) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_video_marker_video_id ON video_marker(video_id)`,
	)
}
