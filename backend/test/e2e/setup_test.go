//go:build e2e
// +build e2e

package e2e

import (
    "context"
    "fmt"
    "os"
    "sync"
    "testing"
    "time"

    "stud_hub/internal/config"
    "stud_hub/internal/repository/gorm-models"
    "stud_hub/internal/repository/gorm-repository"
    redisrepository "stud_hub/internal/repository/redis-repository"
    "stud_hub/internal/usecase"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type TestDBConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type TestRedisConfig struct {
    ConnectionString string
}

var DefaultTestDBConfig = TestDBConfig{
    Host:     "localhost",
    Port:     "5432",
    User:     "test_user",
    Password: "test_password",
    DBName:   "stud_hub_test",
    SSLMode:  "disable",
}

var DefaultTestRedisConfig = TestRedisConfig{
    ConnectionString: "redis://localhost:6379/0",
}

var TestAuthConfig = config.AuthConfig{
    AccessSecretKey:  "test-access-secret-key-for-e2e-tests",
    RefreshSecretKey: "test-refresh-secret-key-for-e2e-tests",
    AccessLifeTime:   15 * time.Minute,
    RefreshLifeTime:  7 * 24 * time.Hour,
}

var (
    globalDB     *gorm.DB
    globalRedis  *redis.Client
    setupOnce    sync.Once
    setupErr     error
    migrationSQL []byte
)

func LoadDBConfigFromEnv() TestDBConfig {
    config := DefaultTestDBConfig
    
    if host := os.Getenv("TEST_DB_HOST"); host != "" {
        config.Host = host
    }
    if port := os.Getenv("TEST_DB_PORT"); port != "" {
        config.Port = port
    }
    if user := os.Getenv("TEST_DB_USER"); user != "" {
        config.User = user
    }
    if password := os.Getenv("TEST_DB_PASSWORD"); password != "" {
        config.Password = password
    }
    if dbname := os.Getenv("TEST_DB_NAME"); dbname != "" {
        config.DBName = dbname
    }
    if sslmode := os.Getenv("TEST_DB_SSLMODE"); sslmode != "" {
        config.SSLMode = sslmode
    }
    
    return config
}

func LoadRedisConfigFromEnv() TestRedisConfig {
    config := DefaultTestRedisConfig
    
    if connStr := os.Getenv("TEST_REDIS_CONNECTION_STRING"); connStr != "" {
        config.ConnectionString = connStr
    }
    
    return config
}

func initGlobalDB() error {
    config := LoadDBConfigFromEnv()
    
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
        config.Host,
        config.Port,
        config.User,
        config.Password,
        config.DBName,
        config.SSLMode,
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    })
    if err != nil {
        return fmt.Errorf("failed to connect to test DB: %v", err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return fmt.Errorf("failed to get underlying sql.DB: %v", err)
    }
    
    if err := sqlDB.Ping(); err != nil {
        return fmt.Errorf("failed to ping database: %v", err)
    }
    
    sqlDB.SetMaxIdleConns(1)
    sqlDB.SetMaxOpenConns(10)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    var exists bool
    db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
    
    if !exists {
        if migrationSQL == nil {
            migrationSQL, err = os.ReadFile("../../deploy/postgresql/001-init.sql")
            if err != nil {
                return fmt.Errorf("failed to read migration file: %v", err)
            }
        }
        
        if err := db.Exec(string(migrationSQL)).Error; err != nil {
            return fmt.Errorf("failed to execute migration: %v", err)
        }
    }
    
    globalDB = db
    return nil
}

func initGlobalRedis() error {
    config := LoadRedisConfigFromEnv()
    
    opts, err := redis.ParseURL(config.ConnectionString)
    if err != nil {
        return fmt.Errorf("failed to parse redis URL: %v", err)
    }
    
    client := redis.NewClient(opts)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("failed to connect to redis: %v", err)
    }
    
    globalRedis = client
    return nil
}

func SetupTestDB(t *testing.T) (*gorm.DB, *redis.Client, func()) {
    t.Helper()
    
    setupOnce.Do(func() {
        if err := initGlobalDB(); err != nil {
            setupErr = err
        }
        if err := initGlobalRedis(); err != nil {
            setupErr = err
        }
    })
    
    if setupErr != nil {
        t.Fatalf("Failed to setup test DB: %v", setupErr)
    }
    
    cleanup := func() {
        tables := []string{
            "queue_slots",
            "queues",
            "lessons",
            "users",
            "groups",
            "teachers",
            "subjects",
            "rooms",
        }
        
        for _, table := range tables {
            globalDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
        }
        
        if globalRedis != nil {
            globalRedis.FlushAll(context.Background())
        }
    }
    
    return globalDB, globalRedis, cleanup
}

func CreateTestGroup(db *gorm.DB, name string) *gormmodels.Group {
    group := &gormmodels.Group{
        ID:   uuid.New(),
        Name: name,
    }
    
    if err := db.Create(group).Error; err != nil {
        return nil
    }
    
    return group
}

