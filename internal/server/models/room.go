package models

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	Name    string `gorm:"uniqueIndex;not null" json:"name"`
	OwnerID uint   `gorm:"not null" json:"owner_id"`
	Owner   User   `gorm:"foreignKey:OwnerID"`
}

func MigrateRoom(db *gorm.DB) error {
	err := db.AutoMigrate(&Room{})

	return err
}
