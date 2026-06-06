package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pornboss/internal/common"
	"pornboss/internal/models"

	"gorm.io/gorm"
)

const maxVideoMarkerNoteLen = 500

// ListVideoMarkers returns markers for a video ordered by time.
func ListVideoMarkers(ctx context.Context, videoID int64) ([]models.VideoMarker, error) {
	if videoID <= 0 {
		return nil, errors.New("invalid video id")
	}
	var items []models.VideoMarker
	err := common.DB.WithContext(ctx).
		Where("video_id = ?", videoID).
		Order("time_sec ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list video markers: %w", err)
	}
	if items == nil {
		items = []models.VideoMarker{}
	}
	return items, nil
}

// CreateVideoMarker inserts a marker for the given video.
func CreateVideoMarker(ctx context.Context, videoID int64, timeSec float64, note string) (*models.VideoMarker, error) {
	if videoID <= 0 {
		return nil, errors.New("invalid video id")
	}
	if timeSec < 0 {
		return nil, errors.New("time must be non-negative")
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return nil, errors.New("note cannot be empty")
	}
	if len(note) > maxVideoMarkerNoteLen {
		return nil, errors.New("note is too long")
	}
	var count int64
	if err := common.DB.WithContext(ctx).Model(&models.Video{}).Where("id = ?", videoID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check video: %w", err)
	}
	if count == 0 {
		return nil, errors.New("video not found")
	}

	marker := models.VideoMarker{
		VideoID: videoID,
		TimeSec: timeSec,
		Note:    note,
	}
	if err := common.DB.WithContext(ctx).Create(&marker).Error; err != nil {
		return nil, fmt.Errorf("create video marker: %w", err)
	}
	return &marker, nil
}

// UpdateVideoMarker updates time and/or note for an existing marker.
func UpdateVideoMarker(ctx context.Context, videoID, markerID int64, timeSec *float64, note *string) (*models.VideoMarker, error) {
	if videoID <= 0 || markerID <= 0 {
		return nil, errors.New("invalid id")
	}
	var marker models.VideoMarker
	if err := common.DB.WithContext(ctx).
		Where("id = ? AND video_id = ?", markerID, videoID).
		First(&marker).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("marker not found")
		}
		return nil, fmt.Errorf("get video marker: %w", err)
	}
	updates := map[string]any{}
	if timeSec != nil {
		if *timeSec < 0 {
			return nil, errors.New("time must be non-negative")
		}
		updates["time_sec"] = *timeSec
	}
	if note != nil {
		trimmed := strings.TrimSpace(*note)
		if trimmed == "" {
			return nil, errors.New("note cannot be empty")
		}
		if len(trimmed) > maxVideoMarkerNoteLen {
			return nil, errors.New("note is too long")
		}
		updates["note"] = trimmed
	}
	if len(updates) == 0 {
		return &marker, nil
	}
	if err := common.DB.WithContext(ctx).Model(&marker).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update video marker: %w", err)
	}
	if err := common.DB.WithContext(ctx).First(&marker, markerID).Error; err != nil {
		return nil, fmt.Errorf("reload video marker: %w", err)
	}
	return &marker, nil
}

// DeleteVideoMarker removes a marker belonging to a video.
func DeleteVideoMarker(ctx context.Context, videoID, markerID int64) error {
	if videoID <= 0 || markerID <= 0 {
		return errors.New("invalid id")
	}
	res := common.DB.WithContext(ctx).
		Where("id = ? AND video_id = ?", markerID, videoID).
		Delete(&models.VideoMarker{})
	if res.Error != nil {
		return fmt.Errorf("delete video marker: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return errors.New("marker not found")
	}
	return nil
}
