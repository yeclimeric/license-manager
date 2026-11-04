# RFC：审计日志模块健壮性增强 —— 资源类型一致性 / 覆盖度 / 脱敏 / 异步化

> 状态：待评审
> 里程碑：架构深化 M1
> 标签：`architecture` `rfc` `audit` `security`
> 关联：#1（许可证签发）· #6（跨切面契约）· #7（前端契约）· 候选 2（架构审计 6 候选中唯一未处理项）

## Problem

审计日志采用 `AuditMiddleware` + `AuditResponseCapture` **旁路中间件**自动记录所有写操作（POST/PUT/PATCH/DELETE），设计上较成熟：旁路写入（失败不阻断业务）、响应体捕获判定成败、路径推断资源、截断防膨胀、全局挂在 `auth` 组覆盖全部用户写操作。

但审计发现该模块存在**1 个真实数据一致性 bug** 与若干**健壮性/合规性缺口**：资源类型（`resource_type`）的写入值从未与规范常量对齐，导致按类型筛选审计记录会落空；此外覆盖度、敏感信息脱敏、写入性能与事务一致性仍有改进空间。

## 已核实现状（代码证据，非臆测）

- `internal/api/middleware/audit.go`：`AuditMiddleware`、`AuditResponseCapture`、`resolveAuditResource`、`buildAuditRecord`、`isUUIDLike`、`truncateAuditBody`。
- `internal/models/audit_log.go`：`AuditResourceAuthorizationCode = "authorization-code"` 等 6 个资源常量 + `AuditStatusSuccess/Failure`。
- 挂载：`router.go:110-113` 在 `auth` 组按 `AuthMiddleware → AuditResponseCapture → AuditMiddleware` 顺序挂载；`api` 组（96 行）与 `admin` 组（238 行）仅挂鉴权，无审计。
- `username` 由 `AuthMiddleware`（auth.go:93-94）与 `APIKeyAuthMiddleware`（api_key.go:52）设置；审计读 `c.GetString("username")`。
- `AuditLogService` 接口仅 `GetAuditLogList` / `GetStatistics` 两个**读**方法；`AuditLogRepository.Create` 存在但**无业务调用点**——写入完全由中间件旁路完成。

## 审计发现

### P0 — 资源类型规范化完全缺失（真实 bug）

- `models.AuditResource*` 常量**全项目零引用**（grep 确认：除定义外无任何消费）。
- `resolveAuditResource` 直接把 URL 路径首段写入 `resource_type`：`/v1/authorization-codes/{id}` → `"authorization-codes"`（复数、连字符），而常量定义是 `"authorization-code"`（单数）。
- 后果：
  1. 常量成为死代码；
  2. 记录的 `resource_type` 是路径形态（如 `authorization-codes` / `product-offerings` / `feature-templates` / `api-keys`），与任何规范不一致；
  3. 前端若按常量 `authorization-code` 筛选 → **查不到** `authorization-codes` 的记录（筛选静默落空）；
  4. 路径风格（连字符、复数、子路径）变化会产生碎片化的 `resource_type` 值，统计分布（`GetStatistics` 的 `ByResource`）也随之失真。

### P1

- **P1-1 资源 ID 识别仅 UUID 形态**：`isUUIDLike` 严格 36 位 hex。API Key 等非 UUID 资源 ID 不会被提取，`resource_id` 留空 → 无法按资源追溯。
- **P1-2 敏感信息未脱敏**：`Detail` 存明文 `reqBody`/`respBody`（截断 4096）。`create-user` / `reset-password` 的密码、`rotate-signing-key` 的密钥、`login` 的凭证等敏感字段可能落库，违反最小暴露原则。
- **P1-3 覆盖缺口**：`/v1/activate`（许可证激活，核心写操作）在 `public`/`api-key` 组（`router.go:104`），**未被审计**；`/v1/login` 失败尝试未审计（安全审计常需记录登录失败）。

### P2

- **P2-1 同步阻塞写入**：每个写请求在请求生命周期内同步 `cfg.DB.Create`，高并发增加响应延迟（失败不阻断，但成功路径同步）。
- **P2-2 事务不一致**：审计旁路单独 `cfg.DB.Create`，不进业务事务。业务事务回滚时审计已落库；审计写入失败仅 `c.Error` 无补偿/重试。

## 候选方案

### 方案 A — 最小改动修 bug（推荐基线）

保留中间件旁路架构，仅：

- 在 `resolveAuditResource` 建「路径段 → 规范常量」映射表（如 `authorization-codes → authorization-code`，与 `models.AuditResource*` 对齐），消除 P0 + 碎片，并让常量从死代码变为单一真源；
- 放宽 `resource_id` 识别（P1-1）：非路径关键字的语义段即视为 ID，或显式从 `gin.Params` 提取；
- `Detail` 序列化前做字段级脱敏（password / token / secret / key 等 → `***`）（P1-2）；
- 把 `/v1/activate` 纳入审计作用域（P1-3），以 API Key 对应的产品/客户端身份写入 `username`/`user_id`。

影响面小、零架构变更、直接修复数据一致性，且让既有常量发挥作用。

### 方案 B — 异步化写入

在 A 基础上，将同步 `cfg.DB.Create` 改为 channel + 单 worker 批量落库（或 `go` fire-and-forget + 连接池隔离）。

- 解决 P2-1，解耦主流程延迟；引入 goroutine 生命周期、panic 恢复、进程退出 flush 等复杂度。当前写 QPS 未到瓶颈，非紧急。

### 方案 C — 事务一致

审计写入纳入业务事务（同 `tx`）或独立审计库 + 失败重试队列。

- 解决 P2-2；但跨事务边界、重试语义、对现有旁路设计的改动最大。审计"最终一致"在旁路场景可接受，收益不足以抵消复杂度。

## 推荐

- **先落地方案 A**（P0 必须修，P1 合规/覆盖缺口应修），不引入架构风险；
- 方案 B 作为后续性能里程碑（当前非紧急）；
- 方案 C 暂不采纳。

## 验证

- P0：单测断言 `resolveAuditResource("/v1/authorization-codes/{uuid}")` 返回 `authorization-code`；grep 确认常量被消费。
- P1-1：非 UUID ID（如 `ak_ab12cd`）能被提取为 `resource_id`。
- P1-2：构造含 `password` 的请求体，断言 `Detail` 中该字段为 `***`。
- P1-3：`/v1/activate` 请求后审计表存在对应记录。
- 回归：后端 `go build ./...` + `go test ./...` 全绿。

## 风险与说明

- P0 修复会改变历史记录的 `resource_type` 值域；存量数据需一次性迁移（UPDATE 映射），否则新旧值并存。RFC 采纳后单列迁移步骤。
- 方案 A 的脱敏是字段名黑名单式，未来新增敏感字段需同步维护；更稳的是基于结构体 tag，但需改请求解析层，超出本 RFC 范围。
