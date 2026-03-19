package internal

import (
	"context"
	"os"
	"stud_hub/internal/config"
	"stud_hub/internal/usecase"
	"stud_hub/util/logger"

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

	_ = usecase.NewAuthUseCase(nil, nil, cfg.AuthConfig)
	_ = usecase.NewGroupUseCase(nil)
	_ = usecase.NewScheduleUseCase(nil, nil, nil, nil, nil, nil)
	_ = usecase.NewQueueUseCase(nil, nil)

}
