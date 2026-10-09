package handlers

import (
	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type SystemConfigHandler struct {
	configService service.SystemConfigService
}

// NewSystemConfigHandler 创建系统配置处理器
func NewSystemConfigHandler(configService service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{configService: configService}
}

// GetPublicConfig 获取公开配置
// @Summary 获取公开配置
// @Description 返回登录页所需的站点信息，未认证即可访问。仅包含标记为公开的配置项
// @Tags 系统配置
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=map[string]string} "查询成功"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/public/config [get]
func (h *SystemConfigHandler) GetPublicConfig(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.configService.GetPublicConfigs(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data.Configs)
}

// GetGroupedConfigs 获取分组配置
// @Summary 获取系统配置
// @Description 按系统/认证/邮件分组返回全部配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.SystemConfigGroupResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 403 {object} models.ErrorResponse "权限不足"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/system-configs [get]
func (h *SystemConfigHandler) GetGroupedConfigs(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.configService.GetGroupedConfigs(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateConfig 创建配置
// @Summary 创建系统配置
// @Description 新增一项系统配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SystemConfigCreateRequest true "配置信息"
// @Success 200 {object} models.APIResponse{data=models.SystemConfig} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 409 {object} models.ErrorResponse "配置键已存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/system-configs [post]
func (h *SystemConfigHandler) CreateConfig(c *gin.Context) {
	var req models.SystemConfigCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.configService.CreateConfig(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateConfig 更新配置
// @Summary 更新系统配置
// @Description 更新一项系统配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Param request body models.SystemConfigUpdateRequest true "配置信息"
// @Success 200 {object} models.APIResponse{data=models.SystemConfig} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 404 {object} models.ErrorResponse "配置不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/system-configs/{key} [put]
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.SystemConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.configService.UpdateConfig(ctx, key, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// BatchUpdateConfigs 批量更新配置
// @Summary 批量更新系统配置
// @Description 一次提交多个配置项的值。配置页同时编辑多个键，逐项保存会产生中间态
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SystemConfigBatchUpdateRequest true "配置项列表"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 404 {object} models.ErrorResponse "配置不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/system-configs/batch [put]
func (h *SystemConfigHandler) BatchUpdateConfigs(c *gin.Context) {
	var req models.SystemConfigBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.configService.BatchUpdateConfigs(ctx, &req); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// DeleteConfig 删除配置
// @Summary 删除系统配置
// @Description 删除一项系统配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 404 {object} models.ErrorResponse "配置不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/system-configs/{key} [delete]
func (h *SystemConfigHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.configService.DeleteConfig(ctx, key); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// GetAPIKeyList 查询 API Key 列表
// @Summary 查询 API Key 列表
// @Description 返回 API Key 列表，仅含前缀不含明文
// @Tags API Key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param keyword query string false "按名称或前缀搜索"
// @Param status query string false "状态筛选"
// @Success 200 {object} models.APIResponse{data=models.APIKeyListResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/api-keys [get]
func (h *SystemConfigHandler) GetAPIKeyList(c *gin.Context) {
	var req models.APIKeyListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.configService.GetAPIKeyList(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateAPIKey 创建 API Key
// @Summary 创建 API Key
// @Description 创建 API Key。响应中的明文密钥仅此一次返回，之后无法再次查看
// @Tags API Key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.APIKeyCreateRequest true "API Key 信息"
// @Success 200 {object} models.APIResponse{data=models.APIKeyCreateResponse} "创建成功，明文密钥仅本次返回"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/api-keys [post]
func (h *SystemConfigHandler) CreateAPIKey(c *gin.Context) {
	var req models.APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	// 记录创建者，供审计追溯
	createdBy := currentUserIDFromContext(c)

	data, err := h.configService.CreateAPIKey(ctx, &req, createdBy)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateAPIKey 更新 API Key
// @Summary 更新 API Key
// @Description 更新 API Key 的名称、状态、有效期或 IP 白名单
// @Tags API Key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "API Key ID"
// @Param request body models.APIKeyUpdateRequest true "API Key 信息"
// @Success 200 {object} models.APIResponse "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 404 {object} models.ErrorResponse "API Key 不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/api-keys/{id} [put]
func (h *SystemConfigHandler) UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.APIKeyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.configService.UpdateAPIKey(ctx, id, &req); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// DeleteAPIKey 删除 API Key
// @Summary 删除 API Key
// @Description 删除 API Key，删除后立即失效
// @Tags API Key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "API Key ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 404 {object} models.ErrorResponse "API Key 不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/api-keys/{id} [delete]
func (h *SystemConfigHandler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.configService.DeleteAPIKey(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// currentUserIDFromContext 取当前登录用户ID，供审计字段使用
func currentUserIDFromContext(c *gin.Context) string {
	if value, exists := c.Get("user_id"); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}
