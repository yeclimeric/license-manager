package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResolveAuditResource(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantType string
		wantID   string
	}{
		{
			name:     "带资源ID的创建路径",
			path:     "/api/v1/authorization-codes",
			wantType: "authorization-codes",
			wantID:   "",
		},
		{
			name:     "UUID 段识别为资源ID",
			path:     "/api/v1/authorization-codes/611d0891-c057-4470-bec6-f82dbaa1df00",
			wantType: "authorization-codes",
			wantID:   "611d0891-c057-4470-bec6-f82dbaa1df00",
		},
		{
			name:     "子动作路径取第一段UUID",
			path:     "/api/v1/authorization-codes/611d0891-c057-4470-bec6-f82dbaa1df00/lock",
			wantType: "authorization-codes",
			wantID:   "611d0891-c057-4470-bec6-f82dbaa1df00",
		},
		{
			name:     "batch 不是资源ID",
			path:     "/api/v1/authorization-codes/batch",
			wantType: "authorization-codes",
			wantID:   "",
		},
		{
			name:     "产品项路径",
			path:     "/api/v1/product-offerings",
			wantType: "product-offerings",
			wantID:   "",
		},
		{
			name:     "公钥路径取产品ID",
			path:     "/api/v1/product-offerings/9f81c0de-1a2b-3c4d-5e6f-7a8b9c0d1e2f/public-key",
			wantType: "product-offerings",
			wantID:   "9f81c0de-1a2b-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			name:     "非 api 前缀路径",
			path:     "/v1/customers",
			wantType: "customers",
			wantID:   "",
		},
		{
			name:     "空路径",
			path:     "/",
			wantType: "",
			wantID:   "",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)

			gotType, gotID := resolveAuditResource(c)
			if gotType != tt.wantType {
				t.Errorf("resourceType = %q, want %q", gotType, tt.wantType)
			}
			if gotID != tt.wantID {
				t.Errorf("resourceID = %q, want %q", gotID, tt.wantID)
			}
		})
	}
}

func TestIsWriteMethod(t *testing.T) {
	writes := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	reads := []string{http.MethodGet, http.MethodHead, http.MethodOptions}

	for _, m := range writes {
		if !isWriteMethod(m) {
			t.Errorf("%s 应视为写操作", m)
		}
	}
	for _, m := range reads {
		if isWriteMethod(m) {
			t.Errorf("%s 不应视为写操作", m)
		}
	}
}

func TestIsFailureResponse(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"成功码", `{"code":"000000","message":"成功"}`, false},
		{"业务错误码", `{"code":"500001","message":"产品项不存在"}`, true},
		{"空响应体", ``, false},
		{"非 JSON", `<html>error</html>`, false},
		{"无 code 字段", `{"data":{"id":"x"}}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFailureResponse(tt.body); got != tt.want {
				t.Errorf("isFailureResponse(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestIsUUIDLike(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"611d0891-c057-4470-bec6-f82dbaa1df00", true},
		{"9F81C0DE-1A2B-3C4D-5E6F-7A8B9C0D1E2F", true},
		{"batch", false},
		{"product-offerings", false},
		{"611d0891c0574470bec6f82dbaa1df00", false},
		{"611d0891-c057-4470-bec6-f82dbaa1df0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := isUUIDLike(tt.value); got != tt.want {
				t.Errorf("isUUIDLike(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestExtractRespID(t *testing.T) {
	body := `{"code":"000000","data":{"id":"abc-123","name":"x"}}`
	if got := extractRespID(body); got != "abc-123" {
		t.Errorf("extractRespID = %q, want %q", got, "abc-123")
	}
	if got := extractRespID(`{"code":"000000","data":null}`); got != "" {
		t.Errorf("data 为 null 时应返回空串, got %q", got)
	}
	if got := extractRespID("not json"); got != "" {
		t.Errorf("非 JSON 应返回空串, got %q", got)
	}
}

func TestTruncateAuditBody(t *testing.T) {
	short := "short body"
	if got := truncateAuditBody(short); got != short {
		t.Errorf("短内容不应被截断, got %q", got)
	}

	long := make([]byte, auditBodyLimit+500)
	for i := range long {
		long[i] = 'a'
	}
	got := truncateAuditBody(string(long))
	if len(got) != auditBodyLimit+len("...[truncated]") {
		t.Errorf("截断后长度 = %d, want %d", len(got), auditBodyLimit+len("...[truncated]"))
	}
}
