package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

const (
	HeaderAdminAPIKey    = "X-Admin-API-Key"
	HeaderInternalAPIKey = "X-Internal-API-Key"
	HeaderUserID         = "X-User-ID"
)

func AdminAuth(expectedKey string, adminAuthService service.AdminAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key := c.GetHeader(HeaderAdminAPIKey); key != "" {
			if expectedKey == "" || key != expectedKey {
				response.HandleError(c, apperror.ErrUnauthorized)
				c.Abort()
				return
			}
			c.Set(string(contextkeys.AdminAPIKeyAuth), true)
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			principal, err := adminAuthService.ValidateToken(token)
			if err == nil {
				c.Set(string(contextkeys.AdminUsername), principal.Username)
				c.Set(string(contextkeys.AdminTenantCode), principal.TenantCode)
				c.Next()
				return
			}
		}

		response.HandleError(c, apperror.ErrUnauthorized)
		c.Abort()
	}
}

func RequireAdminPermission(adminAuthService service.AdminAuthService, allowedPermissions ...string) gin.HandlerFunc {
	permissionSet := make(map[string]struct{}, len(allowedPermissions))
	for _, permission := range allowedPermissions {
		permissionSet[permission] = struct{}{}
	}

	return func(c *gin.Context) {
		if apiKeyAuth, _ := c.Get(string(contextkeys.AdminAPIKeyAuth)); apiKeyAuth == true {
			c.Next()
			return
		}

		usernameValue, ok := c.Get(string(contextkeys.AdminUsername))
		username, validUsername := usernameValue.(string)
		tenantValue, hasTenant := c.Get(string(contextkeys.AdminTenantCode))
		tenantCode, validTenant := tenantValue.(string)
		if !ok || !validUsername || username == "" || !hasTenant || !validTenant || tenantCode == "" {
			response.HandleError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		profile, err := adminAuthService.GetProfile(c.Request.Context(), service.AdminPrincipal{
			Username: username, TenantCode: tenantCode,
		})
		if err != nil {
			response.HandleError(c, err)
			c.Abort()
			return
		}

		if requestedTenant := strings.TrimSpace(c.GetHeader(HeaderTenantCode)); requestedTenant != "" && requestedTenant != tenantCode {
			response.HandleError(c, apperror.ErrForbidden)
			c.Abort()
			return
		}

		for _, role := range profile.User.Roles {
			for _, permission := range role.Permissions {
				if _, allowed := permissionSet[permission.Code]; allowed {
					c.Next()
					return
				}
			}
		}

		response.HandleError(c, apperror.ErrForbidden)
		c.Abort()
	}
}

func InternalAuth(expectedKey string) gin.HandlerFunc {
	return apiKeyAuth(HeaderInternalAPIKey, expectedKey)
}

func apiKeyAuth(header, expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			response.HandleError(c, apperror.ErrInternal)
			c.Abort()
			return
		}

		key := c.GetHeader(header)
		if key == "" || key != expectedKey {
			response.HandleError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(HeaderUserID)
		if userIDStr == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				userIDStr = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if userIDStr == "" {
			response.HandleError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil || userID <= 0 {
			response.HandleError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set(string(contextkeys.UserID), userID)
		c.Next()
	}
}
