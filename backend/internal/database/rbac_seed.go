package database

import (
	"context"
	"errors"
	"log"
	"time"

	"license-manager/internal/models"

	"gorm.io/gorm"
)

// 社区版保留的菜单。与企业版的差异：不含设备台账、API Key、套餐等
// 本项目尚未实现的模块，避免侧边栏出现点不开的死链。
var seedMenus = []models.Menu{
	// 一级菜单
	{ID: "menu-dashboard", NameKey: "menu.dashboard", Type: models.MenuTypeMenu, Path: "/dashboard", Icon: "DataLine", Permission: "dashboard:read", SortOrder: 1},
	{ID: "menu-license", NameKey: "menu.license", Type: models.MenuTypeMenu, Path: "/licenses", Icon: "Lock", Permission: "license:read", SortOrder: 2},
	{ID: "menu-customer", NameKey: "menu.customer", Type: models.MenuTypeMenu, Path: "/customers", Icon: "User", Permission: "customer:read", SortOrder: 3},

	// 产品管理分组
	{ID: "menu-product-management", NameKey: "menu.product_management", Type: models.MenuTypeDirectory, Path: "", Icon: "Box", Permission: "product_offering", SortOrder: 10, IsSystem: true},
	{ID: "menu-feature-template", NameKey: "menu.feature_template", Type: models.MenuTypeMenu, Path: "/feature-templates", Icon: "Files", Permission: "feature_template:read", SortOrder: 11, ParentID: strPtr("menu-product-management")},
	{ID: "menu-product-offering", NameKey: "menu.product_offering", Type: models.MenuTypeMenu, Path: "/products", Icon: "Goods", Permission: "product_offering:read", SortOrder: 12, ParentID: strPtr("menu-product-management")},

	// 管理分组
	{ID: "menu-manage", NameKey: "menu.manage", Type: models.MenuTypeDirectory, Path: "", Icon: "Setting", Permission: "role", SortOrder: 20, IsSystem: true},
	{ID: "menu-role", NameKey: "menu.role", Type: models.MenuTypeMenu, Path: "/roles", Icon: "UserFilled", Permission: "role:read", SortOrder: 21, ParentID: strPtr("menu-manage")},
	{ID: "menu-user", NameKey: "menu.user", Type: models.MenuTypeMenu, Path: "/users", Icon: "User", Permission: "user:read", SortOrder: 22, ParentID: strPtr("menu-manage")},
	{ID: "menu-menu", NameKey: "menu.menu", Type: models.MenuTypeMenu, Path: "/menus", Icon: "Menu", Permission: "menu:read", SortOrder: 23, ParentID: strPtr("menu-manage")},
	{ID: "menu-audit-log", NameKey: "menu.audit_log", Type: models.MenuTypeMenu, Path: "/audit-logs", Icon: "Document", Permission: "audit_log:read", SortOrder: 24, ParentID: strPtr("menu-manage")},
}

// seedPermissionResources 权限资源，对齐企业版实测的 16 项中社区版支持的部分
var seedPermissionResources = []models.PermissionResource{
	{ID: "dashboard", Resource: "dashboard", NameKey: "permission.dashboard", Name: "仪表盘", SortOrder: 1},
	{ID: "license", Resource: "license", NameKey: "permission.license", Name: "许可证管理", SortOrder: 2},
	{ID: "authorization_code", Resource: "authorization_code", NameKey: "permission.authorization_code", Name: "授权码管理", SortOrder: 3},
	{ID: "customer", Resource: "customer", NameKey: "permission.customer", Name: "客户管理", SortOrder: 4},
	{ID: "product_offering", Resource: "product_offering", NameKey: "permission.product_offering", Name: "产品项管理", SortOrder: 5},
	{ID: "feature_template", Resource: "feature_template", NameKey: "permission.feature_template", Name: "功能模板", SortOrder: 6},
	{ID: "role", Resource: "role", NameKey: "permission.role", Name: "角色管理", SortOrder: 7},
	{ID: "user", Resource: "user", NameKey: "permission.user", Name: "用户管理", SortOrder: 8},
	{ID: "menu", Resource: "menu", NameKey: "permission.menu", Name: "菜单管理", SortOrder: 9},
	{ID: "audit_log", Resource: "audit_log", NameKey: "permission.audit_log", Name: "审计日志", SortOrder: 10},
}

