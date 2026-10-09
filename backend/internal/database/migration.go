package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"license-manager/internal/config"
	"license-manager/internal/models"
	"license-manager/pkg/utils"

	"gorm.io/gorm"
)

// AutoMigrate 执行自动数据库迁移
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	log.Println("Starting database auto migration...")

	// 迁移数据库表结构
	err := DB.AutoMigrate(
		&models.Customer{},
		&models.CustomerCodeSequence{},
		&models.User{},
		&models.ProductOffering{},
		&models.AuthorizationCode{},
		&models.License{},
		&models.AuthorizationChange{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := removeLegacyAuthorizationCodeTypeFields(); err != nil {
		return fmt.Errorf("failed to remove legacy authorization code type fields: %w", err)
	}

	// 初始化客户编码序列
	if err := initCustomerCodeSequence(); err != nil {
		return fmt.Errorf("failed to initialize customer code sequence: %w", err)
	}

	// 初始化默认管理员用户
	if err := initDefaultAdminUser(); err != nil {
		return fmt.Errorf("failed to initialize default admin user: %w", err)
	}

	log.Println("Database auto migration completed successfully")
	return nil
}

// removeLegacyAuthorizationCodeTypeFields 清理社区版不再使用的历史字段。
func removeLegacyAuthorizationCodeTypeFields() error {
	migrator := DB.Migrator()
	for _, indexName := range []string{
		"idx_authorization_codes_software_status",
		"idx_authorization_codes_deployment_type",
	} {
		if migrator.HasIndex(&models.AuthorizationCode{}, indexName) {
			if err := migrator.DropIndex(&models.AuthorizationCode{}, indexName); err != nil {
				return err
			}
		}
	}

	for _, columnName := range []string{"deployment_type", "encryption_type"} {
		if migrator.HasColumn(&models.AuthorizationCode{}, columnName) {
			if err := migrator.DropColumn(&models.AuthorizationCode{}, columnName); err != nil {
				return err
			}
		}
	}

	return nil
}

// initCustomerCodeSequence 初始化客户编码序列
func initCustomerCodeSequence() error {
	currentYear := time.Now().Year()

	var sequence models.CustomerCodeSequence
	result := DB.Where("year = ?", currentYear).First(&sequence)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		values := map[string]interface{}{
			"year":            currentYear,
			"sequence_number": 0,
		}

		// 兼容历史测试库保留的 tenant_id 必填列；社区版新库不会新增该列。
		if DB.Migrator().HasColumn(&models.CustomerCodeSequence{}, "tenant_id") {
			var existing struct {
				TenantID string `gorm:"column:tenant_id"`
			}
			if err := DB.Table(models.CustomerCodeSequence{}.TableName()).
				Select("tenant_id").
				Where("tenant_id IS NOT NULL").
				Limit(1).
				Scan(&existing).Error; err != nil {
				return fmt.Errorf("failed to read legacy tenant id: %w", err)
			}
			values["tenant_id"] = existing.TenantID
		}

		if err := DB.Table(models.CustomerCodeSequence{}.TableName()).Create(values).Error; err != nil {
			return fmt.Errorf("failed to create customer code sequence: %w", err)
		}
		log.Printf("Initialized customer code sequence for year %d", currentYear)
	} else if result.Error != nil {
		return fmt.Errorf("failed to query customer code sequence: %w", result.Error)
	}

	return nil
}

// RunMigrationIfEnabled 根据配置决定是否运行迁移
func RunMigrationIfEnabled() error {
	cfg := config.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	if cfg.Database.AutoMigrate {
		return AutoMigrate()
	}

	log.Println("Auto migration is disabled in config")
	return nil
}

// initDefaultAdminUser 初始化默认管理员用户
func initDefaultAdminUser() error {
	var user models.User
	result := DB.Where("username = ?", "admin").First(&user)

	if result.Error != nil {
		// 用户不存在，创建默认管理员
		passwordHash, err := utils.HashPassword("admin@123")
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		user = models.User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: passwordHash,
			FullName:     "系统管理员",
			Role:         "admin",
			Status:       "active",
		}

		if err := DB.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create default admin user: %w", err)
		}

		log.Println("Created default admin user: admin / admin@123")
	} else {
		log.Println("Default admin user already exists")
	}

	return nil
}
