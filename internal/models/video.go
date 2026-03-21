package models

import "time"

// Video represents metadata tracked for a single video file.
type Video struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	DirectoryID  int64     `json:"directory_id" gorm:"index;not null"`
	Path         string    `json:"path" gorm:"index"` // Relative to directory root
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
	Fingerprint  string    `json:"fingerprint" gorm:"uniqueIndex"`
	DurationSec  int64     `json:"duration_sec"` // duration in seconds (rounded)
	PlayCount    int64     `json:"play_count" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Tags         []Tag     `json:"tags,omitempty" gorm:"many2many:video_tag"`
	JavID        *int64    `json:"jav_id" gorm:"index"`
	Jav          *Jav      `json:"jav,omitempty" gorm:"foreignKey:JavID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	DirectoryRef Directory `json:"directory,omitempty" gorm:"foreignKey:DirectoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Hidden       bool      `json:"hidden" gorm:"index"`
}
