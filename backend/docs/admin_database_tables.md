# 管理端数据库表设计文档

> 本文档梳理 License Manager 社区版管理端业务所涉及的数据库表结构，覆盖客户、授权码、许可证和管理员用户。
>
> 文档内容基于：
> - `backend/migrations/*.sql`（表结构、索引、外键）
> - `backend/internal/models/*.go`（GORM 模型定义）
> - `backend/internal/api/routes/router.go`（管理端路由）

---

## 1. 表总览

| 序号 | 表名 | 中文名 | 作用 | 所属迁移文件 |
| --- | --- | --- | --- | --- |
| 1 | `users` | 管理员用户表 | 管理端登录账号、角色权限 | `006_create_users_table.sql` |
| 2 | `customers` | 客户表 | 客户基本信息、分级、状态 | `001_create_customers_table.sql` |
| 3 | `customer_code_sequence` | 客户编码序列表 | 按年度生成客户编码 `CUS-YYYY-NNNN` | `001_create_customers_table.sql` |
| 4 | `authorization_codes` | 授权码表 | 客户授权的"业务配置容器" | `003_create_authorization_codes_table.sql`、`009_update_authorization_codes_code_length.sql` |
| 5 | `authorization_changes` | 授权变更历史表 | 授权码续费/升级/锁定等操作的审计日志 | `005_create_authorization_changes_table.sql` |
| 6 | `licenses` | 许可证表 | 设备激活凭证、心跳、硬件绑定 | `004_create_licenses_table.sql` |

> 仪表盘（`dashboard`）相关数据由 `authorization_codes` 与 `licenses` 派生统计而来，不存在独立表。

---

## 2. 通用约定

- **数据库**：MySQL 8.0+ / MariaDB 10.3+
- **字符集**：`utf8mb4`、`utf8mb4_0900_ai_ci`
- **主键**：使用 UUID 字符串（`VARCHAR(36)`），由 GORM `BeforeCreate` 钩子或数据库 `UUID()` 默认值生成
- **时间戳**：`DATETIME(3)` 匹配 Go `time.Time`；由 GORM 维护，不由数据库默认值控制（部分迁移文件中已显式声明）
- **软删除**：主要业务表（`customers`、`licenses` 等）通过 `deleted_at` 软删除
- **枚举**：使用 `VARCHAR` + 应用层 `oneof` 校验，避免数据库 ENUM 带来的迁移成本
- **JSON 字段**：MySQL 5.7+ 原生 `JSON` 类型，对应 Go 中的 `JSON` 类型

---

## 3. 表结构详解

### 3.1 users — 管理员用户表

管理端操作员/管理员/查看者账号，承载登录认证与角色权限。

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `VARCHAR(36)` | ✅ | UUID | 主键 |
| `username` | `VARCHAR(50)` | ✅ | — | 用户名，唯一 |
| `email` | `VARCHAR(255)` | ✅ | — | 邮箱，唯一 |
| `password_hash` | `VARCHAR(255)` | ✅ | — | bcrypt 哈希，成本因子 12 |
| `full_name` | `VARCHAR(100)` | ✅ | — | 真实姓名 |
| `phone` | `VARCHAR(20)` | ❌ | NULL | 手机号 |
| `role` | `VARCHAR(20)` | ✅ | `viewer` | 角色：`admin` / `operator` / `viewer` |
| `status` | `VARCHAR(20)` | ✅ | `active` | 账号状态：`active` / `disabled` / `locked` |
| `last_login_at` | `DATETIME(3)` | ❌ | NULL | 最后登录时间 |
| `last_login_ip` | `VARCHAR(45)` | ❌ | NULL | 最后登录 IP（兼容 IPv6） |
| `login_attempts` | `INT` | ✅ | `0` | 登录失败次数（连续 5 次失败锁定 30 分钟） |
| `locked_until` | `DATETIME(3)` | ❌ | NULL | 账号锁定到期时间 |
| `created_at` | `DATETIME(3)` | ✅ | — | 创建时间 |
| `updated_at` | `DATETIME(3)` | ✅ | — | 更新时间 |

