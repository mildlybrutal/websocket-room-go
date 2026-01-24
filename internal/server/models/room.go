package models

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	ID        uint       `gorm:"primary key;autoincrement" json:"id"`
	Name      *string    `json:"name"`
	OwnerID   *int       `gorm"foreignkey:user;references:ID" json:"owner_id`
	IsPrivate *bool      `json:"is_private"`
	CreatedAt *time.Time `json:"created_at"`
}

func MigrateRoom(db *gorm.DB) error {
	err := db.AutoMigrate(&Room{})

	return err
}
