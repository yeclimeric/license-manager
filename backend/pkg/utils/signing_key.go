package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// signingKeyBits 签名密钥长度。
// 企业版实测公钥为 RSA 2048（451 字节 PEM），此处与之对齐。
const signingKeyBits = 2048

// cryptoSHA256 签名算法，与许可证验签保持一致
const cryptoSHA256 = crypto.SHA256

// sha256Sum 计算 SHA-256 摘要
func sha256Sum(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

// base64Encode base64 标准编码
func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// ProductSigningKey 产品签名密钥对
type ProductSigningKey struct {
	// PrivateKeyPEM PKCS#1 私钥，仅用于服务端签名，不返回给客户端
	PrivateKeyPEM string
	// PublicKeyPEM PKIX 公钥，供客户端验签
	PublicKeyPEM string
}

// GenerateProductSigningKey 生成产品签名密钥对。
//
// 每个产品项独立持有一对密钥：客户端只需内嵌对应产品的公钥即可验签，
// 无法伪造其他产品的许可证。私钥永不返回给前端。
func GenerateProductSigningKey() (*ProductSigningKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, signingKeyBits)
	if err != nil {
		return nil, fmt.Errorf("生成签名密钥失败: %w", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("序列化公钥失败: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})

	return &ProductSigningKey{
		PrivateKeyPEM: string(privatePEM),
		PublicKeyPEM:  string(publicPEM),
	}, nil
}

// SignWithPrivateKeyPEM 使用 PEM 私钥对数据签名，返回 base64 签名
func SignWithPrivateKeyPEM(privateKeyPEM string, data []byte) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("私钥格式无效")
	}

	var privateKey *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("解析私钥失败: %w", err)
		}
		privateKey = parsed
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("解析私钥失败: %w", err)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("私钥不是 RSA 类型")
		}
		privateKey = rsaKey
	default:
		return "", fmt.Errorf("不支持的私钥类型: %s", block.Type)
	}

	digest := sha256Sum(data)
	signature, err := rsa.SignPSS(rand.Reader, privateKey, cryptoSHA256, digest, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       cryptoSHA256,
	})
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	return base64Encode(signature), nil
}

// VerifyWithPublicKeyPEM 使用 PEM 公钥验证 base64 签名
//
// 与既有 VerifySignature 保持一致：采用 PSS 填充，参数为原始数据。
func VerifyWithPublicKeyPEM(publicKeyPEM string, data []byte, signatureBase64 string) error {
	publicKey, err := LoadRSAPublicKeyFromString(publicKeyPEM)
	if err != nil {
		return err
	}

	// 复用既有验签实现，避免两处填充算法不一致
	return publicKey.VerifySignature(data, signatureBase64)
}
