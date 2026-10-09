package handlers

import (
	"errors"
	"net/http"
	"time"

	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/internal/service"
	"license-manager/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type ProductOfferingHandler struct {
	productOfferingService service.ProductOfferingService
}

// NewProductOfferingHandler 创建产品项处理器
func NewProductOfferingHandler(productOfferingService service.ProductOfferingService) *ProductOfferingHandler {
	return &ProductOfferingHandler{
		productOfferingService: productOfferingService,
	}
}

// respondError 输出错误响应：I18nError 直接用，其余走兜底
func respondError(c *gin.Context, err error) {
	var i18nErr *i18n.I18nError
	if errors.As(err, &i18nErr) {
		c.JSON(i18nErr.HttpCode, models.ErrorResponse{
			Code:      i18nErr.Code,
			Message:   i18nErr.Message,
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	lang := middleware.GetLanguage(c)
	status, errCode, message := i18n.NewI18nErrorResponse("900004", lang)
	c.JSON(status, models.ErrorResponse{
		Code:      errCode,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// respondSuccess 输出成功响应
func respondSuccess(c *gin.Context, data interface{}) {
	lang := middleware.GetLanguage(c)
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    "000000",
		Message: i18n.GetErrorMessage("000000", lang),
		Data:    data,
	})
}

// respondErrorWithCode 输出指定错误码的响应
func respondErrorWithCode(c *gin.Context, code string) {
	lang := middleware.GetLanguage(c)
	status, errCode, message := i18n.NewI18nErrorResponse(code, lang)
	c.JSON(status, models.ErrorResponse{
		Code:      errCode,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetProductOfferingList 查询产品项列表
// @Summary 查询产品项列表
// @Description 分页查询产品项列表，支持关键词与状态筛选
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认1" minimum(1)
// @Param page_size query int false "每页条数，默认20，最大100" minimum(1) maximum(100)
// @Param keyword query string false "搜索关键词(支持名称、别名、编码)"
// @Param status query string false "状态筛选" Enums(active, inactive)
// @Param sort_by query string false "排序字段，默认created_at" Enums(created_at, updated_at, name, code)
// @Param sort_order query string false "排序方向，默认desc" Enums(asc, desc)
// @Success 200 {object} models.APIResponse{data=models.ProductOfferingListResponse} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings [get]
func (h *ProductOfferingHandler) GetProductOfferingList(c *gin.Context) {
	var req models.ProductOfferingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.productOfferingService.GetProductOfferingList(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetProductOffering 获取产品项详情
// @Summary 获取产品项详情
// @Description 根据产品项ID获取详细信息
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "产品项ID"
// @Success 200 {object} models.APIResponse{data=models.ProductOfferingResponse} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "产品项不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings/{id} [get]
func (h *ProductOfferingHandler) GetProductOffering(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.productOfferingService.GetProductOffering(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateProductOffering 创建产品项
// @Summary 创建产品项
// @Description 创建产品项，产品项编码由系统自动生成
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ProductOfferingCreateRequest true "产品项信息"
// @Success 200 {object} models.APIResponse{data=models.ProductOfferingResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings [post]
func (h *ProductOfferingHandler) CreateProductOffering(c *gin.Context) {
	var req models.ProductOfferingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.productOfferingService.CreateProductOffering(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateProductOffering 更新产品项
// @Summary 更新产品项
// @Description 更新产品项信息，产品项编码不可修改
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "产品项ID"
// @Param request body models.ProductOfferingUpdateRequest true "产品项信息"
// @Success 200 {object} models.APIResponse{data=models.ProductOfferingResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "产品项不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings/{id} [put]
func (h *ProductOfferingHandler) UpdateProductOffering(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.ProductOfferingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.productOfferingService.UpdateProductOffering(ctx, id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetProductOfferingPublicKey 获取产品项公钥
// @Summary 获取产品公钥
// @Description 获取产品项的签名公钥，客户端用于验证许可证签名
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "产品项ID"
// @Success 200 {object} models.APIResponse{data=models.ProductOfferingPublicKeyResponse} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "产品项不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings/{id}/public-key [get]
func (h *ProductOfferingHandler) GetProductOfferingPublicKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.productOfferingService.GetProductOfferingPublicKey(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// DeleteProductOffering 删除产品项
// @Summary 删除产品项
// @Description 删除产品项，产品项下存在授权码时不可删除
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "产品项ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "产品项不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings/{id} [delete]
func (h *ProductOfferingHandler) DeleteProductOffering(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.productOfferingService.DeleteProductOffering(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// UpdateProductOfferingStatus 更新产品项状态
// @Summary 更新产品项状态
// @Description 启用或停用产品项
// @Tags 产品项管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "产品项ID"
// @Param request body map[string]string true "状态，active或inactive"
// @Success 200 {object} models.APIResponse{data=models.ProductOfferingResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "产品项不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/product-offerings/{id}/status [patch]
func (h *ProductOfferingHandler) UpdateProductOfferingStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.productOfferingService.UpdateProductOfferingStatus(ctx, id, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}
