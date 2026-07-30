package repository

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"gorm.io/gorm"
)

type RelationshipRepository interface {
	Create(ctx context.Context, rel *domain.UserRelationship) error
	Update(ctx context.Context, rel *domain.UserRelationship) error
	FindByID(ctx context.Context, tenantID, id int64) (*domain.UserRelationship, error)
	ListByUser(ctx context.Context, tenantID, userID int64) ([]domain.UserRelationship, error)
	Delete(ctx context.Context, tenantID, id int64) error
}

type relationshipRepository struct {
	db *gorm.DB
}

func NewRelationshipRepository(db *gorm.DB) RelationshipRepository {
	return &relationshipRepository{db: db}
}

func (r *relationshipRepository) scoped(ctx context.Context, tenantID int64) *gorm.DB {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
}

func (r *relationshipRepository) Create(ctx context.Context, rel *domain.UserRelationship) error {
	return r.db.WithContext(ctx).Create(rel).Error
}

func (r *relationshipRepository) Update(ctx context.Context, rel *domain.UserRelationship) error {
	return r.db.WithContext(ctx).Save(rel).Error
}

func (r *relationshipRepository) FindByID(ctx context.Context, tenantID, id int64) (*domain.UserRelationship, error) {
	var rel domain.UserRelationship
	err := r.scoped(ctx, tenantID).First(&rel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *relationshipRepository) ListByUser(ctx context.Context, tenantID, userID int64) ([]domain.UserRelationship, error) {
	var rels []domain.UserRelationship
	err := r.scoped(ctx, tenantID).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&rels).Error
	return rels, err
}

func (r *relationshipRepository) Delete(ctx context.Context, tenantID, id int64) error {
	return r.scoped(ctx, tenantID).Delete(&domain.UserRelationship{}, "id = ?", id).Error
}
