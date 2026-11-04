package middleware

import (
	"net/http"
	"strings"
	"time"

	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
	"license-manager/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

// getAuthTimestamp 当前时间戳（RFC3339）
func getAuthTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// abortAuthFailure 认证失败统一响应：码取自 i18n 目录（三语齐备），消息随请求语言本地化。
//
// 旧实现自造 AUTH_001~004 从未进 i18n 目录，且 timestamp 硬编码常量，
// 导致认证失败永远中文、时间戳不真实。
func abortAuthFailure(c *gin.Context, httpStatus int, i18nCode string) {
	lang := GetLanguage(c)
	_, code, message := i18n.NewI18nErrorResponse(i18nCode, lang)
	c.JSON(httpStatus, gin.H{
		"code":      code,
		"message":   message,
		"timestamp": getAuthTimestamp(),
	})
	c.Abort()
}

// AuthMiddleware 认证中间件（支持自动延长）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// API Key 中间件已先行认证通过时不再要求 JWT：
		// 两种凭据各自独立，不该要求调用方同时具备两者
		if authType, exists := c.Get("auth_type"); exists && authType == "api_key" {
			c.Next()
			return
		}
		authorizationHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			abortAuthFailure(c, http.StatusUnauthorized, "100004") // 认证信息缺失
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			abortAuthFailure(c, http.StatusUnauthorized, "100002") // 认证格式无效
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			abortAuthFailure(c, http.StatusUnauthorized, "100002") // 不支持的认证类型
			return
		}

		accessToken := fields[1]
		claims, err := utils.ValidateToken(accessToken)
		if err != nil {
			switch err {
			case utils.ErrTokenExpired:
				abortAuthFailure(c, http.StatusUnauthorized, "100001") // 令牌已过期
			default:
				// 格式错误 / 无效 / 其他一律归入 100002
				abortAuthFailure(c, http.StatusUnauthorized, "100002")
			}
			return
		}

		// 检查是否需要自动延长token（单令牌延长机制）
		if newToken, err := utils.RefreshTokenIfNeeded(accessToken); err == nil && newToken != accessToken {
			// 自动延长token有效期，后端直接使用新token，客户端无需感知
			accessToken = newToken
			// 重新解析claims
			newClaims, parseErr := utils.ParseToken(newToken)
			if parseErr == nil {
				claims = newClaims
			}
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set(AuthorizationPayloadKey, claims)

		// 同步写入请求 context：handler 侧经 WithLanguage 包装后
		// ctx 不再是 *gin.Context，服务层只能从这里取值
		ctx := pkgcontext.WithUserID(c.Request.Context(), claims.UserID)
		ctx = pkgcontext.WithUsername(ctx, claims.Username)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			c.Next()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.Next()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			c.Next()
			return
		}

		accessToken := fields[1]
		claims, err := utils.ValidateToken(accessToken)
		if err == nil {
			// Token有效，设置用户信息
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set(AuthorizationPayloadKey, claims)

			ctx := pkgcontext.WithUserID(c.Request.Context(), claims.UserID)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

// AdminOnlyMiddleware 仅超级管理员访问中间件
//
// 内置角色为 super_admin / operator / viewer（见 database/rbac_seed.go），
// 旧实现比对的 "administrator" 是社区版遗留值，对现有角色恒 403，从未生效。
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			abortAuthFailure(c, http.StatusUnauthorized, "100004") // 认证信息缺失
			return
		}

		if role != "super_admin" {
			abortAuthFailure(c, http.StatusForbidden, "100005") // 权限不足
			return
		}

		c.Next()
	}
}
