# Docker 部署指南

本项目提供两种部署：开发环境与生产环境。以下说明已与当前 `docker-compose` 配置、镜像与反向代理实现对齐。

## 开发环境

使用 `docker compose`（Docker Desktop 或 Docker CE 20.10+）：

```bash
# 启动（首次建议带 --build）
docker compose up -d --build

# 查看状态/日志
docker compose ps
docker compose logs -f backend

# 停止
docker compose down
```

### 访问地址（开发）
- 前端: http://localhost:18080
- 后端 API: http://localhost:18888
- MySQL: localhost:13306（数据库、用户、密码见 compose）

## 生产环境

使用 `docker-compose.prod.yml`：

```bash
# 启动生产环境（从源码构建，首次及更新代码后务必带 --build）
docker compose -f docker-compose.prod.yml up -d --build

# 查看服务状态
docker compose -f docker-compose.prod.yml ps
```

### 生产配置说明
- 后端配置文件已通过卷挂载至容器内：
  - 开发：`backend/configs/config.dev.yaml` → 容器内 `/app/backend/cmd/config.yaml`
  - 生产：`backend/configs/config.prod.yaml` → 容器内 `/app/backend/cmd/config.yaml`
- 前端 Nginx 反向代理已指向容器网络内的 `backend:18888`，无需依赖 `host.docker.internal`。
- 健康检查统一使用 `curl`，后端镜像已内置 `curl`。
- 生产 `docker-compose.prod.yml` 已从源码构建后端/前端镜像（不再依赖预构建镜像或 registry）；更新代码后须用 `docker compose -f docker-compose.prod.yml up -d --build` 重新构建。
- 前端默认映射到宿主机高位端口 `18880`，避免与 1panel/OpenResty 的 `80` 冲突；正式访问请用反向代理 + HTTPS。
- **部署配置已收敛到 `.env` 单点真相**：DB 口令（`MYSQL_PASSWORD`）与 JWT 密钥（`JWT_SECRET`）在 `.env` 定义一次，通过环境变量 `DATABASE_PASSWORD` / `AUTH_JWT_SECRET` 注入后端，自动覆盖 `config.prod.yaml`，**无需再手动改 yaml**。首次部署运行 `bash scripts/deploy-prepare.sh` 可一键生成 `.env`、随机 `JWT_SECRET` 与 RSA 密钥。
- 如需隐藏后端对外端口，可在生产 compose 中移除后端端口映射，仅通过前端或网关暴露。

## 生产环境上线前必改项

部署到公网前，必须逐项确认以下安全项，否则存在密钥泄露 / 授权被伪造的风险：

1. **JWT 密钥**：敏感配置已收敛到 `.env`（单点真相）。在 `.env` 设置 `JWT_SECRET` 为强随机串（`openssl rand -hex 32`），或直接运行 `scripts/deploy-prepare.sh` 自动生成并写入。
   - 该值通过环境变量 `AUTH_JWT_SECRET` 注入后端，覆盖 `config.prod.yaml` 的 `auth.jwt.secret`，**无需再手动改 yaml**。
   - 若 `.env` 未设置 `JWT_SECRET`，后端会用 yaml 默认值触发**自动生成随机密钥**并打印 `[WARN]`；该密钥仅内存态，**重启即失效、令牌作废**，仅兜底，不可依赖。
