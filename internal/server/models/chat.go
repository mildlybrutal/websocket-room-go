package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	ID       uint            `gorm:"primary key;autoincrement" json:"id"`
	RoomID   uint            `gorm:"not null;index" json:"room_id"`
	SenderID uint            `gorm:"not null;index" json:"sender_id"`
	Content  string          `gorm:"type:text;not null" json:"content"`
	Metadata json.RawMessage `gorm:"type:jsonb;default:'[]'"`
	Room     Room            `gorm:"foreignKey:RoomID"`
	Sender   User            `gorm:"foreignKey:SenderID"`

	IsEdited bool       `gorm:"default:false" json:"is_edited"`
	EditedAt *time.Time `gorm:"index" json:"edited_at,omitempty"`

	IsDeleted bool `gorm:"default:false;index" json:"is_deleted"`
	DeletedBy uint `gorm:"default:0" json:"deleted_by"`
}

type MessageReadReceipt struct {
	gorm.Model
	MessageID uint      `gorm:"not null;index:idx_message_user,unique" json:"message_id"`
	UserID    uint      `gorm:"not null;index:idx_message_user,unique" json:"user_id"`
	ReadAt    time.Time `gorm:"not null" json:"read_at"`
	Message   Chat      `gorm:"foreignKey:MessageID"`
	User      User      `gorm:"foreignKey:UserID"`
}

func MigrateChat(db *gorm.DB) error {
	err := db.AutoMigrate(&Chat{}, &MessageReadReceipt{})

	return err
}
