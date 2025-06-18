package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/errors"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"gorm.io/gorm"
)

// ConfigService handles nginx configuration management
type ConfigService struct {
	db              *gorm.DB
	nginxConfigPath string
	backupPath      string
	templatePath    string
	authService     *AuthService
}

// NewConfigService creates a new configuration service instance
func NewConfigService(nginxConfigPath, backupPath, templatePath string, authService *AuthService) *ConfigService {
	return &ConfigService{
		db:              database.GetDB(),
		nginxConfigPath: nginxConfigPath,
		backupPath:      backupPath,
		templatePath:    templatePath,
		authService:     authService,
	}
}

// ConfigRequest represents configuration create/update request
type ConfigRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	Type         models.ConfigType      `json:"type" binding:"required"`
	Content      string                 `json:"content" binding:"required"`
	FilePath     string                 `json:"file_path"`
	IsActive     bool                   `json:"is_active"`
	TemplateID   *uint                  `json:"template_id,omitempty"`
	TemplateVars map[string]interface{} `json:"template_vars,omitempty"`
}

// ConfigListResponse represents paginated configuration list
type ConfigListResponse struct {
	Configs []models.NginxConfig `json:"configs"`
	Total   int64                `json:"total"`
	Page    int                  `json:"page"`
	Limit   int                  `json:"limit"`
}

// ValidationResult represents configuration validation result
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
	Output  string   `json:"output"`
}

// CreateConfig creates a new nginx configuration
func (s *ConfigService) CreateConfig(userID uint, req *ConfigRequest) (*models.NginxConfig, error) {
	// Validate config type
	if !req.Type.IsValid() {
		return nil, fmt.Errorf("invalid configuration type")
	}

	// Check for duplicate config name for user
	var existing models.NginxConfig
	err := s.db.Where("name = ? AND user_id = ?", req.Name, userID).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("configuration with this name already exists")
	}

	// Render content from template if template is used
	content := req.Content
	if req.TemplateID != nil {
		rendered, err := s.renderFromTemplate(*req.TemplateID, req.TemplateVars)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		content = rendered
	}

	// Validate configuration content
	validation, err := s.validateConfig(content)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create configuration model
	config := &models.NginxConfig{
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		Status:         models.StatusDraft,
		Content:        content,
		FilePath:       req.FilePath,
		IsActive:       false, // Start as inactive
		UserID:         userID,
		IsValid:        validation.IsValid,
		ValidationTime: time.Now(),
		ValidationLogs: validation.Output,
		TemplateID:     req.TemplateID,
		TemplateVars:   models.JSON(req.TemplateVars),
	}

	// Save to database
	if err := s.db.Create(config).Error; err != nil {
		return nil, err
	}

	// Create initial version
	if err := s.createVersion(config.ID, content, "Initial version", userID); err != nil {
		logger.Warn("Failed to create initial version", logger.Err(err))
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeNginxConfig, config.ID, models.ActionCreated,
		fmt.Sprintf("Created configuration: %s", config.Name))

	return config, nil
}

// UpdateConfig updates an existing configuration
func (s *ConfigService) UpdateConfig(userID uint, id uint, req *ConfigRequest) (*models.NginxConfig, error) {
	// Find existing configuration
	var config models.NginxConfig
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrConfigNotFound
		}
		return nil, err
	}

	// Check permissions
	if config.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, errors.ErrPermissionDenied
		}
	}

	// Check if config is read-only
	if config.IsReadOnly {
		return nil, fmt.Errorf("cannot modify read-only configuration")
	}

	// Create backup before modification
	if err := s.createBackup(config.ID, "Before update", userID); err != nil {
		logger.Warn("Failed to create backup", logger.Err(err))
	}

	// Render content from template if template is used
	content := req.Content
	if req.TemplateID != nil {
		rendered, err := s.renderFromTemplate(*req.TemplateID, req.TemplateVars)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		content = rendered
	}

	// Validate new configuration
	validation, err := s.validateConfig(content)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update configuration
	config.Name = req.Name
	config.Description = req.Description
	config.Type = req.Type
	config.Content = content
	config.FilePath = req.FilePath
	config.IsValid = validation.IsValid
	config.ValidationTime = time.Now()
	config.ValidationLogs = validation.Output
	config.TemplateID = req.TemplateID
	config.TemplateVars = models.JSON(req.TemplateVars)

	// Update status based on validation
	if validation.IsValid {
		if config.Status == models.StatusError {
			config.Status = models.StatusDraft
		}
	} else {
		config.Status = models.StatusError
	}

	// Save to database
	if err := s.db.Save(&config).Error; err != nil {
		return nil, err
	}

	// Create new version
	if err := s.createVersion(config.ID, content, "Configuration updated", userID); err != nil {
		logger.Warn("Failed to create version", logger.Err(err))
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeNginxConfig, config.ID, models.ActionUpdated,
		fmt.Sprintf("Updated configuration: %s", config.Name))

	return &config, nil
}

