package db

import (
	"context"
	"fmt"
	"strings"

	"pornboss/internal/common"
	"pornboss/internal/jav"
	"pornboss/internal/models"
)

// ResolveJavTagNamesToIDs maps tag names (fuzzy) to tag ids visible in metadata providers.
func ResolveJavTagNamesToIDs(ctx context.Context, names []string) (ids []int64, unmatched []string) {
	prov := visibleJavTagProviders()
	for _, raw := range names {
		n := strings.TrimSpace(raw)
		if n == "" {
			continue
		}
		var id int64
		err := common.DB.WithContext(ctx).
			Model(&models.JavTag{}).
			Select("id").
			Where("name = ? COLLATE NOCASE AND provider IN ?", n, prov).
			Limit(1).
			Scan(&id).Error
		if err == nil && id > 0 {
			ids = append(ids, id)
			continue
		}
		err = common.DB.WithContext(ctx).
			Model(&models.JavTag{}).
			Select("id").
			Where("name LIKE ? COLLATE NOCASE AND provider IN ?", "%"+n+"%", prov).
			Order("LENGTH(name) ASC").
			Limit(1).
			Scan(&id).Error
		if err == nil && id > 0 {
			ids = append(ids, id)
			continue
		}
		unmatched = append(unmatched, n)
	}
	return uniqueInt64s(ids), unmatched
}

// ResolveJavIdolNamesToIDs maps idol names to ids for the current metadata language.
func ResolveJavIdolNamesToIDs(ctx context.Context, names []string) (ids []int64, unmatched []string) {
	isEnglish := jav.CurrentMetadataLanguageIsEnglish()
	for _, raw := range names {
		n := strings.TrimSpace(raw)
		if n == "" {
			continue
		}
		var id int64
		err := common.DB.WithContext(ctx).
			Model(&models.JavIdol{}).
			Select("id").
			Where("name = ? COLLATE NOCASE AND COALESCE(is_english, 0) = ?", n, isEnglish).
			Limit(1).
			Scan(&id).Error
		if err == nil && id > 0 {
			ids = append(ids, id)
			continue
		}
		err = common.DB.WithContext(ctx).
			Model(&models.JavIdol{}).
			Select("id").
			Where("name LIKE ? COLLATE NOCASE AND COALESCE(is_english, 0) = ?", "%"+n+"%", isEnglish).
			Order("LENGTH(name) ASC").
			Limit(1).
			Scan(&id).Error
		if err == nil && id > 0 {
			ids = append(ids, id)
			continue
		}
		unmatched = append(unmatched, n)
	}
	return uniqueInt64s(ids), unmatched
}

// ResolveStudioNameToID returns studio id by exact or prefix/substring match.
func ResolveStudioNameToID(ctx context.Context, name string) (int64, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return 0, nil
	}
	var id int64
	if err := common.DB.WithContext(ctx).
		Model(&models.JavStudio{}).
		Select("id").
		Where("name = ? COLLATE NOCASE", n).
		Limit(1).
		Scan(&id).Error; err != nil {
		return 0, fmt.Errorf("resolve studio: %w", err)
	}
	if id > 0 {
		return id, nil
	}
	if err := common.DB.WithContext(ctx).
		Model(&models.JavStudio{}).
		Select("id").
		Where("name LIKE ? COLLATE NOCASE", "%"+n+"%").
		Order("LENGTH(name) ASC").
		Limit(1).
		Scan(&id).Error; err != nil {
		return 0, fmt.Errorf("resolve studio fuzzy: %w", err)
	}
	return id, nil
}
