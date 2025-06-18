package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"gorm.io/gorm"
)

var (
	ErrProxyHostNotFound     = errors.New("proxy host not found")
	ErrInvalidDomainName     = errors.New("invalid domain name")
	ErrNginxConfigGeneration = errors.New("failed to generate nginx configuration")
	ErrNginxReload           = errors.New("failed to reload nginx")
)

// NginxService handles nginx configuration management
type NginxService struct {
	db           *gorm.DB
	configPath   string
	sitesPath    string
	backupPath   string
	templatePath string
	authService  *AuthService
}

// NewNginxService creates a new nginx service instance
func NewNginxService(configPath, sitesPath, backupPath, templatePath string, authService *AuthService) *NginxService {
	return &NginxService{
		db:           database.GetDB(),
		configPath:   configPath,
		sitesPath:    sitesPath,
		backupPath:   backupPath,
		templatePath: templatePath,
		authService:  authService,
	}
}

// ProxyHostRequest represents proxy host create/update request
type ProxyHostRequest struct {
	DomainNames           []string               `json:"domain_names" binding:"required"`
	ForwardScheme         models.ForwardScheme   `json:"forward_scheme" binding:"required"`
	ForwardHost           string                 `json:"forward_host" binding:"required"`
	ForwardPort           int                    `json:"forward_port" binding:"required"`
	AccessListID          *uint                  `json:"access_list_id"`
	CertificateID         *uint                  `json:"certificate_id"`
	SSLForced             bool                   `json:"ssl_forced"`
	CachingEnabled        bool                   `json:"caching_enabled"`
	BlockExploits         bool                   `json:"block_exploits"`
	AllowWebsocketUpgrade bool                   `json:"allow_websocket_upgrade"`
	HTTP2Support          bool                   `json:"http2_support"`
	HSTSEnabled           bool                   `json:"hsts_enabled"`
	HSTSSubdomains        bool                   `json:"hsts_subdomains"`
	AdvancedConfig        string                 `json:"advanced_config"`
	Enabled               bool                   `json:"enabled"`
	Locations             map[string]interface{} `json:"locations"`
}

// CreateProxyHost creates a new proxy host
func (s *NginxService) CreateProxyHost(userID uint, req *ProxyHostRequest) (*models.ProxyHost, error) {
	// Validate domain names
	if err := s.validateDomainNames(req.DomainNames); err != nil {
		return nil, err
	}

	// Check for duplicate domain names
	if err := s.checkDuplicateDomains(0, req.DomainNames); err != nil {
		return nil, err
	}

	// Validate forward scheme
	if !req.ForwardScheme.IsValid() {
		return nil, errors.New("invalid forward scheme")
	}

	// Create proxy host model
	proxyHost := &models.ProxyHost{
		DomainNames:           models.StringArray(req.DomainNames),
		ForwardScheme:         req.ForwardScheme,
		ForwardHost:           req.ForwardHost,
		ForwardPort:           req.ForwardPort,
		AccessListID:          req.AccessListID,
		CertificateID:         req.CertificateID,
		SSLForced:             req.SSLForced,
		CachingEnabled:        req.CachingEnabled,
		BlockExploits:         req.BlockExploits,
		AllowWebsocketUpgrade: req.AllowWebsocketUpgrade,
		HTTP2Support:          req.HTTP2Support,
		HSTSEnabled:           req.HSTSEnabled,
		HSTSSubdomains:        req.HSTSSubdomains,
		AdvancedConfig:        req.AdvancedConfig,
		Enabled:               req.Enabled,
		Locations:             models.JSON(req.Locations),
		UserID:                userID,
	}

	// Save to database
	if err := s.db.Create(proxyHost).Error; err != nil {
		return nil, err
	}

	// Generate nginx configuration
	if err := s.generateConfig(proxyHost); err != nil {
		// Rollback database changes
		s.db.Delete(proxyHost)
		return nil, fmt.Errorf("failed to generate nginx config: %w", err)
	}

	// Reload nginx if enabled
	if proxyHost.Enabled {
		if err := s.reloadNginx(); err != nil {
			logger.Warn("Failed to reload nginx", logger.Err(err))
		}
	}

	return proxyHost, nil
}

