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

	// 签名密钥。创建产品项时自动生成：客户端内嵌公钥验签，
	// 服务端持私钥签发许可证。公钥随产品项一并导出，私钥永不出接口。
	SigningPrivateKey string `gorm:"type:text" json:"-"`
	SigningPublicKey  string `gorm:"type:text" json:"-"`

	// SigningKeyChain 签名密钥链（按时间正序，最后一条为当前签发密钥）。
	//
	// 密钥轮换后旧公钥不能丢弃：已发放的许可证是旧私钥签的，客户端仍需
	// 旧公钥才能验签通过。链上每条记录携带「前任私钥对本条公钥的签名」，
	// 客户端用已内置的根公钥逐级验证，即可把整条链纳入信任——
	// 服务端轮换密钥时，离线客户端无需重新发版。
	SigningKeyChain JSON `gorm:"type:json" json:"-" swaggertype:"object"`

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
	// HasSigningKey 标识该产品项是否已生成签名密钥
	HasSigningKey bool `json:"has_signing_key"`
}

// ProductOfferingPublicKeyResponse 产品公钥响应
//
// 与企业版实测契约一致：只返回产品标识与公钥本身。
type ProductOfferingPublicKeyResponse struct {
	ProductID   string `json:"product_id"`   // 产品项ID
	ProductCode string `json:"product_code"` // 产品项编码
	PublicKey   string `json:"public_key"`   // PEM 格式公钥
	// PublicKeys 全部有效公钥（当前 + 轮换留档）。
	//
	// 客户端可一次性获取全部公钥并内置，密钥轮换后无需重新发版客户端——
	// 只要新公钥随激活/心跳响应下发即可。顺序：当前公钥在前。
	PublicKeys []string `json:"public_keys,omitempty"`
	// KeyChain 完整签名密钥链（含每级的衔接签名）。
	//
	// 客户端以内置的根公钥为锚逐级验证 Signature，即可把整条链纳入信任，
	// 从而在离线场景下也支持密钥轮换。
	KeyChain []SigningKeyChainEntry `json:"key_chain,omitempty"`
}

// SigningKeyChainEntry 签名密钥链条目
//
// Signature 为「前任私钥对本条公钥 PEM 的签名」，算法固定 RSA-PSS-SHA256。
// 客户端以已内置的根公钥为锚，逐级验证 Signature 从而信任整条链。
type SigningKeyChainEntry struct {
	PublicKey string     `json:"public_key"`          // 本条公钥 PEM
	Signature string     `json:"signature,omitempty"` // 前任私钥对本公钥的签名（base64）
	RetiredAt *time.Time `json:"retired_at,omitempty"` // 退役时间；当前签发密钥为 nil
	Reason    string     `json:"reason,omitempty"`     // 轮换原因
}

// ProductOfferingRotateKeyRequest 轮换产品项签名密钥请求
type ProductOfferingRotateKeyRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=200"` // 轮换原因（留档用）
}

// ListResponse 分页响应
type ProductOfferingListResponse struct {
	List       []ProductOfferingResponse `json:"list"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}
