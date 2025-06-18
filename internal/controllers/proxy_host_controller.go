package controllers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/middleware"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// ProxyHostController handles proxy host management
type ProxyHostController struct {
	nginxService *services.NginxService
}

// NewProxyHostController creates a new proxy host controller
func NewProxyHostController(nginxService *services.NginxService) *ProxyHostController {
	return &ProxyHostController{
		nginxService: nginxService,
	}
}

// CreateProxyHostRequest represents the request payload for creating a proxy host
type CreateProxyHostRequest struct {
	DomainNames           []string               `json:"domain_names" binding:"required,min=1"`
	ForwardScheme         models.ForwardScheme   `json:"forward_scheme" binding:"required,oneof=http https"`
	ForwardHost           string                 `json:"forward_host" binding:"required"`
	ForwardPort           int                    `json:"forward_port" binding:"required,min=1,max=65535"`
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
	Meta                  map[string]interface{} `json:"meta"`
}

// UpdateProxyHostRequest represents the request payload for updating a proxy host
type UpdateProxyHostRequest struct {
	CreateProxyHostRequest
}

// ProxyHostListResponse represents a single proxy host in list view
type ProxyHostListResponse struct {
	ID            uint                 `json:"id"`
	DomainNames   []string             `json:"domain_names"`
	ForwardScheme models.ForwardScheme `json:"forward_scheme"`
	ForwardHost   string               `json:"forward_host"`
	ForwardPort   int                  `json:"forward_port"`
	AccessListID  *uint                `json:"access_list_id"`
	CertificateID *uint                `json:"certificate_id"`
	SSLForced     bool                 `json:"ssl_forced"`
	Enabled       bool                 `json:"enabled"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`

	// Computed fields
	PrimaryDomain string `json:"primary_domain"`
	TargetURL     string `json:"target_url"`
	SSLEnabled    bool   `json:"ssl_enabled"`
	HasAccessList bool   `json:"has_access_list"`

	// Related entities (optional)
	Certificate *models.Certificate `json:"certificate,omitempty"`
	AccessList  *models.AccessList  `json:"access_list,omitempty"`
}

// ProxyHostDetailResponse represents a proxy host detail view
type ProxyHostDetailResponse struct {
	ProxyHostListResponse
	CachingEnabled        bool                   `json:"caching_enabled"`
	BlockExploits         bool                   `json:"block_exploits"`
	AllowWebsocketUpgrade bool                   `json:"allow_websocket_upgrade"`
	HTTP2Support          bool                   `json:"http2_support"`
	HSTSEnabled           bool                   `json:"hsts_enabled"`
	HSTSSubdomains        bool                   `json:"hsts_subdomains"`
	AdvancedConfig        string                 `json:"advanced_config"`
	Locations             map[string]interface{} `json:"locations"`
	Meta                  map[string]interface{} `json:"meta"`

	// Nginx configuration
	NginxConfig string `json:"nginx_config,omitempty"`
	ConfigValid bool   `json:"config_valid"`
}

// List returns paginated list of proxy hosts for the current user
func (pc *ProxyHostController) List(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	enabled := c.Query("enabled")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get database connection
	db := database.GetDB()

	// Build query
	query := db.Where("user_id = ?", userID)

	if search != "" {
		query = query.Where("domain_names LIKE ? OR forward_host LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if enabled != "" {
		switch enabled {
		case "true":
			query = query.Where("enabled = ?", true)
		case "false":
			query = query.Where("enabled = ?", false)
		}
	}

	// Count total records
	var total int64
	if err := query.Model(&models.ProxyHost{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count proxy hosts", logger.Err(err), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to count proxy hosts", err)
		return
	}

	// Get paginated results
	var proxyHosts []models.ProxyHost
	offset := (page - 1) * limit
	if err := query.Preload("Certificate").Preload("AccessList").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&proxyHosts).Error; err != nil {
		logger.Error("Failed to fetch proxy hosts", logger.Err(err), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to fetch proxy hosts", err)
		return
	}

	// Convert to response format
	var proxyHostResponses []ProxyHostListResponse
	for _, host := range proxyHosts {
		resp := ProxyHostListResponse{
			ID:            host.ID,
			DomainNames:   host.DomainNames,
			ForwardScheme: host.ForwardScheme,
			ForwardHost:   host.ForwardHost,
			ForwardPort:   host.ForwardPort,
			AccessListID:  host.AccessListID,
			CertificateID: host.CertificateID,
			SSLForced:     host.SSLForced,
			Enabled:       host.Enabled,
			CreatedAt:     host.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     host.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			PrimaryDomain: host.GetPrimaryDomain(),
			TargetURL:     host.GetTargetURL(),
			SSLEnabled:    host.IsSSLEnabled(),
			HasAccessList: host.HasAccessList(),
		}

		if host.Certificate != nil {
			resp.Certificate = host.Certificate
		}
		if host.AccessList != nil {
			resp.AccessList = host.AccessList
		}

		proxyHostResponses = append(proxyHostResponses, resp)
	}

	// Pagination info
	response.SuccessJSONWithLog(c, gin.H{
		"data": proxyHostResponses,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"pages":    (total + int64(limit) - 1) / int64(limit),
			"has_next": page*limit < int(total),
			"has_prev": page > 1,
		},
	}, "Proxy hosts retrieved successfully")
}

