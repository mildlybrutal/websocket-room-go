package repository

import (
	"github.com/mildlybrutal/websocketGo/internal/server/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetOrCreateUser(username string, email string) (*models.User, error) {
	var user models.User

	err := r.db.Where(models.User{
		Username: username,
	}).Attrs(models.User{Email: email}).FirstOrCreate(&user).Error

	return &user, err
}
