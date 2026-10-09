package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
)

type featureTemplateService struct {
	repo repository.FeatureTemplateRepository
}

// NewFeatureTemplateService 创建功能模板服务实例
func NewFeatureTemplateService(repo repository.FeatureTemplateRepository) FeatureTemplateService {
	return &featureTemplateService{repo: repo}
}

// GetFeatureTemplateList 查询功能模板列表
func (s *featureTemplateService) GetFeatureTemplateList(ctx context.Context, req *models.FeatureTemplateListRequest) (*models.FeatureTemplateListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	templates, total, err := s.repo.GetFeatureTemplateList(ctx, req)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	labels, err := s.categoryLabels(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	for i := range templates {
		s.decorate(&templates[i], labels, lang)
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &models.FeatureTemplateListResponse{
		List:       templates,
		Total:      int(total),
		Page:       req.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetFeatureTemplate 获取功能模板详情
func (s *featureTemplateService) GetFeatureTemplate(ctx context.Context, id string) (*models.FeatureTemplate, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	template, err := s.repo.GetFeatureTemplateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrFeatureTemplateNotFound) {
			return nil, i18n.NewI18nError("510001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	labels, err := s.categoryLabels(ctx)
	if err == nil {
		s.decorate(template, labels, lang)
	}
	return template, nil
}

// CreateFeatureTemplate 创建功能模板
//
// 标识符留空时按名称自动生成；自动生成失败（如名称为纯中文）时要求调用方显式提供。
func (s *featureTemplateService) CreateFeatureTemplate(ctx context.Context, req *models.FeatureTemplateCreateRequest) (*models.FeatureTemplate, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, i18n.NewI18nError("510003", lang)
	}

	identifier := strings.TrimSpace(req.Identifier)
	autoIdentifier := identifier == ""
	if !autoIdentifier {
		if err := validateIdentifier(identifier, lang); err != nil {
			return nil, err
		}
	}

	category := req.Category
	if category == "" {
		category = models.CategoryGeneral
	}
	if err := s.validateCategory(ctx, category, lang); err != nil {
		return nil, err
	}

	featureConfig, err := buildFeatureArray(req.FeatureConfig, lang)
	if err != nil {
		return nil, err
	}

	template := &models.FeatureTemplate{
		Name:             name,
		Identifier:       identifier,
		Description:      strings.TrimSpace(req.Description),
		Category:         category,
		FeatureConfig:    featureConfig,
		UsageLimits:      buildJSONObject(req.UsageLimits),
		CustomParameters: buildJSONObject(req.CustomParameters),
		SortOrder:        intOrDefault(req.SortOrder, 0),
		Status:           "active",
	}

	if err := s.repo.CreateFeatureTemplate(ctx, template, autoIdentifier); err != nil {
		switch {
		case errors.Is(err, repository.ErrFeatureTemplateIdentifierExists):
			return nil, i18n.NewI18nError("510002", lang)
		case strings.Contains(err.Error(), "无法转换为合法标识符"):
			return nil, i18n.NewI18nError("510006", lang)
		default:
			return nil, i18n.NewI18nError("900005", lang)
		}
	}

	labels, _ := s.categoryLabels(ctx)
	s.decorate(template, labels, lang)
	return template, nil
}

// UpdateFeatureTemplate 更新功能模板
//
// 标识符不可更新：被客户端 SDK 引用。
func (s *featureTemplateService) UpdateFeatureTemplate(ctx context.Context, id string, req *models.FeatureTemplateUpdateRequest) (*models.FeatureTemplate, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	existing, err := s.repo.GetFeatureTemplateByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrFeatureTemplateNotFound) {
			return nil, i18n.NewI18nError("510001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	if existing.IsSystem {
		return nil, i18n.NewI18nError("510007", lang)
	}

	updated := *existing

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, i18n.NewI18nError("510003", lang)
		}
		updated.Name = name
	}
	if req.Description != nil {
		updated.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		if err := s.validateCategory(ctx, *req.Category, lang); err != nil {
			return nil, err
		}
		updated.Category = *req.Category
	}
	if req.SortOrder != nil {
		updated.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" {
			return nil, i18n.NewI18nError("500010", lang)
		}
		updated.Status = *req.Status
	}
	if req.FeatureConfig != nil {
		featureConfig, err := buildFeatureArray(req.FeatureConfig, lang)
		if err != nil {
			return nil, err
		}
		updated.FeatureConfig = featureConfig
	}
	if req.UsageLimits != nil {
		updated.UsageLimits = buildJSONObject(req.UsageLimits)
	}
	if req.CustomParameters != nil {
		updated.CustomParameters = buildJSONObject(req.CustomParameters)
	}

	// 保留原标识符，防止误改
	updated.Identifier = existing.Identifier

	if err := s.repo.UpdateFeatureTemplate(ctx, &updated); err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	labels, _ := s.categoryLabels(ctx)
	s.decorate(&updated, labels, lang)
	return &updated, nil
}

// DeleteFeatureTemplate 删除功能模板
func (s *featureTemplateService) DeleteFeatureTemplate(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.repo.DeleteFeatureTemplate(ctx, id); err != nil {
		switch {
		case errors.Is(err, repository.ErrFeatureTemplateNotFound):
			return i18n.NewI18nError("510001", lang)
		case errors.Is(err, repository.ErrFeatureTemplateSystemProtected):
			return i18n.NewI18nError("510007", lang)
		default:
			return i18n.NewI18nError("900005", lang)
		}
	}
	return nil
}

// GetCategoryList 查询分类列表
func (s *featureTemplateService) GetCategoryList(ctx context.Context) (*models.FeatureTemplateCategoryListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	categories, err := s.repo.GetCategoryList(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	return &models.FeatureTemplateCategoryListResponse{
		List:  categories,
		Total: len(categories),
	}, nil
}

// CreateCategory 新建自定义分类
func (s *featureTemplateService) CreateCategory(ctx context.Context, category *models.FeatureTemplateCategory) (*models.FeatureTemplateCategory, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	value := strings.TrimSpace(category.Value)
	label := strings.TrimSpace(category.Label)

	if value == "" || label == "" {
		return nil, i18n.NewI18nError("900001", lang)
	}

	// 自定义分类标识需符合 slug 约定
	if err := validateCategoryValue(value, lang); err != nil {
		return nil, err
	}

	// 内置分类标识保留
	switch value {
	case models.CategoryGeneral, models.CategorySensor, models.CategoryCommunication, models.CategoryOther:
		return nil, i18n.NewI18nError("510010", lang)
	}

	row := &models.FeatureTemplateCategory{
		Value:     value,
		Name:      value,
		Label:     label,
		SortOrder: category.SortOrder,
		IsSystem:  false,
		CreatedBy: "admin",
	}

	if err := s.repo.CreateCategory(ctx, row); err != nil {
		if strings.Contains(err.Error(), "已存在") {
			return nil, i18n.NewI18nError("510010", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	row.CanDelete = true
	return row, nil
}

// UpdateCategory 更新自定义分类
func (s *featureTemplateService) UpdateCategory(ctx context.Context, id string, label string, sortOrder *int) (*models.FeatureTemplateCategory, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	categories, err := s.repo.GetCategoryList(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	var target *models.FeatureTemplateCategory
	for i := range categories {
		if categories[i].ID == id {
			target = &categories[i]
			break
		}
	}
	if target == nil {
		return nil, i18n.NewI18nError("510008", lang)
	}
	if target.IsSystem {
		return nil, i18n.NewI18nError("510011", lang)
	}

	row := &models.FeatureTemplateCategory{ID: id, Label: label, SortOrder: target.SortOrder}
	if label = strings.TrimSpace(label); label != "" {
		row.Label = label
	} else {
		row.Label = target.Label
	}
	if sortOrder != nil {
		row.SortOrder = *sortOrder
	}

	if err := s.repo.UpdateCategory(ctx, row); err != nil {
		if errors.Is(err, repository.ErrFeatureTemplateSystemProtected) {
			return nil, i18n.NewI18nError("510011", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	row.CanDelete = target.CanDelete
	row.IsSystem = false
	row.Value = target.Value
	row.Name = target.Name
	return row, nil
}

// DeleteCategory 删除自定义分类
func (s *featureTemplateService) DeleteCategory(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		switch {
		case errors.Is(err, repository.ErrTemplateCategoryNotFound):
			return i18n.NewI18nError("510008", lang)
		case errors.Is(err, repository.ErrFeatureTemplateSystemProtected):
			return i18n.NewI18nError("510011", lang)
		case errors.Is(err, repository.ErrTemplateCategoryInUse):
			return i18n.NewI18nError("510009", lang)
		default:
			return i18n.NewI18nError("900005", lang)
		}
	}
	return nil
}

// decorate 补充多语言展示字段
func (s *featureTemplateService) decorate(template *models.FeatureTemplate, labels map[string]string, lang string) {
	template.StatusDisplay = i18n.GetEnumMessage("feature_template_status", template.Status, lang)
	if label, ok := labels[template.Category]; ok {
		template.CategoryDisplay = label
	} else {
		template.CategoryDisplay = template.Category
	}
	template.CanDelete = !template.IsSystem
}

// categoryLabels 获取分类标签映射
func (s *featureTemplateService) categoryLabels(ctx context.Context) (map[string]string, error) {
	categories, err := s.repo.GetCategoryList(ctx)
	if err != nil {
		return nil, err
	}
	labels := make(map[string]string, len(categories))
	for _, c := range categories {
		labels[c.Value] = c.Label
	}
	return labels, nil
}

// validateCategory 校验分类是否存在
func (s *featureTemplateService) validateCategory(ctx context.Context, value string, lang string) error {
	categories, err := s.repo.GetCategoryList(ctx)
	if err != nil {
		return i18n.NewI18nError("900005", lang)
	}
	for _, c := range categories {
		if c.Value == value {
			return nil
		}
	}
	return i18n.NewI18nError("510008", lang)
}

// validateIdentifier 校验标识符格式
func validateIdentifier(identifier string, lang string) error {
	if identifier == "" {
		return i18n.NewI18nError("510004", lang)
	}
	if !isValidIdentifierFormat(identifier) {
		return i18n.NewI18nError("510005", lang)
	}
	return nil
}

// validateCategoryValue 校验自定义分类标识格式
func validateCategoryValue(value string, lang string) error {
	if !isValidIdentifierFormat(value) {
		return i18n.NewI18nError("510005", lang)
	}
	return nil
}

func isValidIdentifierFormat(v string) bool {
	if v == "" {
		return false
	}
	first := v[0]
	if !(first >= 'a' && first <= 'z') {
		return false
	}
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c != '_' && !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// buildFeatureArray 将功能项列表序列化为 JSON 数组
func buildFeatureArray(items []interface{}, lang string) (models.JSON, error) {
	if items == nil {
		return models.JSON("[]"), nil
	}
	data, err := json.Marshal(items)
	if err != nil {
		return nil, i18n.NewI18nError("510012", lang)
	}
	return models.JSON(data), nil
}

// intOrDefault 返回指针值或默认值
func intOrDefault(v *int, def int) int {
	if v == nil {
		return def
	}
	return *v
}
