package utils

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// buildLicenseFile 构造与激活接口同构的许可证文件信封
func buildLicenseFile(t *testing.T, key *ProductSigningKey, payload map[string]interface{}) (envelope string, signature string, data string) {
	t.Helper()

	dataBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("序列化许可证数据失败: %v", err)
	}

	sig, err := SignWithPrivateKeyPEM(key.PrivateKeyPEM, dataBytes)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	raw, err := json.Marshal(map[string]interface{}{
		"algorithm": "RSA-PSS-SHA256",
		"data":      string(dataBytes),
		"signature": sig,
	})
	if err != nil {
		t.Fatalf("序列化信封失败: %v", err)
	}

	return base64.StdEncoding.EncodeToString(raw), sig, string(dataBytes)
}

// TestLicenseFileVerifiesWithProductKey 激活返回的公钥必须能验证许可证文件
func TestLicenseFileVerifiesWithProductKey(t *testing.T) {
	key, err := GenerateProductSigningKey()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	envelope, _, _ := buildLicenseFile(t, key, map[string]interface{}{
		"license_key":          "LIC-DEVICE-ABC123",
		"authorization_code":   "LIC-COMP-XYZ",
		"hardware_fingerprint": "fp-1",
		"deployment_type":      "standalone",
		"status":               "active",
	})

	raw, err := base64.StdEncoding.DecodeString(envelope)
	if err != nil {
		t.Fatalf("base64 解码失败: %v", err)
	}

	var parsed struct {
		Algorithm string `json:"algorithm"`
		Data      string `json:"data"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("信封解析失败: %v", err)
	}

	if parsed.Algorithm != "RSA-PSS-SHA256" {
		t.Errorf("algorithm = %q, 期望 RSA-PSS-SHA256", parsed.Algorithm)
	}

	if err := VerifyWithPublicKeyPEM(key.PublicKeyPEM, []byte(parsed.Data), parsed.Signature); err != nil {
		t.Errorf("公钥验签失败: %v", err)
	}
}

// TestLicenseFileRejectsOtherProductKey 其他产品的公钥不能验证本产品许可证
func TestLicenseFileRejectsOtherProductKey(t *testing.T) {
	keyA, _ := GenerateProductSigningKey()
	keyB, _ := GenerateProductSigningKey()

	_, _, data := buildLicenseFile(t, keyA, map[string]interface{}{
		"license_key": "LIC-DEVICE-ABC",
		"status":      "active",
	})

	sig, err := SignWithPrivateKeyPEM(keyA.PrivateKeyPEM, []byte(data))
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	if err := VerifyWithPublicKeyPEM(keyB.PublicKeyPEM, []byte(data), sig); err == nil {
		t.Error("跨产品验签竟然通过，说明许可证未按产品隔离")
	}
}

// TestLicenseFileDetectsTampering 篡改许可证内容必须被发现
func TestLicenseFileDetectsTampering(t *testing.T) {
	key, _ := GenerateProductSigningKey()

	_, _, original := buildLicenseFile(t, key, map[string]interface{}{
		"license_key": "LIC-DEVICE-ABC",
		"end_date":    "2026-10-09T23:59:59+08:00",
	})

	sig, err := SignWithPrivateKeyPEM(key.PrivateKeyPEM, []byte(original))
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// 把到期日往后改，试图延长使用期限
	tampered, err := json.Marshal(map[string]interface{}{
		"license_key": "LIC-DEVICE-ABC",
		"end_date":    "2099-12-31T23:59:59+08:00",
	})
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	if err := VerifyWithPublicKeyPEM(key.PublicKeyPEM, tampered, sig); err == nil {
		t.Error("延长到期日后的许可证竟然验签通过")
	}
}

// TestLicenseFileCarriesDeploymentType 部署类型必须随许可证下发
func TestLicenseFileCarriesDeploymentType(t *testing.T) {
	key, _ := GenerateProductSigningKey()

	envelope, _, _ := buildLicenseFile(t, key, map[string]interface{}{
		"license_key":     "LIC-DEVICE-ABC",
		"deployment_type": "cloud",
		"status":          "active",
	})

	raw, _ := base64.StdEncoding.DecodeString(envelope)
	var parsed struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("信封解析失败: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(parsed.Data), &payload); err != nil {
		t.Fatalf("payload 解析失败: %v", err)
	}

	if payload["deployment_type"] != "cloud" {
		t.Errorf("deployment_type = %v, 期望 cloud", payload["deployment_type"])
	}
}
