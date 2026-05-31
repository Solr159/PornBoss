package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605300001_add_jav_idol_favorites.go", addJavIdolFavorites, irreversibleMigration)
}

func addJavIdolFavorites(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE TABLE IF NOT EXISTS "jav_idol_favorite_group" (
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS "jav_idol_favorite_map" (
			jav_idol_favorite_group_id integer,
			jav_idol_id integer,
			created_at datetime,
			PRIMARY KEY (jav_idol_favorite_group_id, jav_idol_id),
			CONSTRAINT fk_jav_idol_favorite_map_jav_idol_favorite_group FOREIGN KEY (jav_idol_favorite_group_id) REFERENCES jav_idol_favorite_group(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_idol_favorite_map_jav_idol FOREIGN KEY (jav_idol_id) REFERENCES jav_idol(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_idol_favorite_group_name ON jav_idol_favorite_group(name)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_idol_favorite_map_jav_idol_id_group_id ON jav_idol_favorite_map(jav_idol_id, jav_idol_favorite_group_id)`,
	)
}
