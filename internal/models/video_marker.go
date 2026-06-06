package models

import "time"

// VideoMarker is a user note anchored at a specific playback time on a video.
type VideoMarker struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	VideoID   int64     `json:"video_id" gorm:"index;not null"`
	TimeSec   float64   `json:"time_sec" gorm:"not null"`
	Note      string    `json:"note" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Video Video `json:"-" gorm:"foreignKey:VideoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
