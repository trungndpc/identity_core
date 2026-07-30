package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type AuthHandler struct {
	adminAuthService service.AdminAuthService
}

func NewAuthHandler(adminAuthService service.AdminAuthService) *AuthHandler {
	return &AuthHandler{adminAuthService: adminAuthService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := h.adminAuthService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	usernameValue, ok := c.Get(string(contextkeys.AdminUsername))
	if !ok {
		response.HandleError(c, apperror.ErrUnauthorized)
		return
	}

	username, ok := usernameValue.(string)
	if !ok || username == "" {
		response.HandleError(c, apperror.ErrUnauthorized)
		return
	}

	result, err := h.adminAuthService.GetProfile(c.Request.Context(), username)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.OK(c, result)
}
