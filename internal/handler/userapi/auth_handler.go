package userapi

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/handler"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type AuthHandler struct {
	zaloAuthService service.ZaloAuthService
}

func NewAuthHandler(zaloAuthService service.ZaloAuthService) *AuthHandler {
	return &AuthHandler{zaloAuthService: zaloAuthService}
}

func (h *AuthHandler) ZaloAuth(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	tenantCode, _ := c.Get(string(contextkeys.TenantCode))
	code, _ := tenantCode.(string)

	var req dto.ZaloAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	result, err := h.zaloAuthService.Authenticate(c.Request.Context(), tenantID, code, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AuthHandler) ResolveZaloPhone(c *gin.Context) {
	var req dto.ZaloPhoneResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}
	phone, err := h.zaloAuthService.ResolvePhone(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, gin.H{"phone": phone})
}

func (h *AuthHandler) RegisterMember(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	userID, err := handler.UserIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	tenantCode, _ := c.Get(string(contextkeys.TenantCode))
	code, _ := tenantCode.(string)

	var req dto.MemberRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	user, err := h.zaloAuthService.RegisterMember(c.Request.Context(), tenantID, code, userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, user)
}
