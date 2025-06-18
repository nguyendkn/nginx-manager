package services

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/errors"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"gorm.io/gorm"
)

// TemplateService handles configuration template management
type TemplateService struct {
	db          *gorm.DB
	authService *AuthService
}

// NewTemplateService creates a new template service instance
func NewTemplateService(authService *AuthService) *TemplateService {
	return &TemplateService{
		db:          database.GetDB(),
		authService: authService,
	}
}

// TemplateRequest represents template create/update request
type TemplateRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Category    models.TemplateCategory `json:"category" binding:"required"`
	Content     string                  `json:"content" binding:"required"`
	Variables   map[string]interface{}  `json:"variables"`
	IsPublic    bool                    `json:"is_public"`
}

// TemplateListResponse represents paginated template list
type TemplateListResponse struct {
	Templates []models.ConfigTemplate `json:"templates"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	Limit     int                     `json:"limit"`
}

// TemplateRenderRequest represents template render request
type TemplateRenderRequest struct {
	Variables map[string]interface{} `json:"variables" binding:"required"`
}

// TemplateRenderResponse represents template render response
type TemplateRenderResponse struct {
	Content string   `json:"content"`
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors,omitempty"`
}

// CreateTemplate creates a new configuration template
func (s *TemplateService) CreateTemplate(userID uint, req *TemplateRequest) (*models.ConfigTemplate, error) {
	// Validate category
	if !req.Category.IsValid() {
		return nil, fmt.Errorf("invalid template category")
	}

	// Check for duplicate template name for user
	var existing models.ConfigTemplate
	err := s.db.Where("name = ? AND user_id = ?", req.Name, userID).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		return nil, errors.ErrTemplateDuplicate
	}

	// Validate template syntax
	if err := s.validateTemplate(req.Content); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Create template model
	tmpl := &models.ConfigTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Content:     req.Content,
		Variables:   models.JSON(req.Variables),
		IsBuiltIn:   false,
		IsPublic:    req.IsPublic,
		UsageCount:  0,
		UserID:      userID,
	}

	// Save to database
	if err := s.db.Create(tmpl).Error; err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeConfigTemplate, tmpl.ID, models.ActionCreated,
		fmt.Sprintf("Created template: %s", tmpl.Name))

	return tmpl, nil
}

// UpdateTemplate updates an existing template
func (s *TemplateService) UpdateTemplate(userID uint, id uint, req *TemplateRequest) (*models.ConfigTemplate, error) {
	// Find existing template
	var tmpl models.ConfigTemplate
	if err := s.db.Where("id = ?", id).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTemplateNotFound
		}
		return nil, err
	}

	// Check permissions
	if tmpl.UserID != userID && !tmpl.IsBuiltIn {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, errors.ErrPermissionDenied
		}
	}

	// Built-in templates can only be modified by admins
	if tmpl.IsBuiltIn {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, fmt.Errorf("built-in templates can only be modified by administrators")
		}
	}

	// Validate category
	if !req.Category.IsValid() {
		return nil, fmt.Errorf("invalid template category")
	}

	// Validate template syntax
	if err := s.validateTemplate(req.Content); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Update template
	tmpl.Name = req.Name
	tmpl.Description = req.Description
	tmpl.Category = req.Category
	tmpl.Content = req.Content
	tmpl.Variables = models.JSON(req.Variables)
	tmpl.IsPublic = req.IsPublic

	// Save to database
	if err := s.db.Save(&tmpl).Error; err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeConfigTemplate, tmpl.ID, models.ActionUpdated,
		fmt.Sprintf("Updated template: %s", tmpl.Name))

	return &tmpl, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(userID uint, id uint) (*models.ConfigTemplate, error) {
	var tmpl models.ConfigTemplate
	query := s.db.Preload("User")

	if err := query.Where("id = ?", id).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTemplateNotFound
		}
		return nil, err
	}

	// Check permissions
	if tmpl.UserID != userID && !tmpl.IsPublic && !tmpl.IsBuiltIn {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, errors.ErrPermissionDenied
		}
	}

	return &tmpl, nil
}

// ListTemplates retrieves templates with pagination and filtering
func (s *TemplateService) ListTemplates(userID uint, page, limit int, category string, includePublic bool) (*TemplateListResponse, error) {
	offset := (page - 1) * limit

	query := s.db.Model(&models.ConfigTemplate{}).Preload("User")

	// Check if user is admin
	isAdmin := s.authService.IsAdmin(userID)

	// Apply access filters
	if !isAdmin {
		if includePublic {
			query = query.Where("user_id = ? OR is_public = true OR is_built_in = true", userID)
		} else {
			query = query.Where("user_id = ?", userID)
		}
	}

	// Filter by category if specified
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var templates []models.ConfigTemplate
	var total int64

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get templates with pagination
	if err := query.Offset(offset).Limit(limit).Find(&templates).Error; err != nil {
		return nil, err
	}

	return &TemplateListResponse{
		Templates: templates,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}, nil
}

