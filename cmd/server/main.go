package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/configs"
	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/middleware"
	"github.com/nguyendkn/nginx-manager/internal/routers"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
)

func main() {
	// Load environment configuration
	env := configs.LoadEnvironment()

	// Initialize logger
	loggerConfig := logger.ConfigFromEnv()
	if err := logger.Initialize(loggerConfig); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Log application startup
	logger.Info("Starting application",
		logger.String("app_name", env.GetAppName()),
		logger.String("app_version", env.GetAppVersion()),
		logger.String("environment", env.GetAppEnvironment()),
		logger.String("log_level", env.GetLogLevel()),
		logger.String("log_encoding", env.GetLogEncoding()),
		logger.String("gin_mode", env.GetGinMode()),
	)

	// Initialize Database
	if err := initializeDatabase(); err != nil {
		logger.Fatal("Failed to initialize database", logger.Err(err))
	}

	// Create Gin router
	r := setupRouter(env)

	// Get port from environment config
	port := env.GetPort()

	logger.Info("Server starting",
		logger.String("port", port),
		logger.String("host", env.GetHost()),
		logger.String("address", env.GetServerAddress()),
	)

	// Start server
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", logger.Err(err))
	}
}

func initializeDatabase() error {
	logger.Info("Initializing database...")

	// Load database configuration
	dbConfig := database.LoadDatabaseConfig()

	// Initialize database connection
	if err := database.InitDatabase(dbConfig); err != nil {
		return err
	}

	// Get database instance
	db := database.GetDB()

	// Run auto-migration
	if err := database.AutoMigrate(db); err != nil {
		return err
	}

	// Seed initial data
	if err := database.SeedData(db); err != nil {
		return err
	}

	// Check database health
	if err := database.CheckDatabaseHealth(db); err != nil {
		return err
	}

	logger.Info("Database initialized successfully")
	return nil
}

func setupRouter(env *configs.Environment) *gin.Engine {
	// Create Gin router without default middleware
	r := gin.New()

	// Add custom middleware
	r.Use(logger.RequestIDMiddleware())
	r.Use(logger.GinLogger())
	r.Use(logger.ErrorLogger())
	r.Use(logger.RecoveryLogger())

	// Add CORS middleware with environment configuration
	r.Use(middleware.CORSMiddleware(env))

	// Setup health routes
	routers.SetupHealthRoutes(r, env)

	// Setup API routes
	routers.SetupAPIRoutes(r)

	return r
}
