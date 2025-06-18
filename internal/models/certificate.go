package models

import (
	"time"
)

// Certificate represents an SSL certificate
type Certificate struct {
	BaseModel
	Name                    string              `json:"name" gorm:"size:255;not null"`
	NiceName                string              `json:"nice_name" gorm:"size:255"`
	Provider                CertificateProvider `json:"provider" gorm:"size:50;not null"`
	DomainNames             StringArray         `json:"domain_names" gorm:"type:text"`
	ExpiresOn               *time.Time          `json:"expires_on"`
	Status                  string              `json:"status" gorm:"size:50;default:'pending'"`
	HasValidation           bool                `json:"has_validation" gorm:"default:false"`
	Certificate             string              `json:"certificate" gorm:"type:longtext"`
	CertificateKey          string              `json:"certificate_key" gorm:"type:longtext"`
	IntermediateCertificate string              `json:"intermediate_certificate" gorm:"type:longtext"`
	Meta                    JSON                `json:"meta" gorm:"type:json"`
	UserID                  uint                `json:"user_id" gorm:"not null;index"`

	// Relationships
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ProxyHosts []ProxyHost `json:"proxy_hosts,omitempty" gorm:"foreignKey:CertificateID"`
	Streams    []Stream    `json:"streams,omitempty" gorm:"foreignKey:CertificateID"`
}

// TableName specifies the table name for Certificate model
func (Certificate) TableName() string {
	return "certificates"
}

// IsExpired checks if the certificate is expired
func (c *Certificate) IsExpired() bool {
	if c.ExpiresOn == nil {
		return true
	}
	return time.Now().After(*c.ExpiresOn)
}

// IsExpiringSoon checks if the certificate expires within the given duration
func (c *Certificate) IsExpiringSoon(within time.Duration) bool {
	if c.ExpiresOn == nil {
		return true
	}
	return time.Now().Add(within).After(*c.ExpiresOn)
}

// DaysUntilExpiry returns the number of days until the certificate expires
func (c *Certificate) DaysUntilExpiry() int {
	if c.ExpiresOn == nil {
		return 0
	}
	duration := time.Until(*c.ExpiresOn)
	return int(duration.Hours() / 24)
}

// IsValid checks if the certificate has valid certificate and key
func (c *Certificate) IsValid() bool {
	return c.Certificate != "" && c.CertificateKey != ""
}

// IsLetsEncrypt checks if this is a Let's Encrypt certificate
func (c *Certificate) IsLetsEncrypt() bool {
	return c.Provider == ProviderLetsEncrypt
}

// IsCustom checks if this is a custom certificate
func (c *Certificate) IsCustom() bool {
	return c.Provider == ProviderCustom
}

// CanRenew checks if the certificate can be renewed
func (c *Certificate) CanRenew() bool {
	// Only Let's Encrypt certificates can be auto-renewed
	if !c.IsLetsEncrypt() {
		return false
	}

	// Should renew if expires within 30 days
	return c.IsExpiringSoon(30 * 24 * time.Hour)
}

// GetPrimaryDomain returns the first domain name (primary domain)
func (c *Certificate) GetPrimaryDomain() string {
	if len(c.DomainNames) > 0 {
		return c.DomainNames[0]
	}
	return ""
}

// HasDomain checks if the certificate covers a specific domain
func (c *Certificate) HasDomain(domain string) bool {
	for _, d := range c.DomainNames {
		if d == domain {
			return true
		}
	}
	return false
}

// AddDomain adds a domain to the certificate if not already present
func (c *Certificate) AddDomain(domain string) {
	if !c.HasDomain(domain) {
		c.DomainNames = append(c.DomainNames, domain)
	}
}

// RemoveDomain removes a domain from the certificate
func (c *Certificate) RemoveDomain(domain string) {
	for i, d := range c.DomainNames {
		if d == domain {
			c.DomainNames = append(c.DomainNames[:i], c.DomainNames[i+1:]...)
			break
		}
	}
}

// SetExpiryFromCertificate parses the certificate and sets the expiry date
func (c *Certificate) SetExpiryFromCertificate() error {
	// This would implement certificate parsing to extract expiry date
	// For now, we'll set it to a default value for Let's Encrypt (90 days)
	if c.IsLetsEncrypt() {
		expiry := time.Now().Add(90 * 24 * time.Hour)
		c.ExpiresOn = &expiry
	}
	return nil
}

// GetMetaValue gets a value from the meta JSON field
func (c *Certificate) GetMetaValue(key string) interface{} {
	if c.Meta != nil {
		return c.Meta[key]
	}
	return nil
}

// SetMetaValue sets a value in the meta JSON field
func (c *Certificate) SetMetaValue(key string, value interface{}) {
	if c.Meta == nil {
		c.Meta = make(JSON)
	}
	c.Meta[key] = value
}

// ClearSensitiveData removes sensitive data (private key) from the model
func (c *Certificate) ClearSensitiveData() {
	c.CertificateKey = ""
}

// DomainTestResult represents individual domain test result
type DomainTestResult struct {
	Domain       string `json:"domain"`
	Reachable    bool   `json:"reachable"`
	SSL          bool   `json:"ssl"`
	Port80       bool   `json:"port_80"`
	Port443      bool   `json:"port_443"`
	Message      string `json:"message"`
	ResponseTime int64  `json:"response_time_ms"`
}
