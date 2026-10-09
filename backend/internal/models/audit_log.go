package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 审计日志状态
const (
	AuditStatusSuccess = "success"
	AuditStatusFailure = "failure"
)

// 审计资源类型
const (
	AuditResourceAuthorizationCode = "authorization-code"
	AuditResourceLicense           = "license"
	AuditResourceCustomer          = "customer"
	AuditResourceProductOffering   = "product-offering"
	AuditResourceFeatureTemplate   = "feature-template"
	AuditResourceUser              = "user"
)

// AuditLog 审计日志模型
//
// 记录写操作（POST/PUT/PATCH/DELETE）的可追溯信息：谁、在什么时间、
// 对什么资源、做了什么、结果如何。读操作不记录，否则日志量会淹没真正
// 需要追溯的变更。
type AuditLog struct {
	ID       string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID   string `gorm:"type:varchar(36);index" json:"user_id"`
	Username string `gorm:"type:varchar(100)" json:"username"`

	// Action 形如 "POST /v1/authorization-codes"，由方法与路径组成
	Action       string `gorm:"type:varchar(255);index" json:"action"`
	ResourceType string `gorm:"type:varchar(64);index" json:"resource_type"`
	ResourceID   string `gorm:"type:varchar(36);index" json:"resource_id"`

	// ProductID 关联产品项，便于按产品追溯（授权类操作会填充）
	ProductID *string `gorm:"type:varchar(36);index" json:"product_id"`

	// Detail 存放请求/响应摘要。请求体与响应体以字符串保存原始内容，
	// 不解析为对象：审计需要保留原文，解析后可能丢失未知字段。
	Detail JSON `gorm:"type:json" json:"detail" swaggertype:"object"`

	IP        string `gorm:"type:varchar(64)" json:"ip"`
	UserAgent string `gorm:"type:varchar(512)" json:"user_agent"`

	Status string `gorm:"type:varchar(20);not null;index" json:"status"`

	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate 创建前自动设置时间戳和ID
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	if a.Status == "" {
		a.Status = AuditStatusSuccess
	}
	return nil
}

// AuditLogDetail 审计详情内容
type AuditLogDetail struct {
	Path     string `json:"path,omitempty"`
	Method   string `json:"method,omitempty"`
	ReqBody  string `json:"req_body,omitempty"`
	RespBody string `json:"resp_body,omitempty"`
	RespID   string `json:"resp_id,omitempty"`
}

// AuditLogListRequest 审计日志列表查询请求
type AuditLogListRequest struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`                   // 页码
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`      // 每页条数
	ResourceType string `form:"resource_type" binding:"omitempty"`                // 资源类型
	UserID       string `form:"user_id" binding:"omitempty"`                      // 操作人
	Status       string `form:"status" binding:"omitempty,oneof=success failure"` // 状态
	Action       string `form:"action" binding:"omitempty"`                       // 动作模糊匹配
	ProductID    string `form:"product_id" binding:"omitempty"`                   // 产品项
	StartDate    string `form:"start_date" binding:"omitempty"`                   // 开始时间
	EndDate      string `form:"end_date" binding:"omitempty"`                     // 结束时间
	SortBy       string `form:"sort_by" binding:"omitempty"`                      // 排序字段
	SortOrder    string `form:"sort_order" binding:"omitempty,oneof=asc desc"`    // 排序方向
}

// AuditLogListResponse 审计日志列表响应
type AuditLogListResponse struct {
	List       []AuditLog `json:"list"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// AuditLogStatisticsResponse 审计统计响应
type AuditLogStatisticsResponse struct {
	Total      int64            `json:"total"`
	Success    int64            `json:"success"`
	Failure    int64            `json:"failure"`
	ByResource map[string]int64 `json:"by_resource"`
	ByStatus   map[string]int64 `json:"by_status"`
}
