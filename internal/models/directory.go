package models

import "time"

// Directory represents a root path that can be scanned for videos.
type Directory struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Path      string    `json:"path" gorm:"uniqueIndex"`
	Missing   bool      `json:"missing" gorm:"index"`
	IsDelete  bool      `json:"is_delete" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
