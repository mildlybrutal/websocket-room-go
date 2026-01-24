package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         uint       `gorm:"primary key;autoincrement" json:"id"`
	Username   *string    `json:"username"`
	Email      *string    `json:"email"`
	Avatar_url *string    `json:"avatar_url"`
	Created_at *time.Time `json:"created_at"`
}

func MigrateUser(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})

	return err
}
