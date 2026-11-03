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
#
# 提示：仓库自带的示例私钥已随仓库公开，生产环境务必用本脚本重新生成并替换。

set -euo pipefail

OUT_DIR="backend/configs"
SIZE=2048
FORCE=0

# 位置参数按出现顺序解析：第 1 个为输出目录，第 2 个为密钥位数。
# 注意：不可用「OUT_DIR 是否仍等于默认值」来判定序号——默认值恰为 backend/configs，
# 显式传入同名目录时会误把第 2 个参数（位数）当作目录，导致生成到 ./2048/ 下。
POSITIONAL=()
for arg in "$@"; do
  case "$arg" in
    --force | -f) FORCE=1 ;;
    *)
      POSITIONAL+=("$arg")
      ;;
  esac
done

if [ "${#POSITIONAL[@]}" -ge 1 ]; then
  OUT_DIR="${POSITIONAL[0]}"
fi
if [ "${#POSITIONAL[@]}" -ge 2 ]; then
  SIZE="${POSITIONAL[1]}"
fi
if [ "${#POSITIONAL[@]}" -gt 2 ]; then
  echo "错误: 参数过多。用法: $0 [输出目录] [密钥位数] [--force]" >&2
  exit 1
fi

# 位数必须是纯数字，避免再次出现「位数被当作目录」的情况
case "$SIZE" in
  '' | *[!0-9]*)
    echo "错误: 密钥位数必须为数字（推荐 2048 或 4096），当前为: $SIZE" >&2
    exit 1
    ;;
esac
if [ "$SIZE" -lt 2048 ]; then
  echo "错误: 密钥位数不得小于 2048，当前为: $SIZE" >&2
  exit 1
fi

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
# OpenSSL 3.0+ 的 genrsa 默认输出 PKCS#8（BEGIN PRIVATE KEY），
# 与后端 Go 工具的标准输出（PKCS#1 / BEGIN RSA PRIVATE KEY）不一致。
# 这里显式要求 -traditional 保证跨版本输出格式统一；
# OpenSSL 1.1.1 不支持该参数，但其默认即为 PKCS#1，故失败时回退。
if ! openssl genrsa -traditional -out "$PRIVATE" "$SIZE" 2>/dev/null; then
  openssl genrsa -out "$PRIVATE" "$SIZE" 2>/dev/null
fi
openssl rsa -in "$PRIVATE" -pubout -out "$PUBLIC" 2>/dev/null
chmod 600 "$PRIVATE"
chmod 644 "$PUBLIC"

# 落盘后自检：确认两个文件确实存在于预期路径且格式正确，
# 避免再次出现「参数被误当作目录、密钥落在意外位置」而无人察觉。
if [ ! -s "$PRIVATE" ] || [ ! -s "$PUBLIC" ]; then
  echo "错误: 密钥文件生成异常 (${PRIVATE} / ${PUBLIC})" >&2
  exit 1
fi
# 私钥接受 PKCS#1（首选，与 Go 工具一致）与 PKCS#8（后端可自动回退解析）
if ! head -1 "$PRIVATE" | grep -qE "BEGIN (RSA )?PRIVATE KEY"; then
  echo "错误: 私钥格式异常，期望 PKCS#1 (BEGIN RSA PRIVATE KEY) 或 PKCS#8 (BEGIN PRIVATE KEY)" >&2
  exit 1
fi
if ! head -1 "$PUBLIC" | grep -q "BEGIN PUBLIC KEY"; then
  echo "错误: 公钥格式异常，期望 PKIX (BEGIN PUBLIC KEY)" >&2
  exit 1
fi
openssl rsa -in "$PRIVATE" -check -noout >/dev/null 2>&1 || {
  echo "错误: 私钥校验失败" >&2
  exit 1
}

echo "✓ 私钥: $PRIVATE (权限 600)"
echo "✓ 公钥: $PUBLIC"
echo "✓ 格式自检通过（$(head -1 "$PRIVATE" | tr -d '-' | sed 's/^ *//;s/ *$//') + PKIX 公钥，可被后端直接解析）"
echo ""
echo "注意:"
echo "  - 私钥仅供服务器端使用，务必保密，勿提交进仓库或对外泄露。"
echo "  - 公钥可分发 / 编译进客户端用于验签。"
echo "  - 替换密钥后，旧密钥签发的授权码将失效，需重新签发。"
