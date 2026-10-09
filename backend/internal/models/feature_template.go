package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FeatureTemplate 功能模板模型
//
// 模板是一组功能开关的集合，用于集中维护可复用的功能配置。创建产品项时
// 可选择模板，许可证激活时关联模板，决定客户端能使用哪些功能。
type FeatureTemplate struct {
	ID          string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(200);not null;index" json:"name"`
	Identifier  string `gorm:"type:varchar(100);uniqueIndex;not null" json:"identifier"`
	Description string `gorm:"type:text" json:"description"`

	// 功能分类，对应 feature_template_categories.value
	Category        string `gorm:"type:varchar(50);not null;default:'general';index" json:"category"`
	CategoryDisplay string `gorm:"-" json:"category_display,omitempty"`

	// 功能项列表。与产品项的 feature_config（对象）不同，
	// 模板需要保留功能项的顺序与描述，因此用数组。
	FeatureConfig    JSON `gorm:"type:json" json:"feature_config" swaggertype:"array"`
	UsageLimits      JSON `gorm:"type:json" json:"usage_limits" swaggertype:"object"`
	CustomParameters JSON `gorm:"type:json" json:"custom_parameters" swaggertype:"object"`

	SortOrder int    `gorm:"not null;default:0" json:"sort_order"`
	Status    string `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`

	// StatusDisplay 多语言状态显示
	StatusDisplay string `gorm:"-" json:"status_display,omitempty"`
	// IsSystem 内置模板不可删除
	IsSystem bool `gorm:"not null;default:false" json:"is_system"`

	CreatedAt time.Time      `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	CanDelete bool `gorm:"-" json:"can_delete"`
}

// FeatureTemplateCategory 功能模板分类模型
type FeatureTemplateCategory struct {
	ID        string    `gorm:"type:varchar(64);primaryKey" json:"id"`
	Value     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"value"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Label     string    `gorm:"type:varchar(100);not null" json:"label"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	IsSystem  bool      `gorm:"not null;default:false" json:"is_system"`
	CreatedBy string    `gorm:"type:varchar(36);not null" json:"created_by"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	TemplateCount int  `gorm:"-" json:"template_count"`
	CanDelete     bool `gorm:"-" json:"can_delete"`
}

// TableName 指定表名
func (FeatureTemplate) TableName() string {
	return "feature_templates"
}

// BeforeCreate 创建前自动设置时间戳和ID
func (f *FeatureTemplate) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}

	now := time.Now()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	if f.UpdatedAt.IsZero() {
		f.UpdatedAt = now
	}

	return nil
}

// TableName 指定表名
func (FeatureTemplateCategory) TableName() string {
	return "feature_template_categories"
}

// BeforeCreate 创建前自动设置时间戳和ID
func (c *FeatureTemplateCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = "ftc_" + uuid.New().String()[:8]
	}

	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}

	return nil
}

// 内置分类的 value 与 ID，对齐企业版实测数据
const (
	CategoryGeneral       = "general"
	CategorySensor        = "sensor"
	CategoryCommunication = "communication"
	CategoryOther         = "other"
)

// FeatureTemplateListRequest 功能模板列表查询请求
type FeatureTemplateListRequest struct {
	Page      int    `form:"page" json:"page"`
	PageSize  int    `form:"page_size" json:"page_size"`
	Keyword   string `form:"keyword" json:"keyword"`
	Category  string `form:"category" json:"category"`
	Status    string `form:"status" json:"status"`
	SortBy    string `form:"sort_by" json:"sort_by"`
	SortOrder string `form:"sort_order" json:"sort_order"`
}

// FeatureTemplateCreateRequest 创建功能模板请求
type FeatureTemplateCreateRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Identifier       string                 `json:"identifier"`
	Description      string                 `json:"description"`
	Category         string                 `json:"category"`
	FeatureConfig    []interface{}          `json:"feature_config"`
	UsageLimits      map[string]interface{} `json:"usage_limits"`
	CustomParameters map[string]interface{} `json:"custom_parameters"`
	SortOrder        *int                   `json:"sort_order"`
}

// FeatureTemplateUpdateRequest 更新功能模板请求
//
// Identifier 不在更新范围内：模板标识符被客户端 SDK 引用，创建后不可更改。
type FeatureTemplateUpdateRequest struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	Category         *string                `json:"category"`
	FeatureConfig    []interface{}          `json:"feature_config"`
	UsageLimits      map[string]interface{} `json:"usage_limits"`
	CustomParameters map[string]interface{} `json:"custom_parameters"`
	SortOrder        *int                   `json:"sort_order"`
	Status           *string                `json:"status"`
}

// FeatureTemplateListResponse 功能模板列表响应
type FeatureTemplateListResponse struct {
	List       []FeatureTemplate `json:"list"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// FeatureTemplateCategoryListResponse 分类列表响应
type FeatureTemplateCategoryListResponse struct {
	List  []FeatureTemplateCategory `json:"list"`
	Total int                       `json:"total"`
}
