package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 系统配置值类型
const (
	ConfigTypeText     = "text"     // 单行文本
	ConfigTypeTextarea = "textarea" // 多行文本
	ConfigTypeSwitch   = "switch"   // 开关
	ConfigTypeImage    = "image"    // 图片（Logo / Favicon）
	ConfigTypeSelect   = "select"   // 下拉选择
	ConfigTypeNumber   = "number"   // 数值
)

// 系统配置分组
const (
	ConfigGroupSystem = "system"
	ConfigGroupAuth   = "auth"
	ConfigGroupMail   = "mail"
)

// SystemConfig 系统配置项。
//
// 键用点号分层的扁平串（system.site_name），前端与公开接口都直接透传，
// 无需在前端再维护一份映射表。
type SystemConfig struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	ConfigKey string `gorm:"type:varchar(100);uniqueIndex;not null" json:"config_key"`
	Value     string `gorm:"type:text" json:"value"`
	ValueType string `gorm:"type:varchar(20);not null;default:'text'" json:"value_type"`
	Group     string `gorm:"type:varchar(50);not null;default:'system';index" json:"group"`
	// Options 下拉候选项，JSON 数组字符串，仅 value_type=select 时有意义
	Options string `gorm:"type:text" json:"options"`
	// Public 是否可由未认证的公开接口读取。
	// Logo 与站点名属于登录页所需，其余配置不应外泄。
	Public      bool           `gorm:"not null;default:false" json:"public"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	SortOrder   int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// BeforeCreate 创建前生成ID
func (c *SystemConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// SystemConfigCreateRequest 创建配置请求
type SystemConfigCreateRequest struct {
	ConfigKey   string `json:"config_key" binding:"required,max=100"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type" binding:"omitempty,oneof=text textarea switch image select number"`
	Group       string `json:"group" binding:"omitempty,oneof=system auth mail"`
	Options     string `json:"options"`
	Public      *bool  `json:"public"`
	Description string `json:"description" binding:"omitempty,max=255"`
	SortOrder   int    `json:"sort_order"`
}

// SystemConfigUpdateRequest 更新配置请求
type SystemConfigUpdateRequest struct {
	Value       *string `json:"value"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int    `json:"sort_order"`
	// Public 控制是否可公开读取
	Public *bool `json:"public"`
}

// SystemConfigBatchUpdateRequest 批量更新配置
//
// 一次提交整份配置而非逐项调用：配置页同时编辑多个键，
// 逐项保存会产生中间态，且部分失败时状态难以回滚。
type SystemConfigBatchUpdateRequest struct {
	Items []SystemConfigBatchItem `json:"items" binding:"required,min=1,dive"`
}

// SystemConfigBatchItem 批量更新项
type SystemConfigBatchItem struct {
	ConfigKey string `json:"config_key" binding:"required"`
	Value     string `json:"value"`
}

// SystemConfigGroup 分组视图
type SystemConfigGroup struct {
	Group   string         `json:"group"`
	Configs []SystemConfig `json:"configs"`
}

// SystemConfigGroupResponse 分组响应
type SystemConfigGroupResponse struct {
	Groups []SystemConfigGroup `json:"groups"`
}

// PublicConfigResponse 公开配置响应
//
// 键值对形式，与企业版实测契约一致。
type PublicConfigResponse struct {
	Configs map[string]string `json:"configs"`
}

// API Key 状态
const (
	APIKeyStatusActive   = "active"
	APIKeyStatusInactive = "inactive"
	APIKeyStatusExpired  = "expired"
)

// APIKeyPrefix API Key 前缀
const APIKeyPrefix = "sk-"

// APIKey API 密钥。
//
// 只保存哈希与前缀：明文仅在创建响应中返回一次，
// 之后无法还原。泄露数据库不等于泄露全部密钥。
type APIKey struct {
	ID   string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);not null" json:"name"`
	// KeyHash 密钥哈希，认证时比对
	KeyHash string `gorm:"type:varchar(64);not null;index" json:"-"`
	// KeyPrefix 前缀（sk- 加 8 位十六进制），用于列表辨识
	KeyPrefix string `gorm:"column:api_key_prefix;type:varchar(20);not null;index" json:"api_key_prefix"`
	UserID    string `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Status    string `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	// ExpiresAt 为空表示永不过期
	ExpiresAt *time.Time `json:"expires_at"`
	// AllowedIPs IP 白名单，JSON 数组字符串，为空表示不限制。
	// 以逗号分隔便于直接匹配请求 IP。
	AllowedIPs string         `gorm:"type:text" json:"-"`
	LastUsedAt *time.Time     `json:"last_used_at"`
	CreatedBy  string         `gorm:"type:varchar(36)" json:"created_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// BeforeCreate 创建前生成ID
func (k *APIKey) BeforeCreate(tx *gorm.DB) error {
	if k.ID == "" {
		k.ID = uuid.New().String()
	}
	return nil
}

// APIKeyCreateRequest 创建 API Key 请求
type APIKeyCreateRequest struct {
	Name   string `json:"name" binding:"required,max=100"`
	UserID string `json:"user_id" binding:"required"`
	// ExpiresAt 为空表示永不过期
	ExpiresAt *time.Time `json:"expires_at"`
	// AllowedIPs IP 白名单，空表示不限制
	AllowedIPs []string `json:"allowed_ips"`
}

// APIKeyCreateResponse 创建响应。
//
// APIKey 字段仅在此返回一次，之后任何接口都不再返回明文。
type APIKeyCreateResponse struct {
	APIKey *APIKey `json:"api_key"`
	// Plaintext 明文密钥，仅创建时返回
	Plaintext string `json:"plaintext"`
}

// APIKeyListItem API Key 列表项
type APIKeyListItem struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"api_key_prefix"`
	UserID     string     `json:"user_id"`
	UserName   string     `json:"user_name"`
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at"`
	AllowedIPs []string   `json:"allowed_ips"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  string     `json:"created_by"`
}

// APIKeyListResponse 列表响应
type APIKeyListResponse struct {
	List  []APIKeyListItem `json:"list"`
	Total int64            `json:"total"`
}

// APIKeyUpdateRequest 更新 API Key 请求
type APIKeyUpdateRequest struct {
	Name       *string    `json:"name" binding:"omitempty,max=100"`
	Status     *string    `json:"status" binding:"omitempty,oneof=active inactive"`
	ExpiresAt  *time.Time `json:"expires_at"`
	AllowedIPs []string   `json:"allowed_ips"`
}

// APIKeyListRequest 列表查询请求
type APIKeyListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive expired"`
}
