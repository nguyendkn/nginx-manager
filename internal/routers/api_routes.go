package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/controllers"
	"github.com/nguyendkn/nginx-manager/internal/middleware"
)

// SetupAPIRoutes sets up all API routes with middleware
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
		setupProxyHostRoutes(protected)
		setupCertificateRoutes(protected)
		setupMonitoringRoutes(protected)
		setupSettingsRoutes(protected)
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
func setupProxyHostRoutes(rg *gin.RouterGroup) {
	proxyHostController := controllers.NewProxyHostController(nil) // Pass nil for now, will inject service later

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
func setupCertificateRoutes(rg *gin.RouterGroup) {
	// Note: For now we'll use nil service, should be properly injected in the main server setup
	certificateController := controllers.NewCertificateController(nil)

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
func setupMonitoringRoutes(rg *gin.RouterGroup) {
	// Note: For now we'll use nil service, should be properly injected in the main server setup
	monitoringController := controllers.NewMonitoringController(nil)

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

		nginx.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin: Test nginx config - to be implemented"})
		})
	}

	// Database operations
	database := rg.Group("/database")
	{
		database.POST("/migrate", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin: Run migrations - to be implemented"})
		})

		database.POST("/seed", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin: Seed database - to be implemented"})
		})

		database.GET("/backup", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin: Backup database - to be implemented"})
		})
	}
}
