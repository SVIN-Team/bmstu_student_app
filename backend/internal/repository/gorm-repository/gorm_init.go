package gormrepository

import (
	"context"
	"os"
	"stud_hub/internal/config"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CreateDB initializes a GORM PostgreSQL connection and runs automigrations for all models.
func CreateDB(ctx context.Context, cfg *config.PostgresConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if cfg.PerformOrmMigration {
		logger.Infof(ctx, "performing gorm migration")
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
	}

	seedFirstAdmin(ctx, db)

	logger.Infof(ctx, "successfully connected to postgres")
	return db, nil
}

func seedFirstAdmin(ctx context.Context, db *gorm.DB) {
	adminEmail := os.Getenv("FIRST_ADMIN_EMAIL")
	adminPassword := os.Getenv("FIRST_ADMIN_PASSWORD")

	err := os.Unsetenv("FIRST_ADMIN_EMAIL")
	if err != nil {
		return
	}
	err = os.Unsetenv("FIRST_ADMIN_PASSWORD")
	if err != nil {
		return
	}

	if adminEmail == "" || adminPassword == "" {
		return
	}

	var count int64
	db.Model(&gormmodels.User{}).Where("role = ?", gormmodels.UserRoleAdmin).Count(&count)
	if count > 0 {
		logger.Warnf(ctx, "FIRST_ADMIN_EMAIL already exists")
		return // Admin already exists
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf(ctx, "failed to hash first admin password: %v", err)
		return
	}
	hashedPassword := string(hashedPasswordBytes)

	adminUser := gormmodels.User{
		Email:        adminEmail,
		PasswordHash: &hashedPassword,
		FirstName:    "Admin",
		LastName:     "System",
		Role:         gormmodels.UserRoleAdmin,
		IsBlocked:    false,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		logger.Errorf(ctx, "failed to create first admin: %v", err)
	} else {
		logger.Infof(ctx, "first admin successfully created with email: %s", adminEmail)
	}
}
