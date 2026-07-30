package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/handler"
	"github.com/updev/galaxy/identity_core/internal/service"
	"github.com/updev/galaxy/identity_core/pkg/response"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Create(c *gin.Context) {
	tenantID, err := handler.TenantIDFromContext(c)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	user, err := h.userService.Create(c.Request.Context(), tenantID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Created(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
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

func (h *UserHandler) Get(c *gin.Context) {
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

func (h *UserHandler) List(c *gin.Context) {
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

func (h *UserHandler) Delete(c *gin.Context) {
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

	if err := h.userService.Delete(c.Request.Context(), tenantID, userID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}

func (h *UserHandler) AssignRoles(c *gin.Context) {
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

	var req dto.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := h.userService.AssignRoles(c.Request.Context(), tenantID, userID, req.RoleIDs); err != nil {
		response.HandleError(c, err)
		return
	}
	response.NoContent(c)
}
