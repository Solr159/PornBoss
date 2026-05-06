package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605060004_backfill_jav_providers.go", migrateBackfillJavProviders, irreversibleMigration)
}

func migrateBackfillJavProviders(ctx context.Context, tx *sql.Tx) error {
	hasIsUser, err := columnExists(ctx, tx, "jav_tag", "is_user")
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jav SET provider = ? WHERE COALESCE(provider, 0) = 0`, providerJavBus); err != nil {
		return err
	}

	if hasIsUser {
		if _, err := tx.ExecContext(ctx,
			`UPDATE jav_tag
			 SET provider = CASE WHEN COALESCE(is_user, 0) = 1 THEN ? ELSE ? END
			 WHERE COALESCE(provider, 0) = 0`,
			providerUser,
			providerJavBus,
		); err != nil {
			return err
		}
	} else if _, err := tx.ExecContext(ctx, `UPDATE jav_tag SET provider = ? WHERE COALESCE(provider, 0) = 0`, providerJavBus); err != nil {
		return err
	}

	if err := execStatements(ctx, tx,
		`DROP INDEX IF EXISTS idx_jav_tag_name_source`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_jav_tag_name_source ON jav_tag(name, provider)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_tag_provider ON jav_tag(provider)`,
	); err != nil {
		return err
	}
	if hasIsUser {
		return dropColumnIfExists(ctx, tx, "jav_tag", "is_user")
	}
	return nil
}
