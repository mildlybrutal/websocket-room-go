package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
}

func MigrateUser(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})

	return err
}
