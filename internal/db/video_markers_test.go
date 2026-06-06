package db

import (
	"context"
	"testing"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/models"
)

func TestVideoMarkersCRUD(t *testing.T) {
	openTestDB(t)
	ctx := context.Background()
	now := time.Unix(1710000000, 0).UTC()

	dir := models.Directory{Path: "/tmp/markers"}
	if err := common.DB.Create(&dir).Error; err != nil {
		t.Fatalf("create directory: %v", err)
	}
	video := models.Video{
		Fingerprint: "marker-test-fp",
		DurationSec: 600,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := common.DB.Create(&video).Error; err != nil {
		t.Fatalf("create video: %v", err)
	}

	created, err := CreateVideoMarker(ctx, video.ID, 125.5, "scene one")
	if err != nil {
		t.Fatalf("CreateVideoMarker: %v", err)
	}
	if created.TimeSec != 125.5 || created.Note != "scene one" {
		t.Fatalf("unexpected created marker: %+v", created)
	}

	items, err := ListVideoMarkers(ctx, video.ID)
	if err != nil {
		t.Fatalf("ListVideoMarkers: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("unexpected list: %+v", items)
	}

	newTime := 200.0
	newNote := "updated"
	updated, err := UpdateVideoMarker(ctx, video.ID, created.ID, &newTime, &newNote)
	if err != nil {
		t.Fatalf("UpdateVideoMarker: %v", err)
	}
	if updated.TimeSec != newTime || updated.Note != newNote {
		t.Fatalf("unexpected updated marker: %+v", updated)
	}

	if err := DeleteVideoMarker(ctx, video.ID, created.ID); err != nil {
		t.Fatalf("DeleteVideoMarker: %v", err)
	}
	items, err = ListVideoMarkers(ctx, video.ID)
	if err != nil {
		t.Fatalf("ListVideoMarkers after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %+v", items)
	}
}
