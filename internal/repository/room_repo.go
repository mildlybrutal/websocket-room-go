package repository

import (
	"github.com/mildlybrutal/websocketGo/internal/server/models"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateRoom(name string, ownerID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		room := &models.Room{
			Name:    name,
			OwnerID: ownerID,
		}

		if err := tx.Create(room).Error; err != nil {
			return err
		}

		// Add owner as a member automatically
		membership := &models.RoomMember{
			UserID: ownerID,
			RoomID: room.ID,
			Role:   "owner",
		}
		if err := tx.Create(membership).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *RoomRepository) AddMember(roomID uint, userID uint) error {
	// check if already a member
	var count int64
	r.db.Model(&models.RoomMember{}).Where("room_id = ? AND user_id = ?", roomID, userID).Count(&count)
	if count > 0 {
		return nil // already member
	}

	member := &models.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   "member",
	}
	return r.db.Create(member).Error
}

func (r *RoomRepository) GetRoomByID(roomID uint) (*models.Room, error) {
	var room models.Room
	err := r.db.First(&room, roomID).Error
	return &room, err
}

func (r *RoomRepository) GetUserRooms(userID uint) ([]models.Room, error) {
	var rooms []models.Room
	// Join with room_members table to get rooms the user is a member of
	err := r.db.Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ?", userID).
		Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepository) GetAllRooms() ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepository) DeleteRoom(roomID uint) error {
	return r.db.Delete(&models.Room{}, roomID).Error
}

func (r *RoomRepository) CheckUserRoomAccess(userID uint, roomID uint) (bool, error) {
	return true, nil // Temporarily allow all access until logic is confirmed for private rooms
	/*
		var members models.RoomMember
		err := r.db.Where("user_id = ? AND room_id = ?", userID, roomID).First(&members).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return false, nil
			}
			return false, err
		}

		return true, nil
	*/
}
