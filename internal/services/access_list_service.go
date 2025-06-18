package services

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/nguyendkn/nginx-manager/internal/database"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"gorm.io/gorm"
)

var (
	ErrAccessListNotFound = errors.New("access list not found")
	ErrAccessListInUse    = errors.New("access list is currently in use")
	ErrInvalidIPFormat    = errors.New("invalid IP address format")
	ErrInvalidCIDRFormat  = errors.New("invalid CIDR format")
)

// AccessListService handles access list management
type AccessListService struct {
	db          *gorm.DB
	authService *AuthService
}

// NewAccessListService creates a new access list service instance
func NewAccessListService(authService *AuthService) *AccessListService {
	return &AccessListService{
		db:          database.GetDB(),
		authService: authService,
	}
}

// AccessListRequest represents access list create/update request
type AccessListRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Items       []AccessListItemRequest `json:"items"`
}

// AccessListItemRequest represents access list item request
type AccessListItemRequest struct {
	Type      models.AccessListItemType  `json:"type" binding:"required"`
	Directive models.AccessListDirective `json:"directive" binding:"required"`
	Address   string                     `json:"address,omitempty"`
	Subnet    string                     `json:"subnet,omitempty"`
	Username  string                     `json:"username,omitempty"`
	Password  string                     `json:"password,omitempty"`
	Comment   string                     `json:"comment,omitempty"`
	Enabled   bool                       `json:"enabled"`
}

// TestIPRequest represents IP testing request
type TestIPRequest struct {
	IPAddress string `json:"ip_address" binding:"required"`
}

// TestIPResponse represents IP testing response
type TestIPResponse struct {
	IPAddress   string                 `json:"ip_address"`
	Allowed     bool                   `json:"allowed"`
	Message     string                 `json:"message"`
	MatchedRule *models.AccessListItem `json:"matched_rule,omitempty"`
}

