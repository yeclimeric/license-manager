#!/usr/bin/env bash
# 构建前端静态资源（Vue 3 + Vite）
#
# 产物：frontend/dist/
#
# 用法：
#   bash scripts/build-frontend.sh                          # 构建并输出到 frontend/dist
#   bash scripts/build-frontend.sh -o /var/www/license-ui   # 构建并同步到指定目录
#   bash scripts/build-frontend.sh --mode test              # 使用 test 模式
#   VITE_OUTDIR=production VITE_ADDRESS_BASE_URL=/production/ \
#     bash scripts/build-frontend.sh                        # 构建到子路径（CI 同款）
#
# 重要：
#   仓库 frontend/.env.production 已配置为
#       VITE_OUTDIR='dist'
#       VITE_ADDRESS_BASE_URL='/'
#   即默认就是「输出到 dist/ + 根路径托管」，与 Dockerfile.frontend 一致。
#
#   Vite 的 env 优先级（实测 Vite 5.4.21）：
#       进程环境变量 (export / Dockerfile ENV)  >  .env.production.local  >  .env.production
#   因此本脚本仍显式 export 这两项：
#     - 防止 .env.production 被改动后构建行为漂移
#     - CI 部署到 /production/ 子路径时用同样的方式覆盖，无需改仓库配置
#
#   若需构建到子路径（如 /production/）：
#     VITE_OUTDIR=production VITE_ADDRESS_BASE_URL=/production/ bash scripts/build-frontend.sh
set -euo pipefail

cd "$(dirname "$0")/.." || exit 1
ROOT="$(pwd)"

MODE="production"
OUTDIR=""
SKIP_INSTALL=0
# 默认与 frontend/.env.production 保持一致（dist + 根路径）；
# 可通过环境变量覆盖，例如构建到 /production/ 子路径。
BUILD_OUTDIR="${VITE_OUTDIR:-dist}"
BUILD_BASE_URL="${VITE_ADDRESS_BASE_URL:-/}"

usage() {
  cat <<EOF
用法: $0 [选项]

选项:
  -o, --output <dir>    构建后同步产物到指定目录（如 nginx 站点目录）
  -m, --mode <mode>     构建模式（默认 production，可选 test）
      --skip-install    跳过 pnpm install（依赖已就绪时更快）
  -h, --help            显示本帮助

环境变量:
  VITE_OUTDIR           输出目录（默认 dist，与 .env.production 一致）
  VITE_ADDRESS_BASE_URL 资源 base（默认 /，根路径托管）
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    -o | --output)
      OUTDIR="$2"
      shift 2
      ;;
    -m | --mode)
      MODE="$2"
      shift 2
      ;;
    --skip-install)
      SKIP_INSTALL=1
      shift
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
if ! command -v node >/dev/null 2>&1; then
  echo "错误: 未找到 node，请先安装 Node 18+（建议 20 LTS）" >&2
  exit 1
fi

# 必须用 pnpm：项目的 pnpm-workspace.yaml 白名单了 esbuild 等包的安装脚本，
# 换 npm 会导致 vite 构建失败。
if ! command -v pnpm >/dev/null 2>&1; then
  echo "错误: 未找到 pnpm。本项目必须使用 pnpm（不是 npm/yarn）" >&2
  echo "      安装: npm i -g pnpm@11" >&2
  exit 1
fi

NODE_VERSION=$(node --version)
echo "==> Node: ${NODE_VERSION} / pnpm: $(pnpm --version)"

# pnpm-workspace.yaml 是必需文件：白名单 esbuild / @parcel/watcher / vue-demi 的安装脚本，
# 缺失会导致 pnpm install 报 ERR_PNPM_IGNORED_BUILDS。
if [ ! -f "$ROOT/frontend/pnpm-workspace.yaml" ]; then
  echo "错误: 缺少 frontend/pnpm-workspace.yaml，该文件不可删除" >&2
  exit 1
fi

# ---- 2. 安装依赖 ----
if [ "$SKIP_INSTALL" -eq 0 ]; then
  echo "==> 安装依赖 (pnpm install) ..."
  if ! (cd "$ROOT/frontend" && pnpm install); then
    echo "错误: 依赖安装失败" >&2
    exit 1
  fi
else
  echo "==> 跳过依赖安装 (--skip-install)"
fi

# ---- 3. 构建 ----
# 显式导出构建参数：Vite 中进程环境变量优先级最高，
# 可覆盖 .env.production，避免仓库配置变动影响构建结果。
export VITE_OUTDIR="$BUILD_OUTDIR"
export VITE_ADDRESS_BASE_URL="$BUILD_BASE_URL"

echo "==> 构建 (mode=${MODE}, outdir=${BUILD_OUTDIR}, base=${BUILD_BASE_URL}) ..."
if [ "$MODE" = "production" ]; then
  BUILD_CMD="build:prod"
else
  BUILD_CMD="build:${MODE}"
fi

if ! (cd "$ROOT/frontend" && pnpm run "$BUILD_CMD"); then
  echo "错误: 前端构建失败" >&2
  exit 1
fi

# ---- 4. 校验产物 ----
DIST="$ROOT/frontend/$BUILD_OUTDIR"
if [ ! -f "$DIST/index.html" ]; then
  echo "错误: 未找到产物 $DIST/index.html" >&2
  exit 1
fi

# 资源引用应与 base 设定一致：base='/' 时形如 /static/...；
# 若出现 /production/ 前缀而 base 不是 /production/，说明配置不匹配（会白屏）
if [ "$BUILD_BASE_URL" != "/production/" ] && grep -q '"/production/' "$DIST/index.html"; then
  echo "错误: 产物资源路径带 /production/ 前缀，与 base=${BUILD_BASE_URL} 不匹配" >&2
  echo "      请确认 VITE_OUTDIR / VITE_ADDRESS_BASE_URL 配置一致" >&2
  exit 1
fi

echo ""
echo "✓ 构建完成: $DIST"
du -sh "$DIST" | awk '{print "  体积: " $1}'
echo "  资源 base: ${BUILD_BASE_URL}"

# ---- 5. 可选：同步到目标目录 ----
if [ -n "$OUTDIR" ]; then
  echo "==> 同步到 $OUTDIR ..."
  mkdir -p "$OUTDIR"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete "$DIST/" "$OUTDIR/"
  else
    rm -rf "${OUTDIR:?}/"*
    cp -r "$DIST/." "$OUTDIR/"
  fi
  echo "✓ 已同步到 $OUTDIR"
  echo "  注意: 站点需配置 history 模式 fallback: try_files \$uri \$uri/ /index.html;"
fi
