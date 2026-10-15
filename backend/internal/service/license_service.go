package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"license-manager/internal/config"
	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
	"license-manager/pkg/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type licenseService struct {
	licenseRepo   repository.LicenseRepository
	productRepo   repository.ProductOfferingRepository
	db            *gorm.DB
	logger        *logrus.Logger
	rsaPrivateKey *utils.RSAPrivateKey // 全局RSA私钥（授权码未关联产品项时的兜底）
}

// NewLicenseService 创建许可证服务实例
func NewLicenseService(
	licenseRepo repository.LicenseRepository,
	productRepo repository.ProductOfferingRepository,
	db *gorm.DB,
	logger *logrus.Logger,
) LicenseService {
	return &licenseService{
		licenseRepo: licenseRepo,
		productRepo: productRepo,
		db:          db,
		logger:      logger,
	}
}

// GetLicenseList 查询许可证列表
func (s *licenseService) GetLicenseList(ctx context.Context, req *models.LicenseListRequest) (*models.LicenseListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang) // 业务错误，不覆盖多语言message
	}

	// 委托给Repository层进行数据访问
	result, err := s.licenseRepo.GetLicenseList(ctx, req)
	if err != nil {
		// 数据库相关错误，使用系统错误码，覆盖message显示详细信息
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 业务逻辑：添加多语言显示字段
	for i := range result.List {
		s.fillDisplayFields(&result.List[i], lang)
	}

	return result, nil
}

// fillDisplayFields 填充列表项多语言显示字段
func (s *licenseService) fillDisplayFields(item *models.LicenseListItem, lang string) {
	// 填充状态显示字段
	item.StatusDisplay = i18n.GetEnumMessage("license_status", item.Status, lang)

	// 填充在线状态显示字段
	if item.IsOnline {
		item.IsOnlineDisplay = i18n.GetEnumMessage("license_online_status", "online", lang)
	} else {
		item.IsOnlineDisplay = i18n.GetEnumMessage("license_online_status", "offline", lang)
	}
}

// GetLicense 获取单个许可证详情
func (s *licenseService) GetLicense(ctx context.Context, id string) (*models.LicenseDetailResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if id == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 委托给Repository层进行数据访问
	license, err := s.licenseRepo.GetLicenseByID(ctx, id)
	if err != nil {
		// 根据Repository错误类型包装为完整的I18nError
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return nil, i18n.NewI18nError("300006", lang) // 许可证不存在
		}

		// 数据库相关错误，使用系统错误码，覆盖message显示详细信息
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 转换为详情响应格式
	response := s.convertToDetailResponse(license, lang)

	return response, nil
}

