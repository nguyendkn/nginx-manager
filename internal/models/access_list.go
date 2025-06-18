package models

// AccessList represents an access control list
type AccessList struct {
	BaseModel
	Name      string `json:"name" gorm:"size:255;not null"`
	Directive string `json:"directive" gorm:"size:10;not null;default:'allow'"`
	Address   string `json:"address" gorm:"size:255;not null"`
	UserID    uint   `json:"user_id" gorm:"not null;index"`
	Meta      JSON   `json:"meta" gorm:"type:json"`

	// Relationships
	User              User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ProxyHosts        []ProxyHost        `json:"proxy_hosts,omitempty" gorm:"foreignKey:AccessListID"`
	AccessListAuths   []AccessListAuth   `json:"access_list_auths,omitempty" gorm:"foreignKey:AccessListID"`
	AccessListClients []AccessListClient `json:"access_list_clients,omitempty" gorm:"foreignKey:AccessListID"`
}

// TableName specifies the table name for AccessList model
func (AccessList) TableName() string {
	return "access_lists"
}

// AccessListAuth represents HTTP authentication for access lists
type AccessListAuth struct {
	BaseModel
	AccessListID uint   `json:"access_list_id" gorm:"not null;index"`
	Username     string `json:"username" gorm:"size:255;not null"`
	Password     string `json:"password" gorm:"size:255;not null"`

	// Relationships
	AccessList AccessList `json:"access_list,omitempty" gorm:"foreignKey:AccessListID"`
}

// TableName specifies the table name for AccessListAuth model
func (AccessListAuth) TableName() string {
	return "access_list_auths"
}

// AccessListClient represents client-based access control
type AccessListClient struct {
	BaseModel
	AccessListID uint        `json:"access_list_id" gorm:"not null;index"`
	Address      string      `json:"address" gorm:"size:255;not null"`
	Directive    string      `json:"directive" gorm:"size:10;not null;default:'allow'"`
	PassAuth     bool        `json:"pass_auth" gorm:"default:true"`
	SatisfyMode  SatisfyMode `json:"satisfy" gorm:"size:10;default:'any'"`

	// Relationships
	AccessList AccessList `json:"access_list,omitempty" gorm:"foreignKey:AccessListID"`
}

// TableName specifies the table name for AccessListClient model
func (AccessListClient) TableName() string {
	return "access_list_clients"
}

// IsAllowDirective checks if this is an allow directive
func (a *AccessList) IsAllowDirective() bool {
	return a.Directive == string(DirectiveAllow)
}

// IsDenyDirective checks if this is a deny directive
func (a *AccessList) IsDenyDirective() bool {
	return a.Directive == string(DirectiveDeny)
}

// GetMetaValue gets a value from the meta JSON field
func (a *AccessList) GetMetaValue(key string) interface{} {
	if a.Meta != nil {
		return a.Meta[key]
	}
	return nil
}

// SetMetaValue sets a value in the meta JSON field
func (a *AccessList) SetMetaValue(key string, value interface{}) {
	if a.Meta == nil {
		a.Meta = make(JSON)
	}
	a.Meta[key] = value
}
