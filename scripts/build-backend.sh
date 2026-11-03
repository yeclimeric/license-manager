#!/usr/bin/env bash
# 构建后端二进制（Go）
#
# 产物：bin/license-manager
#
# 用法：
#   bash scripts/build-backend.sh                       # 构建当前平台
#   bash scripts/build-backend.sh -o /tmp/lm            # 指定输出文件
#   GOOS=linux GOARCH=amd64 bash scripts/build-backend.sh   # 交叉编译
#
# 特性：
#   - CGO_ENABLED=0 静态编译：可在 alpine（Docker 运行时镜像）上直接运行
#   - ldflags="-s -w"：与 Dockerfile.backend 保持一致，去除符号表减小体积
#   - 依赖下载与编译分离：缓存命中时二次构建约 0.7s
#   - 缓存目录默认落在项目 .cache/，跨容器重建保留
set -euo pipefail

cd "$(dirname "$0")/.." || exit 1
ROOT="$(pwd)"

OUTPUT="$ROOT/bin/license-manager"
GO_LDFLAGS="-s -w"

usage() {
  cat <<EOF
用法: $0 [选项]

选项:
  -o, --output <path>   输出文件路径（默认: bin/license-manager）
  -h, --help            显示本帮助

环境变量:
  GOOS / GOARCH         交叉编译目标（如 linux/amd64、linux/arm64）
  GOPROXY               Go 模块代理（未设置时默认 https://goproxy.cn,direct）
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    -o | --output)
      OUTPUT="$2"
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

# ---- 1. 环境检查 ----
if ! command -v go >/dev/null 2>&1; then
  echo "错误: 未找到 go，请先安装 Go 1.23+" >&2
  exit 1
fi

GO_VERSION=$(go env GOVERSION 2>/dev/null || echo "unknown")
GO_MINOR=$(printf '%s' "$GO_VERSION" | sed -E 's/^go([0-9]+)\.([0-9]+).*/\1 \2/')
if ! printf '%s' "$GO_MINOR" | awk '{ exit !($1 > 1 || ($1 == 1 && $2 >= 23)) }'; then
  echo "错误: Go 版本过低（当前 ${GO_VERSION}），本项目需要 Go 1.23+" >&2
  exit 1
fi

echo "==> Go 版本: ${GO_VERSION}"

# ---- 2. 缓存与代理 ----
# 缓存放在项目 .cache/ 下（该目录是挂载目录时跨容器重建保留）。
CACHE_DIR="$ROOT/.cache"
mkdir -p "$CACHE_DIR/mod" "$CACHE_DIR/build"
export GOMODCACHE="$CACHE_DIR/mod"
export GOCACHE="$CACHE_DIR/build"

if [ -z "${GOPROXY:-}" ]; then
  export GOPROXY="https://goproxy.cn,direct"
fi

TARGET_OS="${GOOS:-$(go env GOOS)}"
TARGET_ARCH="${GOARCH:-$(go env GOARCH)}"
echo "==> 目标平台: ${TARGET_OS}/${TARGET_ARCH}"
echo "==> 依赖缓存: ${GOMODCACHE}"
echo "==> 模块代理: ${GOPROXY}"

# ---- 3. 下载依赖 ----
# 与编译分离：便于区分「网络问题」与「编译问题」，也便于失败后单独重试。
echo "==> 下载依赖 ..."
if ! (cd "$ROOT/backend" && go mod download); then
  echo "错误: 依赖下载失败。若为网络问题，可设置 GOPROXY 环境变量后重试" >&2
  exit 1
fi

# ---- 4. 编译 ----
# CGO_ENABLED=0：静态链接，可在 alpine 镜像中直接运行（与 Docker 部署一致）
echo "==> 编译 ..."
mkdir -p "$(dirname "$OUTPUT")"
if ! (cd "$ROOT/backend" && CGO_ENABLED=0 go build -ldflags="$GO_LDFLAGS" -o "$OUTPUT" cmd/main.go); then
  echo "错误: 编译失败" >&2
  exit 1
fi

echo ""
echo "✓ 构建完成: $OUTPUT"
ls -lh "$OUTPUT" | awk '{print "  体积: " $5}'
echo "  目标平台: ${TARGET_OS}/${TARGET_ARCH} (CGO_ENABLED=0, 静态链接)"
echo ""
echo "  部署提示: 产物为静态二进制，可直接复制到 Linux 服务器运行；"
echo "           若目标为 alpine 镜像，请确认架构匹配（linux/amd64 或 linux/arm64）。"
