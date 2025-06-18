package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// JSON type for storing JSON data in database
type JSON map[string]interface{}

// Scan implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSON", value)
	}

	if len(bytes) == 0 {
		*j = make(JSON)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// StringArray type for storing string arrays in database
type StringArray []string

// Scan implements sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}

	if len(bytes) == 0 {
		*sa = []string{}
		return nil
	}

	return json.Unmarshal(bytes, sa)
}

// Value implements driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	return json.Marshal(sa)
}

// Role represents user roles
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	}
	return false
}

// CertificateProvider represents certificate providers
type CertificateProvider string

const (
	ProviderLetsEncrypt CertificateProvider = "letsencrypt"
	ProviderCustom      CertificateProvider = "custom"
)

// IsValid checks if the certificate provider is valid
func (cp CertificateProvider) IsValid() bool {
	switch cp {
	case ProviderLetsEncrypt, ProviderCustom:
		return true
	}
	return false
}

// ForwardScheme represents forward schemes
type ForwardScheme string

const (
	SchemeHTTP  ForwardScheme = "http"
	SchemeHTTPS ForwardScheme = "https"
)

// IsValid checks if the forward scheme is valid
func (fs ForwardScheme) IsValid() bool {
	switch fs {
	case SchemeHTTP, SchemeHTTPS:
		return true
	}
	return false
}

// AccessDirective represents access control directives
type AccessDirective string

const (
	DirectiveAllow AccessDirective = "allow"
	DirectiveDeny  AccessDirective = "deny"
)

// IsValid checks if the access directive is valid
func (ad AccessDirective) IsValid() bool {
	switch ad {
	case DirectiveAllow, DirectiveDeny:
		return true
	}
	return false
}

// SatisfyMode represents access list satisfy modes
type SatisfyMode string

const (
	SatisfyAny SatisfyMode = "any"
	SatisfyAll SatisfyMode = "all"
)

// IsValid checks if the satisfy mode is valid
func (sm SatisfyMode) IsValid() bool {
	switch sm {
	case SatisfyAny, SatisfyAll:
		return true
	}
	return false
}

// AuditAction represents audit log actions
type AuditAction string

const (
	ActionCreated AuditAction = "created"
	ActionUpdated AuditAction = "updated"
	ActionDeleted AuditAction = "deleted"
	ActionLogin   AuditAction = "login"
	ActionLogout  AuditAction = "logout"
)

// IsValid checks if the audit action is valid
func (aa AuditAction) IsValid() bool {
	switch aa {
	case ActionCreated, ActionUpdated, ActionDeleted, ActionLogin, ActionLogout:
		return true
	}
	return false
}

// ObjectType represents audit log object types
type ObjectType string

const (
	ObjectTypeUser            ObjectType = "user"
	ObjectTypeProxyHost       ObjectType = "proxy_host"
	ObjectTypeCertificate     ObjectType = "certificate"
	ObjectTypeAccessList      ObjectType = "access_list"
	ObjectTypeRedirectionHost ObjectType = "redirection_host"
	ObjectTypeStream          ObjectType = "stream"
	ObjectTypeDeadHost        ObjectType = "dead_host"
	ObjectTypeSetting         ObjectType = "setting"
)

// IsValid checks if the object type is valid
func (ot ObjectType) IsValid() bool {
	switch ot {
	case ObjectTypeUser, ObjectTypeProxyHost, ObjectTypeCertificate,
		ObjectTypeAccessList, ObjectTypeRedirectionHost, ObjectTypeStream,
		ObjectTypeDeadHost, ObjectTypeSetting:
		return true
	}
	return false
}
