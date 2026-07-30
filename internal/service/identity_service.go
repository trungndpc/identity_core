package service

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type IdentityService interface {
	Create(ctx context.Context, tenantID, userID int64, req dto.CreateIdentityRequest) (*domain.UserIdentity, error)
	ListByUser(ctx context.Context, tenantID, userID int64) ([]domain.UserIdentity, error)
	Verify(ctx context.Context, tenantID int64, req dto.VerifyIdentityRequest) (*domain.User, error)
}

type identityService struct {
	userRepo     repository.UserRepository
	identityRepo repository.IdentityRepository
}

func NewIdentityService(
	userRepo repository.UserRepository,
	identityRepo repository.IdentityRepository,
) IdentityService {
	return &identityService{
		userRepo:     userRepo,
		identityRepo: identityRepo,
	}
}

func (s *identityService) Create(ctx context.Context, tenantID, userID int64, req dto.CreateIdentityRequest) (*domain.UserIdentity, error) {
	if _, err := s.userRepo.FindByID(ctx, tenantID, userID); err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	provider := req.Provider
	if provider == "" {
		provider = domain.IdentityProviderLocal
	}

	passwordHash := ""
	if req.Password != "" {
		hash, err := hashPassword(req.Password)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to hash password", apperror.ErrInternal.HTTPStatus)
		}
		passwordHash = hash
	}

	identity := &domain.UserIdentity{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: req.ProviderUserID,
		Identity:       req.Identity,
		PasswordHash:   passwordHash,
		Metadata:       toJSON(req.Metadata),
	}

	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return identity, nil
}

func (s *identityService) ListByUser(ctx context.Context, tenantID, userID int64) ([]domain.UserIdentity, error) {
	if _, err := s.userRepo.FindByID(ctx, tenantID, userID); err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}
	identities, err := s.identityRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return identities, nil
}

func (s *identityService) Verify(ctx context.Context, tenantID int64, req dto.VerifyIdentityRequest) (*domain.User, error) {
	provider := req.Provider
	if provider == "" {
		provider = domain.IdentityProviderLocal
	}

	identity, err := s.identityRepo.FindByIdentity(ctx, tenantID, provider, req.Identity)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrUnauthorized)
	}

	if identity.PasswordHash == "" || !checkPassword(identity.PasswordHash, req.Password) {
		return nil, apperror.ErrUnauthorized
	}

	user, err := s.userRepo.FindByIDWithRelations(ctx, tenantID, identity.UserID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	if user.Status != domain.UserStatusActive {
		return nil, apperror.ErrForbidden
	}

	return user, nil
}