// Get returns a single proxy host by ID
func (pc *ProxyHostController) Get(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid proxy host ID", err)
		return
	}

	db := database.GetDB()
	var proxyHost models.ProxyHost
	if err := db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Certificate").Preload("AccessList").
		First(&proxyHost).Error; err != nil {
		logger.Error("Failed to fetch proxy host", logger.Err(err), logger.Uint("id", uint(id)), logger.Uint("user_id", userID))
		response.NotFoundJSONWithLog(c, "Proxy host not found")
		return
	}

	// Generate nginx configuration preview if nginx service is available
	var nginxConfig string
	var configValid bool
	if pc.nginxService != nil {
		nginxConfig, configValid = pc.generateProxyHostConfig(&proxyHost)
	}

	resp := ProxyHostDetailResponse{
		ProxyHostListResponse: ProxyHostListResponse{
			ID:            proxyHost.ID,
			DomainNames:   proxyHost.DomainNames,
			ForwardScheme: proxyHost.ForwardScheme,
			ForwardHost:   proxyHost.ForwardHost,
			ForwardPort:   proxyHost.ForwardPort,
			AccessListID:  proxyHost.AccessListID,
			CertificateID: proxyHost.CertificateID,
			SSLForced:     proxyHost.SSLForced,
			Enabled:       proxyHost.Enabled,
			CreatedAt:     proxyHost.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     proxyHost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			PrimaryDomain: proxyHost.GetPrimaryDomain(),
			TargetURL:     proxyHost.GetTargetURL(),
			SSLEnabled:    proxyHost.IsSSLEnabled(),
			HasAccessList: proxyHost.HasAccessList(),
			Certificate:   proxyHost.Certificate,
			AccessList:    proxyHost.AccessList,
		},
		CachingEnabled:        proxyHost.CachingEnabled,
		BlockExploits:         proxyHost.BlockExploits,
		AllowWebsocketUpgrade: proxyHost.AllowWebsocketUpgrade,
		HTTP2Support:          proxyHost.HTTP2Support,
		HSTSEnabled:           proxyHost.HSTSEnabled,
		HSTSSubdomains:        proxyHost.HSTSSubdomains,
		AdvancedConfig:        proxyHost.AdvancedConfig,
		Locations:             proxyHost.Locations,
		Meta:                  proxyHost.Meta,
		NginxConfig:           nginxConfig,
		ConfigValid:           configValid,
	}

	response.SuccessJSONWithLog(c, resp, "Proxy host retrieved successfully")
}

// Create creates a new proxy host
func (pc *ProxyHostController) Create(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	var req CreateProxyHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	// Validate domain names
	if err := pc.validateDomainNames(req.DomainNames); err != nil {
		response.BadRequestJSONWithLog(c, err.Error(), err)
		return
	}

	// Check for duplicate domains
	if err := pc.checkDuplicateDomains(req.DomainNames, 0); err != nil {
		response.BadRequestJSONWithLog(c, err.Error(), err)
		return
	}

	// Create proxy host model
	proxyHost := models.ProxyHost{
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
		UserID:                userID,
	}

	if req.Locations != nil {
		proxyHost.Locations = models.JSON(req.Locations)
	}
	if req.Meta != nil {
		proxyHost.Meta = models.JSON(req.Meta)
	}

	// Save to database
	db := database.GetDB()
	if err := db.Create(&proxyHost).Error; err != nil {
		logger.Error("Failed to create proxy host", logger.Err(err), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to create proxy host", err)
		return
	}

	// Generate and apply nginx configuration if enabled and service is available
	if proxyHost.Enabled && pc.nginxService != nil {
		if err := pc.applyProxyHostConfig(&proxyHost); err != nil {
			logger.Error("Failed to apply nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", proxyHost.ID))
			// Continue anyway, don't fail the creation
		}
	}

	logger.Info("Proxy host created successfully", logger.Uint("id", proxyHost.ID), logger.Uint("user_id", userID), logger.Any("domains", req.DomainNames))
	response.SuccessJSONWithLog(c, proxyHost, "Proxy host created successfully")
}

