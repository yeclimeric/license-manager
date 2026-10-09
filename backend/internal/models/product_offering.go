package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductOffering 产品项模型
//
// 产品项是授权策略的基础载体：定义"授权什么"。授权码在其之上定义"怎么授权"，
// 许可证是设备级授权的最终产物。
type ProductOffering struct {
	ID          string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(200);not null;index" json:"name"`
	Code        string `gorm:"type:varchar(64);uniqueIndex;not null" json:"code"`
	Alias       string `gorm:"type:varchar(200)" json:"alias,omitempty"`
	Description string `gorm:"type:text" json:"description"`

	// 体验（试用）配置
	IsTrialProduct bool `gorm:"not null;default:false" json:"is_trial_product"`
	TrialDays      *int `gorm:"type:int" json:"trial_days,omitempty"`

	// 默认授权配置，新建授权码时继承
	FeatureConfig    JSON `gorm:"type:json" json:"feature_config" swaggertype:"object"`
	UsageLimits      JSON `gorm:"type:json" json:"usage_limits" swaggertype:"object"`
	CustomParameters JSON `gorm:"type:json" json:"custom_parameters" swaggertype:"object"`

	// 默认授权策略
	MaxActivations *int   `gorm:"type:int" json:"max_activations,omitempty"`
	DurationDays   *int   `gorm:"type:int" json:"duration_days,omitempty"`
	DeploymentMode string `gorm:"type:varchar(20);not null;default:'online'" json:"deployment_mode"`

	Status        string `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	StatusDisplay string `gorm:"-" json:"status_display,omitempty"`

	CreatedAt time.Time      `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联与统计
	AuthorizationCodes []AuthorizationCode `gorm:"foreignKey:ProductOfferingID" json:"-"`
	AuthCodeCount      int                 `gorm:"-" json:"auth_code_count"`
	UserCount          int                 `gorm:"-" json:"user_count"`

	CanEdit   bool `gorm:"-" json:"can_edit"`
	CanDelete bool `gorm:"-" json:"can_delete"`
}

// TableName 指定表名
func (ProductOffering) TableName() string {
	return "product_offerings"
}

// BeforeCreate 创建前自动设置时间戳和ID
func (p *ProductOffering) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	return nil
}

// DisplayName 返回界面展示名称：优先别名，为空回落编码
func (p *ProductOffering) DisplayName() string {
	if p.Alias != "" {
		return p.Alias
	}
	return p.Code
}

// ProductOfferingListRequest 产品项列表查询请求
type ProductOfferingListRequest struct {
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
	Keyword   string `form:"keyword" json:"keyword"`
	Status    string `form:"status" json:"status"`
	SortBy    string `form:"sort_by" json:"sort_by"`
	SortOrder string `form:"sort_order" json:"sort_order"`
}

// ProductOfferingCreateRequest 新建产品项请求
type ProductOfferingCreateRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Alias            string                 `json:"alias"`
	Description      string                 `json:"description"`
	IsTrialProduct   bool                   `json:"is_trial_product"`
	TrialDays        *int                   `json:"trial_days"`
	FeatureConfig    map[string]interface{} `json:"feature_config"`
	UsageLimits      map[string]interface{} `json:"usage_limits"`
	CustomParameters map[string]interface{} `json:"custom_parameters"`
	MaxActivations   *int                   `json:"max_activations"`
	DurationDays     *int                   `json:"duration_days"`
	DeploymentMode   string                 `json:"deployment_mode"`
}

// ProductOfferingUpdateRequest 更新产品项请求
//
// Code 不在更新范围内：编码由系统生成且不可更改（D4 决策）。
type ProductOfferingUpdateRequest struct {
	Name             *string                `json:"name"`
	Alias            *string                `json:"alias"`
	Description      *string                `json:"description"`
	IsTrialProduct   *bool                  `json:"is_trial_product"`
	TrialDays        *int                   `json:"trial_days"`
	FeatureConfig    map[string]interface{} `json:"feature_config"`
	UsageLimits      map[string]interface{} `json:"usage_limits"`
	CustomParameters map[string]interface{} `json:"custom_parameters"`
	MaxActivations   *int                   `json:"max_activations"`
	DurationDays     *int                   `json:"duration_days"`
	DeploymentMode   *string                `json:"deployment_mode"`
	Status           *string                `json:"status"`
}

// ProductOfferingResponse 产品项响应
type ProductOfferingResponse struct {
	*ProductOffering
	TrialDisplay string `json:"trial_display,omitempty"`
}

// ListResponse 分页响应
type ProductOfferingListResponse struct {
	List       []ProductOfferingResponse `json:"list"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}
