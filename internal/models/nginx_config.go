package models

import (
	"time"
)

// ConfigType represents the type of nginx configuration
type ConfigType string

const (
	ConfigTypeMain     ConfigType = "main"     // Main nginx.conf
	ConfigTypeServer   ConfigType = "server"   // Server block configuration
	ConfigTypeUpstream ConfigType = "upstream" // Upstream configuration
	ConfigTypeLocation ConfigType = "location" // Location block configuration
	ConfigTypeCustom   ConfigType = "custom"   // Custom configuration files
)

// IsValid checks if the config type is valid
func (ct ConfigType) IsValid() bool {
	switch ct {
	case ConfigTypeMain, ConfigTypeServer, ConfigTypeUpstream, ConfigTypeLocation, ConfigTypeCustom:
		return true
	}
	return false
}

// ConfigStatus represents the status of a configuration
type ConfigStatus string

const (
	StatusDraft    ConfigStatus = "draft"    // Draft configuration
	StatusActive   ConfigStatus = "active"   // Active configuration
	StatusInactive ConfigStatus = "inactive" // Inactive configuration
	StatusError    ConfigStatus = "error"    // Configuration with errors
)

// IsValid checks if the config status is valid
func (cs ConfigStatus) IsValid() bool {
	switch cs {
	case StatusDraft, StatusActive, StatusInactive, StatusError:
		return true
	}
	return false
}

// TemplateCategory represents template categories
type TemplateCategory string

const (
	CategoryProxy       TemplateCategory = "proxy"        // Reverse proxy templates
	CategoryLoadBalance TemplateCategory = "load_balance" // Load balancing templates
	CategorySSL         TemplateCategory = "ssl"          // SSL/TLS templates
	CategoryCache       TemplateCategory = "cache"        // Caching templates
	CategorySecurity    TemplateCategory = "security"     // Security templates
	CategoryCustom      TemplateCategory = "custom"       // Custom templates
)

// IsValid checks if the template category is valid
func (tc TemplateCategory) IsValid() bool {
	switch tc {
	case CategoryProxy, CategoryLoadBalance, CategorySSL, CategoryCache, CategorySecurity, CategoryCustom:
		return true
	}
	return false
}

// NginxConfig represents an nginx configuration file
type NginxConfig struct {
	BaseModel
	Name        string       `json:"name" gorm:"not null;uniqueIndex:idx_config_name_user"`
	Description string       `json:"description"`
	Type        ConfigType   `json:"type" gorm:"not null"`
	Status      ConfigStatus `json:"status" gorm:"default:'draft'"`
	Content     string       `json:"content" gorm:"type:text"`
	FilePath    string       `json:"file_path"` // Path to the actual nginx config file
	IsActive    bool         `json:"is_active" gorm:"default:false"`
	IsReadOnly  bool         `json:"is_read_only" gorm:"default:false"` // System configs are read-only
	UserID      uint         `json:"user_id" gorm:"not null;uniqueIndex:idx_config_name_user"`

	// Validation
	IsValid        bool      `json:"is_valid" gorm:"default:false"`
	ValidationTime time.Time `json:"validation_time"`
	ValidationLogs string    `json:"validation_logs" gorm:"type:text"`

	// Relationships
	User     User            `json:"user" gorm:"foreignKey:UserID"`
	Versions []ConfigVersion `json:"versions" gorm:"foreignKey:ConfigID"`
	Backups  []ConfigBackup  `json:"backups" gorm:"foreignKey:ConfigID"`

	// Template information
	TemplateID       *uint           `json:"template_id,omitempty"`
	TemplateTemplate *ConfigTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	TemplateVars     JSON            `json:"template_vars" gorm:"type:jsonb"`
}