// CreateLicense 创建许可证
func (s *licenseService) CreateLicense(ctx context.Context, req *models.LicenseCreateRequest) (*models.License, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 获取授权码信息以获取客户ID
	authCode, err := s.licenseRepo.GetAuthorizationCodeByID(ctx, req.AuthorizationCodeID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, i18n.NewI18nError("300001", lang) // 授权码不存在
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 生成许可证密钥
	licenseKey, err := s.generateLicenseKey()
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	now := time.Now()

	// 构建许可证实体，手动添加的离线许可证直接视为已激活
	license := &models.License{
		LicenseKey:          licenseKey,
		AuthorizationCodeID: req.AuthorizationCodeID,
		CustomerID:          authCode.CustomerID, // 从授权码获取客户ID
		HardwareFingerprint: req.HardwareFingerprint,
		ActivationIP:        req.ActivationIP,
		Status:              "active",
		ActivatedAt:         &now,
	}

	// 设置设备信息
	if req.DeviceInfo != nil {
		deviceInfoBytes, err := json.Marshal(req.DeviceInfo)
		if err == nil {
			license.DeviceInfo = models.JSON(deviceInfoBytes)
		}
	}

	// 委托给Repository层进行数据创建
	if err := s.licenseRepo.CreateLicense(ctx, license); err != nil {
		// 根据错误类型包装为完整的I18nError
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, i18n.NewI18nError("300007", lang) // 许可证已存在
		}

		// 数据库相关错误
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return license, nil
}

// convertToDetailResponse 转换为详情响应格式
func (s *licenseService) convertToDetailResponse(license *models.License, lang string) *models.LicenseDetailResponse {
	response := &models.LicenseDetailResponse{
		ID:                  license.ID,
		LicenseKey:          license.LicenseKey,
		AuthorizationCodeID: license.AuthorizationCodeID,
		CustomerID:          license.CustomerID,
		HardwareFingerprint: license.HardwareFingerprint,
		ActivationIP:        license.ActivationIP,
		Status:              license.Status,
		IsOnline:            license.IsOnline,
		LastOnlineIP:        license.LastOnlineIP,
		CreatedAt:           license.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           license.UpdatedAt.Format(time.RFC3339),
	}

	// 格式化时间字段
	if license.ActivatedAt != nil {
		activatedAt := license.ActivatedAt.Format(time.RFC3339)
		response.ActivatedAt = &activatedAt
	}
	if license.LastHeartbeat != nil {
		lastHeartbeat := license.LastHeartbeat.Format(time.RFC3339)
		response.LastHeartbeat = &lastHeartbeat
	}
	if license.ConfigUpdatedAt != nil {
		configUpdatedAt := license.ConfigUpdatedAt.Format(time.RFC3339)
		response.ConfigUpdatedAt = &configUpdatedAt
	}

	// 设置关联数据
	if license.AuthorizationCode != nil {
		response.AuthorizationCode = license.AuthorizationCode.Code
	}
	if license.Customer != nil {
		response.CustomerName = license.Customer.CustomerName
	}

	// 解析JSON字段
	if len(license.DeviceInfo) > 0 {
		var deviceInfo map[string]interface{}
		if err := json.Unmarshal(license.DeviceInfo, &deviceInfo); err == nil {
			response.DeviceInfo = deviceInfo
		}
	}
	if len(license.UsageData) > 0 {
		var usageData map[string]interface{}
		if err := json.Unmarshal(license.UsageData, &usageData); err == nil {
			response.UsageData = usageData
		}
	}

	// 填充多语言显示字段
	response.StatusDisplay = i18n.GetEnumMessage("license_status", license.Status, lang)
	if license.IsOnline {
		response.IsOnlineDisplay = i18n.GetEnumMessage("license_online_status", "online", lang)
	} else {
		response.IsOnlineDisplay = i18n.GetEnumMessage("license_online_status", "offline", lang)
	}

	return response
}

// RevokeLicense 撤销许可证
func (s *licenseService) RevokeLicense(ctx context.Context, id string, req *models.LicenseRevokeRequest) (*models.License, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 业务逻辑：参数验证
	if id == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 先查询现有许可证
	existingLicense, err := s.licenseRepo.GetLicenseByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return nil, i18n.NewI18nError("300006", lang) // 许可证不存在
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 检查许可证是否已经被撤销
	if existingLicense.Status == "revoked" {
		return nil, i18n.NewI18nError("300007", lang) // 许可证已被撤销
	}

	// 更新许可证状态为撤销
	existingLicense.Status = "revoked"

	// 如果提供了撤销原因，保存到使用数据中
	if req.Reason != "" {
		revokeData := map[string]interface{}{
			"revoked_at":    time.Now().Format(time.RFC3339),
			"revoke_reason": req.Reason,
		}

		// 保留原有的使用数据
		if len(existingLicense.UsageData) > 0 {
			var existingUsageData map[string]interface{}
			if err := json.Unmarshal(existingLicense.UsageData, &existingUsageData); err == nil {
				for k, v := range existingUsageData {
					revokeData[k] = v
				}
			}
		}

		usageDataBytes, err := json.Marshal(revokeData)
		if err == nil {
			existingLicense.UsageData = models.JSON(usageDataBytes)
		}
	}

	// 委托给Repository层进行数据更新
	if err := s.licenseRepo.UpdateLicense(ctx, existingLicense); err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return existingLicense, nil
}

// DeleteLicense 删除已撤销的许可证。
func (s *licenseService) DeleteLicense(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)
	if id == "" {
		return i18n.NewI18nError("900001", lang)
	}

	existingLicense, err := s.licenseRepo.GetLicenseByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return i18n.NewI18nError("300006", lang)
		}
		return i18n.NewI18nError("900004", lang, err.Error())
	}
	if existingLicense.Status != "revoked" {
		return i18n.NewI18nError("300012", lang)
	}
	if err := s.licenseRepo.DeleteLicensePermanently(ctx, existingLicense); err != nil {
		return i18n.NewI18nError("900004", lang, err.Error())
	}
	return nil
}

