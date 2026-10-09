package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"gorm.io/gorm"

	"license-manager/internal/models"
)

// apiKeyHexLength API Key 随机部分长度（十六进制字符数）。
//
// 32 个十六进制字符即 128 位熵，配合前缀足够长；
// 更长只会让请求头膨胀而无实际收益。
const apiKeyHexLength = 32

type systemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建系统配置数据访问实例
func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

// GetByKey 按键获取配置
func (r *systemConfigRepository) GetByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	if err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSystemConfigNotFound
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return &config, nil
}

// GetAll 获取全部配置（按分组、排序）
func (r *systemConfigRepository) GetAll(ctx context.Context) ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	if err := r.db.WithContext(ctx).
		Order("`group` ASC, sort_order ASC, config_key ASC").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}
	return configs, nil
}

// GetPublic 获取可公开的配置
func (r *systemConfigRepository) GetPublic(ctx context.Context) ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	if err := r.db.WithContext(ctx).
		Where("public = ?", true).
		Order("sort_order ASC, config_key ASC").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list public configs: %w", err)
	}
	return configs, nil
}

// Create 创建配置
func (r *systemConfigRepository) Create(ctx context.Context, config *models.SystemConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrSystemConfigDuplicate
		}
		return fmt.Errorf("failed to create config: %w", err)
	}
	return nil
}

// Update 更新配置
func (r *systemConfigRepository) Update(ctx context.Context, config *models.SystemConfig) error {
	result := r.db.WithContext(ctx).Model(&models.SystemConfig{}).
		Where("config_key = ?", config.ConfigKey).
		Select("value", "description", "sort_order", "public", "updated_at").
		Updates(config)
	if result.Error != nil {
		return fmt.Errorf("failed to update config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSystemConfigNotFound
	}
	return nil
}

// UpdateValueOnly 仅更新值。
//
// 用于批量保存：键已存在，不应改动类型与分组等结构属性。
func (r *systemConfigRepository) UpdateValueOnly(ctx context.Context, key, value string) error {
	result := r.db.WithContext(ctx).Model(&models.SystemConfig{}).
		Where("config_key = ?", key).
		Updates(map[string]interface{}{"value": value, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("failed to update config value: %w", result.Error)
	}
	return nil
}

// Delete 删除配置
func (r *systemConfigRepository) Delete(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).
		Where("config_key = ?", key).Delete(&models.SystemConfig{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSystemConfigNotFound
	}
	return nil
}

// GetAPIKeyList 查询 API Key 列表
func (r *systemConfigRepository) GetAPIKeyList(ctx context.Context, req *models.APIKeyListRequest) ([]models.APIKeyListItem, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	type row struct {
		ID         string
		Name       string
		KeyPrefix  string
		UserID     string
		UserName   string
		Status     string
		ExpiresAt  *time.Time
		AllowedIPs string
		LastUsedAt *time.Time
		CreatedAt  time.Time
		CreatedBy  string
	}

	query := r.db.WithContext(ctx).Table("api_keys k").
		Select(`k.id, k.name, k.api_key_prefix AS key_prefix, k.user_id,
			COALESCE(u.username, '') AS user_name, k.status, k.expires_at,
			k.allowed_ips, k.last_used_at, k.created_at, k.created_by`).
		Joins("LEFT JOIN users u ON u.id = k.user_id")

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("k.name LIKE ? OR k.api_key_prefix LIKE ?", like, like)
	}
	if req.Status != "" {
		query = query.Where("k.status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count api keys: %w", err)
	}

	var rows []row
	if err := query.Order("k.created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list api keys: %w", err)
	}

	items := make([]models.APIKeyListItem, 0, len(rows))
	now := time.Now()
	for _, item := range rows {
		status := item.Status
		// 过期是时间的函数，不依赖定时任务去改状态字段
		if status == models.APIKeyStatusActive && item.ExpiresAt != nil && item.ExpiresAt.Before(now) {
			status = models.APIKeyStatusExpired
		}
		items = append(items, models.APIKeyListItem{
			ID: item.ID, Name: item.Name, KeyPrefix: item.KeyPrefix,
			UserID: item.UserID, UserName: item.UserName, Status: status,
			ExpiresAt:  item.ExpiresAt,
			AllowedIPs: splitAndTrim(item.AllowedIPs),
			LastUsedAt: item.LastUsedAt,
			CreatedAt:  item.CreatedAt, CreatedBy: item.CreatedBy,
		})
	}

	return items, total, nil
}

