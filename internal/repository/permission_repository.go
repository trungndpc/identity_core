package repository

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *domain.Permission) error
	FindByID(ctx context.Context, id int64) (*domain.Permission, error)
	FindByCode(ctx context.Context, code string) (*domain.Permission, error)
	List(ctx context.Context) ([]domain.Permission, error)
	FindByIDs(ctx context.Context, ids []int64) ([]domain.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) FindByID(ctx context.Context, id int64) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindByCode(ctx context.Context, code string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).First(&permission, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).Order("module ASC, code ASC").Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) FindByIDs(ctx context.Context, ids []int64) ([]domain.Permission, error) {
	var permissions []domain.Permission
	if len(ids) == 0 {
		return permissions, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}
