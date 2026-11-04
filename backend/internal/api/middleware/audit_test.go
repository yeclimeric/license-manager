package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResolveAuditResource(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		params   gin.Params
		wantType string
		wantID   string
	}{
		{
			name:     "授权码列表路径规范化",
			path:     "/api/v1/authorization-codes",
			wantType: "authorization-code",
			wantID:   "",
		},
		{
			name:     "UUID 段识别为资源ID",
			path:     "/api/v1/authorization-codes/611d0891-c057-4470-bec6-f82dbaa1df00",
			wantType: "authorization-code",
			wantID:   "611d0891-c057-4470-bec6-f82dbaa1df00",
		},
		{
			name:     "子动作路径取第一段UUID",
			path:     "/api/v1/authorization-codes/611d0891-c057-4470-bec6-f82dbaa1df00/lock",
			wantType: "authorization-code",
			wantID:   "611d0891-c057-4470-bec6-f82dbaa1df00",
		},
		{
			name:     "batch 不是资源ID",
			path:     "/api/v1/authorization-codes/batch",
			wantType: "authorization-code",
			wantID:   "",
		},
		{
			name:     "产品项路径规范化",
			path:     "/api/v1/product-offerings",
			wantType: "product-offering",
			wantID:   "",
		},
		{
			name:     "公钥路径取产品ID",
			path:     "/api/v1/product-offerings/9f81c0de-1a2b-3c4d-5e6f-7a8b9c0d1e2f/public-key",
			wantType: "product-offering",
			wantID:   "9f81c0de-1a2b-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			name:     "客户路径规范化",
			path:     "/v1/customers",
			wantType: "customer",
			wantID:   "",
		},
		{
			name:     "用户路径规范化",
			path:     "/api/v1/users",
			wantType: "user",
			wantID:   "",
		},
		{
			name:     "API Key 路径规范化",
			path:     "/api/v1/api-keys",
			wantType: "api-key",
			wantID:   "",
		},
		{
			name:     "非 UUID 资源ID由路由参数提取",
			path:     "/api/v1/api-keys/ak_ab12cd34ef",
			params:   gin.Params{{Key: "id", Value: "ak_ab12cd34ef"}},
			wantType: "api-key",
			wantID:   "ak_ab12cd34ef",
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
			if tt.params != nil {
				c.Params = tt.params
			}

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

func TestRedactSensitive(t *testing.T) {
	// 敏感字段被替换为 ***，非敏感字段保留
	checkObj := func(t *testing.T, body string, sensitive, preserved []string) {
		t.Helper()
		got := redactSensitive(body)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(got), &data); err != nil {
			t.Fatalf("redactSensitive 结果非 JSON: %v (got %q)", err, got)
		}
		for _, k := range sensitive {
			if v, ok := data[k]; !ok || v != "***" {
				t.Errorf("敏感字段 %q 应被脱敏为 ***, got %v", k, v)
			}
		}
		for _, k := range preserved {
			if _, ok := data[k]; !ok {
				t.Errorf("非敏感字段 %q 应保留, got %v", k, data)
			}
		}
	}

	t.Run("密码字段脱敏", func(t *testing.T) {
		checkObj(t, `{"username":"a","password":"secret123"}`, []string{"password"}, []string{"username"})
	})
	t.Run("token 字段脱敏", func(t *testing.T) {
		checkObj(t, `{"token":"abc.def.ghi"}`, []string{"token"}, nil)
	})
	t.Run("嵌套对象脱敏", func(t *testing.T) {
		got := redactSensitive(`{"user":{"api_key":"k1"},"name":"x"}`)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(got), &data); err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		user, ok := data["user"].(map[string]interface{})
		if !ok || user["api_key"] != "***" {
			t.Errorf("嵌套 api_key 应脱敏, got %v", data)
		}
		if data["name"] != "x" {
			t.Errorf("name 应保留, got %v", data["name"])
		}
	})
	t.Run("非 JSON 原样返回", func(t *testing.T) {
		if got := redactSensitive("not json"); got != "not json" {
			t.Errorf("非 JSON 应原样返回, got %q", got)
		}
	})
	t.Run("数组元素脱敏", func(t *testing.T) {
		got := redactSensitive(`[{"password":"p"}]`)
		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(got), &arr); err != nil || len(arr) != 1 || arr[0]["password"] != "***" {
			t.Errorf("数组内 password 应脱敏, got %q", got)
		}
	})
}
