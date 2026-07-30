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
)

type adminAuthServiceStub struct {
	profile *dto.AdminProfileResponse
	err     error
}

func (s adminAuthServiceStub) Login(context.Context, string, string) (*dto.AdminLoginResponse, error) {
	return nil, nil
}

func (s adminAuthServiceStub) GetProfile(context.Context, string) (*dto.AdminProfileResponse, error) {
	return s.profile, s.err
}

func (s adminAuthServiceStub) ValidateToken(string) (string, error) {
	return "", nil
}

func TestRequireAdminRoleAllowsMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminUsername), "admin")
	})
	engine.Use(RequireAdminRole(adminAuthServiceStub{
		profile: &dto.AdminProfileResponse{
			User: domain.User{Roles: []domain.Role{{Code: "super_admin"}}},
		},
	}, "super_admin"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestRequireAdminRoleRejectsMissingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminUsername), "editor")
	})
	engine.Use(RequireAdminRole(adminAuthServiceStub{
		profile: &dto.AdminProfileResponse{
			User: domain.User{Roles: []domain.Role{{Code: "editor"}}},
		},
	}, "super_admin"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, responseRecorder.Code)
	}
}

func TestRequireAdminRoleAllowsSystemAPIKeyAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(contextkeys.AdminAPIKeyAuth), true)
	})
	engine.Use(RequireAdminRole(adminAuthServiceStub{}, "super_admin"))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	responseRecorder := httptest.NewRecorder()
	engine.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}
