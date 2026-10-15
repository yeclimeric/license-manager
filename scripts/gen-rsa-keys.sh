#!/usr/bin/env bash
#
# 生成生产用 RSA 密钥对
#
# 仓库自带的 rsa_private_key.pem / rsa_public_key.pem 私钥已随仓库公开，
# 生产部署必须重新生成专属密钥对并替换，否则授权签名可被伪造。
#
# 用法:
#   ./scripts/gen-rsa-keys.sh                      # 默认输出到 backend/configs，2048 位
#   ./scripts/gen-rsa-keys.sh ./mykeys 4096        # 自定义目录与位数
#   ./scripts/gen-rsa-keys.sh backend/configs 2048 --force   # 覆盖已存在文件
#
# 输出格式与 backend/cmd/gen-rsa-keys 一致：
#   私钥 PKCS#1 (BEGIN RSA PRIVATE KEY)，公钥 PKIX (BEGIN PUBLIC KEY)。
#   与后端 x509.ParsePKCS1PrivateKey / ParsePKIXPublicKey 兼容。

set -euo pipefail

OUT_DIR="backend/configs"
SIZE=2048
FORCE=0

for arg in "$@"; do
  case "$arg" in
    --force | -f) FORCE=1 ;;
    *)
      if [ "$OUT_DIR" = "backend/configs" ]; then
        OUT_DIR="$arg"
      else
        SIZE="$arg"
      fi
      ;;
  esac
done

PRIVATE="$OUT_DIR/rsa_private_key.pem"
PUBLIC="$OUT_DIR/rsa_public_key.pem"

if ! command -v openssl >/dev/null 2>&1; then
  echo "错误: 未找到 openssl，请先安装 OpenSSL。" >&2
  exit 1
fi

if [ -f "$PRIVATE" ] && [ "$FORCE" -ne 1 ]; then
  echo "错误: $PRIVATE 已存在（不覆盖示例/现有密钥）。如需覆盖请加 --force。" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

echo "生成 RSA 密钥对 (${SIZE} 位) -> $OUT_DIR"
openssl genrsa -out "$PRIVATE" "$SIZE" 2>/dev/null
openssl rsa -in "$PRIVATE" -pubout -out "$PUBLIC" 2>/dev/null
chmod 600 "$PRIVATE"
chmod 644 "$PUBLIC"

echo "✓ 私钥: $PRIVATE (权限 600)"
echo "✓ 公钥: $PUBLIC"
echo ""
echo "注意:"
echo "  - 私钥仅供服务器端使用，务必保密，勿提交进仓库或对外泄露。"
echo "  - 公钥可分发 / 编译进客户端用于验签。"
echo "  - 替换密钥后，旧密钥签发的授权码将失效，需重新签发。"
