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
		{"合法编码", "PO-a2b3c4d5e6f7", false},
		{"合法编码全数字", "PO-234567892345", false},
		{"长度不足", "PO-a2b3", true},
		{"长度超长", "PO-a2b3c4d5e6f7g8", true},
		{"旧前缀po_已废弃", "po_a2b3c4d5e6f7", true},
		{"前缀错误", "XX-a2b3c4d5e6f7", true},
		{"缺少前缀", "a2b3c4d5e6f7", true},
		{"含大写字母", "PO-A2b3c4d5e6f7", true},
		{"含易混字符0", "PO-a2b3c4d5e6f0", true},
		{"含易混字符o", "PO-a2b3c4d5e6fo", true},
		{"含易混字符l", "PO-a2b3c4d5e6fl", true},
		{"含易混字符i", "PO-a2b3c4d5e6fi", true},
		{"含连字符", "PO-a2b3-c4d5e6f", true},
		{"含下划线", "PO-a2b3_c4d5e6f", true},
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
