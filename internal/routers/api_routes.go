package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/controllers"
	"github.com/nguyendkn/nginx-manager/internal/middleware"
	"github.com/nguyendkn/nginx-manager/internal/services"
)

// ServiceContainer holds all the initialized services
type ServiceContainer struct {
	AuthService         *services.AuthService
	CertificateService  *services.CertificateService
	MonitoringService   *services.MonitoringService
	AnalyticsService    *services.AnalyticsService
	NotificationService *services.NotificationService
	ConfigService       *services.ConfigService
	TemplateService     *services.TemplateService
	AccessListService   *services.AccessListService
	NginxService        *services.NginxService
}

// SetupAPIRoutes sets up all API routes with middleware (backward compatibility)
func SetupAPIRoutes(r *gin.Engine) {
	// Apply global rate limiting
	r.Use(middleware.GeneralRateLimitMiddleware())

	// API v1 group
	v1 := r.Group("/api/v1")

	// Setup auth routes
	setupAuthRoutes(v1)

	// Setup protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		setupUserRoutes(protected)
		setupProxyHostRoutes(protected, nil)
		setupCertificateRoutes(protected, nil)
		setupMonitoringRoutes(protected, nil)
		setupSettingsRoutes(protected)
		setupNginxConfigRoutes(protected, nil)
		setupTemplateRoutes(protected, nil)
		setupAnalyticsRoutes(protected, nil)
	}

	// Setup admin routes (require admin role)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminOnlyMiddleware())
	{
		setupAdminRoutes(admin)
	}
}

// SetupAPIRoutesWithServices sets up all API routes with proper service injection
func SetupAPIRoutesWithServices(r *gin.Engine, services *ServiceContainer) {
	// Apply global rate limiting
	r.Use(middleware.GeneralRateLimitMiddleware())

	// API v1 group
	v1 := r.Group("/api/v1")

	// Setup auth routes
	setupAuthRoutes(v1)

	// Setup protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		setupUserRoutes(protected)
		setupProxyHostRoutes(protected, nil)
		setupCertificateRoutes(protected, services.CertificateService)
		setupMonitoringRoutes(protected, services.MonitoringService)
		setupSettingsRoutes(protected)
		setupNginxConfigRoutes(protected, services.ConfigService)
		setupTemplateRoutes(protected, services.TemplateService)
		setupAnalyticsRoutes(protected, services.AnalyticsService)
	}

	// Setup admin routes (require admin role)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminOnlyMiddleware())
	{
		setupAdminRoutes(admin)
	}
}

// setupAuthRoutes sets up authentication routes
func setupAuthRoutes(rg *gin.RouterGroup) {
	authController := controllers.NewAuthController()

	auth := rg.Group("/auth")
	// Apply stricter rate limiting for auth endpoints
	auth.Use(middleware.AuthRateLimitMiddleware())
	{
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)

		// Protected auth routes
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/logout", authController.Logout)
			protected.GET("/profile", authController.GetProfile)
			protected.PUT("/profile", authController.UpdateProfile)
			protected.POST("/change-password", authController.ChangePassword)
			protected.POST("/validate", authController.ValidateToken)
		}
	}
}

// setupUserRoutes sets up user management routes
func setupUserRoutes(rg *gin.RouterGroup) {
	// User routes will be implemented later
	users := rg.Group("/users")
	{
		users.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "List users - to be implemented"})
		})
		users.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get user - to be implemented"})
		})
		users.PUT("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Update user - to be implemented"})
		})
		users.DELETE("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Delete user - to be implemented"})
		})
	}
}

// setupProxyHostRoutes sets up proxy host management routes
func setupProxyHostRoutes(rg *gin.RouterGroup, service interface{}) {
	proxyHostController := controllers.NewProxyHostController(nil)

	proxyHosts := rg.Group("/proxy-hosts")
	{
		proxyHosts.GET("", proxyHostController.List)
		proxyHosts.POST("", proxyHostController.Create)
		proxyHosts.GET("/:id", proxyHostController.Get)
		proxyHosts.PUT("/:id", proxyHostController.Update)
		proxyHosts.DELETE("/:id", proxyHostController.Delete)
		proxyHosts.POST("/:id/toggle", proxyHostController.Toggle)
		proxyHosts.POST("/bulk-toggle", proxyHostController.BulkToggle)
	}
}

// setupCertificateRoutes sets up certificate management routes
func setupCertificateRoutes(rg *gin.RouterGroup, service *services.CertificateService) {
	certificateController := controllers.NewCertificateController(service)

	certificates := rg.Group("/certificates")
	{
		certificates.GET("", certificateController.ListCertificates)
		certificates.POST("", certificateController.CreateCertificate)
		certificates.GET("/expiring-soon", certificateController.GetExpiringSoon)
		certificates.POST("/test", certificateController.TestCertificate)
		certificates.GET("/:id", certificateController.GetCertificate)
		certificates.PUT("/:id", certificateController.UpdateCertificate)
		certificates.DELETE("/:id", certificateController.DeleteCertificate)
		certificates.POST("/:id/upload", certificateController.UploadCertificate)
		certificates.POST("/:id/renew", certificateController.RenewCertificate)
	}
}

