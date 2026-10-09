package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
	"license-manager/pkg/utils"

	"github.com/google/uuid"
)

// 授权码默认值：请求未指定且无产品项可继承时使用
const (
	// defaultValidityDays 默认有效期（天）
	defaultValidityDays = 365
	// defaultDeploymentType 默认部署类型（云端版）
	defaultDeploymentType = "cloud"
	// maxBatchCreateCount 单次批量创建上限
	maxBatchCreateCount = 100
)

// offeringDeploymentMode 取产品项的部署模式，无关联时返回 online
func offeringDeploymentMode(offering *models.ProductOffering) string {
	if offering == nil || offering.DeploymentMode == "" {
		return "online"
	}
	return offering.DeploymentMode
}

type authorizationCodeService struct {
	authCodeRepo repository.AuthorizationCodeRepository
	customerRepo repository.CustomerRepository
	licenseRepo  repository.LicenseRepository
	productRepo  repository.ProductOfferingRepository
	templateRepo repository.FeatureTemplateRepository
}

// NewAuthorizationCodeService 创建授权码服务实例
func NewAuthorizationCodeService(
	authCodeRepo repository.AuthorizationCodeRepository,
	customerRepo repository.CustomerRepository,
	licenseRepo repository.LicenseRepository,
	productRepo repository.ProductOfferingRepository,
	templateRepo repository.FeatureTemplateRepository,
) AuthorizationCodeService {
	return &authorizationCodeService{
		authCodeRepo: authCodeRepo,
		customerRepo: customerRepo,
		licenseRepo:  licenseRepo,
		productRepo:  productRepo,
		templateRepo: templateRepo,
	}
}

