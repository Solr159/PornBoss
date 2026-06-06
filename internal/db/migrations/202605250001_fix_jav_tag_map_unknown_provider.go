package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605250001_fix_jav_tag_map_unknown_provider.go", fixJavTagMapUnknownProvider, irreversibleMigration)
}

// User-edited tag maps were saved with provider=0 (unknown) and filtered out on read.
func fixJavTagMapUnknownProvider(ctx context.Context, tx *sql.Tx) error {
	if err := execDB(ctx, tx, `DELETE FROM jav_tag_map
		WHERE provider = 0
		  AND EXISTS (
		    SELECT 1 FROM jav_tag_map j2
		    WHERE j2.jav_id = jav_tag_map.jav_id
		      AND j2.jav_tag_id = jav_tag_map.jav_tag_id
		      AND j2.provider = ?)`, providerUser); err != nil {
		return err
	}
	return execDB(ctx, tx, `UPDATE jav_tag_map SET provider = ? WHERE provider = 0`, providerUser)
}
