package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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
		// 链的根条目无前驱签名：客户端内置这把公钥作为信任锚
		SigningKeyChain: models.JSON(mustMarshalJSON([]models.SigningKeyChainEntry{{
			PublicKey: signingKey.PublicKeyPEM,
		}})),
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
		PublicKeys:  CollectSigningPublicKeys(offering),
		KeyChain:    ParseSigningKeyChain(offering.SigningKeyChain),
	}, nil
}

// RotateSigningKey 轮换产品项签名密钥
//
// 生成新密钥对并将其设为当前签发密钥，旧当前公钥移入链上留档。
// 旧私钥不保留（无法再签发新许可），但旧公钥必须留档：已发放的许可证
// 是旧私钥签的，客户端仍需旧公钥才能验签通过。
//
// 关键设计：轮换时用「旧私钥」为「新公钥」签名，写入链上新条目。
// 客户端只要信任链上的根公钥（首次内置的那把），就能逐级验证整条链，
// 把轮换后的新公钥纳入信任——密钥轮换无需重新发版客户端。
//
// 客户端也可一次取全部公钥（PublicKeys）直接内置，效果等价。
func (s *productOfferingService) RotateSigningKey(ctx context.Context, id string, req *models.ProductOfferingRotateKeyRequest) (*models.ProductOfferingPublicKeyResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	offering, err := s.repo.GetProductOfferingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductOfferingNotFound) {
			return nil, i18n.NewI18nError("500001", lang)
		}
		return nil, i18n.NewI18nError("900005", lang)
	}

	newKey, genErr := utils.GenerateProductSigningKey()
	if genErr != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	reason := ""
	if req != nil {
		reason = strings.TrimSpace(req.Reason)
	}

	chain := ParseSigningKeyChain(offering.SigningKeyChain)
	now := time.Now()

	if offering.SigningPrivateKey != "" && offering.SigningPublicKey != "" {
		// 旧当前密钥退役：留档其公钥，并由旧私钥为新公钥背书
		if len(chain) > 0 {
			chain[len(chain)-1].RetiredAt = &now
			chain[len(chain)-1].Reason = reason
		} else {
			// 链为空（升级前创建的产品项）：补一条根条目
			chain = append(chain, models.SigningKeyChainEntry{
				PublicKey: offering.SigningPublicKey,
				RetiredAt: &now,
				Reason:    reason,
			})
		}

		// 用旧私钥为新公钥签名，形成链的衔接证明
		sig, signErr := utils.SignWithPrivateKeyPEM(offering.SigningPrivateKey, []byte(newKey.PublicKeyPEM))
		if signErr != nil {
			return nil, i18n.NewI18nError("900005", lang)
		}
		chain = append(chain, models.SigningKeyChainEntry{
			PublicKey: newKey.PublicKeyPEM,
			Signature: sig,
		})
	} else {
		// 无密钥产品项：直接补生成，作为链的根（无前驱签名）
		chain = append(chain, models.SigningKeyChainEntry{
			PublicKey: newKey.PublicKeyPEM,
		})
	}

	offering.SigningPrivateKey = newKey.PrivateKeyPEM
	offering.SigningPublicKey = newKey.PublicKeyPEM
	offering.SigningKeyChain = models.JSON(mustMarshalJSON(chain))

	if err := s.repo.UpdateSigningKeys(ctx, offering); err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	return &models.ProductOfferingPublicKeyResponse{
		ProductID:   offering.ID,
		ProductCode: offering.Code,
		PublicKey:   offering.SigningPublicKey,
		PublicKeys:  CollectSigningPublicKeys(offering),
		KeyChain:    chain,
	}, nil
}

// ParseSigningKeyChain 解析签名密钥链
func ParseSigningKeyChain(raw models.JSON) []models.SigningKeyChainEntry {
	if len(raw) == 0 {
		return nil
	}
	var entries []models.SigningKeyChainEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil
	}
	return entries
}

// CollectSigningPublicKeys 收集产品项全部有效公钥
//
// 顺序：当前签发公钥在前，链上历史公钥在后（由新到旧）。
// 客户端按顺序逐一尝试验签，任一通过即视为签名可信。
func CollectSigningPublicKeys(offering *models.ProductOffering) []string {
	if offering == nil {
		return nil
	}
	keys := make([]string, 0, 4)
	if offering.SigningPublicKey != "" {
		keys = append(keys, offering.SigningPublicKey)
	}
	chain := ParseSigningKeyChain(offering.SigningKeyChain)
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].PublicKey != "" {
			keys = append(keys, chain[i].PublicKey)
		}
	}
	return keys
}

func mustMarshalJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
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
