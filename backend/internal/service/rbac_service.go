package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
)

// roleNamePattern 角色标识格式：小写字母开头，可含小写字母、数字、下划线
var roleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type rbacService struct {
	repo     repository.RBACRepository
	userRepo repository.UserRepository
}

// NewRBACService 创建菜单权限角色服务实例
func NewRBACService(repo repository.RBACRepository, userRepo repository.UserRepository) RBACService {
	return &rbacService{repo: repo, userRepo: userRepo}
}

// GetUserMenuTree 获取当前用户可见的菜单树
func (s *rbacService) GetUserMenuTree(ctx context.Context) (*models.MenuTreeResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	userID := pkgcontext.GetUserIDFromContext(ctx)
	if userID == "" {
		// 未登录时返回空树而非报错，登录页需要渲染而非崩溃
		return &models.MenuTreeResponse{List: []models.Menu{}}, nil
	}

	menus, err := s.repo.GetMenuTree(ctx, userID)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	if menus == nil {
		menus = []models.Menu{}
	}

	return &models.MenuTreeResponse{List: menus, Total: len(menus)}, nil
}

// GetAllMenuTree 获取全部菜单树（菜单管理界面用）
func (s *rbacService) GetAllMenuTree(ctx context.Context) (*models.MenuTreeResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	menus, err := s.repo.GetMenuTree(ctx, "")
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	if menus == nil {
		menus = []models.Menu{}
	}

	return &models.MenuTreeResponse{List: menus, Total: len(menus)}, nil
}

// GetPermissionMatrix 获取权限矩阵
func (s *rbacService) GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrixResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	resources, err := s.repo.GetPermissionMatrix(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	if resources == nil {
		resources = []models.PermissionResource{}
	}

	return &models.PermissionMatrixResponse{
		Resources: resources,
		Actions:   models.StandardActions,
	}, nil
}

// GetRoleList 获取角色列表
func (s *rbacService) GetRoleList(ctx context.Context, keyword string) ([]models.Role, int, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	roles, err := s.repo.GetRoleList(ctx, keyword)
	if err != nil {
		return nil, 0, i18n.NewI18nError("900005", lang)
	}
	return roles, len(roles), nil
}

