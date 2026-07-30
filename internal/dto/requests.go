package dto

import (
	"encoding/json"
	"time"
)

type CreateTenantRequest struct {
	Code     string          `json:"code" binding:"required,min=2,max=64"`
	Name     string          `json:"name" binding:"required,min=1,max=255"`
	Status   string          `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	Metadata json.RawMessage `json:"metadata"`
}

type UpdateTenantRequest struct {
	Name     *string         `json:"name" binding:"omitempty,min=1,max=255"`
	Status   *string         `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	Metadata json.RawMessage `json:"metadata"`
}

type CreateUserRequest struct {
	FullName    string          `json:"full_name"`
	DisplayName string          `json:"display_name"`
	AvatarURL   string          `json:"avatar_url"`
	Gender      string          `json:"gender"`
	Birthday    *time.Time      `json:"birthday"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Address     string          `json:"address"`
	City        string          `json:"city"`
	District    string          `json:"district"`
	Ward        string          `json:"ward"`
	Username    string          `json:"username"`
	Status      string          `json:"status" binding:"omitempty,oneof=active inactive banned"`
	Metadata    json.RawMessage `json:"metadata"`
	Password    string          `json:"password"`
	RoleIDs     []int64         `json:"role_ids"`
}

type UpdateUserRequest struct {
	FullName    *string         `json:"full_name"`
	DisplayName *string         `json:"display_name"`
	AvatarURL   *string         `json:"avatar_url"`
	Gender      *string         `json:"gender"`
	Birthday    *time.Time      `json:"birthday"`
	Email       *string         `json:"email"`
	Phone       *string         `json:"phone"`
	Address     *string         `json:"address"`
	City        *string         `json:"city"`
	District    *string         `json:"district"`
	Ward        *string         `json:"ward"`
	Username    *string         `json:"username"`
	Status      *string         `json:"status" binding:"omitempty,oneof=active inactive banned"`
	Metadata    json.RawMessage `json:"metadata"`
}

type CreateIdentityRequest struct {
	Provider       string          `json:"provider" binding:"required"`
	ProviderUserID string          `json:"provider_user_id"`
	Identity       string          `json:"identity" binding:"required"`
	Password       string          `json:"password"`
	Metadata       json.RawMessage `json:"metadata"`
}

type CreateRelationshipRequest struct {
	FromUserID       int64           `json:"from_user_id" binding:"required,gt=0"`
	ToUserID         int64           `json:"to_user_id" binding:"required,gt=0"`
	RelationshipType string          `json:"relationship_type" binding:"required"`
	Status           string          `json:"status" binding:"omitempty,oneof=active inactive pending"`
	Metadata         json.RawMessage `json:"metadata"`
}

type CreateRoleRequest struct {
	Code          string  `json:"code" binding:"required,min=2,max=64"`
	Name          string  `json:"name" binding:"required,min=1,max=255"`
	Description   string  `json:"description"`
	IsSystemRole  bool    `json:"is_system_role"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description   *string `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=128"`
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Module      string `json:"module" binding:"required,min=1,max=64"`
	Description string `json:"description"`
}

type AssignRolesRequest struct {
	RoleIDs []int64 `json:"role_ids" binding:"required"`
}

type VerifyIdentityRequest struct {
	Identity string `json:"identity" binding:"required"`
	Password string `json:"password" binding:"required"`
	Provider string `json:"provider" binding:"omitempty"`
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ZaloAuthRequest struct {
	AccessToken string `json:"access_token"`
	PhoneToken  string `json:"phone_token"`
	ZaloID      string `json:"zalo_id"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Phone       string `json:"phone"`
}

type ZaloPhoneResolveRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
	PhoneToken  string `json:"phone_token" binding:"required"`
}

type MemberRegisterRequest struct {
	FullName  string `json:"full_name" binding:"required,min=1,max=255"`
	Phone     string `json:"phone" binding:"required,min=8,max=32"`
	Email     string `json:"email" binding:"required,email,max=255"`
	AvatarURL string `json:"avatar_url" binding:"required,min=1,max=512"`
	City      string `json:"city" binding:"omitempty,max=128"`
	Ward      string `json:"ward" binding:"omitempty,max=128"`
}

type ListUsersQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive banned"`
	Search   string `form:"search"`
}

func (q *ListUsersQuery) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
}
