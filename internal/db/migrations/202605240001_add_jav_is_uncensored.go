package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605240001_add_jav_is_uncensored.go", addJavIsUncensored, irreversibleMigration)
}

func addJavIsUncensored(ctx context.Context, tx *sql.Tx) error {
	return addColumnIfMissing(ctx, tx, "jav", "is_uncensored", "numeric")
}