// DeleteTemplate deletes a template
func (s *TemplateService) DeleteTemplate(userID uint, id uint) error {
	// Find template
	var tmpl models.ConfigTemplate
	if err := s.db.Where("id = ?", id).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrTemplateNotFound
		}
		return err
	}

	// Check permissions
	if tmpl.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return errors.ErrPermissionDenied
		}
	}

	// Built-in templates cannot be deleted
	if tmpl.IsBuiltIn {
		return fmt.Errorf("built-in templates cannot be deleted")
	}

	// Check if template is in use
	var count int64
	if err := s.db.Model(&models.NginxConfig{}).Where("template_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.ErrTemplateInUse
	}

	// Delete from database (soft delete due to BaseModel)
	if err := s.db.Delete(&tmpl).Error; err != nil {
		return err
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeConfigTemplate, tmpl.ID, models.ActionDeleted,
		fmt.Sprintf("Deleted template: %s", tmpl.Name))

	return nil
}

// RenderTemplate renders a template with given variables
func (s *TemplateService) RenderTemplate(userID uint, id uint, req *TemplateRenderRequest) (*TemplateRenderResponse, error) {
	// Get template
	tmpl, err := s.GetTemplate(userID, id)
	if err != nil {
		return nil, err
	}

	// Parse template
	t, err := template.New("template").Parse(tmpl.Content)
	if err != nil {
		return &TemplateRenderResponse{
			Content: "",
			IsValid: false,
			Errors:  []string{fmt.Sprintf("Template parse error: %s", err.Error())},
		}, nil
	}

	// Render template with variables
	var result strings.Builder
	if err := t.Execute(&result, req.Variables); err != nil {
		return &TemplateRenderResponse{
			Content: "",
			IsValid: false,
			Errors:  []string{fmt.Sprintf("Template execution error: %s", err.Error())},
		}, nil
	}

	// Increment usage count
	s.incrementUsageCount(id)

	return &TemplateRenderResponse{
		Content: result.String(),
		IsValid: true,
		Errors:  []string{},
	}, nil
}

// GetCategories returns all available template categories
func (s *TemplateService) GetCategories() []string {
	return []string{
		string(models.CategoryProxy),
		string(models.CategoryLoadBalance),
		string(models.CategorySSL),
		string(models.CategoryCache),
		string(models.CategorySecurity),
		string(models.CategoryCustom),
	}
}

