package storage

import (
	"fmt"
	"log"

	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/server/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(config *common.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return db, err
	}

	sqlDB, err := db.DB()

	if err == nil {
		sqlDB.SetMaxOpenConns(config.MaxCons)
	}

	if err := AutoMigrate(db); err != nil {
		log.Printf("Migration failed: %v", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Room{},
		&models.RoomMember{},
		&models.Chat{},
	)
}
