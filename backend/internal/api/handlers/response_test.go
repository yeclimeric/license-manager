package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"license-manager/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type responseBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}

func newResponseTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/probe", nil)
	return c, w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) responseBody {
	t.Helper()
	var body responseBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON %q: %v", w.Body.String(), err)
	}
	return body
}

// 回归：respondError 用 errors.As，包装过的 I18nError 仍应被识别，
// 而不是退化成 500/900004。原 handleI18nError 用类型断言做不到这点。
func TestRespondErrorUnwrapsI18nError(t *testing.T) {
	c, w := newResponseTestContext()
	inner := i18n.NewI18nError("300006", "zh-CN")
	wrapped := fmt.Errorf("repository layer: %w", inner)

	respondError(c, wrapped)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (from 300006)", w.Code)
	}
	body := decodeBody(t, w)
	if body.Code != "300006" {
		t.Errorf("code = %q, want 300006（不得退化为 900004）", body.Code)
	}
	if body.Timestamp == "" {
		t.Error("timestamp should be set")
	}
}

// 非 I18nError 走 900004 兜底，状态码由码表决定（500）。
func TestRespondErrorFallback(t *testing.T) {
	c, w := newResponseTestContext()
	respondError(c, fmt.Errorf("plain db failure"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if body := decodeBody(t, w); body.Code != "900004" {
		t.Errorf("code = %q, want 900004", body.Code)
	}
}

// 指定错误码时走统一状态码映射，不硬编码。
func TestRespondErrorWithCodeUsesCodeTable(t *testing.T) {
	c, w := newResponseTestContext()
	respondErrorWithCode(c, "900001")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (from 900001)", w.Code)
	}
	if body := decodeBody(t, w); body.Code != "900001" {
		t.Errorf("code = %q, want 900001", body.Code)
	}
}

func TestRespondSuccessShape(t *testing.T) {
	c, w := newResponseTestContext()
	respondSuccess(c, map[string]int{"n": 1})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := decodeBody(t, w)
	if body.Code != "000000" {
		t.Errorf("code = %q, want 000000", body.Code)
	}
	if body.Message == "" {
		t.Error("success message should not be empty")
	}
}
