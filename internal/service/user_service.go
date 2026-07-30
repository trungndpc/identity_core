package service

import (
	"context"

	"github.com/updev/galaxy/identity_core/internal/domain"
	"github.com/updev/galaxy/identity_core/internal/dto"
	"github.com/updev/galaxy/identity_core/internal/repository"
	"github.com/updev/galaxy/identity_core/pkg/apperror"
)

type UserService interface {
	Create(ctx context.Context, tenantID int64, req dto.CreateUserRequest) (*domain.User, error)
	Update(ctx context.Context, tenantID, userID int64, req dto.UpdateUserRequest) (*domain.User, error)
	GetByID(ctx context.Context, tenantID, userID int64, withRelations bool) (*domain.User, error)
	List(ctx context.Context, tenantID int64, query dto.ListUsersQuery) ([]domain.User, int64, error)
	Delete(ctx context.Context, tenantID, userID int64) error
	AssignRoles(ctx context.Context, tenantID, userID int64, roleIDs []int64) error
}

type userService struct {
	userRepo     repository.UserRepository
	identityRepo repository.IdentityRepository
	roleRepo     repository.RoleRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	identityRepo repository.IdentityRepository,
	roleRepo repository.RoleRepository,
) UserService {
	return &userService{
		userRepo:     userRepo,
		identityRepo: identityRepo,
		roleRepo:     roleRepo,
	}
}

func (s *userService) Create(ctx context.Context, tenantID int64, req dto.CreateUserRequest) (*domain.User, error) {
	status := req.Status
	if status == "" {
		status = domain.UserStatusActive
	}

	user := &domain.User{
		TenantID:    tenantID,
		FullName:    req.FullName,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
		Gender:      req.Gender,
		Birthday:    req.Birthday,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		City:        req.City,
		District:    req.District,
		Ward:        req.Ward,
		Username:    req.Username,
		Status:      status,
		Metadata:    toJSON(req.Metadata),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}

	if req.Password != "" || req.Username != "" || req.Email != "" || req.Phone != "" {
		identity := req.Username
		if identity == "" {
			identity = req.Email
		}
		if identity == "" {
			identity = req.Phone
		}
		if identity != "" {
			passwordHash := ""
			if req.Password != "" {
				hash, err := hashPassword(req.Password)
				if err != nil {
					return nil, apperror.Wrap(err, apperror.ErrInternal.Code, "failed to hash password", apperror.ErrInternal.HTTPStatus)
				}
				passwordHash = hash
			}
			uid := &domain.UserIdentity{
				UserID:       user.ID,
				Provider:     domain.IdentityProviderLocal,
				Identity:     identity,
				PasswordHash: passwordHash,
			}
			if err := s.identityRepo.Create(ctx, uid); err != nil {
				return nil, mapDBError(err, apperror.ErrInternal)
			}
		}
	}

	if len(req.RoleIDs) > 0 {
		if err := s.validateRolesInTenant(ctx, tenantID, req.RoleIDs); err != nil {
			return nil, err
		}
		if err := s.roleRepo.AssignToUser(ctx, user.ID, req.RoleIDs); err != nil {
			return nil, mapDBError(err, apperror.ErrInternal)
		}
	}

	return s.userRepo.FindByIDWithRelations(ctx, tenantID, user.ID)
}

func (s *userService) Update(ctx context.Context, tenantID, userID int64, req dto.UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}

	applyUserUpdate(user, req)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, mapDBError(err, apperror.ErrInternal)
	}
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, tenantID, userID int64, withRelations bool) (*domain.User, error) {
	if withRelations {
		user, err := s.userRepo.FindByIDWithRelations(ctx, tenantID, userID)
		if err != nil {
			return nil, mapDBError(err, apperror.ErrNotFound)
		}
		return user, nil
	}

	user, err := s.userRepo.FindByID(ctx, tenantID, userID)
	if err != nil {
		return nil, mapDBError(err, apperror.ErrNotFound)
	}
	return user, nil
}

func (s *userService) List(ctx context.Context, tenantID int64, query dto.ListUsersQuery) ([]domain.User, int64, error) {
	query.Normalize()
	users, total, err := s.userRepo.List(ctx, repository.UserFilter{
		TenantID: tenantID,
		Status:   query.Status,
		Search:   query.Search,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return nil, 0, mapDBError(err, apperror.ErrInternal)
	}
	return users, total, nil
}

func (s *userService) Delete(ctx context.Context, tenantID, userID int64) error {
	if err := s.userRepo.Delete(ctx, tenantID, userID); err != nil {
		return mapDBError(err, apperror.ErrNotFound)
	}
	return nil
}

func (s *userService) AssignRoles(ctx context.Context, tenantID, userID int64, roleIDs []int64) error {
	if _, err := s.userRepo.FindByID(ctx, tenantID, userID); err != nil {
		return mapDBError(err, apperror.ErrNotFound)
	}
	if err := s.validateRolesInTenant(ctx, tenantID, roleIDs); err != nil {
		return err
	}
	return mapDBError(s.roleRepo.AssignToUser(ctx, userID, roleIDs), apperror.ErrInternal)
}

func (s *userService) validateRolesInTenant(ctx context.Context, tenantID int64, roleIDs []int64) error {
	for _, roleID := range roleIDs {
		if _, err := s.roleRepo.FindByID(ctx, tenantID, roleID); err != nil {
			return apperror.New("ROLE_NOT_FOUND", "one or more roles not found in tenant", apperror.ErrBadRequest.HTTPStatus)
		}
	}
	return nil
}

func applyUserUpdate(user *domain.User, req dto.UpdateUserRequest) {
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Birthday != nil {
		user.Birthday = req.Birthday
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.City != nil {
		user.City = *req.City
	}
	if req.District != nil {
		user.District = *req.District
	}
	if req.Ward != nil {
		user.Ward = *req.Ward
	}
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if len(req.Metadata) > 0 {
		user.Metadata = toJSON(req.Metadata)
	}
}