// Update updates an existing proxy host
func (pc *ProxyHostController) Update(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid proxy host ID", err)
		return
	}

	var req UpdateProxyHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	// Find existing proxy host
	db := database.GetDB()
	var proxyHost models.ProxyHost
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&proxyHost).Error; err != nil {
		response.NotFoundJSONWithLog(c, "Proxy host not found")
		return
	}

	// Validate domain names
	if err := pc.validateDomainNames(req.DomainNames); err != nil {
		response.BadRequestJSONWithLog(c, err.Error(), err)
		return
	}

	// Check for duplicate domains (excluding current host)
	if err := pc.checkDuplicateDomains(req.DomainNames, uint(id)); err != nil {
		response.BadRequestJSONWithLog(c, err.Error(), err)
		return
	}

	// Update fields
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

	if req.Locations != nil {
		proxyHost.Locations = models.JSON(req.Locations)
	}
	if req.Meta != nil {
		proxyHost.Meta = models.JSON(req.Meta)
	}

	// Save changes
	if err := db.Save(&proxyHost).Error; err != nil {
		logger.Error("Failed to update proxy host", logger.Err(err), logger.Uint("id", uint(id)), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to update proxy host", err)
		return
	}

	// Update nginx configuration
	if pc.nginxService != nil {
		if proxyHost.Enabled {
			if err := pc.applyProxyHostConfig(&proxyHost); err != nil {
				logger.Error("Failed to apply nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", proxyHost.ID))
			}
		} else {
			if err := pc.removeProxyHostConfig(&proxyHost); err != nil {
				logger.Error("Failed to remove nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", proxyHost.ID))
			}
		}
	}

	logger.Info("Proxy host updated successfully", logger.Uint("id", proxyHost.ID), logger.Uint("user_id", userID))
	response.SuccessJSONWithLog(c, proxyHost, "Proxy host updated successfully")
}

// Delete deletes a proxy host
func (pc *ProxyHostController) Delete(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid proxy host ID", err)
		return
	}

	// Find existing proxy host
	db := database.GetDB()
	var proxyHost models.ProxyHost
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&proxyHost).Error; err != nil {
		response.NotFoundJSONWithLog(c, "Proxy host not found")
		return
	}

	// Remove nginx configuration
	if pc.nginxService != nil {
		if err := pc.removeProxyHostConfig(&proxyHost); err != nil {
			logger.Error("Failed to remove nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", proxyHost.ID))
			// Continue anyway
		}
	}

	// Delete from database
	if err := db.Delete(&proxyHost).Error; err != nil {
		logger.Error("Failed to delete proxy host", logger.Err(err), logger.Uint("id", uint(id)), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to delete proxy host", err)
		return
	}

	logger.Info("Proxy host deleted successfully", logger.Uint("id", uint(id)), logger.Uint("user_id", userID))
	response.SuccessJSONWithLog(c, gin.H{"id": id}, "Proxy host deleted successfully")
}

// Toggle toggles the enabled status of a proxy host
func (pc *ProxyHostController) Toggle(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid proxy host ID", err)
		return
	}

	// Find existing proxy host
	db := database.GetDB()
	var proxyHost models.ProxyHost
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&proxyHost).Error; err != nil {
		response.NotFoundJSONWithLog(c, "Proxy host not found")
		return
	}

	// Toggle enabled status
	proxyHost.Enabled = !proxyHost.Enabled

	// Save changes
	if err := db.Save(&proxyHost).Error; err != nil {
		logger.Error("Failed to toggle proxy host", logger.Err(err), logger.Uint("id", uint(id)), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to toggle proxy host status", err)
		return
	}

	// Update nginx configuration
	if pc.nginxService != nil {
		if proxyHost.Enabled {
			if err := pc.applyProxyHostConfig(&proxyHost); err != nil {
				logger.Error("Failed to apply nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", proxyHost.ID))
			}
		} else {
			if err := pc.removeProxyHostConfig(&proxyHost); err != nil {
				logger.Error("Failed to remove nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", proxyHost.ID))
			}
		}
	}

	action := "disabled"
	if proxyHost.Enabled {
		action = "enabled"
	}

	logger.Info("Proxy host toggled successfully", logger.Uint("id", uint(id)), logger.Uint("user_id", userID), logger.Bool("enabled", proxyHost.Enabled))
	response.SuccessJSONWithLog(c, gin.H{
		"id":      proxyHost.ID,
		"enabled": proxyHost.Enabled,
	}, "Proxy host "+action+" successfully")
}

