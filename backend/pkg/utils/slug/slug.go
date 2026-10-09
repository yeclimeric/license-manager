package slug

import (
	"strings"
	"unicode"
)

// Convert 将任意文本转为 slug：保留小写字母、数字与下划线，
// 其余字符转为下划线，压缩连续下划线并去除首尾下划线。
//
// 用于从人类可读名称生成代码标识符，中文等非 ASCII 字符会被丢弃，
// 因此调用方需处理"转换结果为空"的情况。
func Convert(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	lastUnderscore := true // 行首视为已分隔，避免产生首下划线
	for _, r := range text {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastUnderscore = false
		case unicode.IsDigit(r):
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}

	return strings.Trim(b.String(), "_")
}
