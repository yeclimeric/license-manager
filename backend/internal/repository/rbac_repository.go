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

// wildcardPermission 通配权限，超级管理员持有
const wildcardPermission = "*"

type rbacRepository struct {
	db *gorm.DB
}

// NewRBACRepository 创建 RBAC 数据访问实例
func NewRBACRepository(db *gorm.DB) RBACRepository {
	return &rbacRepository{db: db}
}

// GetMenuList 查询菜单列表（扁平）
func (r *rbacRepository) GetMenuList(ctx context.Context) ([]models.Menu, error) {
	var menus []models.Menu
	if err := r.db.WithContext(ctx).
		Order("sort_order ASC").
		Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("failed to list menus: %w", err)
	}
	return menus, nil
}

// GetMenuTree 查询用户可见的菜单树
//
// userID 为空时返回全部启用菜单（用于角色配置界面）。
// 用户持通配权限时同样返回全部。
func (r *rbacRepository) GetMenuTree(ctx context.Context, userID string) ([]models.Menu, error) {
	menus, err := r.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}

	if userID != "" {
		allowed, isWildcard, err := r.allowedMenuIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !isWildcard {
			filtered := make([]models.Menu, 0, len(menus))
			for _, menu := range menus {
				if allowed[menu.ID] {
					filtered = append(filtered, menu)
				}
			}
			menus = filtered
		}
	}

	// 统计子菜单数量，供角色配置界面判断能否删除
	childCount := map[string]int{}
	for _, menu := range menus {
		if menu.ParentID != nil {
			childCount[*menu.ParentID]++
		}
	}

	tree := r.buildMenuTree(menus, childCount)
	return tree, nil
}

// buildMenuTree 将扁平菜单组装为树
func (r *rbacRepository) buildMenuTree(menus []models.Menu, childCount map[string]int) []models.Menu {
	byParent := map[string][]models.Menu{}

	for _, menu := range menus {
		// 不可见或停用的节点不参与组装
		if menu.Hidden || menu.Status != "active" {
			continue
		}
		if menu.ParentID == nil || *menu.ParentID == "" {
			continue
		}
		byParent[*menu.ParentID] = append(byParent[*menu.ParentID], menu)
	}

	// attach 组装指定父节点的子菜单
	var attach func(parentID string, depth int) []models.Menu
	attach = func(parentID string, depth int) []models.Menu {
		// 深度上限防御环形引用
		if depth > 5 {
			return nil
		}
		children := byParent[parentID]
		result := make([]models.Menu, 0, len(children))
		for _, child := range children {
			child.Children = attach(child.ID, depth+1)
			// 分组下无可见子项时不显示为空目录
			if child.Type == models.MenuTypeDirectory && len(child.Children) == 0 {
				continue
			}
			result = append(result, child)
		}
		return result
	}

	roots := make([]models.Menu, 0, len(menus))
	for _, menu := range menus {
		if menu.Hidden || menu.Status != "active" {
			continue
		}
		if menu.ParentID != nil && *menu.ParentID != "" {
			continue
		}

		menu.Children = attach(menu.ID, 1)
		if menu.Type == models.MenuTypeDirectory && len(menu.Children) == 0 {
			continue
		}
		menu.CanDelete = !menu.IsSystem && childCount[menu.ID] == 0
		for i := range menu.Children {
			menu.Children[i].CanDelete = !menu.Children[i].IsSystem &&
				childCount[menu.Children[i].ID] == 0
		}
		roots = append(roots, menu)
	}

	return roots
}

// allowedMenuIDs 返回用户可见菜单ID集合
func (r *rbacRepository) allowedMenuIDs(ctx context.Context, userID string) (map[string]bool, bool, error) {
	// 判断是否持有通配权限
	var wildcardCount int64
	if err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Joins("JOIN user_roles ur ON ur.role_id = role_permissions.role_id").
		Where("ur.user_id = ? AND role_permissions.permission = ?", userID, wildcardPermission).
		Count(&wildcardCount).Error; err != nil {
		return nil, false, fmt.Errorf("failed to check wildcard permission: %w", err)
	}
	if wildcardCount > 0 {
		return nil, true, nil
	}

	var menuIDs []string
	if err := r.db.WithContext(ctx).Model(&models.RoleMenu{}).
		Joins("JOIN user_roles ur ON ur.role_id = role_menus.role_id").
		Where("ur.user_id = ?", userID).
		Distinct().
		Pluck("role_menus.menu_id", &menuIDs).Error; err != nil {
		return nil, false, fmt.Errorf("failed to list user menus: %w", err)
	}

	ids := make(map[string]bool, len(menuIDs))
	for _, id := range menuIDs {
		ids[id] = true
	}
	return ids, false, nil
}

// GetUserPermissions 返回用户全部权限码
func (r *rbacRepository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	var permissions []string
	if err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Joins("JOIN user_roles ur ON ur.role_id = role_permissions.role_id").
		Where("ur.user_id = ?", userID).
		Distinct().
		Pluck("role_permissions.permission", &permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to list user permissions: %w", err)
	}
	return permissions, nil
}

