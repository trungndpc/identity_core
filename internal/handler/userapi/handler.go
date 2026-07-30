package userapi

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/handler"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type Handler struct {
	userService         service.UserService
	identityService     service.IdentityService
	relationshipService service.RelationshipService
}

func NewHandler(
	userService service.UserService,
	identityService service.IdentityService,
	relationshipService service.RelationshipService,
) *Handler {
	return &Handler{
		userService:         userService,
		identityService:     identityService,
		relationshipService: relationshipService,
	}
}

func (h *Handler) GetProfile(c *gin.Context) {
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

	user, err := h.userService.GetByID(c.Request.Context(), tenantID, userID, true)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, user)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
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

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	user, err := h.userService.Update(c.Request.Context(), tenantID, userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, user)
}

func (h *Handler) ListIdentities(c *gin.Context) {
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

	identities, err := h.identityService.ListByUser(c.Request.Context(), tenantID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, identities)
}

func (h *Handler) ListRelationships(c *gin.Context) {
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

	rels, err := h.relationshipService.ListByUser(c.Request.Context(), tenantID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, rels)
}
