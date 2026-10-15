package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadEnvOverride 验证敏感配置可由环境变量覆盖（单点真相收敛到 .env）：
// DATABASE_PASSWORD 覆盖 database.password，AUTH_JWT_SECRET 覆盖 auth.jwt.secret。
func TestLoadEnvOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
server:
  host: "0.0.0.0"
  port: 18888
database:
  host: "mysql"
  port: 3306
  username: "license_user"
  password: "from-yaml"
  database: "license_manager"
auth:
  jwt:
    secret: "from-yaml"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	t.Setenv("DATABASE_PASSWORD", "env-password")
	t.Setenv("AUTH_JWT_SECRET", "env-secret")

	if err := Load(cfgPath); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := GetConfig().Database.Password; got != "env-password" {
		t.Errorf("database.password = %q, want %q", got, "env-password")
	}
	if got := GetConfig().Auth.JWT.Secret; got != "env-secret" {
		t.Errorf("auth.jwt.secret = %q, want %q", got, "env-secret")
	}
}

// TestLoadEnvFallback 验证环境变量缺失时回退到配置文件值（不破坏现有行为）。
func TestLoadEnvFallback(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
database:
  password: "yaml-password"
auth:
  jwt:
    secret: "yaml-secret"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	// 确保环境变量未设置（viper 会把空字符串 env 也视为覆盖，故用 Unsetenv 模拟缺失）
	os.Unsetenv("DATABASE_PASSWORD")
	os.Unsetenv("AUTH_JWT_SECRET")

	if err := Load(cfgPath); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := GetConfig().Database.Password; got != "yaml-password" {
		t.Errorf("database.password = %q, want %q", got, "yaml-password")
	}
	if got := GetConfig().Auth.JWT.Secret; got != "yaml-secret" {
		t.Errorf("auth.jwt.secret = %q, want %q", got, "yaml-secret")
	}
}
