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
	licenseService := service.NewLicenseService(licenseRepo, db, log)
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

	// 健康检测接口（无需认证）
	router.GET("/health", systemHandler.HealthCheck)

	// Swagger文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// API路由组
	api := router.Group("/api")
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
			auth.POST("/customers", customerHandler.CreateCustomer)
			auth.PUT("/customers/:id", customerHandler.UpdateCustomer)
			auth.DELETE("/customers/:id", customerHandler.DeleteCustomer)
			auth.PATCH("/customers/:id/status", customerHandler.UpdateCustomerStatus)

			// 枚举管理
			auth.GET("/enums", enumHandler.GetAllEnums)
			auth.GET("/enums/:type", enumHandler.GetEnumsByType)

			// 产品项管理
			auth.GET("/v1/product-offerings", productOfferingHandler.GetProductOfferingList)
			auth.POST("/v1/product-offerings", productOfferingHandler.CreateProductOffering)
			auth.GET("/v1/product-offerings/:id", productOfferingHandler.GetProductOffering)
			auth.GET("/v1/product-offerings/:id/public-key", productOfferingHandler.GetProductOfferingPublicKey)
			auth.PUT("/v1/product-offerings/:id", productOfferingHandler.UpdateProductOffering)
			auth.DELETE("/v1/product-offerings/:id", productOfferingHandler.DeleteProductOffering)
			auth.PATCH("/v1/product-offerings/:id/status", productOfferingHandler.UpdateProductOfferingStatus)

			// 功能模板管理
			auth.GET("/v1/feature-templates", featureTemplateHandler.GetFeatureTemplateList)
			auth.POST("/v1/feature-templates", featureTemplateHandler.CreateFeatureTemplate)
			auth.GET("/v1/feature-templates/categories", featureTemplateHandler.GetCategoryList)
			auth.GET("/v1/feature-templates/:id", featureTemplateHandler.GetFeatureTemplate)
			auth.PUT("/v1/feature-templates/:id", featureTemplateHandler.UpdateFeatureTemplate)
			auth.DELETE("/v1/feature-templates/:id", featureTemplateHandler.DeleteFeatureTemplate)
			auth.POST("/v1/feature-template-categories", featureTemplateHandler.CreateCategory)
			auth.PUT("/v1/feature-template-categories/:id", featureTemplateHandler.UpdateCategory)
			auth.DELETE("/v1/feature-template-categories/:id", featureTemplateHandler.DeleteCategory)

			// 审计日志
			auth.GET("/v1/audit-logs", auditLogHandler.GetAuditLogList)
			auth.GET("/v1/audit-logs/statistics", auditLogHandler.GetStatistics)

			// 菜单与权限
			auth.GET("/v1/menus", rbacHandler.GetUserMenuTree)
			auth.GET("/v1/menus/all", rbacHandler.GetAllMenuTree)
			auth.POST("/v1/menus", rbacHandler.CreateMenu)
			auth.PUT("/v1/menus/:id", rbacHandler.UpdateMenu)
			auth.DELETE("/v1/menus/:id", rbacHandler.DeleteMenu)
			auth.GET("/v1/permissions", rbacHandler.GetPermissionMatrix)

			// 角色管理
			auth.GET("/v1/roles", rbacHandler.GetRoleList)
			auth.POST("/v1/roles", rbacHandler.CreateRole)
			auth.GET("/v1/roles/:id", rbacHandler.GetRole)
			auth.PUT("/v1/roles/:id", rbacHandler.UpdateRole)
			auth.DELETE("/v1/roles/:id", rbacHandler.DeleteRole)

			// 用户角色绑定
			auth.GET("/v1/users/:userId/roles", rbacHandler.GetUserRoles)
			auth.PUT("/v1/users/:userId/roles", rbacHandler.SetUserRoles)

			// 授权码管理
			auth.GET("/v1/authorization-codes", authCodeHandler.GetAuthorizationCodeList)
			auth.POST("/v1/authorization-codes", authCodeHandler.CreateAuthorizationCode)
			auth.POST("/v1/authorization-codes/batch", authCodeHandler.BatchCreateAuthorizationCodes)
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
		admin.Use(middleware.AuthMiddleware(), middleware.AdminOnlyMiddleware())
		{
			admin.GET("/system/info", systemHandler.GetSystemInfo)
		}
	}

	return router
}
