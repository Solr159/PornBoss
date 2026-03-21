package models

import "time"

// Config stores simple key/value settings persisted in the database.
type Config struct {
	Key       string    `json:"key" gorm:"primaryKey;size:64"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
