package models

import (
	"time"

	"gorm.io/gorm"
)

type RoomMember struct {
	gorm.Model
	UserID   uint       `gorm:"many2many:user" json:"user_id"`
	RoomID   *uint      `gorm:"many2many:room" json:"room_id"`
	Role     *string    `gorm:"many2many:user" json:"user_id"`
	JoinedAt *time.Time `gorm:"many2many:user" json:"user_id"`
}

func MigrateRoomMember(db *gorm.DB) error {
	err := db.AutoMigrate(&RoomMember{})

	return err
}
