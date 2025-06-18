package models

// ProxyHost represents a proxy host configuration
type ProxyHost struct {
	BaseModel
	DomainNames           StringArray   `json:"domain_names" gorm:"type:text"`
	ForwardScheme         ForwardScheme `json:"forward_scheme" gorm:"size:10;not null"`
	ForwardHost           string        `json:"forward_host" gorm:"size:255;not null"`
	ForwardPort           int           `json:"forward_port" gorm:"not null"`
	AccessListID          *uint         `json:"access_list_id" gorm:"index"`
	CertificateID         *uint         `json:"certificate_id" gorm:"index"`
	SSLForced             bool          `json:"ssl_forced" gorm:"default:false"`
	CachingEnabled        bool          `json:"caching_enabled" gorm:"default:false"`
	BlockExploits         bool          `json:"block_exploits" gorm:"default:true"`
	AllowWebsocketUpgrade bool          `json:"allow_websocket_upgrade" gorm:"default:false"`
	HTTP2Support          bool          `json:"http2_support" gorm:"default:true"`
	HSTSEnabled           bool          `json:"hsts_enabled" gorm:"default:false"`
	HSTSSubdomains        bool          `json:"hsts_subdomains" gorm:"default:false"`
	AdvancedConfig        string        `json:"advanced_config" gorm:"type:text"`
	Enabled               bool          `json:"enabled" gorm:"default:true"`
	Locations             JSON          `json:"locations" gorm:"type:json"`
	Meta                  JSON          `json:"meta" gorm:"type:json"`
	UserID                uint          `json:"user_id" gorm:"not null;index"`

	// Relationships
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AccessList  *AccessList  `json:"access_list,omitempty" gorm:"foreignKey:AccessListID"`
	Certificate *Certificate `json:"certificate,omitempty" gorm:"foreignKey:CertificateID"`
}

// TableName specifies the table name for ProxyHost model
func (ProxyHost) TableName() string {
	return "proxy_hosts"
}

// GetPrimaryDomain returns the first domain name (primary domain)
func (p *ProxyHost) GetPrimaryDomain() string {
	if len(p.DomainNames) > 0 {
		return p.DomainNames[0]
	}
	return ""
}

// HasDomain checks if the proxy host covers a specific domain
func (p *ProxyHost) HasDomain(domain string) bool {
	for _, d := range p.DomainNames {
		if d == domain {
			return true
		}
	}
	return false
}

// AddDomain adds a domain to the proxy host if not already present
func (p *ProxyHost) AddDomain(domain string) {
	if !p.HasDomain(domain) {
		p.DomainNames = append(p.DomainNames, domain)
	}
}

// RemoveDomain removes a domain from the proxy host
func (p *ProxyHost) RemoveDomain(domain string) {
	for i, d := range p.DomainNames {
		if d == domain {
			p.DomainNames = append(p.DomainNames[:i], p.DomainNames[i+1:]...)
			break
		}
	}
}

// IsSSLEnabled checks if SSL is enabled for this proxy host
func (p *ProxyHost) IsSSLEnabled() bool {
	return p.CertificateID != nil && *p.CertificateID > 0
}

// GetTargetURL returns the target URL for proxying
func (p *ProxyHost) GetTargetURL() string {
	return string(p.ForwardScheme) + "://" + p.ForwardHost + ":" + string(rune(p.ForwardPort))
}

// HasAccessList checks if an access list is configured
func (p *ProxyHost) HasAccessList() bool {
	return p.AccessListID != nil && *p.AccessListID > 0
}

// GetMetaValue gets a value from the meta JSON field
func (p *ProxyHost) GetMetaValue(key string) interface{} {
	if p.Meta != nil {
		return p.Meta[key]
	}
	return nil
}

// SetMetaValue sets a value in the meta JSON field
func (p *ProxyHost) SetMetaValue(key string, value interface{}) {
	if p.Meta == nil {
		p.Meta = make(JSON)
	}
	p.Meta[key] = value
}

// AddLocation adds a custom location configuration
func (p *ProxyHost) AddLocation(path string, config map[string]interface{}) {
	if p.Locations == nil {
		p.Locations = make(JSON)
	}
	p.Locations[path] = config
}

// RemoveLocation removes a custom location configuration
func (p *ProxyHost) RemoveLocation(path string) {
	if p.Locations != nil {
		delete(p.Locations, path)
	}
}
