package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTempConfig 写入临时配置文件并返回路径。
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return p
}

// TestJWTSecretReplacementByMode 验证 JWT 密钥兜底替换按 server.mode 区分：
//   - 空值 / 已知占位值（please-change-in-prod、viper 默认值）：任何模式都替换为随机强密钥。
//   - 开发环境固定密钥（license-manager-dev-secret-key-32-bytes）：随仓库公开，
//     在 release 模式下必须替换；在 debug 模式下保留，避免本地调试时令牌频繁失效。
func TestJWTSecretReplacementByMode(t *testing.T) {
	cases := []struct {
		name       string
		mode       string
		secret     string
		wantRandom bool
	}{
		{"空值_任意模式都替换", "debug", "", true},
		{"占位符_任意模式都替换", "release", "please-change-in-prod", true},
		{"viper默认值_任意模式都替换", "debug", "license-manager-default-secret-key", true},
		{"开发固定密钥_release下替换", "release", devJWTSecret, true},
		{"开发固定密钥_debug下保留", "debug", devJWTSecret, false},
		{"自定义强密钥_始终保留", "release", "my-own-strong-secret-value", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfgPath := writeTempConfig(t,
				"server:\n  mode: \""+c.mode+"\"\nauth:\n  jwt:\n    secret: \""+c.secret+"\"\n")

			os.Unsetenv("AUTH_JWT_SECRET")
			if err := Load(cfgPath); err != nil {
				t.Fatalf("Load: %v", err)
			}

			got := GetConfig().Auth.JWT.Secret
			isRandom := got != c.secret && len(got) == 64 // 32 字节 hex
			if isRandom != c.wantRandom {
				t.Errorf("mode=%s secret=%q -> 是否已随机化 = %v, 期望 %v (实际值长度 %d)",
					c.mode, c.secret, isRandom, c.wantRandom, len(got))
			}
		})
	}
}

// TestJWTSecretEnvOverridesMode 验证环境变量始终优先，且不受 mode 判定影响。
func TestJWTSecretEnvOverridesMode(t *testing.T) {
	cfgPath := writeTempConfig(t,
		"server:\n  mode: \"release\"\nauth:\n  jwt:\n    secret: \""+devJWTSecret+"\"\n")

	t.Setenv("AUTH_JWT_SECRET", "env-provided-strong-secret")

	if err := Load(cfgPath); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := GetConfig().Auth.JWT.Secret; got != "env-provided-strong-secret" {
		t.Errorf("auth.jwt.secret = %q, 期望环境变量值", got)
	}
}
