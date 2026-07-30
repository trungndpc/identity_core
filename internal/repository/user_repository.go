package repository

import (
	"context"
	"fmt"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"gorm.io/gorm"
)

type UserFilter struct {
	TenantID int64
	Status   string
	Search   string
	Page     int
	PageSize int
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, tenantID, userID int64) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByIDWithRelations(ctx context.Context, tenantID, userID int64) (*domain.User, error)
	List(ctx context.Context, filter UserFilter) ([]domain.User, int64, error)
	Delete(ctx context.Context, tenantID, userID int64) error
	ExistsInTenant(ctx context.Context, tenantID, userID int64) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) scoped(ctx context.Context, tenantID int64) *gorm.DB {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, tenantID, userID int64) (*domain.User, error) {
	var user domain.User
	err := r.scoped(ctx, tenantID).First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByIDWithRelations(ctx context.Context, tenantID, userID int64) (*domain.User, error) {
	var user domain.User
	err := r.scoped(ctx, tenantID).
		Preload("Identities").
		Preload("Roles").
		Preload("Roles.Permissions").
		First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, filter UserFilter) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.scoped(ctx, filter.TenantID).Model(&domain.User{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Search != "" {
		like := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where(
			"full_name ILIKE ? OR display_name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR username ILIKE ?",
			like, like, like, like, like,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&users).Error
	return users, total, err
}

func (r *userRepository) Delete(ctx context.Context, tenantID, userID int64) error {
	return r.scoped(ctx, tenantID).Delete(&domain.User{}, "id = ?", userID).Error
}

func (r *userRepository) ExistsInTenant(ctx context.Context, tenantID, userID int64) (bool, error) {
	var count int64
	err := r.scoped(ctx, tenantID).Model(&domain.User{}).Where("id = ?", userID).Count(&count).Error
	return count > 0, err
}
