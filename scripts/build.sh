#!/usr/bin/env bash
# 构建前后端（等价于依次执行 build-backend.sh 与 build-frontend.sh）
#
# 保留此脚本以兼容既有调用方；新脚本功能更完整（环境检查、交叉编译、产物校验、
# 输出目录同步），建议直接使用：
#   bash scripts/build-backend.sh
#   bash scripts/build-frontend.sh
set -euo pipefail

cd "$(dirname "$0")/.." || exit 1

echo "==> 构建后端 ..."
bash scripts/build-backend.sh

echo ""
echo "==> 构建前端 ..."
bash scripts/build-frontend.sh

echo ""
echo "✓ 全部构建完成"
echo "  后端二进制: ./bin/license-manager"
echo "  前端产物:   ./frontend/dist"
