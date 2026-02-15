package repository

import (
	"time"

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

func (r *ChatRepository) DeleteMessage(messageID uint, deleterID uint) error {
	now := time.Now()
	return r.db.Model(&models.Chat{}).
		Where("id = ? AND sender_id = ? AND is_deleted = false", messageID, deleterID).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": deleterID,
			"content":    "[Message deleted]",
		}).Error
}

func (r *ChatRepository) EditMessage(messageID uint, newContent string, editorID uint) error {
	now := time.Now()

	return r.db.Model(&models.Chat{}).
		Where("id = ? AND sender_id = ? AND is_deleted = false", messageID, editorID).
		Updates(map[string]any{
			"content":    newContent,
			"is_edited":  true,
			"edited_at":  now,
			"updated_at": now,
		}).Error
}

func (r *ChatRepository) MarkMessageAsRead(messageID uint, userID uint) error {
	receipt := &models.MessageReadReceipt{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}
	return r.db.Where(models.MessageReadReceipt{
		MessageID: messageID,
		UserID:    userID,
	}).FirstOrCreate(receipt).Error
}

func (r *ChatRepository) GetMessageReadReceipts(messageID uint) ([]models.MessageReadReceipt, error) {
	var receipts []models.MessageReadReceipt
	err := r.db.Where("message_id = ?", messageID).
		Preload("User").
		Order("read_at desc").
		Find(&receipts).Error
	return receipts, err
}

func (r *ChatRepository) GetUnreadMessagesForUser(roomID uint, userID uint) ([]models.Chat, error) {
	var messages []models.Chat
	err := r.db.Where("room_id = ? AND sender_id != ? AND is_deleted = false", roomID, userID).
		Where("id NOT IN (SELECT message_id FROM message_read_receipts WHERE user_id = ?)", userID).
		Preload("Sender").
		Order("created_at asc").
		Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) GetMessageByID(messageID uint) (*models.Chat, error) {
	var message models.Chat
	err := r.db.Preload("Sender").First(&message, messageID).Error
	return &message, err
}

func (r *ChatRepository) GetRoomHistory(roomID uint, limit int) ([]models.Chat, error) {
	var messages []models.Chat
	err := r.db.Where("room_id = ?", roomID).
		Preload("Sender").
		Order("created_at asc").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