// CreateBuiltInTemplates creates default built-in templates
func (s *TemplateService) CreateBuiltInTemplates() error {
	templates := s.getBuiltInTemplates()

	for _, tmpl := range templates {
		// Check if template already exists
		var existing models.ConfigTemplate
		err := s.db.Where("name = ? AND is_built_in = true", tmpl.Name).First(&existing).Error
		if err == nil {
			// Template already exists, skip
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		// Create built-in template
		if err := s.db.Create(&tmpl).Error; err != nil {
			logger.Error("Failed to create built-in template",
				logger.String("name", tmpl.Name),
				logger.Err(err))
		}
	}

	return nil
}

// validateTemplate validates template syntax
func (s *TemplateService) validateTemplate(content string) error {
	// Parse template to check syntax
	_, err := template.New("test").Parse(content)
	return err
}

// incrementUsageCount increments the usage count for a template
func (s *TemplateService) incrementUsageCount(templateID uint) {
	if err := s.db.Model(&models.ConfigTemplate{}).Where("id = ?", templateID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error; err != nil {
		logger.Warn("Failed to increment template usage count",
			logger.Uint("template_id", templateID),
			logger.Err(err))
	}
}

// getBuiltInTemplates returns the built-in template definitions
func (s *TemplateService) getBuiltInTemplates() []models.ConfigTemplate {
	return []models.ConfigTemplate{
		// Basic Reverse Proxy Template
		{
			Name:        "Basic Reverse Proxy",
			Description: "Simple reverse proxy configuration for forwarding requests to upstream servers",
			Category:    models.CategoryProxy,
			Content: `server {
    listen 80;
    listen [::]:80;
    server_name {{.domain}};

    location / {
        proxy_pass {{.upstream_url}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`,
			Variables: models.JSON{
				"domain": map[string]interface{}{
					"type":        "string",
					"description": "Domain name for the server",
					"required":    true,
					"example":     "example.com",
				},
				"upstream_url": map[string]interface{}{
					"type":        "string",
					"description": "Upstream server URL",
					"required":    true,
					"example":     "http://localhost:3000",
				},
			},
			IsBuiltIn: true,
			IsPublic:  true,
			UserID:    1, // System user
		},

		// SSL/HTTPS Template
		{
			Name:        "SSL/HTTPS Proxy",
			Description: "Reverse proxy with SSL termination and HTTP to HTTPS redirect",
			Category:    models.CategorySSL,
			Content: `server {
    listen 80;
    listen [::]:80;
    server_name {{.domain}};
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name {{.domain}};

    ssl_certificate {{.ssl_cert_path}};
    ssl_certificate_key {{.ssl_key_path}};
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass {{.upstream_url}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`,
			Variables: models.JSON{
				"domain": map[string]interface{}{
					"type":        "string",
					"description": "Domain name for the server",
					"required":    true,
					"example":     "example.com",
				},
				"upstream_url": map[string]interface{}{
					"type":        "string",
					"description": "Upstream server URL",
					"required":    true,
					"example":     "http://localhost:3000",
				},
				"ssl_cert_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to SSL certificate file",
					"required":    true,
					"example":     "/etc/ssl/certs/example.com.crt",
				},
				"ssl_key_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to SSL private key file",
					"required":    true,
					"example":     "/etc/ssl/private/example.com.key",
				},
			},
			IsBuiltIn: true,
			IsPublic:  true,
			UserID:    1,
		},

		// Load Balancer Template
		{
			Name:        "Load Balancer",
			Description: "Load balancing configuration with multiple upstream servers",
			Category:    models.CategoryLoadBalance,
			Content: `upstream {{.upstream_name}} {
    {{.load_balance_method}};
    {{range .servers}}server {{.}};
    {{end}}
}

server {
    listen 80;
    listen [::]:80;
    server_name {{.domain}};

    location / {
        proxy_pass http://{{.upstream_name}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Health checks
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
}`,
			Variables: models.JSON{
				"domain": map[string]interface{}{
					"type":        "string",
					"description": "Domain name for the server",
					"required":    true,
					"example":     "example.com",
				},
				"upstream_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the upstream group",
					"required":    true,
					"example":     "backend_servers",
				},
				"load_balance_method": map[string]interface{}{
					"type":        "string",
					"description": "Load balancing method",
					"required":    false,
					"example":     "least_conn",
					"options":     []string{"", "least_conn", "ip_hash", "hash $request_uri"},
				},
				"servers": map[string]interface{}{
					"type":        "array",
					"description": "List of upstream server addresses",
					"required":    true,
					"example":     []string{"127.0.0.1:3000", "127.0.0.1:3001", "127.0.0.1:3002"},
				},
			},
			IsBuiltIn: true,
			IsPublic:  true,
			UserID:    1,
		},

		// Static File Server Template
		{
			Name:        "Static File Server",
			Description: "Static file serving with caching and gzip compression",
			Category:    models.CategoryCache,
			Content: `server {
    listen 80;
    listen [::]:80;
    server_name {{.domain}};
    root {{.root_path}};
    index {{.index_files}};

    # Gzip compression
    gzip on;
    gzip_types
        text/plain
        text/css
        text/js
        text/xml
        text/javascript
        application/javascript
        application/xml+rss
        application/json;

    # Static file caching
    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Main location
    location / {
        try_files $uri $uri/ =404;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
}`,
			Variables: models.JSON{
				"domain": map[string]interface{}{
					"type":        "string",
					"description": "Domain name for the server",
					"required":    true,
					"example":     "static.example.com",
				},
				"root_path": map[string]interface{}{
					"type":        "string",
					"description": "Root directory path for static files",
					"required":    true,
					"example":     "/var/www/html",
				},
				"index_files": map[string]interface{}{
					"type":        "string",
					"description": "Index file names",
					"required":    false,
					"example":     "index.html index.htm",
				},
			},
			IsBuiltIn: true,
			IsPublic:  true,
			UserID:    1,
		},

		// WebSocket Proxy Template
		{
			Name:        "WebSocket Proxy",
			Description: "WebSocket proxy configuration with proper upgrade headers",
			Category:    models.CategoryProxy,
			Content: `server {
    listen 80;
    listen [::]:80;
    server_name {{.domain}};

    location {{.ws_path}} {
        proxy_pass {{.upstream_url}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket specific timeouts
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }

    {{if .include_regular_proxy}}
    location / {
        proxy_pass {{.upstream_url}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    {{end}}
}`,
			Variables: models.JSON{
				"domain": map[string]interface{}{
					"type":        "string",
					"description": "Domain name for the server",
					"required":    true,
					"example":     "ws.example.com",
				},
				"upstream_url": map[string]interface{}{
					"type":        "string",
					"description": "Upstream WebSocket server URL",
					"required":    true,
					"example":     "http://localhost:3000",
				},
				"ws_path": map[string]interface{}{
					"type":        "string",
					"description": "WebSocket endpoint path",
					"required":    false,
					"example":     "/ws",
				},
				"include_regular_proxy": map[string]interface{}{
					"type":        "boolean",
					"description": "Include regular HTTP proxy for non-WebSocket requests",
					"required":    false,
					"example":     true,
				},
			},
			IsBuiltIn: true,
			IsPublic:  true,
			UserID:    1,
		},
	}
}

// logAuditEvent logs an audit event
func (s *TemplateService) logAuditEvent(userID uint, objectType models.ObjectType, objectID uint, action models.AuditAction, description string) {
	auditLog := &models.AuditLog{
		UserID:      userID,
		Action:      action,
		ObjectType:  objectType,
		ObjectID:    objectID,
		Description: description,
		IPAddress:   "",
		UserAgent:   "",
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		logger.Error("Failed to create audit log", logger.Err(err))
	}
}
