package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	BaseModel
	Email       string      `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Name        string      `json:"name" gorm:"size:100;not null"`
	Nickname    string      `json:"nickname" gorm:"size:100"`
	Avatar      string      `json:"avatar" gorm:"type:text"`
	Password    string      `json:"-" gorm:"size:255;not null"`
	Roles       StringArray `json:"roles" gorm:"type:text"`
	IsDisabled  bool        `json:"is_disabled" gorm:"default:false"`
	LastLoginAt *time.Time  `json:"last_login_at"`

	// Relationships
	ProxyHosts       []ProxyHost       `json:"proxy_hosts,omitempty" gorm:"foreignKey:UserID"`
	Certificates     []Certificate     `json:"certificates,omitempty" gorm:"foreignKey:UserID"`
	AccessLists      []AccessList      `json:"access_lists,omitempty" gorm:"foreignKey:UserID"`
	RedirectionHosts []RedirectionHost `json:"redirection_hosts,omitempty" gorm:"foreignKey:UserID"`
	Streams          []Stream          `json:"streams,omitempty" gorm:"foreignKey:UserID"`
	DeadHosts        []DeadHost        `json:"dead_hosts,omitempty" gorm:"foreignKey:UserID"`
	AuditLogs        []AuditLog        `json:"audit_logs,omitempty" gorm:"foreignKey:UserID"`
	Tokens           []Token           `json:"tokens,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to hash password
func (u *User) BeforeCreate() error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}

	// Set default role if empty
	if len(u.Roles) == 0 {
		u.Roles = StringArray{string(RoleUser)}
	}

	return nil
}

// BeforeUpdate hook to hash password if changed
func (u *User) BeforeUpdate() error {
	if u.Password != "" {
		// Check if password is already hashed
		if !u.IsPasswordHashed() {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			u.Password = string(hashedPassword)
		}
	}
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// SetPassword hashes and sets the password
func (u *User) SetPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// IsPasswordHashed checks if the password is already hashed
func (u *User) IsPasswordHashed() bool {
	return len(u.Password) == 60 && u.Password[:4] == "$2a$"
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role Role) bool {
	for _, r := range u.Roles {
		if r == string(role) {
			return true
		}
	}
	return false
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// AddRole adds a role to the user if not already present
func (u *User) AddRole(role Role) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, string(role))
	}
}

// RemoveRole removes a role from the user
func (u *User) RemoveRole(role Role) {
	for i, r := range u.Roles {
		if r == string(role) {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			break
		}
	}
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// CanManageUser checks if this user can manage another user
func (u *User) CanManageUser(targetUser *User) bool {
	// Admin can manage all users
	if u.IsAdmin() {
		return true
	}

	// Users can only manage themselves
	return u.ID == targetUser.ID
}
