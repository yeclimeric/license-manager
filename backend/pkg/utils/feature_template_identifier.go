package utils

import (
	"fmt"
	"strings"
	"unicode"

	"license-manager/pkg/utils/slug"
)

// featureTemplateIdentifierMaxLength 标识符最大长度，与 DB varchar(100) 对齐
const featureTemplateIdentifierMaxLength = 100

// GenerateFeatureTemplateIdentifier 从模板名称生成标识符。
//
// 模板标识符会被客户端 SDK 读取并写入许可证，因此必须稳定且符合编程语言
// 变量命名约定：仅保留小写字母、数字与下划线，首字符必须为字母。
func GenerateFeatureTemplateIdentifier(name string) (string, error) {
	converted := slug.Convert(name)
	if converted == "" {
		return "", fmt.Errorf("模板名称无法转换为合法标识符: %q", name)
	}

	// 首字符必须是字母，避免纯数字开头
	if unicode.IsDigit(rune(converted[0])) {
		converted = "ft_" + converted
	}

	if len(converted) > featureTemplateIdentifierMaxLength {
		// 保留前缀并追加随机后缀，避免长名称截断后互相冲突
		suffix, err := generateRandomFromCharset(8, productCodeChars)
		if err != nil {
			return "", fmt.Errorf("生成标识符后缀失败: %w", err)
		}
		keep := featureTemplateIdentifierMaxLength - len(suffix) - 1
		if keep < 1 {
			keep = 1
		}
		converted = strings.TrimRight(converted[:keep], "_") + "_" + suffix
	}

	return converted, nil
}

// ValidateFeatureTemplateIdentifierFormat 校验标识符格式：
// 仅小写字母、数字、下划线，且以字母开头。
func ValidateFeatureTemplateIdentifierFormat(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("标识符不能为空")
	}
	if len(identifier) > featureTemplateIdentifierMaxLength {
		return fmt.Errorf("标识符长度不能超过 %d 个字符", featureTemplateIdentifierMaxLength)
	}
	if unicode.IsDigit(rune(identifier[0])) {
		return fmt.Errorf("标识符必须以字母开头")
	}
	for _, c := range identifier {
		if c != '_' && !unicode.IsDigit(c) && !(c >= 'a' && c <= 'z') {
			return fmt.Errorf("标识符只能包含小写字母、数字和下划线")
		}
	}
	return nil
}
