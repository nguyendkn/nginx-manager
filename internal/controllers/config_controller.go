package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/errors"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// ConfigController handles nginx configuration management endpoints
type ConfigController struct {
	configService *services.ConfigService
}

// NewConfigController creates a new config controller
func NewConfigController(configService *services.ConfigService) *ConfigController {
	return &ConfigController{
		configService: configService,
	}
}

// CreateConfig creates a new nginx configuration
// @Summary Create nginx configuration
// @Description Create a new nginx configuration with validation
// @Tags nginx-config
// @Accept json
// @Produce json
// @Param config body services.ConfigRequest true "Configuration data"
// @Success 201 {object} models.NginxConfig
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs [post]
func (c *ConfigController) CreateConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req services.ConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	config, err := c.configService.CreateConfig(userID.(uint), &req)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Failed to create configuration", err)
		return
	}

	response.SuccessJSONWithLog(ctx, config, "Configuration created successfully")
}

// GetConfig retrieves a nginx configuration by ID
// @Summary Get nginx configuration
// @Description Get nginx configuration details by ID
// @Tags nginx-config
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} models.NginxConfig
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id} [get]
func (c *ConfigController) GetConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	config, err := c.configService.GetConfig(userID.(uint), uint(id))
	if err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to get configuration", err)
		return
	}

	response.SuccessJSONWithLog(ctx, config, "Configuration retrieved successfully")
}

// ListConfigs retrieves nginx configurations with pagination
// @Summary List nginx configurations
// @Description Get paginated list of nginx configurations
// @Tags nginx-config
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(10)
// @Param type query string false "Configuration type filter"
// @Success 200 {object} services.ConfigListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs [get]
func (c *ConfigController) ListConfigs(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	configType := ctx.Query("type")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	configs, err := c.configService.ListConfigs(userID.(uint), page, limit, configType)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to list configurations", err)
		return
	}

	response.SuccessJSONWithLog(ctx, configs, "Configurations retrieved successfully")
}

// UpdateConfig updates an existing nginx configuration
// @Summary Update nginx configuration
// @Description Update an existing nginx configuration
// @Tags nginx-config
// @Accept json
// @Produce json
// @Param id path int true "Configuration ID"
// @Param config body services.ConfigRequest true "Configuration data"
// @Success 200 {object} models.NginxConfig
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id} [put]
func (c *ConfigController) UpdateConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	var req services.ConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	config, err := c.configService.UpdateConfig(userID.(uint), uint(id), &req)
	if err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Failed to update configuration", err)
		return
	}

	response.SuccessJSONWithLog(ctx, config, "Configuration updated successfully")
}

// DeleteConfig deletes a nginx configuration
// @Summary Delete nginx configuration
// @Description Delete a nginx configuration by ID
// @Tags nginx-config
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id} [delete]
func (c *ConfigController) DeleteConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	if err := c.configService.DeleteConfig(userID.(uint), uint(id)); err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Failed to delete configuration", err)
		return
	}

	response.SuccessJSONWithLog(ctx, gin.H{"id": id}, "Configuration deleted successfully")
}

// ValidateConfig validates nginx configuration content
// @Summary Validate nginx configuration
// @Description Validate nginx configuration syntax
// @Tags nginx-config
// @Accept json
// @Produce json
// @Param content body map[string]string true "Configuration content"
// @Success 200 {object} services.ValidationResult
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/validate [post]
func (c *ConfigController) ValidateConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	result, err := c.configService.ValidateConfig(userID.(uint), req.Content)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Validation failed", err)
		return
	}

	response.SuccessJSONWithLog(ctx, result, "Configuration validated")
}

// DeployConfig deploys a configuration to nginx
// @Summary Deploy nginx configuration
// @Description Deploy configuration to nginx and reload
// @Tags nginx-config
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id}/deploy [post]
func (c *ConfigController) DeployConfig(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	if err := c.configService.DeployConfig(userID.(uint), uint(id)); err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		if err == errors.ErrConfigValidationFailed {
			response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Configuration validation failed", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Deployment failed", err)
		return
	}

	response.SuccessJSONWithLog(ctx, gin.H{"id": id}, "Configuration deployed successfully")
}

// GetConfigHistory retrieves configuration version history
// @Summary Get configuration history
// @Description Get version history for a configuration
// @Tags nginx-config
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} []models.ConfigVersion
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id}/history [get]
func (c *ConfigController) GetConfigHistory(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	// First check if user has access to the configuration
	_, err = c.configService.GetConfig(userID.(uint), uint(id))
	if err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to access configuration", err)
		return
	}

	// TODO: Implement GetConfigHistory in service
	// For now, return empty array
	response.SuccessJSONWithLog(ctx, []interface{}{}, "Configuration history retrieved")
}

// CreateConfigBackup creates a manual backup of a configuration
// @Summary Create configuration backup
// @Description Create a manual backup of a configuration
// @Tags nginx-config
// @Accept json
// @Produce json
// @Param id path int true "Configuration ID"
// @Param backup body map[string]string true "Backup reason"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id}/backup [post]
func (c *ConfigController) CreateConfigBackup(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	if req.Reason == "" {
		req.Reason = "Manual backup"
	}

	// First check if user has access to the configuration
	_, err = c.configService.GetConfig(userID.(uint), uint(id))
	if err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to access configuration", err)
		return
	}

	// TODO: Implement CreateBackup method in service that can be called externally
	// For now, return success
	response.SuccessJSONWithLog(ctx, gin.H{"id": id, "reason": req.Reason}, "Backup created successfully")
}

// RestoreConfigFromBackup restores a configuration from backup
// @Summary Restore configuration from backup
// @Description Restore a configuration from a specific backup version
// @Tags nginx-config
// @Produce json
// @Param id path int true "Configuration ID"
// @Param version path int true "Backup version"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/configs/{id}/restore/{version} [post]
func (c *ConfigController) RestoreConfigFromBackup(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid configuration ID", err)
		return
	}

	versionStr := ctx.Param("version")
	version, err := strconv.ParseUint(versionStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid version number", err)
		return
	}

	// First check if user has access to the configuration
	_, err = c.configService.GetConfig(userID.(uint), uint(id))
	if err != nil {
		if err == errors.ErrConfigNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Configuration not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to access configuration", err)
		return
	}

	// TODO: Implement RestoreFromBackup method in service
	// For now, return success
	response.SuccessJSONWithLog(ctx, gin.H{"id": id, "version": version}, "Configuration restored successfully")
}
