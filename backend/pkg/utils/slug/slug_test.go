package slug

import "testing"

func TestConvert(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"纯小写英文", "pro", "pro"},
		{"大写转小写", "Pro", "pro"},
		{"混合大小写下划线", "Pro_Version_2", "pro_version_2"},
		{"空格转下划线", "pro version", "pro_version"},
		{"连字符转下划线", "pro-version", "pro_version"},
		{"压缩连续下划线", "pro__version", "pro_version"},
		{"去除首尾下划线", "__pro__", "pro"},
		{"保留数字", "version2final", "version2final"},
		{"中文全部丢弃", "功能模板", ""},
		{"中文与英文混合", "Pro版模板", "pro"},
		{"仅符号", "---", ""},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.input); got != tt.want {
				t.Errorf("Convert(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
