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
	if err != nil{
		logger.Errorf(ctx, "fatal: cannot create database: %v", err)
		return
	}
	rd, err := redisrepository.InitRedis(ctx, &cfg.RepositoryConfig.RedisConfig)
	if err != nil{
		logger.Errorf(ctx, "fatal: cannot connect to redis: %v", err)
		return
	}

	token_repo := redisrepository.NewTokenRepository(rd)
	user_repo := gormrepository.NewUserRepository(db)
	group_repo := gormrepository.NewGroupRepository(db)
	teacher_repo := gormrepository.NewTeacherRepository(db)
	subject_repo := gormrepository.NewSubjectRepository(db)
	classroom_repo := gormrepository.NewClassroomRepository(db)
	queue_repo := gormrepository.NewQueueRepository(db)
	queue_slots_repo := gormrepository.NewQueueSlotsRepository(db)
	lesson_repo := gormrepository.NewLessonRepository(db)


	auth := usecase.NewAuthUseCase(token_repo, user_repo, cfg.AuthConfig)
	_ = usecase.NewGroupUseCase(group_repo)
	_ = usecase.NewScheduleUseCase(lesson_repo, subject_repo, group_repo, teacher_repo, classroom_repo, queue_repo)
	_ = usecase.NewQueueUseCase(queue_repo, queue_slots_repo, user_repo)

	logger.Infof(ctx, "Приложение запущено! ^w^")
	go tmp_app(ctx, auth)
}


func tmp_app(ctx context.Context, auth *usecase.AuthUseCase) {
	ticker := time.NewTicker(time.Second * 10)
	select {
	case <-ctx.Done(): {
		return
	}
	case <-ticker.C: {
		logger.Infof(ctx, "Создать тестового пользователя")
		a, b, err := auth.SignUp(ctx, models.User{
			FirstName: "Test",
			LastName: "Dog",
			Patronymic: "Patron",
			Email: fmt.Sprintf("email@%s.com",time.Now().String()),
		})
		if err != nil {
			logger.Errorf(ctx, "cannot create user: %v", err)
		} else {
			logger.Infof(ctx, "created user, got the %s and %s as tokens", a, b)
		}
	}
	}
	
	logger.Infof(ctx, "Всё!")
}