// GetRole 获取角色详情
func (s *rbacService) GetRole(ctx context.Context, id string) (*models.Role, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, i18n.NewI18nError("600001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}
	return role, nil
}

// CreateRole 创建角色
func (s *rbacService) CreateRole(ctx context.Context, req *models.RoleCreateRequest) (*models.Role, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	name := strings.TrimSpace(req.Name)
	if !roleNamePattern.MatchString(name) {
		return nil, i18n.NewI18nError("600005", lang)
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 校验权限码必须来自已定义的资源，避免写入永远无法匹配的权限
	if err := s.validatePermissions(ctx, req.Permissions, lang); err != nil {
		return nil, err
	}

	role := &models.Role{
		Name:        name,
		DisplayName: displayName,
		Description: strings.TrimSpace(req.Description),
		Status:      "active",
	}

	if err := s.repo.CreateRole(ctx, role, req.Permissions, req.MenuIDs); err != nil {
		if errors.Is(err, repository.ErrRoleDuplicate) {
			return nil, i18n.NewI18nError("600002", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	return s.repo.GetRoleByID(ctx, role.ID)
}

// UpdateRole 更新角色
func (s *rbacService) UpdateRole(ctx context.Context, id string, req *models.RoleUpdateRequest) (*models.Role, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	existing, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, i18n.NewI18nError("600001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	// 内置角色的标识与归属不可改，但展示名与权限可以调整
	updated := *existing
	if req.DisplayName != nil && strings.TrimSpace(*req.DisplayName) != "" {
		updated.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	if req.Description != nil {
		updated.Description = strings.TrimSpace(*req.Description)
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" {
			return nil, i18n.NewI18nError("500010", lang)
		}
		updated.Status = *req.Status
	}

	if req.Permissions != nil {
		if err := s.validatePermissions(ctx, req.Permissions, lang); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateRole(ctx, &updated, req.Permissions, req.MenuIDs); err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, i18n.NewI18nError("600001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	return s.repo.GetRoleByID(ctx, id)
}

// DeleteRole 删除角色
func (s *rbacService) DeleteRole(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.repo.DeleteRole(ctx, id); err != nil {
		switch {
		case errors.Is(err, repository.ErrRoleNotFound):
			return i18n.NewI18nError("600001", lang)
		case errors.Is(err, repository.ErrRoleSystemProtected):
			return i18n.NewI18nError("600003", lang)
		case errors.Is(err, repository.ErrRoleInUse):
			return i18n.NewI18nError("600004", lang)
		default:
			return i18n.NewI18nError("900005", lang)
		}
	}
	return nil
}

// SetUserRoles 设置用户绑定的角色
func (s *rbacService) SetUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if len(roleIDs) == 0 {
		return i18n.NewI18nError("600014", lang)
	}

	// 确认用户存在，避免为不存在的用户建立角色关联
	if _, err := s.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return i18n.NewI18nError("600013", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}

	// 确认每个角色都存在
	for _, roleID := range roleIDs {
		if strings.TrimSpace(roleID) == "" {
			continue
		}
		if _, err := s.repo.GetRoleByID(ctx, roleID); err != nil {
			if errors.Is(err, repository.ErrRoleNotFound) {
				return i18n.NewI18nError("600001", lang)
			}
			return i18n.NewI18nError("900005", lang)
		}
	}

	if err := s.repo.SetUserRoles(ctx, userID, roleIDs); err != nil {
		return i18n.NewI18nError("900005", lang)
	}
	return nil
}

// GetUserRoles 获取用户绑定的角色ID
func (s *rbacService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	roleIDs, err := s.repo.GetUserRoleIDs(ctx, userID)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	return roleIDs, nil
}

// CreateMenu 创建菜单
func (s *rbacService) CreateMenu(ctx context.Context, menu *models.Menu) (*models.Menu, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if strings.TrimSpace(menu.NameKey) == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}
	if menu.Type == "" {
		menu.Type = models.MenuTypeMenu
	}
	if menu.Type != models.MenuTypeMenu && menu.Type != models.MenuTypeDirectory {
		return nil, i18n.NewI18nError("900001", lang)
	}
	// 菜单项必须有路径，分组目录没有
	if menu.Type == models.MenuTypeMenu && strings.TrimSpace(menu.Path) == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	menu.NameKey = strings.TrimSpace(menu.NameKey)
	menu.Path = strings.TrimSpace(menu.Path)
	menu.Permission = strings.TrimSpace(menu.Permission)

	if err := s.repo.CreateMenu(ctx, menu); err != nil {
		switch {
		case errors.Is(err, repository.ErrMenuDuplicate):
			return nil, i18n.NewI18nError("600007", lang)
		case errors.Is(err, repository.ErrMenuParentNotFound):
			return nil, i18n.NewI18nError("600011", lang)
		case errors.Is(err, repository.ErrMenuCyclicReference):
			return nil, i18n.NewI18nError("600012", lang)
		default:
			return nil, i18n.NewI18nError("900005", lang)
		}
	}
	return menu, nil
}

// UpdateMenu 更新菜单
func (s *rbacService) UpdateMenu(ctx context.Context, id string, menu *models.Menu) (*models.Menu, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	existing, err := s.repo.GetMenuByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrMenuNotFound) {
			return nil, i18n.NewI18nError("600006", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	// 内置菜单归属系统，仅允许调整展示属性
	if existing.IsSystem {
		if menu.Type != "" && menu.Type != existing.Type {
			return nil, i18n.NewI18nError("600008", lang)
		}
		if menu.Path != "" && menu.Path != existing.Path {
			return nil, i18n.NewI18nError("600008", lang)
		}
		if menu.Permission != "" && menu.Permission != existing.Permission {
			return nil, i18n.NewI18nError("600008", lang)
		}
	}

	updated := *existing
	if menu.NameKey != "" {
		updated.NameKey = strings.TrimSpace(menu.NameKey)
	}
	if menu.Type != "" {
		updated.Type = menu.Type
	}
	if menu.Path != "" {
		updated.Path = strings.TrimSpace(menu.Path)
	}
	if menu.Icon != "" {
		updated.Icon = menu.Icon
	}
	if menu.ParentID != nil {
		updated.ParentID = menu.ParentID
	}
	if menu.Permission != "" {
		updated.Permission = strings.TrimSpace(menu.Permission)
	}
	updated.SortOrder = menu.SortOrder
	updated.Hidden = menu.Hidden
	if menu.Status != "" {
		updated.Status = menu.Status
	}
	updated.Description = strings.TrimSpace(menu.Description)
	updated.ID = id

	if err := s.repo.UpdateMenu(ctx, &updated); err != nil {
		switch {
		case errors.Is(err, repository.ErrMenuNotFound):
			return nil, i18n.NewI18nError("600006", lang)
		case errors.Is(err, repository.ErrMenuParentNotFound):
			return nil, i18n.NewI18nError("600011", lang)
		case errors.Is(err, repository.ErrMenuCyclicReference):
			return nil, i18n.NewI18nError("600012", lang)
		default:
			return nil, i18n.NewI18nError("900005", lang)
		}
	}
	return &updated, nil
}

// DeleteMenu 删除菜单
func (s *rbacService) DeleteMenu(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.repo.DeleteMenu(ctx, id); err != nil {
		switch {
		case errors.Is(err, repository.ErrMenuNotFound):
			return i18n.NewI18nError("600006", lang)
		case errors.Is(err, repository.ErrMenuSystemProtected):
			return i18n.NewI18nError("600008", lang)
		case errors.Is(err, repository.ErrMenuHasChildren):
			return i18n.NewI18nError("600009", lang)
		case errors.Is(err, repository.ErrMenuInUse):
			return i18n.NewI18nError("600010", lang)
		default:
			return i18n.NewI18nError("900005", lang)
		}
	}
	return nil
}

// validatePermissions 校验权限码均为已定义的资源动作组合
func (s *rbacService) validatePermissions(ctx context.Context, permissions []string, lang string) error {
	if len(permissions) == 0 {
		return nil
	}

	resources, err := s.repo.GetPermissionMatrix(ctx)
	if err != nil {
		return i18n.NewI18nError("900005", lang)
	}

	// 允许通配权限（超级管理员）
	allowed := map[string]bool{"*": true}
	for _, resource := range resources {
		for _, action := range resource.Actions {
			allowed[resource.Resource+":"+action.Code] = true
		}
	}

	for _, permission := range permissions {
		if !allowed[permission] {
			return i18n.NewI18nError("600015", lang)
		}
	}
	return nil
}
