package service

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type RelationshipService interface {
	Create(ctx context.Context, tenantID int64, req dto.CreateRelationshipRequest) (*domain.UserRelationship, error)
	ListByUser(ctx context.Context, tenantID, userID int64) ([]domain.UserRelationship, error)
	Delete(ctx context.Context, tenantID, id int64) error
}

type relationshipService struct {
	repo     repository.RelationshipRepository
	userRepo repository.UserRepository
}

func NewRelationshipService(
	repo repository.RelationshipRepository,
	userRepo repository.UserRepository,
) RelationshipService {
	return &relationshipService{repo: repo, userRepo: userRepo}
}

func (s *relationshipService) Create(ctx context.Context, tenantID int64, req dto.CreateRelationshipRequest) (*domain.UserRelationship, error) {
	if req.FromUserID == req.ToUserID {
		return nil, apperror.New("INVALID_RELATIONSHIP", "cannot create relationship with self", apperror.ErrBadRequest.HTTPStatus)
	}

	fromExists, err := s.userRepo.ExistsInTenant(ctx, tenantID, req.FromUserID)
	if err != nil || !fromExists {
		return nil, apperror.New("FROM_USER_NOT_FOUND", "from_user not found in tenant", apperror.ErrBadRequest.HTTPStatus)
	}

	toExists, err := s.userRepo.ExistsInTenant(ctx, tenantID, req.ToUserID)
	if err != nil || !toExists {
		return nil, apperror.New("TO_USER_NOT_FOUND", "to_user not found in tenant", apperror.ErrBadRequest.HTTPStatus)
	}

	status := req.Status
	if status == "" {
		status = domain.RelationshipStatusActive
	}

	rel := &domain.UserRelationship{
		TenantID:         tenantID,
		FromUserID:       req.FromUserID,
		ToUserID:         req.ToUserID,
		RelationshipType: req.RelationshipType,
		Status:           status,
		Metadata:         toJSON(req.Metadata),
	}

	if err := s.repo.Create(ctx, rel); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return rel, nil
}

func (s *relationshipService) ListByUser(ctx context.Context, tenantID, userID int64) ([]domain.UserRelationship, error) {
	if _, err := s.userRepo.FindByID(ctx, tenantID, userID); err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}
	rels, err := s.repo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return rels, nil
}

func (s *relationshipService) Delete(ctx context.Context, tenantID, id int64) error {
	if err := s.repo.Delete(ctx, tenantID, id); err != nil {
		return mapDBError(err, apperror.ErrNotFound)
	}
	return nil
}