// GenerateLicenseFile 生成许可证文件
func (s *licenseService) GenerateLicenseFile(ctx context.Context, id string) ([]byte, string, string, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)
	s.logger.Infof("[GenerateLicenseFile] 开始生成许可证文件，license_id: %s", id)

	// 业务逻辑：参数验证
	if id == "" {
		s.logger.Error("[GenerateLicenseFile] 错误：许可证ID为空")
		return nil, "", "", i18n.NewI18nError("900001", lang)
	}

	// 获取许可证信息
	s.logger.Infof("[GenerateLicenseFile] 查询许可证信息，license_id: %s", id)
	license, err := s.licenseRepo.GetLicenseByID(ctx, id)
	if err != nil {
		s.logger.Errorf("[GenerateLicenseFile] 查询许可证失败，license_id: %s, error: %v", id, err)
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return nil, "", "", i18n.NewI18nError("300006", lang) // 许可证不存在
		}
		return nil, "", "", i18n.NewI18nError("900004", lang, err.Error())
	}

	s.logger.Infof("[GenerateLicenseFile] 许可证信息查询成功，license_key: %s, status: %s", license.LicenseKey, license.Status)

	// 检查许可证状态
	if license.Status == "revoked" {
		s.logger.Warnf("[GenerateLicenseFile] 许可证已被撤销，license_id: %s", id)
		return nil, "", "", i18n.NewI18nError("300007", lang) // 许可证已被撤销
	}

	// 构建许可证文件内容
	licenseFileData := map[string]interface{}{
		"license_key":           license.LicenseKey,
		"authorization_code_id": license.AuthorizationCodeID,
		"hardware_fingerprint":  license.HardwareFingerprint,
		"status":                license.Status,
		"activated_at":          license.ActivatedAt,
		"config_updated_at":     license.ConfigUpdatedAt,
		"generated_at":          time.Now().Format(time.RFC3339),
	}

	// 包含授权码信息
	if license.AuthorizationCode != nil {
		licenseFileData["authorization_code"] = license.AuthorizationCode.Code
		periodStart, periodEnd := resolveLicensePeriod(license, license.AuthorizationCode)
		licenseFileData["start_date"] = periodStart
		licenseFileData["end_date"] = periodEnd
		licenseFileData["max_activations"] = license.AuthorizationCode.MaxActivations
		licenseFileData["deployment_type"] = license.AuthorizationCode.DeploymentType
		licenseFileData["validity_mode"] = license.AuthorizationCode.ValidityMode

		// 包含功能配置
		if featureConfig := parseJSONField(license.AuthorizationCode.FeatureConfig); len(featureConfig) > 0 {
			licenseFileData["feature_config"] = featureConfig
		}

		if usageLimits := parseJSONField(license.AuthorizationCode.UsageLimits); len(usageLimits) > 0 {
			licenseFileData["usage_limits"] = usageLimits
		}

		if customParameters := parseJSONField(license.AuthorizationCode.CustomParameters); len(customParameters) > 0 {
			licenseFileData["custom_parameters"] = customParameters
		}
	}

	// 序列化许可证数据
	licenseJSON, err := json.Marshal(licenseFileData)
	if err != nil {
		return nil, "", "", i18n.NewI18nError("300009", lang) // 许可证文件生成失败
	}

	// 使用RSA数字签名
	encryptedData, err := s.signLicenseFile(licenseJSON, license.AuthorizationCode)
	if err != nil {
		return nil, "", "", i18n.NewI18nError("300009", lang) // 许可证文件生成失败
	}

	// 生成文件名
	fileName := fmt.Sprintf("license_%s.lic", license.LicenseKey)

	return encryptedData, fileName, license.LicenseKey, nil
}

