package main

import (
	"fmt"
	"log"

	"github.com/updev/galaxy/identity_core/internal/config"
	"github.com/updev/galaxy/identity_core/internal/database"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/internal/router"
	"github.com/updev/galaxy/identity_core/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate db: %v", err)
	}

	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	identityRepo := repository.NewIdentityRepository(db)
	relationshipRepo := repository.NewRelationshipRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	znsNotifier := service.NewStubZNSNotifier()
	userTokenService := service.NewUserTokenService(cfg)
	zaloAuthService := service.NewZaloAuthService(
		userRepo,
		identityRepo,
		roleRepo,
		service.NewHTTPZaloClient(cfg.ZaloAppSecretKey),
		znsNotifier,
		userTokenService,
		cfg.ZaloAuthDevMode,
	)

	deps := router.Dependencies{
		Config:              cfg,
		AdminAuthService:    service.NewAdminAuthService(cfg, service.NewIdentityService(userRepo, identityRepo), userRepo, tenantRepo),
		TenantService:       service.NewTenantService(tenantRepo),
		UserService:         service.NewUserService(userRepo, identityRepo, roleRepo),
		IdentityService:     service.NewIdentityService(userRepo, identityRepo),
		RelationshipService: service.NewRelationshipService(relationshipRepo, userRepo),
		RoleService:         service.NewRoleService(roleRepo, permissionRepo),
		PermissionService:   service.NewPermissionService(permissionRepo),
		ZaloAuthService:     zaloAuthService,
		UserTokenService:    userTokenService,
	}

	r := router.Setup(deps)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("identity_core server starting on %s (env=%s)", addr, cfg.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
