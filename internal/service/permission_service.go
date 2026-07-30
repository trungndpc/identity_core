package service

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type PermissionService interface {
	Create(ctx context.Context, req dto.CreatePermissionRequest) (*domain.Permission, error)
	List(ctx context.Context) ([]domain.Permission, error)
}

type permissionService struct {
	repo repository.PermissionRepository
}

func NewPermissionService(repo repository.PermissionRepository) PermissionService {
	return &permissionService{repo: repo}
}

func (s *permissionService) Create(ctx context.Context, req dto.CreatePermissionRequest) (*domain.Permission, error) {
	if _, err := s.repo.FindByCode(ctx, req.Code); err == nil {
		return nil, apperror.New("PERMISSION_EXISTS", "permission code already exists", apperror.ErrConflict.HTTPStatus)
	}

	permission := &domain.Permission{
		Code:        req.Code,
		Name:        req.Name,
		Module:      req.Module,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, permission); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return permission, nil
}

func (s *permissionService) List(ctx context.Context) ([]domain.Permission, error) {
	permissions, err := s.repo.List(ctx)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return permissions, nil
}
