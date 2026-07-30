package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/service"
)

type adminAuthServiceStub struct {
	profile *dto.AdminProfileResponse
	err     error
}

func (s adminAuthServiceStub) Login(context.Context, string, string) (*dto.AdminLoginResponse, error) {
	return nil, nil
}

func (s adminAuthServiceStub) GetProfile(context.Context, service.AdminPrincipal) (*dto.AdminProfileResponse, error) {
	return s.profile, s.err
}

func (s adminAuthServiceStub) ValidateToken(string) (service.AdminPrincipal, error) {
	return service.AdminPrincipal{}, nil
}

func TestRequireAdminPermissionAllowsMatchingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminUsername), "admin")
		c.Set(string(contextkeys.AdminTenantCode), "root")
	})
	engine.Use(RequireAdminPermission(adminAuthServiceStub{
		profile: &dto.AdminProfileResponse{
			User: domain.User{Roles: []domain.Role{{
				Code:        "super_admin",
				Permissions: []domain.Permission{{Code: "tenants.manage"}},
			}}},
		},
	}, "tenants.manage"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestRequireAdminPermissionRejectsMissingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminUsername), "editor")
		c.Set(string(contextkeys.AdminTenantCode), "sgs")
	})
	engine.Use(RequireAdminPermission(adminAuthServiceStub{
		profile: &dto.AdminProfileResponse{
			User: domain.User{Roles: []domain.Role{{Code: "editor"}}},
		},
	}, "tenants.manage"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, responseRecorder.Code)
	}
}

func TestRequireAdminPermissionAllowsSystemAPIKeyAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminAPIKeyAuth), true)
	})
	engine.Use(RequireAdminPermission(adminAuthServiceStub{}, "tenants.manage"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestRequireAdminPermissionRejectsDifferentTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminUsername), "sgs")
		c.Set(string(contextkeys.AdminTenantCode), "sgs")
	})
	engine.Use(RequireAdminPermission(adminAuthServiceStub{
		profile: &dto.AdminProfileResponse{
			User: domain.User{Roles: []domain.Role{{
				Permissions: []domain.Permission{{Code: "users.manage"}},
			}}},
		},
	}, "users.manage"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(HeaderTenantCode, "knauf")
	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, responseRecorder.Code)
	}
}
