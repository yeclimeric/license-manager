package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"license-manager/internal/config"
	"license-manager/internal/models"
)

type licenseRepository struct {
	db *gorm.DB
}

// NewLicenseRepository 创建许可证数据访问实例
func NewLicenseRepository(db *gorm.DB) LicenseRepository {
	return &licenseRepository{
		db: db,
	}
}

// GetLicenseList 查询许可证列表
func (r *licenseRepository) GetLicenseList(ctx context.Context, req *models.LicenseListRequest) (*models.LicenseListResponse, error) {
	// 设置默认值
	req.Page, req.PageSize = NormalizePage(req.Page, req.PageSize)
	if req.Sort == "" {
		req.Sort = "created_at"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	// 构建查询
	query := r.db.Model(&models.License{}).
		Select(`licenses.id, licenses.license_key, licenses.authorization_code_id, 
				authorization_codes.code as authorization_code, customers.customer_name,
				licenses.hardware_fingerprint, licenses.status, licenses.activation_ip,
				licenses.last_online_ip, licenses.activated_at, licenses.last_heartbeat`).
		Joins("LEFT JOIN authorization_codes ON licenses.authorization_code_id = authorization_codes.id").
		Joins("LEFT JOIN customers ON licenses.customer_id = customers.id")

	// 授权码ID筛选
	if req.AuthorizationCodeID != "" {
		query = query.Where("licenses.authorization_code_id = ?", req.AuthorizationCodeID)
	}

	// 客户ID筛选
	if req.CustomerID != "" {
		query = query.Where("licenses.customer_id = ?", req.CustomerID)
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where("licenses.status = ?", req.Status)
	}

	// 在线状态筛选
	if req.IsOnline != nil && *req.IsOnline != "" {
		cfg := config.GetConfig()
		heartbeatTimeoutSeconds := cfg.License.HeartbeatTimeout
		if heartbeatTimeoutSeconds <= 0 {
			heartbeatTimeoutSeconds = 300 // 默认5分钟
		}
		onlineThreshold := time.Now().Add(-time.Duration(heartbeatTimeoutSeconds) * time.Second)

		if *req.IsOnline == "true" {
			query = query.Where("licenses.last_heartbeat >= ?", onlineThreshold)
		} else if *req.IsOnline == "false" {
			query = query.Where("licenses.last_heartbeat < ? OR licenses.last_heartbeat IS NULL", onlineThreshold)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 排序和分页：白名单收敛，避免请求字段直接拼进 SQL
	orderBy := SafeOrderClause(req.Sort, req.Order, "licenses.",
		"created_at", "updated_at", "activated_at", "last_heartbeat")
	offset := (req.Page - 1) * req.PageSize

	var licenses []struct {
		ID                  string     `json:"id"`
		LicenseKey          string     `json:"license_key"`
		AuthorizationCodeID string     `json:"authorization_code_id"`
		AuthorizationCode   string     `json:"authorization_code"`
		CustomerName        string     `json:"customer_name"`
		HardwareFingerprint string     `json:"hardware_fingerprint"`
		Status              string     `json:"status"`
		ActivationIP        *string    `json:"activation_ip"`
		LastOnlineIP        *string    `json:"last_online_ip"`
		ActivatedAt         *time.Time `json:"activated_at"`
		LastHeartbeat       *time.Time `json:"last_heartbeat"`
	}

	if err := query.Order(orderBy).Limit(req.PageSize).Offset(offset).Scan(&licenses).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式
	list := make([]models.LicenseListItem, len(licenses))

	cfg := config.GetConfig()
	heartbeatTimeoutSeconds := cfg.License.HeartbeatTimeout
	if heartbeatTimeoutSeconds <= 0 {
		heartbeatTimeoutSeconds = 300 // 默认5分钟
	}
	onlineThreshold := time.Now().Add(-time.Duration(heartbeatTimeoutSeconds) * time.Second)

	for i, license := range licenses {
		// 计算在线状态
		isOnline := false
		if license.LastHeartbeat != nil {
			isOnline = license.LastHeartbeat.After(onlineThreshold)
		}

		// 格式化时间
		var activatedAtStr, lastHeartbeatStr *string
		if license.ActivatedAt != nil {
			str := license.ActivatedAt.Format(time.RFC3339)
			activatedAtStr = &str
		}
		if license.LastHeartbeat != nil {
			str := license.LastHeartbeat.Format(time.RFC3339)
			lastHeartbeatStr = &str
		}

		list[i] = models.LicenseListItem{
			ID:                  license.ID,
			LicenseKey:          license.LicenseKey,
			AuthorizationCodeID: license.AuthorizationCodeID,
			AuthorizationCode:   license.AuthorizationCode,
			CustomerName:        license.CustomerName,
			HardwareFingerprint: license.HardwareFingerprint,
			Status:              license.Status,
			IsOnline:            isOnline,
			ActivationIP:        license.ActivationIP,
			LastOnlineIP:        license.LastOnlineIP,
			ActivatedAt:         activatedAtStr,
			LastHeartbeat:       lastHeartbeatStr,
		}
	}

	// 计算总页数（统一契约，含除零保护）
	totalPages := CalcTotalPages(total, req.PageSize)

	return &models.LicenseListResponse{
		List:       list,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetLicenseByID 根据ID获取许可证信息
func (r *licenseRepository) GetLicenseByID(ctx context.Context, id string) (*models.License, error) {
	var license models.License

	err := r.db.Preload("AuthorizationCode").Preload("Customer").
		Where("id = ?", id).First(&license).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}

	// 计算在线状态
	if license.LastHeartbeat != nil {
		cfg := config.GetConfig()
		heartbeatTimeoutSeconds := cfg.License.HeartbeatTimeout
		if heartbeatTimeoutSeconds <= 0 {
			heartbeatTimeoutSeconds = 300 // 默认5分钟
		}
		onlineThreshold := time.Now().Add(-time.Duration(heartbeatTimeoutSeconds) * time.Second)
		license.IsOnline = license.LastHeartbeat.After(onlineThreshold)
	}

	return &license, nil
}

// CreateLicense 创建许可证
func (r *licenseRepository) CreateLicense(ctx context.Context, license *models.License) error {
	return r.db.Create(license).Error
}

// UpdateLicense 更新许可证信息
func (r *licenseRepository) UpdateLicense(ctx context.Context, license *models.License) error {
	return r.db.WithContext(ctx).Save(license).Error
}

// DeleteLicensePermanently 物理删除许可证，释放授权码与硬件指纹唯一约束。
func (r *licenseRepository) DeleteLicensePermanently(ctx context.Context, license *models.License) error {
	return r.db.WithContext(ctx).Unscoped().Delete(license).Error
}

// CheckAuthorizationCodeExists 检查授权码是否存在
func (r *licenseRepository) CheckAuthorizationCodeExists(ctx context.Context, authCodeID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.AuthorizationCode{}).Where("id = ?", authCodeID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAuthorizationCodeByID 根据ID获取授权码信息
func (r *licenseRepository) GetAuthorizationCodeByID(ctx context.Context, authCodeID string) (*models.AuthorizationCode, error) {
	var authCode models.AuthorizationCode
	err := r.db.Where("id = ?", authCodeID).First(&authCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAuthorizationCodeNotFound
		}
		return nil, err
	}
	return &authCode, nil
}

// GetAuthorizationCodeByCode 根据授权码获取授权码信息
func (r *licenseRepository) GetAuthorizationCodeByCode(ctx context.Context, code string) (*models.AuthorizationCode, error) {
	var authCode models.AuthorizationCode
	err := r.db.Where("code = ?", code).First(&authCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAuthorizationCodeNotFound
		}
		return nil, err
	}
	return &authCode, nil
}

// GetLicenseByKey 根据许可证密钥获取许可证信息
func (r *licenseRepository) GetLicenseByKey(ctx context.Context, licenseKey string) (*models.License, error) {
	var license models.License

	err := r.db.Preload("AuthorizationCode").Preload("Customer").
		Where("license_key = ?", licenseKey).First(&license).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}

	// 计算在线状态
	if license.LastHeartbeat != nil {
		cfg := config.GetConfig()
		heartbeatTimeoutSeconds := cfg.License.HeartbeatTimeout
		if heartbeatTimeoutSeconds <= 0 {
			heartbeatTimeoutSeconds = 300 // 默认5分钟
		}
		onlineThreshold := time.Now().Add(-time.Duration(heartbeatTimeoutSeconds) * time.Second)
		license.IsOnline = license.LastHeartbeat.After(onlineThreshold)
	}

	return &license, nil
}

// GetActiveLicenseCount 获取指定授权码的激活许可证数量
func (r *licenseRepository) GetActiveLicenseCount(ctx context.Context, authCodeID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.License{}).
		Where("authorization_code_id = ? AND status = ?", authCodeID, "active").
		Count(&count).Error
	return count, err
}

// CheckAuthorizationCodeHasLicenses 检查授权码是否有关联的许可证
func (r *licenseRepository) CheckAuthorizationCodeHasLicenses(ctx context.Context, authCodeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.License{}).
		Where("authorization_code_id = ?", authCodeID).
		Count(&count).Error
	return count > 0, err
}
