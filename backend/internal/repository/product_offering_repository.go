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

// codeGenerateRetries 编码冲突时的重试次数。
// 12 位随机字符的碰撞概率极低，重试仅作为兜底。
const codeGenerateRetries = 5

type productOfferingRepository struct {
	db *gorm.DB
}

// NewProductOfferingRepository 创建产品项数据访问实例
func NewProductOfferingRepository(db *gorm.DB) ProductOfferingRepository {
	return &productOfferingRepository{db: db}
}

// GetProductOfferingList 查询产品项列表
func (r *productOfferingRepository) GetProductOfferingList(ctx context.Context, req *models.ProductOfferingListRequest) ([]models.ProductOffering, int64, error) {
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

	query := r.db.WithContext(ctx).Model(&models.ProductOffering{})

	// 关键词模糊搜索：名称、别名、编码
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		term := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR alias LIKE ? OR code LIKE ?", term, term, term)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count product offerings: %w", err)
	}

	// 白名单排序字段，防止 SQL 注入
	sortColumn := "created_at"
	switch req.SortBy {
	case "name", "code", "created_at", "updated_at":
		sortColumn = req.SortBy
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	var offerings []models.ProductOffering
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order(sortColumn + " " + sortOrder).Offset(offset).Limit(req.PageSize).Find(&offerings).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list product offerings: %w", err)
	}

	// 填充授权码统计
	counts, err := r.authCodeCounts(ctx, extractIDs(offerings))
	if err != nil {
		return nil, 0, err
	}
	for i := range offerings {
		offerings[i].AuthCodeCount = int(counts[offerings[i].ID])
		// UserCount 暂复用授权码数，产品项级用户统计在后续阶段补充
		offerings[i].UserCount = int(counts[offerings[i].ID])
	}

	return offerings, total, nil
}

// GetProductOfferingByID 根据ID获取产品项
func (r *productOfferingRepository) GetProductOfferingByID(ctx context.Context, id string) (*models.ProductOffering, error) {
	var offering models.ProductOffering
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&offering).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductOfferingNotFound
		}
		return nil, fmt.Errorf("failed to get product offering: %w", err)
	}

	// 统计字段为非持久化列，需聚合后单独回填
	counts, err := r.authCodeCounts(ctx, []string{id})
	if err != nil {
		return nil, err
	}
	offering.AuthCodeCount = int(counts[id])
	offering.UserCount = int(counts[id])

	return &offering, nil
}

// GetProductOfferingByCode 根据编码获取产品项
func (r *productOfferingRepository) GetProductOfferingByCode(ctx context.Context, code string) (*models.ProductOffering, error) {
	var offering models.ProductOffering
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&offering).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductOfferingNotFound
		}
		return nil, fmt.Errorf("failed to get product offering by code: %w", err)
	}
	return &offering, nil
}

// CreateProductOffering 创建产品项，编码由服务端生成
func (r *productOfferingRepository) CreateProductOffering(ctx context.Context, offering *models.ProductOffering) error {
	var lastErr error
	for i := 0; i < codeGenerateRetries; i++ {
		code, err := utils.GenerateProductCode()
		if err != nil {
			return fmt.Errorf("failed to generate product code: %w", err)
		}
		offering.Code = code

		if err := r.db.WithContext(ctx).Create(offering).Error; err != nil {
			if isDuplicateKeyError(err) {
				lastErr = ErrProductOfferingDuplicate
				continue
			}
			return fmt.Errorf("failed to create product offering: %w", err)
		}
		return nil
	}
	return lastErr
}

// UpdateProductOffering 更新产品项
//
// 编码不在更新范围内：系统生成且不可更改（D4 决策）。
func (r *productOfferingRepository) UpdateProductOffering(ctx context.Context, offering *models.ProductOffering) error {
	// Select 显式限定可更新字段，防止误改 code
	err := r.db.WithContext(ctx).
		Model(&models.ProductOffering{}).
		Where("id = ?", offering.ID).
		Select("name", "alias", "description", "is_trial_product", "trial_days",
			"feature_config", "usage_limits", "custom_parameters",
			"max_activations", "duration_days", "deployment_mode", "status", "updated_at").
		Updates(offering).Error
	if err != nil {
		return fmt.Errorf("failed to update product offering: %w", err)
	}
	return nil
}

// DeleteProductOffering 删除产品项
//
// 若产品项下仍有授权码则拒绝删除，避免授权码成为孤儿。
func (r *productOfferingRepository) DeleteProductOffering(ctx context.Context, id string) error {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuthorizationCode{}).
		Where("product_offering_id = ?", id).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count authorization codes: %w", err)
	}
	if count > 0 {
		return ErrProductOfferingInUse
	}

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ProductOffering{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete product offering: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProductOfferingNotFound
	}
	return nil
}

// UpdateProductOfferingStatus 更新产品项状态
func (r *productOfferingRepository) UpdateProductOfferingStatus(ctx context.Context, id string, status string) error {
	result := r.db.WithContext(ctx).
		Model(&models.ProductOffering{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update product offering status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrProductOfferingNotFound
	}
	return nil
}

// authCodeCounts 聚合指定产品项下的授权码数量
func (r *productOfferingRepository) authCodeCounts(ctx context.Context, ids []string) (map[string]int64, error) {
	counts := make(map[string]int64, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}

	type aggRow struct {
		ProductOfferingID string
		Total             int64
	}

	var rows []aggRow
	err := r.db.WithContext(ctx).
		Model(&models.AuthorizationCode{}).
		Select("product_offering_id, COUNT(*) as total").
		Where("product_offering_id IN ?", ids).
		Group("product_offering_id").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate authorization codes: %w", err)
	}

	for _, row := range rows {
		counts[row.ProductOfferingID] = row.Total
	}
	return counts, nil
}

// extractIDs 提取产品项 ID 列表
func extractIDs(offerings []models.ProductOffering) []string {
	ids := make([]string, 0, len(offerings))
	for _, o := range offerings {
		ids = append(ids, o.ID)
	}
	return ids
}

// isDuplicateKeyError 判断是否为唯一键冲突
func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