// ConfigVersion represents a version of a configuration
type ConfigVersion struct {
	BaseModel
	ConfigID  uint   `json:"config_id" gorm:"not null"`
	Version   int    `json:"version" gorm:"not null"`
	Content   string `json:"content" gorm:"type:text"`
	Comment   string `json:"comment"`
	IsBackup  bool   `json:"is_backup" gorm:"default:false"`
	CreatedBy uint   `json:"created_by" gorm:"not null"`

	// Relationships
	Config        NginxConfig `json:"config" gorm:"foreignKey:ConfigID"`
	CreatedByUser User        `json:"created_by_user" gorm:"foreignKey:CreatedBy"`
}

// ConfigBackup represents a backup of a configuration
type ConfigBackup struct {
	BaseModel
	ConfigID   uint   `json:"config_id" gorm:"not null"`
	BackupName string `json:"backup_name" gorm:"not null"`
	Content    string `json:"content" gorm:"type:text"`
	FilePath   string `json:"file_path"` // Path to backup file
	Reason     string `json:"reason"`    // Reason for backup
	AutoBackup bool   `json:"auto_backup" gorm:"default:true"`
	CreatedBy  uint   `json:"created_by" gorm:"not null"`

	// Relationships
	Config        NginxConfig `json:"config" gorm:"foreignKey:ConfigID"`
	CreatedByUser User        `json:"created_by_user" gorm:"foreignKey:CreatedBy"`
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	BaseModel
	Name        string           `json:"name" gorm:"not null;uniqueIndex:idx_template_name_user"`
	Description string           `json:"description"`
	Category    TemplateCategory `json:"category" gorm:"not null"`
	Content     string           `json:"content" gorm:"type:text"`
	Variables   JSON             `json:"variables" gorm:"type:jsonb"` // Template variable definitions
	IsBuiltIn   bool             `json:"is_built_in" gorm:"default:false"`
	IsPublic    bool             `json:"is_public" gorm:"default:false"`
	UsageCount  int              `json:"usage_count" gorm:"default:0"`
	UserID      uint             `json:"user_id" gorm:"not null;uniqueIndex:idx_template_name_user"`

	// Relationships
	User    User          `json:"user" gorm:"foreignKey:UserID"`
	Configs []NginxConfig `json:"configs" gorm:"foreignKey:TemplateID"`
}

// ConfigApproval represents configuration change approval workflow
type ConfigApproval struct {
	BaseModel
	ConfigID    uint           `json:"config_id" gorm:"not null"`
	RequestedBy uint           `json:"requested_by" gorm:"not null"`
	ApprovedBy  *uint          `json:"approved_by,omitempty"`
	Status      ApprovalStatus `json:"status" gorm:"default:'pending'"`
	Content     string         `json:"content" gorm:"type:text"`
	Comment     string         `json:"comment"`
	ApprovedAt  *time.Time     `json:"approved_at,omitempty"`

	// Relationships
	Config          NginxConfig `json:"config" gorm:"foreignKey:ConfigID"`
	RequestedByUser User        `json:"requested_by_user" gorm:"foreignKey:RequestedBy"`
	ApprovedByUser  *User       `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
}

// ApprovalStatus represents approval status
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
	ApprovalCanceled ApprovalStatus = "canceled"
)

// IsValid checks if the approval status is valid
func (as ApprovalStatus) IsValid() bool {
	switch as {
	case ApprovalPending, ApprovalApproved, ApprovalRejected, ApprovalCanceled:
		return true
	}
	return false
}

// TableName returns the table name for NginxConfig
func (NginxConfig) TableName() string {
	return "nginx_configs"
}

// TableName returns the table name for ConfigVersion
func (ConfigVersion) TableName() string {
	return "config_versions"
}

// TableName returns the table name for ConfigBackup
func (ConfigBackup) TableName() string {
	return "config_backups"
}

// TableName returns the table name for ConfigTemplate
func (ConfigTemplate) TableName() string {
	return "config_templates"
}

// TableName returns the table name for ConfigApproval
func (ConfigApproval) TableName() string {
	return "config_approvals"
}
