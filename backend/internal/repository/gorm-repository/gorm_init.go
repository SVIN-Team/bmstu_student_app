package gormrepository

import (
	"context"
	"stud_hub/internal/config"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CreateDB initializes a GORM PostgreSQL connection and runs automigrations for all models.
func CreateDB(ctx context.Context, cfg *config.PostgresConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&gormmodels.Group{},
		&gormmodels.User{},
		&gormmodels.Subject{},
		&gormmodels.Teacher{},
		&gormmodels.Room{},
		&gormmodels.Lesson{},
		&gormmodels.Queue{},
		&gormmodels.QueueSlot{},
	); err != nil {
		logger.Errorf(ctx, "gorm auto-migrate failed: %v", err)
		return nil, err
	}
	logger.Infof(ctx, "successfullly connected to postgres")
	return db, nil
}