2. **数据库口令**：只需在 `.env` 设置 `MYSQL_PASSWORD`（及 `MYSQL_ROOT_PASSWORD`）。它通过环境变量 `DATABASE_PASSWORD` 注入后端，自动与 `config.prod.yaml` 的 `database.password` 保持一致，**无需在两处分别修改**。
3. **RSA 密钥对**：仓库自带的 `rsa_private_key.pem` / `rsa_public_key.pem` 私钥已随仓库公开，**生产必须重新生成专属密钥对并替换**，否则授权签名可被伪造。`scripts/deploy-prepare.sh` 会在密钥缺失时自动生成；也可手动运行 `./scripts/gen-rsa-keys.sh`（或仓库内置 Go 工具 `backend/cmd/gen-rsa-keys`），默认输出到 `backend/configs`，覆盖需加 `--force`。
4. **默认管理员**：首次启动自动 seeding 的 `admin / admin@123`，登录后立即修改密码。
5. **端口冲突**：1panel 自带 OpenResty 占用宿主机 `80/443`。本仓库提供的 `docker-compose.1panel.yml` 前端映射到高位端口 `18880`，**不要**改回 `80`；正式访问请用 1panel「网站」做反向代理 + Let's Encrypt HTTPS。
6. **暴露面**：公网仅暴露前端（经反代），后端 `18888` 不映射到宿主机。
7. **Redis**：代码未实际使用，已从编排中移除，无需部署；如自行启用请先确认业务确有依赖。
8. **配置相对路径**：`config.prod.yaml` 中的 `i18n.config_path`、`license.rsa.*_key_path` 相对 cwd（`/app/backend/cmd`）解析，编排已挂载 `./backend/configs` 保证路径有效，勿删该挂载。

## 1panel 部署（推荐）

若服务器已安装 1panel，可直接用本项目提供的 `docker-compose.1panel.yml`：

1. 将仓库 clone 到服务器某目录（如 `/opt/1panel/docker/compose/license-manager`）。
2. 1panel「Docker → 编排 → 创建编排」，编排目录选该仓库根目录，compose 文件选 `docker-compose.1panel.yml`（或把内容粘贴进去）。
3. 创建并启动，1panel 会执行 `docker compose up -d --build` 从源码构建镜像。
4. 1panel「网站 → 创建网站 → 反向代理」，目标 `http://127.0.0.1:18880`，再一键申请 HTTPS。
5. 启动前先按上方「生产环境上线前必改项」修改 `config.prod.yaml`。

## 服务说明

### MySQL 8.0
- 端口：3306（生产）/ 13306（开发宿主机映射）
- 库：`license_manager`，用户：`license_user`
- 字符集/排序：`utf8mb4`/`utf8mb4_unicode_ci`
- 开启慢查询日志（生产）

### 后端服务
- 端口：18888
- 健康检查：`/health`
- 配置：通过卷挂载 `config.dev.yaml` / `config.prod.yaml` 为容器内 `config.yaml`
- 日志：标准输出 + `./logs` 挂载目录

### 前端服务
- 端口：18880（生产）/ 18080（开发映射）
- 反向代理：`/api/` → `backend:18888`
- 静态资源：启用 Gzip 与长缓存

### Redis（已省略）
- 本项目代码未实际使用 Redis，生产编排已移除该服务以降低攻击面与资源占用；如确有依赖，可参照历史版本自行添加 redis 服务与 `redis_data` 卷。

## 常用命令

```bash
# 重新构建镜像
docker compose build

# 启动指定服务
docker compose up -d mysql backend

# 进入容器
docker compose exec backend sh
docker compose exec mysql mysql -uroot -p

# 备份/恢复数据库
docker compose exec mysql mysqldump -uroot -p license_manager > backup.sql
docker compose exec -T mysql mysql -uroot -p license_manager < backup.sql

# 清理未使用资源
docker system prune -f
```

## 故障排查

1) MySQL 启动失败
```bash
docker compose logs mysql
docker compose down -v && docker compose up -d
```

2) 后端连接数据库失败
```bash
# 确认 MySQL 健康
docker compose ps

# 检查容器内配置（已挂载到 /app/backend/cmd/config.yaml）
docker compose exec backend cat /app/backend/cmd/config.yaml
```

3) 前端无法访问后端 API
```bash
# 容器内连通性
docker compose exec frontend sh -c "curl -sS http://backend:18888/health"

# 检查 nginx 配置
docker compose exec frontend cat /etc/nginx/conf.d/default.conf
```

## 注意事项

- 生产环境请修改默认账号、数据库口令与 JWT 秘钥（见 `backend/configs/config.prod.yaml`）。
- 如使用较新 Docker，YAML 顶部 `version` 字段会被忽略，可按需移除提示。
- `deploy.resources` 仅在 Swarm/K8s 生效，单机 Compose 可忽略。
- 如置于公网，建议仅暴露前端端口，通过前端或网关转发后端 API。