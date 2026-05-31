package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605310001_add_jav_idol_favorite_order.go", addJavIdolFavoriteOrder, irreversibleMigration)
}

func addJavIdolFavoriteOrder(ctx context.Context, tx *sql.Tx) error {
	if err := addColumnIfMissing(ctx, tx, "jav_idol_favorite_group", "sort_order", "integer NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := addColumnIfMissing(ctx, tx, "jav_idol_favorite_map", "sort_order", "integer NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	return execStatements(ctx, tx,
		`UPDATE "jav_idol_favorite_group" SET sort_order = id WHERE COALESCE(sort_order, 0) = 0`,
		`UPDATE "jav_idol_favorite_map" SET sort_order = jav_idol_id WHERE COALESCE(sort_order, 0) = 0`,
		`CREATE INDEX IF NOT EXISTS idx_jav_idol_favorite_group_sort_order ON jav_idol_favorite_group(sort_order)`,
		`CREATE INDEX IF NOT EXISTS idx_jav_idol_favorite_map_sort_order ON jav_idol_favorite_map(sort_order)`,
	)
}
