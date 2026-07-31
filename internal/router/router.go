package router

import (
	"github.com/gin-gonic/gin"
	"github.com/updev/galaxy/identity_core/internal/config"
	"github.com/updev/galaxy/identity_core/internal/contextkeys"
	"github.com/updev/galaxy/identity_core/internal/handler/admin"
	"github.com/updev/galaxy/identity_core/internal/handler/internalapi"
	"github.com/updev/galaxy/identity_core/internal/handler/userapi"
	"github.com/updev/galaxy/identity_core/internal/middleware"
	"github.com/updev/galaxy/identity_core/internal/service"
)

type Dependencies struct {
	Config              *config.Config
	AdminAuthService    service.AdminAuthService
	TenantService       service.TenantService
	UserService         service.UserService
	IdentityService     service.IdentityService
	RelationshipService service.RelationshipService
	RoleService         service.RoleService
	PermissionService   service.PermissionService
	ZaloAuthService     service.ZaloAuthService
	UserTokenService    service.UserTokenService
}

func Setup(deps Dependencies) *gin.Engine {
	if deps.Config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.CORS(deps.Config.CORSAllowedOrigins), gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// ── Admin API: quản trị hệ thống, không yêu cầu tenant cho tenant management ──
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.RequestSource(contextkeys.SourceAdmin))

	adminAuthHandler := admin.NewAuthHandler(deps.AdminAuthService)
	adminGroup.POST("/auth/login", adminAuthHandler.Login)

	adminProtected := adminGroup.Group("")
	adminProtected.Use(middleware.AdminAuth(deps.Config.AdminAPIKey, deps.AdminAuthService))
	adminProtected.GET("/auth/me", adminAuthHandler.Me)

	adminTenantHandler := admin.NewTenantHandler(deps.TenantService)
	tenantAdmin := adminProtected.Group("")
	tenantAdmin.Use(middleware.RequireAdminPermission(deps.AdminAuthService, "tenants.manage"))
	tenantAdmin.POST("/tenants", adminTenantHandler.Create)
	tenantAdmin.GET("/tenants", adminTenantHandler.List)
	tenantAdmin.GET("/tenants/:id", adminTenantHandler.Get)
	tenantAdmin.PUT("/tenants/:id", adminTenantHandler.Update)

	adminTenantScoped := adminProtected.Group("")
	adminTenantScoped.Use(middleware.TenantRequired(deps.TenantService))

	adminUserHandler := admin.NewUserHandler(deps.UserService)
	userAdmin := adminTenantScoped.Group("")
	userAdmin.Use(middleware.RequireAdminPermission(deps.AdminAuthService, "users.manage"))
	userAdmin.POST("/users", adminUserHandler.Create)
	userAdmin.GET("/users", adminUserHandler.List)
	userAdmin.GET("/users/:id", adminUserHandler.Get)
	userAdmin.PUT("/users/:id", adminUserHandler.Update)
	userAdmin.DELETE("/users/:id", adminUserHandler.Delete)
	userAdmin.PUT("/users/:id/roles", adminUserHandler.AssignRoles)

	adminRoleHandler := admin.NewRoleHandler(deps.RoleService)
	roleAdmin := adminTenantScoped.Group("")
	roleAdmin.Use(middleware.RequireAdminPermission(deps.AdminAuthService, "roles.manage"))
	roleAdmin.POST("/roles", adminRoleHandler.Create)
	roleAdmin.GET("/roles", adminRoleHandler.List)
	roleAdmin.GET("/roles/:id", adminRoleHandler.Get)
	roleAdmin.PUT("/roles/:id", adminRoleHandler.Update)
	roleAdmin.DELETE("/roles/:id", adminRoleHandler.Delete)

	adminPermissionHandler := admin.NewPermissionHandler(deps.PermissionService)
	permissionAdmin := adminProtected.Group("")
	permissionAdmin.Use(middleware.RequireAdminPermission(deps.AdminAuthService, "permissions.manage"))
	permissionAdmin.POST("/permissions", adminPermissionHandler.Create)
	permissionAdmin.GET("/permissions", adminPermissionHandler.List)

	adminRelHandler := admin.NewRelationshipHandler(deps.RelationshipService)
	relationshipAdmin := adminTenantScoped.Group("")
	relationshipAdmin.Use(middleware.RequireAdminPermission(deps.AdminAuthService, "relationships.manage"))
	relationshipAdmin.POST("/relationships", adminRelHandler.Create)
	relationshipAdmin.GET("/users/:id/relationships", adminRelHandler.ListByUser)
	relationshipAdmin.DELETE("/relationships/:id", adminRelHandler.Delete)

	// ── Internal API: service-to-service integration ──
	internalGroup := v1.Group("/internal")
	internalGroup.Use(
		middleware.RequestSource(contextkeys.SourceInternal),
		middleware.InternalAuth(deps.Config.InternalAPIKey),
		middleware.TenantRequired(deps.TenantService),
	)

	internalHandler := internalapi.NewHandler(
		deps.UserService,
		deps.IdentityService,
		deps.RelationshipService,
	)
	internalGroup.GET("/users/:id", internalHandler.GetUser)
	internalGroup.GET("/users", internalHandler.ListUsers)
	internalGroup.POST("/auth/verify", internalHandler.VerifyIdentity)
	internalGroup.GET("/users/:id/relationships", internalHandler.GetUserRelationships)

	// ── User API: end-user self-service ──
	userGroup := v1.Group("/user")
	userGroup.Use(
		middleware.RequestSource(contextkeys.SourceUser),
		middleware.TenantRequired(deps.TenantService),
	)

	authHandler := userapi.NewAuthHandler(deps.ZaloAuthService)
	userGroup.POST("/auth/zalo", authHandler.ZaloAuth)
	userGroup.POST("/auth/zalo/phone", authHandler.ResolveZaloPhone)

	userAuthed := userGroup.Group("")
	userAuthed.Use(middleware.UserAuth(deps.UserTokenService, deps.Config.AllowLegacyUserID))

	userHandler := userapi.NewHandler(
		deps.UserService,
		deps.IdentityService,
		deps.RelationshipService,
	)
	userAuthed.GET("/me", userHandler.GetProfile)
	userAuthed.PUT("/me", userHandler.UpdateProfile)
	userAuthed.GET("/me/identities", userHandler.ListIdentities)
	userAuthed.GET("/me/relationships", userHandler.ListRelationships)
	userAuthed.POST("/members/register", authHandler.RegisterMember)

	return r
}
