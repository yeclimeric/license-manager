package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
)

// PermissionChecker 权限校验函数，由路由层注入 RBAC 服务实现。
// 与中间件解耦，避免 middleware 反向依赖 service。
type PermissionChecker func(c *gin.Context, permission string) (bool, error)

// RequirePermission 权限校验中间件
//
// 按 RBAC 角色权限校验，权限码形如 resource:action。
// 持通配权限（超级管理员）者放行。
//
// 校验失败返回 403：错误码在响应体中，与项目既有的业务错误约定一致。
func RequirePermission(permission string, checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 无校验器时不拦截，避免因装配缺失导致接口不可用
		if checker == nil {
			c.Next()
			return
		}

		userID := currentUserID(c)
		if userID == "" {
			abortForbidden(c, "100004")
			return
		}

		allowed, err := checker(c, permission)
		if err != nil {
			// 校验过程出错时按拒绝处理，不因鉴权故障放行
			abortForbidden(c, "100005")
			return
		}

		if !allowed {
			abortForbidden(c, "100006")
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 多权限任一满足即可
func RequireAnyPermission(permissions []string, checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checker == nil {
			c.Next()
			return
		}

		userID := currentUserID(c)
		if userID == "" {
			abortForbidden(c, "100004")
			return
		}

		for _, permission := range permissions {
			allowed, err := checker(c, permission)
			if err != nil {
				continue
			}
			if allowed {
				c.Next()
				return
			}
		}

		abortForbidden(c, "100006")
	}
}

// currentUserID 从 gin 上下文取当前用户ID。
//
// 鉴权中间件已把用户ID写入请求 context，此处优先从 context 取，
// 退回 gin 存储以兼容未经包装的调用路径。
func currentUserID(c *gin.Context) string {
	if userID := pkgcontext.GetUserIDFromContext(c.Request.Context()); userID != "" {
		return userID
	}
	if value, exists := c.Get("user_id"); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

// abortForbidden 中止请求并返回权限错误
func abortForbidden(c *gin.Context, i18nCode string) {
	lang := GetLanguage(c)
	status, code, message := i18n.NewI18nErrorResponse(i18nCode, lang)
	c.JSON(status, gin.H{
		"code":      code,
		"message":   message,
		"timestamp": getCurrentTimestamp(),
	})
	c.Abort()
}

// NewGormPermissionChecker 基于数据库的权限校验实现。
//
// 作为兜底：路由层未注入 RBAC 服务时使用，直接查关联表判断。
func NewGormPermissionChecker(db *gorm.DB) PermissionChecker {
	return func(c *gin.Context, permission string) (bool, error) {
		if permission == "" {
			return true, nil
		}

		userID := currentUserID(c)
		if userID == "" {
			return false, nil
		}

		var count int64
		err := db.WithContext(c.Request.Context()).
			Table("role_permissions rp").
			Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
			Where("ur.user_id = ?", userID).
			Where("rp.permission IN ?", []string{permission, "*"}).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}
}
