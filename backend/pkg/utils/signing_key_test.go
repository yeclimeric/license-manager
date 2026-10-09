package utils

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
)

func TestGenerateProductSigningKey(t *testing.T) {
	key, err := GenerateProductSigningKey()
	if err != nil {
		t.Fatalf("GenerateProductSigningKey() error = %v", err)
	}

	if !strings.HasPrefix(key.PublicKeyPEM, "-----BEGIN PUBLIC KEY-----") {
		t.Errorf("公钥 PEM 头不正确: %q", strings.SplitN(key.PublicKeyPEM, "\n", 2)[0])
	}
	if !strings.HasSuffix(strings.TrimSpace(key.PublicKeyPEM), "-----END PUBLIC KEY-----") {
		t.Error("公钥 PEM 尾不正确")
	}
	if !strings.Contains(key.PrivateKeyPEM, "PRIVATE KEY") {
		t.Error("私钥 PEM 类型不正确")
	}

	// 公钥必须可解析
	if _, err := LoadRSAPublicKeyFromString(key.PublicKeyPEM); err != nil {
		t.Errorf("生成的公钥无法被解析: %v", err)
	}

	// 密钥长度应为 2048 位
	block, _ := pem.Decode([]byte(key.PublicKeyPEM))
	if block == nil {
		t.Fatal("公钥 PEM 解析失败")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("解析公钥失败: %v", err)
	}
	type sizeAder interface{ Size() int }
	if s, ok := pub.(sizeAder); ok {
		if got := s.Size() * 8; got != signingKeyBits {
			t.Errorf("密钥长度 = %d 位, 期望 %d 位", got, signingKeyBits)
		}
	}
}

func TestSignAndVerifyRoundTrip(t *testing.T) {
	key, err := GenerateProductSigningKey()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	payload := []byte(`{"product_code":"po_test","license_code":"LIC-ABC"}`)
	signature, err := SignWithPrivateKeyPEM(key.PrivateKeyPEM, payload)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	if err := VerifyWithPublicKeyPEM(key.PublicKeyPEM, payload, signature); err != nil {
		t.Errorf("验签失败: %v", err)
	}
}

func TestVerifyRejectsTamperedPayload(t *testing.T) {
	key, err := GenerateProductSigningKey()
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	original := []byte(`{"days":365}`)
	signature, err := SignWithPrivateKeyPEM(key.PrivateKeyPEM, original)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// 内容被篡改后必须验签失败
	tampered := []byte(`{"days":9999}`)
	if err := VerifyWithPublicKeyPEM(key.PublicKeyPEM, tampered, signature); err == nil {
		t.Error("被篡改的内容竟然验签通过")
	}
}

func TestVerifyRejectsOtherProductKey(t *testing.T) {
	keyA, err := GenerateProductSigningKey()
	if err != nil {
		t.Fatalf("生成密钥 A 失败: %v", err)
	}
	keyB, err := GenerateProductSigningKey()
	if err != nil {
		t.Fatalf("生成密钥 B 失败: %v", err)
	}

	payload := []byte("license-data")
	signature, err := SignWithPrivateKeyPEM(keyA.PrivateKeyPEM, payload)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	// 跨产品验签必须失败：A 签的许可证不能用 B 的公钥验证
	if err := VerifyWithPublicKeyPEM(keyB.PublicKeyPEM, payload, signature); err == nil {
		t.Error("跨产品验签竟然通过，说明密钥未按产品隔离")
	}
}

func TestSignWithInvalidPrivateKey(t *testing.T) {
	if _, err := SignWithPrivateKeyPEM("not-a-pem", []byte("data")); err == nil {
		t.Error("无效私钥应当报错")
	}
	if _, err := SignWithPrivateKeyPEM("", []byte("data")); err == nil {
		t.Error("空私钥应当报错")
	}
}