**索引**：`username`、`email`（唯一）、`role`、`status`、`last_login_at`、`created_at`、`locked_until`；复合索引 `(role, status)`、`(status, locked_until)`。

**业务规则**：
- 密码使用 bcrypt 哈希存储。
- 连续 5 次登录失败锁定 30 分钟，由应用层控制。
- 默认管理员账号需要在应用初始化时创建。

---

### 3.2 customers — 客户表

管理端管理的客户主数据，是授权码和许可证的归属主体。

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `VARCHAR(36)` | ✅ | UUID | 主键 |
| `customer_code` | `VARCHAR(20)` | ✅ | — | 客户编码，唯一，格式 `CUS-YYYY-NNNN` |
| `customer_name` | `VARCHAR(200)` | ✅ | — | 客户名称 |
| `customer_type` | `VARCHAR(20)` | ✅ | `enterprise` | 客户类型：`individual` / `enterprise` / `government` / `education` |
| `contact_person` | `VARCHAR(100)` | ✅ | — | 联系人姓名 |
| `contact_title` | `VARCHAR(100)` | ❌ | NULL | 联系人职位 |
| `email` | `VARCHAR(255)` | ❌ | NULL | 邮箱 |
| `phone` | `VARCHAR(20)` | ❌ | NULL | 电话 |
| `address` | `TEXT` | ❌ | NULL | 地址 |
| `company_size` | `VARCHAR(20)` | ❌ | NULL | 企业规模：`small` / `medium` / `large` / `enterprise` |
| `customer_level` | `VARCHAR(20)` | ✅ | `basic` | 客户等级：`basic`(1星) / `active`(2星) / `vip`(3星) / `strategic`(4星) / `other`(5星)；展示为星级 |
| `status` | `VARCHAR(20)` | ✅ | `active` | 状态：`active` / `disabled` |
| `description` | `TEXT` | ❌ | NULL | 备注/描述 |
| `created_at` | `DATETIME(3)` | ✅ | — | 创建时间 |
| `updated_at` | `DATETIME(3)` | ✅ | — | 更新时间 |
| `created_by` | `VARCHAR(36)` | ✅ | — | 创建人（管理员用户 ID） |
| `updated_by` | `VARCHAR(36)` | ❌ | NULL | 更新人 |
| `deleted_at` | `DATETIME(3)` | ❌ | NULL | 软删除时间 |

**索引**：`customer_code`（唯一）、`customer_name`、`customer_type`、`customer_level`、`status`、`created_at`、`created_by`、`deleted_at`。

**外键**：作为 `authorization_codes.customer_id`、`licenses.customer_id` 的父表（通常 `ON DELETE RESTRICT`）。

**编码生成规则**：
- 格式：`CUS-YYYY-NNNN`，如 `CUS-2026-0001`。
- 按年递增，通过 `customer_code_sequence` 表配合应用层生成。

---

### 3.3 customer_code_sequence — 客户编码序列表

支撑客户编码 `CUS-YYYY-NNNN` 按年度自增。

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `year` | `INT` | ✅ | — | 主键，年份 |
| `sequence_number` | `INT` | ✅ | `0` | 当前年度序列号 |

**使用方式**：应用层先 `UPDATE ... SET sequence_number = sequence_number + 1`，再读取当前序列号并格式化为 `CUS-{year}-{seq:04d}`。

---

### 3.4 authorization_codes — 授权码表