// generateLicenseKey 生成许可证密钥
func (s *licenseService) generateLicenseKey() (string, error) {
	// 生成12字节随机数据
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 转换为16进制字符串并格式化
	hex := strings.ToUpper(hex.EncodeToString(bytes))
	return fmt.Sprintf("LIC-DEVICE-%s", hex), nil
}

// ActivateLicense 激活许可证
func (s *licenseService) ActivateLicense(ctx context.Context, req *models.ActivateRequest, clientIP string) (*models.ActivateResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 参数验证
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 兼容“产品激活码”形式：{授权码}&{payload}
	// 激活流程只需要授权码本体，不依赖payload。
	req.AuthorizationCode = strings.TrimSpace(req.AuthorizationCode)
	if idx := strings.Index(req.AuthorizationCode, "&"); idx > 0 {
		req.AuthorizationCode = strings.TrimSpace(req.AuthorizationCode[:idx])
	}

	if req.AuthorizationCode == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 获取授权码信息
	authCode, err := s.licenseRepo.GetAuthorizationCodeByCode(ctx, req.AuthorizationCode)
	if err != nil {
		if errors.Is(err, repository.ErrAuthorizationCodeNotFound) {
			return nil, i18n.NewI18nError("300001", lang) // 授权码不存在
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 检查授权码状态和有效期
	now := time.Now()
	if authCode.IsLocked {
		return nil, i18n.NewI18nError("300003", lang) // 授权码已被锁定
	}
	if now.Before(authCode.StartDate) || now.After(authCode.EndDate) {
		return nil, i18n.NewI18nError("300001", lang) // 授权码已过期
	}

	// 使用事务确保并发安全
	var response *models.ActivateResponse
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 检查当前激活数量
		count, err := s.licenseRepo.GetActiveLicenseCount(ctx, authCode.ID)
		if err != nil {
			return err
		}

		// 检查是否已存在相同硬件指纹的许可证
		var existingLicense models.License
		err = tx.Where("authorization_code_id = ? AND hardware_fingerprint = ?",
			authCode.ID, req.HardwareFingerprint).First(&existingLicense).Error

		if err == nil {
			// 已存在，直接激活
			existingLicense.Status = "active"
			existingLicense.ActivationIP = &clientIP
			now := time.Now()
			existingLicense.ActivatedAt = &now
			existingLicense.LastHeartbeat = &now
			existingLicense.LastOnlineIP = &clientIP

			if err := tx.Save(&existingLicense).Error; err != nil {
				return err
			}

			// 生成许可证文件
			licenseFile, err := s.generateLicenseFileContent(&existingLicense, authCode)
			if err != nil {
				return err
			}

			response = &models.ActivateResponse{
				LicenseKey:        existingLicense.LicenseKey,
				LicenseFile:       licenseFile,
				HeartbeatInterval: 300, // 5分钟
				PublicKey:         s.productPublicKey(authCode),
			}
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// 检查激活数量限制
		if count >= int64(authCode.MaxActivations) {
			return repository.ErrLicenseNotFound // 使用已有错误，表示激活数量已达上限
		}

		// 生成新的许可证
		licenseKey, err := s.generateLicenseKey()
		if err != nil {
			return err
		}

		license := &models.License{
			LicenseKey:          licenseKey,
			AuthorizationCodeID: authCode.ID,
			CustomerID:          authCode.CustomerID,
			HardwareFingerprint: req.HardwareFingerprint,
			ActivationIP:        &clientIP,
			Status:              "active",
			ActivatedAt:         &now,
			LastHeartbeat:       &now,
			LastOnlineIP:        &clientIP,
		}

		// 设置设备信息
		if req.DeviceInfo != nil {
			deviceInfoBytes, err := json.Marshal(req.DeviceInfo)
			if err == nil {
				license.DeviceInfo = models.JSON(deviceInfoBytes)
			}
		}

		if err := tx.Create(license).Error; err != nil {
			return err
		}

		// 生成许可证文件
		licenseFile, err := s.generateLicenseFileContent(license, authCode)
		if err != nil {
			return err
		}

		response = &models.ActivateResponse{
			LicenseKey:        license.LicenseKey,
			LicenseFile:       licenseFile,
			HeartbeatInterval: 300,
			PublicKey:         s.productPublicKey(authCode),
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return nil, i18n.NewI18nError("300004", lang) // 激活数量已达上限
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return response, nil
}

// Heartbeat 心跳检测
func (s *licenseService) Heartbeat(ctx context.Context, req *models.HeartbeatRequest, clientIP string) (*models.HeartbeatResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 参数验证
	if req == nil {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 获取许可证信息
	license, err := s.licenseRepo.GetLicenseByKey(ctx, req.LicenseKey)
	if err != nil {
		if errors.Is(err, repository.ErrLicenseNotFound) {
			return nil, i18n.NewI18nError("300006", lang) // 许可证不存在
		}
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 检查许可证状态
	if license.Status == "revoked" {
		return nil, i18n.NewI18nError("300007", lang) // 许可证已被撤销
	}

	// 校验硬件指纹：许可证密钥可被复制，若不校验指纹，
	// 复制到另一台机器后心跳仍然通过，等于把许可证送出去。
	if req.HardwareFingerprint != "" && license.HardwareFingerprint != "" &&
		req.HardwareFingerprint != license.HardwareFingerprint {
		return nil, i18n.NewI18nError("300008", lang) // 硬件指纹不匹配
	}

	// 检查许可证是否过期：企业版心跳不感知过期与锁定，此处补上。
	// 客户端依赖心跳判定是否继续运行，不校验等于永久放行。
	if license.AuthorizationCode != nil {
		now := time.Now()
		if license.AuthorizationCode.IsLocked {
			return nil, i18n.NewI18nError("300003", lang) // 授权码已被锁定
		}
		_, periodEnd := resolveLicensePeriod(license, license.AuthorizationCode)
		if now.After(periodEnd) {
			return nil, i18n.NewI18nError("300016", lang) // 许可证已过期
		}
	}

	// 更新心跳时间和使用数据
	now := time.Now()
	license.LastHeartbeat = &now
	license.LastOnlineIP = &clientIP

	// 更新使用数据
	if req.UsageData != nil {
		usageDataBytes, err := json.Marshal(req.UsageData)
		if err == nil {
			license.UsageData = models.JSON(usageDataBytes)
		}
	}

	// 检查配置是否有更新
	configUpdated := false
	var licenseFile *string

	if req.ConfigUpdatedAt != nil && license.AuthorizationCode != nil {
		clientConfigTime, err := time.Parse(time.RFC3339, *req.ConfigUpdatedAt)
		if err == nil && license.AuthorizationCode.UpdatedAt.After(clientConfigTime) {
			configUpdated = true
			// 生成新的许可证文件
			fileContent, err := s.generateLicenseFileContent(license, license.AuthorizationCode)
			if err == nil {
				licenseFile = &fileContent
				license.ConfigUpdatedAt = &now
			}
		}
	}

	// 保存更新
	if err := s.licenseRepo.UpdateLicense(ctx, license); err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	response := &models.HeartbeatResponse{
		Status:            license.Status,
		ConfigUpdated:     configUpdated,
		LicenseFile:       licenseFile,
		HeartbeatInterval: 300,
	}

	return response, nil
}

// generateLicenseFileContent 生成许可证文件内容
func (s *licenseService) generateLicenseFileContent(license *models.License, authCode *models.AuthorizationCode) (string, error) {
	// 构建许可证文件内容
	licenseFileData := map[string]interface{}{
		"license_key":           license.LicenseKey,
		"authorization_code_id": license.AuthorizationCodeID,
		"hardware_fingerprint":  license.HardwareFingerprint,
		"status":                license.Status,
		"activated_at":          license.ActivatedAt,
		"config_updated_at":     license.ConfigUpdatedAt,
		"generated_at":          time.Now().Format(time.RFC3339),
	}

	if authCode != nil {
		licenseFileData["authorization_code"] = authCode.Code
		// 起止时间按有效期模式计算：从激活起算时不能沿用授权码的日期
		periodStart, periodEnd := resolveLicensePeriod(license, authCode)
		licenseFileData["start_date"] = periodStart
		licenseFileData["end_date"] = periodEnd
		licenseFileData["max_activations"] = authCode.MaxActivations
		// 部署类型随许可证下发：客户端据此决定校验单机型还是云端型
		licenseFileData["deployment_type"] = authCode.DeploymentType
		// 有效期模式一并下发，客户端才知道该以哪个日期为准
		licenseFileData["validity_mode"] = authCode.ValidityMode

		// 包含功能配置等
		if featureConfig := parseJSONField(authCode.FeatureConfig); len(featureConfig) > 0 {
			licenseFileData["feature_config"] = featureConfig
		}

		if usageLimits := parseJSONField(authCode.UsageLimits); len(usageLimits) > 0 {
			licenseFileData["usage_limits"] = usageLimits
		}

		if customParameters := parseJSONField(authCode.CustomParameters); len(customParameters) > 0 {
			licenseFileData["custom_parameters"] = customParameters
		}
	}

	// 序列化
	licenseJSON, err := json.Marshal(licenseFileData)
	if err != nil {
		return "", err
	}

	// 使用RSA数字签名
	encryptedData, err := s.signLicenseFile(licenseJSON, authCode)
	if err != nil {
		return "", err
	}

	return string(encryptedData), nil
}

// loadRSAPrivateKey 加载RSA私钥（懒加载）
func (s *licenseService) loadRSAPrivateKey() (*utils.RSAPrivateKey, error) {
	if s.rsaPrivateKey != nil {
		return s.rsaPrivateKey, nil
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not initialized")
	}

	privateKeyPath := cfg.License.RSA.PrivateKeyPath
	if privateKeyPath == "" {
		return nil, fmt.Errorf("RSA private key path not configured")
	}

	key, err := utils.LoadRSAPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载RSA私钥失败: %w", err)
	}

	s.rsaPrivateKey = key
	return key, nil
}

// GetLicenseSigningPublicKey 返回该许可证实际签发私钥所配对的公钥 PEM。
// 密钥选择逻辑与 signWithProductKey 保持镜像一致：
// 产品项配置了签名密钥对时返回产品项公钥，否则回退全局配置公钥文件。
// 客户端下载的授权 zip 必须携带此公钥，否则离线验签必然失败。
func (s *licenseService) GetLicenseSigningPublicKey(ctx context.Context, id string) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("license id is empty")
	}

	license, err := s.licenseRepo.GetLicenseByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询许可证失败: %w", err)
	}

	if s.productRepo != nil && license.AuthorizationCode != nil &&
		license.AuthorizationCode.ProductOfferingID != nil && *license.AuthorizationCode.ProductOfferingID != "" {
		offering, err := s.productRepo.GetProductOfferingByID(ctx, *license.AuthorizationCode.ProductOfferingID)
		if err == nil && offering.SigningPrivateKey != "" && offering.SigningPublicKey != "" {
			return []byte(offering.SigningPublicKey), nil
		}
		// 产品项存在但密钥不完整（升级前创建）时回退全局公钥，与签名兜底一致
	}

	cfg := config.GetConfig()
	if cfg == nil || cfg.License.RSA.PublicKeyPath == "" {
		return nil, fmt.Errorf("RSA public key path not configured")
	}
	return os.ReadFile(filepath.Clean(cfg.License.RSA.PublicKeyPath))
}

// signLicenseFile 使用RSA数字签名许可证文件
//
// 优先使用授权码所关联产品项的私钥：客户端只需内嵌对应产品的公钥，
// 无法伪造其他产品的许可证。未关联产品项时回退到全局密钥。
func (s *licenseService) signLicenseFile(data []byte, authCode *models.AuthorizationCode) ([]byte, error) {
	signature, err := s.signWithProductKey(data, authCode)
	if err != nil {
		return nil, err
	}

	// 构建签名后的许可证文件结构
	licenseWithSignature := map[string]interface{}{
		"data":      string(data), // 原始数据（JSON字符串）
		"signature": signature,    // 数字签名
		"algorithm": "RSA-PSS-SHA256",
	}

	// 序列化为JSON
	licenseJSON, err := json.Marshal(licenseWithSignature)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}

	// 使用base64编码
	encoded := base64.StdEncoding.EncodeToString(licenseJSON)

	return []byte(encoded), nil
}

// signWithProductKey 使用产品项私钥签名，缺失时回退全局密钥
func (s *licenseService) signWithProductKey(data []byte, authCode *models.AuthorizationCode) (string, error) {
	if s.productRepo != nil && authCode != nil && authCode.ProductOfferingID != nil && *authCode.ProductOfferingID != "" {
		offering, err := s.productRepo.GetProductOfferingByID(context.Background(), *authCode.ProductOfferingID)
		if err == nil && offering.SigningPrivateKey != "" {
			return utils.SignWithPrivateKeyPEM(offering.SigningPrivateKey, data)
		}
		// 产品项存在但无密钥（升级前创建）时回退，不阻断激活
	}

	// 兜底：全局文件密钥
	privateKey, err := s.loadRSAPrivateKey()
	if err != nil {
		return "", fmt.Errorf("加载RSA私钥失败: %w", err)
	}
	return privateKey.SignData(data)
}

// productPublicKey 返回授权码所属产品项的公钥，无关联时返回空串
func (s *licenseService) productPublicKey(authCode *models.AuthorizationCode) string {
	if s.productRepo == nil || authCode == nil ||
		authCode.ProductOfferingID == nil || *authCode.ProductOfferingID == "" {
		return ""
	}
	offering, err := s.productRepo.GetProductOfferingByID(context.Background(), *authCode.ProductOfferingID)
	if err != nil {
		return ""
	}
	return offering.SigningPublicKey
}

// parseJSONField 将数据库中的 JSON 字段解析为 map，兼容被字符串包裹的情况
func parseJSONField(raw models.JSON) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err == nil {
		return result
	}

	// 兼容字段以字符串形式保存 JSON 的情况，例如 "\"{\\\"key\\\":123}\""
	var rawString string
	if err := json.Unmarshal(raw, &rawString); err != nil {
		return nil
	}

	if err := json.Unmarshal([]byte(rawString), &result); err != nil {
		return nil
	}
	return result
}

