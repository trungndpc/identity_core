package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/updev/galaxy/identity_core/internal/config"
	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type AdminAuthService interface {
	Login(ctx context.Context, username, password string) (*dto.AdminLoginResponse, error)
	GetProfile(ctx context.Context, principal AdminPrincipal) (*dto.AdminProfileResponse, error)
	ValidateToken(tokenString string) (AdminPrincipal, error)
}

type AdminPrincipal struct {
	Username   string
	TenantCode string
}

type adminAuthService struct {
	cfg             *config.Config
	identityService IdentityService
	userRepo        repository.UserRepository
	tenantRepo      repository.TenantRepository
}

type adminJWTClaims struct {
	Username   string `json:"username"`
	TenantCode string `json:"tenant_code"`
	jwt.RegisteredClaims
}

func NewAdminAuthService(
	cfg *config.Config,
	identityService IdentityService,
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
) AdminAuthService {
	return &adminAuthService{
		cfg:             cfg,
		identityService: identityService,
		userRepo:        userRepo,
		tenantRepo:      tenantRepo,
	}
}

func (s *adminAuthService) Login(ctx context.Context, username, password string) (*dto.AdminLoginResponse, error) {
	if s.cfg.JWTSecret == "" {
		return nil, apperror.ErrInternal
	}

	adminUser, err := s.userRepo.FindByUsernameGlobally(ctx, username)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	tenant, err := s.tenantRepo.FindByID(ctx, adminUser.TenantID)
	if err != nil || tenant.Status != domain.TenantStatusActive {
		return nil, apperror.ErrUnauthorized
	}

	user, err := s.identityService.Verify(ctx, adminUser.TenantID, dto.VerifyIdentityRequest{
		Identity: username,
		Password: password,
	})
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.JWTExpiryHours) * time.Hour)
	claims := adminJWTClaims{
		Username:   user.Username,
		TenantCode: tenant.Code,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "admin",
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to sign token", apperror.ErrInternal.HTTPStatus)
	}

	return &dto.AdminLoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *sanitizeUserForResponse(user),
		Tenant:    *tenant,
	}, nil
}

func (s *adminAuthService) GetProfile(ctx context.Context, principal AdminPrincipal) (*dto.AdminProfileResponse, error) {
	tenant, err := s.tenantRepo.FindByCode(ctx, principal.TenantCode)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrUnauthorized)
	}

	user, err := s.userRepo.FindByUsername(ctx, tenant.ID, principal.Username)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrUnauthorized)
	}

	if user.Status != domain.UserStatusActive {
		return nil, apperror.ErrUnauthorized
	}

	if tenant.Status != domain.TenantStatusActive {
		return nil, apperror.ErrUnauthorized
	}

	return &dto.AdminProfileResponse{
		User:   *sanitizeUserForResponse(user),
		Tenant: *tenant,
	}, nil
}

func sanitizeUserForResponse(user *domain.User) *domain.User {
	if user == nil {
		return &domain.User{}
	}

	responseUser := *user
	responseUser.Identities = nil
	return &responseUser
}

func (s *adminAuthService) ValidateToken(tokenString string) (AdminPrincipal, error) {
	if s.cfg.JWTSecret == "" {
		return AdminPrincipal{}, apperror.ErrUnauthorized
	}

	claims := &adminJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid || claims.Username == "" || claims.TenantCode == "" {
		return AdminPrincipal{}, apperror.ErrUnauthorized
	}

	return AdminPrincipal{Username: claims.Username, TenantCode: claims.TenantCode}, nil
}
