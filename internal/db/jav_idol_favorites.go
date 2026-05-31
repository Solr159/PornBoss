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

// JavIdolFavoriteGroupSummary represents an idol favorite group with a visible idol count.
type JavIdolFavoriteGroupSummary struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func buildIdolFavoriteCountQuery(ctx context.Context) *gorm.DB {
	return common.DB.WithContext(ctx).
		Table("jav_idol_favorite_map").
		Select("jav_idol_id, COUNT(DISTINCT jav_idol_favorite_group_id) AS favorite_count").
		Group("jav_idol_id")
}

// ListJavIdolFavoriteGroups returns user-created idol favorite groups.
func ListJavIdolFavoriteGroups(ctx context.Context, directoryIDs []int64) ([]JavIdolFavoriteGroupSummary, error) {
	var groups []JavIdolFavoriteGroupSummary
	isEnglish := jav.CurrentMetadataLanguageIsEnglish()
	soloIdols := buildVisibleSoloIdolSampleQuery(ctx, directoryIDs, isEnglish)
	if err := common.DB.WithContext(ctx).
		Table("jav_idol_favorite_group jifg").
		Select("jifg.id, jifg.name, COUNT(DISTINCT CASE WHEN solo_idols.sample_code IS NOT NULL AND COALESCE(ji.is_english, 0) = ? THEN jifm.jav_idol_id END) AS count", isEnglish).
		Joins("LEFT JOIN jav_idol_favorite_map jifm ON jifm.jav_idol_favorite_group_id = jifg.id").
		Joins("LEFT JOIN jav_idol ji ON ji.id = jifm.jav_idol_id").
		Joins("LEFT JOIN (?) solo_idols ON solo_idols.jav_idol_id = jifm.jav_idol_id", soloIdols).
		Group("jifg.id, jifg.name").
		Order("jifg.name").
		Scan(&groups).Error; err != nil {
		return nil, fmt.Errorf("list jav idol favorite groups: %w", err)
	}
	return groups, nil
}

// CreateJavIdolFavoriteGroup creates a favorite group or returns the existing group with the same name.
func CreateJavIdolFavoriteGroup(ctx context.Context, name string) (*models.JavIdolFavoriteGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("favorite group name cannot be empty")
	}

	var group models.JavIdolFavoriteGroup
	err := common.DB.WithContext(ctx).Where("name = ?", name).First(&group).Error
	if err == nil {
		return &group, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find jav idol favorite group %q: %w", name, err)
	}
	group = models.JavIdolFavoriteGroup{Name: name}
	if err := common.DB.WithContext(ctx).Create(&group).Error; err != nil {
		return nil, fmt.Errorf("create jav idol favorite group %q: %w", name, err)
	}
	return &group, nil
}

// ListJavIdolFavoriteGroupIDs returns favorite group IDs containing the idol.
func ListJavIdolFavoriteGroupIDs(ctx context.Context, idolID int64) ([]int64, error) {
	if idolID <= 0 {
		return nil, errors.New("idol id must be positive")
	}
	var ids []int64
	if err := common.DB.WithContext(ctx).
		Table("jav_idol_favorite_map").
		Where("jav_idol_id = ?", idolID).
		Order("jav_idol_favorite_group_id").
		Pluck("jav_idol_favorite_group_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("list jav idol favorite group ids: %w", err)
	}
	return ids, nil
}

// ReplaceJavIdolFavoriteGroups replaces all favorite group memberships for one idol.
func ReplaceJavIdolFavoriteGroups(ctx context.Context, idolID int64, groupIDs []int64) error {
	if idolID <= 0 {
		return errors.New("idol id must be positive")
	}
	cleanGroupIDs := uniqueInt64s(groupIDs)

	return common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var idolCount int64
		if err := tx.Model(&models.JavIdol{}).Where("id = ?", idolID).Count(&idolCount).Error; err != nil {
			return fmt.Errorf("find jav idol: %w", err)
		}
		if idolCount == 0 {
			return gorm.ErrRecordNotFound
		}

		if len(cleanGroupIDs) > 0 {
			var groupCount int64
			if err := tx.Model(&models.JavIdolFavoriteGroup{}).Where("id IN ?", cleanGroupIDs).Count(&groupCount).Error; err != nil {
				return fmt.Errorf("find jav idol favorite groups: %w", err)
			}
			if groupCount != int64(len(cleanGroupIDs)) {
				return errors.New("invalid favorite_group_id")
			}
		}

		if err := tx.Where("jav_idol_id = ?", idolID).Delete(&models.JavIdolFavoriteMap{}).Error; err != nil {
			return fmt.Errorf("delete jav idol favorite maps: %w", err)
		}
		if len(cleanGroupIDs) == 0 {
			return nil
		}

		now := time.Now()
		rows := make([]models.JavIdolFavoriteMap, 0, len(cleanGroupIDs))
		for _, groupID := range cleanGroupIDs {
			rows = append(rows, models.JavIdolFavoriteMap{
				JavIdolFavoriteGroupID: groupID,
				JavIdolID:              idolID,
				CreatedAt:              now,
			})
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return fmt.Errorf("insert jav idol favorite maps: %w", err)
		}
		return nil
	})
}
