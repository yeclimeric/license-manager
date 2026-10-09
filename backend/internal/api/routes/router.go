package routes

import (
	_ "license-manager/docs/swagger" // swagger docs
	"license-manager/internal/api/handlers"
	"license-manager/internal/api/middleware"
	"license-manager/internal/config"
	"license-manager/internal/database"
	"license-manager/internal/repository"
	"license-manager/internal/service"
	"license-manager/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	// 权限校验器：直接查关联表，避免 middleware 反向依赖 service 层
	permChecker := middleware.NewGormPermissionChecker(database.GetDB())
	router := gin.New()

	// 全局中间件
	router.Use(middleware.CustomLoggerMiddleware())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// 多语言中间件
	cfg := config.GetConfig()
	if cfg.I18n.Enable {
		i18nConfig := &middleware.I18nConfig{
			Enable:       cfg.I18n.Enable,
			DefaultLang:  cfg.I18n.DefaultLang,
			SupportLangs: cfg.I18n.SupportLangs,
		}
		router.Use(middleware.I18nMiddleware(i18nConfig))
	}

	// 初始化数据访问层
	db := database.GetDB()
	customerRepo := repository.NewCustomerRepository(db)
	userRepo := repository.NewUserRepository(db)
	authCodeRepo := repository.NewAuthorizationCodeRepository(db)
	licenseRepo := repository.NewLicenseRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	productOfferingRepo := repository.NewProductOfferingRepository(db)
	featureTemplateRepo := repository.NewFeatureTemplateRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	rbacRepo := repository.NewRBACRepository(db)

	// 获取logger实例
	log := logger.GetLogger()

	// 初始化服务层
	authService := service.NewAuthService(userRepo)
	systemService := service.NewSystemService()
	customerService := service.NewCustomerService(customerRepo)
	authCodeService := service.NewAuthorizationCodeService(authCodeRepo, customerRepo, licenseRepo, productOfferingRepo, featureTemplateRepo)
	enumService := service.NewEnumService()
	licenseService := service.NewLicenseService(licenseRepo, productOfferingRepo, db, log)
	dashboardService := service.NewDashboardService(dashboardRepo)
	productOfferingService := service.NewProductOfferingService(productOfferingRepo)
	featureTemplateService := service.NewFeatureTemplateService(featureTemplateRepo)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	rbacService := service.NewRBACService(rbacRepo, userRepo)

	// 初始化处理器层
	authHandler := handlers.NewAuthHandler(authService)
	systemHandler := handlers.NewSystemHandler(systemService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	enumHandler := handlers.NewEnumHandler(enumService)
	authCodeHandler := handlers.NewAuthorizationCodeHandler(authCodeService)
	licenseHandler := handlers.NewLicenseHandler(licenseService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	productOfferingHandler := handlers.NewProductOfferingHandler(productOfferingService)
	featureTemplateHandler := handlers.NewFeatureTemplateHandler(featureTemplateService)
	auditLogHandler := handlers.NewAuditLogHandler(auditLogService)
	rbacHandler := handlers.NewRBACHandler(rbacService)

	// 系统配置与 API Key
	systemConfigRepo := repository.NewSystemConfigRepository(db)
	systemConfigService := service.NewSystemConfigService(systemConfigRepo, userRepo)
	systemConfigHandler := handlers.NewSystemConfigHandler(systemConfigService)
	apiKeyChecker := middleware.NewGormAPIKeyChecker(db, systemConfigRepo)
	userManagementHandler := handlers.NewUserManagementHandler(service.NewUserManagementService(userRepo))

	// 健康检测接口（无需认证）
	router.GET("/health", systemHandler.HealthCheck)

	// Swagger文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// API路由组
	api := router.Group("/api")
	// API Key 与登录态并行：两者形态不同，据此区分调用方
	api.Use(middleware.APIKeyAuthMiddleware(apiKeyChecker))
	{
		// 公开接口（无需认证）
		public := api.Group("")
		{
			public.POST("/v1/login", authHandler.Login)

			// 许可证激活接口（客户端软件使用）
			public.POST("/v1/activate", licenseHandler.ActivateLicense)
			public.POST("/v1/heartbeat", licenseHandler.Heartbeat)
		}

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		// 审计只关注写操作，挂在鉴权之后以便记录操作人
		auth.Use(middleware.AuditResponseCapture())
		auth.Use(middleware.AuditMiddleware(middleware.AuditConfig{DB: db}))
		{
			// 认证相关
			auth.POST("/v1/logout", authHandler.Logout)
			auth.POST("/v1/auth/refresh", authHandler.RefreshToken)

			// 客户管理
			auth.GET("/customers", customerHandler.GetCustomerList)
			auth.GET("/customers/:id", customerHandler.GetCustomer)
			auth.POST("/customers", middleware.RequirePermission("customer:create", permChecker), customerHandler.CreateCustomer)
			auth.PUT("/customers/:id", middleware.RequirePermission("customer:update", permChecker), customerHandler.UpdateCustomer)
			auth.DELETE("/customers/:id", middleware.RequirePermission("customer:delete", permChecker), customerHandler.DeleteCustomer)
			auth.PATCH("/customers/:id/status", customerHandler.UpdateCustomerStatus)

			// 枚举管理
			auth.GET("/enums", enumHandler.GetAllEnums)
			auth.GET("/enums/:type", enumHandler.GetEnumsByType)

			// 产品项管理
			auth.GET("/v1/product-offerings", productOfferingHandler.GetProductOfferingList)
			auth.POST("/v1/product-offerings", middleware.RequirePermission("product_offering:create", permChecker), productOfferingHandler.CreateProductOffering)
			auth.GET("/v1/product-offerings/:id", productOfferingHandler.GetProductOffering)
			auth.GET("/v1/product-offerings/:id/public-key", productOfferingHandler.GetProductOfferingPublicKey)
			auth.PUT("/v1/product-offerings/:id", productOfferingHandler.UpdateProductOffering)
			auth.DELETE("/v1/product-offerings/:id", productOfferingHandler.DeleteProductOffering)
			auth.PATCH("/v1/product-offerings/:id/status", productOfferingHandler.UpdateProductOfferingStatus)

			// 功能模板管理
			auth.GET("/v1/feature-templates", featureTemplateHandler.GetFeatureTemplateList)
			auth.POST("/v1/feature-templates", middleware.RequirePermission("feature_template:create", permChecker), featureTemplateHandler.CreateFeatureTemplate)
			auth.GET("/v1/feature-templates/categories", featureTemplateHandler.GetCategoryList)
			auth.GET("/v1/feature-templates/:id", featureTemplateHandler.GetFeatureTemplate)
			auth.PUT("/v1/feature-templates/:id", featureTemplateHandler.UpdateFeatureTemplate)
			auth.DELETE("/v1/feature-templates/:id", featureTemplateHandler.DeleteFeatureTemplate)
			auth.POST("/v1/feature-template-categories", featureTemplateHandler.CreateCategory)
			auth.PUT("/v1/feature-template-categories/:id", featureTemplateHandler.UpdateCategory)
			auth.DELETE("/v1/feature-template-categories/:id", featureTemplateHandler.DeleteCategory)

			// 审计日志
			auth.GET("/v1/audit-logs", middleware.RequirePermission("audit_log:read", permChecker), auditLogHandler.GetAuditLogList)
			auth.GET("/v1/audit-logs/statistics", middleware.RequirePermission("audit_log:read", permChecker), auditLogHandler.GetStatistics)

			// 菜单与权限。/menus 无需权限：任何登录用户都要读自己的菜单
			auth.GET("/v1/menus", rbacHandler.GetUserMenuTree)
			auth.GET("/v1/permissions", rbacHandler.GetPermissionMatrix)
			auth.GET("/v1/menus/all", middleware.RequirePermission("menu:read", permChecker), rbacHandler.GetAllMenuTree)
			auth.POST("/v1/menus", middleware.RequirePermission("menu:create", permChecker), rbacHandler.CreateMenu)
			auth.PUT("/v1/menus/:id", middleware.RequirePermission("menu:update", permChecker), rbacHandler.UpdateMenu)
			auth.DELETE("/v1/menus/:id", middleware.RequirePermission("menu:delete", permChecker), rbacHandler.DeleteMenu)

			// 角色管理
			auth.GET("/v1/roles", middleware.RequirePermission("role:read", permChecker), rbacHandler.GetRoleList)
			auth.POST("/v1/roles", middleware.RequirePermission("role:create", permChecker), rbacHandler.CreateRole)
			auth.GET("/v1/roles/:id", middleware.RequirePermission("role:read", permChecker), rbacHandler.GetRole)
			auth.PUT("/v1/roles/:id", middleware.RequirePermission("role:update", permChecker), rbacHandler.UpdateRole)
			auth.DELETE("/v1/roles/:id", middleware.RequirePermission("role:delete", permChecker), rbacHandler.DeleteRole)

			// 系统配置
			auth.GET("/v1/system-configs", middleware.RequirePermission("system_config:read", permChecker), systemConfigHandler.GetGroupedConfigs)
			auth.POST("/v1/system-configs", middleware.RequirePermission("system_config:create", permChecker), systemConfigHandler.CreateConfig)
			// batch 必须在 :key 之前
			auth.PUT("/v1/system-configs/batch", middleware.RequirePermission("system_config:update", permChecker), systemConfigHandler.BatchUpdateConfigs)
			auth.PUT("/v1/system-configs/:key", middleware.RequirePermission("system_config:update", permChecker), systemConfigHandler.UpdateConfig)
			auth.DELETE("/v1/system-configs/:key", middleware.RequirePermission("system_config:delete", permChecker), systemConfigHandler.DeleteConfig)

			// API Key
			auth.GET("/v1/api-keys", middleware.RequirePermission("api_key:read", permChecker), systemConfigHandler.GetAPIKeyList)
			auth.POST("/v1/api-keys", middleware.RequirePermission("api_key:create", permChecker), systemConfigHandler.CreateAPIKey)
			auth.PUT("/v1/api-keys/:id", middleware.RequirePermission("api_key:update", permChecker), systemConfigHandler.UpdateAPIKey)
			auth.DELETE("/v1/api-keys/:id", middleware.RequirePermission("api_key:delete", permChecker), systemConfigHandler.DeleteAPIKey)

			// 用户角色绑定（按ID写入，供角色页分配）
			auth.PUT("/v1/users/:id/roles", middleware.RequirePermission("user:update", permChecker), rbacHandler.SetUserRoles)

			// 用户管理
			auth.GET("/v1/users", middleware.RequirePermission("user:read", permChecker), userManagementHandler.GetUserList)
			auth.POST("/v1/users", middleware.RequirePermission("user:create", permChecker), userManagementHandler.CreateUser)
			auth.GET("/v1/users/:id", middleware.RequirePermission("user:read", permChecker), userManagementHandler.GetUserDetail)
			auth.PUT("/v1/users/:id", middleware.RequirePermission("user:update", permChecker), userManagementHandler.UpdateUser)
			auth.DELETE("/v1/users/:id", middleware.RequirePermission("user:delete", permChecker), userManagementHandler.DeleteUser)
			auth.POST("/v1/users/:id/reset-password", middleware.RequirePermission("user:update", permChecker), userManagementHandler.ResetPassword)
			auth.GET("/v1/users/:id/roles", middleware.RequirePermission("user:read", permChecker), userManagementHandler.GetUserRoles)

			// 授权码管理
			auth.GET("/v1/authorization-codes", authCodeHandler.GetAuthorizationCodeList)
			auth.POST("/v1/authorization-codes", middleware.RequirePermission("authorization_code:create", permChecker), authCodeHandler.CreateAuthorizationCode)
			// 批量创建上限 100 条，按单次创建权限管控
			auth.POST("/v1/authorization-codes/batch", middleware.RequirePermission("authorization_code:create", permChecker), authCodeHandler.BatchCreateAuthorizationCodes)
			auth.GET("/v1/authorization-codes/:id", authCodeHandler.GetAuthorizationCode)
			auth.GET("/v1/authorization-codes/:id/download", authCodeHandler.DownloadAuthorizationFile)
			auth.PUT("/v1/authorization-codes/:id", authCodeHandler.UpdateAuthorizationCode)
			auth.PUT("/v1/authorization-codes/:id/lock", authCodeHandler.LockUnlockAuthorizationCode)
			auth.DELETE("/v1/authorization-codes/:id", authCodeHandler.DeleteAuthorizationCode)
			auth.GET("/v1/authorization-codes/:id/changes", authCodeHandler.GetAuthorizationChangeList)

			// 许可证管理
			auth.GET("/v1/licenses", licenseHandler.GetLicenseList)
			auth.GET("/v1/licenses/:id", licenseHandler.GetLicense)
			auth.POST("/v1/licenses", licenseHandler.CreateLicense)
			auth.PUT("/v1/licenses/:id/revoke", licenseHandler.RevokeLicense)
			auth.DELETE("/v1/licenses/:id", licenseHandler.DeleteLicense)
			auth.GET("/v1/licenses/:id/download", licenseHandler.DownloadLicenseFile)

			// 统计分析
			auth.GET("/v1/stats/overview", licenseHandler.GetStatsOverview)

			// 仪表盘接口
			auth.GET("/v1/dashboard/home", dashboardHandler.GetHome)
			auth.GET("/v1/dashboard/trends", dashboardHandler.GetBusinessTrends)
			auth.GET("/v1/dashboard/authorization-trend", dashboardHandler.GetAuthorizationTrend)
			auth.GET("/v1/dashboard/recent-authorizations", dashboardHandler.GetRecentAuthorizations)
		}

		// 管理员接口
		admin := api.Group("/v1/admin")
		// 修复：原实现比对 "administrator"，而角色实际取值为 admin/operator/viewer，
		// 导致该分组内接口无人可通过。改由权限码校验。
		admin.Use(middleware.AuthMiddleware(),
			middleware.RequirePermission("system_config:read", permChecker))
		{
			admin.GET("/system/info", systemHandler.GetSystemInfo)
		}
	}

	// 公开配置：登录页需在未认证时读取站点信息
	router.GET("/api/public/config", systemConfigHandler.GetPublicConfig)

	return router
}
