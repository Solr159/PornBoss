package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605160001_add_collections.go", addCollections, irreversibleMigration)
}

func addCollections(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE TABLE IF NOT EXISTS "collection" (
			id integer PRIMARY KEY AUTOINCREMENT,
			name text NOT NULL,
			description text,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_collection_name ON collection(name)`,
		`CREATE TABLE IF NOT EXISTS "jav_collection" (
			collection_id integer NOT NULL,
			jav_id integer NOT NULL,
			created_at datetime,
			PRIMARY KEY (collection_id, jav_id),
			CONSTRAINT fk_jav_collection_collection FOREIGN KEY (collection_id) REFERENCES collection(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_collection_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_collection_jav_id ON jav_collection(jav_id)`,
	)
}