// GetAPIKeyByID 按ID获取 API Key
func (r *systemConfigRepository) GetAPIKeyByID(ctx context.Context, id string) (*models.APIKey, error) {
	var key models.APIKey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}
	return &key, nil
}

// CreateAPIKey 创建 API Key
func (r *systemConfigRepository) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	if err := r.db.WithContext(ctx).Create(key).Error; err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}
	return nil
}

// UpdateAPIKey 更新 API Key
func (r *systemConfigRepository) UpdateAPIKey(ctx context.Context, key *models.APIKey) error {
	result := r.db.WithContext(ctx).Model(&models.APIKey{}).
		Where("id = ?", key.ID).
		Select("name", "status", "expires_at", "allowed_ips", "updated_at").
		Updates(key)
	if result.Error != nil {
		return fmt.Errorf("failed to update api key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// DeleteAPIKey 删除 API Key
func (r *systemConfigRepository) DeleteAPIKey(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.APIKey{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete api key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// GetAPIKeyByHash 按哈希查找 API Key
func (r *systemConfigRepository) GetAPIKeyByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	var key models.APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("failed to get api key by hash: %w", err)
	}
	return &key, nil
}

// TouchAPIKeyUsage 记录最近使用时间
func (r *systemConfigRepository) TouchAPIKeyUsage(ctx context.Context, id string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.APIKey{}).
		Where("id = ?", id).Update("last_used_at", now).Error; err != nil {
		return fmt.Errorf("failed to touch api key usage: %w", err)
	}
	return nil
}

// GenerateAPIKey 生成 API Key 明文与哈希。
//
// 返回明文仅此一次，数据库只存哈希与前缀。
func GenerateAPIKey() (plaintext string, prefix string, hash string, err error) {
	buf := make([]byte, apiKeyHexLength/2)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random: %w", err)
	}

	random := hex.EncodeToString(buf)
	plaintext = models.APIKeyPrefix + random
	// 前缀取前 8 位十六进制，列表里足以辨识且不泄露全量
	prefix = plaintext[:len(models.APIKeyPrefix)+8]

	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])

	return plaintext, prefix, hash, nil
}

// IPWhitelistMatches 判断 IP 是否在白名单内。
//
// 白名单为空表示不限制。支持 CIDR 写法，便于限定网段。
func IPWhitelistMatches(allowedIPs string, ip string) bool {
	allowed := splitAndTrim(allowedIPs)
	if len(allowed) == 0 {
		return true
	}
	for _, entry := range allowed {
		if entry == ip {
			return true
		}
		if strings.Contains(entry, "/") {
			if cidrContains(entry, ip) {
				return true
			}
		}
	}
	return false
}

// cidrContains 判断 IP 是否落在 CIDR 网段内
func cidrContains(cidr, ip string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	addr := net.ParseIP(ip)
	if addr == nil {
		return false
	}
	return network.Contains(addr)
}

// splitAndTrim 逗号分隔字符串转切片
func splitAndTrim(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// VerifyAPIKey 校验 API Key。
//
// 返回所属用户ID与是否放行。校验不通过时不返回错误：
// 密钥无效是调用方的常态，不是服务端故障，两者不该混为一谈。
func (r *systemConfigRepository) VerifyAPIKey(ctx context.Context, plaintext, hash, clientIP string) (string, bool, error) {
	key, err := r.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if err == ErrAPIKeyNotFound {
			return "", false, nil
		}
		return "", false, err
	}

	if key.Status != models.APIKeyStatusActive {
		return "", false, nil
	}

	// 过期每次现算：状态字段可能仍是 active 而时间已过
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return "", false, nil
	}

	if !IPWhitelistMatches(key.AllowedIPs, clientIP) {
		return "", false, nil
	}

	// 记录使用时间；失败不影响本次请求的成败
	_ = r.TouchAPIKeyUsage(ctx, key.ID)

	return key.UserID, true, nil
}