// HasPermission 判断用户是否持有指定权限
func (r *rbacRepository) HasPermission(ctx context.Context, userID string, permission string) (bool, error) {
	if permission == "" {
		return true, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Joins("JOIN user_roles ur ON ur.role_id = role_permissions.role_id").
		Where("ur.user_id = ? AND (role_permissions.permission = ? OR role_permissions.permission = ?)",
			userID, permission, wildcardPermission).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return count > 0, nil
}

// GetPermissionMatrix 查询权限矩阵
func (r *rbacRepository) GetPermissionMatrix(ctx context.Context) ([]models.PermissionResource, error) {
	var resources []models.PermissionResource
	if err := r.db.WithContext(ctx).
		Order("sort_order ASC").
		Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to list permission resources: %w", err)
	}

	for i := range resources {
		resources[i].Actions = r.actionsForResource(resources[i].Resource)
	}
	return resources, nil
}

// actionsForResource 返回资源支持的动作
//
// 仪表盘与审计日志只有读取动作，其余资源支持完整增删改查。
func (r *rbacRepository) actionsForResource(resource string) []models.PermissionAction {
	switch resource {
	case "dashboard", "audit_log":
		return []models.PermissionAction{{Code: models.ActionRead, Name: "查看"}}
	default:
		return models.StandardActions
	}
}

// GetRoleList 查询角色列表
func (r *rbacRepository) GetRoleList(ctx context.Context, keyword string) ([]models.Role, error) {
	query := r.db.WithContext(ctx).Model(&models.Role{})

	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("name LIKE ? OR display_name LIKE ?", like, like)
	}

	var roles []models.Role
	if err := query.Order("is_system DESC, created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	for i := range roles {
		r.decorateRole(ctx, &roles[i])
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (r *rbacRepository) GetRoleByID(ctx context.Context, id string) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	r.decorateRole(ctx, &role)
	return &role, nil
}

// decorateRole 补充角色的权限、菜单与用户数
func (r *rbacRepository) decorateRole(ctx context.Context, role *models.Role) {
	var permissions []string
	if err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Where("role_id = ?", role.ID).
		Pluck("permission", &permissions).Error; err == nil {
		role.Permissions = permissions
	}

	var menuIDs []string
	if err := r.db.WithContext(ctx).Model(&models.RoleMenu{}).
		Where("role_id = ?", role.ID).
		Pluck("menu_id", &menuIDs).Error; err == nil {
		role.MenuIDs = menuIDs
	}

	var userCount int64
	if err := r.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("role_id = ?", role.ID).
		Count(&userCount).Error; err == nil {
		role.UserCount = int(userCount)
	}

	role.CanDelete = !role.IsSystem && role.UserCount == 0
}

// CreateRole 创建角色
func (r *rbacRepository) CreateRole(ctx context.Context, role *models.Role, permissions []string, menuIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			if isDuplicateKeyError(err) {
				return ErrRoleDuplicate
			}
			return fmt.Errorf("failed to create role: %w", err)
		}

		now := time.Now()
		for _, permission := range permissions {
			if err := tx.Create(&models.RolePermission{
				RoleID: role.ID, Permission: permission, CreatedAt: now,
			}).Error; err != nil {
				return fmt.Errorf("failed to bind permission: %w", err)
			}
		}
		for _, menuID := range menuIDs {
			if err := tx.Create(&models.RoleMenu{
				RoleID: role.ID, MenuID: menuID, CreatedAt: now,
			}).Error; err != nil {
				return fmt.Errorf("failed to bind menu: %w", err)
			}
		}
		return nil
	})
}

// UpdateRole 更新角色
//
// 标识符 Name 不在更新范围内：被用户角色关联引用。
func (r *rbacRepository) UpdateRole(ctx context.Context, role *models.Role, permissions []string, menuIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Role{}).Where("id = ?", role.ID).
			Select("display_name", "description", "status", "updated_at").
			Updates(role)
		if result.Error != nil {
			return fmt.Errorf("failed to update role: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrRoleNotFound
		}

		// 全量替换关联：先清后插，语义与"提交完整集合"一致
		if permissions != nil {
			if err := tx.Where("role_id = ?", role.ID).Delete(&models.RolePermission{}).Error; err != nil {
				return fmt.Errorf("failed to clear permissions: %w", err)
			}
			now := time.Now()
			for _, permission := range permissions {
				if err := tx.Create(&models.RolePermission{
					RoleID: role.ID, Permission: permission, CreatedAt: now,
				}).Error; err != nil {
					return fmt.Errorf("failed to bind permission: %w", err)
				}
			}
		}
		if menuIDs != nil {
			if err := tx.Where("role_id = ?", role.ID).Delete(&models.RoleMenu{}).Error; err != nil {
				return fmt.Errorf("failed to clear menus: %w", err)
			}
			now := time.Now()
			for _, menuID := range menuIDs {
				if err := tx.Create(&models.RoleMenu{
					RoleID: role.ID, MenuID: menuID, CreatedAt: now,
				}).Error; err != nil {
					return fmt.Errorf("failed to bind menu: %w", err)
				}
			}
		}
		return nil
	})
}

