#!/bin/bash
#
# 本地开发环境启动脚本
#
# 用法：
#   ./scripts/dev.sh                # 用 config.dev.yaml（docker-compose 的 mysql 主机）
#   ./scripts/dev.sh config.local.yaml
#   LM_CONFIG=config.local.yaml ./scripts/dev.sh
#
# 注意：config.local.yaml 含真实数据库凭据，已被 .gitignore 排除，
# 不会随仓库分发。换机器开发时需自行创建。

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# 配置选择：命令行参数优先，其次环境变量，最后 dev
CONFIG_FILE="${1:-${LM_CONFIG:-config.dev.yaml}}"
CONFIG_PATH="$ROOT_DIR/backend/configs/$CONFIG_FILE"

if [ ! -f "$CONFIG_PATH" ]; then
  echo "错误：找不到配置文件 $CONFIG_PATH"
  echo "可选：$(cd "$ROOT_DIR/backend/configs" && ls config.*.yaml | tr '\n' ' ')"
  exit 1
fi

# 端口从配置里读，避免脚本里的注释与实际端口不一致
BACKEND_PORT="$(grep -A 2 '^server:' "$CONFIG_PATH" | grep 'port:' | head -1 | sed 's/.*port: *//' | tr -d '"' )"
BACKEND_PORT="${BACKEND_PORT:-18888}"
FRONTEND_PORT="${LM_FRONTEND_PORT:-3000}"

PIDS=()

cleanup() {
  echo ""
  echo "停止服务..."
  for pid in "${PIDS[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      kill -TERM "$pid" 2>/dev/null || true
    fi
  done
  sleep 1
  for pid in "${PIDS[@]}"; do
    kill -KILL "$pid" 2>/dev/null || true
  done
  # go run 会先编译再启动真正的二进制，杀掉 go 进程后编译产物仍占着端口，
  # 因此按端口再清一次。只清本脚本启动的这两个端口，不动其他服务。
  for port in "$BACKEND_PORT" "$FRONTEND_PORT"; do
    pids_on_port="$(lsof -ti "$port" 2>/dev/null || true)"
    if [ -n "$pids_on_port" ]; then
      echo "$pids_on_port" | xargs kill -KILL 2>/dev/null || true
    fi
  done
  echo "已停止"
}
trap cleanup EXIT INT TERM

echo "配置：$CONFIG_FILE"

# 启动后端
echo "启动后端 (:$BACKEND_PORT)..."
cd "$ROOT_DIR/backend"
go run ./cmd -config "$CONFIG_PATH" &
BACKEND_JOB=$!
PIDS+=($BACKEND_JOB)

# 等后端就绪再起前端，否则前端首屏请求会失败
echo "等待后端就绪..."
for i in $(seq 1 60); do
  if curl -sf "http://localhost:$BACKEND_PORT/health" >/dev/null 2>&1; then
    echo "后端已就绪"
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "警告：后端 60 秒内未就绪，继续启动前端"
  fi
  sleep 1
done

# 启动前端
# 用 pnpm：项目已迁移，npm run dev 不存在（脚本名是 serve）
echo "启动前端 (:$FRONTEND_PORT)..."
cd "$ROOT_DIR/frontend"
pnpm run serve --port "$FRONTEND_PORT" &
FRONTEND_JOB=$!
PIDS+=($FRONTEND_JOB)

sleep 3

echo ""
echo "─────────────────────────────"
echo "  后端  http://localhost:$BACKEND_PORT"
echo "  前端  http://localhost:$FRONTEND_PORT"
echo ""
echo "  默认账号  admin / admin@123"
echo "  按 Ctrl+C 停止"
echo "─────────────────────────────"

# 后台任务不会阻塞 wait，显式等待用户中断
while true; do
  # 任一进程意外退出则整体退出，避免只剩半边服务
  for pid in "${PIDS[@]}"; do
    if ! kill -0 "$pid" 2>/dev/null; then
      echo "进程 $pid 已退出，正在停止其余服务"
      exit 1
    fi
  done
  sleep 2
done