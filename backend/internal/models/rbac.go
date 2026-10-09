package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 菜单类型
const (
	// MenuTypeDirectory 分组目录，仅用于组织子菜单，不对应页面
	MenuTypeDirectory = "directory"
	// MenuTypeMenu 实际菜单项，对应一个路由
	MenuTypeMenu = "menu"
)

// Menu 菜单模型
//
// 菜单与角色是多对多关系：角色关联菜单后，用户获得该角色即可看到对应菜单。
// NameKey 存 i18n 键（如 menu.product_management）而非翻译文本，
// 这样切换语言时菜单标签随之变化，无需改动数据库。
type Menu struct {
	ID       string  `gorm:"type:varchar(36);primaryKey" json:"id"`
	NameKey  string  `gorm:"type:varchar(100);not null" json:"name_key"`
	Type     string  `gorm:"type:varchar(20);not null;default:'menu';index" json:"type"`
	Path     string  `gorm:"type:varchar(200);not null;default:''" json:"path"`
	Icon     string  `gorm:"type:varchar(64)" json:"icon"`
	ParentID *string `gorm:"type:varchar(36);index" json:"parent_id"`

	// Permission 该菜单所需权限码（如 product_offering:read）。
	// 目录节点填资源名（如 product_offering），表示其下所有子菜单的权限前缀。
	Permission string `gorm:"type:varchar(100);index" json:"permission"`

	SortOrder   int    `gorm:"not null;default:0" json:"sort_order"`
	Status      string `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	IsSystem    bool   `gorm:"not null;default:false" json:"is_system"`
	Hidden      bool   `gorm:"not null;default:false" json:"hidden"`
	Description string `gorm:"type:varchar(255)" json:"description"`

	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Children []Menu `gorm:"-" json:"children"`

	CanDelete bool `gorm:"-" json:"can_delete"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}

// BeforeCreate 创建前自动设置时间戳和ID
func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	if m.Status == "" {
		m.Status = "active"
	}
	if m.Type == "" {
		m.Type = MenuTypeMenu
	}
	return nil
}

// PermissionResource 权限资源模型
//
// 描述"可以对什么做什么"，是权限矩阵的行。动作固定为
// read/create/update/delete，不存表而由代码定义。
type PermissionResource struct {
	ID        string `gorm:"type:varchar(64);primaryKey" json:"id"`
	Resource  string `gorm:"type:varchar(64);uniqueIndex;not null" json:"resource"`
	NameKey   string `gorm:"type:varchar(100);not null" json:"name_key"`
	Name      string `gorm:"type:varchar(100);not null" json:"name"`
	SortOrder int    `gorm:"not null;default:0" json:"sort_order"`

	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	Actions []PermissionAction `gorm:"-" json:"actions"`
}

// TableName 指定表名
func (PermissionResource) TableName() string {
	return "permission_resources"
}

// BeforeCreate 创建前设置时间戳。
//
// 主键取 Resource 本身（不使用 UUID），因此这里无需生成 ID。
func (p *PermissionResource) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	return nil
}

// PermissionAction 权限动作定义
type PermissionAction struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Role 角色模型
type Role struct {
	ID          string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"type:varchar(100);not null" json:"display_name"`
	Description string `gorm:"type:varchar(255)" json:"description"`

	// IsSystem 内置角色不可删除
	IsSystem bool   `gorm:"not null;default:false" json:"is_system"`
	Status   string `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`

	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Permissions []string `gorm:"-" json:"permissions,omitempty"`
	MenuIDs     []string `gorm:"-" json:"menu_ids,omitempty"`
	UserCount   int      `gorm:"-" json:"user_count"`
	CanDelete   bool     `gorm:"-" json:"can_delete"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// BeforeCreate 创建前自动设置时间戳和ID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = now
	}
	if r.Status == "" {
		r.Status = "active"
	}
	return nil
}

// RolePermission 角色-权限关联
type RolePermission struct {
	RoleID     string    `gorm:"type:varchar(36);primaryKey" json:"role_id"`
	Permission string    `gorm:"type:varchar(100);primaryKey" json:"permission"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// RoleMenu 角色-菜单关联
type RoleMenu struct {
	RoleID    string    `gorm:"type:varchar(36);primaryKey" json:"role_id"`
	MenuID    string    `gorm:"type:varchar(36);primaryKey" json:"menu_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return "role_menus"
}

// UserRole 用户-角色关联
type UserRole struct {
	UserID    string    `gorm:"type:varchar(36);primaryKey" json:"user_id"`
	RoleID    string    `gorm:"type:varchar(36);primaryKey" json:"role_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}

// 标准权限动作
const (
	ActionRead   = "read"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// StandardActions 标准动作定义，顺序即展示顺序
var StandardActions = []PermissionAction{
	{Code: ActionRead, Name: "查看"},
	{Code: ActionCreate, Name: "创建"},
	{Code: ActionUpdate, Name: "更新"},
	{Code: ActionDelete, Name: "删除"},
}

// MenuTreeResponse 菜单树响应
type MenuTreeResponse struct {
	List  []Menu `json:"list"`
	Total int    `json:"total"`
}

// PermissionMatrixResponse 权限矩阵响应
type PermissionMatrixResponse struct {
	Resources []PermissionResource `json:"resources"`
	Actions   []PermissionAction   `json:"actions"`
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Name        string   `json:"name" binding:"required"`
	DisplayName string   `json:"display_name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	MenuIDs     []string `json:"menu_ids"`
}

// RoleUpdateRequest 更新角色请求
//
// 不含 Name：角色标识被用户绑定关系引用，创建后不可更改。
type RoleUpdateRequest struct {
	DisplayName *string  `json:"display_name"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	Permissions []string `json:"permissions"`
	MenuIDs     []string `json:"menu_ids"`
}
