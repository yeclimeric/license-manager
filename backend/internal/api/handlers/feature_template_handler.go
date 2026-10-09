package handlers

import (
	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type FeatureTemplateHandler struct {
	featureTemplateService service.FeatureTemplateService
}

// NewFeatureTemplateHandler 创建功能模板处理器
func NewFeatureTemplateHandler(featureTemplateService service.FeatureTemplateService) *FeatureTemplateHandler {
	return &FeatureTemplateHandler{
		featureTemplateService: featureTemplateService,
	}
}

// GetFeatureTemplateList 查询功能模板列表
// @Summary 查询功能模板列表
// @Description 分页查询功能模板列表，支持关键词、分类与状态筛选
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1" minimum(1)
// @Param page_size query int false "每页条数，默认20，最大100" minimum(1) maximum(100)
// @Param keyword query string false "搜索关键词(支持名称、标识符、描述)"
// @Param category query string false "分类筛选"
// @Param status query string false "状态筛选" Enums(active, inactive)
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplateListResponse} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-templates [get]
func (h *FeatureTemplateHandler) GetFeatureTemplateList(c *gin.Context) {
	var req models.FeatureTemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.GetFeatureTemplateList(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetFeatureTemplate 获取功能模板详情
// @Summary 获取功能模板详情
// @Description 根据功能模板ID获取详细信息
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "功能模板ID"
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplate} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "功能模板不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-templates/{id} [get]
func (h *FeatureTemplateHandler) GetFeatureTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.GetFeatureTemplate(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateFeatureTemplate 创建功能模板
// @Summary 创建功能模板
// @Description 创建功能模板，标识符留空时由服务端按名称生成
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.FeatureTemplateCreateRequest true "功能模板信息"
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplate} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-templates [post]
func (h *FeatureTemplateHandler) CreateFeatureTemplate(c *gin.Context) {
	var req models.FeatureTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.CreateFeatureTemplate(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateFeatureTemplate 更新功能模板
// @Summary 更新功能模板
// @Description 更新功能模板信息，标识符不可修改
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "功能模板ID"
// @Param request body models.FeatureTemplateUpdateRequest true "功能模板信息"
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplate} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "功能模板不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-templates/{id} [put]
func (h *FeatureTemplateHandler) UpdateFeatureTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.FeatureTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.UpdateFeatureTemplate(ctx, id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// DeleteFeatureTemplate 删除功能模板
// @Summary 删除功能模板
// @Description 删除功能模板，内置模板不可删除
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "功能模板ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "功能模板不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-templates/{id} [delete]
func (h *FeatureTemplateHandler) DeleteFeatureTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.featureTemplateService.DeleteFeatureTemplate(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// GetCategoryList 查询功能模板分类列表
// @Summary 查询分类列表
// @Description 查询功能模板分类，含各分类下的模板数量
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplateCategoryListResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-templates/categories [get]
func (h *FeatureTemplateHandler) GetCategoryList(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.GetCategoryList(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateCategory 新建功能模板分类
// @Summary 新建分类
// @Description 新建自定义功能模板分类
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "分类信息，含value与label"
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplateCategory} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-template-categories [post]
func (h *FeatureTemplateHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Value     string `json:"value" binding:"required"`
		Label     string `json:"label" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.CreateCategory(ctx, &models.FeatureTemplateCategory{
		Value:     req.Value,
		Label:     req.Label,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateCategory 更新功能模板分类
// @Summary 更新分类
// @Description 更新自定义功能模板分类，内置分类不可修改
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "分类ID"
// @Param request body map[string]interface{} true "分类信息，含label与sort_order"
// @Success 200 {object} models.APIResponse{data=models.FeatureTemplateCategory} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "分类不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-template-categories/{id} [put]
func (h *FeatureTemplateHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req struct {
		Label     string `json:"label"`
		SortOrder *int   `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.featureTemplateService.UpdateCategory(ctx, id, req.Label, req.SortOrder)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// DeleteCategory 删除功能模板分类
// @Summary 删除分类
// @Description 删除自定义分类，分类下仍有模板时不可删除
// @Tags 功能模板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "分类ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "分类不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/feature-template-categories/{id} [delete]
func (h *FeatureTemplateHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.featureTemplateService.DeleteCategory(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}
