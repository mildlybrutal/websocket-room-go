package storage

import (
	"fmt"

	"github.com/mildlybrutal/websocketGo/internal/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(config *common.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s maxcons=%d",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode, config.MaxCons,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return db, err
	}

	return db, nil
}
