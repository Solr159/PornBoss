package models

import "time"

// Collection is a user-defined playlist of JAV entries.
type Collection struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex:idx_collection_name;not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JavCollection is the join table between collections and jav rows.
type JavCollection struct {
	CollectionID int64     `json:"collection_id" gorm:"primaryKey;not null"`
	JavID        int64     `json:"jav_id" gorm:"primaryKey;not null;index:idx_jav_collection_jav_id"`
	CreatedAt    time.Time `json:"created_at"`

	Collection Collection `json:"-" gorm:"foreignKey:CollectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Jav        Jav        `json:"-" gorm:"foreignKey:JavID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
