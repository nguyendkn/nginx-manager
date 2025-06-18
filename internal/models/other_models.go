package models

import "time"

// Stream represents a TCP/UDP stream proxy
type Stream struct {
	BaseModel
	IncomingPort   int    `json:"incoming_port" gorm:"not null;uniqueIndex"`
	ForwardingHost string `json:"forwarding_host" gorm:"size:255;not null"`
	ForwardingPort int    `json:"forwarding_port" gorm:"not null"`
	TCP            bool   `json:"tcp" gorm:"default:true"`
	UDP            bool   `json:"udp" gorm:"default:false"`
	Enabled        bool   `json:"enabled" gorm:"default:true"`
	CertificateID  *uint  `json:"certificate_id" gorm:"index"`
	SSLTermination bool   `json:"ssl_termination" gorm:"default:false"`
	UserID         uint   `json:"user_id" gorm:"not null;index"`

	// Relationships
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Certificate *Certificate `json:"certificate,omitempty" gorm:"foreignKey:CertificateID"`
}

// TableName specifies the table name for Stream model
func (Stream) TableName() string {
	return "streams"
}

// RedirectionHost represents a redirection host
type RedirectionHost struct {
	BaseModel
	DomainNames       StringArray `json:"domain_names" gorm:"type:text"`
	ForwardScheme     string      `json:"forward_scheme" gorm:"size:10;not null"`
	ForwardDomainName string      `json:"forward_domain_name" gorm:"size:255;not null"`
	StatusCode        int         `json:"status_code" gorm:"default:301"`
	PreservePath      bool        `json:"preserve_path" gorm:"default:true"`
	Enabled           bool        `json:"enabled" gorm:"default:true"`
	CertificateID     *uint       `json:"certificate_id" gorm:"index"`
	AdvancedConfig    string      `json:"advanced_config" gorm:"type:text"`
	Meta              JSON        `json:"meta" gorm:"type:json"`
	UserID            uint        `json:"user_id" gorm:"not null;index"`

	// Relationships
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Certificate *Certificate `json:"certificate,omitempty" gorm:"foreignKey:CertificateID"`
}

// TableName specifies the table name for RedirectionHost model
func (RedirectionHost) TableName() string {
	return "redirection_hosts"
}

// DeadHost represents a 404 host configuration
type DeadHost struct {
	BaseModel
	DomainNames   StringArray `json:"domain_names" gorm:"type:text"`
	CertificateID *uint       `json:"certificate_id" gorm:"index"`
	Enabled       bool        `json:"enabled" gorm:"default:true"`
	Meta          JSON        `json:"meta" gorm:"type:json"`
	UserID        uint        `json:"user_id" gorm:"not null;index"`

	// Relationships
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Certificate *Certificate `json:"certificate,omitempty" gorm:"foreignKey:CertificateID"`
}

// TableName specifies the table name for DeadHost model
func (DeadHost) TableName() string {
	return "dead_hosts"
}

// AuditLog represents an audit log entry
type AuditLog struct {
	BaseModel
	UserID      uint        `json:"user_id" gorm:"not null;index"`
	Action      AuditAction `json:"action" gorm:"size:50;not null"`
	ObjectType  ObjectType  `json:"object_type" gorm:"size:50;not null"`
	ObjectID    uint        `json:"object_id" gorm:"not null;index"`
	Description string      `json:"description" gorm:"type:text"`
	IPAddress   string      `json:"ip_address" gorm:"size:45"`
	UserAgent   string      `json:"user_agent" gorm:"type:text"`
	Meta        JSON        `json:"meta" gorm:"type:json"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Token represents an API token
type Token struct {
	BaseModel
	Name      string     `json:"name" gorm:"size:255;not null"`
	Secret    string     `json:"secret" gorm:"size:255;not null;uniqueIndex"`
	ExpiresAt *time.Time `json:"expires_at"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	UserID    uint       `json:"user_id" gorm:"not null;index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for Token model
func (Token) TableName() string {
	return "tokens"
}

// Setting represents system settings
type Setting struct {
	BaseModel
	ID    string `json:"id" gorm:"primaryKey;size:255"`
	Name  string `json:"name" gorm:"size:255;not null"`
	Value JSON   `json:"value" gorm:"type:json"`
	Meta  JSON   `json:"meta" gorm:"type:json"`
}

// TableName specifies the table name for Setting model
func (Setting) TableName() string {
	return "settings"
}

// Helper methods for Stream
func (s *Stream) IsSSLEnabled() bool {
	return s.CertificateID != nil && *s.CertificateID > 0 && s.SSLTermination
}

func (s *Stream) IsTCPEnabled() bool {
	return s.TCP
}

func (s *Stream) IsUDPEnabled() bool {
	return s.UDP
}

// Helper methods for RedirectionHost
func (r *RedirectionHost) GetPrimaryDomain() string {
	if len(r.DomainNames) > 0 {
		return r.DomainNames[0]
	}
	return ""
}

func (r *RedirectionHost) IsSSLEnabled() bool {
	return r.CertificateID != nil && *r.CertificateID > 0
}

// Helper methods for DeadHost
func (d *DeadHost) GetPrimaryDomain() string {
	if len(d.DomainNames) > 0 {
		return d.DomainNames[0]
	}
	return ""
}

func (d *DeadHost) IsSSLEnabled() bool {
	return d.CertificateID != nil && *d.CertificateID > 0
}

// Helper methods for Token
func (t *Token) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

func (t *Token) IsValid() bool {
	return t.IsActive && !t.IsExpired()
}

// Helper methods for Setting
func (s *Setting) GetValue() interface{} {
	if s.Value != nil {
		if value, exists := s.Value["value"]; exists {
			return value
		}
	}
	return nil
}

func (s *Setting) SetValue(value interface{}) {
	if s.Value == nil {
		s.Value = make(JSON)
	}
	s.Value["value"] = value
}
