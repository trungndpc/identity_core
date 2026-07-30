package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
)

func RequestSource(source string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(contextkeys.RequestSource), source)
		c.Next()
	}
}
