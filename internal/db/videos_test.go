package db

import (
	"context"
	"testing"
	"time"

	"pornboss/internal/models"
)

func TestListVideosSortByDurationDirections(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/media"}
	if err := db.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}

	shortVideo := models.Video{
		DirectoryID: dir.ID,
		Path:        "short.mp4",
		Filename:    "short.mp4",
		Fingerprint: "video-fp-short",
		DurationSec: 90,
		ModifiedAt:  now,
		CreatedAt:   now,
	}
	longVideo := models.Video{
		DirectoryID: dir.ID,
		Path:        "long.mp4",
		Filename:    "long.mp4",
		Fingerprint: "video-fp-long",
		DurationSec: 180,
		ModifiedAt:  now,
		CreatedAt:   now.Add(time.Second),
	}
	if err := db.Create(&shortVideo).Error; err != nil {
		t.Fatalf("create short video: %v", err)
	}
	if err := db.Create(&longVideo).Error; err != nil {
		t.Fatalf("create long video: %v", err)
	}

	items, err := ListVideos(ctx, 20, 0, nil, "", "duration", nil)
	if err != nil {
		t.Fatalf("ListVideos duration: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected item count: got %d want 2", len(items))
	}
	if items[0].ID != longVideo.ID {
		t.Fatalf("unexpected first video: got %d want %d", items[0].ID, longVideo.ID)
	}

	items, err = ListVideos(ctx, 20, 0, nil, "", "duration_asc", nil)
	if err != nil {
		t.Fatalf("ListVideos duration_asc: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected asc item count: got %d want 2", len(items))
	}
	if items[0].ID != shortVideo.ID {
		t.Fatalf("unexpected asc first video: got %d want %d", items[0].ID, shortVideo.ID)
	}
}
