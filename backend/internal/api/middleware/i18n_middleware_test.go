package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgcontext "license-manager/pkg/context"

	"github.com/gin-gonic/gin"
)

// newI18nTestEngine 装配仅含 I18nMiddleware 的测试路由，
// 最终处理器把 request context 与 gin 键里的语言各读一次，供断言。
func newI18nTestEngine(t *testing.T, cfg *I18nConfig) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(I18nMiddleware(cfg))
	r.GET("/probe", func(c *gin.Context) {
		ctxLang := pkgcontext.GetLanguageFromContext(c.Request.Context())
		ginKeyLang := GetLanguage(c)
		c.JSON(http.StatusOK, gin.H{"ctx_lang": ctxLang, "gin_key_lang": ginKeyLang})
	})
	return r
}

// 回归：I18nMiddleware 必须把语言写进 request context（不只是 gin 键）。
// 服务层拿到的是 ctx.Request.Context()，gin 对非 string 键的 Value 不做兜底，
// 只有 request context 才是服务层取语言的唯一可靠通道。
func TestI18nMiddlewareWritesRequestContext(t *testing.T) {
	r := newI18nTestEngine(t, &I18nConfig{Enable: true, DefaultLang: "zh-CN", SupportLangs: []string{"zh-CN", "en-US", "ja-JP"}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe?lang=en-US", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if got := parseProbe(w.Body.String()); got.CtxLang != "en-US" {
		t.Errorf("request context language = %q, want en-US（服务层 i18n 依赖此值）", got.CtxLang)
	}
	if got := parseProbe(w.Body.String()); got.GinKeyLang != "en-US" {
		t.Errorf("gin key language = %q, want en-US", got.GinKeyLang)
	}
}

// 无 ?lang 时落到默认语言，且同样双写。
func TestI18nMiddlewareDefaultsWhenNoLangParam(t *testing.T) {
	r := newI18nTestEngine(t, &I18nConfig{Enable: true, DefaultLang: "ja-JP", SupportLangs: []string{"ja-JP"}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	r.ServeHTTP(w, req)

	if got := parseProbe(w.Body.String()); got.CtxLang != "ja-JP" {
		t.Errorf("request context language = %q, want ja-JP（默认）", got.CtxLang)
	}
}

// 回归：config 为 nil 不再 panic（原实现 i18n.go:31-35 会解引用 nil）。
func TestI18nMiddlewareNilConfigDoesNotPanic(t *testing.T) {
	r := newI18nTestEngine(t, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("nil config should not panic, got %d", w.Code)
	}
}

// 回归：Accept-Language 头参与语言判定。
func TestI18nMiddlewareRespectsAcceptLanguage(t *testing.T) {
	r := newI18nTestEngine(t, &I18nConfig{Enable: true, DefaultLang: "zh-CN", SupportLangs: []string{"zh-CN", "en-US", "ja-JP"}})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Accept-Language", "ja-JP")
	r.ServeHTTP(w, req)

	if got := parseProbe(w.Body.String()); got.CtxLang != "ja-JP" {
		t.Errorf("request context language = %q, want ja-JP（Accept-Language）", got.CtxLang)
	}
}

type probe struct {
	CtxLang   string `json:"ctx_lang"`
	GinKeyLang string `json:"gin_key_lang"`
}

func parseProbe(body string) probe {
	var p probe
	_ = json.Unmarshal([]byte(body), &p)
	return p
}
