package repository

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	Update(ctx context.Context, role *domain.Role) error
	FindByID(ctx context.Context, tenantID, roleID int64) (*domain.Role, error)
	FindByCode(ctx context.Context, tenantID int64, code string) (*domain.Role, error)
	List(ctx context.Context, tenantID int64) ([]domain.Role, error)
	Delete(ctx context.Context, tenantID, roleID int64) error
	AssignPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error
	AssignToUser(ctx context.Context, userID int64, roleIDs []int64) error
	AddRoleToUser(ctx context.Context, userID, roleID int64) error
	ListUserRoles(ctx context.Context, tenantID, userID int64) ([]domain.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) scoped(ctx context.Context, tenantID int64) *gorm.DB {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) FindByID(ctx context.Context, tenantID, roleID int64) (*domain.Role, error) {
	var role domain.Role
	err := r.scoped(ctx, tenantID).
		Preload("Permissions").
		First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByCode(ctx context.Context, tenantID int64, code string) (*domain.Role, error) {
	var role domain.Role
	err := r.scoped(ctx, tenantID).First(&role, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context, tenantID int64) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.scoped(ctx, tenantID).Preload("Permissions").Order("created_at ASC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Delete(ctx context.Context, tenantID, roleID int64) error {
	return r.scoped(ctx, tenantID).Delete(&domain.Role{}, "id = ?", roleID).Error
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error; err != nil {
			return err
		}
		for _, pid := range permissionIDs {
			rp := domain.RolePermission{RoleID: roleID, PermissionID: pid}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepository) AssignToUser(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.UserRole{}).Error; err != nil {
			return err
		}
		for _, rid := range roleIDs {
			ur := domain.UserRole{UserID: userID, RoleID: rid}
			if err := tx.Create(&ur).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepository) AddRoleToUser(ctx context.Context, userID, roleID int64) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&domain.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *roleRepository) ListUserRoles(ctx context.Context, tenantID, userID int64) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.tenant_id = ?", userID, tenantID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}