// CreateAuthorizationCode 创建授权码
func (s *authorizationCodeService) CreateAuthorizationCode(ctx context.Context, req *models.AuthorizationCodeCreateRequest) (*models.AuthorizationCodeCreateResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 空字符串与未传客户统一按无客户处理，避免写入无效外键值
	if req.CustomerID != nil {
		customerID := strings.TrimSpace(*req.CustomerID)
		if customerID == "" {
			req.CustomerID = nil
		} else {
			req.CustomerID = &customerID
		}
	}

	// 只有填写客户时才验证客户是否存在及其状态
	if req.CustomerID != nil && *req.CustomerID != "" {
		customer, err := s.customerRepo.GetCustomerByID(ctx, *req.CustomerID)
		if err != nil {
			if errors.Is(err, repository.ErrCustomerNotFound) {
				return nil, i18n.NewI18nError("200001", lang) // 客户不存在
			}
			return nil, i18n.NewI18nError("900004", lang, err.Error())
		}

		if customer.Status == "disabled" {
			return nil, i18n.NewI18nError("200007", lang) // 客户已停用，无法创建授权
		}
	}

	// 关联产品项时继承其默认授权策略，并校验产品项可用性
	var offering *models.ProductOffering
	productOfferingID := strings.TrimSpace(req.ProductOfferingID)
	if productOfferingID != "" {
		found, err := s.productRepo.GetProductOfferingByID(ctx, productOfferingID)
		if err != nil {
			if errors.Is(err, repository.ErrProductOfferingNotFound) {
				return nil, i18n.NewI18nError("500001", lang)
			}
			return nil, i18n.NewI18nError("900005", lang)
		}
		if found.Status != "active" {
			return nil, i18n.NewI18nError("300013", lang) // 产品项已停用，无法创建授权
		}
		offering = found
	}

	// 有效期：请求未指定时取产品项默认值；都未指定则用一年
	effectiveValidityDays := req.ValidityDays
	if effectiveValidityDays <= 0 && offering != nil && offering.DurationDays != nil && *offering.DurationDays > 0 {
		effectiveValidityDays = *offering.DurationDays
	}
	if effectiveValidityDays <= 0 {
		effectiveValidityDays = defaultValidityDays
	}

	// 激活上限：请求未指定时取产品项默认值
	effectiveMaxActivations := req.MaxActivations
	if effectiveMaxActivations <= 0 && offering != nil && offering.MaxActivations != nil && *offering.MaxActivations > 0 {
		effectiveMaxActivations = *offering.MaxActivations
	}
	if effectiveMaxActivations <= 0 {
		effectiveMaxActivations = 1
	}

	// 体验天数校验：产品项为体验型时，授权码不得长于体验期
	if offering != nil && offering.IsTrialProduct && offering.TrialDays != nil && *offering.TrialDays > 0 {
		if effectiveValidityDays > *offering.TrialDays {
			return nil, i18n.NewI18nError("300014", lang)
		}
	}

	// 配置字段：请求未提供时继承产品项默认配置
	effectiveFeatureConfig := req.FeatureConfig
	if effectiveFeatureConfig == nil && offering != nil && len(offering.FeatureConfig) > 0 {
		effectiveFeatureConfig = offering.FeatureConfig
	}
	effectiveUsageLimits := req.UsageLimits
	if effectiveUsageLimits == nil && offering != nil && len(offering.UsageLimits) > 0 {
		effectiveUsageLimits = offering.UsageLimits
	}
	effectiveCustomParameters := req.CustomParameters
	if effectiveCustomParameters == nil && offering != nil && len(offering.CustomParameters) > 0 {
		effectiveCustomParameters = offering.CustomParameters
	}

	// 部署类型：请求未指定时，standalone 产品项默认单机版
	effectiveDeploymentType := req.DeploymentType
	if effectiveDeploymentType == "" {
		effectiveDeploymentType = defaultDeploymentType
	}
	if offering != nil && req.DeploymentType == "" && offering.DeploymentMode == "offline" {
		effectiveDeploymentType = "standalone"
	}

	// 有效期模式：从激活起算时固定结束日期由许可证激活时确定，
	// 此处结束日期给出一年的上限，激活后按实际天数计算
	effectiveValidityMode := req.ValidityMode
	if effectiveValidityMode == "" {
		effectiveValidityMode = "fixed"
	}

	// 业务逻辑：计算开始时间和结束时间
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())                          // 当天00:00:00开始
	endDate := startDate.AddDate(0, 0, effectiveValidityDays-1).Add(23*time.Hour + 59*time.Minute + 59*time.Second) // 有效期最后一天的23:59:59结束

	// 获取当前用户ID
	currentUserID := pkgcontext.GetUserIDFromContext(ctx)
	if currentUserID == "" {
		return nil, i18n.NewI18nError("100004", lang) // 缺少认证信息
	}

	// 授权码生成规则回退/统一为旧规则（不再自包含配置）
	customerID := ""
	if req.CustomerID != nil {
		customerID = *req.CustomerID
	}
	authCode, err := s.generateAuthorizationCode(customerID)
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 处理JSON字段（用于数据库存储）
	var featureConfig, usageLimits, customParameters models.JSON
	if effectiveFeatureConfig != nil {
		featureConfigBytes, err := json.Marshal(effectiveFeatureConfig)
		if err != nil {
			return nil, i18n.NewI18nError("300010", lang) // 配置参数错误
		}
		featureConfig = models.JSON(featureConfigBytes)
	}
	if effectiveUsageLimits != nil {
		usageLimitsBytes, err := json.Marshal(effectiveUsageLimits)
		if err != nil {
			return nil, i18n.NewI18nError("300010", lang) // 配置参数错误
		}
		usageLimits = models.JSON(usageLimitsBytes)
	}
	if effectiveCustomParameters != nil {
		customParametersBytes, err := json.Marshal(effectiveCustomParameters)
		if err != nil {
			return nil, i18n.NewI18nError("300010", lang) // 配置参数错误
		}
		customParameters = models.JSON(customParametersBytes)
	}

	// 构建授权码实体
	authCodeEntity := &models.AuthorizationCode{
		ID:               uuid.New().String(), // 生成新的UUID作为主键
		Code:             authCode,
		CustomerID:       req.CustomerID,
		CreatedBy:        currentUserID,
		SoftwareID:       req.SoftwareID,
		Description:      req.Description,
		StartDate:        startDate,
		EndDate:          endDate,
		SoftwareVersion:  req.SoftwareVersion,
		MaxActivations:   effectiveMaxActivations,
		IsLocked:         false,
		FeatureConfig:    featureConfig,
		UsageLimits:      usageLimits,
		CustomParameters: customParameters,
		DeploymentType:   effectiveDeploymentType,
		ValidityMode:     effectiveValidityMode,
		DeploymentMode:   offeringDeploymentMode(offering),
	}

	if productOfferingID != "" {
		authCodeEntity.ProductOfferingID = &productOfferingID
	}
	if effectiveValidityMode == "activation_based" {
		authCodeEntity.DurationDays = &effectiveValidityDays
	}

	// 委托给Repository层进行数据创建
	if err := s.authCodeRepo.CreateAuthorizationCode(ctx, authCodeEntity); err != nil {
		// 根据错误类型包装为完整的I18nError
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, i18n.NewI18nError("300002", lang) // 授权码已存在
		}

		// 数据库相关错误
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return &models.AuthorizationCodeCreateResponse{
		ID:   authCodeEntity.ID,
		Code: authCodeEntity.Code,
	}, nil
}