func CreateTestSubject(db *gorm.DB, name string) *gormmodels.Subject {
    subject := &gormmodels.Subject{
        ID:   uuid.New(),
        Name: name,
    }
    
    if err := db.Create(subject).Error; err != nil {
        return nil
    }
    
    return subject
}

func CreateTestTeacher(db *gorm.DB, firstName, lastName string) *gormmodels.Teacher {
    teacher := &gormmodels.Teacher{
        ID:        uuid.New(),
        FirstName: firstName,
        LastName:  lastName,
    }
    
    if err := db.Create(teacher).Error; err != nil {
        return nil
    }
    
    return teacher
}

func CreateTestClassroom(db *gorm.DB, name string) *gormmodels.Room {
    room := &gormmodels.Room{
        ID:   uuid.New(),
        Name: &name,
    }
    
    if err := db.Create(room).Error; err != nil {
        return nil
    }
    
    return room
}

func CreateTestLesson(db *gorm.DB, groupID, subjectID, teacherID, roomID uuid.UUID, startsAt, endsAt time.Time) *gormmodels.Lesson {
    lessonType := gormmodels.LessonTypeLecture 
    lesson := &gormmodels.Lesson{
        ID:         uuid.New(),
        GroupID:    groupID,
        SubjectID:  subjectID,
        TeacherID:  &teacherID,
        RoomID:     &roomID,
        Type:       lessonType,
        StartsAt:   startsAt,
        EndsAt:     endsAt,
    }
    
    if err := db.Create(lesson).Error; err != nil {
        return nil
    }
    
    return lesson
}

type TestServices struct {
    DB    *gorm.DB
    Redis *redis.Client
    
    UserRepo       *gormrepository.UserRepository
    TokenRepo      *redisrepository.TokenRepository
    GroupRepo      *gormrepository.GroupRepository
    SubjectRepo    *gormrepository.SubjectRepository
    TeacherRepo    *gormrepository.TeacherRepository
    ClassroomRepo  *gormrepository.ClassroomRepository
    LessonRepo     *gormrepository.LessonRepository
    QueueRepo      *gormrepository.QueueRepository
    QueueSlotsRepo *gormrepository.QueueSlotsRepository
    
    AuthUC     *usecase.AuthUseCase
    GroupUC    *usecase.GroupUseCase
    SubjectUC  *usecase.SubjectUseCase
    ScheduleUC *usecase.ScheduleUseCase
    QueueUC    *usecase.QueueUseCase
    UserUC     *usecase.UserUseCase
    AdminUC    *usecase.AdminUseCase
}

func NewTestServices(db *gorm.DB, redisClient *redis.Client) *TestServices {
    userRepo := gormrepository.NewUserRepository(db)
    tokenRepo := redisrepository.NewTokenRepository(redisClient)
    groupRepo := gormrepository.NewGroupRepository(db)
    subjectRepo := gormrepository.NewSubjectRepository(db)
    teacherRepo := gormrepository.NewTeacherRepository(db)
    classroomRepo := gormrepository.NewClassroomRepository(db)
    lessonRepo := gormrepository.NewLessonRepository(db)
    queueRepo := gormrepository.NewQueueRepository(db)
    queueSlotsRepo := gormrepository.NewQueueSlotsRepository(db)
    
    authUC := usecase.NewAuthUseCase(tokenRepo, userRepo, TestAuthConfig)
    groupUC := usecase.NewGroupUseCase(groupRepo)
    subjectUC := usecase.NewSubjectUseCase(subjectRepo)
    scheduleUC := usecase.NewScheduleUseCase(lessonRepo, subjectRepo, groupRepo, teacherRepo, classroomRepo, queueRepo)
    queueUC := usecase.NewQueueUseCase(queueRepo, queueSlotsRepo, userRepo, lessonRepo)
    userUC := usecase.NewUserUseCase(userRepo, queueSlotsRepo, queueRepo)
    adminUC := usecase.NewAdminUseCase(userRepo)
    
    return &TestServices{
        DB:    db,
        Redis: redisClient,
        
        UserRepo:       userRepo,
        TokenRepo:      tokenRepo,
        GroupRepo:      groupRepo,
        SubjectRepo:    subjectRepo,
        TeacherRepo:    teacherRepo,
        ClassroomRepo:  classroomRepo,
        LessonRepo:     lessonRepo,
        QueueRepo:      queueRepo,
        QueueSlotsRepo: queueSlotsRepo,
        
        AuthUC:     authUC,
        GroupUC:    groupUC,
        SubjectUC:  subjectUC,
        ScheduleUC: scheduleUC,
        QueueUC:    queueUC,
        UserUC:     userUC,
        AdminUC:    adminUC,
    }
}

func SetupTestServices(t *testing.T) (*TestServices, func()) {
    t.Helper()
    
    db, redisClient, cleanup := SetupTestDB(t)
    services := NewTestServices(db, redisClient)
    
    return services, cleanup
}

func WithTestContext(t *testing.T) (context.Context, context.CancelFunc) {
    t.Helper()
    return context.WithTimeout(context.Background(), 30*time.Second)
}