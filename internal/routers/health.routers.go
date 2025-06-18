package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/configs"
	"github.com/nguyendkn/nginx-manager/internal/controllers"
)

// SetupHealthRoutes sets up health-related routes
func SetupHealthRoutes(router *gin.Engine, env *configs.Environment) {
	// Create health controller instance
	healthController := controllers.NewHealthController(env)

	// Health check routes
	router.GET("/health", healthController.HealthCheck)
	router.GET("/ping", healthController.Ping)
}

// SetupHealthRoutesWithGroup sets up health-related routes with a route group
func SetupHealthRoutesWithGroup(router *gin.Engine, env *configs.Environment, prefix string) {
	// Create health controller instance
	healthController := controllers.NewHealthController(env)

	// Create route group
	healthGroup := router.Group(prefix)
	{
		healthGroup.GET("/health", healthController.HealthCheck)
		healthGroup.GET("/ping", healthController.Ping)
	}
}
