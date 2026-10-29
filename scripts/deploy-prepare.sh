#!/usr/bin/env bash
# 部署准备脚本：把「部署者需手动改的多处配置」收敛为「只改 .env + 跑本脚本」。
# 作用：
#   1. 若 .env 不存在，从 .env.example 复制（部署者唯一需要编辑的文件）。
#   2. 若 JWT_SECRET 未设置，生成强随机值写入 .env。
#   3. 若 RSA 密钥缺失，生成专属密钥对到 backend/configs（覆盖示例密钥需 --force）。
set -euo pipefail

cd "$(dirname "$0")/.." || exit 1
echo "==> 仓库根目录: $(pwd)"

# 1. .env
if [ -f .env ]; then
  echo "[skip] .env 已存在，保留现有配置（如需重置请删除 .env 后重跑）"
else
  cp .env.example .env
  echo "[ok] 已从 .env.example 创建 .env"
fi

# 2. JWT_SECRET
if grep -Eq '^JWT_SECRET=.+' .env 2>/dev/null; then
  echo "[skip] JWT_SECRET 已设置"
else
  SECRET=$(openssl rand -hex 32)
  if grep -Eq '^JWT_SECRET=' .env; then
    sed -i.bak 's/^JWT_SECRET=.*/JWT_SECRET='"$SECRET"'/' .env && rm -f .env.bak
  else
    printf 'JWT_SECRET=%s\n' "$SECRET" >> .env
  fi
  echo "[ok] 已生成随机 JWT_SECRET 并写入 .env"
fi

# 3. RSA 密钥
if [ -f backend/configs/rsa_private_key.pem ] && [ -f backend/configs/rsa_public_key.pem ]; then
  echo "[skip] RSA 密钥已存在（重新生成请运行: bash scripts/gen-rsa-keys.sh backend/configs 2048 --force）"
else
  bash scripts/gen-rsa-keys.sh backend/configs 2048 --force
  echo "[ok] 已生成 RSA 密钥对到 backend/configs"
fi

echo ""
echo "==> 部署准备完成。"
echo "    1) 按需编辑 .env（至少修改 MYSQL_ROOT_PASSWORD / MYSQL_PASSWORD；JWT_SECRET 已自动生成）"
echo "    2) 启动："
echo "       - 通用/1panel 服务器: docker compose -f docker-compose.prod.yml up -d --build"
echo "       - 1panel 面板:        用 docker-compose.1panel.yml 创建编排并启动"
echo "    注：backend 通过环境变量 DATABASE_PASSWORD / AUTH_JWT_SECRET 读取敏感配置，"
echo "        无需再手动修改 backend/configs/config.prod.yaml。"