// BulkToggle toggles multiple proxy hosts
func (pc *ProxyHostController) BulkToggle(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	var req struct {
		IDs     []uint `json:"ids" binding:"required,min=1"`
		Enabled bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request payload", err)
		return
	}

	db := database.GetDB()
	// Update proxy hosts
	result := db.Where("id IN ? AND user_id = ?", req.IDs, userID).
		Updates(models.ProxyHost{Enabled: req.Enabled})

	if result.Error != nil {
		logger.Error("Failed to bulk toggle proxy hosts", logger.Err(result.Error), logger.Uint("user_id", userID))
		response.InternalServerErrorJSONWithLog(c, "Failed to update proxy hosts", result.Error)
		return
	}

	// Get updated proxy hosts for nginx config update
	if pc.nginxService != nil {
		var proxyHosts []models.ProxyHost
		if err := db.Where("id IN ? AND user_id = ?", req.IDs, userID).Find(&proxyHosts).Error; err != nil {
			logger.Error("Failed to fetch updated proxy hosts", logger.Err(err), logger.Uint("user_id", userID))
			// Continue anyway
		} else {
			// Update nginx configurations
			for _, host := range proxyHosts {
				if host.Enabled {
					if err := pc.applyProxyHostConfig(&host); err != nil {
						logger.Error("Failed to apply nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", host.ID))
					}
				} else {
					if err := pc.removeProxyHostConfig(&host); err != nil {
						logger.Error("Failed to remove nginx configuration", logger.Err(err), logger.Uint("proxy_host_id", host.ID))
					}
				}
			}
		}
	}

	action := "disabled"
	if req.Enabled {
		action = "enabled"
	}

	logger.Info("Proxy hosts bulk toggled successfully", logger.Int64("count", result.RowsAffected), logger.Uint("user_id", userID), logger.Bool("enabled", req.Enabled))
	response.SuccessJSONWithLog(c, gin.H{
		"updated": result.RowsAffected,
		"enabled": req.Enabled,
	}, strconv.FormatInt(result.RowsAffected, 10)+" proxy hosts "+action+" successfully")
}

// validateDomainNames validates a list of domain names
func (pc *ProxyHostController) validateDomainNames(domains []string) error {
	if len(domains) == 0 {
		return errors.New("at least one domain name is required")
	}

	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			return errors.New("domain name cannot be empty")
		}

		// Basic domain validation (you can add more sophisticated validation)
		if len(domain) > 253 {
			return errors.New("domain name too long: " + domain)
		}

		if strings.Contains(domain, " ") {
			return errors.New("domain name cannot contain spaces: " + domain)
		}
	}

	return nil
}

// checkDuplicateDomains checks if any domain already exists in other proxy hosts
func (pc *ProxyHostController) checkDuplicateDomains(domains []string, excludeID uint) error {
	db := database.GetDB()
	for _, domain := range domains {
		var count int64
		query := db.Model(&models.ProxyHost{}).Where("domain_names LIKE ?", "%"+domain+"%")

		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}

		if err := query.Count(&count).Error; err != nil {
			return errors.New("failed to check domain uniqueness")
		}

		if count > 0 {
			return errors.New("domain already exists: " + domain)
		}
	}

	return nil
}

// Helper methods for nginx configuration (simplified for now)
func (pc *ProxyHostController) generateProxyHostConfig(proxyHost *models.ProxyHost) (string, bool) {
	// This is a simplified implementation
	// In a real implementation, you would generate actual nginx config
	config := "# Generated nginx config for " + proxyHost.GetPrimaryDomain() + "\n"
	config += "server {\n"
	config += "    server_name " + strings.Join(proxyHost.DomainNames, " ") + ";\n"
	config += "    location / {\n"
	config += "        proxy_pass " + proxyHost.GetTargetURL() + ";\n"
	config += "    }\n"
	config += "}\n"

	return config, true
}

func (pc *ProxyHostController) applyProxyHostConfig(proxyHost *models.ProxyHost) error {
	// This is a simplified implementation
	// In a real implementation, you would write the config to nginx sites and reload
	logger.Info("Applying nginx configuration", logger.Uint("proxy_host_id", proxyHost.ID))
	return nil
}

func (pc *ProxyHostController) removeProxyHostConfig(proxyHost *models.ProxyHost) error {
	// This is a simplified implementation
	// In a real implementation, you would remove the config file and reload nginx
	logger.Info("Removing nginx configuration", logger.Uint("proxy_host_id", proxyHost.ID))
	return nil
}
