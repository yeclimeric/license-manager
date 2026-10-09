package utils

import (
	"strings"
	"testing"
)

func TestGenerateProductCode(t *testing.T) {
	code, err := GenerateProductCode()
	if err != nil {
		t.Fatalf("GenerateProductCode() error = %v", err)
	}

	if err := ValidateProductCodeFormat(code); err != nil {
		t.Errorf("生成的编码未通过自身格式校验: code=%q err=%v", code, err)
	}

	if !strings.HasPrefix(code, productCodePrefix) {
		t.Errorf("编码缺少前缀 %s: %q", productCodePrefix, code)
	}

	if got := len(code); got != len(productCodePrefix)+productCodeLength {
		t.Errorf("编码长度 = %d, 期望 %d", got, len(productCodePrefix)+productCodeLength)
	}
}

func TestGenerateProductCodeUniqueness(t *testing.T) {
	// 生成 1000 个，验证随机性未导致碰撞
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		code, err := GenerateProductCode()
		if err != nil {
			t.Fatalf("GenerateProductCode() error = %v", err)
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("编码重复: %q", code)
		}
		seen[code] = struct{}{}
	}
}

func TestValidateProductCodeFormat(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"合法编码", "po_a1b2c3d4e5f6", false},
		{"合法编码全数字", "po_0123456789ab", false},
		{"长度不足", "po_a1b2", true},
		{"长度超长", "po_a1b2c3d4e5f6g7", true},
		{"前缀错误", "xx_a1b2c3d4e5f6", true},
		{"缺少前缀", "a1b2c3d4e5f6", true},
		{"含大写字母", "po_A1b2c3d4e5f6", true},
		{"含连字符", "po_a1b2-c3d4e5", true},
		{"含下划线", "po_a1b2_c3d4e5", true},
		{"空字符串", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProductCodeFormat(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProductCodeFormat(%q) error = %v, wantErr = %v", tt.code, err, tt.wantErr)
			}
		})
	}
}
