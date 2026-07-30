package dto

import (
	"time"

	"github.com/updev/galaxy/identity_core/internal/domain"
)

type AdminLoginResponse struct {
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	User      domain.User    `json:"user"`
	Tenant    domain.Tenant  `json:"tenant"`
}

type AdminProfileResponse struct {
	User   domain.User  `json:"user"`
	Tenant domain.Tenant `json:"tenant"`
}

type ZaloAuthResponse struct {
	UserID   int64       `json:"user_id"`
	User     domain.User `json:"user"`
	IsMember bool        `json:"is_member"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Items []interface{}  `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}
