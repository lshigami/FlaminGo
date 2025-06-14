package integrationtest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	config_pkg "queue_system/config"
	"queue_system/database"
	"queue_system/internal/controller"
	"queue_system/internal/model"
	"queue_system/internal/repository"
	"queue_system/internal/service"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"gorm.io/gorm"
)

type TestApp struct {
	DB     *gorm.DB
	Router *gin.Engine
	Config *config_pkg.Config
}

var globalTestApp *TestApp

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..")
	return projectRoot
}

func SetupTestApp(tLogger MinimalLogger) (*TestApp, error) {
	projectRoot := getProjectRoot()

	viper.SetConfigName(".env.test")
	viper.SetConfigType("env")
	viper.AddConfigPath(projectRoot)
	viper.AutomaticEnv()

	errReadTest := viper.ReadInConfig()

	if errReadTest != nil {
		tLogger.Logf("Warning: .env.test file not found or error reading it from %s: %v. Trying .env.", projectRoot, errReadTest)
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
		errReadEnv := viper.ReadInConfig()

		if errReadEnv != nil {
			return nil, fmt.Errorf("failed to read .env.test and .env files from %s. Last error for .env: %w", projectRoot, errReadEnv)
		}
		tLogger.Logf("Successfully read .env file from %s.", projectRoot)
	} else {
		tLogger.Logf("Successfully read .env.test file from %s.", projectRoot)
	}

	cfg, err := config_pkg.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to process config values after viper read: %w", err)
	}

	//override cfg.Database.Name
	testDBName := viper.GetString("DATABASE_NAME_TEST")
	cfg.Database.Name = testDBName

	//Make sure DATABASE_NAME_TEST is exists
	defaultDbConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password)
	sqlDb, err := sql.Open("pgx", defaultDbConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default 'postgres' database to check/create test DB: %w", err)
	}
	defer sqlDb.Close()

	var exists bool
	checkDbExistsQuery := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '%s')", testDBName)
	err = sqlDb.QueryRow(checkDbExistsQuery).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database '%s' exists: %w", testDBName, err)
	}

	if !exists {
		tLogger.Logf("Database '%s' does not exist. Attempting to create it...", testDBName)
		_, err = sqlDb.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
		if err != nil {
			return nil, fmt.Errorf("failed to create database '%s': %w", testDBName, err)
		}
		tLogger.Logf("Database '%s' created successfully.", testDBName)
	} else {
		tLogger.Logf("Database '%s' already exists.", testDBName)
	}

	// Set up test
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database '%s' with GORM: %w", testDBName, err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Appointment{})
	if err != nil {
		gormSQLDB, _ := db.DB()
		if gormSQLDB != nil {
			gormSQLDB.Close()
		}
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userCtrl := controller.NewUserController(userSvc)

	apptRepo := repository.NewAppointmentRepository(db)
	apptSvc := service.NewAppointmentService(apptRepo, userRepo, db)
	apptCtrl := controller.NewAppointmentController(apptSvc)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	apiV1 := router.Group("/api/v1")
	userRoutes := apiV1.Group("/users")
	{
		userRoutes.POST("", userCtrl.CreateUser)
		userRoutes.GET("/:id", userCtrl.GetUserById)
	}
	apptRoutes := apiV1.Group("/appointments")
	{
		apptRoutes.POST("", apptCtrl.CreateAppointment)
		apptRoutes.GET("/:id", apptCtrl.GetAppointmentByID)
	}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return &TestApp{
		DB:     db,
		Router: router,
		Config: cfg,
	}, nil
}

type MinimalLogger interface {
	Logf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func ClearTables(t *testing.T, db *gorm.DB, tables ...interface{}) {
	for _, table := range tables {
		err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error
		if err != nil {
			t.Fatalf("Failed to clear table for model %T: %v", table, err)
		}
	}
}

func MakeRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBodyBytes []byte
	var err error

	if body != nil {
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		t.Fatalf("Failed to create new HTTP request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func CheckTestEnv(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests: RUN_INTEGRATION_TESTS env var not set")
	}
}

func CreateUserInDB(t *testing.T, db *gorm.DB, user *model.User) *model.User {
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("Failed to create user for test setup: %v", err)
	}
	return user
}
