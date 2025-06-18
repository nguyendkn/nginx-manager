package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrUnauthorized       = errors.New("unauthorized access")
)

// AuthService handles authentication and authorization
type AuthService struct {
	db        *gorm.DB
	jwtSecret string
}

// NewAuthService creates a new auth service instance
func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{
		db:        database.GetDB(),
		jwtSecret: jwtSecret,
	}
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID uint               `json:"user_id"`
	Email  string             `json:"email"`
	Roles  models.StringArray `json:"roles"`
	jwt.RegisteredClaims
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User         *models.User `json:"user"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login authenticates user and returns JWT tokens
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrInvalidCredentials
		}
		logger.Error("Database error during login", logger.Err(err))
		return nil, err
	}

	// Check if user is disabled
	if user.IsDisabled {
		return nil, ErrUserDisabled
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	// Update last login time
	user.UpdateLastLogin()
	if err := s.db.Save(&user).Error; err != nil {
		logger.Warn("Failed to update last login time", logger.Err(err))
	}

	// Generate JWT tokens
	token, err := s.GenerateToken(&user, 24*time.Hour) // 24 hours
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(&user, 7*24*time.Hour) // 7 days
	if err != nil {
		return nil, err
	}

	// Clear sensitive data
	user.Password = ""

	return &LoginResponse{
		User:         &user,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

// RefreshToken refreshes JWT token using refresh token
func (s *AuthService) RefreshToken(req *RefreshTokenRequest) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Find user
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", claims.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is disabled
	if user.IsDisabled {
		return nil, ErrUserDisabled
	}

	// Generate new tokens
	token, err := s.GenerateToken(&user, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(&user, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	// Clear sensitive data
	user.Password = ""

	return &LoginResponse{
		User:         &user,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

// GenerateToken generates a JWT access token
func (s *AuthService) GenerateToken(user *models.User, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "nginx-manager",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GenerateRefreshToken generates a JWT refresh token
func (s *AuthService) GenerateRefreshToken(user *models.User, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "nginx-manager-refresh",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken validates JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// GetCurrentUser gets user from token
func (s *AuthService) GetCurrentUser(tokenString string) (*models.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", claims.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if user.IsDisabled {
		return nil, ErrUserDisabled
	}

	// Clear sensitive data
	user.Password = ""
	return &user, nil
}

// HasPermission checks if user has specific permission
func (s *AuthService) HasPermission(userID uint, permission string) bool {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return false
	}

	// Admin has all permissions
	if user.IsAdmin() {
		return true
	}

	// Add specific permission checking logic here
	// For now, regular users have basic permissions
	basicPermissions := []string{
		"proxy_hosts:read",
		"certificates:read",
		"settings:read",
	}

	for _, p := range basicPermissions {
		if p == permission {
			return true
		}
	}

	return false
}

// RequireAdmin checks if user has admin role
func (s *AuthService) RequireAdmin(userID uint) error {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return ErrUserNotFound
	}

	if !user.IsAdmin() {
		return ErrUnauthorized
	}

	return nil
}

// CanManageResource checks if user can manage a specific resource
func (s *AuthService) CanManageResource(userID uint, resourceUserID uint) bool {
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return false
	}

	// Admin can manage all resources
	if user.IsAdmin() {
		return true
	}

	// Users can only manage their own resources
	return userID == resourceUserID
}

// Logout invalidates user session (if using token blacklist)
func (s *AuthService) Logout(tokenString string) error {
	// In a production system, you might want to implement token blacklisting
	// For now, we'll just validate the token
	_, err := s.ValidateToken(tokenString)
	return err
}
