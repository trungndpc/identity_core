package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/handler"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type RelationshipHandler struct {
	relationshipService service.RelationshipService
}

func NewRelationshipHandler(relationshipService service.RelationshipService) *RelationshipHandler {
	return &RelationshipHandler{relationshipService: relationshipService}
}

func (h *RelationshipHandler) Create(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.CreateRelationshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	rel, err := h.relationshipService.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, rel)
}

func (h *RelationshipHandler) ListByUser(c *gin.Context) {
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

func (h *RelationshipHandler) Delete(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	id, err := handler.ParseIDParam(c, "id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := h.relationshipService.Delete(c.Request.Context(), tenantID, id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
