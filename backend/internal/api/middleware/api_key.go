package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	pkgcontext "license-manager/pkg/context"
)

// APIKeyChecker API Key 校验函数，由路由层注入。
type APIKeyChecker func(ctx *gin.Context, plaintext string) (userID string, allowed bool, err error)

// APIKeyAuthMiddleware API Key 认证中间件。
//
// 与登录态并行而非互斥：既支持 Authorization: Bearer <jwt>，
// 也支持 X-API-Key 头。两者的密钥形态不同（JWT 有三段、API Key 无点号），
// 据此区分，避免让调用方声明自己用的是哪一种。
//
// API Key 通过后按该密钥所属用户注入 user_id，
// 因此受保护的接口无需为两种认证方式各写一套分支。
func APIKeyAuthMiddleware(checker APIKeyChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		plaintext := extractAPIKey(c)
		if plaintext == "" {
			// 没有 API Key 时不拦截，交由后续的登录态中间件处理
			c.Next()
			return
		}

		if checker == nil {
			abortForbidden(c, "100005")
			return
		}

		userID, allowed, err := checker(c, plaintext)
		if err != nil {
			abortForbidden(c, "100005")
			return
		}
		if !allowed {
			abortForbidden(c, "100007")
			return
		}

		c.Set("auth_type", "api_key")
		if userID != "" {
			c.Set("user_id", userID)
			c.Request = c.Request.WithContext(pkgcontext.WithUserID(c.Request.Context(), userID))
		}
		c.Next()
	}
}

// extractAPIKey 从请求中提取 API Key 明文。
//
// 只接受 X-API-Key 头：Authorization 承载登录态令牌，
// 混用会让「拿到 JWT 的人是不是也能拿到 API Key」变得难以判断。
func extractAPIKey(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader("X-API-Key"))
}

// HashAPIKey 计算 API Key 哈希，与写入时口径一致
func HashAPIKey(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// NewGormAPIKeyChecker 基于数据库的 API Key 校验实现。
//
// 校验顺序：存在 → 状态启用 → 未过期 → 来源 IP 在白名单内。
// 过期是时间的函数，每次请求现算，不依赖定时任务改状态字段。
func NewGormAPIKeyChecker(db *gorm.DB, checker SystemConfigChecker) APIKeyChecker {
	return func(c *gin.Context, plaintext string) (string, bool, error) {
		if checker == nil {
			return "", false, nil
		}

		userID, allowed, err := checker.VerifyAPIKey(c.Request.Context(), plaintext,
			HashAPIKey(plaintext), c.ClientIP())
		if err != nil {
			return "", false, err
		}
		return userID, allowed, nil
	}
}

// SystemConfigChecker 配置仓储的最小依赖，避免中间件依赖具体实现
type SystemConfigChecker interface {
	VerifyAPIKey(ctx context.Context, plaintext, hash, clientIP string) (userID string, allowed bool, err error)
}
