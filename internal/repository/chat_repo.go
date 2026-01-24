package repository

import (
	"github.com/mildlybrutal/websocketGo/internal/server/models"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveMessage(chat *models.Chat) error {
	return r.db.Create(chat).Error
}

func (r *ChatRepository) GetRoomHistory(roomID uint, limit int) ([]models.Chat, error) {
	var messages []models.Chat
	err := r.db.Where("room_id = ?", roomID).
		Order("created_at asc").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
