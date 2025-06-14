package integrationtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"                  // Để xử lý đường dẫn config tốt hơn
	config_pkg "queue_system/config" // Đổi tên import để tránh trùng với biến config
	"queue_system/database"
	"queue_system/internal/controller"
	"queue_system/internal/model"
	"queue_system/internal/repository"
	"queue_system/internal/service"
	"runtime" // Để lấy đường dẫn gốc của project
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestApp holds all the necessary components for running integration tests.
type TestApp struct {
	DB     *gorm.DB
	Router *gin.Engine
	Config *config_pkg.Config // Sử dụng tên import đã đổi
}

// Global instance, initialized by TestMain via SetupTestApp
var globalTestApp *TestApp

// getProjectRoot returns the absolute path to the project root.
// This is a bit of a heuristic but works for typical Go project structures.
func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0) // Get path to current file (helper.go)
	// Project root is expected to be two levels up from test/integrationtest/
	// test/integrationtest -> test -> project_root
	// Hoặc bạn có thể dùng một file marker (ví dụ go.mod) để tìm project root
	// Dưới đây là giả định đi lên 2 thư mục từ thư mục chứa helper.go
	// test/integrationtest -> test -> project_root
	// Hoặc nếu helper.go ở test/ thì là một cấp
	// Nếu helper.go ở integrationtest/ thì là 2 cấp
	// Giả sử helper.go nằm trong integrationtest/
	projectRoot := filepath.Join(filepath.Dir(b), "..", "..")
	return projectRoot
}

// SetupTestApp initializes the application components for testing.
// It's designed to be called once by TestMain.
func SetupTestApp(t *testing.T) *TestApp {
	// Load config for test
	projectRoot := getProjectRoot()
	envTestPath := filepath.Join(projectRoot, ".env.test")
	envPath := filepath.Join(projectRoot, ".env")

	viper.SetConfigName(".env.test")
	viper.AddConfigPath(projectRoot) // Chỉ cần đường dẫn đến thư mục gốc
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		t.Logf("Warning: .env.test file (%s) not found or error reading it: %v. Trying .env.", envTestPath, err)
		viper.SetConfigName(".env") // Expect .env in project root
		// viper.AddConfigPath(projectRoot) // Đã add rồi
		if errReadEnv := viper.ReadInConfig(); errReadEnv != nil {
			t.Fatalf("Failed to read .env file (%s) as well: %v", envPath, errReadEnv)
		}
	}

	testDBName := viper.GetString("DATABASE_NAME_TEST")
	if testDBName == "" {
		originalDBName := viper.GetString("DATABASE_NAME")
		if originalDBName == "" {
			t.Fatal("DATABASE_NAME is not set in any .env file.")
		}
		testDBName = originalDBName + "_test_integration" // Thêm hậu tố rõ ràng
		t.Logf("DATABASE_NAME_TEST not set, using derived name: %s. THIS IS RISKY. Please set DATABASE_NAME_TEST in .env.test.", testDBName)
	}
	viper.Set("DATABASE_NAME", testDBName) // Crucial override

	cfg, err := config_pkg.NewConfig() // Sử dụng tên import đã đổi
	require.NoError(t, err, "Failed to load config for test")

	// Setup database
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err, "Failed to connect to test database: %s", cfg.Database.Name)

	// Migrate schema
	err = db.AutoMigrate(&model.User{}, &model.Appointment{})
	require.NoError(t, err, "Failed to migrate test database")

	// Setup Gin router with real dependencies
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userCtrl := controller.NewUserController(userSvc)

	apptRepo := repository.NewAppointmentRepository(db)
	apptSvc := service.NewAppointmentService(apptRepo, userRepo, db)
	apptCtrl := controller.NewAppointmentController(apptSvc)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Register routes
	apiV1 := router.Group("/api/v1")
	userRoutes := apiV1.Group("/users")
	{
		userRoutes.POST("", userCtrl.CreateUser)
		userRoutes.GET("/:id", userCtrl.GetUserById) // Đã sửa tên hàm
		// Thêm các route khác của User nếu bạn đã implement và muốn test
		// userRoutes.GET("", userCtrl.GetAllUsers)
		// userRoutes.PUT("/:id", userCtrl.UpdateUser)
		// userRoutes.DELETE("/:id", userCtrl.DeleteUser)
	}
	apptRoutes := apiV1.Group("/appointments")
	{
		apptRoutes.POST("", apptCtrl.CreateAppointment)
		apptRoutes.GET("/:id", apptCtrl.GetAppointmentByID)
		// Thêm các route khác của Appointment nếu bạn đã implement và muốn test
		// apptRoutes.GET("/user", apptCtrl.GetAppointmentsByUserID)
		// apptRoutes.GET("/participant", apptCtrl.GetAppointmentsByParticipantID)
		// apptRoutes.PUT("/:id", apptCtrl.UpdateAppointment)
		// apptRoutes.PATCH("/:id/cancel", apptCtrl.CancelAppointment)
	}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return &TestApp{
		DB:     db,
		Router: router,
		Config: cfg,
	}
}

// ClearTables truncates all specified tables to ensure a clean state for each test.
func ClearTables(t *testing.T, db *gorm.DB, tables ...interface{}) {
	for _, table := range tables {
		err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error
		require.NoError(t, err, fmt.Sprintf("Failed to clear table for model %T", table))
	}
}

// MakeRequest is a helper to perform HTTP requests against the test router.
func MakeRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBodyBytes []byte
	var err error

	if body != nil {
		reqBodyBytes, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBodyBytes))
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// CheckTestEnv checks if the integration tests should run.
func CheckTestEnv(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration tests: RUN_INTEGRATION_TESTS env var not set")
	}
}

// CreateUserInDB is a helper to directly create a user in DB for test setup.
func CreateUserInDB(t *testing.T, db *gorm.DB, user *model.User) *model.User {
	err := db.Create(user).Error
	require.NoError(t, err, "Failed to create user for test setup")
	return user
}
