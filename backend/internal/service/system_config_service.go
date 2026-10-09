package service

import (
	"context"
	"errors"
	"strings"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
)

type systemConfigService struct {
	configRepo repository.SystemConfigRepository
	userRepo   repository.UserRepository
}

// NewSystemConfigService 创建系统配置服务实例
func NewSystemConfigService(
	configRepo repository.SystemConfigRepository,
	userRepo repository.UserRepository,
) SystemConfigService {
	return &systemConfigService{configRepo: configRepo, userRepo: userRepo}
}

// GetGroupedConfigs 获取按分组组织的系统配置
func (s *systemConfigService) GetGroupedConfigs(ctx context.Context) (*models.SystemConfigGroupResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	configs, err := s.configRepo.GetAll(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	// 按分组聚合，保持数据库给定的排序
	order := []string{models.ConfigGroupSystem, models.ConfigGroupAuth, models.ConfigGroupMail}
	grouped := map[string][]models.SystemConfig{}
	for _, config := range configs {
		grouped[config.Group] = append(grouped[config.Group], config)
	}

	groups := make([]models.SystemConfigGroup, 0, len(grouped))
	// 先按预置分组顺序输出，避免依赖 map 迭代顺序
	for _, name := range order {
		if items, ok := grouped[name]; ok {
			groups = append(groups, models.SystemConfigGroup{Group: name, Configs: items})
			delete(grouped, name)
		}
	}
	// 兜底：分组常量未覆盖的新分组
	for name, items := range grouped {
		groups = append(groups, models.SystemConfigGroup{Group: name, Configs: items})
	}

	return &models.SystemConfigGroupResponse{Groups: groups}, nil
}

// GetPublicConfigs 获取公开配置。
//
// 仅返回 public=true 的项：邮件服务器这类信息不应暴露给未登录访问者。
func (s *systemConfigService) GetPublicConfigs(ctx context.Context) (*models.PublicConfigResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	configs, err := s.configRepo.GetPublic(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	result := make(map[string]string, len(configs))
	for _, config := range configs {
		result[config.ConfigKey] = config.Value
	}

	return &models.PublicConfigResponse{Configs: result}, nil
}

// GetConfigByKey 按键获取配置
func (s *systemConfigService) GetConfigByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	config, err := s.configRepo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, repository.ErrSystemConfigNotFound) {
			return nil, i18n.NewI18nError("800001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}
	return config, nil
}

// CreateConfig 创建配置
func (s *systemConfigService) CreateConfig(ctx context.Context, req *models.SystemConfigCreateRequest) (*models.SystemConfig, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	key := strings.TrimSpace(req.ConfigKey)
	if key == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	valueType := req.ValueType
	if valueType == "" {
		valueType = models.ConfigTypeText
	}
	group := req.Group
	if group == "" {
		group = models.ConfigGroupSystem
	}

	public := false
	if req.Public != nil {
		public = *req.Public
	}

	config := &models.SystemConfig{
		ConfigKey:   key,
		Value:       req.Value,
		ValueType:   valueType,
		Group:       group,
		Options:     req.Options,
		Public:      public,
		Description: strings.TrimSpace(req.Description),
		SortOrder:   req.SortOrder,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {
		if errors.Is(err, repository.ErrSystemConfigDuplicate) {
			return nil, i18n.NewI18nError("800002", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	return config, nil
}

// UpdateConfig 更新配置
func (s *systemConfigService) UpdateConfig(ctx context.Context, key string, req *models.SystemConfigUpdateRequest) (*models.SystemConfig, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	config, err := s.configRepo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, repository.ErrSystemConfigNotFound) {
			return nil, i18n.NewI18nError("800001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	if req.Value != nil {
		config.Value = *req.Value
	}
	if req.Description != nil {
		config.Description = strings.TrimSpace(*req.Description)
	}
	if req.SortOrder != nil {
		config.SortOrder = *req.SortOrder
	}
	if req.Public != nil {
		config.Public = *req.Public
	}

	if err := s.configRepo.Update(ctx, config); err != nil {
		if errors.Is(err, repository.ErrSystemConfigNotFound) {
			return nil, i18n.NewI18nError("800001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	return config, nil
}

// BatchUpdateConfigs 批量更新配置值
func (s *systemConfigService) BatchUpdateConfigs(ctx context.Context, req *models.SystemConfigBatchUpdateRequest) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 逐项存在性检查：静默跳过不存在的键会让调用方以为保存成功
	for _, item := range req.Items {
		if _, err := s.configRepo.GetByKey(ctx, strings.TrimSpace(item.ConfigKey)); err != nil {
			if errors.Is(err, repository.ErrSystemConfigNotFound) {
				return i18n.NewI18nError("800001", lang)
			}
			return i18n.NewI18nError("900005", lang)
		}
	}

	for _, item := range req.Items {
		if err := s.configRepo.UpdateValueOnly(ctx, strings.TrimSpace(item.ConfigKey), item.Value); err != nil {
			return i18n.NewI18nError("900005", lang)
		}
	}

	return nil
}

// DeleteConfig 删除配置
func (s *systemConfigService) DeleteConfig(ctx context.Context, key string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.configRepo.Delete(ctx, key); err != nil {
		if errors.Is(err, repository.ErrSystemConfigNotFound) {
			return i18n.NewI18nError("800001", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}
	return nil
}

// GetAPIKeyList 查询 API Key 列表
func (s *systemConfigService) GetAPIKeyList(ctx context.Context, req *models.APIKeyListRequest) (*models.APIKeyListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if req == nil {
		req = &models.APIKeyListRequest{}
	}

	items, total, err := s.configRepo.GetAPIKeyList(ctx, req)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	if items == nil {
		items = []models.APIKeyListItem{}
	}

	return &models.APIKeyListResponse{List: items, Total: total}, nil
}

// CreateAPIKey 创建 API Key
func (s *systemConfigService) CreateAPIKey(ctx context.Context, req *models.APIKeyCreateRequest, createdBy string) (*models.APIKeyCreateResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 确认用户存在：密钥必须归属于某个账号，否则权限无从判断
	if _, err := s.userRepo.GetUserByID(ctx, req.UserID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, i18n.NewI18nError("600013", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	plaintext, prefix, hash, err := repository.GenerateAPIKey()
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	key := &models.APIKey{
		Name:       strings.TrimSpace(req.Name),
		KeyHash:    hash,
		KeyPrefix:  prefix,
		UserID:     req.UserID,
		Status:     models.APIKeyStatusActive,
		ExpiresAt:  req.ExpiresAt,
		AllowedIPs: strings.Join(req.AllowedIPs, ","),
		CreatedBy:  createdBy,
	}

	if err := s.configRepo.CreateAPIKey(ctx, key); err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	// 明文仅此一次返回，数据库只存哈希
	return &models.APIKeyCreateResponse{APIKey: key, Plaintext: plaintext}, nil
}

// UpdateAPIKey 更新 API Key
func (s *systemConfigService) UpdateAPIKey(ctx context.Context, id string, req *models.APIKeyUpdateRequest) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	key, err := s.configRepo.GetAPIKeyByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAPIKeyNotFound) {
			return i18n.NewI18nError("800003", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}

	if req.Name != nil && *req.Name != "" {
		key.Name = strings.TrimSpace(*req.Name)
	}
	if req.Status != nil {
		key.Status = *req.Status
	}
	if req.ExpiresAt != nil {
		key.ExpiresAt = req.ExpiresAt
	}
	if req.AllowedIPs != nil {
		key.AllowedIPs = strings.Join(req.AllowedIPs, ",")
	}

	if err := s.configRepo.UpdateAPIKey(ctx, key); err != nil {
		if errors.Is(err, repository.ErrAPIKeyNotFound) {
			return i18n.NewI18nError("800003", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}
	return nil
}

// DeleteAPIKey 删除 API Key
func (s *systemConfigService) DeleteAPIKey(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.configRepo.DeleteAPIKey(ctx, id); err != nil {
		if errors.Is(err, repository.ErrAPIKeyNotFound) {
			return i18n.NewI18nError("800003", lang)
		}
		return i18n.NewI18nError("900005", lang)
	}
	return nil
}

// GetConfigValue 读取单个配置值，供其他服务复用
func (s *systemConfigService) GetConfigValue(ctx context.Context, key string) string {
	config, err := s.configRepo.GetByKey(ctx, key)
	if err != nil {
		return ""
	}
	return config.Value
}
