package service

import (
	"context"
	"errors"
	"strings"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
	"gorm.io/gorm"
)

const MemberRoleCode = "member"

type ZaloAuthService interface {
	Authenticate(ctx context.Context, tenantID int64, tenantCode string, req dto.ZaloAuthRequest) (*dto.ZaloAuthResponse, error)
	RegisterMember(ctx context.Context, tenantID int64, tenantCode string, userID int64, req dto.MemberRegisterRequest) (*domain.User, error)
	ResolvePhone(ctx context.Context, req dto.ZaloPhoneResolveRequest) (string, error)
}

type zaloAuthService struct {
	userRepo     repository.UserRepository
	identityRepo repository.IdentityRepository
	roleRepo     repository.RoleRepository
	zaloClient   ZaloClient
	zns          ZNSNotifier
	devMode      bool
}

func NewZaloAuthService(
	userRepo repository.UserRepository,
	identityRepo repository.IdentityRepository,
	roleRepo repository.RoleRepository,
	zaloClient ZaloClient,
	zns ZNSNotifier,
	devMode bool,
) ZaloAuthService {
	return &zaloAuthService{
		userRepo:     userRepo,
		identityRepo: identityRepo,
		roleRepo:     roleRepo,
		zaloClient:   zaloClient,
		zns:          zns,
		devMode:      devMode,
	}
}

func (s *zaloAuthService) Authenticate(ctx context.Context, tenantID int64, tenantCode string, req dto.ZaloAuthRequest) (*dto.ZaloAuthResponse, error) {
	profile, err := s.resolveProfile(ctx, req)
	if err != nil {
		return nil, err
	}

	identity, err := s.identityRepo.FindByProviderUserID(ctx, tenantID, domain.IdentityProviderZalo, profile.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		identity, err = s.identityRepo.FindByIdentity(ctx, tenantID, domain.IdentityProviderZalo, profile.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mapDBError(err, apperror.ErrInternal)
		}
	}

	var user *domain.User
	if identity != nil {
		user, err = s.userRepo.FindByIDWithRelations(ctx, tenantID, identity.UserID)
		if err != nil {
			return nil, mapDBError(err, apperror.ErrNotFound)
		}
		updated := false
		if profile.Name != "" && user.FullName == "" {
			user.FullName = profile.Name
			user.DisplayName = profile.Name
			updated = true
		}
		if profile.Avatar != "" && user.AvatarURL == "" {
			user.AvatarURL = profile.Avatar
			updated = true
		}
		if profile.Phone != "" && user.Phone == "" {
			user.Phone = profile.Phone
			updated = true
		}
		if updated {
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, mapDBError(err, apperror.ErrInternal)
			}
			user, err = s.userRepo.FindByIDWithRelations(ctx, tenantID, user.ID)
			if err != nil {
				return nil, mapDBError(err, apperror.ErrNotFound)
			}
		}
	} else {
		user = &domain.User{
			TenantID:    tenantID,
			FullName:    profile.Name,
			DisplayName: profile.Name,
			AvatarURL:   profile.Avatar,
			Phone:       profile.Phone,
			Status:      domain.UserStatusActive,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, mapDBError(err, apperror.ErrInternal)
		}
		zaloIdentity := &domain.UserIdentity{
			UserID:         user.ID,
			Provider:       domain.IdentityProviderZalo,
			ProviderUserID: profile.ID,
			Identity:       profile.ID,
		}
		if err := s.identityRepo.Create(ctx, zaloIdentity); err != nil {
			return nil, mapDBError(err, apperror.ErrInternal)
		}
		user, err = s.userRepo.FindByIDWithRelations(ctx, tenantID, user.ID)
		if err != nil {
			return nil, mapDBError(err, apperror.ErrNotFound)
		}
	}

	_ = tenantCode
	return &dto.ZaloAuthResponse{
		UserID:   user.ID,
		User:     *user,
		IsMember: hasMemberRole(user.Roles),
	}, nil
}

