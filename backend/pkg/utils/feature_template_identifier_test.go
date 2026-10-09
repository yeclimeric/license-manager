package utils

import (
	"strings"
	"testing"
)

func TestGenerateFeatureTemplateIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"小写英文", "pro", "pro", false},
		{"大写英文", "Pro", "pro", false},
		{"带空格", "Pro Version", "pro_version", false},
		{"带连字符", "pro-version", "pro_version", false},
		{"以数字开头", "2fast", "ft_2fast", false},
		{"纯数字", "123", "ft_123", false},
		{"纯中文无结果", "功能模板", "", true},
		{"英文中文混合", "Pro模板", "pro", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFeatureTemplateIdentifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GenerateFeatureTemplateIdentifier(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("GenerateFeatureTemplateIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if err := ValidateFeatureTemplateIdentifierFormat(got); err != nil {
				t.Errorf("生成的标识符未通过自身校验: %q -> %v", got, err)
			}
		})
	}
}

func TestGenerateFeatureTemplateIdentifierTruncation(t *testing.T) {
	// 超长名称应被截断并追加随机后缀，且总长不超过上限
	long := strings.Repeat("verylongproductname", 20)
	got, err := GenerateFeatureTemplateIdentifier(long)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > featureTemplateIdentifierMaxLength {
		t.Errorf("标识符长度 %d 超过上限 %d", len(got), featureTemplateIdentifierMaxLength)
	}
	if err := ValidateFeatureTemplateIdentifierFormat(got); err != nil {
		t.Errorf("截断后的标识符未通过校验: %q -> %v", got, err)
	}
	if !strings.Contains(got, "_") {
		t.Errorf("截断后的标识符应包含随机后缀分隔符: %q", got)
	}
}

func TestValidateFeatureTemplateIdentifierFormat(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantErr    bool
	}{
		{"合法小写", "pro_version", false},
		{"合法含数字", "version2", false},
		{"空字符串", "", true},
		{"含大写", "ProVersion", true},
		{"含连字符", "pro-version", true},
		{"含空格", "pro version", true},
		{"数字开头", "2pro", true},
		{"超长", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFeatureTemplateIdentifierFormat(tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFeatureTemplateIdentifierFormat(%q) error = %v, wantErr = %v", tt.identifier, err, tt.wantErr)
			}
		})
	}
}