// setupMonitoringRoutes sets up monitoring and real-time metrics routes
func setupMonitoringRoutes(rg *gin.RouterGroup, service *services.MonitoringService) {
	monitoringController := controllers.NewMonitoringController(service)

	monitoring := rg.Group("/monitoring")
	{
		monitoring.GET("/dashboard", monitoringController.GetDashboardStats)
		monitoring.GET("/system-metrics", monitoringController.GetSystemMetrics)
		monitoring.GET("/nginx-status", monitoringController.GetNginxStatus)
		monitoring.GET("/activity-feed", monitoringController.GetActivityFeed)
		monitoring.GET("/ws", monitoringController.HandleWebSocket)
		monitoring.POST("/nginx/control", monitoringController.ControlNginx)
	}
}

// setupSettingsRoutes sets up settings management routes
func setupSettingsRoutes(rg *gin.RouterGroup) {
	// Settings routes will be implemented later
	settings := rg.Group("/settings")
	{
		settings.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get settings - to be implemented"})
		})
		settings.PUT("", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Update settings - to be implemented"})
		})
		settings.GET("/:key", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get setting by key - to be implemented"})
		})
		settings.PUT("/:key", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Update setting by key - to be implemented"})
		})
	}
}

// setupAdminRoutes sets up admin-only routes
func setupAdminRoutes(rg *gin.RouterGroup) {
	// System administration routes
	rg.GET("/system/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "System health - to be implemented"})
	})

	rg.GET("/system/stats", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "System statistics - to be implemented"})
	})

	// User management for admins
	rg.GET("/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin: List all users - to be implemented"})
	})

	rg.POST("/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin: Create user - to be implemented"})
	})

	// System logs
	rg.GET("/logs", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin: Get system logs - to be implemented"})
	})

	// Certificate management for all users
	rg.GET("/certificates", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin: List all certificates - to be implemented"})
	})

	// Proxy host management for all users
	rg.GET("/proxy-hosts", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin: List all proxy hosts - to be implemented"})
	})

	// Nginx configuration management
	nginx := rg.Group("/nginx")
	{
		nginx.POST("/reload", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin: Reload nginx - to be implemented"})
		})

		nginx.GET("/config", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin: Get nginx config - to be implemented"})
		})
	}
}

// setupNginxConfigRoutes sets up nginx configuration management routes
func setupNginxConfigRoutes(rg *gin.RouterGroup, service *services.ConfigService) {
	configController := controllers.NewConfigController(service)

	configs := rg.Group("/nginx/configs")
	{
		configs.GET("", configController.ListConfigs)
		configs.POST("", configController.CreateConfig)
		configs.GET("/:id", configController.GetConfig)
		configs.PUT("/:id", configController.UpdateConfig)
		configs.DELETE("/:id", configController.DeleteConfig)
		configs.POST("/validate", configController.ValidateConfig)
		configs.POST("/:id/deploy", configController.DeployConfig)
		configs.GET("/:id/history", configController.GetConfigHistory)
		configs.POST("/:id/backup", configController.CreateConfigBackup)
		configs.POST("/:id/restore/:version", configController.RestoreConfigFromBackup)
	}
}

// setupTemplateRoutes sets up configuration template management routes
func setupTemplateRoutes(rg *gin.RouterGroup, service *services.TemplateService) {
	templateController := controllers.NewTemplateController(service)

	templates := rg.Group("/nginx/templates")
	{
		templates.GET("", templateController.ListTemplates)
		templates.POST("", templateController.CreateTemplate)
		templates.GET("/categories", templateController.GetCategories)
		templates.POST("/init-builtin", templateController.InitializeBuiltInTemplates)
		templates.GET("/:id", templateController.GetTemplate)
		templates.PUT("/:id", templateController.UpdateTemplate)
		templates.DELETE("/:id", templateController.DeleteTemplate)
		templates.POST("/:id/render", templateController.RenderTemplate)
	}
}

// setupAnalyticsRoutes sets up analytics and monitoring routes
func setupAnalyticsRoutes(rg *gin.RouterGroup, service *services.AnalyticsService) {
	analyticsController := controllers.NewAnalyticsController(service)

	analytics := rg.Group("/analytics")
	{
		// Historical Metrics Routes
		metricsGroup := analytics.Group("/metrics")
		{
			metricsGroup.POST("/query", analyticsController.QueryMetrics)
			metricsGroup.GET("/:type/:name", analyticsController.GetHistoricalMetrics)
		}

		// System Analytics Routes
		systemGroup := analytics.Group("/system")
		{
			systemGroup.GET("/summary", analyticsController.GetSystemMetricsSummary)
		}

		// Alert Management Routes
		alertsGroup := analytics.Group("/alerts")
		{
			// Alert Rules
			rulesGroup := alertsGroup.Group("/rules")
			{
				rulesGroup.POST("", analyticsController.CreateAlertRule)
				rulesGroup.GET("", analyticsController.GetAlertRules)
				rulesGroup.PUT("/:id", analyticsController.UpdateAlertRule)
				rulesGroup.DELETE("/:id", analyticsController.DeleteAlertRule)
			}

			// Alert Instances
			alertsGroup.GET("/instances", analyticsController.GetAlertInstances)
		}

		// Dashboard Management Routes
		dashboardsGroup := analytics.Group("/dashboards")
		{
			dashboardsGroup.POST("", analyticsController.CreateDashboard)
			dashboardsGroup.GET("", analyticsController.GetDashboards)
			dashboardsGroup.GET("/:id", analyticsController.GetDashboard)
			dashboardsGroup.PUT("/:id", analyticsController.UpdateDashboard)
			dashboardsGroup.DELETE("/:id", analyticsController.DeleteDashboard)
		}
	}
}
