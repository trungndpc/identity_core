package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

const HeaderTenantCode = "X-Tenant-Code"

func TenantRequired(tenantService service.TenantService) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.GetHeader(HeaderTenantCode)
		if code == "" {
			response.HandleError(c, apperror.ErrTenantRequired)
			c.Abort()
			return
		}

		tenant, err := tenantService.GetByCode(c.Request.Context(), code)
		if err != nil {
			response.HandleError(c, apperror.ErrTenantInvalid)
			c.Abort()
			return
		}

		if tenant.Status != domain.TenantStatusActive {
			response.HandleError(c, apperror.ErrTenantInvalid)
			c.Abort()
			return
		}

		c.Set(string(contextkeys.TenantID), tenant.ID)
		c.Set(string(contextkeys.TenantCode), tenant.Code)
		c.Next()
	}
}
