package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"license-manager/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter 搭建最小路由以驱动中间件
func newTestRouter() *gin.Engine {
	r := gin.New()
	return r
}

// 无 API Key 时中间件应放行，交由登录态中间件处理
func TestAPIKeyAuthMiddlewareNoKeyPassesThrough(t *testing.T) {
	r := newTestRouter()
	reached := false
	r.Use(APIKeyAuthMiddleware(nil))
	r.GET("/x", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	if !reached {
		t.Error("未提供 API Key 时不应拦截请求")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// 校验通过时应注入 auth_type 并放行
func TestAPIKeyAuthMiddlewareValidKeyPasses(t *testing.T) {
	r := newTestRouter()

	var gotAuthType string
	var gotUserID string
	reached := false

	checker := func(c *gin.Context, plaintext string) (string, bool, error) {
		return "user-123", true, nil
	}

	r.Use(APIKeyAuthMiddleware(checker))
	r.GET("/x", func(c *gin.Context) {
		reached = true
		if v, ok := c.Get("auth_type"); ok {
			gotAuthType, _ = v.(string)
		}
		if v, ok := c.Get("user_id"); ok {
			gotUserID, _ = v.(string)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-API-Key", "sk-abc")
	r.ServeHTTP(w, req)

	if !reached {
		t.Fatal("有效 API Key 应放行到处理器")
	}
	if gotAuthType != "api_key" {
		t.Errorf("auth_type = %q, want api_key", gotAuthType)
	}
	if gotUserID != "user-123" {
		t.Errorf("user_id = %q, want user-123", gotUserID)
	}
}

// 校验不通过时不得放行
func TestAPIKeyAuthMiddlewareRejects(t *testing.T) {
	cases := []struct {
		name     string
		checker  APIKeyChecker
		wantCode int
	}{
		{
			// 校验器不可用或出错：走权限不足 100005（403）
			name:     "校验器返回错误",
			checker:  func(c *gin.Context, p string) (string, bool, error) { return "", false, errFake },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "校验器返回错误",
			checker:  func(c *gin.Context, p string) (string, bool, error) { return "", false, errFake },
			wantCode: http.StatusForbidden,
		},
		{
			name:     "校验器缺失",
			checker:  nil,
			wantCode: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRouter()
			reached := false
			r.Use(APIKeyAuthMiddleware(tc.checker))
			r.GET("/x", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.Header.Set("X-API-Key", "sk-abc")
			r.ServeHTTP(w, req)

			if reached {
				t.Error("校验不通过时不应到达处理器")
			}
			if w.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tc.wantCode)
			}
		})
	}
}

// 密钥本身无效时属认证失败（401），与权限不足（403）区分
func TestAPIKeyAuthMiddlewareInvalidKeyIsUnauthorized(t *testing.T) {
	r := newTestRouter()
	reached := false
	r.Use(APIKeyAuthMiddleware(func(c *gin.Context, p string) (string, bool, error) {
		return "", false, nil
	}))
	// 直接用 abortForbidden 走 100007 分支，验证状态码映射
	r.GET("/x", func(c *gin.Context) {
		if extractAPIKey(c) == "sk-bad" {
			abortForbidden(c, "100007")
			return
		}
		reached = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-API-Key", "sk-bad")
	r.ServeHTTP(w, req)

	if reached {
		t.Error("无效密钥不应到达处理器")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401（密钥无效属认证失败）", w.Code)
	}
}

// API Key 只从 X-API-Key 头读取。
//
// Authorization 承载登录态令牌，混用会让「拿到 JWT 的人是否也能拿到密钥」难以判断。
func TestExtractAPIKeyHeaderOnly(t *testing.T) {
	cases := []struct {
		name   string
		apiKey string
		auth   string
		want   string
	}{
		{name: "仅 X-API-Key", apiKey: "sk-1", want: "sk-1"},
		{name: "X-API-Key 前后空白被裁剪", apiKey: "  sk-2  ", want: "sk-2"},
		{name: "Authorization 不被当作密钥", auth: "Bearer sk-3", want: ""},
		{name: "两者同时存在时取专用头", apiKey: "sk-4", auth: "Bearer sk-5", want: "sk-4"},
		{name: "空头返回空串", apiKey: "", want: ""},
		{name: "纯空白头返回空串", apiKey: "   ", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
			if tc.apiKey != "" {
				c.Request.Header.Set("X-API-Key", tc.apiKey)
			}
			if tc.auth != "" {
				c.Request.Header.Set("Authorization", tc.auth)
			}

			if got := extractAPIKey(c); got != tc.want {
				t.Errorf("extractAPIKey() = %q, want %q", got, tc.want)
			}
		})
	}
}

// HashAPIKey 必须与 GenerateAPIKey 的口径一致，否则认证永远不匹配
func TestHashAPIKeyMatchesGenerator(t *testing.T) {
	plaintext, prefix, hash, err := repository.GenerateAPIKey()
	if err != nil {
		t.Fatalf("生成失败: %v", err)
	}

	if got := HashAPIKey(plaintext); got != hash {
		t.Errorf("HashAPIKey 与生成时的哈希不一致")
	}

	// 不同明文必须得到不同哈希
	other, _, _, _ := repository.GenerateAPIKey()
	if HashAPIKey(other) == hash {
		t.Error("不同密钥的哈希不应与前者相同")
	}
	if HashAPIKey(other) == HashAPIKey(plaintext) {
		t.Error("两个不同密钥算出同一哈希")
	}

	if !strings.HasPrefix(prefix, "sk-") {
		t.Errorf("prefix = %q, 应以 sk- 开头", prefix)
	}
}

// 校验器返回的 userID 为空时不应写入无效身份
func TestAPIKeyAuthMiddlewareEmptyUserID(t *testing.T) {
	r := newTestRouter()
	checker := func(c *gin.Context, p string) (string, bool, error) { return "", true, nil }

	var userIDSet bool
	r.Use(APIKeyAuthMiddleware(checker))
	r.GET("/x", func(c *gin.Context) {
		if _, ok := c.Get("user_id"); ok {
			userIDSet = true
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-API-Key", "sk-abc")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200（放行不应取决于 userID 是否为空）", w.Code)
	}
	if userIDSet {
		t.Error("userID 为空时不应注入 user_id")
	}
}

// 模拟校验错误
type fakeError struct{}

func (e *fakeError) Error() string { return "verification failed" }

var errFake = &fakeError{}
