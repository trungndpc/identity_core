package service

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type RoleService interface {
	Create(ctx context.Context, tenantID int64, req dto.CreateRoleRequest) (*domain.Role, error)
	Update(ctx context.Context, tenantID, roleID int64, req dto.UpdateRoleRequest) (*domain.Role, error)
	GetByID(ctx context.Context, tenantID, roleID int64) (*domain.Role, error)
	List(ctx context.Context, tenantID int64) ([]domain.Role, error)
	Delete(ctx context.Context, tenantID, roleID int64) error
}

type roleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

func NewRoleService(roleRepo repository.RoleRepository, permissionRepo repository.PermissionRepository) RoleService {
	return &roleService{roleRepo: roleRepo, permissionRepo: permissionRepo}
}

func (s *roleService) Create(ctx context.Context, tenantID int64, req dto.CreateRoleRequest) (*domain.Role, error) {
	if _, err := s.roleRepo.FindByCode(ctx, tenantID, req.Code); err == nil {
		return nil, apperror.New("ROLE_EXISTS", "role code already exists in tenant", apperror.ErrConflict.HTTPStatus)
	}

	role := &domain.Role{
		TenantID:     tenantID,
		Code:         req.Code,
		Name:         req.Name,
		Description:  req.Description,
		IsSystemRole: req.IsSystemRole,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}

	if len(req.PermissionIDs) > 0 {
		if err := s.validatePermissions(ctx, req.PermissionIDs); err != nil {
			return nil, err
		}
		if err := s.roleRepo.AssignPermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, mapDBError(err, apperror.ErrInternal)
		}
	}

	return s.roleRepo.FindByID(ctx, tenantID, role.ID)
}

func (s *roleService) Update(ctx context.Context, tenantID, roleID int64, req dto.UpdateRoleRequest) (*domain.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}

	if req.PermissionIDs != nil {
		if err := s.validatePermissions(ctx, req.PermissionIDs); err != nil {
			return nil, err
		}
		if err := s.roleRepo.AssignPermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, mapDBError(err, apperror.ErrInternal)
		}
	}

	return s.roleRepo.FindByID(ctx, tenantID, role.ID)
}

func (s *roleService) GetByID(ctx context.Context, tenantID, roleID int64) (*domain.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}
	return role, nil
}

func (s *roleService) List(ctx context.Context, tenantID int64) ([]domain.Role, error) {
	roles, err := s.roleRepo.List(ctx, tenantID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return roles, nil
}

func (s *roleService) Delete(ctx context.Context, tenantID, roleID int64) error {
	role, err := s.roleRepo.FindByID(ctx, tenantID, roleID)
	if err != nil {
		return mapDBError(err, apperror.ErrNotFound)
	}
	if role.IsSystemRole {
		return apperror.New("SYSTEM_ROLE", "cannot delete system role", apperror.ErrForbidden.HTTPStatus)
	}
	return mapDBError(s.roleRepo.Delete(ctx, tenantID, roleID), apperror.ErrInternal)
}

func (s *roleService) validatePermissions(ctx context.Context, ids []int64) error {
	perms, err := s.permissionRepo.FindByIDs(ctx, ids)
	if err != nil {
		return mapDBError(err, apperror.ErrInternal)
	}
	if len(perms) != len(ids) {
		return apperror.New("PERMISSION_NOT_FOUND", "one or more permissions not found", apperror.ErrBadRequest.HTTPStatus)
	}
	return nil
}