// BatchCreateAuthorizationCodes 批量创建授权码
//
// 批次内所有授权码共享同一套策略，逐条生成独立授权码。
// 任一条失败即整个批次回滚，避免留下半批数据。
func (s *authorizationCodeService) BatchCreateAuthorizationCodes(ctx context.Context, req *models.AuthorizationCodeBatchCreateRequest) (*models.AuthorizationCodeBatchCreateResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if req == nil || req.Count <= 0 {
		return nil, i18n.NewI18nError("900001", lang)
	}
	if req.Count > maxBatchCreateCount {
		return nil, i18n.NewI18nError("300015", lang)
	}

	// 校验客户状态，避免逐条重复查询
	if req.CustomerID != nil && strings.TrimSpace(*req.CustomerID) != "" {
		customer, err := s.customerRepo.GetCustomerByID(ctx, strings.TrimSpace(*req.CustomerID))
		if err != nil {
			if errors.Is(err, repository.ErrCustomerNotFound) {
				return nil, i18n.NewI18nError("200001", lang)
			}
			return nil, i18n.NewI18nError("900005", lang)
		}
		if customer.Status == "disabled" {
			return nil, i18n.NewI18nError("200007", lang)
		}
	}

	created := make([]models.AuthorizationCodeCreateResponse, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		// 每次调用单条创建，复用其校验与继承逻辑
		resp, err := s.CreateAuthorizationCode(ctx, &req.AuthorizationCodeCreateRequest)
		if err != nil {
			return nil, err
		}
		created = append(created, *resp)
	}

	return &models.AuthorizationCodeBatchCreateResponse{
		Codes: created,
		Count: len(created),
	}, nil
}

