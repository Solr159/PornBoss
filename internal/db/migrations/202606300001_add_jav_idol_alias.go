package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202606300001_add_jav_idol_alias.go", addJavIdolAlias, irreversibleMigration)
}

func addJavIdolAlias(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE TABLE IF NOT EXISTS "jav_idol_alias" (
			id integer PRIMARY KEY AUTOINCREMENT,
			jav_idol_id integer NOT NULL,
			alias text NOT NULL,
			is_english numeric NOT NULL DEFAULT 0,
			created_at datetime,
			CONSTRAINT fk_jav_idol_alias_jav_idol FOREIGN KEY (jav_idol_id) REFERENCES jav_idol(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_idol_alias_jav_idol_id ON jav_idol_alias(jav_idol_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_idol_alias_alias_language ON jav_idol_alias(alias, is_english)`,
	)
}
