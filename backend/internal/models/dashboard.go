package models

import "time"

// DashboardAuthorizationTrendRequest 授权趋势查询请求
type DashboardAuthorizationTrendRequest struct {
	Type      string `form:"type" binding:"required,oneof=week month custom" json:"type"` // 时间类型: week/month/custom
	StartDate string `form:"start_date" json:"start_date,omitempty"`                      // 开始日期 (YYYY-MM-DD格式，当type为custom时必填)
	EndDate   string `form:"end_date" json:"end_date,omitempty"`                          // 结束日期 (YYYY-MM-DD格式，当type为custom时必填)
	Timezone  string `form:"timezone" json:"timezone,omitempty"`                          // 时区 (如: Asia/Shanghai, UTC等，默认使用服务器本地时区)
}

// DashboardAuthorizationTrendResponse 授权趋势响应
type DashboardAuthorizationTrendResponse struct {
	Period    TrendPeriod  `json:"period"`     // 时间段信息
	TrendData []TrendData  `json:"trend_data"` // 趋势数据
	Summary   TrendSummary `json:"summary"`    // 汇总信息
}

// TrendPeriod 时间段信息
type TrendPeriod struct {
	Type               string `json:"type"`                // 时间类型
	StartDate          string `json:"start_date"`          // 开始日期
	EndDate            string `json:"end_date"`            // 结束日期
	DescriptionDisplay string `json:"description_display"` // 多语言描述
}

// TrendData 单日趋势数据
type TrendData struct {
	Date                  string `json:"date"`                   // 日期 (YYYY-MM-DD)
	TotalAuthorizations   int64  `json:"total_authorizations"`   // 当日授权总数
	NewAuthorizations     int64  `json:"new_authorizations"`     // 当日新增授权数
	ExpiredAuthorizations int64  `json:"expired_authorizations"` // 当日过期授权数
}

// TrendSummary 趋势汇总信息
type TrendSummary struct {
	TotalCount   int64   `json:"total_count"`   // 期间总授权数
	NewCount     int64   `json:"new_count"`     // 期间新增授权数
	ExpiredCount int64   `json:"expired_count"` // 期间过期授权数
	GrowthRate   float64 `json:"growth_rate"`   // 增长率(百分比)
}

// DashboardRecentAuthorizationsRequest 最近授权列表查询请求
type DashboardRecentAuthorizationsRequest struct {
	Limit      int    `form:"limit" json:"limit,omitempty"`             // 返回数量限制（默认20，最大100）
	CustomerID string `form:"customer_id" json:"customer_id,omitempty"` // 客户ID筛选（可选）
	Status     string `form:"status" json:"status,omitempty"`           // 状态筛选 (normal/locked/expired) 可选
}

// DashboardRecentAuthorizationsResponse 最近授权列表响应
type DashboardRecentAuthorizationsResponse struct {
	List  []RecentAuthorization `json:"list"`  // 授权列表
	Total int64                 `json:"total"` // 符合条件的授权总数量
}

// RecentAuthorization 最近授权信息
type RecentAuthorization struct {
	ID                 string    `json:"id"`                  // 授权码ID
	Code               string    `json:"code"`                // 授权码
	CustomerID         *string   `json:"customer_id"`         // 客户ID，NULL表示无客户
	CustomerName       string    `json:"customer_name"`       // 客户名称
	Description        string    `json:"description"`         // 授权描述
	Status             string    `json:"status"`              // 状态 (normal/locked/expired)
	StatusDisplay      string    `json:"status_display"`      // 状态显示文本(多语言)
	StartDate          time.Time `json:"start_date"`          // 生效时间
	EndDate            time.Time `json:"end_date"`            // 到期时间
	MaxActivations     int       `json:"max_activations"`     // 最大激活数量
	CurrentActivations int       `json:"current_activations"` // 当前激活数量
	// ValidityMode 有效期模式：fixed(固定期限)/activation_based(从激活起算)
	// 前端据此提示：起算模式下到期日因设备而异，不能直接展示授权码的 end_date
	ValidityMode string    `json:"validity_mode"`
	ExpiresAt    time.Time `json:"expires_at"` // 到期时间（从激活起算时按最近激活推算）
	// ProductOfferingName 关联产品项名称，未关联时为空
	ProductOfferingName string    `json:"product_offering_name"`
	CreatedAt           time.Time `json:"created_at"` // 创建时间
	UpdatedAt           time.Time `json:"updated_at"` // 更新时间
}

// DashboardHomeResponse 仪表盘首页聚合数据。
type DashboardHomeResponse struct {
	GeneratedAt          time.Time                   `json:"generated_at"`
	Overview             DashboardOverview           `json:"overview"`
	ExpiryReminders      DashboardExpiryReminders    `json:"expiry_reminders"`
	RecentAuthorizations []RecentAuthorization       `json:"recent_authorizations"`
	RecentActivations    []DashboardRecentActivation `json:"recent_activations"`
}

// DashboardOverview 首页核心业务概览。
type DashboardOverview struct {
	ValidAuthorizations      DashboardValidAuthorizations `json:"valid_authorizations"`
	ActivatedDevices         DashboardActivatedDevices    `json:"activated_devices"`
	RemainingActivationSlots int64                        `json:"remaining_activation_slots"`
}

type DashboardValidAuthorizations struct {
	Total        int64 `json:"total"`
	Activated    int64 `json:"activated"`
	NotActivated int64 `json:"not_activated"`
}

// DashboardActivatedDevices 已激活设备。
//
// 拆分手工激活与体验激活：体验型产品不占用客户的激活名额，
// 运营需要分开看才能判断真实用量。
type DashboardActivatedDevices struct {
	Total int64 `json:"total"`
	// Manual 手工授权码激活的设备
	Manual int64 `json:"manual"`
	// Trial 体验型产品激活的设备
	Trial int64 `json:"trial"`
}

type DashboardExpiryReminders struct {
	Due7Days     DashboardExpiryReminder `json:"due_7_days"`
	Due8To30Days DashboardExpiryReminder `json:"due_8_to_30_days"`
	Expired      DashboardExpiryReminder `json:"expired"`
}

type DashboardExpiryReminder struct {
	AuthorizationCount  int64 `json:"authorization_count"`
	AffectedDeviceCount int64 `json:"affected_device_count"`
}

// DashboardRecentActivation 最近完成的设备激活。
type DashboardRecentActivation struct {
	ID                  string    `json:"id"`
	AuthorizationCodeID string    `json:"authorization_code_id"`
	CustomerName        string    `json:"customer_name"`
	Description         string    `json:"description"`
	HardwareFingerprint string    `json:"hardware_fingerprint"`
	ActivatedAt         time.Time `json:"activated_at"`
	EndDate             time.Time `json:"end_date"`
}

// DashboardBusinessTrendsRequest 双业务趋势查询请求。
type DashboardBusinessTrendsRequest struct {
	Period    string `form:"period" json:"period"`
	StartDate string `form:"start_date" json:"start_date,omitempty"`
	EndDate   string `form:"end_date" json:"end_date,omitempty"`
	Timezone  string `form:"timezone" json:"timezone,omitempty"`
}

type DashboardBusinessTrendsResponse struct {
	Period                     DashboardBusinessTrendPeriod `json:"period"`
	AuthorizationCreationTrend []DashboardTrendPoint        `json:"authorization_creation_trend"`
	ActivationTrend            []DashboardTrendPoint        `json:"activation_trend"`
}

type DashboardBusinessTrendPeriod struct {
	Type      string `json:"type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type DashboardTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
