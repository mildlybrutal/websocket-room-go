package models

import (
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	ID       uint              `gorm:"primary key;autoincrement" json:"id"`
	RoomID   uint              `gorm:"not null;index" json:"room_id"`
	SenderID uint              `gorm:"not null;index" json:"sender_id"`
	Content  string            `gorm:"type:text;not null" json:"content"`
	Metadata pgtype.JSONBCodec `gorm:"type:jsonb;default:'[]'"`
	Room     Room              `gorm:"foreignKey:RoomID"`
	Sender   User              `gorm:"foreignKey:SenderID"`
}

func MigrateChat(db *gorm.DB) error {
	err := db.AutoMigrate(&Chat{})

	return err
}
