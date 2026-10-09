package middleware

import (
	"net/http"
	"strings"

	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

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
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"error":     "AUTH_001",
				"message":   "缺少认证令牌",
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"error":     "AUTH_002",
				"message":   "认证令牌格式无效",
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"error":     "AUTH_002",
				"message":   "不支持的认证类型",
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		accessToken := fields[1]
		claims, err := utils.ValidateToken(accessToken)
		if err != nil {
			var errorCode, message string
			switch err {
			case utils.ErrTokenExpired:
				errorCode = "AUTH_003"
				message = "令牌已过期"
			case utils.ErrTokenMalformed:
				errorCode = "AUTH_002"
				message = "令牌格式错误"
			default:
				errorCode = "AUTH_002"
				message = "令牌无效"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"error":     errorCode,
				"message":   message,
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
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

// AdminOnlyMiddleware 仅管理员访问中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":      http.StatusUnauthorized,
				"error":     "AUTH_001",
				"message":   "认证信息缺失",
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		if role != "administrator" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":      http.StatusForbidden,
				"error":     "AUTH_004",
				"message":   "权限不足",
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return "2024-07-30T12:00:00Z" // 简化实现，实际应该使用time.Now().Format(time.RFC3339)
}
