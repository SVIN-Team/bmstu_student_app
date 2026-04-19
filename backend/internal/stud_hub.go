package internal

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "time"

    _ "stud_hub/docs"
    "stud_hub/internal/config"
    http2 "stud_hub/internal/handler/http"
    "stud_hub/internal/handler/middleware"
    "stud_hub/internal/models"
    gormrepository "stud_hub/internal/repository/gorm-repository"
    redisrepository "stud_hub/internal/repository/redis-repository"
    "stud_hub/internal/usecase"
    "stud_hub/util/logger"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func Run(cfg *config.ApplicationConfig) {
    logger.InitLogger(cfg.LoggerConfig.Level, os.Stdout)
    requestID := uuid.New().String()
    ctx := logger.ContextWithRequestID(context.Background(), requestID)

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

    // Initialize use cases
    authUseCase := usecase.NewAuthUseCase(tokenRepo, userRepo, cfg.AuthConfig)
    groupUseCase := usecase.NewGroupUseCase(groupRepo)
    scheduleUseCase := usecase.NewScheduleUseCase(lessonRepo, subjectRepo, groupRepo, teacherRepo, classroomRepo, queueRepo)
    queueUseCase := usecase.NewQueueUseCase(queueRepo, queueSlotsRepo, userRepo, lessonRepo)
    userUseCase := usecase.NewUserUseCase(userRepo, queueSlotsRepo, queueRepo)
    adminUseCase := usecase.NewAdminUseCase(userRepo)
    subjectUseCase := usecase.NewSubjectUseCase(subjectRepo)
    if cfg.TestMode {
        logger.Infof(ctx, "Приложение запущено! ^w^")
        tmpApp(ctx, authUseCase)
    }

    // Initialize handlers
    authHandler := http2.NewAuthHandler(authUseCase, groupUseCase, &cfg.AuthConfig)
    scheduleHandler := http2.NewScheduleHandler(scheduleUseCase, userUseCase)
    queueHandler := http2.NewQueueHandler(queueUseCase, groupUseCase, subjectUseCase, userUseCase)
    queueSlotsHandler := http2.NewQueueSlotsHandler(queueUseCase, queueUseCase)
    userHandler := http2.NewUserHandler(userUseCase, groupUseCase)
    adminHandler := http2.NewAdminHandler(adminUseCase, groupUseCase)

    authMiddleware := middleware.NewAuthMiddleware(cfg.AuthConfig, userUseCase)

    // Setup router
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.RequestIDMiddleware())

    // Health check
    r.GET("/api/v1/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "application is healthy",
        })
    })

    if cfg.EnableSwagger {
        r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    // Authentication routes (public)
    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/register", authHandler.SignUp)
        auth.POST("/login", authHandler.SignIn)
        auth.POST("/refresh", authHandler.Refresh)
        auth.POST("/logout", authMiddleware.AuthRequired(), authHandler.SignOut)        // requires auth
        auth.POST("/logout-all", authMiddleware.AuthRequired(), authHandler.SignOutAll) // requires auth
    }

    // Legacy routes (for compatibility)
    //r.POST("/api/v1/users", authHandler.SignUp)
    //r.POST("/api/v1/sessions", authHandler.SignIn)

    // Protected routes (require authentication)
    api := r.Group("/api/v1")
    api.Use(authMiddleware.AuthRequired(), authMiddleware.AuthRequiredWithRole())
    {
        // Lessons/Schedule
        lessons := api.Group("/lessons")
        {
            lessons.GET("", scheduleHandler.GetLessons)
            lessons.GET("/:id", scheduleHandler.GetLessonByID)

            lessonsAdmin := lessons.Group("")
            lessonsAdmin.POST("/imports", authMiddleware.AdminRequired(), scheduleHandler.ImportSchedule)
        }

        // Queues
        queues := api.Group("/queues")
        {
            queues.GET("", queueHandler.GetQueues)
            queues.POST("", authMiddleware.HeadmanRequired(), queueHandler.CreateQueue) // headman only
            queues.GET("/:queue_id", queueHandler.GetQueueByID)
            queues.PATCH("/:queue_id", authMiddleware.HeadmanRequired(), queueHandler.UpdateQueue)  // headman only
            queues.DELETE("/:queue_id", authMiddleware.HeadmanRequired(), queueHandler.DeleteQueue) // headman only

            // Queue slots
            queues.GET("/:queue_id/slots", queueSlotsHandler.GetQueueSlots)
            queues.POST("/:queue_id/slots", queueSlotsHandler.SignUpForQueue)
            queues.GET("/:queue_id/slots/:slot_id", queueSlotsHandler.GetQueueSlot)
            queues.PATCH("/:queue_id/slots/:slot_id", authMiddleware.HeadmanRequired(), queueSlotsHandler.UpdateQueueSlot) // headman only
            queues.DELETE("/:queue_id/slots/me", queueSlotsHandler.CancelQueueSlot)

            // Slot transfers
            queues.POST("/:queue_id/transfers", authMiddleware.HeadmanRequired(), queueSlotsHandler.TransferFailedSlots) // headman only
        }

        // User profile
        users := api.Group("/users")
        {
            users.GET("/me", userHandler.GetCurrentUser)
            users.PATCH("/me", userHandler.UpdateCurrentUser)
            users.GET("/me/slots", userHandler.GetCurrentUserSlots)
            users.PUT("/me/headman-role", authMiddleware.HeadmanRequired(), userHandler.TransferHeadmanRole) // headman only
        }

        // Admin routes
        admin := api.Group("/admin")
        admin.Use(authMiddleware.AdminRequired()) // Require admin role
        {
            // User management
            adminUsers := admin.Group("/users")
            {
                adminUsers.GET("", adminHandler.GetUsers)
                adminUsers.GET("/:id", adminHandler.GetUserByID)
                adminUsers.PATCH("/:id", adminHandler.UpdateUser)
                adminUsers.DELETE("/:id", adminHandler.DeleteUser)
            }

            // Group management
            adminGroups := admin.Group("/groups")
            {
                adminGroups.GET("", adminHandler.GetGroups)
                adminGroups.POST("", adminHandler.CreateGroup)
                adminGroups.GET("/:id", adminHandler.GetGroupByID)
                adminGroups.PATCH("/:id", adminHandler.UpdateGroup)
                adminGroups.DELETE("/:id", adminHandler.DeleteGroup)
            }
            // TODO: Subjects, Teachers, Rooms CRUD
        }
    }

    gin.SetMode(cfg.RoutingConfig.GinMode)

    err = r.Run(fmt.Sprintf(":%d", cfg.RoutingConfig.Port))
    if err != nil {
        logger.Errorf(ctx, "fatal: cannot start server: %v", err)
        return
    }
}

func tmpApp(ctx context.Context, auth *usecase.AuthUseCase) {
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
                _, _, _, err := auth.SignUp(ctx, models.User{
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