func (s *zaloAuthService) RegisterMember(ctx context.Context, tenantID int64, tenantCode string, userID int64, req dto.MemberRegisterRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	user.FullName = strings.TrimSpace(req.FullName)
	user.DisplayName = strings.TrimSpace(req.FullName)
	user.Phone = strings.TrimSpace(req.Phone)
	user.Email = strings.TrimSpace(req.Email)
	user.AvatarURL = strings.TrimSpace(req.AvatarURL)
	user.City = strings.TrimSpace(req.City)
	user.Ward = strings.TrimSpace(req.Ward)
	user.Status = domain.UserStatusActive

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}

	role, err := s.ensureMemberRole(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if err := s.roleRepo.AddRoleToUser(ctx, userID, role.ID); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}

	user, err = s.userRepo.FindByIDWithRelations(ctx, tenantID, userID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	s.zns.NotifyMemberRegistered(ctx, tenantCode, user.ID, user.Phone, user.FullName)
	return user, nil
}

func (s *zaloAuthService) resolveProfile(ctx context.Context, req dto.ZaloAuthRequest) (*ZaloProfile, error) {
	var profile *ZaloProfile
	var err error

	if strings.TrimSpace(req.AccessToken) != "" {
		profile, err = s.zaloClient.GetProfile(ctx, strings.TrimSpace(req.AccessToken))
		if err != nil {
			return nil, err
		}
	} else if s.devMode && strings.TrimSpace(req.ZaloID) != "" {
		profile = &ZaloProfile{
			ID:     strings.TrimSpace(req.ZaloID),
			Name:   strings.TrimSpace(req.Name),
			Avatar: strings.TrimSpace(req.AvatarURL),
			Phone:  strings.TrimSpace(req.Phone),
		}
	} else {
		return nil, apperror.New("ZALO_AUTH_REQUIRED", "access_token is required (or zalo_id in ZALO_AUTH_DEV_MODE)", apperror.ErrBadRequest.HTTPStatus)
	}

	// Prefer client-supplied name/avatar from getUserInfo when Graph picture is empty.
	if profile.Name == "" && strings.TrimSpace(req.Name) != "" {
		profile.Name = strings.TrimSpace(req.Name)
	}
	if profile.Avatar == "" && strings.TrimSpace(req.AvatarURL) != "" {
		profile.Avatar = strings.TrimSpace(req.AvatarURL)
	}
	if profile.Phone == "" && strings.TrimSpace(req.Phone) != "" {
		profile.Phone = strings.TrimSpace(req.Phone)
	}

	// Exchange getPhoneNumber token → phone (server-side only). Soft-fail so
	// missing secret / denied phone still allows login; user can type phone.
	if strings.TrimSpace(req.PhoneToken) != "" && strings.TrimSpace(req.AccessToken) != "" {
		phone, phoneErr := s.zaloClient.ResolvePhoneNumber(ctx, strings.TrimSpace(req.AccessToken), strings.TrimSpace(req.PhoneToken))
		if phoneErr == nil && phone != "" {
			profile.Phone = phone
		}
	}

	return profile, nil
}

func (s *zaloAuthService) ResolvePhone(ctx context.Context, req dto.ZaloPhoneResolveRequest) (string, error) {
	return s.zaloClient.ResolvePhoneNumber(ctx, req.AccessToken, req.PhoneToken)
}

func (s *zaloAuthService) ensureMemberRole(ctx context.Context, tenantID int64) (*domain.Role, error) {
	role, err := s.roleRepo.FindByCode(ctx, tenantID, MemberRoleCode)
	if err == nil {
		return role, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	role = &domain.Role{
		TenantID:     tenantID,
		Code:         MemberRoleCode,
		Name:         "Thành viên",
		Description:  "SGS Academy member",
		IsSystemRole: true,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return role, nil
}

func hasMemberRole(roles []domain.Role) bool {
	for _, role := range roles {
		if role.Code == MemberRoleCode {
			return true
		}
	}
	return false
}
