package integration

import (
    "context"
    "fmt"
    "os"
    "sync"
    "testing"
    "time"

    "stud_hub/internal/repository/gorm-models"

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

var DefaultTestDBConfig = TestDBConfig{
    Host:     "localhost",
    Port:     "5432",
    User:     "test_user",
    Password: "test_password",
    DBName:   "stud_hub_test",
    SSLMode:  "disable",
}

var (
    globalDB     *gorm.DB
    setupOnce    sync.Once
    setupErr     error
    migrationSQL []byte
)

func LoadConfigFromEnv() TestDBConfig {
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

func initGlobalDB() error {
    config := LoadConfigFromEnv()
    
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

func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
    t.Helper()
    
    setupOnce.Do(func() {
        setupErr = initGlobalDB()
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
    }
    
    return globalDB, cleanup
}

func GetTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    
    setupOnce.Do(func() {
        setupErr = initGlobalDB()
    })
    
    if setupErr != nil {
        t.Fatalf("Failed to setup test DB: %v", setupErr)
    }
    
    return globalDB
}

func ClearTables(db *gorm.DB, tables ...string) error {
    for _, table := range tables {
        if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
            return err
        }
    }
    return nil
}

func ClearAllTables(db *gorm.DB) error {
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
    
    db.Exec("SET CONSTRAINTS ALL DEFERRED")
    
    for _, table := range tables {
        if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
            return err
        }
    }
    
    return nil
}

func BeginTransaction(t *testing.T, db *gorm.DB) (*gorm.DB, func()) {
    t.Helper()
    
    tx := db.Begin()
    if tx.Error != nil {
        t.Fatalf("Failed to begin transaction: %v", tx.Error)
    }
    
    cleanup := func() {
        tx.Rollback()
    }
    
    return tx, cleanup
}

func CreateTestUser(db *gorm.DB, email, firstName, lastName string) *gormmodels.User {
    user := &gormmodels.User{
        Email:     email,
        FirstName: firstName,
        LastName:  lastName,
    }
    
    if err := db.Create(user).Error; err != nil {
        return nil
    }
    
    return user
}

func CreateTestGroup(db *gorm.DB, name string) *gormmodels.Group {
    group := &gormmodels.Group{
        Name: name,
    }
    
    if err := db.Create(group).Error; err != nil {
        return nil
    }
    
    return group
}

func CreateTestSubject(db *gorm.DB, name string) *gormmodels.Subject {
    subject := &gormmodels.Subject{
        Name: name,
    }
    
    if err := db.Create(subject).Error; err != nil {
        return nil
    }
    
    return subject
}

func CreateTestTeacher(db *gorm.DB, firstName, lastName string) *gormmodels.Teacher {
    teacher := &gormmodels.Teacher{
        FirstName: firstName,
        LastName:  lastName,
    }
    
    if err := db.Create(teacher).Error; err != nil {
        return nil
    }
    
    return teacher
}

func WithTestContext(t *testing.T) (context.Context, context.CancelFunc) {
    t.Helper()
    return context.WithTimeout(context.Background(), 10*time.Second)
}

func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
    t.Helper()
    if err != nil {
        if len(msgAndArgs) > 0 {
            t.Fatalf("Expected no error, got %v: %v", err, msgAndArgs[0])
        } else {
            t.Fatalf("Expected no error, got %v", err)
        }
    }
}

func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
    t.Helper()
    if err == nil {
        if len(msgAndArgs) > 0 {
            t.Fatalf("Expected error, got nil: %v", msgAndArgs[0])
        } else {
            t.Fatalf("Expected error, got nil")
        }
    }
}