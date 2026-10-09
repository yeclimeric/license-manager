package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"license-manager/internal/models"
)

// userAssignment 把角色名解析为角色记录
//
// 按名称而非ID匹配：角色可能被改名，ID 才是稳定标识，
// 但对外契约（企业版实测）用的是角色名，此处做一次解析。
func (r *userRepository) resolveRolesByNames(ctx context.Context, names []string) ([]models.Role, error) {
	if len(names) == 0 {
		return nil, nil
	}

	var roles []models.Role
	if err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to resolve roles: %w", err)
	}

	if len(roles) != len(dedupeStrings(names)) {
		found := map[string]bool{}
		for _, role := range roles {
			found[role.Name] = true
		}
		for _, name := range dedupeStrings(names) {
			if !found[name] {
				return nil, ErrRoleNotFound
			}
		}
	}
	return roles, nil
}

// dedupeStrings 去重并保持顺序
func dedupeStrings(items []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}

// GetUserRoles 查询用户绑定的角色
func (r *userRepository) GetUserRoles(ctx context.Context, userID string) ([]models.UserRoleItem, error) {
	type row struct {
		ID          string
		Name        string
		DisplayName string
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("roles").
		Select("roles.id, roles.name, roles.display_name").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Order("roles.is_system DESC, roles.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	items := make([]models.UserRoleItem, 0, len(rows))
	for _, item := range rows {
		items = append(items, models.UserRoleItem{
			ID: item.ID, Name: item.Name, DisplayName: item.DisplayName,
		})
	}
	return items, nil
}

// batchUserRoles 批量查询多个用户的角色
func (r *userRepository) batchUserRoles(ctx context.Context, userIDs []string) (map[string][]models.UserRoleItem, error) {
	result := map[string][]models.UserRoleItem{}
	if len(userIDs) == 0 {
		return result, nil
	}

	// 字段平铺而非嵌套结构体：GORM 的 Scan 无法映射嵌套 struct 字段
	type row struct {
		UserID      string
		RoleID      string
		Name        string
		DisplayName string
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("user_roles").
		Select("user_roles.user_id AS user_id, roles.id AS role_id, roles.name AS name, roles.display_name AS display_name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id IN ?", userIDs).
		Order("roles.is_system DESC, roles.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to batch user roles: %w", err)
	}

	for _, item := range rows {
		result[item.UserID] = append(result[item.UserID], models.UserRoleItem{
			ID: item.RoleID, Name: item.Name, DisplayName: item.DisplayName,
		})
	}
	return result, nil
}

// GetUserList 查询用户列表
func (r *userRepository) GetUserList(ctx context.Context, req *models.UserListRequest) ([]models.UserListItem, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	sortBy := "created_at"
	switch req.Sort {
	case "updated_at", "username", "full_name":
		sortBy = req.Sort
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.Order, "asc") {
		sortOrder = "ASC"
	}

	query := r.db.WithContext(ctx).Model(&models.User{})

	if keyword := strings.TrimSpace(req.Search); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?", like, like, like)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var users []models.User
	if err := query.Order(sortBy + " " + sortOrder).
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	roleMap, err := r.batchUserRoles(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	items := make([]models.UserListItem, 0, len(users))
	for _, user := range users {
		roles := roleMap[user.ID]
		if roles == nil {
			roles = []models.UserRoleItem{}
		}
		items = append(items, models.UserListItem{
			ID: user.ID, Username: user.Username, Email: user.Email,
			FullName: user.FullName, Phone: user.Phone,
			Role: user.Role, Status: user.Status,
			Remark: nil, LastLoginAt: user.LastLoginAt,
			CreatedAt: user.CreatedAt, Roles: roles,
		})
	}

	return items, total, nil
}

// GetUserDetail 查询用户详情（含登录安全信息）
func (r *userRepository) GetUserDetail(ctx context.Context, id string) (*models.UserDetailResponse, error) {
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	roles, err := r.GetUserRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []models.UserRoleItem{}
	}

	return &models.UserDetailResponse{
		UserListItem: models.UserListItem{
			ID: user.ID, Username: user.Username, Email: user.Email,
			FullName: user.FullName, Phone: user.Phone,
			Role: user.Role, Status: user.Status,
			LastLoginAt: user.LastLoginAt, CreatedAt: user.CreatedAt, Roles: roles,
		},
		LoginAttempts: user.LoginAttempts,
		LockedUntil:   user.LockedUntil,
		LastOnlineIP:  user.LastLoginIP,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// CreateUser 创建用户并绑定角色
func (r *userRepository) CreateUser(ctx context.Context, user *models.User, roleNames []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			if isDuplicateKeyError(err) {
				return ErrUserDuplicate
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		roles, err := r.resolveRolesByNames(ctx, roleNames)
		if err != nil {
			return err
		}
		now := time.Now()
		for _, role := range roles {
			if err := tx.Create(&models.UserRole{
				UserID: user.ID, RoleID: role.ID, CreatedAt: now,
			}).Error; err != nil {
				return fmt.Errorf("failed to assign role: %w", err)
			}
		}
		return nil
	})
}

// ReplaceUserRoles 覆盖式设置用户角色
func (r *userRepository) ReplaceUserRoles(ctx context.Context, userID string, roleNames []string) error {
	roles, err := r.resolveRolesByNames(ctx, roleNames)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return ErrUserAssignmentEmpty
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return fmt.Errorf("failed to clear user roles: %w", err)
		}
		now := time.Now()
		for _, role := range roles {
			if err := tx.Create(&models.UserRole{
				UserID: userID, RoleID: role.ID, CreatedAt: now,
			}).Error; err != nil {
				return fmt.Errorf("failed to assign role: %w", err)
			}
		}
		return nil
	})
}

// CountUsersByStatus 统计各状态用户数
func (r *userRepository) CountUsersByStatus(ctx context.Context) (map[string]int64, error) {
	type row struct {
		Status string
		Total  int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).Model(&models.User{}).
		Select("status, COUNT(*) AS total").
		Group("status").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to count users by status: %w", err)
	}

	result := map[string]int64{}
	for _, item := range rows {
		result[item.Status] = item.Total
	}
	return result, nil
}

// IsLastSuperAdmin 判断是否为最后一个超级管理员
//
// 用于阻止把最后一个管理员降级或删除，否则系统将无人可管。
func (r *userRepository) IsLastSuperAdmin(ctx context.Context, userID string) (bool, error) {
	// 取超级管理员角色ID
	var superRole models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", "super_admin").First(&superRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to load super admin role: %w", err)
	}

	// 该用户是否持有超级管理员
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, superRole.ID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check super admin assignment: %w", err)
	}
	if count == 0 {
		return false, nil
	}

	// 系统中还有几个超级管理员
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("role_id = ?", superRole.ID).
		Count(&total).Error; err != nil {
		return false, fmt.Errorf("failed to count super admins: %w", err)
	}
	return total <= 1, nil
}