// GetStatsOverview returns dashboard overview statistics
func (s *licenseService) GetStatsOverview(ctx context.Context) (*models.StatsOverviewResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// Load heartbeat timeout config
	cfg := config.GetConfig()
	offlineTimeout := time.Duration(cfg.License.HeartbeatTimeout) * time.Second

	now := time.Now()
	onlineThreshold := now.Add(-offlineTimeout)

	// Time boundaries
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	expire7d := now.AddDate(0, 0, 7)
	expire30d := now.AddDate(0, 0, 30)

	type authorizationStats struct {
		TotalAuthCodes     int64 `gorm:"column:total_auth_codes"`
		MonthNewAuthCodes  int64 `gorm:"column:month_new_auth_codes"`
		ExpiringIn7Days    int64 `gorm:"column:expiring_in_7days"`
		ExpiringIn30Days   int64 `gorm:"column:expiring_in_30days"`
		PreviousTotalCodes int64 `gorm:"column:previous_total_codes"`
	}
	type licenseStats struct {
		ActiveLicenses         int64 `gorm:"column:active_licenses"`
		TodayNewLicenses       int64 `gorm:"column:today_new_licenses"`
		YesterdayNewLicenses   int64 `gorm:"column:yesterday_new_licenses"`
		AbnormalAlerts         int64 `gorm:"column:abnormal_alerts"`
		PreviousActiveLicenses int64 `gorm:"column:previous_active_licenses"`
	}

	var authorizationData authorizationStats
	if err := s.db.WithContext(ctx).Model(&models.AuthorizationCode{}).Select(`
		COUNT(*) AS total_auth_codes,
		COALESCE(SUM(CASE WHEN created_at >= ? THEN 1 ELSE 0 END), 0) AS month_new_auth_codes,
		COALESCE(SUM(CASE WHEN end_date > ? AND end_date <= ? AND is_locked = false THEN 1 ELSE 0 END), 0) AS expiring_in_7days,
		COALESCE(SUM(CASE WHEN end_date > ? AND end_date <= ? AND is_locked = false THEN 1 ELSE 0 END), 0) AS expiring_in_30days,
		COALESCE(SUM(CASE WHEN created_at < ? THEN 1 ELSE 0 END), 0) AS previous_total_codes`,
		monthStart, now, expire7d, now, expire30d, monthStart).
		Scan(&authorizationData).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	var licenseData licenseStats
	if err := s.db.WithContext(ctx).Model(&models.License{}).Select(`
		COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active_licenses,
		COALESCE(SUM(CASE WHEN created_at >= ? THEN 1 ELSE 0 END), 0) AS today_new_licenses,
		COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN 1 ELSE 0 END), 0) AS yesterday_new_licenses,
		COALESCE(SUM(CASE WHEN status = 'active' AND ((last_heartbeat IS NOT NULL AND last_heartbeat < ?) OR (last_heartbeat IS NULL AND created_at < ?)) THEN 1 ELSE 0 END), 0) AS abnormal_alerts,
		COALESCE(SUM(CASE WHEN status = 'active' AND created_at < ? THEN 1 ELSE 0 END), 0) AS previous_active_licenses`,
		todayStart, yesterdayStart, todayStart, onlineThreshold, onlineThreshold, monthStart).
		Scan(&licenseData).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	stats := models.StatsOverviewResponse{
		TotalAuthCodes:       authorizationData.TotalAuthCodes,
		ActiveLicenses:       licenseData.ActiveLicenses,
		TodayNewLicenses:     licenseData.TodayNewLicenses,
		YesterdayNewLicenses: licenseData.YesterdayNewLicenses,
		MonthNewAuthCodes:    authorizationData.MonthNewAuthCodes,
		ExpiringIn7Days:      authorizationData.ExpiringIn7Days,
		ExpiringIn30Days:     authorizationData.ExpiringIn30Days,
		AbnormalAlerts:       licenseData.AbnormalAlerts,
	}
	if authorizationData.PreviousTotalCodes > 0 {
		stats.GrowthRate.AuthCodesMoM = float64(stats.TotalAuthCodes-authorizationData.PreviousTotalCodes) / float64(authorizationData.PreviousTotalCodes) * 100
	} else if stats.TotalAuthCodes > 0 {
		stats.GrowthRate.AuthCodesMoM = 100
	}
	if licenseData.PreviousActiveLicenses > 0 {
		stats.GrowthRate.LicensesMoM = float64(stats.ActiveLicenses-licenseData.PreviousActiveLicenses) / float64(licenseData.PreviousActiveLicenses) * 100
	} else if stats.ActiveLicenses > 0 {
		stats.GrowthRate.LicensesMoM = 100
	}

	return &stats, nil
}
