package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/jav"
	"pornboss/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CollectionSummary is returned by list APIs.
type CollectionSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	SampleCode  string `json:"sample_code"`
}

// CreateCollection inserts a new collection.
func CreateCollection(ctx context.Context, name, description string) (*models.Collection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("collection name cannot be empty")
	}
	description = strings.TrimSpace(description)
	col := models.Collection{Name: name, Description: description}
	if err := common.DB.WithContext(ctx).Create(&col).Error; err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}
	return &col, nil
}

// GetCollection loads a collection by id.
func GetCollection(ctx context.Context, id int64) (*models.Collection, error) {
	if id <= 0 {
		return nil, errors.New("invalid collection id")
	}
	var col models.Collection
	if err := common.DB.WithContext(ctx).First(&col, id).Error; err != nil {
		return nil, fmt.Errorf("get collection: %w", err)
	}
	return &col, nil
}

// UpdateCollection updates name and/or description.
func UpdateCollection(ctx context.Context, id int64, name *string, description *string) (*models.Collection, error) {
	if id <= 0 {
		return nil, errors.New("invalid collection id")
	}
	var col models.Collection
	if err := common.DB.WithContext(ctx).First(&col, id).Error; err != nil {
		return nil, fmt.Errorf("get collection: %w", err)
	}
	updates := map[string]any{}
	if name != nil {
		n := strings.TrimSpace(*name)
		if n == "" {
			return nil, errors.New("collection name cannot be empty")
		}
		updates["name"] = n
	}
	if description != nil {
		updates["description"] = strings.TrimSpace(*description)
	}
	if len(updates) == 0 {
		return &col, nil
	}
	if err := common.DB.WithContext(ctx).Model(&col).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update collection: %w", err)
	}
	if err := common.DB.WithContext(ctx).First(&col, id).Error; err != nil {
		return nil, fmt.Errorf("reload collection: %w", err)
	}
	return &col, nil
}

// DeleteCollection removes a collection and its jav links.
func DeleteCollection(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid collection id")
	}
	if err := common.DB.WithContext(ctx).Delete(&models.Collection{}, id).Error; err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	return nil
}

// ListCollections returns all collections with member counts.
func ListCollections(ctx context.Context) ([]CollectionSummary, error) {
	var rows []CollectionSummary
	err := common.DB.WithContext(ctx).
		Table("collection c").
		Select(`c.id, c.name, c.description, COUNT(jc.jav_id) AS count,
			(SELECT j.code FROM jav_collection jc2
			 JOIN jav j ON j.id = jc2.jav_id
			 WHERE jc2.collection_id = c.id
			 ORDER BY j.code LIMIT 1) AS sample_code`).
		Joins("LEFT JOIN jav_collection jc ON jc.collection_id = c.id").
		Group("c.id, c.name, c.description").
		Order("c.name COLLATE NOCASE").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	return rows, nil
}

type javCollectionRow struct {
	JavID        int64  `gorm:"column:jav_id"`
	CollectionID int64  `gorm:"column:collection_id"`
	Name         string `gorm:"column:name"`
}

// attachJavCollections fills Collections on each Jav item for API responses.
func attachJavCollections(ctx context.Context, items []models.Jav) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		if item.ID > 0 {
			ids = append(ids, item.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var rows []javCollectionRow
	if err := common.DB.WithContext(ctx).
		Table("jav_collection jc").
		Select("jc.jav_id, c.id AS collection_id, c.name").
		Joins("JOIN collection c ON c.id = jc.collection_id").
		Where("jc.jav_id IN ?", ids).
		Order("c.name COLLATE NOCASE").
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("load jav collections: %w", err)
	}

	byJavID := make(map[int64][]models.JavCollectionBrief, len(ids))
	for _, row := range rows {
		byJavID[row.JavID] = append(byJavID[row.JavID], models.JavCollectionBrief{
			ID:   row.CollectionID,
			Name: row.Name,
		})
	}
	for i := range items {
		items[i].Collections = byJavID[items[i].ID]
	}
	return nil
}

// AddJavsToCollection adds jav rows to a collection (idempotent).
func AddJavsToCollection(ctx context.Context, collectionID int64, javIDs []int64) error {
	if collectionID <= 0 || len(javIDs) == 0 {
		return nil
	}
	clean := uniqueInt64s(javIDs)
	if len(clean) == 0 {
		return nil
	}
	var col models.Collection
	if err := common.DB.WithContext(ctx).First(&col, collectionID).Error; err != nil {
		return fmt.Errorf("find collection: %w", err)
	}
	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		rows := make([]models.JavCollection, 0, len(clean))
		for _, jid := range clean {
			rows = append(rows, models.JavCollection{CollectionID: collectionID, JavID: jid, CreatedAt: now})
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return fmt.Errorf("insert jav_collection: %w", err)
		}
		return nil
	})
}

// RemoveJavsFromCollection removes jav rows from a collection.
func RemoveJavsFromCollection(ctx context.Context, collectionID int64, javIDs []int64) error {
	if collectionID <= 0 || len(javIDs) == 0 {
		return nil
	}
	clean := uniqueInt64s(javIDs)
	if len(clean) == 0 {
		return nil
	}
	if err := common.DB.WithContext(ctx).
		Where("collection_id = ? AND jav_id IN ?", collectionID, clean).
		Delete(&models.JavCollection{}).Error; err != nil {
		return fmt.Errorf("remove jav from collection: %w", err)
	}
	return nil
}

// ListJavIDsInCollection returns jav ids belonging to a collection.
func ListJavIDsInCollection(ctx context.Context, collectionID int64) ([]int64, error) {
	if collectionID <= 0 {
		return nil, errors.New("invalid collection id")
	}
	var ids []int64
	if err := common.DB.WithContext(ctx).
		Model(&models.JavCollection{}).
		Where("collection_id = ?", collectionID).
		Pluck("jav_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("list jav ids in collection: %w", err)
	}
	return ids, nil
}

// ListJavsInCollectionForAnalysis loads up to limit Jav rows with tags and idols for AI prompts.
func ListJavsInCollectionForAnalysis(ctx context.Context, collectionID int64, limit int) ([]models.Jav, error) {
	if collectionID <= 0 {
		return nil, errors.New("invalid collection id")
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 300 {
		limit = 300
	}
	isEnglish := jav.CurrentMetadataLanguageIsEnglish()
	var items []models.Jav
	q := common.DB.WithContext(ctx).
		Model(&models.Jav{}).
		Joins("JOIN jav_collection jc ON jc.jav_id = jav.id AND jc.collection_id = ?", collectionID).
		Preload("Idols", "COALESCE(is_english, 0) = ?", isEnglish).
		Order("jav.code").
		Limit(limit)
	if err := q.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list javs for analysis: %w", err)
	}
	if err := attachVisibleJavTags(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}
