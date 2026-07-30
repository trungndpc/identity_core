package service

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type TenantService interface {
	Create(ctx context.Context, req dto.CreateTenantRequest) (*domain.Tenant, error)
	Update(ctx context.Context, id int64, req dto.UpdateTenantRequest) (*domain.Tenant, error)
	GetByID(ctx context.Context, id int64) (*domain.Tenant, error)
	GetByCode(ctx context.Context, code string) (*domain.Tenant, error)
	List(ctx context.Context, page, pageSize int) ([]domain.Tenant, int64, error)
}

type tenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) Create(ctx context.Context, req dto.CreateTenantRequest) (*domain.Tenant, error) {
	if _, err := s.repo.FindByCode(ctx, req.Code); err == nil {
		return nil, apperror.New("TENANT_EXISTS", "tenant code already exists", apperror.ErrConflict.HTTPStatus)
	}

	status := req.Status
	if status == "" {
		status = domain.TenantStatusActive
	}

	tenant := &domain.Tenant{
		Code:     req.Code,
		Name:     req.Name,
		Status:   status,
		Metadata: toJSON(req.Metadata),
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return tenant, nil
}

func (s *tenantService) Update(ctx context.Context, id int64, req dto.UpdateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if len(req.Metadata) > 0 {
		tenant.Metadata = toJSON(req.Metadata)
	}

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return tenant, nil
}

func (s *tenantService) GetByID(ctx context.Context, id int64) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}
	return tenant, nil
}

func (s *tenantService) GetByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	tenant, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}
	return tenant, nil
}

func (s *tenantService) List(ctx context.Context, page, pageSize int) ([]domain.Tenant, int64, error) {
	tenants, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, mapDBError(err, apperror.ErrInternal)
	}
	return tenants, total, nil
}
