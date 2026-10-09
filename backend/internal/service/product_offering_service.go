package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
	"license-manager/pkg/utils"
)

// maxTrialDays 体验天数上限，防止误填导致授权永不过期
const maxTrialDays = 3650

// validDeploymentModes 允许的部署模式
var validDeploymentModes = map[string]bool{
	"online":  true,
	"offline": true,
	"hybrid":  true,
}

type productOfferingService struct {
	repo repository.ProductOfferingRepository
}

// NewProductOfferingService 创建产品项服务实例
func NewProductOfferingService(repo repository.ProductOfferingRepository) ProductOfferingService {
	return &productOfferingService{repo: repo}
}

// GetProductOfferingList 查询产品项列表
func (s *productOfferingService) GetProductOfferingList(ctx context.Context, req *models.ProductOfferingListRequest) (*models.ProductOfferingListResponse, error) {
	offerings, total, err := s.repo.GetProductOfferingList(ctx, req)
	if err != nil {
		return nil, i18n.NewI18nError("900005", pkgcontext.GetLanguageFromContext(ctx))
	}

	lang := pkgcontext.GetLanguageFromContext(ctx)
	items := make([]models.ProductOfferingResponse, 0, len(offerings))
	for _, offering := range offerings {
		items = append(items, s.toResponse(offering, lang))
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &models.ProductOfferingListResponse{
		List:       items,
		Total:      int(total),
		Page:       req.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProductOffering 获取产品项详情
func (s *productOfferingService) GetProductOffering(ctx context.Context, id string) (*models.ProductOfferingResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	offering, err := s.repo.GetProductOfferingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductOfferingNotFound) {
			return nil, i18n.NewI18nError("500001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	resp := s.toResponse(*offering, lang)
	return &resp, nil
}

// CreateProductOffering 创建产品项
//
// 编码不在请求中，由服务端生成（D4 决策）。
func (s *productOfferingService) CreateProductOffering(ctx context.Context, req *models.ProductOfferingCreateRequest) (*models.ProductOfferingResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, i18n.NewI18nError("500003", lang)
	}

	if err := validateTrial(req.IsTrialProduct, req.TrialDays, lang); err != nil {
		return nil, err
	}

	mode := req.DeploymentMode
	if mode == "" {
		mode = "online"
	}
	if !validDeploymentModes[mode] {
		return nil, i18n.NewI18nError("500009", lang)
	}

	// 创建时生成产品专属签名密钥对：客户端用公钥验签，私钥仅服务端持有。
	// 密钥不提供轮换入口——轮换会使已发放的许可证全部失效。
	signingKey, err := utils.GenerateProductSigningKey()
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	offering := &models.ProductOffering{
		Name:              name,
		Alias:             strings.TrimSpace(req.Alias),
		Description:       strings.TrimSpace(req.Description),
		IsTrialProduct:    req.IsTrialProduct,
		TrialDays:         req.TrialDays,
		MaxActivations:    req.MaxActivations,
		DurationDays:      req.DurationDays,
		DeploymentMode:    mode,
		Status:            "active",
		FeatureConfig:     buildJSONObject(req.FeatureConfig),
		UsageLimits:       buildJSONObject(req.UsageLimits),
		CustomParameters:  buildJSONObject(req.CustomParameters),
		SigningPrivateKey: signingKey.PrivateKeyPEM,
		SigningPublicKey:  signingKey.PublicKeyPEM,
	}

	if err := s.repo.CreateProductOffering(ctx, offering); err != nil {
		if errors.Is(err, repository.ErrProductOfferingDuplicate) {
			return nil, i18n.NewI18nError("500002", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	resp := s.toResponse(*offering, lang)
	return &resp, nil
}

// UpdateProductOffering 更新产品项
//
// 编码不可更新：请求结构中不含 Code 字段，服务层亦不读取。
func (s *productOfferingService) UpdateProductOffering(ctx context.Context, id string, req *models.ProductOfferingUpdateRequest) (*models.ProductOfferingResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	existing, err := s.repo.GetProductOfferingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductOfferingNotFound) {
			return nil, i18n.NewI18nError("500001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	updated := *existing

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, i18n.NewI18nError("500003", lang)
		}
		updated.Name = name
	}
	if req.Alias != nil {
		updated.Alias = strings.TrimSpace(*req.Alias)
	}
	if req.Description != nil {
		updated.Description = strings.TrimSpace(*req.Description)
	}
	if req.MaxActivations != nil {
		updated.MaxActivations = req.MaxActivations
	}
	if req.DurationDays != nil {
		updated.DurationDays = req.DurationDays
	}
	if req.DeploymentMode != nil {
		if !validDeploymentModes[*req.DeploymentMode] {
			return nil, i18n.NewI18nError("500009", lang)
		}
		updated.DeploymentMode = *req.DeploymentMode
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" {
			return nil, i18n.NewI18nError("500010", lang)
		}
		updated.Status = *req.Status
	}

	// 体验开关与天数需成对校验：先取请求值，未提供则沿用现值
	isTrial := updated.IsTrialProduct
	if req.IsTrialProduct != nil {
		isTrial = *req.IsTrialProduct
	}
	trialDays := updated.TrialDays
	if req.TrialDays != nil {
		trialDays = req.TrialDays
	}
	if err := validateTrial(isTrial, trialDays, lang); err != nil {
		return nil, err
	}
	updated.IsTrialProduct = isTrial
	updated.TrialDays = trialDays

	if req.FeatureConfig != nil {
		updated.FeatureConfig = buildJSONObject(req.FeatureConfig)
	}
	if req.UsageLimits != nil {
		updated.UsageLimits = buildJSONObject(req.UsageLimits)
	}
	if req.CustomParameters != nil {
		updated.CustomParameters = buildJSONObject(req.CustomParameters)
	}

	// 保留原编码，防止意外覆盖
	updated.Code = existing.Code

	if err := s.repo.UpdateProductOffering(ctx, &updated); err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	resp := s.toResponse(updated, lang)
	return &resp, nil
}

// DeleteProductOffering 删除产品项
func (s *productOfferingService) DeleteProductOffering(ctx context.Context, id string) error {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if err := s.repo.DeleteProductOffering(ctx, id); err != nil {
		switch {
		case errors.Is(err, repository.ErrProductOfferingNotFound):
			return i18n.NewI18nError("500001", lang)
		case errors.Is(err, repository.ErrProductOfferingInUse):
			return i18n.NewI18nError("500004", lang)
		default:
			return i18n.NewI18nError("900005", lang)
		}
	}
	return nil
}

// UpdateProductOfferingStatus 启用/停用产品项
func (s *productOfferingService) UpdateProductOfferingStatus(ctx context.Context, id string, status string) (*models.ProductOfferingResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if status != "active" && status != "inactive" {
		return nil, i18n.NewI18nError("500010", lang)
	}

	if err := s.repo.UpdateProductOfferingStatus(ctx, id, status); err != nil {
		if errors.Is(err, repository.ErrProductOfferingNotFound) {
			return nil, i18n.NewI18nError("500001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	offering, err := s.repo.GetProductOfferingByID(ctx, id)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	resp := s.toResponse(*offering, lang)
	return &resp, nil
}

// toResponse 补充多语言展示字段
func (s *productOfferingService) toResponse(offering models.ProductOffering, lang string) models.ProductOfferingResponse {
	offering.StatusDisplay = i18n.GetEnumMessage("product_offering_status", offering.Status, lang)

	resp := models.ProductOfferingResponse{ProductOffering: &offering}

	// 体验列展示：体验/15 天、体验/30 天，无体验则留空
	if offering.IsTrialProduct && offering.TrialDays != nil && *offering.TrialDays > 0 {
		resp.TrialDisplay = fmt.Sprintf("%s / %d %s",
			i18n.GetEnumMessage("product_offering_trial", "enabled", lang),
			*offering.TrialDays,
			i18n.GetEnumMessage("product_offering_trial", "days", lang))
	}

	// 有授权码时不可删除
	resp.CanDelete = offering.AuthCodeCount == 0
	resp.CanEdit = true
	resp.HasSigningKey = offering.SigningPublicKey != ""

	return resp
}

// GetProductOfferingPublicKey 获取产品项公钥
//
// 仅返回公钥。历史产品项可能未生成密钥（升级前创建），此时返回空公钥
// 而非报错，让调用方能区分"无密钥"与"无此产品"。
func (s *productOfferingService) GetProductOfferingPublicKey(ctx context.Context, id string) (*models.ProductOfferingPublicKeyResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	offering, err := s.repo.GetProductOfferingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductOfferingNotFound) {
			return nil, i18n.NewI18nError("500001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	return &models.ProductOfferingPublicKeyResponse{
		ProductID:   offering.ID,
		ProductCode: offering.Code,
		PublicKey:   offering.SigningPublicKey,
	}, nil
}

// validateTrial 校验体验配置
func validateTrial(isTrial bool, trialDays *int, lang string) error {
	if !isTrial {
		return nil
	}
	if trialDays == nil {
		return i18n.NewI18nError("500007", lang)
	}
	if *trialDays <= 0 {
		return i18n.NewI18nError("500006", lang)
	}
	if *trialDays > maxTrialDays {
		return i18n.NewI18nError("500008", lang)
	}
	return nil
}

// buildJSONObject 将 map 序列化为 models.JSON，nil 转为空对象
func buildJSONObject(m map[string]interface{}) models.JSON {
	if len(m) == 0 {
		return models.JSON("{}")
	}
	data, err := json.Marshal(m)
	if err != nil {
		// map 来自 JSON 请求体，序列化失败说明结构异常，退回空对象而非中断写入
		return models.JSON("{}")
	}
	return models.JSON(data)
}
