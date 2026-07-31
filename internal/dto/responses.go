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
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	UserID      int64       `json:"user_id"`
	User        domain.User `json:"user"`
	IsMember    bool        `json:"is_member"`
}

type UserTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
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
