package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// 权限门在异常情况下必须拒绝而非放行。
//
// 鉴权出故障时若选择放行，等于把「服务不可用」变成「安全门失效」。
func TestRequirePermissionFailClosed(t *testing.T) {
	cases := []struct {
		name    string
		checker PermissionChecker
	}{
		{
			name:    "校验器报错",
			checker: func(c *gin.Context, p string) (bool, error) { return false, errors.New("db down") },
		},
		{
			name:    "校验器明确拒绝",
			checker: func(c *gin.Context, p string) (bool, error) { return false, nil },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRouter()
			reached := false
			r.Use(func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() })
			r.GET("/x", RequirePermission("user:read", tc.checker), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

			if reached {
				t.Error("鉴权未通过时不应到达处理器")
			}
			if w.Code == http.StatusOK {
				t.Error("鉴权未通过时不应返回 200")
			}
		})
	}
}

// 有权限时正常放行
func TestRequirePermissionAllows(t *testing.T) {
	r := newTestRouter()
	reached := false
	r.Use(func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() })
	r.GET("/x", RequirePermission("user:read", func(c *gin.Context, p string) (bool, error) {
		return true, nil
	}), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if !reached {
		t.Error("有权限时应放行")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// 校验器缺省时放行。
//
// 这是有意的取舍：装配遗漏时若一律拒绝，整个服务会不可用，
// 那比单点失守更难发现。因此此处放行，由装配测试保证校验器确实被注入。
func TestRequirePermissionNilCheckerPasses(t *testing.T) {
	r := newTestRouter()
	reached := false
	r.Use(func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() })
	r.GET("/x", RequirePermission("user:read", nil), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if !reached {
		t.Error("校验器缺省时应放行（避免装配遗漏导致服务不可用）")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// 未登录（无 user_id）时不得通过权限校验
func TestRequirePermissionWithoutUserRejects(t *testing.T) {
	r := newTestRouter()
	reached := false
	r.GET("/x", RequirePermission("user:read", func(c *gin.Context, p string) (bool, error) {
		return true, nil
	}), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if reached {
		t.Error("未登录时应拒绝")
	}
}

// 空权限码会交给校验器判断，行为取决于校验器如何解释。
//
// 这里钉住现状而非期望值：若将来改为「空码即放行」，此用例会提示同步更新。
func TestRequirePermissionEmptyCodeDelegatesToChecker(t *testing.T) {
	r := newTestRouter()
	var gotPermission string
	reached := false
	r.Use(func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() })
	r.GET("/x", RequirePermission("", func(c *gin.Context, p string) (bool, error) {
		gotPermission = p
		return true, nil
	}), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if gotPermission != "" {
		t.Errorf("校验器收到的权限码 = %q, want 空串", gotPermission)
	}
	if !reached {
		t.Error("校验器返回允许时应放行")
	}
}

// 多权限任一满足即放行
func TestRequireAnyPermission(t *testing.T) {
	cases := []struct {
		name    string
		granted bool
		wantOK  bool
	}{
		{name: "全部满足", granted: true, wantOK: true},
		{name: "全部不满足", granted: false, wantOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRouter()
			reached := false
			r.Use(func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() })
			r.GET("/x", RequireAnyPermission(
				[]string{"a:read", "b:read"},
				func(c *gin.Context, p string) (bool, error) { return tc.granted, nil },
			), func(c *gin.Context) {
				reached = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

			if reached != tc.wantOK {
				t.Errorf("reached = %v, want %v", reached, tc.wantOK)
			}
		})
	}
}

// 权限码列表为空时不应放行
func TestRequireAnyPermissionEmptyListRejects(t *testing.T) {
	r := newTestRouter()
	reached := false
	r.Use(func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() })
	r.GET("/x", RequireAnyPermission([]string{}, func(c *gin.Context, p string) (bool, error) {
		return true, nil
	}), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if reached {
		t.Error("权限码列表为空不应等同于放行")
	}
}

// currentUserID 从上下文读取用户ID
func TestCurrentUserID(t *testing.T) {
	cases := []struct {
		name   string
		userID string
		set    bool
		want   string
	}{
		{name: "存在用户ID", userID: "u-1", set: true, want: "u-1"},
		{name: "未设置", set: false, want: ""},
		{name: "空字符串", userID: "", set: true, want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
			if tc.set {
				c.Set("user_id", tc.userID)
			}

			if got := currentUserID(c); got != tc.want {
				t.Errorf("currentUserID() = %q, want %q", got, tc.want)
			}
		})
	}
}
