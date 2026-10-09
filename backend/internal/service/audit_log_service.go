package service

import (
	"context"

	"license-manager/internal/models"
	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
)

type auditLogService struct {
	repo repository.AuditLogRepository
}

// NewAuditLogService 创建审计日志服务实例
func NewAuditLogService(repo repository.AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

// GetAuditLogList 查询审计日志列表
func (s *auditLogService) GetAuditLogList(ctx context.Context, req *models.AuditLogListRequest) (*models.AuditLogListResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	if req == nil {
		req = &models.AuditLogListRequest{}
	}

	logs, total, err := s.repo.GetAuditLogList(ctx, req)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}

	// 空结果返回空数组而非 null，前端无需判空
	if logs == nil {
		logs = []models.AuditLog{}
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &models.AuditLogListResponse{
		List:       logs,
		Total:      int(total),
		Page:       req.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStatistics 查询审计日志统计
func (s *auditLogService) GetStatistics(ctx context.Context) (*models.AuditLogStatisticsResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	stats, err := s.repo.GetStatistics(ctx)
	if err != nil {
		return nil, i18n.NewI18nError("900005", lang)
	}
	return stats, nil
}
