#!/usr/bin/env bash
# MySQL 授权脚本（容器化 / 跨主机场景）
#
# 用途：为 License Manager 创建数据库与专用用户，并按来源主机授权。
# 背景：MySQL 的用户由「用户名 + 来源 host」共同标识，
#      跨主机/跨容器访问时，授权的 host 必须覆盖实际来源 IP，
#      否则报 ERROR 1045 (28000): Access denied for user 'x'@'<来源IP>'。
#
# 用法（在 MySQL 所在主机或能直连 MySQL 的机器上执行）：
#   bash scripts/mysql-grant.sh <MySQL地址> <授权来源IP> [用户名] [口令] [库名]
#
# 示例（1panel 容器场景，后端容器 IP 为 172.18.0.2）：
#   bash scripts/mysql-grant.sh 172.18.0.3 172.18.0.2 license_user '你的强口令' license_manager
#
# 授权来源 IP 传 % 表示放行任意来源（内网环境可用）。
# 默认值：地址 127.0.0.1、来源 %、用户 license_user、库 license_manager
# 口令必须显式传入（不设默认值，避免误用弱口令）。
#
# ⚠️ MySQL 的用户名严格区分「连字符 -」与「下划线 _」：
#    'license-manager' 与 'license_manager' 是两个不同的用户！
#    报 Access denied 时，先核对 mysql.user 表的 user 列与后端配置是否完全一致。

set -euo pipefail

MYSQL_HOST="${1:-127.0.0.1}"
GRANT_HOST="${2:-%}"
DB_USER="${3:-license_user}"
DB_PASS="${4:-}"
DB_NAME="${5:-license_manager}"

if [ -z "$DB_PASS" ]; then
  echo "错误: 必须提供数据库口令。用法: $0 <MySQL地址> <授权来源IP> <用户名> <口令> [库名]" >&2
  exit 1
fi

if ! command -v mysql >/dev/null 2>&1; then
  echo "错误: 未找到 mysql 客户端。请在能访问 MySQL 的机器上执行，" >&2
  echo "      或进 MySQL 容器内执行: docker exec -it <mysql容器> bash" >&2
  exit 1
fi

echo "==> 目标: $DB_USER@$GRANT_HOST  →  MySQL($MYSQL_HOST)  库: $DB_NAME"

# ---- 用户名相似性检查（本次事故的教训）----
# MySQL 的用户名区分「连字符 -」与「下划线 _」：'license-manager' ≠ 'license_manager'。
# 若 mysql.user 中存在仅在 -/_ 等字符上不同的近似用户名，直接给出提示，
# 避免反复出现「Access denied」却查不出原因。
existing_users=$(mysql -h "$MYSQL_HOST" -uroot -p -N -B -e \
  "SELECT DISTINCT user FROM mysql.user;" 2>/dev/null || true)

if [ -n "$existing_users" ]; then
  normalized_target=$(printf '%s' "$DB_USER" | tr -d '_-')
  similar=""
  for u in $existing_users; do
    [ "$u" = "$DB_USER" ] && continue
    normalized_u=$(printf '%s' "$u" | tr -d '_-')
    if [ "$normalized_u" = "$normalized_target" ]; then
      similar="$similar '$u'"
    fi
  done
  if [ -n "$similar" ]; then
    echo ""
    echo "⚠️ 警告: MySQL 中存在与 '${DB_USER}' 极为相似的用户名:${similar}"
    echo "   MySQL 严格区分连字符(-)与下划线(_)，它们是不同的用户！"
    echo "   若后端配置用的是 '${DB_USER}'，则必须存在 '${DB_USER}' 这个用户；"
    echo "   或者把后端配置改成已有用户名。请用下面的命令核对："
    echo "     SELECT user, host FROM mysql.user;"
  fi
fi

# 口令中若含单引号或反斜杠，需转义后再嵌入 SQL 字符串字面量，
# 否则会破坏语句导致授权失败（表现为「授权没生效」）。
esc_pwd() {
  printf '%s' "$1" | sed "s/\\\\/\\\\\\\\/g; s/'/\\\\'/g"
}
SQL_PASS="$(esc_pwd "$DB_PASS")"

# 用 SQL 变量传参，避免口令中的特殊字符破坏语句
mysql -h "$MYSQL_HOST" -uroot -p <<SQL
-- 1) 建库（utf8mb4，授权码/证书含中文与 emoji）
CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\`
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 2) 创建用户并授权
--    用户由「用户名 + 来源 host」共同标识。
--    来源 host 需覆盖后端实际发起的连接来源 IP，否则 Access denied。
CREATE USER IF NOT EXISTS '${DB_USER}'@'${GRANT_HOST}' IDENTIFIED BY '${SQL_PASS}';
ALTER USER '${DB_USER}'@'${GRANT_HOST}' IDENTIFIED BY '${SQL_PASS}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'${GRANT_HOST}';

-- 3) 同时放行 localhost（同机场景，如调试脚本）
CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED BY '${SQL_PASS}';
ALTER USER '${DB_USER}'@'localhost' IDENTIFIED BY '${SQL_PASS}';
GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'localhost';

FLUSH PRIVILEGES;

-- 4) 回显确认
SELECT user, host, plugin FROM mysql.user WHERE user = '${DB_USER}';
SQL

echo ""
echo "✓ 授权完成。请用以下方式验证（把地址/端口/来源替换为实际值）："
echo "  mysql -h ${MYSQL_HOST} -u${DB_USER} -p -e \"SELECT 1;\" ${DB_NAME}"
echo ""
echo "  注意：后端配置 database.username 必须是 '${DB_USER}'，"
echo "        database.database 才是 '${DB_NAME}'，两者不要填反。"
