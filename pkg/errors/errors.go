package errors

import "errors"

// Common service errors
var (
	// Template errors
	ErrTemplateNotFound     = errors.New("template not found")
	ErrTemplateRenderFailed = errors.New("template render failed")
	ErrTemplateValidation   = errors.New("template validation failed")
	ErrTemplateDuplicate    = errors.New("template with this name already exists")
	ErrTemplateInUse        = errors.New("template is in use")

	// Configuration errors
	ErrConfigNotFound         = errors.New("configuration not found")
	ErrConfigValidationFailed = errors.New("configuration validation failed")
	ErrConfigInUse            = errors.New("configuration is in use")

	// General errors
	ErrBackupFailed     = errors.New("backup operation failed")
	ErrPermissionDenied = errors.New("permission denied")
)
