package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/configs"
	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/middleware"
	"github.com/nguyendkn/nginx-manager/internal/routers"
	"github.com/nguyendkn/nginx-manager/internal/services"
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

	// Initialize Services
	serviceContainer := initializeServices()

	// Create Gin router
	r := setupRouter(env, serviceContainer)

	// Start background services
	startBackgroundServices(serviceContainer)

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

func initializeServices() *routers.ServiceContainer {
	logger.Info("Initializing services...")

	db := database.GetDB()

	// Configuration paths (should come from environment)
	nginxConfigPath := "/etc/nginx/nginx.conf"
	sitesPath := "/etc/nginx/sites-available"
	backupPath := "/var/lib/nginx-manager/backups"
	templatePath := "/var/lib/nginx-manager/templates"
	certPath := "/etc/nginx/ssl/certs"
	keyPath := "/etc/nginx/ssl/private"
	jwtSecret := "your-jwt-secret-key" // TODO: Get from environment

	// Initialize core services
	authService := services.NewAuthService(jwtSecret)
	nginxService := services.NewNginxService(nginxConfigPath, sitesPath, backupPath, templatePath, authService)
	notificationService := services.NewNotificationService()

	// Initialize dependent services
	certificateService := services.NewCertificateService(certPath, keyPath, authService)
	accessListService := services.NewAccessListService(authService)
	configService := services.NewConfigService(nginxConfigPath, backupPath, templatePath, authService)
	templateService := services.NewTemplateService(authService)
	monitoringService := services.NewMonitoringService(nginxService)

	// Initialize analytics service (depends on monitoring service)
	analyticsService := services.NewAnalyticsService(db, monitoringService, notificationService)

	logger.Info("Services initialized successfully")

	return &routers.ServiceContainer{
		AuthService:         authService,
		CertificateService:  certificateService,
		MonitoringService:   monitoringService,
		AnalyticsService:    analyticsService,
		NotificationService: notificationService,
		ConfigService:       configService,
		TemplateService:     templateService,
		AccessListService:   accessListService,
		NginxService:        nginxService,
	}
}

func startBackgroundServices(services *routers.ServiceContainer) {
	logger.Info("Starting background services...")

	// Start analytics metrics collection every 5 minutes
	go func() {
		ctx := context.Background()
		services.AnalyticsService.StartMetricsCollection(ctx, 5*time.Minute)
	}()

	// Start metrics cleanup every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if err := services.AnalyticsService.CleanupExpiredMetrics(); err != nil {
				logger.Error("Failed to cleanup expired metrics", logger.Err(err))
			}
		}
	}()

	logger.Info("Background services started")
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

func setupRouter(env *configs.Environment, services *routers.ServiceContainer) *gin.Engine {
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

	// Setup API routes with service injection
	routers.SetupAPIRoutesWithServices(r, services)

	return r
}
