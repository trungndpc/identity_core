package domain

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tenant struct {
	BaseModel
	Code     string         `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name     string         `gorm:"size:255;not null" json:"name"`
	Status   string         `gorm:"size:32;not null;default:active;index" json:"status"`
	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`
}

func (Tenant) TableName() string { return "tenants" }

type User struct {
	BaseModel
	TenantID int64 `gorm:"not null;index" json:"tenant_id"`

	FullName    string     `gorm:"size:255" json:"full_name"`
	DisplayName string     `gorm:"size:255" json:"display_name"`
	AvatarURL   string     `gorm:"size:512" json:"avatar_url"`
	Gender      string     `gorm:"size:16" json:"gender"`
	Birthday    *time.Time `json:"birthday,omitempty"`

	Email    string `gorm:"size:255;index" json:"email"`
	Phone    string `gorm:"size:32;index" json:"phone"`
	Address  string `gorm:"size:512" json:"address"`
	City     string `gorm:"size:128" json:"city"`
	District string `gorm:"size:128" json:"district"`
	Ward     string `gorm:"size:128" json:"ward"`

	Username string         `gorm:"size:128;index" json:"username"`
	Status   string         `gorm:"size:32;not null;default:active;index" json:"status"`
	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	Tenant     Tenant         `gorm:"foreignKey:TenantID" json:"-"`
	Identities []UserIdentity `gorm:"foreignKey:UserID" json:"identities,omitempty"`
	Roles      []Role         `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

type UserIdentity struct {
	BaseModel
	UserID         int64          `gorm:"not null;index" json:"user_id"`
	Provider       string         `gorm:"size:64;not null;index" json:"provider"`
	ProviderUserID string         `gorm:"size:255" json:"provider_user_id"`
	Identity       string         `gorm:"size:255;not null;index" json:"identity"`
	PasswordHash   string         `gorm:"size:255" json:"-"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserIdentity) TableName() string { return "user_identities" }

type UserRelationship struct {
	BaseModel
	TenantID         int64          `gorm:"not null;index" json:"tenant_id"`
	FromUserID       int64          `gorm:"not null;index" json:"from_user_id"`
	ToUserID         int64          `gorm:"not null;index" json:"to_user_id"`
	RelationshipType string         `gorm:"size:64;not null;index" json:"relationship_type"`
	Status           string         `gorm:"size:32;not null;default:active;index" json:"status"`
	Metadata         datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`

	FromUser User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}

func (UserRelationship) TableName() string { return "user_relationships" }

type Role struct {
	BaseModel
	TenantID     int64  `gorm:"not null;index:idx_role_tenant_code,unique" json:"tenant_id"`
	Code         string `gorm:"size:64;not null;index:idx_role_tenant_code,unique" json:"code"`
	Name         string `gorm:"size:255;not null" json:"name"`
	Description  string `gorm:"size:512" json:"description"`
	IsSystemRole bool   `gorm:"not null;default:false" json:"is_system_role"`

	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

type UserRole struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index:idx_user_role,unique" json:"user_id"`
	RoleID    int64     `gorm:"not null;index:idx_user_role,unique" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserRole) TableName() string { return "user_roles" }

type Permission struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:128;uniqueIndex;not null" json:"code"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Module      string    `gorm:"size:64;not null;index" json:"module"`
	Description string    `gorm:"size:512" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       int64     `gorm:"not null;index:idx_role_permission,unique" json:"role_id"`
	PermissionID int64     `gorm:"not null;index:idx_role_permission,unique" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermission) TableName() string { return "role_permissions" }
