package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605230001_move_jav_tag_provider_to_map.go", moveJavTagProviderToMap, irreversibleMigration)
}

func moveJavTagProviderToMap(ctx context.Context, tx *sql.Tx) error {
	if err := execStatements(ctx, tx,
		`DROP TABLE IF EXISTS "__new_jav_tag_map"`,
		`DROP TABLE IF EXISTS "__new_jav_tag"`,
		`CREATE TABLE "__new_jav_tag" (
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			is_user numeric NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE "__new_jav_tag_map" (
			jav_id integer,
			jav_tag_id integer,
			provider integer NOT NULL DEFAULT 0,
			created_at datetime,
			PRIMARY KEY (jav_id, jav_tag_id, provider),
			CONSTRAINT fk_jav_tag_map_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_tag_map_jav_tag FOREIGN KEY (jav_tag_id) REFERENCES jav_tag(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
	); err != nil {
		return err
	}

	if err := execDB(ctx, tx, `INSERT INTO "__new_jav_tag" (name, is_user, created_at, updated_at)
		SELECT jt.name,
			CASE WHEN COALESCE(jt.provider, 0) = ? THEN 1 ELSE 0 END AS is_user,
			MIN(jt.created_at),
			MAX(jt.updated_at)
		FROM "jav_tag" jt
		GROUP BY jt.name, CASE WHEN COALESCE(jt.provider, 0) = ? THEN 1 ELSE 0 END`, providerUser, providerUser); err != nil {
		return err
	}
	if err := execDB(ctx, tx, `INSERT OR IGNORE INTO "__new_jav_tag_map" (jav_id, jav_tag_id, provider, created_at)
		SELECT jtm.jav_id,
			new_jt.id,
			COALESCE(jt.provider, 0) AS provider,
			MIN(jtm.created_at)
		FROM "jav_tag_map" jtm
		JOIN "jav_tag" jt ON jt.id = jtm.jav_tag_id
		JOIN "__new_jav_tag" new_jt ON (
			((new_jt.name = jt.name) OR (new_jt.name IS NULL AND jt.name IS NULL))
			AND new_jt.is_user = CASE WHEN COALESCE(jt.provider, 0) = ? THEN 1 ELSE 0 END
		)
		GROUP BY jtm.jav_id, new_jt.id, COALESCE(jt.provider, 0)`, providerUser); err != nil {
		return err
	}

	return execStatements(ctx, tx,
		`DROP TABLE "jav_tag_map"`,
		`DROP TABLE "jav_tag"`,
		`ALTER TABLE "__new_jav_tag" RENAME TO "jav_tag"`,
		`ALTER TABLE "__new_jav_tag_map" RENAME TO "jav_tag_map"`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_tag_name_user ON jav_tag(name, is_user)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_tag_map_provider ON jav_tag_map(provider)`,
	)
}
