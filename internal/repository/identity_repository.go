package repository

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"gorm.io/gorm"
)

type IdentityRepository interface {
	Create(ctx context.Context, identity *domain.UserIdentity) error
	FindByUserID(ctx context.Context, userID int64) ([]domain.UserIdentity, error)
	FindByIdentity(ctx context.Context, tenantID int64, provider, identity string) (*domain.UserIdentity, error)
	FindByProviderUserID(ctx context.Context, tenantID int64, provider, providerUserID string) (*domain.UserIdentity, error)
	Delete(ctx context.Context, id int64) error
}

type identityRepository struct {
	db *gorm.DB
}

func NewIdentityRepository(db *gorm.DB) IdentityRepository {
	return &identityRepository{db: db}
}

func (r *identityRepository) Create(ctx context.Context, identity *domain.UserIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
}

func (r *identityRepository) FindByUserID(ctx context.Context, userID int64) ([]domain.UserIdentity, error) {
	var identities []domain.UserIdentity
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&identities).Error
	return identities, err
}

func (r *identityRepository) FindByIdentity(ctx context.Context, tenantID int64, provider, identity string) (*domain.UserIdentity, error) {
	var result domain.UserIdentity
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = user_identities.user_id").
		Where("users.tenant_id = ? AND user_identities.provider = ? AND user_identities.identity = ?", tenantID, provider, identity).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *identityRepository) FindByProviderUserID(ctx context.Context, tenantID int64, provider, providerUserID string) (*domain.UserIdentity, error) {
	var result domain.UserIdentity
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = user_identities.user_id").
		Where("users.tenant_id = ? AND user_identities.provider = ? AND user_identities.provider_user_id = ?", tenantID, provider, providerUserID).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *identityRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UserIdentity{}, "id = ?", id).Error
}
