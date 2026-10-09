package handlers

import (
	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	auditLogService service.AuditLogService
}

// NewAuditLogHandler 创建审计日志处理器
func NewAuditLogHandler(auditLogService service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogService: auditLogService,
	}
}

// GetAuditLogList 查询审计日志列表
// @Summary 查询审计日志列表
// @Description 分页查询操作审计日志，支持资源类型、操作人、状态等筛选
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" minimum(1)
// @Param page_size query int false "每页条数" minimum(1) maximum(100)
// @Param resource_type query string false "资源类型筛选"
// @Param user_id query string false "操作人ID"
// @Param status query string false "状态筛选" Enums(success, failure)
// @Param action query string false "动作模糊匹配"
// @Param start_date query string false "开始时间"
// @Param end_date query string false "结束时间"
// @Success 200 {object} models.APIResponse{data=models.AuditLogListResponse} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/audit-logs [get]
func (h *AuditLogHandler) GetAuditLogList(c *gin.Context) {
	var req models.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.auditLogService.GetAuditLogList(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetStatistics 查询审计统计
// @Summary 查询审计统计
// @Description 查询审计日志总数、成功失败分布与资源类型分布
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.AuditLogStatisticsResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/audit-logs/statistics [get]
func (h *AuditLogHandler) GetStatistics(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.auditLogService.GetStatistics(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}
