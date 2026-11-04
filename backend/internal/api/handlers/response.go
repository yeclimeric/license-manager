package handlers

import (
	"errors"
	"net/http"
	"time"

	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// 错误/成功响应的单一真源。
//
// 历史上有 4 套并行实现：定义在 product_offering_handler.go 的 respond*、
// dashboard_handler.go 的 respondDashboard*、error_handler.go 的
// handleI18nError（用类型断言、且从未被调用），以及散落在 8 个文件里的
// models.ErrorResponse 字面量。后两类会绕过统一的错误码/语言/时间戳逻辑。
//
// 现在把函数实现集中到本文件；dashboard 与死代码改为复用这里。
// 业务字面量暂不强行替换：部分调用点会把 err.Error() 拼进 message，
// 直接替换会改变响应内容，需单独评估。

// respondError 输出错误响应：I18nError 直接用，其余走兜底。
// 用 errors.As 而非类型断言，包装过的 error 也能正确识别。
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