// UpdateProxyHost updates an existing proxy host
func (s *NginxService) UpdateProxyHost(userID uint, id uint, req *ProxyHostRequest) (*models.ProxyHost, error) {
	// Find existing proxy host
	var proxyHost models.ProxyHost
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&proxyHost).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProxyHostNotFound
		}
		return nil, err
	}

	// Check admin permission for cross-user management
	if proxyHost.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, err
		}
	}

	// Validate domain names
	if err := s.validateDomainNames(req.DomainNames); err != nil {
		return nil, err
	}

	// Check for duplicate domain names (excluding current proxy host)
	if err := s.checkDuplicateDomains(id, req.DomainNames); err != nil {
		return nil, err
	}

	// Backup current configuration
	if err := s.backupConfig(&proxyHost); err != nil {
		logger.Warn("Failed to backup config", logger.Err(err))
	}

	// Update proxy host
	proxyHost.DomainNames = models.StringArray(req.DomainNames)
	proxyHost.ForwardScheme = req.ForwardScheme
	proxyHost.ForwardHost = req.ForwardHost
	proxyHost.ForwardPort = req.ForwardPort
	proxyHost.AccessListID = req.AccessListID
	proxyHost.CertificateID = req.CertificateID
	proxyHost.SSLForced = req.SSLForced
	proxyHost.CachingEnabled = req.CachingEnabled
	proxyHost.BlockExploits = req.BlockExploits
	proxyHost.AllowWebsocketUpgrade = req.AllowWebsocketUpgrade
	proxyHost.HTTP2Support = req.HTTP2Support
	proxyHost.HSTSEnabled = req.HSTSEnabled
	proxyHost.HSTSSubdomains = req.HSTSSubdomains
	proxyHost.AdvancedConfig = req.AdvancedConfig
	proxyHost.Enabled = req.Enabled
	proxyHost.Locations = models.JSON(req.Locations)

	// Save to database
	if err := s.db.Save(&proxyHost).Error; err != nil {
		return nil, err
	}

	// Regenerate nginx configuration
	if err := s.generateConfig(&proxyHost); err != nil {
		return nil, fmt.Errorf("failed to regenerate nginx config: %w", err)
	}

	// Reload nginx
	if err := s.reloadNginx(); err != nil {
		logger.Warn("Failed to reload nginx", logger.Err(err))
	}

	return &proxyHost, nil
}

// DeleteProxyHost deletes a proxy host
func (s *NginxService) DeleteProxyHost(userID uint, id uint) error {
	// Find proxy host
	var proxyHost models.ProxyHost
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&proxyHost).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProxyHostNotFound
		}
		return err
	}

	// Check admin permission for cross-user management
	if proxyHost.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return err
		}
	}

	// Backup configuration before deletion
	if err := s.backupConfig(&proxyHost); err != nil {
		logger.Warn("Failed to backup config before deletion", logger.Err(err))
	}

	// Delete from database
	if err := s.db.Delete(&proxyHost).Error; err != nil {
		return err
	}

	// Remove nginx configuration file
	if err := s.removeConfig(&proxyHost); err != nil {
		logger.Warn("Failed to remove nginx config", logger.Err(err))
	}

	// Reload nginx
	if err := s.reloadNginx(); err != nil {
		logger.Warn("Failed to reload nginx", logger.Err(err))
	}

	return nil
}

// GetProxyHost gets a single proxy host
func (s *NginxService) GetProxyHost(userID uint, id uint) (*models.ProxyHost, error) {
	var proxyHost models.ProxyHost
	query := s.db.Preload("User").Preload("Certificate").Preload("AccessList")

	// Admin can see all proxy hosts
	if s.authService.RequireAdmin(userID) == nil {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where("id = ? AND user_id = ?", id, userID)
	}

	if err := query.First(&proxyHost).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProxyHostNotFound
		}
		return nil, err
	}

	return &proxyHost, nil
}

// ListProxyHosts lists proxy hosts with pagination
func (s *NginxService) ListProxyHosts(userID uint, offset, limit int) ([]models.ProxyHost, int64, error) {
	var proxyHosts []models.ProxyHost
	var total int64

	query := s.db.Model(&models.ProxyHost{}).Preload("User").Preload("Certificate").Preload("AccessList")

	// Admin can see all proxy hosts
	if s.authService.RequireAdmin(userID) != nil {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&proxyHosts).Error; err != nil {
		return nil, 0, err
	}

	return proxyHosts, total, nil
}

// validateDomainNames validates domain name format
func (s *NginxService) validateDomainNames(domains []string) error {
	if len(domains) == 0 {
		return errors.New("at least one domain name is required")
	}

	for _, domain := range domains {
		if strings.TrimSpace(domain) == "" {
			return ErrInvalidDomainName
		}
		// Add more domain validation logic here
	}

	return nil
}

// checkDuplicateDomains checks for duplicate domain names
func (s *NginxService) checkDuplicateDomains(excludeID uint, domains []string) error {
	for _, domain := range domains {
		var count int64
		query := s.db.Model(&models.ProxyHost{}).Where("JSON_EXTRACT(domain_names, '$') LIKE ?", "%"+domain+"%")

		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}

		if err := query.Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return fmt.Errorf("domain %s is already in use", domain)
		}
	}

	return nil
}

