package models

import "time"

// Tag represents a label that can be assigned to videos.
type Tag struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VideoTag is the join table between videos and tags.
type VideoTag struct {
	VideoID   int64 `gorm:"primaryKey"`
	Video     Video `gorm:"foreignKey:VideoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TagID     int64 `gorm:"primaryKey"`
	Tag       Tag   `gorm:"foreignKey:TagID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time
}
