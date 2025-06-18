package models

import (
	"errors"
	"fmt"
	"net"
)

// AccessList represents an access control list
type AccessList struct {
	BaseModel
	Name        string           `json:"name" gorm:"size:255;not null"`
	Description string           `json:"description" gorm:"type:text"`
	Items       []AccessListItem `json:"items" gorm:"foreignKey:AccessListID"`
	UserID      uint             `json:"user_id" gorm:"not null;index"`

	// Relationships
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ProxyHosts []ProxyHost `json:"proxy_hosts,omitempty" gorm:"foreignKey:AccessListID"`
}

// TableName specifies the table name for AccessList model
func (AccessList) TableName() string {
	return "access_lists"
}

// AccessListItem represents an individual access control rule
type AccessListItem struct {
	BaseModel
	AccessListID uint                `json:"access_list_id" gorm:"not null;index"`
	Type         AccessListItemType  `json:"type" gorm:"size:20;not null"`
	Directive    AccessListDirective `json:"directive" gorm:"size:10;not null"`

	// For IP-based rules
	Address string `json:"address,omitempty" gorm:"size:255"`
	Subnet  string `json:"subnet,omitempty" gorm:"size:255"`

	// For authentication-based rules
	Username string `json:"username,omitempty" gorm:"size:255"`
	Password string `json:"password,omitempty" gorm:"size:255"`

	// Additional configuration
	Comment string `json:"comment,omitempty" gorm:"type:text"`
	Enabled bool   `json:"enabled" gorm:"default:true"`

	// Relationship
	AccessList AccessList `json:"access_list,omitempty" gorm:"foreignKey:AccessListID"`
}

// TableName specifies the table name for AccessListItem model
func (AccessListItem) TableName() string {
	return "access_list_items"
}

// AccessListItemType represents the type of access control rule
type AccessListItemType string

const (
	AccessListItemTypeIP   AccessListItemType = "ip"
	AccessListItemTypeAuth AccessListItemType = "auth"
	AccessListItemTypeCIDR AccessListItemType = "cidr"
)

// IsValid checks if the access list item type is valid
func (t AccessListItemType) IsValid() bool {
	switch t {
	case AccessListItemTypeIP, AccessListItemTypeAuth, AccessListItemTypeCIDR:
		return true
	default:
		return false
	}
}

// AccessListDirective represents allow or deny directive
type AccessListDirective string

const (
	AccessListDirectiveAllow AccessListDirective = "allow"
	AccessListDirectiveDeny  AccessListDirective = "deny"
)

// IsValid checks if the access list directive is valid
func (d AccessListDirective) IsValid() bool {
	switch d {
	case AccessListDirectiveAllow, AccessListDirectiveDeny:
		return true
	default:
		return false
	}
}

// Access List Model Methods

// IsEmpty checks if the access list has no items
func (al *AccessList) IsEmpty() bool {
	return len(al.Items) == 0
}

// HasIPRules checks if the access list contains IP-based rules
func (al *AccessList) HasIPRules() bool {
	for _, item := range al.Items {
		if item.Type == AccessListItemTypeIP || item.Type == AccessListItemTypeCIDR {
			return true
		}
	}
	return false
}

// HasAuthRules checks if the access list contains authentication-based rules
func (al *AccessList) HasAuthRules() bool {
	for _, item := range al.Items {
		if item.Type == AccessListItemTypeAuth {
			return true
		}
	}
	return false
}

// GetEnabledItems returns only enabled access list items
func (al *AccessList) GetEnabledItems() []AccessListItem {
	var enabledItems []AccessListItem
	for _, item := range al.Items {
		if item.Enabled {
			enabledItems = append(enabledItems, item)
		}
	}
	return enabledItems
}

// CheckIPAccess checks if an IP address is allowed or denied
func (al *AccessList) CheckIPAccess(ipAddress string) (bool, error) {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false, ErrInvalidIPAddress
	}

	hasAllowRules := false
	hasExplicitAllow := false
	hasExplicitDeny := false

	for _, item := range al.GetEnabledItems() {
		if item.Type != AccessListItemTypeIP && item.Type != AccessListItemTypeCIDR {
			continue
		}

		allowed, err := item.MatchesIP(ipAddress)
		if err != nil {
			continue
		}

		if allowed {
			if item.Directive == AccessListDirectiveAllow {
				hasAllowRules = true
				hasExplicitAllow = true
			} else if item.Directive == AccessListDirectiveDeny {
				hasExplicitDeny = true
				break // Deny rules take precedence
			}
		}
	}

	// If there's an explicit deny, access is denied
	if hasExplicitDeny {
		return false, nil
	}

	// If there are allow rules but no explicit allow, access is denied
	if hasAllowRules && !hasExplicitAllow {
		return false, nil
	}

	// If there are no allow rules, access is allowed (default behavior)
	// If there's an explicit allow, access is allowed
	return true, nil
}

