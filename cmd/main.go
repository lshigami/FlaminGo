package main

import (
	"context"
	"net/http"
	"queue_system/config"
	"queue_system/database"
	"queue_system/internal/controller"
	"queue_system/internal/repository"
	"queue_system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

func main() {

	app := fx.New(
		fx.Provide(
			config.NewConfig,
			database.NewDatabase,
			NewGinEngine,
		),
		fx.Invoke(RegisterRoutesAndStartServer),
		fx.Provide(
			repository.NewUserRepository,
			service.NewUserService,
			controller.NewUserController,
		),
		fx.Provide(
			repository.NewAppointmentRepository,
			service.NewAppointmentService,
			controller.NewAppointmentController,
		),
	)

	// Start the application
	if err := app.Start(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to start application")
	}

	<-app.Done()
}

func NewGinEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	return router
}

func RegisterRoutesAndStartServer(
	cfg *config.Config,
	router *gin.Engine,
	lc fx.Lifecycle,
	userController *controller.UserController,
	appointmentController *controller.AppointmentController,
) {

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	//User routes
	userRoutes := router.Group("/api/v1/users")
	{
		userRoutes.POST("/", userController.CreateUser)
		userRoutes.GET("/:id", userController.GetUserById)
	}

	//Appointment routes
	appointmentRoutes := router.Group("/api/v1/appointments")
	{
		appointmentRoutes.POST("/", appointmentController.CreateAppointment)
		appointmentRoutes.GET("/:id", appointmentController.GetAppointmentByID)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info().Msgf("Starting HTTP server on port %s", cfg.Server.Port)
			go func() {
				log.Info().Msg("Server is about to listen")
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal().Err(err).Msg("Failed to start HTTP server")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("Stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})

}
