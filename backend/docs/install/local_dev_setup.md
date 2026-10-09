---
title: 本地开发环境配置说明
---

# 本地开发环境配置说明

## 环境依赖

| 工具 | 版本 | 状态 |
| --- | --- | --- |
| Go | 1.23+（本机 go1.27.0 darwin/arm64） | 已就绪 |
| Node.js | >= 18（本机 v24.14.0） | 已就绪 |
| pnpm | 11.x（本机 11.17.0） | 已就绪 |
| MySQL | 8+ | 见下方「数据库」 |

Redis **不需要**。项目文档提到 Redis 缓存模块，但 `backend/` 代码中无任何 redis 引用，仅设计文档存在。

## 数据库（必需）

项目强依赖 MySQL：`backend/internal/database/migration.go` 在启动时执行 GORM `AutoMigrate`，自动建 6 张表并创建默认管理员 `admin / admin@123`。没有数据库后端无法启动（会卡在 200 次 × 3 秒的连接重试后 `log.Fatalf`）。

### 配置方式

真实连接信息写在 `backend/configs/config.local.yaml`，该文件已在 `.gitignore` 中，不进 Git。仓库提供模板：

```bash
cp backend/configs/config.local.yaml.example backend/configs/config.local.yaml
# 编辑 host / username / password
```

`backend/configs/config.local.yaml.example` 是可提交的模板，含全部字段说明。

### 表结构来源

- 启动时自动迁移：`backend/internal/database/migration.go`（权威来源）
- 历史 SQL 脚本：`backend/migrations/*.sql`（001 ~ 011，仅参考，不是运行时路径）

### 当前远程库状态（2026-10-09 已联通）

| 项 | 值 |
| --- | --- |
| 主机 | `cc.casibrain.lan` → 192.168.0.182:3306 |
| 版本 | MySQL 8.0.45 Community Server |
| 账号 | `lmcenter`@`%` |
| 库名 | `license_manager`（由 DBA 手工创建并授权） |
| 字符集 | `utf8mb4` / `utf8mb4_unicode_ci` |

首次启动时 `AutoMigrate` 自动创建 6 张表并初始化管理员：

| 表 | 初始行数 |
| --- | --- |
| `users` | 1（`admin`） |
| `customer_code_sequence` | 1（2026 年序列） |
| `customers` | 0 |
| `authorization_codes` | 0 |
| `licenses` | 0 |
| `authorization_changes` | 0 |

### 路径字段依赖工作目录

配置里的相对路径（`i18n.config_path`、`license.rsa.*_key_path`）都是按**工作目录**解析，不是按 `-config` 文件位置。仓库中所有配置写 `../configs/...`，对应工作目录为 `backend/cmd/`（Docker 场景）。

本地从 `backend/` 启动时，`config.local.yaml` 里必须改成不带 `../`：

```yaml
i18n:
  config_path: "configs/i18n/errors"
license:
  rsa:
    private_key_path: "configs/rsa_private_key.pem"
    public_key_path: "configs/rsa_public_key.pem"
```

漏改会导致 `多语言初始化失败: ... no such file or directory`（`main.go:98` 直接 `log.Fatalf`）。

## 启动后端

```bash
cd backend
go run ./cmd -config configs/config.local.yaml
```

注意：`-config` 的相对路径基于 **`backend/` 目录**（不是 `backend/cmd/`）。`backend/docs/install/dev_install.md` 里写的 `cd cmd && go run . -config configs/config.dev.yaml` 路径是错的 —— `configs` 不在 `cmd/` 下。

配置查找顺序（`backend/cmd/main.go`）：

1. `-config` / `-c` 显式指定的路径
2. `config.yaml`（当前工作目录，Docker 用）
3. `../configs/config.yaml`（本地开发回退路径）

健康检查：

```bash
curl http://localhost:18888/health
```

## 启动前端

```bash
cd frontend
pnpm install
pnpm run serve   # http://localhost:3000
```

### 接口地址

`frontend/.env.development` 指向远程 `43.136.177.43:18888`。本地开发用 `frontend/.env.development.local` 覆盖为 `http://localhost:18888`（Vite 中 `.local` 文件优先级更高，且已被 `.gitignore` 的 `frontend/*.local` 覆盖）。

### pnpm 11 构建脚本白名单

pnpm 11 不再读取 `package.json` 里的 `pnpm` 字段。`frontend/pnpm-workspace.yaml` 中 `allowBuilds` 放行 `esbuild`、`@parcel/watcher`、`vue-demi` 的安装脚本，否则 `pnpm run serve` 会因 `ERR_PNPM_IGNORED_BUILDS` 失败。

## 端口

| 服务 | 端口 |
| --- | --- |
| 后端 API | 18888 |
| 前端 dev server | 3000 |
| MySQL（远程） | 由你提供 |

## 已知事项

- `backend/docs/install/dev_install.md` 中的配置路径示例有误，正确写法见上文「启动后端」。
- `frontend/src/api/https/index.ts:12` 有 `console.log('API Base URL:', envUrl)`，开发环境会打印。生产构建会被 terser 移除（`drop_console`），无需处理。
- Vite `server.watch.usePolling: true`（`vite.config.ts:104`），为 WSL2 兼容而设，macOS 上会略增 CPU 占用，属既有配置。
- 根目录 `.env.example` 是 Docker Compose 生产编排用的，与本地 Go 运行无关。
- pnpm 11 不再读 `package.json` 的 `pnpm` 字段，`frontend/pnpm-workspace.yaml` 的 `allowBuilds` 是必需项，勿删。