// ValidateRules validates all access list items
func (al *AccessList) ValidateRules() []string {
	var errors []string

	for i, item := range al.Items {
		if errs := item.Validate(); len(errs) > 0 {
			for _, err := range errs {
				errors = append(errors, fmt.Sprintf("Item %d: %s", i+1, err))
			}
		}
	}

	return errors
}

// Access List Item Methods

// MatchesIP checks if an IP address matches this access list item
func (ali *AccessListItem) MatchesIP(ipAddress string) (bool, error) {
	if ali.Type != AccessListItemTypeIP && ali.Type != AccessListItemTypeCIDR {
		return false, ErrInvalidAccessListItemType
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false, ErrInvalidIPAddress
	}

	switch ali.Type {
	case AccessListItemTypeIP:
		// Direct IP match
		targetIP := net.ParseIP(ali.Address)
		if targetIP == nil {
			return false, ErrInvalidIPAddress
		}
		return ip.Equal(targetIP), nil

	case AccessListItemTypeCIDR:
		// CIDR range match
		_, cidr, err := net.ParseCIDR(ali.Subnet)
		if err != nil {
			return false, err
		}
		return cidr.Contains(ip), nil

	default:
		return false, ErrInvalidAccessListItemType
	}
}

// Validate validates an access list item
func (ali *AccessListItem) Validate() []string {
	var errors []string

	// Validate type
	if !ali.Type.IsValid() {
		errors = append(errors, "invalid access list item type")
	}

	// Validate directive
	if !ali.Directive.IsValid() {
		errors = append(errors, "invalid access list directive")
	}

	// Type-specific validation
	switch ali.Type {
	case AccessListItemTypeIP:
		if ali.Address == "" {
			errors = append(errors, "IP address is required for IP type")
		} else if net.ParseIP(ali.Address) == nil {
			errors = append(errors, "invalid IP address format")
		}

	case AccessListItemTypeCIDR:
		if ali.Subnet == "" {
			errors = append(errors, "subnet is required for CIDR type")
		} else if _, _, err := net.ParseCIDR(ali.Subnet); err != nil {
			errors = append(errors, "invalid CIDR format")
		}

	case AccessListItemTypeAuth:
		if ali.Username == "" {
			errors = append(errors, "username is required for auth type")
		}
		if ali.Password == "" {
			errors = append(errors, "password is required for auth type")
		}
	}

	return errors
}

// GetDisplayName returns a human-readable name for the access list item
func (ali *AccessListItem) GetDisplayName() string {
	switch ali.Type {
	case AccessListItemTypeIP:
		return ali.Address
	case AccessListItemTypeCIDR:
		return ali.Subnet
	case AccessListItemTypeAuth:
		return ali.Username
	default:
		return "Unknown"
	}
}

// GetNginxRule generates the nginx configuration rule for this item
func (ali *AccessListItem) GetNginxRule() string {
	if !ali.Enabled {
		return ""
	}

	switch ali.Type {
	case AccessListItemTypeIP:
		return fmt.Sprintf("%s %s;", ali.Directive, ali.Address)
	case AccessListItemTypeCIDR:
		return fmt.Sprintf("%s %s;", ali.Directive, ali.Subnet)
	case AccessListItemTypeAuth:
		// Auth rules are handled differently in nginx (auth_basic)
		return ""
	default:
		return ""
	}
}

// IsAuthItem checks if this is an authentication-based item
func (ali *AccessListItem) IsAuthItem() bool {
	return ali.Type == AccessListItemTypeAuth
}

// IsIPItem checks if this is an IP-based item
func (ali *AccessListItem) IsIPItem() bool {
	return ali.Type == AccessListItemTypeIP || ali.Type == AccessListItemTypeCIDR
}

// SetPassword sets and hashes the password for auth items
func (ali *AccessListItem) SetPassword(password string) error {
	if ali.Type != AccessListItemTypeAuth {
		return ErrInvalidAccessListItemType
	}

	// In a real implementation, this would hash the password
	// For now, we'll store it as-is (should be hashed in production)
	ali.Password = password
	return nil
}

// CheckPassword validates the password for auth items
func (ali *AccessListItem) CheckPassword(password string) bool {
	if ali.Type != AccessListItemTypeAuth {
		return false
	}

	// In a real implementation, this would compare hashed passwords
	return ali.Password == password
}

// Error definitions
var (
	ErrInvalidIPAddress          = errors.New("invalid IP address")
	ErrInvalidAccessListItemType = errors.New("invalid access list item type")
)
