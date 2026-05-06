package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060001_create_schema.go", migrateCreateSchema, irreversibleMigration)
}

func migrateCreateSchema(ctx context.Context, tx *sql.Tx) error {
	return execStatements(ctx, tx,
		`CREATE TABLE IF NOT EXISTS directory (
			id integer PRIMARY KEY AUTOINCREMENT,
			path text,
			missing numeric,
			is_delete numeric,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS jav (
			id integer PRIMARY KEY AUTOINCREMENT,
			code text,
			title text,
			release_unix integer,
			duration_min integer,
			provider integer NOT NULL DEFAULT 0,
			fetched_at datetime,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS video (
			id integer PRIMARY KEY AUTOINCREMENT,
			directory_id integer NOT NULL,
			path text,
			filename text,
			size integer,
			modified_at datetime,
			fingerprint text,
			duration_sec integer,
			play_count integer NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime,
			jav_id integer,
			hidden numeric,
			CONSTRAINT fk_video_directory FOREIGN KEY (directory_id) REFERENCES directory(id) ON UPDATE CASCADE ON DELETE RESTRICT,
			CONSTRAINT fk_video_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS video_location (
			id integer PRIMARY KEY AUTOINCREMENT,
			video_id integer NOT NULL,
			directory_id integer NOT NULL,
			relative_path text NOT NULL,
			filename text,
			modified_at datetime,
			jav_id integer,
			is_delete numeric,
			created_at datetime,
			updated_at datetime,
			CONSTRAINT fk_video_location_video FOREIGN KEY (video_id) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_video_location_directory FOREIGN KEY (directory_id) REFERENCES directory(id) ON UPDATE CASCADE ON DELETE RESTRICT
		)`,
		`CREATE TABLE IF NOT EXISTS tag (
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS video_tag (
			video_id integer,
			tag_id integer,
			created_at datetime,
			PRIMARY KEY (video_id, tag_id),
			CONSTRAINT fk_video_tag_video FOREIGN KEY (video_id) REFERENCES video(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_video_tag_tag FOREIGN KEY (tag_id) REFERENCES tag(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS config (
			key text PRIMARY KEY,
			value text,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS jav_tag (
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			provider integer NOT NULL DEFAULT 0,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS jav_idol (
			id integer PRIMARY KEY AUTOINCREMENT,
			name text,
			is_english numeric NOT NULL DEFAULT 0,
			roman_name text,
			japanese_name text,
			chinese_name text,
			height_cm integer,
			birth_date datetime,
			bust integer,
			waist integer,
			hips integer,
			cup integer,
			created_at datetime,
			updated_at datetime
		)`,
		`CREATE TABLE IF NOT EXISTS jav_tag_map (
			jav_id integer,
			jav_tag_id integer,
			created_at datetime,
			PRIMARY KEY (jav_id, jav_tag_id),
			CONSTRAINT fk_jav_tag_map_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_tag_map_jav_tag FOREIGN KEY (jav_tag_id) REFERENCES jav_tag(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS jav_idol_map (
			jav_id integer,
			jav_idol_id integer,
			created_at datetime,
			PRIMARY KEY (jav_id, jav_idol_id),
			CONSTRAINT fk_jav_idol_map_jav FOREIGN KEY (jav_id) REFERENCES jav(id) ON UPDATE CASCADE ON DELETE CASCADE,
			CONSTRAINT fk_jav_idol_map_jav_idol FOREIGN KEY (jav_idol_id) REFERENCES jav_idol(id) ON UPDATE CASCADE ON DELETE CASCADE
		)`,
	)
}