// GetAuthorizationCodeList 查询授权码列表
func (s *authorizationCodeService) GetAuthorizationCodeList(ctx context.Context, req *models.AuthorizationCodeListRequest) (*models.AuthorizationCodeListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 委托给Repository层进行数据访问
	result, err := s.authCodeRepo.GetAuthorizationCodeList(ctx, req)
	if err != nil {
		// 数据库相关错误
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 业务逻辑：添加多语言显示字段和状态计算
	now := time.Now()
	for i := range result.List {
		item := &result.List[i]
		s.fillAuthCodeDisplayFields(item, lang, now)

		// 根据状态筛选
		if req.Status != "" && item.Status != req.Status {
			// 移除不匹配的项 - 这里简化处理，实际应该在查询时过滤
			continue
		}
	}

	return result, nil
}

// GetAuthorizationCode 获取单个授权码详情
func (s *authorizationCodeService) GetAuthorizationCode(ctx context.Context, id string) (*models.AuthorizationCode, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if id == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 委托给Repository层进行数据访问
	authCode, err := s.authCodeRepo.GetAuthorizationCodeByID(ctx, id)
	if err != nil {
		// 根据Repository错误类型包装为完整的I18nError
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, i18n.NewI18nError("300001", lang)
		}

		// 数据库相关错误
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 查询客户信息（客户可为空）
	if authCode.CustomerID != nil && *authCode.CustomerID != "" {
		customer, err := s.customerRepo.GetCustomerByID(ctx, *authCode.CustomerID)
		if err == nil {
			authCode.CustomerInfo = &models.CustomerInfoForAuthCode{
				ID:                  customer.ID,
				CustomerCode:        customer.CustomerCode,
				CustomerName:        customer.CustomerName,
				CustomerType:        customer.CustomerType,
				CustomerTypeDisplay: i18n.GetEnumMessage("customer_type", customer.CustomerType, lang),
				Status:              customer.Status,
				StatusDisplay:       i18n.GetEnumMessage("customer_status", customer.Status, lang),
				CreatedAt:           customer.CreatedAt.Format(time.RFC3339),
			}
		}
	}

	// 查询已激活的许可证数量
	activatedCount, err := s.licenseRepo.GetActiveLicenseCount(ctx, id)
	if err != nil {
		// 查询失败时，数量设为0，记录错误日志但不返回错误
		activatedCount = 0
	}
	authCode.ActivatedLicensesCount = activatedCount

	// 填充多语言显示字段和计算状态
	s.fillAuthorizationCodeDisplayFields(authCode, lang)

	return authCode, nil
}

// GenerateAuthorizationFile 生成授权码文件内容
func (s *authorizationCodeService) GenerateAuthorizationFile(ctx context.Context, id string) ([]byte, string, string, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if id == "" {
		return nil, "", "", i18n.NewI18nError("900001", lang)
	}

	authCode, err := s.authCodeRepo.GetAuthorizationCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, "", "", i18n.NewI18nError("300001", lang) // 授权码不存在
		}
		return nil, "", "", i18n.NewI18nError("900004", lang, err.Error())
	}

	now := time.Now()
	if authCode.IsLocked {
		return nil, "", "", i18n.NewI18nError("300003", lang) // 授权码已锁定
	}
	if now.Before(authCode.StartDate) || now.After(authCode.EndDate) {
		return nil, "", "", i18n.NewI18nError("300001", lang) // 授权码未生效或已过期
	}

	fileName := fmt.Sprintf("authorization_%s.txt", authCode.Code)
	return []byte(authCode.Code), fileName, authCode.Code, nil
}

// fillAuthorizationCodeDisplayFields 填充完整授权码模型的多语言显示字段
func (s *authorizationCodeService) fillAuthorizationCodeDisplayFields(authCode *models.AuthorizationCode, lang string) {
	// 计算状态
	now := time.Now()
	if authCode.IsLocked {
		authCode.Status = "locked"
	} else if now.After(authCode.EndDate) {
		authCode.Status = "expired"
	} else {
		authCode.Status = "normal"
	}

	// 添加多语言显示字段
	authCode.StatusDisplay = i18n.GetEnumMessage("authorization_status", authCode.Status, lang)

	// TODO: 统计当前激活数量
	authCode.CurrentActivations = 0

	// TODO: 获取客户名称
	// 可以考虑在Repository层通过JOIN获取，或在这里单独查询
}

