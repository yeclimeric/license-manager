package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"license-manager/internal/models"
	"license-manager/pkg/utils"
)

// identifierGenerateRetries 标识符冲突时的重试次数
const identifierGenerateRetries = 5

// 内置分类的展示标签，对齐企业版实测数据
var systemCategoryLabels = map[string]string{
	models.CategoryGeneral:       "通用",
	models.CategorySensor:        "传感器",
	models.CategoryCommunication: "通信",
	models.CategoryOther:         "其他",
}

// 内置分类的固定 ID，保证多实例部署下不重复插入
var systemCategoryIDs = map[string]string{
	models.CategoryGeneral:       "ftc_general_001",
	models.CategorySensor:        "ftc_sensor_001",
	models.CategoryCommunication: "ftc_communication_001",
	models.CategoryOther:         "ftc_other_001",
}

type featureTemplateRepository struct {
	db *gorm.DB
}

// NewFeatureTemplateRepository 创建功能模板数据访问实例
func NewFeatureTemplateRepository(db *gorm.DB) FeatureTemplateRepository {
	return &featureTemplateRepository{db: db}
}

// SeedSystemCategories 插入内置分类（幂等）
func (r *featureTemplateRepository) SeedSystemCategories(ctx context.Context) error {
	for idx, category := range []string{
		models.CategoryGeneral,
		models.CategorySensor,
		models.CategoryCommunication,
		models.CategoryOther,
	} {
		row := models.FeatureTemplateCategory{
			ID:        systemCategoryIDs[category],
			Value:     category,
			Name:      category,
			Label:     systemCategoryLabels[category],
			SortOrder: idx + 1,
			IsSystem:  true,
			CreatedBy: "system",
		}

		exists, err := r.categoryExists(ctx, category)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil && !isDuplicateKeyError(err) {
			return fmt.Errorf("failed to seed category %s: %w", category, err)
		}
	}

	return nil
}

func (r *featureTemplateRepository) categoryExists(ctx context.Context, value string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.FeatureTemplateCategory{}).
		Where("value = ?", value).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check category: %w", err)
	}
	return count > 0, nil
}

// GetFeatureTemplateList 查询功能模板列表
func (r *featureTemplateRepository) GetFeatureTemplateList(ctx context.Context, req *models.FeatureTemplateListRequest) ([]models.FeatureTemplate, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	query := r.db.WithContext(ctx).Model(&models.FeatureTemplate{})

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		term := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR identifier LIKE ? OR description LIKE ?", term, term, term)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count feature templates: %w", err)
	}

	// 白名单排序字段，防止 SQL 注入
	sortColumn := "created_at"
	switch req.SortBy {
	case "name", "identifier", "created_at", "updated_at", "sort_order":
		sortColumn = req.SortBy
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	var templates []models.FeatureTemplate
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order(sortColumn + " " + sortOrder).Offset(offset).Limit(req.PageSize).Find(&templates).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list feature templates: %w", err)
	}

	return templates, total, nil
}

// GetFeatureTemplateByID 根据ID获取功能模板
func (r *featureTemplateRepository) GetFeatureTemplateByID(ctx context.Context, id string) (*models.FeatureTemplate, error) {
	var template models.FeatureTemplate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFeatureTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get feature template: %w", err)
	}
	return &template, nil
}

// CreateFeatureTemplate 创建功能模板
//
// 标识符为空时由服务端按名称生成；已提供则校验格式。
func (r *featureTemplateRepository) CreateFeatureTemplate(ctx context.Context, template *models.FeatureTemplate, autoIdentifier bool) error {
	if !autoIdentifier {
		if err := utils.ValidateFeatureTemplateIdentifierFormat(template.Identifier); err != nil {
			return err
		}
		if err := r.createWithIdentifier(ctx, template); err != nil {
			return err
		}
		return nil
	}

	var lastErr error
	for i := 0; i < identifierGenerateRetries; i++ {
		identifier, err := utils.GenerateFeatureTemplateIdentifier(template.Name)
		if err != nil {
			return err
		}
		template.Identifier = identifier

		if err := r.createWithIdentifier(ctx, template); err != nil {
			if errors.Is(err, ErrFeatureTemplateIdentifierExists) {
				lastErr = err
				continue
			}
			return err
		}
		return nil
	}
	return lastErr
}

