package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/handler"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type RoleHandler struct {
	roleService service.RoleService
}

func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) Create(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	role, err := h.roleService.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	roleID, err := handler.ParseIDParam(c, "id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	role, err := h.roleService.Update(c.Request.Context(), tenantID, roleID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, role)
}

func (h *RoleHandler) Get(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	roleID, err := handler.ParseIDParam(c, "id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	role, err := h.roleService.GetByID(c.Request.Context(), tenantID, roleID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, role)
}

func (h *RoleHandler) List(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	roles, err := h.roleService.List(c.Request.Context(), tenantID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, roles)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	roleID, err := handler.ParseIDParam(c, "id")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if err := h.roleService.Delete(c.Request.Context(), tenantID, roleID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
