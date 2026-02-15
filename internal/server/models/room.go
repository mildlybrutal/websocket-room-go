package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	Name    string `gorm:"uniqueIndex;not null" json:"name"`
	OwnerID uint   `gorm:"not null" json:"owner_id"`
	Owner   User   `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}

// MarshalJSON implements custom JSON serialization to ensure ID is lowercase
func (r Room) MarshalJSON() ([]byte, error) {
	type RoomJSON struct {
		ID      uint   `json:"id"`
		Name    string `json:"name"`
		OwnerID uint   `json:"owner_id"`
	}
	return json.Marshal(RoomJSON{
		ID:      r.ID,
		Name:    r.Name,
		OwnerID: r.OwnerID,
	})
}

func MigrateRoom(db *gorm.DB) error {
	err := db.AutoMigrate(&Room{})

	return err
}
