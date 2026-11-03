package utils

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// GenerateLegacyAuthorizationCode 生成旧规则授权码：
// LIC-{客户ID前4位}-{12位随机}-{4位校验码}
func GenerateLegacyAuthorizationCode(customerID string) (string, error) {
	customerCode := "COMP"
	if len(customerID) >= 4 {
		customerCode = strings.ToUpper(customerID[:4])
	}

	randomStr, err := generateSecureRandomString(12)
	if err != nil {
		return "", err
	}

	checksum := generateChecksum62(customerCode + randomStr)
	return fmt.Sprintf("LIC-%s-%s-%s", customerCode, randomStr, checksum), nil
}

// authCodeRandomChars 授权码随机段字符集：62 字符集剔除易混淆的 0/O/1/l/I，
// 电话/工单口头转述不再有二义性；57 字符 × 12 位仍有约 2×10^21 种组合
const authCodeRandomChars = "ABCDEFGHJKMNPQRSTUVWXYZ23456789abcdefghjkmnpqrstuvwxyz"

func generateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = authCodeRandomChars[b%byte(len(authCodeRandomChars))]
	}
	return string(bytes), nil
}

func generateChecksum62(input string) string {
	sum := 0
	for _, char := range input {
		sum += int(char)
	}
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	checksum := ""
	for i := 0; i < 4; i++ {
		checksum += string(chars[sum%len(chars)])
		sum = sum / len(chars)
	}
	return checksum
}

