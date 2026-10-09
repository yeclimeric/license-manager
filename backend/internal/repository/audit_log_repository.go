package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"license-manager/internal/models"
)

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository 创建审计日志数据访问实例
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create 写入一条审计日志
func (r *auditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLogList 查询审计日志列表
func (r *auditLogRepository) GetAuditLogList(ctx context.Context, req *models.AuditLogListRequest) ([]models.AuditLog, int64, error) {
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

	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	if req.ResourceType != "" {
		query = query.Where("resource_type = ?", req.ResourceType)
	}
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.ProductID != "" {
		query = query.Where("product_id = ?", req.ProductID)
	}
	if action := strings.TrimSpace(req.Action); action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}
	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 白名单排序字段，防止 SQL 注入
	sortColumn := "created_at"
	switch sortBy {
	case "created_at", "action", "resource_type", "status":
		sortColumn = sortBy
	}

	var logs []models.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Order(sortColumn + " " + sortOrder).Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, total, nil
}

// GetStatistics 统计审计日志总数与分布
func (r *auditLogRepository) GetStatistics(ctx context.Context) (*models.AuditLogStatisticsResponse, error) {
	stats := &models.AuditLogStatisticsResponse{
		ByResource: map[string]int64{},
		ByStatus:   map[string]int64{},
	}

	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	type row struct {
		Key   string
		Total int64
	}

	var statusRows []row
	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Select("status as `key`, COUNT(*) as total").
		Group("status").Find(&statusRows).Error; err != nil {
		return nil, fmt.Errorf("failed to group by status: %w", err)
	}
	for _, r := range statusRows {
		stats.ByStatus[r.Key] = r.Total
	}
	stats.Success = stats.ByStatus[models.AuditStatusSuccess]
	stats.Failure = stats.ByStatus[models.AuditStatusFailure]

	var resourceRows []row
	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Select("resource_type as `key`, COUNT(*) as total").
		Group("resource_type").Find(&resourceRows).Error; err != nil {
		return nil, fmt.Errorf("failed to group by resource type: %w", err)
	}
	for _, r := range resourceRows {
		stats.ByResource[r.Key] = r.Total
	}

	return stats, nil
}