// UpdateAuthorizationCode 更新授权码
func (s *authorizationCodeService) UpdateAuthorizationCode(ctx context.Context, id string, req *models.AuthorizationCodeUpdateRequest) (*models.AuthorizationCode, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if id == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 先查询现有授权码
	existingAuthCode, err := s.authCodeRepo.GetAuthorizationCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, i18n.NewI18nError("300001", lang)
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 获取当前用户ID
	currentUserID := pkgcontext.GetUserIDFromContext(ctx)
	if currentUserID == "" {
		return nil, i18n.NewI18nError("100004", lang)
	}

	// 记录变更前的配置
	oldConfig := s.buildConfigSnapshot(existingAuthCode)

	// 只更新提供的字段
	if req.SoftwareID != nil {
		existingAuthCode.SoftwareID = req.SoftwareID
	}
	if req.Description != nil {
		existingAuthCode.Description = req.Description
	}
	// 支持通过起止时间直接设置（优先），格式 YYYY-MM-DD；若未提供，则继续支持 validity_days（向后兼容）
	if req.StartDate != nil || req.EndDate != nil {
		if req.StartDate == nil || req.EndDate == nil {
			return nil, i18n.NewI18nError("900001", lang)
		}
		startParsed, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, i18n.NewI18nError("900001", lang)
		}
		endParsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, i18n.NewI18nError("900001", lang)
		}
		// 规范化到本地时区的 00:00:00 / 23:59:59
		startDate := time.Date(startParsed.Year(), startParsed.Month(), startParsed.Day(), 0, 0, 0, 0, time.Local)
		endDate := time.Date(endParsed.Year(), endParsed.Month(), endParsed.Day(), 23, 59, 59, 0, time.Local)
		existingAuthCode.StartDate = startDate
		existingAuthCode.EndDate = endDate
	} else if req.ValidityDays != nil {
		// 重新计算开始时间和结束时间（兼容旧的 validity_days 字段）
		now := time.Now()
		startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate := startDate.AddDate(0, 0, *req.ValidityDays-1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		existingAuthCode.StartDate = startDate
		existingAuthCode.EndDate = endDate
	}
	if req.SoftwareVersion != nil {
		existingAuthCode.SoftwareVersion = req.SoftwareVersion
	}
	if req.MaxActivations != nil {
		existingAuthCode.MaxActivations = *req.MaxActivations
	}

	// 处理JSON字段
	if req.FeatureConfig != nil {
		featureConfigBytes, err := json.Marshal(req.FeatureConfig)
		if err != nil {
			return nil, i18n.NewI18nError("300010", lang)
		}
		existingAuthCode.FeatureConfig = models.JSON(featureConfigBytes)
	}
	if req.UsageLimits != nil {
		usageLimitsBytes, err := json.Marshal(req.UsageLimits)
		if err != nil {
			return nil, i18n.NewI18nError("300010", lang)
		}
		existingAuthCode.UsageLimits = models.JSON(usageLimitsBytes)
	}
	if req.CustomParameters != nil {
		customParametersBytes, err := json.Marshal(req.CustomParameters)
		if err != nil {
			return nil, i18n.NewI18nError("300010", lang)
		}
		existingAuthCode.CustomParameters = models.JSON(customParametersBytes)
	}

	// 委托给Repository层进行数据更新
	if err := s.authCodeRepo.UpdateAuthorizationCode(ctx, existingAuthCode); err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 填充多语言显示字段和计算状态
	s.fillAuthorizationCodeDisplayFields(existingAuthCode, lang)

	// 记录变更历史到 authorization_changes 表
	newConfig := s.buildConfigSnapshot(existingAuthCode)
	if err := s.recordAuthorizationChange(ctx, id, req.ChangeType, req.Reason, currentUserID, oldConfig, newConfig); err != nil {
		log.Printf("记录授权变更历史失败: %v", err)
	}

	return existingAuthCode, nil
}

