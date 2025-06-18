package controllers

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/middleware"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// AuthController handles authentication endpoints
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController() *AuthController {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "nginx-manager-secret"
	}

	return &AuthController{
		authService: services.NewAuthService(jwtSecret),
	}
}

// Login authenticates user and returns JWT token
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body services.LoginRequest true "Login credentials"
// @Success 200 {object} services.LoginResponse "Login successful"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	loginResponse, err := ac.authService.Login(&req)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			response.UnauthorizedJSONWithLog(c, "Invalid email or password")
		case services.ErrUserDisabled:
			response.UnauthorizedJSONWithLog(c, "User account is disabled")
		default:
			response.InternalServerErrorJSONWithLog(c, "Login failed", err)
		}
		return
	}

	response.SuccessJSONWithLog(c, loginResponse, "Login successful")
}

// RefreshToken refreshes JWT token
// @Summary Refresh token
// @Description Refresh JWT token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body services.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} services.LoginResponse "Token refreshed"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Invalid token"
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req services.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	loginResponse, err := ac.authService.RefreshToken(&req)
	if err != nil {
		switch err {
		case services.ErrTokenInvalid, services.ErrTokenExpired:
			response.UnauthorizedJSONWithLog(c, "Invalid or expired refresh token")
		case services.ErrUserNotFound:
			response.UnauthorizedJSONWithLog(c, "User not found")
		case services.ErrUserDisabled:
			response.UnauthorizedJSONWithLog(c, "User account is disabled")
		default:
			response.InternalServerErrorJSONWithLog(c, "Token refresh failed", err)
		}
		return
	}

	response.SuccessJSONWithLog(c, loginResponse, "Token refreshed successfully")
}

// Logout logs out user (invalidates token)
// @Summary Logout user
// @Description Logout user and invalidate token
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "Logout successful"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		response.UnauthorizedJSONWithLog(c, "Authorization token required")
		return
	}

	if err := ac.authService.Logout(token); err != nil {
		response.UnauthorizedJSONWithLog(c, "Invalid token")
		return
	}

	response.SuccessJSONWithLog(c, nil, "Logout successful")
}

// GetProfile gets current user profile
// @Summary Get user profile
// @Description Get current authenticated user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User "User profile"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/auth/profile [get]
func (ac *AuthController) GetProfile(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	response.SuccessJSONWithLog(c, user, "User profile retrieved")
}

// UpdateProfile updates current user profile
// @Summary Update user profile
// @Description Update current authenticated user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body UpdateProfileRequest true "Profile data"
// @Success 200 {object} models.User "Updated profile"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/auth/profile [put]
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	// Here you would implement profile update logic
	// For now, we'll just return success
	profileData := map[string]interface{}{
		"user_id": userID,
		"message": "Profile update functionality to be implemented",
	}
	response.SuccessJSONWithLog(c, profileData, "Profile updated successfully")
}

// ChangePassword changes user password
// @Summary Change password
// @Description Change current user password
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body ChangePasswordRequest true "Password data"
// @Success 200 {object} response.Response "Password changed"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/auth/change-password [post]
func (ac *AuthController) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	// Here you would implement password change logic
	// For now, we'll just return success
	passwordData := map[string]interface{}{
		"user_id": userID,
		"message": "Password change functionality to be implemented",
	}
	response.SuccessJSONWithLog(c, passwordData, "Password changed successfully")
}

// ValidateToken validates JWT token
// @Summary Validate token
// @Description Validate JWT token and return user info
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "Token valid"
// @Failure 401 {object} response.Response "Invalid token"
// @Router /api/v1/auth/validate [post]
func (ac *AuthController) ValidateToken(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "Invalid token")
		return
	}

	tokenData := map[string]interface{}{
		"user":  user,
		"valid": true,
	}
	response.SuccessJSONWithLog(c, tokenData, "Token is valid")
}

// Request structs
type UpdateProfileRequest struct {
	Name     string `json:"name" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}
