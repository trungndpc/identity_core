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
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			username, err := adminAuthService.ValidateToken(token)
			if err == nil {
				c.Set(string(contextkeys.AdminUsername), username)
				c.Next()
				return
			}
		}

		response.HandleError(c, apperror.ErrUnauthorized)
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