func (r *featureTemplateRepository) createWithIdentifier(ctx context.Context, template *models.FeatureTemplate) error {
	if err := r.db.WithContext(ctx).Create(template).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrFeatureTemplateIdentifierExists
		}
		return fmt.Errorf("failed to create feature template: %w", err)
	}
	return nil
}

// UpdateFeatureTemplate 更新功能模板
//
// 标识符不在更新范围内：被客户端 SDK 引用，创建后不可更改。
func (r *featureTemplateRepository) UpdateFeatureTemplate(ctx context.Context, template *models.FeatureTemplate) error {
	err := r.db.WithContext(ctx).
		Model(&models.FeatureTemplate{}).
		Where("id = ?", template.ID).
		Select("name", "description", "category", "feature_config",
			"usage_limits", "custom_parameters", "sort_order", "status", "updated_at").
		Updates(template).Error
	if err != nil {
		return fmt.Errorf("failed to update feature template: %w", err)
	}
	return nil
}

// DeleteFeatureTemplate 删除功能模板
func (r *featureTemplateRepository) DeleteFeatureTemplate(ctx context.Context, id string) error {
	var template models.FeatureTemplate
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFeatureTemplateNotFound
		}
		return fmt.Errorf("failed to load feature template: %w", err)
	}

	if template.IsSystem {
		return ErrFeatureTemplateSystemProtected
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.FeatureTemplate{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete feature template: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFeatureTemplateNotFound
	}
	return nil
}

// GetCategoryList 查询功能模板分类列表，含各分类下的模板数
func (r *featureTemplateRepository) GetCategoryList(ctx context.Context) ([]models.FeatureTemplateCategory, error) {
	var categories []models.FeatureTemplateCategory
	if err := r.db.WithContext(ctx).
		Order("sort_order ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// 统计每个分类下的模板数
	type aggRow struct {
		Category string
		Total    int64
	}
	var rows []aggRow
	if err := r.db.WithContext(ctx).
		Model(&models.FeatureTemplate{}).
		Select("category, COUNT(*) as total").
		Group("category").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate templates: %w", err)
	}
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Category] = row.Total
	}

	for i := range categories {
		categories[i].TemplateCount = int(counts[categories[i].Value])
		// 内置分类不可删除；自定义分类下无模板时才可删除
		categories[i].CanDelete = !categories[i].IsSystem && counts[categories[i].Value] == 0
	}

	return categories, nil
}

// CreateCategory 新建自定义分类
func (r *featureTemplateRepository) CreateCategory(ctx context.Context, category *models.FeatureTemplateCategory) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("分类标识已存在")
		}
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

// UpdateCategory 更新自定义分类
func (r *featureTemplateRepository) UpdateCategory(ctx context.Context, category *models.FeatureTemplateCategory) error {
	result := r.db.WithContext(ctx).
		Model(&models.FeatureTemplateCategory{}).
		Where("id = ? AND is_system = ?", category.ID, false).
		Select("label", "sort_order", "updated_at").
		Updates(category)
	if result.Error != nil {
		return fmt.Errorf("failed to update category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFeatureTemplateSystemProtected
	}
	return nil
}

// DeleteCategory 删除自定义分类
func (r *featureTemplateRepository) DeleteCategory(ctx context.Context, id string) error {
	var category models.FeatureTemplateCategory
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTemplateCategoryNotFound
		}
		return fmt.Errorf("failed to load category: %w", err)
	}

	if category.IsSystem {
		return ErrFeatureTemplateSystemProtected
	}

	// 分类下仍有模板时拒绝删除
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.FeatureTemplate{}).
		Where("category = ?", category.Value).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count templates in category: %w", err)
	}
	if count > 0 {
		return ErrTemplateCategoryInUse
	}

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.FeatureTemplateCategory{}).Error; err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}
