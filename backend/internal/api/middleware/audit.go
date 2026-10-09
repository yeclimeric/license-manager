package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"license-manager/internal/models"
	pkgcontext "license-manager/pkg/context"
)

// auditBodyLimit 单次审计记录的请求/响应体最大保留字节数。
// 超长内容截断，避免日志表被大 payload 撑大。
const auditBodyLimit = 4096

// AuditConfig 审计中间件配置
type AuditConfig struct {
	DB *gorm.DB
}

// AuditMiddleware 写操作审计中间件
//
// 只记录 POST/PUT/PATCH/DELETE，读操作不记录。审计写入是旁路操作：
// 无论记录成功与否都不影响业务响应，因此错误仅记日志不阻断请求。
func AuditMiddleware(cfg AuditConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.DB == nil || !isWriteMethod(c.Request.Method) {
			c.Next()
			return
		}

		// 读取请求体供审计留存，之后复位供业务处理
		var reqBody string
		if c.Request.Body != nil {
			if body, err := io.ReadAll(c.Request.Body); err == nil {
				reqBody = truncateAuditBody(string(body))
				c.Request.Body = io.NopCloser(bytes.NewReader(body))
			}
		}

		start := time.Now()
		c.Next()
		elapsed := time.Since(start)

		// 若响应体未被审计包装器捕获（中间件顺序变化），仍记录其余字段
		respBody := ""
		if captured, ok := c.Get(auditResponseKey); ok {
			switch v := captured.(type) {
			case *bytes.Buffer:
				respBody = v.String()
			case string:
				respBody = v
			}
		}

		record := buildAuditRecord(c, reqBody, respBody, elapsed)
		if record == nil {
			return
		}

		if err := cfg.DB.Create(record).Error; err != nil {
			// 审计失败不阻断业务，仅记录到请求上下文
			c.Error(err) //nolint:errcheck
		}
	}
}

// auditResponseKey 响应体在 gin.Context 中的存储键
const auditResponseKey = "audit_response_body"

// AuditResponseWriter 包装 gin.ResponseWriter 以捕获响应体，
// 供审计中间件读取。写入行为与原 writer 完全一致。
type AuditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 写入响应并记录内容
func (w *AuditResponseWriter) Write(data []byte) (int, error) {
	if w.body != nil {
		w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

// WriteString 写入字符串响应并记录内容
func (w *AuditResponseWriter) WriteString(s string) (int, error) {
	if w.body != nil {
		w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

// captureResponseBody 包装当前 writer 以捕获响应体。
// 需在业务处理前调用，因此单独提供供路由层挂载。
func captureResponseBody(c *gin.Context) {
	body := &bytes.Buffer{}
	c.Writer = &AuditResponseWriter{ResponseWriter: c.Writer, body: body}
	c.Set(auditResponseKey, body)
}

// AuditResponseCapture 响应体捕获中间件。
//
// 与 AuditMiddleware 配套使用，须挂在写接口之前：
// 审计需要看到响应体才能判定业务成功与否。
func AuditResponseCapture() gin.HandlerFunc {
	return func(c *gin.Context) {
		captureResponseBody(c)
		c.Next()
	}
}

// isWriteMethod 是否为写操作
func isWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// buildAuditRecord 构建审计记录
func buildAuditRecord(c *gin.Context, reqBody string, respBody string, elapsed time.Duration) *models.AuditLog {
	resourceType, resourceID := resolveAuditResource(c)

	record := &models.AuditLog{
		UserID:       pkgcontext.GetUserIDFromContext(c.Request.Context()),
		Username:     c.GetString("username"),
		Action:       c.Request.Method + " " + c.Request.URL.Path,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		Status:       models.AuditStatusSuccess,
	}

	// 业务错误码以 4xx/5xx 之外的形式返回（响应体带 code），因此按响应体判断
	detail := models.AuditLogDetail{
		Path:     c.Request.URL.Path,
		Method:   c.Request.Method,
		ReqBody:  reqBody,
		RespBody: truncateAuditBody(respBody),
	}

	if detail.RespBody != "" {
		detail.RespID = extractRespID(detail.RespBody)
		if isFailureResponse(detail.RespBody) {
			record.Status = models.AuditStatusFailure
		}
	}

	// 状态码异常（鉴权失败等）同样记为失败
	if c.Writer.Status() >= http.StatusBadRequest {
		record.Status = models.AuditStatusFailure
	}

	// 授权码列表带 customerId 查询参数，便于按客户追溯
	if productID := c.Query("product_id"); productID != "" {
		record.ProductID = &productID
	}

	if data, err := json.Marshal(detail); err == nil {
		record.Detail = models.JSON(data)
	}

	return record
}

// resolveAuditResource 从路径推断资源类型与资源ID
//
// 路径形如 /api/v1/authorization-codes/{id} 或
// /api/v1/authorization-codes/{id}/lock，需要跳过 uuid 段与子动作段。
func resolveAuditResource(c *gin.Context) (resourceType string, resourceID string) {
	segments := splitPath(c.Request.URL.Path)

	// 去掉 api 与版本前缀，形如 /api/v1/xxx 或 /v1/xxx
	if len(segments) >= 2 && segments[0] == "api" {
		segments = segments[2:]
	} else if len(segments) >= 1 && segments[0] == "api" {
		segments = segments[1:]
	}
	// 剩余首段若为版本号（v1/v2 等）同样跳过
	if len(segments) >= 1 && isVersionSegment(segments[0]) {
		segments = segments[1:]
	}
	if len(segments) == 0 {
		return "", ""
	}

	resourceType = segments[0]

	// 下一段若是 UUID 形态则为资源ID，否则视为子动作（如 batch）
	if len(segments) >= 2 && isUUIDLike(segments[1]) {
		resourceID = segments[1]
	}

	return resourceType, resourceID
}

// splitPath 拆分并清理路径
func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	return cleaned
}

// isVersionSegment 判断是否为版本号路径段（v1、v2 等）
func isVersionSegment(segment string) bool {
	if len(segment) < 2 || (segment[0] != 'v' && segment[0] != 'V') {
		return false
	}
	for i := 1; i < len(segment); i++ {
		if segment[i] < '0' || segment[i] > '9' {
			return false
		}
	}
	return true
}

// isUUIDLike 判断是否形如 UUID
func isUUIDLike(segment string) bool {
	if len(segment) != 36 {
		return false
	}
	for i, c := range segment {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !isHex {
			return false
		}
	}
	return true
}

// isFailureResponse 判断响应体是否为业务失败
func isFailureResponse(respBody string) bool {
	if respBody == "" {
		return false
	}
	var envelope struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(respBody), &envelope); err != nil {
		return false
	}
	// 000000 为成功码，其余视为失败
	return envelope.Code != "" && envelope.Code != "000000"
}

// extractRespID 从响应体提取新建资源的ID
func extractRespID(respBody string) string {
	var envelope struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &envelope); err != nil {
		return ""
	}
	return envelope.Data.ID
}

// truncateAuditBody 截断过长的审计内容
func truncateAuditBody(body string) string {
	if len(body) <= auditBodyLimit {
		return body
	}
	return body[:auditBodyLimit] + "...[truncated]"
}
