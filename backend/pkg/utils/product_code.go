package utils

import (
	"crypto/rand"
	"fmt"
)

// productCodePrefix 产品项编码前缀：Product Element 缩写，
// 大写连字符风格与授权码 LIC- 一致；刻意避开含 O 的缩写，
// 防止与数字 0 混淆
const productCodePrefix = "PE-"

// productCodeLength 随机部分长度
const productCodeLength = 12

// productCodeChars 随机字符集：小写字母 + 数字，
// 已剔除易混淆字符 0/o/1/l/i，人工转述与工单录入不再有二义性
const productCodeChars = "23456789abcdefghjkmnpqrstuvwxyz"

// GenerateProductCode 生成产品项编码：PE- + 12位无混淆随机字符。
//
// 编码由系统生成而非用户填写，创建后不可更改（D4 决策）。使用随机而非递增序列，
// 避免从编码推测业务量级；PE- 前缀保证与授权码（LIC-）在日志和工单中可区分。
func GenerateProductCode() (string, error) {
	randomStr, err := generateRandomFromCharset(productCodeLength, productCodeChars)
	if err != nil {
		return "", err
	}
	return productCodePrefix + randomStr, nil
}

// generateRandomFromCharset 使用加密安全随机源生成指定长度的字符串
func generateRandomFromCharset(length int, chars string) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes), nil
}

// ValidateProductCodeFormat 校验产品项编码格式：PO- + 12位无混淆小写字符。
//
// 仅用于服务端防御性校验，正常编码一律由 GenerateProductCode 生成。
func ValidateProductCodeFormat(code string) error {
	if len(code) != len(productCodePrefix)+productCodeLength {
		return fmt.Errorf("产品项编码长度必须为 %d 个字符", len(productCodePrefix)+productCodeLength)
	}
	if code[:len(productCodePrefix)] != productCodePrefix {
		return fmt.Errorf("产品项编码必须以 %s 开头", productCodePrefix)
	}
	for _, c := range code[len(productCodePrefix):] {
		if !containsRune(productCodeChars, c) {
			return fmt.Errorf("产品项编码只能包含无混淆的小写字母和数字（不含 0/o/1/l/i）")
		}
	}
	return nil
}

func containsRune(charset string, r rune) bool {
	for _, c := range charset {
		if c == r {
			return true
		}
	}
	return false
}