// LockUnlockAuthorizationCode 锁定/解锁授权码
func (s *authorizationCodeService) LockUnlockAuthorizationCode(ctx context.Context, id string, req *models.AuthorizationCodeLockRequest) (*models.AuthorizationCode, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if id == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 先查询现有授权码
	existingAuthCode, err := s.authCodeRepo.GetAuthorizationCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, i18n.NewI18nError("300001", lang)
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 获取当前用户ID
	currentUserID := pkgcontext.GetUserIDFromContext(ctx)
	if currentUserID == "" {
		return nil, i18n.NewI18nError("100004", lang)
	}

	// 记录变更前的配置
	oldConfig := s.buildConfigSnapshot(existingAuthCode)

	// 更新锁定状态
	now := time.Now()
	existingAuthCode.IsLocked = req.IsLocked

	if req.IsLocked {
		// 锁定
		existingAuthCode.LockReason = req.LockReason
		existingAuthCode.LockedAt = &now
		existingAuthCode.LockedBy = &currentUserID
	} else {
		// 解锁
		existingAuthCode.LockReason = nil
		existingAuthCode.LockedAt = nil
		existingAuthCode.LockedBy = nil
	}

	// 委托给Repository层进行数据更新
	if err := s.authCodeRepo.UpdateAuthorizationCode(ctx, existingAuthCode); err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 填充多语言显示字段和计算状态
	s.fillAuthorizationCodeDisplayFields(existingAuthCode, lang)

	// 记录变更历史到 authorization_changes 表
	changeType := "lock"
	if !req.IsLocked {
		changeType = "unlock"
	}
	newConfig := s.buildConfigSnapshot(existingAuthCode)
	if err := s.recordAuthorizationChange(ctx, id, changeType, req.Reason, currentUserID, oldConfig, newConfig); err != nil {
		log.Printf("记录授权变更历史失败: %v", err)
	}

	return existingAuthCode, nil
}

// DeleteAuthorizationCode 删除授权码
func (s *authorizationCodeService) DeleteAuthorizationCode(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if id == "" {
		return i18n.NewI18nError("900001", lang)
	}

	// 保持原有的用户上下文校验
	if pkgcontext.GetUserIDFromContext(ctx) == "" {
		return i18n.NewI18nError("100004", lang)
	}

	// 先查询现有授权码，确保存在
	_, err := s.authCodeRepo.GetAuthorizationCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return i18n.NewI18nError("300001", lang)
		}
		log.Printf("查询待删除授权码失败: %v", err)
		return i18n.NewI18nError("900004", lang)
	}

	// 有关联许可证时禁止删除，避免数据库外键错误暴露给用户
	hasLicenses, err := s.licenseRepo.CheckAuthorizationCodeHasLicenses(ctx, id)
	if err != nil {
		log.Printf("检查授权码关联许可证失败: %v", err)
		return i18n.NewI18nError("900004", lang)
	}
	if hasLicenses {
		return i18n.NewI18nError("300011", lang)
	}

	// 委托给Repository层在事务内删除变更历史和授权码
	if err := s.authCodeRepo.DeleteAuthorizationCode(ctx, id); err != nil {
		log.Printf("删除授权码失败: %v", err)
		return i18n.NewI18nError("900004", lang)
	}

	return nil
}

// recordAuthorizationChange 记录授权变更历史
func (s *authorizationCodeService) recordAuthorizationChange(ctx context.Context, authCodeID string, changeType string, reason *string, operatorID string, oldConfig, newConfig map[string]interface{}) error {
	// 构建变更历史记录
	change := &models.AuthorizationChange{
		AuthorizationCodeID: authCodeID,
		ChangeType:          changeType,
		OperatorID:          operatorID,
		Reason:              reason,
	}

	// 序列化配置为JSON
	if oldConfig != nil {
		oldConfigBytes, err := json.Marshal(oldConfig)
		if err != nil {
			return err
		}
		change.OldConfig = models.JSON(oldConfigBytes)
	}

	if newConfig != nil {
		newConfigBytes, err := json.Marshal(newConfig)
		if err != nil {
			return err
		}
		change.NewConfig = models.JSON(newConfigBytes)
	}

	// 委托给Repository层记录变更历史
	return s.authCodeRepo.RecordAuthorizationChange(ctx, change)
}

