package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

func TenantIDFromContext(c *gin.Context) (int64, error) {
	value, exists := c.Get(string(contextkeys.TenantID))
	if !exists {
		return 0, apperror.ErrTenantRequired
	}
	id, ok := value.(int64)
	if !ok {
		return 0, apperror.ErrInternal
	}
	return id, nil
}

func UserIDFromContext(c *gin.Context) (int64, error) {
	value, exists := c.Get(string(contextkeys.UserID))
	if !exists {
		return 0, apperror.ErrUnauthorized
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return 0, apperror.ErrUnauthorized
		}
		return id, nil
	default:
		return 0, apperror.ErrUnauthorized
	}
}

func ParseIDParam(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperror.ErrBadRequest
	}
	return id, nil
}
