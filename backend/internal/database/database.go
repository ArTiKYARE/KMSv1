package database

import (
	"fmt"
	"log"

	"github.com/kms-messenger/backend/internal/config"
	"github.com/kms-messenger/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) error {
	var err error
	
	dbLogger := logger.Default.LogMode(logger.Info)
	if cfg.AppEnv == "production" {
		dbLogger = logger.Default.LogMode(logger.Warn)
	}

	DB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: dbLogger,
	})
	
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connection established and migrated successfully")
	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Chat{},
		&models.ChatMember{},
		&models.Message{},
	)
}

func GetDB() *gorm.DB {
	return DB
}
