package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// 认证失败响应的稳定形状。旧实现是 {code:int(401), error:"AUTH_00x", message:中文, timestamp:硬编码}，
// 三语语言包从未收录 AUTH_ 码，导致认证失败永远中文且时间戳恒为 2024-07-30T12:00:00Z。
type authFailureBody struct {
	Code      interface{} `json:"code"`
	Error     *string     `json:"error"`
	Message   string      `json:"message"`
	Timestamp string      `json:"timestamp"`
}

func newAuthTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(I18nMiddleware(&I18nConfig{Enable: true, DefaultLang: "zh-CN", SupportLangs: []string{"zh-CN", "en-US", "ja-JP"}}))
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// 缺令牌：码必须是 i18n 目录中的字符串码，不再自造 AUTH_001。
func TestAuthFailureUsesI18nCode(t *testing.T) {
	r := newAuthTestEngine()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	var body authFailureBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %s", w.Body.String())
	}
	if code, ok := body.Code.(string); !ok || code != "100004" {
		t.Errorf("code = %v (%T), want string 100004（i18n 认证码，非 int 401 / 非 AUTH_001）", body.Code, body.Code)
	}
	if body.Error != nil {
		t.Errorf("response should not carry legacy error field, got %q", *body.Error)
	}
	if body.Message == "" {
		t.Error("message should not be empty")
	}
}

// 回归：timestamp 必须是真实 RFC3339 时间，不再是硬编码常量。
func TestAuthFailureTimestampIsReal(t *testing.T) {
	r := newAuthTestEngine()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

	var body authFailureBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %s", w.Body.String())
	}
	const legacy = "2024-07-30T12:00:00Z"
	if body.Timestamp == legacy {
		t.Fatalf("timestamp is still the hardcoded legacy constant %q", legacy)
	}
	parsed, err := time.Parse(time.RFC3339, body.Timestamp)
	if err != nil {
		t.Fatalf("timestamp %q is not valid RFC3339: %v", body.Timestamp, err)
	}
	if time.Since(parsed) > time.Minute || time.Since(parsed) < -time.Minute {
		t.Errorf("timestamp %q is not close to now", body.Timestamp)
	}
}

// 无效 Bearer 类型归入 100002（认证无效），而非硬编码中文。
func TestAuthFailureInvalidBearerType(t *testing.T) {
	r := newAuthTestEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic abc")
	r.ServeHTTP(w, req)

	var body authFailureBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %s", w.Body.String())
	}
	if code, ok := body.Code.(string); !ok || code != "100002" {
		t.Errorf("code = %v, want string 100002（认证无效）", body.Code)
	}
}

// AdminOnly 权限不足用 100005（i18n 权限码），不再是 AUTH_004。
func TestAdminOnlyUsesI18nCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("role", "viewer"); c.Next() })
	r.Use(AdminOnlyMiddleware())
	r.GET("/admin", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	var body authFailureBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %s", w.Body.String())
	}
	if code, ok := body.Code.(string); !ok || code != "100005" {
		t.Errorf("code = %v, want string 100005（权限不足）", body.Code)
	}
}
