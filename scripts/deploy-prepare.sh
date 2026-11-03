#!/usr/bin/env bash
# 部署准备脚本：把「部署者需手动改的多处配置」收敛为「只改 .env + 跑本脚本」。
# 作用：
#   1. 若 .env 不存在，从 .env.example 复制（部署者唯一需要编辑的文件）。
#   2. 若 JWT_SECRET 未设置，生成强随机值写入 .env。
#   3. 若 RSA 密钥缺失，或仍是仓库自带的示例私钥（已公开，可被伪造），
#      自动重新生成专属密钥对到 backend/configs。
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
# 仓库自带的示例密钥私钥已随仓库公开，生产必须使用专属密钥。
# 因此不能只看「文件是否存在」——示例密钥永远存在，必须额外比对内容指纹：
# 若当前私钥仍与仓库示例一致，则强制重新生成。
EXAMPLE_PRIVATE_FINGERPRINT="ae21892f74116542ba70c9851922b76d45a6012daa9efe4bd6fe5e9e8c704f32"
PRIVATE_KEY="backend/configs/rsa_private_key.pem"
PUBLIC_KEY="backend/configs/rsa_public_key.pem"

private_fingerprint() {
  [ -f "$1" ] || return 1
  shasum -a 256 "$1" 2>/dev/null | awk '{print $1}' || sha256sum "$1" 2>/dev/null | awk '{print $1}'
}

current_fingerprint="$(private_fingerprint "$PRIVATE_KEY" || true)"

if [ -z "$current_fingerprint" ]; then
  echo "[info] RSA 密钥缺失，正在生成专属密钥对到 backend/configs ..."
  bash scripts/gen-rsa-keys.sh backend/configs 2048 --force
  echo "[ok] 已生成 RSA 密钥对到 backend/configs"
elif [ "$current_fingerprint" = "$EXAMPLE_PRIVATE_FINGERPRINT" ]; then
  echo "[warn] 检测到当前 RSA 私钥仍是仓库自带的示例密钥（私钥已随仓库公开，可被伪造）。"
  echo "[warn] 正在自动重新生成专属密钥对 ..."
  bash scripts/gen-rsa-keys.sh backend/configs 2048 --force
  echo "[ok] 已替换为新的专属 RSA 密钥对"
else
  echo "[skip] RSA 密钥已存在且非仓库示例密钥，保持不变"
  echo "       如需重新生成: bash scripts/gen-rsa-keys.sh backend/configs 2048 --force"
fi

# 4. 保护生产密钥，防止被 git 覆盖或误提交
# 仓库中的示例密钥是被 git 追踪的，若不处理：
#   - 服务器执行 git pull 时会用仓库示例覆盖本地生产私钥 → 授权签名被替换
#   - 执行 git add . / git commit 时可能把生产私钥提交进仓库
# 用 skip-worktree 让 git 忽略该文件的本地改动（不影响其他文件的正常 pull）。
if git rev-parse --git-dir >/dev/null 2>&1; then
  for pem in "$PRIVATE_KEY" "$PUBLIC_KEY"; do
    if git ls-files --error-unmatch "$pem" >/dev/null 2>&1; then
      if git update-index --skip-worktree "$pem" 2>/dev/null; then
        echo "[ok] 已为 $pem 设置 git skip-worktree（git pull 不会覆盖，git commit 不会误提交）"
      fi
    fi
  done
  echo "       如需恢复 git 追踪: git update-index --no-skip-worktree $PRIVATE_KEY"
fi

echo ""
echo "==> 部署准备完成。"
echo "    1) 按需编辑 .env（至少修改 MYSQL_ROOT_PASSWORD / MYSQL_PASSWORD；JWT_SECRET 已自动生成）"
echo "    2) 启动："
echo "       - 通用/1panel 服务器: docker compose -f docker-compose.prod.yml up -d --build"
echo "       - 1panel 面板:        用 docker-compose.1panel.yml 创建编排并启动"
echo "    注：backend 通过环境变量 DATABASE_PASSWORD / AUTH_JWT_SECRET 读取敏感配置，"
echo "        无需再手动修改 backend/configs/config.prod.yaml。"