管理端创建并下发给客户的**业务配置容器**，客户在客户端软件使用授权码完成激活。是整个授权体系的核心表。

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `VARCHAR(36)` | ✅ | UUID | 主键 |
| `code` | `VARCHAR(1000)` | ✅ | — | 唯一授权码，软件身份凭证（支持自包含配置/HMAC 签名） |
| `customer_id` | `VARCHAR(36)` | ✅ | — | 所属客户 ID |
| `created_by` | `VARCHAR(36)` | ✅ | — | 创建人（管理员用户 ID） |
| `software_id` | `VARCHAR(50)` | ❌ | NULL | 目标软件产品 ID |
| `description` | `TEXT` | ❌ | NULL | 授权码描述 |
| `start_date` | `DATETIME(3)` | ✅ | — | 授权生效时间 |
| `end_date` | `DATETIME(3)` | ✅ | — | 授权失效时间 |
| `software_version` | `VARCHAR(50)` | ❌ | NULL | 支持的软件版本范围 |
| `max_activations` | `INT` | ✅ | `1` | 最大激活次数 |
| `is_locked` | `BOOLEAN` | ✅ | `FALSE` | 是否被锁定 |
| `lock_reason` | `TEXT` | ❌ | NULL | 锁定原因 |
| `locked_at` | `DATETIME(3)` | ❌ | NULL | 锁定时间 |
| `locked_by` | `VARCHAR(36)` | ❌ | NULL | 锁定操作人（管理员用户 ID） |
| `feature_config` | `JSON` | ❌ | NULL | 功能配置（软件产品基础功能开关） |
| `usage_limits` | `JSON` | ❌ | NULL | 使用量限制（如用户数、API 调用次数） |
| `custom_parameters` | `JSON` | ❌ | NULL | 自定义参数（用户定制功能） |
| `created_at` | `DATETIME(3)` | ✅ | — | 创建时间 |
| `updated_at` | `DATETIME(3)` | ✅ | — | 更新时间 |

**索引**：
- `code` 的唯一约束以**前缀索引**实现：`UNIQUE INDEX idx_authorization_codes_code_prefix (code(255))`，避开 InnoDB 3072 字节索引长度限制。
- 普通索引：`customer_id`、`software_id`、`start_date`、`end_date`、`created_by`、`created_at`、`is_locked`、`locked_at`。
- 复合索引：`(customer_id, start_date, end_date)`。

**外键**：`customer_id → customers(id) ON DELETE RESTRICT`。

**状态字段**（虚字段，由应用层计算）：
- `normal`：当前时间在 `start_date` 与 `end_date` 之间，且 `is_locked = false`。
- `locked`：`is_locked = true`。
- `expired`：当前时间超过 `end_date`。

**字段长度演进**：最初 `code` 为 `VARCHAR(100)`，迁移 `009_update_authorization_codes_code_length.sql` 将其扩展为 `VARCHAR(1000)`，并改造为前缀唯一索引，以支持 HMAC 自包含配置授权码。

---

### 3.5 authorization_changes — 授权变更历史表

记录授权码所有变更操作的不可变审计日志。

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `VARCHAR(36)` | ✅ | UUID | 主键 |
| `authorization_code_id` | `VARCHAR(36)` | ✅ | — | 授权码 ID |
| `change_type` | `VARCHAR(30)` | ✅ | — | 变更类型：`renewal` / `upgrade` / `limit_change` / `feature_toggle` / `lock` / `unlock` / `other` |
| `old_config` | `JSON` | ❌ | NULL | 变更前配置快照 |
| `new_config` | `JSON` | ❌ | NULL | 变更后配置快照 |
| `operator_id` | `VARCHAR(36)` | ✅ | — | 操作人（管理员用户 ID） |
| `reason` | `TEXT` | ❌ | NULL | 变更原因 |
| `created_at` | `DATETIME(3)` | ✅ | — | 记录创建时间 |

**索引**：`authorization_code_id`、`change_type`、`operator_id`、`created_at`；复合索引 `(authorization_code_id, change_type)`、`(authorization_code_id, created_at)`、`(operator_id, created_at)`。

**外键**：`authorization_code_id → authorization_codes(id) ON DELETE CASCADE`。

**特殊约束**：
- 表只有 `created_at`，没有 `updated_at`、`deleted_at`，历史记录不可修改或删除。
- `old_config` / `new_config` 存储完整快照，便于审计与回滚。

---

### 3.6 licenses — 许可证表