// GetConfig retrieves a configuration by ID
func (s *ConfigService) GetConfig(userID uint, id uint) (*models.NginxConfig, error) {
	var config models.NginxConfig
	query := s.db.Preload("User").Preload("Versions").Preload("TemplateTemplate")

	if err := query.Where("id = ?", id).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrConfigNotFound
		}
		return nil, err
	}

	// Check permissions
	if config.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, errors.ErrPermissionDenied
		}
	}

	return &config, nil
}

// ListConfigs retrieves configurations with pagination
func (s *ConfigService) ListConfigs(userID uint, page, limit int, configType string) (*ConfigListResponse, error) {
	offset := (page - 1) * limit

	query := s.db.Model(&models.NginxConfig{}).Preload("User")

	// Check if user is admin
	isAdmin := s.authService.IsAdmin(userID)
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by config type if specified
	if configType != "" {
		query = query.Where("type = ?", configType)
	}

	var configs []models.NginxConfig
	var total int64

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get configs with pagination
	if err := query.Offset(offset).Limit(limit).Find(&configs).Error; err != nil {
		return nil, err
	}

	return &ConfigListResponse{
		Configs: configs,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// DeleteConfig deletes a configuration
func (s *ConfigService) DeleteConfig(userID uint, id uint) error {
	// Find configuration
	var config models.NginxConfig
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrConfigNotFound
		}
		return err
	}

	// Check permissions
	if config.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return errors.ErrPermissionDenied
		}
	}

	// Check if config is read-only
	if config.IsReadOnly {
		return fmt.Errorf("cannot delete read-only configuration")
	}

	// Check if config is active
	if config.IsActive {
		return fmt.Errorf("cannot delete active configuration")
	}

	// Create final backup
	if err := s.createBackup(config.ID, "Before deletion", userID); err != nil {
		logger.Warn("Failed to create backup before deletion", logger.Err(err))
	}

	// Delete from database (soft delete due to BaseModel)
	if err := s.db.Delete(&config).Error; err != nil {
		return err
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeNginxConfig, config.ID, models.ActionDeleted,
		fmt.Sprintf("Deleted configuration: %s", config.Name))

	return nil
}

// DeployConfig deploys a configuration to nginx
func (s *ConfigService) DeployConfig(userID uint, id uint) error {
	// Find configuration
	var config models.NginxConfig
	if err := s.db.Where("id = ?", id).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrConfigNotFound
		}
		return err
	}

	// Check permissions
	if config.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return errors.ErrPermissionDenied
		}
	}

	// Validate configuration
	if !config.IsValid {
		return errors.ErrConfigValidationFailed
	}

	// Create backup before deployment
	if err := s.createBackup(config.ID, "Before deployment", userID); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Write configuration to file
	if err := s.writeConfigToFile(&config); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Test nginx configuration
	if err := s.testNginxConfig(); err != nil {
		return fmt.Errorf("nginx test failed: %w", err)
	}

	// Reload nginx
	if err := s.reloadNginx(); err != nil {
		return fmt.Errorf("nginx reload failed: %w", err)
	}

	// Update config status
	config.Status = models.StatusActive
	config.IsActive = true
	if err := s.db.Save(&config).Error; err != nil {
		return err
	}

	// Log audit event
	s.logAuditEvent(userID, models.ObjectTypeNginxConfig, config.ID, models.ActionUpdated,
		fmt.Sprintf("Deployed configuration: %s", config.Name))

	return nil
}

// ValidateConfig validates nginx configuration syntax
func (s *ConfigService) ValidateConfig(userID uint, content string) (*ValidationResult, error) {
	return s.validateConfig(content)
}