// generateConfig generates nginx configuration for proxy host
func (s *NginxService) generateConfig(proxyHost *models.ProxyHost) error {
	// Load certificate if specified
	var certificate *models.Certificate
	if proxyHost.CertificateID != nil {
		if err := s.db.Where("id = ?", *proxyHost.CertificateID).First(&certificate).Error; err != nil {
			logger.Warn("Failed to load certificate", logger.Err(err))
		}
	}

	// Load access list if specified
	var accessList *models.AccessList
	if proxyHost.AccessListID != nil {
		if err := s.db.Preload("AccessListAuths").Preload("AccessListClients").
			Where("id = ?", *proxyHost.AccessListID).First(&accessList).Error; err != nil {
			logger.Warn("Failed to load access list", logger.Err(err))
		}
	}

	// Generate configuration content
	configContent, err := s.renderTemplate(proxyHost, certificate, accessList)
	if err != nil {
		return err
	}

	// Write configuration file
	configFile := filepath.Join(s.sitesPath, fmt.Sprintf("proxy_host_%d.conf", proxyHost.ID))
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return err
	}

	return nil
}

// renderTemplate renders nginx configuration template
func (s *NginxService) renderTemplate(proxyHost *models.ProxyHost, certificate *models.Certificate, accessList *models.AccessList) (string, error) {
	templateFile := filepath.Join(s.templatePath, "proxy_host.tmpl")

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		// Fallback to basic template
		return s.generateBasicConfig(proxyHost, certificate, accessList), nil
	}

	data := map[string]interface{}{
		"ProxyHost":   proxyHost,
		"Certificate": certificate,
		"AccessList":  accessList,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateBasicConfig generates basic nginx configuration
func (s *NginxService) generateBasicConfig(proxyHost *models.ProxyHost, certificate *models.Certificate, accessList *models.AccessList) string {
	var config strings.Builder

	// Server block
	config.WriteString("server {\n")

	// Listen directives
	if certificate != nil && certificate.IsValid() {
		config.WriteString("    listen 443 ssl")
		if proxyHost.HTTP2Support {
			config.WriteString(" http2")
		}
		config.WriteString(";\n")

		// SSL configuration
		config.WriteString(fmt.Sprintf("    ssl_certificate /etc/nginx/certificates/cert_%d.pem;\n", certificate.ID))
		config.WriteString(fmt.Sprintf("    ssl_certificate_key /etc/nginx/certificates/key_%d.pem;\n", certificate.ID))
	} else {
		config.WriteString("    listen 80;\n")
	}

	// Server names
	config.WriteString("    server_name")
	for _, domain := range proxyHost.DomainNames {
		config.WriteString(" " + domain)
	}
	config.WriteString(";\n")

	// Access control
	if accessList != nil {
		config.WriteString("    # Access control\n")
		// Add access control directives
	}

	// Proxy configuration
	config.WriteString("    location / {\n")
	config.WriteString(fmt.Sprintf("        proxy_pass %s;\n", proxyHost.GetTargetURL()))
	config.WriteString("        proxy_set_header Host $host;\n")
	config.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
	config.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
	config.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")

	if proxyHost.AllowWebsocketUpgrade {
		config.WriteString("        proxy_set_header Upgrade $http_upgrade;\n")
		config.WriteString("        proxy_set_header Connection \"upgrade\";\n")
	}

	config.WriteString("    }\n")

	// Custom locations
	if len(proxyHost.Locations) > 0 {
		// Add custom location blocks
	}

	// Advanced configuration
	if proxyHost.AdvancedConfig != "" {
		config.WriteString("\n    # Advanced configuration\n")
		config.WriteString("    " + strings.ReplaceAll(proxyHost.AdvancedConfig, "\n", "\n    ") + "\n")
	}

	config.WriteString("}\n")

	// HTTP to HTTPS redirect if SSL is forced
	if proxyHost.SSLForced && certificate != nil {
		config.WriteString("\nserver {\n")
		config.WriteString("    listen 80;\n")
		config.WriteString("    server_name")
		for _, domain := range proxyHost.DomainNames {
			config.WriteString(" " + domain)
		}
		config.WriteString(";\n")
		config.WriteString("    return 301 https://$server_name$request_uri;\n")
		config.WriteString("}\n")
	}

	return config.String()
}

// backupConfig creates a backup of current configuration
func (s *NginxService) backupConfig(proxyHost *models.ProxyHost) error {
	configFile := filepath.Join(s.sitesPath, fmt.Sprintf("proxy_host_%d.conf", proxyHost.ID))

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // No config to backup
	}

	backupFile := filepath.Join(s.backupPath, fmt.Sprintf("proxy_host_%d_%d.conf.bak",
		proxyHost.ID, time.Now().Unix()))

	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupPath, 0755); err != nil {
		return err
	}

	// Copy file
	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	return os.WriteFile(backupFile, content, 0644)
}

// removeConfig removes nginx configuration file
func (s *NginxService) removeConfig(proxyHost *models.ProxyHost) error {
	configFile := filepath.Join(s.sitesPath, fmt.Sprintf("proxy_host_%d.conf", proxyHost.ID))
	return os.Remove(configFile)
}

// reloadNginx reloads nginx configuration
func (s *NginxService) reloadNginx() error {
	// In production, this would execute nginx reload command
	// For now, we'll just log the action
	logger.Info("Nginx configuration reloaded")
	return nil
}
