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

# ---- 1.5 运行时诊断信息 ----
# 打印容器内关键路径，便于排查「缓存目录没有内容」「日志不刷新」这类问题。
echo "[start] ---- 运行环境诊断 ----"
echo "[start] 容器内工作目录(启动时): $(pwd)"
echo "[start] 源码是否挂载: $([ -f backend/go.mod ] && echo '是 (/app 为 CODE_DIR 挂载点)' || echo '否')"
if [ -d /go/pkg/mod ]; then
  echo "[start] 官方 mod 缓存挂载点: /go/pkg/mod ($(du -sh /go/pkg/mod 2>/dev/null | cut -f1))"
else
  echo "[start] 官方 mod 缓存挂载点: /go/pkg/mod (不存在)"
fi
echo "[start] 项目内 .cache: $([ -d .cache ] && du -sh .cache 2>/dev/null | cut -f1 || echo '不存在')"
echo "[start] 预编译产物: $([ -x bin/license-manager ] && ls -lh bin/license-manager | awk '{print $5}' || echo '不存在')"
echo "[start] 启动时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "[start] ------------------------"

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
# Go 缓存说明（依据 1panel 官方 Go 运行环境模板 apps/go/<版本>/docker-compose.yml）：
#   官方已持久化挂载 mod 缓存：  ./mod:/go/pkg/mod
#   即容器内 /go/pkg/mod 映射到运行环境目录下的 mod/，随运行环境持久化，容器重建不丢。
#   因此【不要】改写 GOMODCACHE——那会绕开官方已挂载的持久目录，造成两份 mod 缓存。
#   只有编译缓存 GOCACHE 默认在容器可写层（/root/.cache/go-build），重建后会丢，
#   故把它指向项目内 .cache/build（项目目录是挂载目录，同样持久）。
if command -v go >/dev/null 2>&1; then
  BUILD_CACHE_DIR="$PROJECT_ROOT/.cache/build"
  if mkdir -p "$BUILD_CACHE_DIR" 2>/dev/null; then
    GOCACHE="$BUILD_CACHE_DIR"
    export GOCACHE
    echo "[start] Go 编译缓存: ${BUILD_CACHE_DIR} (持久化)"
  else
    echo "[start] 提示: 无法创建 ${BUILD_CACHE_DIR}，编译缓存将写入容器可写层（重建后失效）"
  fi
  # GOMODCACHE 保持官方挂载位置，不覆盖
  echo "[start] Go 依赖缓存: $(go env GOMODCACHE) (官方挂载，持久化)"

  # 国内服务器直连 proxy.golang.org 常常很慢或超时，优先用 goproxy.cn
  # （用户已在 1panel 环境变量中设置 GOPROXY 时不覆盖）
  if [ -z "${GOPROXY:-}" ]; then
    export GOPROXY="https://goproxy.cn,direct"
    echo "[start] GOPROXY: ${GOPROXY} (国内加速，可在环境变量中覆盖)"
  fi
fi

BIN="$PROJECT_ROOT/bin/license-manager"
if [ -x "$BIN" ]; then
  echo "[start] 使用已有可执行文件: $BIN"
elif command -v go >/dev/null 2>&1; then
  echo "[start] 未找到预编译产物，开始编译（首次需下载依赖，约 1-3 分钟）..."
  echo "[start] 依赖缓存目录: $(go env GOMODCACHE)"
  # 先单独下载依赖：便于区分「网络问题」与「编译问题」，也便于失败时单独重试
  if ! (cd "$PROJECT_ROOT/backend" && go mod download); then
    echo "[start] 错误: 依赖下载失败。若为网络问题，可设置环境变量 GOPROXY=https://goproxy.cn,direct" >&2
    exit 1
  fi
  echo "[start] 依赖下载完成，开始编译 ..."
  if ! (cd "$PROJECT_ROOT/backend" && go build -o "$BIN" cmd/main.go); then
    echo "[start] 错误: 编译失败" >&2
    exit 1
  fi
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
