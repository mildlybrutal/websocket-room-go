package models

import (
	"time"

	"gorm.io/gorm"
)

type RoomMember struct {
	gorm.Model
	UserID   uint      `gorm:"primaryKey" json:"user_id"`
	RoomID   uint      `gorm:"primaryKey" json:"room_id"`
	Role     string    `gorm:"default:member" json:"role"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
	User     User      `gorm:"foreignKey:UserID"`
	Room     Room      `gorm:"foreignKey:RoomID"`
}

func MigrateRoomMember(db *gorm.DB) error {
	err := db.AutoMigrate(&RoomMember{})

	return err
}
