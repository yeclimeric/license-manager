package service

import (
	"context"
	"license-manager/internal/models"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, req *models.LoginRequest, clientIP string) (*models.LoginData, error)
	RefreshToken(token string) (string, error)
	ValidateToken(token string) error
}

// SystemService 系统服务接口
type SystemService interface {
	GetHealthStatus() *models.HealthResponse
	GetSystemInfo() map[string]interface{}
}

// CustomerService 客户服务接口
type CustomerService interface {
	GetCustomerList(ctx context.Context, req *models.CustomerListRequest) (*models.CustomerListResponse, error)
	GetCustomer(ctx context.Context, id string) (*models.Customer, error)
	CreateCustomer(ctx context.Context, req *models.CustomerCreateRequest) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, id string, req *models.CustomerUpdateRequest) (*models.Customer, error)
	DeleteCustomer(ctx context.Context, id string) error
	UpdateCustomerStatus(ctx context.Context, id string, req *models.CustomerStatusUpdateRequest) (*models.Customer, error)
}

// EnumService 枚举服务接口
type EnumService interface {
	GetAllEnums(ctx context.Context) (*models.EnumListResponse, error)
	GetEnumsByType(ctx context.Context, enumType string) (*models.EnumTypeResponse, error)
}

// RBACService 菜单权限角色服务接口
type RBACService interface {
	// GetUserMenuTree 获取当前用户可见菜单树
	GetUserMenuTree(ctx context.Context) (*models.MenuTreeResponse, error)
	// GetAllMenuTree 获取全部菜单树
	GetAllMenuTree(ctx context.Context) (*models.MenuTreeResponse, error)
	// GetPermissionMatrix 获取权限矩阵
	GetPermissionMatrix(ctx context.Context) (*models.PermissionMatrixResponse, error)
	// GetRoleList 获取角色列表
	GetRoleList(ctx context.Context, keyword string) ([]models.Role, int, error)
	// GetRole 获取角色详情
	GetRole(ctx context.Context, id string) (*models.Role, error)
	// CreateRole 创建角色
	CreateRole(ctx context.Context, req *models.RoleCreateRequest) (*models.Role, error)
	// UpdateRole 更新角色
	UpdateRole(ctx context.Context, id string, req *models.RoleUpdateRequest) (*models.Role, error)
	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, id string) error
	// SetUserRoles 设置用户绑定的角色
	SetUserRoles(ctx context.Context, userID string, roleIDs []string) error
	// GetUserRoles 获取用户绑定的角色ID
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
	// CreateMenu 创建菜单
	CreateMenu(ctx context.Context, menu *models.Menu) (*models.Menu, error)
	// UpdateMenu 更新菜单
	UpdateMenu(ctx context.Context, id string, menu *models.Menu) (*models.Menu, error)
	// DeleteMenu 删除菜单
	DeleteMenu(ctx context.Context, id string) error
}

// AuditLogService 审计日志服务接口
type AuditLogService interface {
	// GetAuditLogList 查询审计日志列表
	GetAuditLogList(ctx context.Context, req *models.AuditLogListRequest) (*models.AuditLogListResponse, error)
	// GetStatistics 查询审计日志统计
	GetStatistics(ctx context.Context) (*models.AuditLogStatisticsResponse, error)
}

// FeatureTemplateService 功能模板服务接口
type FeatureTemplateService interface {
	// GetFeatureTemplateList 查询功能模板列表
	GetFeatureTemplateList(ctx context.Context, req *models.FeatureTemplateListRequest) (*models.FeatureTemplateListResponse, error)
	// GetFeatureTemplate 获取功能模板详情
	GetFeatureTemplate(ctx context.Context, id string) (*models.FeatureTemplate, error)
	// CreateFeatureTemplate 创建功能模板，标识符可自动生成
	CreateFeatureTemplate(ctx context.Context, req *models.FeatureTemplateCreateRequest) (*models.FeatureTemplate, error)
	// UpdateFeatureTemplate 更新功能模板，标识符不可更改
	UpdateFeatureTemplate(ctx context.Context, id string, req *models.FeatureTemplateUpdateRequest) (*models.FeatureTemplate, error)
	// DeleteFeatureTemplate 删除功能模板
	DeleteFeatureTemplate(ctx context.Context, id string) error
	// GetCategoryList 查询分类列表
	GetCategoryList(ctx context.Context) (*models.FeatureTemplateCategoryListResponse, error)
	// CreateCategory 新建自定义分类
	CreateCategory(ctx context.Context, category *models.FeatureTemplateCategory) (*models.FeatureTemplateCategory, error)
	// UpdateCategory 更新自定义分类
	UpdateCategory(ctx context.Context, id string, label string, sortOrder *int) (*models.FeatureTemplateCategory, error)
	// DeleteCategory 删除自定义分类
	DeleteCategory(ctx context.Context, id string) error
}

