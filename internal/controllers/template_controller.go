package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/errors"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// TemplateController handles configuration template management endpoints
type TemplateController struct {
	templateService *services.TemplateService
}

// NewTemplateController creates a new template controller
func NewTemplateController(templateService *services.TemplateService) *TemplateController {
	return &TemplateController{
		templateService: templateService,
	}
}

// CreateTemplate creates a new configuration template
// @Summary Create configuration template
// @Description Create a new nginx configuration template
// @Tags nginx-templates
// @Accept json
// @Produce json
// @Param template body services.TemplateRequest true "Template data"
// @Success 201 {object} models.ConfigTemplate
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates [post]
func (c *TemplateController) CreateTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var req services.TemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	template, err := c.templateService.CreateTemplate(userID.(uint), &req)
	if err != nil {
		if err == errors.ErrTemplateDuplicate {
			response.ErrorJSONWithLog(ctx, http.StatusConflict, "Template with this name already exists", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Failed to create template", err)
		return
	}

	response.SuccessJSONWithLog(ctx, template, "Template created successfully")
}

// GetTemplate retrieves a configuration template by ID
// @Summary Get configuration template
// @Description Get configuration template details by ID
// @Tags nginx-templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} models.ConfigTemplate
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates/{id} [get]
func (c *TemplateController) GetTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid template ID", err)
		return
	}

	template, err := c.templateService.GetTemplate(userID.(uint), uint(id))
	if err != nil {
		if err == errors.ErrTemplateNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Template not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to get template", err)
		return
	}

	response.SuccessJSONWithLog(ctx, template, "Template retrieved successfully")
}

// ListTemplates retrieves configuration templates with pagination
// @Summary List configuration templates
// @Description Get paginated list of configuration templates
// @Tags nginx-templates
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(10)
// @Param category query string false "Template category filter"
// @Param include_public query bool false "Include public templates" default(true)
// @Success 200 {object} services.TemplateListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates [get]
func (c *TemplateController) ListTemplates(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	category := ctx.Query("category")
	includePublic, _ := strconv.ParseBool(ctx.DefaultQuery("include_public", "true"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	templates, err := c.templateService.ListTemplates(userID.(uint), page, limit, category, includePublic)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to list templates", err)
		return
	}

	response.SuccessJSONWithLog(ctx, templates, "Templates retrieved successfully")
}

// UpdateTemplate updates an existing configuration template
// @Summary Update configuration template
// @Description Update an existing configuration template
// @Tags nginx-templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param template body services.TemplateRequest true "Template data"
// @Success 200 {object} models.ConfigTemplate
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates/{id} [put]
func (c *TemplateController) UpdateTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid template ID", err)
		return
	}

	var req services.TemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	template, err := c.templateService.UpdateTemplate(userID.(uint), uint(id), &req)
	if err != nil {
		if err == errors.ErrTemplateNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Template not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Failed to update template", err)
		return
	}

	response.SuccessJSONWithLog(ctx, template, "Template updated successfully")
}

// DeleteTemplate deletes a configuration template
// @Summary Delete configuration template
// @Description Delete a configuration template by ID
// @Tags nginx-templates
// @Produce json
// @Param id path int true "Template ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates/{id} [delete]
func (c *TemplateController) DeleteTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid template ID", err)
		return
	}

	if err := c.templateService.DeleteTemplate(userID.(uint), uint(id)); err != nil {
		if err == errors.ErrTemplateNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Template not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		if err == errors.ErrTemplateInUse {
			response.ErrorJSONWithLog(ctx, http.StatusConflict, "Template is in use", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Failed to delete template", err)
		return
	}

	response.SuccessJSONWithLog(ctx, gin.H{"id": id}, "Template deleted successfully")
}

// RenderTemplate renders a template with given variables
// @Summary Render configuration template
// @Description Render a template with provided variables
// @Tags nginx-templates
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param render body services.TemplateRenderRequest true "Template variables"
// @Success 200 {object} services.TemplateRenderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates/{id}/render [post]
func (c *TemplateController) RenderTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid template ID", err)
		return
	}

	var req services.TemplateRenderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	result, err := c.templateService.RenderTemplate(userID.(uint), uint(id), &req)
	if err != nil {
		if err == errors.ErrTemplateNotFound {
			response.ErrorJSONWithLog(ctx, http.StatusNotFound, "Template not found", err)
			return
		}
		if err == errors.ErrPermissionDenied {
			response.ErrorJSONWithLog(ctx, http.StatusForbidden, "Permission denied", err)
			return
		}
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to render template", err)
		return
	}

	response.SuccessJSONWithLog(ctx, result, "Template rendered successfully")
}

// GetCategories returns all available template categories
// @Summary Get template categories
// @Description Get list of all available template categories
// @Tags nginx-templates
// @Produce json
// @Success 200 {object} []string
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates/categories [get]
func (c *TemplateController) GetCategories(ctx *gin.Context) {
	_, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	categories := c.templateService.GetCategories()
	response.SuccessJSONWithLog(ctx, categories, "Categories retrieved successfully")
}

// InitializeBuiltInTemplates initializes built-in templates (admin only)
// @Summary Initialize built-in templates
// @Description Initialize default built-in configuration templates (admin only)
// @Tags nginx-templates
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /api/v1/nginx/templates/init-builtin [post]
func (c *TemplateController) InitializeBuiltInTemplates(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.ErrorJSONWithLog(ctx, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// TODO: Add admin check
	// For now, allow any authenticated user
	_ = userID

	if err := c.templateService.CreateBuiltInTemplates(); err != nil {
		response.ErrorJSONWithLog(ctx, http.StatusInternalServerError, "Failed to initialize built-in templates", err)
		return
	}

	response.SuccessJSONWithLog(ctx, gin.H{"message": "Built-in templates initialized"}, "Built-in templates created successfully")
}
