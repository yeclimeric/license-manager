#!/bin/sh
# 1panel Go 运行环境 —— 后端启动脚本
#
# 用法：在 1panel「创建 Go 运行环境」表单的「启动命令」字段填：
#   sh deploy/1panel/start-backend.sh
# （运行目录填项目仓库根目录，即包含 backend/ 与 deploy/ 的那一层）
#
# 为什么需要这个脚本：
#   后端配置中的 i18n.config_path 与 license.rsa.*_key_path 是【相对当前工作目录】解析的，
#   不是相对配置文件位置。从 backend/ 启动时对应 configs/...，
#   从 backend/cmd/ 启动时对应 ../configs/...（容器场景）。
#   cwd 错了会导致启动即 log.Fatalf（找不到 i18n 语言包 / RSA 密钥）。
#   本脚本自动定位项目根并切换到 backend/，消除对 cwd 的依赖。
#
# 兼容性：使用 POSIX sh 语法，兼容 bash / alpine busybox ash。

set -eu

# ---- 1. 定位项目根（从脚本所在位置向上找含 backend/cmd 的目录）----
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
PROJECT_ROOT=$(cd "$SCRIPT_DIR/../.." && pwd)

# 若脚本被复制到别处，回退为从当前工作目录向上探测
if [ ! -d "$PROJECT_ROOT/backend/cmd" ]; then
  PROJECT_ROOT=$(pwd)
  while [ "$PROJECT_ROOT" != "/" ] && [ ! -d "$PROJECT_ROOT/backend/cmd" ]; do
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
  done
fi

if [ ! -d "$PROJECT_ROOT/backend/cmd" ]; then
  echo "[start] 错误: 未找到 backend/cmd 目录，无法定位项目根" >&2
  echo "[start] 请确认「运行目录」已设置为包含 backend/ 的项目根目录" >&2
  exit 1
fi

cd "$PROJECT_ROOT"
echo "[start] 项目根: $PROJECT_ROOT"

# ---- 2. 环境变量兜底（1panel 表单未填时给出明确提示）----
# 敏感配置通过环境变量注入，覆盖 config 文件中的占位值：
#   DATABASE_PASSWORD -> database.password
#   AUTH_JWT_SECRET  -> auth.jwt.secret
if [ -z "${DATABASE_PASSWORD:-}" ]; then
  echo "[start] 提示: 未设置 DATABASE_PASSWORD，将使用配置文件中的 database.password" >&2
fi
if [ -z "${AUTH_JWT_SECRET:-}" ]; then
  echo "[start] 提示: 未设置 AUTH_JWT_SECRET，将使用配置文件中的 auth.jwt.secret" >&2
  echo "[start]       若为默认值，后端会自动生成随机密钥（重启即失效，不可依赖）" >&2
fi

# ---- 3. 准备配置文件 ----
CONFIG_FILE="$PROJECT_ROOT/backend/configs/config.local.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
  echo "[start] 错误: 缺少配置文件 $CONFIG_FILE" >&2
  echo "[start] 请先执行: cp backend/configs/config.local.yaml.example backend/configs/config.local.yaml" >&2
  echo "[start] 并按注释填写数据库地址/账号/口令" >&2
  exit 1
fi

# ---- 4. 准备可执行文件（优先用预编译产物，缺失时现场编译）----
BIN="$PROJECT_ROOT/bin/license-manager"
if [ -x "$BIN" ]; then
  echo "[start] 使用已有可执行文件: $BIN"
elif command -v go >/dev/null 2>&1; then
  echo "[start] 未找到预编译产物，开始编译 ..."
  (cd "$PROJECT_ROOT/backend" && go build -o "$BIN" cmd/main.go)
  echo "[start] 编译完成: $BIN"
else
  echo "[start] 错误: 无可执行文件且容器内无 go 工具链" >&2
  echo "[start] 请在宿主机编译后上传: cd backend && go build -o ../bin/license-manager cmd/main.go" >&2
  exit 1
fi

# ---- 5. 启动 ----
# cwd 必须是 backend/：配置中的相对路径（i18n.config_path、license.rsa.*_key_path）
# 基于工作目录解析，从 backend/ 启动时对应 configs/...
cd "$PROJECT_ROOT/backend"
echo "[start] 工作目录: $(pwd)"
echo "[start] 配置文件: configs/config.local.yaml"
echo "[start] 监听地址: 0.0.0.0:18888"

# exec 让后端成为容器主进程（PID 1），才能正确接收停止信号
exec "$BIN" -config configs/config.local.yaml
