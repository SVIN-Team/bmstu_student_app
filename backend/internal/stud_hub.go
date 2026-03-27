package internal

import (
	"context"
	"fmt"
	"os"
	"stud_hub/internal/config"
	"stud_hub/internal/models"
	gormrepository "stud_hub/internal/repository/gorm-repository"
	redisrepository "stud_hub/internal/repository/redis-repository"
	"stud_hub/internal/usecase"
	"stud_hub/util/logger"
	"time"

	"github.com/google/uuid"
)

func Run(cfg *config.ApplicationConfig) {
	logger.InitLogger(cfg.LoggerConfig.Level, os.Stdout)
	requestID := uuid.New().String()
	ctx := logger.ContextWithRequestID(context.Background(), requestID)
	logger.Infof(context.Background(), "Начало работы приложения")
	logger.Warnf(ctx, "Warn")
	logger.Errorf(ctx, "Error")
	logger.Debugf(ctx, "Debug")

	db, err := gormrepository.CreateDB(ctx, &cfg.RepositoryConfig.PostgresConfig)
	if err != nil {
		logger.Errorf(ctx, "fatal: cannot create database: %v", err)
		return
	}
	rd, err := redisrepository.InitRedis(ctx, &cfg.RepositoryConfig.RedisConfig)
	if err != nil {
		logger.Errorf(ctx, "fatal: cannot connect to redis: %v", err)
		return
	}

	tokenRepo := redisrepository.NewTokenRepository(rd)
	userRepo := gormrepository.NewUserRepository(db)
	groupRepo := gormrepository.NewGroupRepository(db)
	teacherRepo := gormrepository.NewTeacherRepository(db)
	subjectRepo := gormrepository.NewSubjectRepository(db)
	classroomRepo := gormrepository.NewClassroomRepository(db)
	queueRepo := gormrepository.NewQueueRepository(db)
	queueSlotsRepo := gormrepository.NewQueueSlotsRepository(db)
	lessonRepo := gormrepository.NewLessonRepository(db)

	auth := usecase.NewAuthUseCase(tokenRepo, userRepo, cfg.AuthConfig)
	_ = usecase.NewGroupUseCase(groupRepo)
	_ = usecase.NewScheduleUseCase(lessonRepo, subjectRepo, groupRepo, teacherRepo, classroomRepo, queueRepo)
	_ = usecase.NewQueueUseCase(queueRepo, queueSlotsRepo, userRepo)
	if cfg.TestMode {
		logger.Infof(ctx, "Приложение запущено! ^w^")
		tmp_app(ctx, auth)
	}
}

func tmp_app(ctx context.Context, auth *usecase.AuthUseCase) {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			{
				logger.Infof(ctx, "stopped loop")
				ticker.Stop()
				return
			}
		case <-ticker.C:
			{
				logger.Infof(ctx, "Создать тестового пользователя")
				_, _, err := auth.SignUp(ctx, models.User{
					FirstName:    "Test",
					LastName:     "Dog",
					Patronymic:   "Patron",
					PasswordHash: "password",
					Email:        fmt.Sprintf("email@%s.com", uuid.New().String()),
				})
				if err != nil {
					logger.Errorf(ctx, "cannot create user: %v", err)
				} else {
					logger.Infof(ctx, "created user successfully")
				}
			}
		}
	}
}
