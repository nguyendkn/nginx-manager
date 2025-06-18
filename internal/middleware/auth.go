package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "nginx-manager-secret" // Default secret for development
	}

	authService := services.NewAuthService(jwtSecret)

	return gin.HandlerFunc(func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			response.ErrorJSON(c, http.StatusUnauthorized, "Authorization token required", nil)
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			response.ErrorJSON(c, http.StatusUnauthorized, "Invalid token", err)
			c.Abort()
			return
		}

		// Get current user
		user, err := authService.GetCurrentUser(token)
		if err != nil {
			response.ErrorJSON(c, http.StatusUnauthorized, "Invalid user", err)
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)
		c.Set("auth_service", authService)

		c.Next()
	})
}

// OptionalAuthMiddleware creates optional JWT authentication middleware
func OptionalAuthMiddleware() gin.HandlerFunc {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "nginx-manager-secret"
	}

	authService := services.NewAuthService(jwtSecret)

	return gin.HandlerFunc(func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token != "" {
			claims, err := authService.ValidateToken(token)
			if err == nil {
				user, err := authService.GetCurrentUser(token)
				if err == nil {
					c.Set("user", user)
					c.Set("user_id", claims.UserID)
					c.Set("user_email", claims.Email)
					c.Set("user_roles", claims.Roles)
				}
			}
		}

		c.Set("auth_service", authService)
		c.Next()
	})
}

// AdminOnlyMiddleware ensures only admin users can access the endpoint
func AdminOnlyMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authService, exists := c.Get("auth_service")
		if !exists {
			response.ErrorJSON(c, http.StatusInternalServerError, "Auth service not available", nil)
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorJSON(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		auth := authService.(*services.AuthService)
		if err := auth.RequireAdmin(userID.(uint)); err != nil {
			response.ErrorJSON(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequirePermissionMiddleware checks if user has specific permission
func RequirePermissionMiddleware(permission string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authService, exists := c.Get("auth_service")
		if !exists {
			response.ErrorJSON(c, http.StatusInternalServerError, "Auth service not available", nil)
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorJSON(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		auth := authService.(*services.AuthService)
		if !auth.HasPermission(userID.(uint), permission) {
			response.ErrorJSON(c, http.StatusForbidden, "Insufficient permissions", fmt.Errorf("required permission: %s", permission))
			c.Abort()
			return
		}

		c.Next()
	})
}

// ResourceOwnerMiddleware checks if user owns the resource or is admin
func ResourceOwnerMiddleware(resourceUserIDParam string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authService, exists := c.Get("auth_service")
		if !exists {
			response.ErrorJSON(c, http.StatusInternalServerError, "Auth service not available", nil)
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			response.ErrorJSON(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		// Get resource user ID from URL parameter
		resourceUserIDStr := c.Param(resourceUserIDParam)
		if resourceUserIDStr == "" {
			response.ErrorJSON(c, http.StatusBadRequest, "Resource user ID required", nil)
			c.Abort()
			return
		}

		// Parse resource user ID
		var resourceUserID uint
		if _, err := fmt.Sscanf(resourceUserIDStr, "%d", &resourceUserID); err != nil {
			response.ErrorJSON(c, http.StatusBadRequest, "Invalid resource user ID", err)
			c.Abort()
			return
		}

		auth := authService.(*services.AuthService)
		if !auth.CanManageResource(userID.(uint), resourceUserID) {
			response.ErrorJSON(c, http.StatusForbidden, "Cannot access this resource", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// extractTokenFromHeader extracts JWT token from Authorization header
func extractTokenFromHeader(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if bearerToken == "" {
		return ""
	}

	if strings.HasPrefix(bearerToken, "Bearer ") {
		return strings.TrimPrefix(bearerToken, "Bearer ")
	}

	return bearerToken
}

// GetCurrentUser gets current user from context
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	u, ok := user.(*models.User)
	return u, ok
}

// GetCurrentUserID gets current user ID from context
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetAuthService gets auth service from context
func GetAuthService(c *gin.Context) (*services.AuthService, bool) {
	authService, exists := c.Get("auth_service")
	if !exists {
		return nil, false
	}

	auth, ok := authService.(*services.AuthService)
	return auth, ok
}