// DeleteRole 删除角色
func (r *rbacRepository) DeleteRole(ctx context.Context, id string) error {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to load role: %w", err)
	}

	if role.IsSystem {
		return ErrRoleSystemProtected
	}

	// 已分配给用户的角色不可删除，否则用户会失去全部权限
	var userCount int64
	if err := r.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count role users: %w", err)
	}
	if userCount > 0 {
		return ErrRoleInUse
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&models.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Role{}, "id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}
		return nil
	})
}

// GetUserRoleIDs 返回用户绑定的角色ID列表
func (r *rbacRepository) GetUserRoleIDs(ctx context.Context, userID string) ([]string, error) {
	var roleIDs []string
	if err := r.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}
	return roleIDs, nil
}

// SetUserRoles 设置用户绑定的角色（覆盖式）
func (r *rbacRepository) SetUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return fmt.Errorf("failed to clear user roles: %w", err)
		}
		now := time.Now()
		for _, roleID := range roleIDs {
			if err := tx.Create(&models.UserRole{
				UserID: userID, RoleID: roleID, CreatedAt: now,
			}).Error; err != nil {
				return fmt.Errorf("failed to bind user role: %w", err)
			}
		}
		return nil
	})
}

// CreateMenu 创建菜单
func (r *rbacRepository) CreateMenu(ctx context.Context, menu *models.Menu) error {
	if err := r.validateParent(ctx, menu.ParentID, menu.ID); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(menu).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrMenuDuplicate
		}
		return fmt.Errorf("failed to create menu: %w", err)
	}
	return nil
}

// UpdateMenu 更新菜单
func (r *rbacRepository) UpdateMenu(ctx context.Context, menu *models.Menu) error {
	if err := r.validateParent(ctx, menu.ParentID, menu.ID); err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Model(&models.Menu{}).Where("id = ?", menu.ID).
		Select("name_key", "type", "path", "icon", "parent_id", "permission",
			"sort_order", "status", "hidden", "description", "updated_at").
		Updates(menu)
	if result.Error != nil {
		return fmt.Errorf("failed to update menu: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrMenuNotFound
	}
	return nil
}

// validateParent 校验父菜单存在且不会形成环
func (r *rbacRepository) validateParent(ctx context.Context, parentID *string, selfID string) error {
	if parentID == nil || *parentID == "" {
		return nil
	}

	// 父菜单自身必须存在
	var parent models.Menu
	if err := r.db.WithContext(ctx).Where("id = ?", *parentID).First(&parent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuParentNotFound
		}
		return fmt.Errorf("failed to load parent menu: %w", err)
	}

	// 沿父链向上检查，selfID 不能成为自己的祖先
	visited := map[string]bool{selfID: true}
	current := *parentID
	for depth := 0; depth < 10; depth++ {
		if visited[current] {
			return ErrMenuCyclicReference
		}
		visited[current] = true

		var node models.Menu
		if err := r.db.WithContext(ctx).Select("id", "parent_id").Where("id = ?", current).First(&node).Error; err != nil {
			// 父链断裂时按合法处理，删除父节点后的历史数据不应阻塞
			return nil
		}
		if node.ParentID == nil || *node.ParentID == "" {
			return nil
		}
		current = *node.ParentID
	}
	return ErrMenuCyclicReference
}

// DeleteMenu 删除菜单
func (r *rbacRepository) DeleteMenu(ctx context.Context, id string) error {
	var menu models.Menu
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return fmt.Errorf("failed to load menu: %w", err)
	}

	// 内置归属判断先于一切：内置菜单即便被角色绑定也不可删除，
	// 否则运营会看到"已绑定角色"而误以为可以先解绑再删
	if menu.IsSystem {
		return ErrMenuSystemProtected
	}

	// 存在子菜单时拒绝删除，避免产生孤儿节点
	var childCount int64
	if err := r.db.WithContext(ctx).Model(&models.Menu{}).
		Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return fmt.Errorf("failed to count child menus: %w", err)
	}
	if childCount > 0 {
		return ErrMenuHasChildren
	}

	// 已绑定给角色的菜单不可删除，否则角色会指向不存在的菜单
	var roleCount int64
	if err := r.db.WithContext(ctx).Model(&models.RoleMenu{}).
		Where("menu_id = ?", id).Count(&roleCount).Error; err != nil {
		return fmt.Errorf("failed to count role menus: %w", err)
	}
	if roleCount > 0 {
		return ErrMenuInUse
	}

	if err := r.db.WithContext(ctx).Delete(&models.Menu{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete menu: %w", err)
	}
	return nil
}

// GetMenuByID 根据ID获取菜单
func (r *rbacRepository) GetMenuByID(ctx context.Context, id string) (*models.Menu, error) {
	var menu models.Menu
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuNotFound
		}
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}
	return &menu, nil
}
