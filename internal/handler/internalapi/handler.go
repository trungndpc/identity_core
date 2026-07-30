package internalapi

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/handler"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type Handler struct {
	userService       service.UserService
	identityService   service.IdentityService
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

func (h *Handler) GetUser(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	userID, err := handler.ParseIDParam(c, "id")
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

func (h *Handler) ListUsers(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var query dto.ListUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.HandleError(c, err)
		return
	}
	query.Normalize()

	users, total, err := h.userService.List(c.Request.Context(), tenantID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	response.OK(c, gin.H{
		"items": users,
		"meta": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *Handler) VerifyIdentity(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.VerifyIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	user, err := h.identityService.Verify(c.Request.Context(), tenantID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, user)
}

func (h *Handler) GetUserRelationships(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	userID, err := handler.ParseIDParam(c, "id")
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
