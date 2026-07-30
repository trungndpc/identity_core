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
	adminProtected.POST("/tenants", adminTenantHandler.Create)
	adminProtected.GET("/tenants", adminTenantHandler.List)
	adminProtected.GET("/tenants/:id", adminTenantHandler.Get)
	adminProtected.PUT("/tenants/:id", adminTenantHandler.Update)

	adminTenantScoped := adminProtected.Group("")
	adminTenantScoped.Use(middleware.TenantRequired(deps.TenantService))

	adminUserHandler := admin.NewUserHandler(deps.UserService)
	adminTenantScoped.POST("/users", adminUserHandler.Create)
	adminTenantScoped.GET("/users", adminUserHandler.List)
	adminTenantScoped.GET("/users/:id", adminUserHandler.Get)
	adminTenantScoped.PUT("/users/:id", adminUserHandler.Update)
	adminTenantScoped.DELETE("/users/:id", adminUserHandler.Delete)
	adminTenantScoped.PUT("/users/:id/roles", adminUserHandler.AssignRoles)

	adminRoleHandler := admin.NewRoleHandler(deps.RoleService)
	adminTenantScoped.POST("/roles", adminRoleHandler.Create)
	adminTenantScoped.GET("/roles", adminRoleHandler.List)
	adminTenantScoped.GET("/roles/:id", adminRoleHandler.Get)
	adminTenantScoped.PUT("/roles/:id", adminRoleHandler.Update)
	adminTenantScoped.DELETE("/roles/:id", adminRoleHandler.Delete)

	adminPermissionHandler := admin.NewPermissionHandler(deps.PermissionService)
	adminProtected.POST("/permissions", adminPermissionHandler.Create)
	adminProtected.GET("/permissions", adminPermissionHandler.List)

	adminRelHandler := admin.NewRelationshipHandler(deps.RelationshipService)
	adminTenantScoped.POST("/relationships", adminRelHandler.Create)
	adminTenantScoped.GET("/users/:id/relationships", adminRelHandler.ListByUser)
	adminTenantScoped.DELETE("/relationships/:id", adminRelHandler.Delete)

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
	userAuthed.Use(middleware.UserAuth())

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