设备激活凭证，与授权码一对多，记录硬件绑定与心跳信息。

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `id` | `VARCHAR(36)` | ✅ | UUID | 主键 |
| `license_key` | `VARCHAR(200)` | ✅ | — | 许可证密钥，设备特定，唯一 |
| `authorization_code_id` | `VARCHAR(36)` | ✅ | — | 关联授权码 ID |
| `customer_id` | `VARCHAR(36)` | ✅ | — | 客户 ID（冗余字段便于查询） |
| `hardware_fingerprint` | `VARCHAR(200)` | ✅ | — | 绑定的硬件指纹 |
| `device_info` | `JSON` | ❌ | NULL | 设备信息（CPU、内存等） |
| `activation_ip` | `VARCHAR(45)` | ❌ | NULL | 激活时 IP |
| `status` | `VARCHAR(20)` | ✅ | `inactive` | 状态：`active` / `inactive` / `revoked` |
| `activated_at` | `DATETIME(3)` | ❌ | NULL | 激活时间 |
| `last_heartbeat` | `DATETIME(3)` | ❌ | NULL | 最后心跳时间 |
| `last_online_ip` | `VARCHAR(45)` | ❌ | NULL | 最后在线 IP |
| `config_updated_at` | `DATETIME(3)` | ❌ | NULL | 配置更新时间 |
| `usage_data` | `JSON` | ❌ | NULL | 使用数据（软件上报统计信息） |
| `created_at` | `DATETIME(3)` | ✅ | — | 创建时间 |
| `updated_at` | `DATETIME(3)` | ✅ | — | 更新时间 |
| `deleted_at` | `DATETIME(3)` | ❌ | NULL | 软删除时间 |

**索引**：`license_key`（唯一）、`authorization_code_id`、`customer_id`、`hardware_fingerprint`、`status`、`activated_at`、`last_heartbeat`、`config_updated_at`、`created_at`、`deleted_at`；
复合索引：`(customer_id, status)`、`(authorization_code_id, status)`、`(status, activated_at)`、`(last_heartbeat, status)`。

**唯一约束**：`UNIQUE KEY uk_licenses_auth_hardware (authorization_code_id, hardware_fingerprint)` — 同一授权码下同一硬件指纹只能有一条许可证。

**外键**：
- `authorization_code_id → authorization_codes(id) ON DELETE CASCADE`
- `customer_id → customers(id) ON DELETE RESTRICT`

**在线状态**（虚字段，由应用层基于 `last_heartbeat` 计算）：
- `online`：最近 5 分钟内有心跳。
- `offline`：5 分钟–24 小时之间。
- `abnormal`：超过 24 小时未上报。

---


## 4. 表关系图

```
customers ──< authorization_codes ──< licenses
                    │
                    └──< authorization_changes

users
```

## 5. 关键设计要点

1. **UUID 主键**：业务表使用 `VARCHAR(36)` UUID 主键，避免自增 ID 暴露业务量级，便于分布式部署。
2. **前缀唯一索引**：`authorization_codes.code` 因字段较长，采用 `code(255)` 前缀唯一索引避开 InnoDB 3072 字节索引长度限制，碰撞概率极低。
3. **软删除一致性**：`customers`、`licenses` 等使用 `deleted_at DATETIME(3)` 软删除，`authorization_codes` 与 `authorization_changes` 不做软删除（前者的状态由字段表达，后者作为审计日志不可变）。
4. **状态虚字段**：授权码的 `status`（`normal`/`locked`/`expired`）与许可证的 `is_online`（`online`/`offline`/`abnormal`）均由应用层基于时间字段计算，避免每次写入维护。
5. **变更审计**：所有对授权码的写操作（创建、更新、锁定、解锁）均在 `authorization_changes` 留下 `old_config`/`new_config` 快照，便于审计与回滚分析。
6. **外键策略**：
   - 业务父表（如 `customers`）使用 `ON DELETE RESTRICT`，防止级联误删。
   - `authorization_changes → authorization_codes` 使用 `CASCADE`，授权码删除时一并清理历史。

---

## 6. 字段演进记录

| 迁移文件 | 影响表 | 变更说明 |
| --- | --- | --- |
| `009_update_authorization_codes_code_length.sql` | `authorization_codes` | `code VARCHAR(100) → VARCHAR(1000)`，唯一索引改为 `code(255)` 前缀索引，支持 HMAC 自包含配置授权码 |

---

## 7. 文档维护

- 任何对上述表结构的变更需同步：
  1. 新增/修改 `backend/migrations/*.sql`；
  2. 更新对应 `backend/internal/models/*.go`；
  3. 同步更新本文档。
- 新增管理端业务表时，需在此文档"表总览"与"表关系图"中同步登记。
