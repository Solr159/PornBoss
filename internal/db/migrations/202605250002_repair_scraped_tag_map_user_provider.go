package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddNamedMigrationContext("202605250002_repair_scraped_tag_map_user_provider.go", repairScrapedTagMapUserProvider, irreversibleMigration)
}

// Scraped tags were incorrectly stored with provider=user after 202605250001.
func repairScrapedTagMapUserProvider(ctx context.Context, tx *sql.Tx) error {
	return execDB(ctx, tx, `UPDATE jav_tag_map
		SET provider = COALESCE(
			(SELECT jtm2.provider
			 FROM jav_tag_map jtm2
			 WHERE jtm2.jav_tag_id = jav_tag_map.jav_tag_id
			   AND jtm2.provider NOT IN (0, ?)
			 LIMIT 1),
			?)
		WHERE provider = ?
		  AND jav_tag_id IN (SELECT id FROM jav_tag WHERE is_user = 0)`,
		providerUser, providerJavBus, providerUser)
}
