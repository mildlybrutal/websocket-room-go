package models

import (
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	ID          uint              `gorm:"primary key;autoincrement" json:"id"`
	RoomID      uint              `gorm:"many2many:room" json:"room_id"`
	SenderID    string            `json:"string"`
	MessageType string            `json:"string"`
	Content     string            `json:"string"`
	Metadata    pgtype.JSONBCodec `gorm:"type:jsonb;default:'[]'"`
}

func MigrateChat(db *gorm.DB) error {
	err := db.AutoMigrate(&Chat{})

	return err
}
