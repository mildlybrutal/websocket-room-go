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
	room := &models.Room{
		Name:    name,
		OwnerID: ownerID,
	}

	return r.db.Create(room).Error
}

func (r *RoomRepository) GetRoomByID(roomID uint) (*models.Room, error) {
	var room models.Room
	err := r.db.First(&room, roomID).Error
	return &room, err
}

func (r *RoomRepository) GetUserRooms(userID uint) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.Where("owner_id = ?", userID).Find(&rooms).Error
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
	var members models.RoomMember

	err := r.db.Where("user_id = ? AND room_id = ?", userID, roomID).First(&members).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