// CreateAccessList creates a new access list
func (s *AccessListService) CreateAccessList(userID uint, req *AccessListRequest) (*models.AccessList, error) {
	// Validate request
	if err := s.validateAccessListRequest(req); err != nil {
		return nil, err
	}

	// Create access list
	accessList := &models.AccessList{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create access list
	if err := tx.Create(accessList).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create access list items
	for _, itemReq := range req.Items {
		item := &models.AccessListItem{
			AccessListID: accessList.ID,
			Type:         itemReq.Type,
			Directive:    itemReq.Directive,
			Address:      itemReq.Address,
			Subnet:       itemReq.Subnet,
			Username:     itemReq.Username,
			Password:     itemReq.Password,
			Comment:      itemReq.Comment,
			Enabled:      itemReq.Enabled,
		}

		// Hash password for auth items
		if item.Type == models.AccessListItemTypeAuth && item.Password != "" {
			if err := item.SetPassword(item.Password); err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load the complete access list with items
	if err := s.db.Preload("Items").Find(accessList).Error; err != nil {
		return nil, err
	}

	return accessList, nil
}

// UpdateAccessList updates an existing access list
func (s *AccessListService) UpdateAccessList(userID uint, id uint, req *AccessListRequest) (*models.AccessList, error) {
	// Find existing access list
	var accessList models.AccessList
	if err := s.db.Preload("Items").Where("id = ? AND user_id = ?", id, userID).First(&accessList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccessListNotFound
		}
		return nil, err
	}

	// Check admin permission for cross-user management
	if accessList.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return nil, err
		}
	}

	// Validate request
	if err := s.validateAccessListRequest(req); err != nil {
		return nil, err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update access list
	accessList.Name = req.Name
	accessList.Description = req.Description

	if err := tx.Save(&accessList).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Delete existing items
	if err := tx.Where("access_list_id = ?", accessList.ID).Delete(&models.AccessListItem{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create new items
	for _, itemReq := range req.Items {
		item := &models.AccessListItem{
			AccessListID: accessList.ID,
			Type:         itemReq.Type,
			Directive:    itemReq.Directive,
			Address:      itemReq.Address,
			Subnet:       itemReq.Subnet,
			Username:     itemReq.Username,
			Password:     itemReq.Password,
			Comment:      itemReq.Comment,
			Enabled:      itemReq.Enabled,
		}

		// Hash password for auth items
		if item.Type == models.AccessListItemTypeAuth && item.Password != "" {
			if err := item.SetPassword(item.Password); err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Load the updated access list with items
	if err := s.db.Preload("Items").Find(&accessList).Error; err != nil {
		return nil, err
	}

	return &accessList, nil
}

// DeleteAccessList deletes an access list
func (s *AccessListService) DeleteAccessList(userID uint, id uint) error {
	// Find access list
	var accessList models.AccessList
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&accessList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAccessListNotFound
		}
		return err
	}

	// Check admin permission for cross-user management
	if accessList.UserID != userID {
		if err := s.authService.RequireAdmin(userID); err != nil {
			return err
		}
	}

	// Check if access list is in use
	var count int64
	if err := s.db.Model(&models.ProxyHost{}).Where("access_list_id = ?", id).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return ErrAccessListInUse
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete access list items first
	if err := tx.Where("access_list_id = ?", id).Delete(&models.AccessListItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete access list
	if err := tx.Delete(&accessList).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// GetAccessList gets a single access list
func (s *AccessListService) GetAccessList(userID uint, id uint) (*models.AccessList, error) {
	var accessList models.AccessList
	query := s.db.Preload("Items")

	// Check if user has admin privileges to view all access lists
	if err := s.authService.RequireAdmin(userID); err != nil {
		// Non-admin users can only see their own access lists
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Where("id = ?", id).First(&accessList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccessListNotFound
		}
		return nil, err
	}

	return &accessList, nil
}

// ListAccessLists gets access lists with pagination
func (s *AccessListService) ListAccessLists(userID uint, offset, limit int) ([]models.AccessList, int64, error) {
	var accessLists []models.AccessList
	var total int64

	query := s.db.Model(&models.AccessList{}).Preload("Items")

	// Check if user has admin privileges to view all access lists
	if err := s.authService.RequireAdmin(userID); err != nil {
		// Non-admin users can only see their own access lists
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&accessLists).Error; err != nil {
		return nil, 0, err
	}

	return accessLists, total, nil
}

// TestIP tests an IP address against an access list
func (s *AccessListService) TestIP(userID uint, id uint, req *TestIPRequest) (*TestIPResponse, error) {
	// Get access list
	accessList, err := s.GetAccessList(userID, id)
	if err != nil {
		return nil, err
	}

	// Validate IP address
	if net.ParseIP(req.IPAddress) == nil {
		return &TestIPResponse{
			IPAddress: req.IPAddress,
			Allowed:   false,
			Message:   "Invalid IP address format",
		}, nil
	}

	// Test IP access
	allowed, err := accessList.CheckIPAccess(req.IPAddress)
	if err != nil {
		return &TestIPResponse{
			IPAddress: req.IPAddress,
			Allowed:   false,
			Message:   fmt.Sprintf("Error checking IP access: %v", err),
		}, nil
	}

	// Find matching rule (if any)
	var matchedRule *models.AccessListItem
	for _, item := range accessList.GetEnabledItems() {
		if item.IsIPItem() {
			if matches, _ := item.MatchesIP(req.IPAddress); matches {
				matchedRule = &item
				break
			}
		}
	}

	message := "No matching rules found"
	if matchedRule != nil {
		if allowed {
			message = fmt.Sprintf("Allowed by rule: %s %s", matchedRule.Directive, matchedRule.GetDisplayName())
		} else {
			message = fmt.Sprintf("Denied by rule: %s %s", matchedRule.Directive, matchedRule.GetDisplayName())
		}
	} else if allowed {
		message = "Allowed by default (no deny rules matched)"
	} else {
		message = "Denied by default (allow rules exist but none matched)"
	}

	return &TestIPResponse{
		IPAddress:   req.IPAddress,
		Allowed:     allowed,
		Message:     message,
		MatchedRule: matchedRule,
	}, nil
}

// ValidateAccessList validates an access list configuration
func (s *AccessListService) ValidateAccessList(userID uint, id uint) ([]string, error) {
	// Get access list
	accessList, err := s.GetAccessList(userID, id)
	if err != nil {
		return nil, err
	}

	// Validate rules
	return accessList.ValidateRules(), nil
}

// ExportAccessList exports access list rules in nginx format
func (s *AccessListService) ExportAccessList(userID uint, id uint) (string, error) {
	// Get access list
	accessList, err := s.GetAccessList(userID, id)
	if err != nil {
		return "", err
	}

	var config strings.Builder
	config.WriteString(fmt.Sprintf("# Access List: %s\n", accessList.Name))
	if accessList.Description != "" {
		config.WriteString(fmt.Sprintf("# Description: %s\n", accessList.Description))
	}
	config.WriteString("\n")

	// Generate IP rules
	for _, item := range accessList.GetEnabledItems() {
		if item.IsIPItem() {
			if rule := item.GetNginxRule(); rule != "" {
				if item.Comment != "" {
					config.WriteString(fmt.Sprintf("# %s\n", item.Comment))
				}
				config.WriteString(rule + "\n")
			}
		}
	}

	// Generate auth rules
	authItems := []models.AccessListItem{}
	for _, item := range accessList.GetEnabledItems() {
		if item.IsAuthItem() {
			authItems = append(authItems, item)
		}
	}

	if len(authItems) > 0 {
		config.WriteString("\n# HTTP Authentication\n")
		config.WriteString("auth_basic \"Restricted Area\";\n")
		config.WriteString("auth_basic_user_file /etc/nginx/.htpasswd;\n")
	}

	return config.String(), nil
}

// ImportAccessList imports access list rules from nginx configuration
func (s *AccessListService) ImportAccessList(userID uint, name string, config string) (*models.AccessList, error) {
	// Parse nginx configuration and create access list
	// This is a simplified implementation
	items := []AccessListItemRequest{}

	lines := strings.Split(config, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse allow/deny rules
		if strings.HasPrefix(line, "allow ") || strings.HasPrefix(line, "deny ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				directive := models.AccessListDirective(parts[0])
				address := strings.TrimSuffix(parts[1], ";")

				itemType := models.AccessListItemTypeIP
				itemReq := AccessListItemRequest{
					Type:      itemType,
					Directive: directive,
					Address:   address,
					Enabled:   true,
				}

				// Check if it's a CIDR notation
				if strings.Contains(address, "/") {
					itemReq.Type = models.AccessListItemTypeCIDR
					itemReq.Subnet = address
					itemReq.Address = ""
				}

				items = append(items, itemReq)
			}
		}
	}

	// Create access list request
	req := &AccessListRequest{
		Name:        name,
		Description: "Imported from nginx configuration",
		Items:       items,
	}

	return s.CreateAccessList(userID, req)
}

// validateAccessListRequest validates an access list request
func (s *AccessListService) validateAccessListRequest(req *AccessListRequest) error {
	if req.Name == "" {
		return errors.New("access list name is required")
	}

	for i, item := range req.Items {
		if !item.Type.IsValid() {
			return fmt.Errorf("item %d: invalid type", i+1)
		}

		if !item.Directive.IsValid() {
			return fmt.Errorf("item %d: invalid directive", i+1)
		}

		switch item.Type {
		case models.AccessListItemTypeIP:
			if item.Address == "" {
				return fmt.Errorf("item %d: IP address is required", i+1)
			}
			if net.ParseIP(item.Address) == nil {
				return fmt.Errorf("item %d: invalid IP address format", i+1)
			}

		case models.AccessListItemTypeCIDR:
			if item.Subnet == "" {
				return fmt.Errorf("item %d: subnet is required", i+1)
			}
			if _, _, err := net.ParseCIDR(item.Subnet); err != nil {
				return fmt.Errorf("item %d: invalid CIDR format", i+1)
			}

		case models.AccessListItemTypeAuth:
			if item.Username == "" {
				return fmt.Errorf("item %d: username is required", i+1)
			}
			if item.Password == "" {
				return fmt.Errorf("item %d: password is required", i+1)
			}
		}
	}

	return nil
}
