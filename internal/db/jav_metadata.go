package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JavMetadataPatch carries user edits for a single JAV row. Non-nil / non-empty fields are applied.
// When Lock is nil or true, each applied field sets its corresponding lock so future scrapes skip it.
type JavMetadataPatch struct {
	Title       *string
	TitleEn     *string
	StudioID    *int64
	StudioName  string
	ClearStudio bool
	SeriesID    *int64
	SeriesName  string
	ClearSeries bool
	TagIDs      []int64 // nil = leave tags unchanged; non-nil replaces all tag associations
	Lock        *bool
}

func patchLockEnabled(patch JavMetadataPatch) bool {
	if patch.Lock == nil {
		return true
	}
	return *patch.Lock
}

// PatchJavMetadata applies user corrections to a JAV row.
func PatchJavMetadata(ctx context.Context, javID int64, patch JavMetadataPatch, isEnglish bool) (*models.Jav, error) {
	if javID <= 0 {
		return nil, errors.New("jav id must be positive")
	}
	lock := patchLockEnabled(patch)

	var out models.Jav
	err := common.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rec models.Jav
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&rec, javID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("jav not found")
			}
			return fmt.Errorf("load jav: %w", err)
		}

		updates := map[string]any{}

		if patch.Title != nil {
			rec.Title = strings.TrimSpace(*patch.Title)
			updates["title"] = rec.Title
			if lock {
				updates["title_locked"] = true
				rec.TitleLocked = true
			}
		}
		if patch.TitleEn != nil {
			rec.TitleEn = strings.TrimSpace(*patch.TitleEn)
			updates["title_en"] = rec.TitleEn
			if lock {
				updates["title_en_locked"] = true
				rec.TitleEnLocked = true
			}
		}

		if patch.ClearStudio {
			rec.StudioID = nil
			updates["studio_id"] = nil
			if lock {
				updates["studio_locked"] = true
				rec.StudioLocked = true
			}
		} else if patch.StudioID != nil && *patch.StudioID > 0 {
			var studio models.JavStudio
			if err := tx.First(&studio, *patch.StudioID).Error; err != nil {
				return fmt.Errorf("find studio: %w", err)
			}
			rec.StudioID = &studio.ID
			updates["studio_id"] = studio.ID
			if lock {
				updates["studio_locked"] = true
				rec.StudioLocked = true
			}
		} else if name := strings.TrimSpace(patch.StudioName); name != "" {
			studioRec, err := ensureStudioTx(tx, name)
			if err != nil {
				return err
			}
			rec.StudioID = &studioRec.ID
			updates["studio_id"] = studioRec.ID
			if lock {
				updates["studio_locked"] = true
				rec.StudioLocked = true
			}
		}

		seriesColumn := "series_id"
		seriesLockColumn := "series_locked"
		if isEnglish {
			seriesColumn = "series_en_id"
			seriesLockColumn = "series_en_locked"
		}

		if patch.ClearSeries {
			updates[seriesColumn] = nil
			if isEnglish {
				rec.SeriesEnID = nil
			} else {
				rec.SeriesID = nil
			}
			if lock {
				updates[seriesLockColumn] = true
				if isEnglish {
					rec.SeriesEnLocked = true
				} else {
					rec.SeriesLocked = true
				}
			}
		} else if patch.SeriesID != nil && *patch.SeriesID > 0 {
			var series models.JavSeries
			if err := tx.First(&series, *patch.SeriesID).Error; err != nil {
				return fmt.Errorf("find series: %w", err)
			}
			if isEnglish {
				rec.SeriesEnID = &series.ID
			} else {
				rec.SeriesID = &series.ID
			}
			updates[seriesColumn] = series.ID
			if lock {
				updates[seriesLockColumn] = true
				if isEnglish {
					rec.SeriesEnLocked = true
				} else {
					rec.SeriesLocked = true
				}
			}
		} else if name := strings.TrimSpace(patch.SeriesName); name != "" {
			seriesRec, err := ensureSeriesWithStudioTx(tx, name, isEnglish, rec.StudioID)
			if err != nil {
				return err
			}
			if isEnglish {
				rec.SeriesEnID = &seriesRec.ID
			} else {
				rec.SeriesID = &seriesRec.ID
			}
			updates[seriesColumn] = seriesRec.ID
			if lock {
				updates[seriesLockColumn] = true
				if isEnglish {
					rec.SeriesEnLocked = true
				} else {
					rec.SeriesLocked = true
				}
			}
		}

		if patch.TagIDs != nil {
			if err := replaceJavAllTagsTx(tx, rec.ID, patch.TagIDs); err != nil {
				return err
			}
			if lock {
				updates["tags_locked"] = true
				rec.TagsLocked = true
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(&models.Jav{}).Where("id = ?", rec.ID).Updates(updates).Error; err != nil {
				return fmt.Errorf("update jav metadata: %w", err)
			}
		}

		out = rec
		return nil
	})
	if err != nil {
		return nil, err
	}

	return loadJavForResponse(ctx, out.ID, isEnglish)
}

func loadJavForResponse(ctx context.Context, javID int64, isEnglish bool) (*models.Jav, error) {
	query := common.DB.WithContext(ctx).Preload("Studio")
	if isEnglish {
		query = query.Preload("SeriesEn")
	} else {
		query = query.Preload("Series")
	}
	var rec models.Jav
	if err := query.First(&rec, javID).Error; err != nil {
		return nil, fmt.Errorf("load jav: %w", err)
	}
	items := []models.Jav{rec}
	if err := attachVisibleJavTags(ctx, items); err != nil {
		return nil, err
	}
	return &items[0], nil
}

func replaceJavAllTagsTx(tx *gorm.DB, javID int64, tagIDs []int64) error {
	if javID == 0 {
		return errors.New("jav id cannot be zero")
	}
	cleanIDs := uniqueInt64s(tagIDs)
	if len(cleanIDs) > 0 {
		var count int64
		if err := tx.Model(&models.JavTag{}).Where("id IN ?", cleanIDs).Count(&count).Error; err != nil {
			return fmt.Errorf("find jav tags: %w", err)
		}
		if count != int64(len(cleanIDs)) {
			return errors.New("invalid tag_id")
		}
	}
	if err := tx.Where("jav_id = ?", javID).Delete(&models.JavTagMap{}).Error; err != nil {
		return fmt.Errorf("delete jav tag maps: %w", err)
	}
	if len(cleanIDs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]models.JavTagMap, 0, len(cleanIDs))
	for _, tagID := range cleanIDs {
		rows = append(rows, models.JavTagMap{JavID: javID, JavTagID: tagID, CreatedAt: now})
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
		return fmt.Errorf("insert jav tag maps: %w", err)
	}
	return nil
}

func javRecordLockedForSeries(rec *models.Jav, isEnglish bool) bool {
	if rec == nil {
		return false
	}
	if isEnglish {
		return rec.SeriesEnLocked
	}
	return rec.SeriesLocked
}
