package service

import (
	"context"
	"errors"
	"strings"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
	"license-manager/pkg/utils"
)

// legacyRoleToRoleName 旧单角色值到角色名的映射。
//
// 兼容 v1 调用方：它们只传 role=admin/operator/viewer，
// 迁移期这些值需映射到等价角色，避免旧客户端创建出无角色的账号。
var legacyRoleToRoleName = map[string]string{
	"admin":    "super_admin",
	"operator": "operator",
	"viewer":   "viewer",
}

type userManagementService struct {
	userRepo repository.UserRepository
}

// NewUserManagementService 创建用户管理服务实例
func NewUserManagementService(userRepo repository.UserRepository) UserManagementService {
	return &userManagementService{userRepo: userRepo}
}

// resolveRoleNames 解析待分配的角色名。
//
// 优先使用 role_assignments；未提供时回退到 role 字段的映射，
// 使 v1 与 v2 两种调用方式都能创建出带角色的账号。
func resolveRoleNames(assignments []models.UserRoleAssignmentInput, legacyRole string) []string {
	if len(assignments) > 0 {
		names := make([]string, 0, len(assignments))
		for _, item := range assignments {
			if role := strings.TrimSpace(item.Role); role != "" {
				names = append(names, role)
			}
		}
		if len(names) > 0 {
			return names
		}
	}

	if mapped, ok := legacyRoleToRoleName[strings.TrimSpace(legacyRole)]; ok {
		return []string{mapped}
	}
	return nil
}

// GetUserList 查询用户列表
func (s *userManagementService) GetUserList(ctx context.Context, req *models.UserListRequest) (*models.UserListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if req == nil {
		req = &models.UserListRequest{}
	}

	items, total, err := s.userRepo.GetUserList(ctx, req)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	if items == nil {
		items = []models.UserListItem{}
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &models.UserListResponse{
		List:       items,
		Total:      total,
		Page:       req.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserDetail 查询用户详情
func (s *userManagementService) GetUserDetail(ctx context.Context, id string) (*models.UserDetailResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	detail, err := s.userRepo.GetUserDetail(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, i18n.NewI18nError("600013", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}
	return detail, nil
}

// CreateUser 创建用户
func (s *userManagementService) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.UserDetailResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	roleNames := resolveRoleNames(req.RoleAssignments, req.Role)
	if len(roleNames) == 0 {
		return nil, i18n.NewI18nError("600014", lang)
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// role 字段保持同步，供 v1 接口读取
	legacyRole := req.Role
	if legacyRole == "" {
		legacyRole = "viewer"
	}

	user := &models.User{
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(req.FullName),
		Phone:        req.Phone,
		Role:         legacyRole,
		Status:       req.Status,
	}

	if err := s.userRepo.CreateUser(ctx, user, roleNames); err != nil {
		if errors.Is(err, repository.ErrUserDuplicate) {
			return nil, i18n.NewI18nError("200011", lang)
		}
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, i18n.NewI18nError("600001", lang)
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return s.userRepo.GetUserDetail(ctx, user.ID)
}

// UpdateUser 更新用户
func (s *userManagementService) UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest) (*models.UserDetailResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, i18n.NewI18nError("600013", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	if req.Email != nil && *req.Email != "" {
		user.Email = strings.TrimSpace(*req.Email)
	}
	if req.FullName != nil && *req.FullName != "" {
		user.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Status != nil && *req.Status != "" {
		user.Status = *req.Status
	}

	if req.Password != nil && *req.Password != "" {
		hash, hashErr := utils.HashPassword(*req.Password)
		if hashErr != nil {
			return nil, i18n.NewI18nError("900004", lang, hashErr.Error())
		}
		user.PasswordHash = hash
		// 重设密码后解除锁定，避免新密码仍被锁定挡住
		user.LoginAttempts = 0
		user.LockedUntil = nil
	}

	// 降级最后一个超级管理员会导致系统无人可管
	if req.Status != nil && *req.Status != models.UserStatusActive {
		last, checkErr := s.userRepo.IsLastSuperAdmin(ctx, id)
		if checkErr != nil {
			return nil, i18n.NewI18nError("900005", lang)
		}
		if last {
			return nil, i18n.NewI18nError("600017", lang)
		}
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserDuplicate) {
			return nil, i18n.NewI18nError("200011", lang)
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	if req.RoleAssignments != nil {
		names := resolveRoleNames(req.RoleAssignments, "")
		if len(names) == 0 {
			return nil, i18n.NewI18nError("600014", lang)
		}
		if err := s.userRepo.ReplaceUserRoles(ctx, id, names); err != nil {
			if errors.Is(err, repository.ErrRoleNotFound) {
				return nil, i18n.NewI18nError("600001", lang)
			}
			return nil, i18n.NewI18nError("900005", lang)
		}
	}

	return s.userRepo.GetUserDetail(ctx, id)
}

// DeleteUser 删除用户
func (s *userManagementService) DeleteUser(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if _, err := s.userRepo.GetUserByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return i18n.NewI18nError("600013", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}

	last, err := s.userRepo.IsLastSuperAdmin(ctx, id)
	if err != nil {
		return i18n.NewI18nError("900005", lang)
	}
	if last {
		return i18n.NewI18nError("600017", lang)
	}

	if err := s.userRepo.DeleteUser(ctx, id); err != nil {
		return i18n.NewI18nError("900004", lang, err.Error())
	}
	return nil
}

// ResetPassword 重置用户密码
func (s *userManagementService) ResetPassword(ctx context.Context, id string, newPassword string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return i18n.NewI18nError("600013", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return i18n.NewI18nError("900004", lang, err.Error())
	}

	user.PasswordHash = hash
	// 重设密码即解除锁定：管理员已经确认了身份
	user.LoginAttempts = 0
	user.LockedUntil = nil

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return i18n.NewI18nError("900004", lang, err.Error())
	}
	return nil
}

// GetUserRoles 查询用户角色
func (s *userManagementService) GetUserRoles(ctx context.Context, userID string) ([]models.UserRoleItem, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	if roles == nil {
		roles = []models.UserRoleItem{}
	}
	return roles, nil
}