// validateConfig performs nginx configuration validation
func (s *ConfigService) validateConfig(content string) (*ValidationResult, error) {
	// Create temporary file
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("nginx_test_%d.conf", time.Now().UnixNano()))
	defer os.Remove(tempFile)

	// Write content to temporary file
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Run nginx -t on the temporary file
	cmd := exec.Command("nginx", "-t", "-c", tempFile)
	output, err := cmd.CombinedOutput()

	result := &ValidationResult{
		IsValid: err == nil,
		Output:  string(output),
		Errors:  []string{},
	}

	if err != nil {
		// Parse nginx error output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "test is successful") {
				result.Errors = append(result.Errors, line)
			}
		}
	}

	return result, nil
}

// writeConfigToFile writes configuration content to nginx config file
func (s *ConfigService) writeConfigToFile(config *models.NginxConfig) error {
	if config.FilePath == "" {
		return fmt.Errorf("file path not specified")
	}

	// Ensure directory exists
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write content to file
	if err := os.WriteFile(config.FilePath, []byte(config.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// testNginxConfig tests nginx configuration
func (s *ConfigService) testNginxConfig() error {
	cmd := exec.Command("nginx", "-t")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx test failed: %s", string(output))
	}
	return nil
}

// reloadNginx reloads nginx configuration
func (s *ConfigService) reloadNginx() error {
	cmd := exec.Command("nginx", "-s", "reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx reload failed: %s", string(output))
	}
	return nil
}

// renderFromTemplate renders configuration from template
func (s *ConfigService) renderFromTemplate(templateID uint, vars map[string]interface{}) (string, error) {
	// Get template
	var tmpl models.ConfigTemplate
	if err := s.db.Where("id = ?", templateID).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.ErrTemplateNotFound
		}
		return "", err
	}

	// Parse template
	t, err := template.New("config").Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("template parse failed: %w", err)
	}

	// Render template with variables
	var result strings.Builder
	if err := t.Execute(&result, vars); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return result.String(), nil
}

// createVersion creates a new configuration version
func (s *ConfigService) createVersion(configID uint, content, comment string, userID uint) error {
	// Get latest version number
	var latestVersion models.ConfigVersion
	err := s.db.Where("config_id = ?", configID).Order("version DESC").First(&latestVersion).Error
	version := 1
	if err == nil {
		version = latestVersion.Version + 1
	}

	// Create new version
	newVersion := &models.ConfigVersion{
		ConfigID:  configID,
		Version:   version,
		Content:   content,
		Comment:   comment,
		CreatedBy: userID,
	}

	return s.db.Create(newVersion).Error
}

// createBackup creates a configuration backup
func (s *ConfigService) createBackup(configID uint, reason string, userID uint) error {
	// Get configuration
	var config models.NginxConfig
	if err := s.db.Where("id = ?", configID).First(&config).Error; err != nil {
		return err
	}

	// Generate backup name
	backupName := fmt.Sprintf("%s_backup_%d", config.Name, time.Now().Unix())
	backupFilePath := filepath.Join(s.backupPath, backupName+".conf")

	// Create backup
	backup := &models.ConfigBackup{
		ConfigID:   configID,
		BackupName: backupName,
		Content:    config.Content,
		FilePath:   backupFilePath,
		Reason:     reason,
		AutoBackup: true,
		CreatedBy:  userID,
	}

	// Save backup to database
	if err := s.db.Create(backup).Error; err != nil {
		return err
	}

	// Write backup file
	if err := os.MkdirAll(s.backupPath, 0755); err != nil {
		return err
	}

	return os.WriteFile(backupFilePath, []byte(config.Content), 0644)
}

// logAuditEvent logs an audit event
func (s *ConfigService) logAuditEvent(userID uint, objectType models.ObjectType, objectID uint, action models.AuditAction, description string) {
	auditLog := &models.AuditLog{
		UserID:      userID,
		Action:      action,
		ObjectType:  objectType,
		ObjectID:    objectID,
		Description: description,
		IPAddress:   "", // This should be populated from the request context
		UserAgent:   "", // This should be populated from the request context
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		logger.Error("Failed to create audit log", logger.Err(err))
	}
}
