package database

import (
	"context"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"

	"license-manager/internal/models"
)

// seedSystemConfigs 默认系统配置。
//
// 键名与公开接口的契约对齐：公开接口把这些键直接透传给前端，
// 前端据此渲染登录页，不需要再维护一份映射。
var seedSystemConfigs = []models.SystemConfig{
	{
		ConfigKey:   "system.site_name",
		Value:       "软件授权管理系统",
		ValueType:   models.ConfigTypeText,
		Group:       models.ConfigGroupSystem,
		Public:      true,
		Description: "站点名称，显示在登录页与浏览器标题",
		SortOrder:   1,
	},
	{
		ConfigKey:   "system.logo_url",
		Value:       "",
		ValueType:   models.ConfigTypeImage,
		Group:       models.ConfigGroupSystem,
		Public:      true,
		Description: "站点 Logo，显示在登录页左上角",
		SortOrder:   2,
	},
	{
		ConfigKey:   "system.favicon_url",
		Value:       "",
		ValueType:   models.ConfigTypeImage,
		Group:       models.ConfigGroupSystem,
		Public:      true,
		Description: "站点图标，显示在浏览器标签页",
		SortOrder:   3,
	},
	{
		ConfigKey:   "system.show_copyright",
		Value:       "false",
		ValueType:   models.ConfigTypeSwitch,
		Group:       models.ConfigGroupSystem,
		Public:      true,
		Description: "是否在侧边栏底部显示版权信息",
		SortOrder:   4,
	},
	{
		ConfigKey:   "auth.show_username_password",
		Value:       "true",
		ValueType:   models.ConfigTypeSwitch,
		Group:       models.ConfigGroupAuth,
		Public:      true,
		Description: "是否显示用户名密码登录方式",
		SortOrder:   10,
	},
	{
		ConfigKey:   "auth.session_timeout_hours",
		Value:       "24",
		ValueType:   models.ConfigTypeNumber,
		Group:       models.ConfigGroupAuth,
		Description: "登录会话有效期（小时）",
		SortOrder:   11,
	},
	{
		ConfigKey:   "auth.max_login_attempts",
		Value:       "5",
		ValueType:   models.ConfigTypeNumber,
		Group:       models.ConfigGroupAuth,
		Description: "连续登录失败达到该次数后锁定账号",
		SortOrder:   12,
	},
	{
		ConfigKey:   "auth.lock_duration_minutes",
		Value:       "30",
		ValueType:   models.ConfigTypeNumber,
		Group:       models.ConfigGroupAuth,
		Description: "账号锁定时长（分钟）",
		SortOrder:   13,
	},
	{
		ConfigKey:   "mail.enabled",
		Value:       "false",
		ValueType:   models.ConfigTypeSwitch,
		Group:       models.ConfigGroupMail,
		Description: "是否启用邮件发送（忘记密码等功能依赖此项）",
		SortOrder:   20,
	},
	{
		ConfigKey:   "mail.host",
		Value:       "",
		ValueType:   models.ConfigTypeText,
		Group:       models.ConfigGroupMail,
		Description: "SMTP 服务器地址",
		SortOrder:   21,
	},
	{
		ConfigKey:   "mail.port",
		Value:       "465",
		ValueType:   models.ConfigTypeNumber,
		Group:       models.ConfigGroupMail,
		Description: "SMTP 端口",
		SortOrder:   22,
	},
	{
		ConfigKey:   "mail.username",
		Value:       "",
		ValueType:   models.ConfigTypeText,
		Group:       models.ConfigGroupMail,
		Description: "SMTP 账号",
		SortOrder:   23,
	},
	{
		ConfigKey:   "mail.from",
		Value:       "",
		ValueType:   models.ConfigTypeText,
		Group:       models.ConfigGroupMail,
		Description: "发件人地址",
		SortOrder:   24,
	},
}

// initSystemConfigs 初始化默认系统配置（幂等）。
//
// 已存在的键不覆盖值：运营在界面上改过的配置不该被重启冲掉。
// 仅在缺失时补齐，以及为新增的键补默认值。
func initSystemConfigs() error {
	ctx := context.Background()

	for i := range seedSystemConfigs {
		config := seedSystemConfigs[i]

		var existing models.SystemConfig
		err := DB.WithContext(ctx).Where("config_key = ?", config.ConfigKey).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check system config %s: %w", config.ConfigKey, err)
		}

		if err := DB.WithContext(ctx).Create(&config).Error; err != nil {
			return fmt.Errorf("failed to create system config %s: %w", config.ConfigKey, err)
		}
		log.Printf("Seeded system config: %s", config.ConfigKey)
	}

	return nil
}

// dropLegacyAPIKeyPrefixColumn 清理早期版本遗留的 key_prefix 列。
//
// 模型标注了 column:api_key_prefix，AutoMigrate 因此新建了该列，
// 而旧迁移留下的 key_prefix 因 NOT NULL 无默认值会让插入失败。
// 该列从未被写入过有效数据，直接删除即可。
func dropLegacyAPIKeyPrefixColumn() error {
	migrator := DB.Migrator()

	if !migrator.HasColumn(&models.APIKey{}, "key_prefix") {
		return nil
	}
	if !migrator.HasColumn(&models.APIKey{}, "api_key_prefix") {
		return nil
	}

	if err := migrator.DropColumn(&models.APIKey{}, "key_prefix"); err != nil {
		return fmt.Errorf("failed to drop api_keys.key_prefix: %w", err)
	}
	log.Println("Dropped legacy api_keys.key_prefix column")
	return nil
}