// ProductOfferingService 产品项服务接口
type ProductOfferingService interface {
	// GetProductOfferingList 查询产品项列表
	GetProductOfferingList(ctx context.Context, req *models.ProductOfferingListRequest) (*models.ProductOfferingListResponse, error)
	// GetProductOffering 获取产品项详情
	GetProductOffering(ctx context.Context, id string) (*models.ProductOfferingResponse, error)
	// CreateProductOffering 创建产品项，编码由服务端生成
	CreateProductOffering(ctx context.Context, req *models.ProductOfferingCreateRequest) (*models.ProductOfferingResponse, error)
	// UpdateProductOffering 更新产品项，编码不可更改
	UpdateProductOffering(ctx context.Context, id string, req *models.ProductOfferingUpdateRequest) (*models.ProductOfferingResponse, error)
	// DeleteProductOffering 删除产品项
	DeleteProductOffering(ctx context.Context, id string) error
	// GetProductOfferingPublicKey 获取产品项公钥
	GetProductOfferingPublicKey(ctx context.Context, id string) (*models.ProductOfferingPublicKeyResponse, error)
	// UpdateProductOfferingStatus 启用/停用产品项
	UpdateProductOfferingStatus(ctx context.Context, id string, status string) (*models.ProductOfferingResponse, error)
}

// AuthorizationCodeService 授权码服务接口
type AuthorizationCodeService interface {
	CreateAuthorizationCode(ctx context.Context, req *models.AuthorizationCodeCreateRequest) (*models.AuthorizationCodeCreateResponse, error)
	BatchCreateAuthorizationCodes(ctx context.Context, req *models.AuthorizationCodeBatchCreateRequest) (*models.AuthorizationCodeBatchCreateResponse, error)
	GetAuthorizationCodeList(ctx context.Context, req *models.AuthorizationCodeListRequest) (*models.AuthorizationCodeListResponse, error)
	GetAuthorizationCode(ctx context.Context, id string) (*models.AuthorizationCode, error)
	UpdateAuthorizationCode(ctx context.Context, id string, req *models.AuthorizationCodeUpdateRequest) (*models.AuthorizationCode, error)
	LockUnlockAuthorizationCode(ctx context.Context, id string, req *models.AuthorizationCodeLockRequest) (*models.AuthorizationCode, error)
	DeleteAuthorizationCode(ctx context.Context, id string) error
	GetAuthorizationChangeList(ctx context.Context, authCodeID string, req *models.AuthorizationChangeListRequest) (*models.AuthorizationChangeListResponse, error)
	GenerateAuthorizationFile(ctx context.Context, id string) ([]byte, string, string, error)
}

// LicenseService 许可证服务接口
type LicenseService interface {
	GetLicenseList(ctx context.Context, req *models.LicenseListRequest) (*models.LicenseListResponse, error)
	GetLicense(ctx context.Context, id string) (*models.LicenseDetailResponse, error)
	CreateLicense(ctx context.Context, req *models.LicenseCreateRequest) (*models.License, error)
	RevokeLicense(ctx context.Context, id string, req *models.LicenseRevokeRequest) (*models.License, error)
	DeleteLicense(ctx context.Context, id string) error
	GenerateLicenseFile(ctx context.Context, id string) ([]byte, string, string, error)

	// 客户端激活和心跳接口
	ActivateLicense(ctx context.Context, req *models.ActivateRequest, clientIP string) (*models.ActivateResponse, error)
	Heartbeat(ctx context.Context, req *models.HeartbeatRequest, clientIP string) (*models.HeartbeatResponse, error)

	// 统计接口
	GetStatsOverview(ctx context.Context) (*models.StatsOverviewResponse, error)
}

// DashboardService 仪表盘服务接口
type DashboardService interface {
	// 获取首页聚合数据
	GetHome(ctx context.Context) (*models.DashboardHomeResponse, error)

	// 获取授权码创建和许可证激活趋势
	GetBusinessTrends(ctx context.Context, req *models.DashboardBusinessTrendsRequest) (*models.DashboardBusinessTrendsResponse, error)

	// 获取授权趋势数据
	GetAuthorizationTrend(ctx context.Context, req *models.DashboardAuthorizationTrendRequest) (*models.DashboardAuthorizationTrendResponse, error)

	// 获取最近授权列表
	GetRecentAuthorizations(ctx context.Context, req *models.DashboardRecentAuthorizationsRequest) (*models.DashboardRecentAuthorizationsResponse, error)
}