// seedRoles 内置角色
var seedRoles = []struct {
	Role        models.Role
	Permissions []string
	MenuIDs     []string
}{
	{
		Role: models.Role{
			Name: "super_admin", DisplayName: "超级管理员",
			Description: "系统内置角色: 超级管理员", IsSystem: true,
		},
		// 超级管理员拥有全部权限，此处仅标记通配
		Permissions: []string{"*"},
		MenuIDs: []string{
			"menu-dashboard", "menu-license", "menu-customer",
			"menu-product-management", "menu-feature-template", "menu-product-offering",
			"menu-manage", "menu-role", "menu-user", "menu-menu", "menu-audit-log",
		},
	},
	{
		Role: models.Role{
			Name: "operator", DisplayName: "运营人员",
			Description: "系统内置角色: 运营人员", IsSystem: true,
		},
		Permissions: []string{
			"dashboard:read",
			"license:read", "authorization_code:read", "authorization_code:create",
			"customer:read", "customer:create", "customer:update",
			"audit_log:read",
		},
		MenuIDs: []string{
			"menu-dashboard", "menu-license", "menu-customer", "menu-manage", "menu-audit-log",
		},
	},
	{
		Role: models.Role{
			Name: "viewer", DisplayName: "只读用户",
			Description: "系统内置角色: 只读用户", IsSystem: true,
		},
		Permissions: []string{
			"dashboard:read", "license:read", "authorization_code:read", "customer:read",
		},
		MenuIDs: []string{"menu-dashboard", "menu-license", "menu-customer"},
	},
}

// initRBACData 初始化 RBAC 基础数据（幂等）
func initRBACData() error {
	ctx := context.Background()
	db := DB
	now := time.Now()

	// 菜单
	for i := range seedMenus {
		menu := seedMenus[i]
		var existing models.Menu
		err := db.WithContext(ctx).Where("id = ?", menu.ID).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := db.WithContext(ctx).Create(&menu).Error; err != nil {
			return err
		}
	}

	// 权限资源
	for i := range seedPermissionResources {
		resource := seedPermissionResources[i]
		// 先按主键查，再按业务键查：上次迁移失败可能留下部分行，
		// FirstOrCreate 只看主键，遇到残留的唯一键冲突会报错。
		var existing models.PermissionResource
		err := db.WithContext(ctx).Where("id = ? OR resource = ?", resource.ID, resource.Resource).First(&existing).Error
		switch {
		case err == nil:
			continue // 已存在
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}
		if err := db.WithContext(ctx).Create(&resource).Error; err != nil {
			return err
		}
	}

	// 角色及其权限、菜单关联
	for _, seed := range seedRoles {
		role := seed.Role
		var existing models.Role
		err := db.WithContext(ctx).Where("name = ?", role.Name).First(&existing).Error
		switch {
		case err == nil:
			role = existing
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		default:
			if err := db.WithContext(ctx).Create(&role).Error; err != nil {
				return err
			}
		}

		// 关联表用 FirstOrCreate 的批量替代：逐条判断存在再插入
		for _, permission := range seed.Permissions {
			var rp models.RolePermission
			err := db.WithContext(ctx).
				Where("role_id = ? AND permission = ?", role.ID, permission).
				First(&rp).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rp = models.RolePermission{RoleID: role.ID, Permission: permission, CreatedAt: now}
				if err := db.WithContext(ctx).Create(&rp).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		for _, menuID := range seed.MenuIDs {
			var rm models.RoleMenu
			err := db.WithContext(ctx).
				Where("role_id = ? AND menu_id = ?", role.ID, menuID).
				First(&rm).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rm = models.RoleMenu{RoleID: role.ID, MenuID: menuID, CreatedAt: now}
				if err := db.WithContext(ctx).Create(&rm).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
	}

	// 既有用户按原 role 字段绑定对应角色，保留原有行为
	return backfillUserRoles(ctx, db)
}

// backfillUserRoles 将 users.role 字段映射为 user_roles 关联。
//
// 社区版原本只有 admin/operator/viewer 三个硬编码角色，
// 迁移后保持语义不变：admin 拥有全部权限。
func backfillUserRoles(ctx context.Context, db *gorm.DB) error {
	var users []struct {
		ID   string
		Role string
	}
	if err := db.WithContext(ctx).Model(&models.User{}).
		Select("id", "role").Find(&users).Error; err != nil {
		return err
	}

	roleNameToID := map[string]string{}
	var roles []models.Role
	if err := db.WithContext(ctx).Find(&roles).Error; err != nil {
		return err
	}
	for _, r := range roles {
		roleNameToID[r.Name] = r.ID
	}

	now := time.Now()
	for _, user := range users {
		// admin 直接视为超级管理员
		roleName := user.Role
		if roleName == "admin" {
			roleName = "super_admin"
		}

		roleID, ok := roleNameToID[roleName]
		if !ok {
			roleID = roleNameToID["viewer"]
		}
		if roleID == "" {
			continue
		}

		var ur models.UserRole
		err := db.WithContext(ctx).
			Where("user_id = ? AND role_id = ?", user.ID, roleID).
			First(&ur).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ur = models.UserRole{UserID: user.ID, RoleID: roleID, CreatedAt: now}
			if err := db.WithContext(ctx).Create(&ur).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	log.Println("RBAC seed data initialized")
	return nil
}

// strPtr 返回字符串指针
func strPtr(s string) *string {
	return &s
}