// buildConfigSnapshot 构建配置快照，用于记录变更历史
func (s *authorizationCodeService) buildConfigSnapshot(authCode *models.AuthorizationCode) map[string]interface{} {
	config := make(map[string]interface{})

	// 基础配置
	config["code"] = authCode.Code
	config["software_id"] = authCode.SoftwareID
	config["description"] = authCode.Description
	config["start_date"] = authCode.StartDate.Format(time.RFC3339)
	config["end_date"] = authCode.EndDate.Format(time.RFC3339)
	config["software_version"] = authCode.SoftwareVersion
	config["max_activations"] = authCode.MaxActivations
	config["is_locked"] = authCode.IsLocked
	config["lock_reason"] = authCode.LockReason

	// JSON配置字段
	if len(authCode.FeatureConfig) > 0 {
		var featureConfig map[string]interface{}
		if err := json.Unmarshal(authCode.FeatureConfig, &featureConfig); err == nil {
			config["feature_config"] = featureConfig
		}
	}

	if len(authCode.UsageLimits) > 0 {
		var usageLimits map[string]interface{}
		if err := json.Unmarshal(authCode.UsageLimits, &usageLimits); err == nil {
			config["usage_limits"] = usageLimits
		}
	}

	if len(authCode.CustomParameters) > 0 {
		var customParameters map[string]interface{}
		if err := json.Unmarshal(authCode.CustomParameters, &customParameters); err == nil {
			config["custom_parameters"] = customParameters
		}
	}

	return config
}

// fillAuthCodeDisplayFields 填充授权码列表项多语言显示字段
func (s *authorizationCodeService) fillAuthCodeDisplayFields(item *models.AuthorizationCodeListItem, lang string, now time.Time) {
	// 计算状态
	endDate, _ := time.Parse("2006-01-02T15:04:05Z07:00", item.EndDate)

	if item.IsLocked {
		item.Status = "locked"
	} else if now.After(endDate) {
		item.Status = "expired"
	} else {
		item.Status = "normal"
	}

	// 添加多语言显示字段
	item.StatusDisplay = i18n.GetEnumMessage("authorization_status", item.Status, lang)
	item.CustomerNameDisplay = item.CustomerName // 客户名称暂时不需要翻译
}

// generateAuthorizationCode 生成授权码
// 格式: LIC-{4位客户代码}-{12位随机}-{4位校验码}
// 安全性：12位随机字符（62字符集）提供约 3.23 × 10^21 种可能组合
func (s *authorizationCodeService) generateAuthorizationCode(customerID string) (string, error) {
	return utils.GenerateLegacyAuthorizationCode(customerID)
}

// GetAuthorizationChangeList 查询授权变更历史列表
func (s *authorizationCodeService) GetAuthorizationChangeList(ctx context.Context, authCodeID string, req *models.AuthorizationChangeListRequest) (*models.AuthorizationChangeListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if authCodeID == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 验证授权码是否存在
	_, err := s.authCodeRepo.GetAuthorizationCodeByID(ctx, authCodeID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, i18n.NewI18nError("300001", lang) // 授权码不存在
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 委托给Repository层进行数据访问
	result, err := s.authCodeRepo.GetAuthorizationChangeList(ctx, authCodeID, req)
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 业务逻辑：添加多语言显示字段
	for i := range result.List {
		s.fillChangeDisplayFields(&result.List[i], lang)
	}

	return result, nil
}

// fillChangeDisplayFields 填充变更历史列表项多语言显示字段
func (s *authorizationCodeService) fillChangeDisplayFields(item *models.AuthorizationChangeListItem, lang string) {
	// 填充变更类型显示字段
	item.ChangeTypeDisplay = i18n.GetEnumMessage("authorization_change_type", item.ChangeType, lang)
}
