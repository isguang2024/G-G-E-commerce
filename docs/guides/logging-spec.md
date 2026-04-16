# 日志与可观测性规范（Logging Spec）

> MaBen 全栈日志系统的真源文档。覆盖后端结构化日志、请求链路 ID 透传、业务审计日志、
> 前端统一 logger 与批量上报、OpenAPI `/telemetry/logs` 摄取端点。
>
> **约束范围**：任何新 handler / service / Vue 页面必须按本规范落地。整改老代码时遵循
> 同一套命名与契约，不要另起炉灶。
>
> **当前性能画像（2026-04）**：单实例审计/遥测链路目标持续吞吐提升到 `5k~10k events/s`
>（依赖 batch flush、Info/Debug 采样、审计月分区与降级写盘策略共同生效）。
>
> **真源**：
> - 后端上下文 logger：`backend/internal/pkg/logger/`
> - Request ID 中间件：`backend/internal/api/middleware/request_id.go`
> - 业务审计服务：`backend/internal/modules/observability/audit/`
> - 前端日志摄取服务：`backend/internal/modules/observability/telemetry/`
> - 前端统一 logger：`frontend/src/utils/logger/index.ts`
> - OpenAPI 契约：`backend/api/openapi/domains/telemetry/{paths,schemas}.yaml`
> - 数据表：`backend/internal/pkg/database/migrations/00026_audit_logs.sql`
>   （同一迁移内建 `audit_logs` 与 `telemetry_logs`）

---

## 1. 整体数据流

```
┌──────────────┐   1. 生成 uuid v7 ─────► X-Request-Id ────► 2. 回显响应头
│  前端 SPA   │   ◄───── 4. 后端 zap 日志（含 request_id）────────────
│              │     5. logger.error(...) → 批量 sendBeacon
│              │     ───────────────────────────────────────────────►
└──────┬───────┘                                                    │
       │ HTTP                                                        ▼
┌──────▼───────────────┐    3a. middleware.RequestID             ┌──────────────────────┐
│  Gin router          │ ─► ctx 注入 request_id / actor /        │  zap 根 logger       │
│  RequestID → Logger  │    tenant / app / workspace             │  (控制台 + 文件)     │
│  → Recovery → ...    │                                          └──────────────────────┘
└──────┬───────────────┘                                                    ▲
       │                                                                    │
       ▼                                                                    │
┌──────────────┐   3b. ogen handler →                                       │
│ 业务 handler │   audit.Recorder.Record(ctx, Event{...})                   │
│ (domain svc) │   telemetry.Ingester.Ingest(ctx, entries)                  │
└──────┬───────┘                                                            │
       │ 异步 channel + worker                                              │
       ▼                                                                    │
┌──────────────────────┐     ┌──────────────────────┐                       │
│  audit_logs (jsonb)  │     │  telemetry_logs      │◄──────────────────────┘
│  append-only         │     │  append-only          │
└──────────────────────┘     └──────────────────────┘
```

关键不变量：
1. **只有一个 request_id**：前端生成（或后端生成后回显）→ 中间件写入 gin.Context + request.Context + 响应头
   → audit / telemetry / zap 都读同一个 key → DB 列同名 `request_id`。
2. **中间件顺序固定**：`RequestID → Logger → Recovery → AppContext → DynamicAppSecurity`。
   调整顺序等于破坏链路，PR 审查时必须显式解释。
3. **所有入库路径都是 append-only**：审计与遥测合规要求。TRUNCATE / 分区清理由运维负责，应用层
   不提供更新/删除入口。

---

## 2. 后端结构化日志（zap）

### 2.1 根 logger 初始化

`backend/cmd/server/main.go` 启动时调用 `logger.NewWithOptions(Options{...})`，随后
`logger.SetBase(zlog)` 把它挂为全局根。任何包都可以通过 `logger.Base()` /
`logger.With(ctx)` 派生子 logger，无需 DI 传递。

```go
zlog, err := logger.NewWithOptions(logger.Options{
    Level:  cfg.Log.Level,   // debug|info|warn|error
    Output: cfg.Log.Output,  // stdout | stderr | /path/to/file
    Format: cfg.Log.Format,  // json（默认）| console
    Sampling: &logger.Sampling{
        Initial:    cfg.Log.Sampling.Initial,
        Thereafter: cfg.Log.Sampling.Thereafter,
    },
})
if err != nil { log.Fatalf(...) }
defer zlog.Sync()
logger.SetBase(zlog)
```

生产默认 `json` 格式 + `ISO8601` 时间戳 + 小写 level。采样用 `zapcore.NewSamplerWithOptions`，
防止同一事件刷屏把日志系统打爆；只在 `Initial/Thereafter > 0` 时启用。

### 2.2 context key 与子 logger

`logger` 包用私有 `ctxKey struct{name string}` 作为 key 类型，**禁止裸字符串作为 context key**。
对外导出 Setter/Getter：

| Setter                    | Getter                         | 用途                        |
|---------------------------|--------------------------------|-----------------------------|
| `WithRequestID(ctx, id)`  | `RequestIDFromContext(ctx)`    | 请求链路 ID                 |
| `WithActor(ctx, id, typ)` | `ActorIDFromContext(ctx)` / `ActorTypeFromContext(ctx)` | 当前 actor 身份 |
| `WithTenant(ctx, id)`     | `TenantFromContext(ctx)`       | 租户，缺省 `"default"`      |
| `WithApp(ctx, key)`       | `AppFromContext(ctx)`          | 当前 app_key                |
| `WithWorkspace(ctx, id)`  | `WorkspaceFromContext(ctx)`    | 协作空间 ID                 |
| `WithClient(ctx, ip, ua)` | `ClientIPFromContext(ctx)` / `UserAgentFromContext(ctx)` | 客户端信息 |

`logger.With(ctx)` 读取上面全部 key，**非空才追加**。Handler / service / 后台 goroutine 里的
规范写法：

```go
func (s *appService) Create(ctx context.Context, req CreateAppRequest) error {
    log := logger.With(ctx)
    log.Info("app.create.start", zap.String("app_key", req.Key))
    if err := s.repo.Insert(ctx, row); err != nil {
        log.Error("app.create.db_failed", zap.Error(err))
        return err
    }
    log.Info("app.create.success")
    return nil
}
```

### 2.3 事件命名（dot-case）

zap 消息字段 **必须** 是稳定的 `dot-case` 事件名：`domain.entity.action[.outcome]`。
这和前端 `logger.info('dashboard.mounted', ...)` 保持同一种命名，方便后端按 `event` 字段
聚合同一业务线、跨层追踪。

禁止的写法：

- `log.Info("创建 app 成功")` —— 面向人的句子，无法聚合
- `log.Info("Create App Success")` —— 空格/驼峰，字段命名不稳定
- `log.Info(fmt.Sprintf("create app %s", key))` —— 拼接信息，把上下文挤进 message

正确：`log.Info("app.create.success", zap.String("app_key", key))`。

### 2.4 Request ID 中间件

`middleware.RequestID()` 是每个请求的第一个中间件：

- 读请求头 `X-Request-Id`（常量 `middleware.RequestIDHeader`）；合法（≤64 ASCII 可见字符）则透传；
- 否则生成 UUID v7（带时序），失败降级 v4；
- 同时写入 `gin.Context["request_id"]`、`request.Context`（`logger.WithRequestID`）、响应头；
- 后续所有 middleware / handler 从 ctx 读 request_id，不要再走 gin.Context 字符串 key。

前端拦截器在 response 到达时调用 `setLastRequestId(id)`，把响应头写入 `frontend/src/utils/logger`
的模块变量；后续 `logger.xxx()` 的 `request_id` 字段即来自此值。

---

## 3. 业务审计日志（audit）

### 3.1 调用模型

`audit.Recorder.Record(ctx, Event{...})` 是唯一入口。它：

- 从 ctx 读 request_id / actor / tenant / app / workspace / IP / UA，**无需在 Event 里重填**；
- 把 `Before/After/Metadata` JSON 序列化并递归脱敏（`DefaultRedactFields` + 运维侧扩展）；
- 入异步 channel，由 worker 池用独立 5s ctx 写 DB；channel 满则"drop-newest + Warn 日志"；
- **不返回业务错误**，业务侧可以安全忽略返回值；关闭时 `recorder.Shutdown(ctx)` drain channel。

```go
h.audit.Record(ctx, audit.Event{
    Action:       "system.app.create",     // 必填，dot-case
    ResourceType: "app",
    ResourceID:   created.Key,
    Outcome:      audit.OutcomeSuccess,    // 空时按 ErrorCode 推断
    HTTPStatus:   http.StatusCreated,
    Before:       nil,
    After:        created,                 // GORM 模型可直接传；脱敏自动处理
    Metadata: map[string]any{
        "source": "admin-panel",
    },
})
```

### 3.2 何时必须写审计

以下场景 **必须** 调用 `audit.Recorder.Record`：

| 类别         | 必须审计的动作                                                    |
|--------------|-------------------------------------------------------------------|
| 认证         | 登录成功 / 失败、登出、OAuth 首次绑定、密码修改                   |
| 授权         | 角色变更、权限授予/撤销、应用作用域变更                           |
| 租户级变更   | 应用创建/删除、注册入口（Register Entry）增删改、菜单空间增删改  |
| 敏感数据     | 登录模板编辑、回调密钥轮换、API Endpoint 同步/注册                |
| 批量操作     | 消息发送 / 重发、批量导入/导出                                    |
| 拒绝操作     | 权限校验失败 → `Outcome: audit.OutcomeDenied, ErrorCode: "..."`   |

只读查询**不写审计**（用普通 zap 日志即可），否则 audit_logs 会被查询流量撑爆。

### 3.3 Outcome / ErrorCode 推断

| 场景                 | Outcome      | ErrorCode                  |
|----------------------|--------------|----------------------------|
| 业务成功             | `success`    | `""`                       |
| 权限校验失败         | `denied`     | 对应 apperr code（如 `permission.denied`） |
| 业务异常 / 参数错误  | `error`      | 对应 apperr code（如 `param.invalid`）     |

Outcome 留空时，service 会按 ErrorCode 是否为空自动推断 `success`/`error`；`denied` 必须显式传。

### 3.4 脱敏规则

`audit.DefaultRedactFields` 包含 `password / token / secret / authorization / cookie / credit_card /
id_card / phone` 等高敏关键字。递归匹配 key 名（不区分大小写），命中后替换为 `[REDACTED]`。

扩展方式：配置文件 `audit.redact_fields: [...]`，运行时覆盖默认集合。前端 logger 使用同一套
`REDACT_KEYS`；telemetry ingest 入库前还会兜一遍。**两层脱敏都是硬约束**：前端是软承诺（可能被
绕过），后端是最终防线。

---

## 4. 前端日志（telemetry）

### 4.1 统一 logger 单例

业务侧通过 `import { logger } from '@/utils/logger'` 使用。API：

```ts
logger.info(event: string, context?: Record<string, unknown>): void
logger.warn(event: string, context?: Record<string, unknown>): void
logger.error(event: string, contextOrError?: Record<string, unknown> | Error): void
logger.debug(event: string, context?: Record<string, unknown>): void
logger.setUser(userId: string): void   // 登录/登出 hook
logger.setRoute(path: string): void    // vue-router afterEach hook
logger.flush(): void                    // 手动 flush（测试用）
```

事件名使用 `dot-case`（同后端）。第二参数是结构化字段 map，**不要拼字符串**。`logger.error`
允许直接传 Error 对象，内部会转成 `{err: Error}`。

### 4.2 生命周期挂点

- `frontend/src/router/guards/afterEach.ts` 里调用 `logger.setRoute(to.fullPath)`
- `frontend/src/domains/auth/store.ts` 里
  - `setUserInfo` 调用 `logger.setUser(userId)`
  - `clearSessionState` 调用 `logger.setUser('')`
- `frontend/src/utils/http/error.ts` 把所有 HTTP 错误通过 `logger.error('http.error', {...})` 上报
- `frontend/src/utils/sys/error-handle.ts` 处理 `vueErrorHandler` / `scriptErrorHandler` /
  `registerPromiseErrorHandler` / `registerResourceErrorHandler`，全部走 logger

### 4.3 上报契约（LogEntry）

字段命名 **snake_case**，直接对齐 OpenAPI `TelemetryLogEntry`（见
`backend/api/openapi/domains/telemetry/schemas.yaml`）。**禁止**在中间拦截器里被误转 camelCase。

```ts
interface LogEntry {
  level: 'debug' | 'info' | 'warn' | 'error'
  event: string                                      // dot-case
  timestamp: string                                  // ISO8601 UTC
  request_id?: string                                // 最近一次 X-Request-Id
  context?: Record<string, unknown>
  error?: { name?: string; message?: string; stack?: string }
  route?: string
  user_id?: string                                   // 未登录为空
  session_id: string                                 // tab 生命周期内唯一
  user_agent: string
  viewport: { w: number; h: number }
}
```

### 4.4 传输策略

- **批量**：`batchSize=20`，`flushIntervalMs=3000`，到条或到时触发 flush
- **优先 `sendBeacon`**：页面卸载也能送达；失败 fallback 到 `fetch(..., {keepalive: true})`
- **Unload hook**：`pagehide` + `visibilitychange(hidden)` 强制 flush；比 `beforeunload` 在移动端更可靠
- **最低上报级别**：`minReportLevel='info'`，`debug` 只打控制台（仅 DEV）
- **永远异步**：`logger.xxx()` 不返回 Promise，业务不感知传输失败

### 4.5 前端脱敏（REDACT_KEYS）

与后端 `DefaultRedactFields` 对齐。命中的 key（不区分大小写）替换为 `[REDACTED]`。
新增敏感字段时两端同时改：

- 前端：`frontend/src/utils/logger/index.ts` `REDACT_KEYS`
- 后端：`backend/internal/modules/observability/audit/redact.go` `DefaultRedactFields`

---

## 5. Ingest 端点 `/api/v1/telemetry/logs`

### 5.1 OpenAPI 契约

| 字段                    | 约束                                         |
|-------------------------|----------------------------------------------|
| `operationId`           | `ingestTelemetryLogs`                        |
| `x-access-mode`         | `public`（未登录/登录皆可；内部租户级限流兜底） |
| `x-tenant-scoped`       | `true`                                       |
| `x-app-scope`           | `none`                                       |
| Request body            | `TelemetryIngestRequest { entries: [1..100] }` |
| Response                | `TelemetryIngestResponse { accepted: int32, dropped: int32 }` |

**端点永远不返回 4xx 表示业务拒绝**。超限 batch 统一返回 `{accepted:0, dropped:N}`，
避免前端进入重试死循环。4xx 仅保留给协议错误（schema mismatch、超 `maxItems` 等）。

### 5.2 服务端限流（token bucket）

`telemetry.service` 维护两个 `sync.Map` 承载的 token bucket：

- **per-session**（key = `session_id`）：配置 `telemetry.session_rate_limit`（默认 60 条/秒）
- **per-ip**（key = `client_ip`）：配置 `telemetry.ip_rate_limit`（默认 600 条/秒）
- burst 容量 = rate × 3；空闲 bucket 由 `BucketIdleTTL`（默认 5 分钟）GC 清理

任一维度超限，整个 batch 全部 dropped，service 端写 `telemetry.rate_limited` warn 日志；
不调用 DB，不返回 4xx。

### 5.3 服务端入库流程

```
Handler.IngestTelemetryLogs
  ├── Entry[] = ogen struct → telemetry.Entry[]
  ├── sessionKey = entries[0].session_id
  ├── ipKey = logger.ClientIPFromContext(ctx)
  └── telemetry.Ingester.Ingest(ctx, entries, sessionKey, ipKey)
       ├── rate limit check
       ├── for each entry:
       │    ├── build TelemetryLog row（level 归并 / 长度截断 / 二次脱敏）
       │    └── 入 queue（满则 dropped++）
       └── return Result{Accepted, Dropped}
```

关键配置（`config.TelemetryConfig`）：

| 字段                | 默认值 | 说明                                             |
|---------------------|--------|--------------------------------------------------|
| `ingest_enabled`    | `true` | 总开关；关闭时 handler 注入 `telemetry.Noop{}`  |
| `max_batch_size`    | `100`  | OpenAPI 层硬限，和 schema `maxItems` 一致       |
| `session_rate_limit`| `60`   | per-session 每秒条数上限                         |
| `ip_rate_limit`     | `600`  | per-ip 每秒条数上限                              |
| `payload_max_bytes` | `8192` | 单条 entry 序列化后最大字节；超限尾部截断      |

### 5.4 Graceful shutdown

`main.go` 的 `defer telemetryIngester.Shutdown(shutdownCtx)` 会：

1. `close(stopCh)` 通知 GC goroutine 退出
2. `close(queue)` 让 workers 消费完剩余 row
3. 等待 `wg.Wait()` 或 5s 超时

超时返回 error（main 侧写 warn 日志），不阻塞进程退出。

---

## 6. 数据表结构

建表迁移：`00026_audit_logs.sql`（`audit_logs` + `telemetry_logs` 同一文件）。

### 6.1 `audit_logs`

| 列              | 类型           | 说明                            |
|-----------------|----------------|---------------------------------|
| `id`            | BIGSERIAL PK   | 自增                            |
| `ts`            | TIMESTAMPTZ    | 业务事件时间                    |
| `request_id`    | varchar(64)    | 请求链路 ID                     |
| `tenant_id`     | varchar(64)    | 默认 `default`                  |
| `actor_id/type` | varchar(64/32) | 当前用户 ID / anonymous/user/api_key/system |
| `app_key`       | varchar(64)    | 当前 app_key                    |
| `workspace_id`  | varchar(64)    | 协作空间                        |
| `action`        | varchar(128)   | dot-case 事件名                 |
| `resource_type/id` | varchar     | 受影响资源                      |
| `outcome`       | varchar(16)    | success/denied/error            |
| `error_code`    | varchar(32)    | 对齐 apperr code                |
| `http_status`   | integer        | HTTP 状态码                     |
| `ip/user_agent` | varchar/text   | 客户端信息                      |
| `before_json/after_json/metadata` | jsonb | 脱敏后 JSON              |
| `created_at`    | TIMESTAMPTZ    | 写入时间                        |

索引：`(tenant_id, ts)`、`(tenant_id, actor_id, ts)`、`(tenant_id, resource_type, resource_id, ts)`、
`(tenant_id, action, ts)`、`request_id`（partial）、`(tenant_id, outcome, ts) WHERE outcome <> 'success'`。

#### 6.1.1 月分区与归档（`audit_logs`）

从 `00028_audit_logs_monthly_partition.sql` 开始，`audit_logs` 调整为声明式分区主表：

- 分区键：`PARTITION BY RANGE (ts)`；
- 主键：`(id, ts)` 复合主键（满足 PG 分区唯一约束）；
- 保底分区：`audit_logs_default`（兜住越界历史数据）；
- 月分区：`audit_logs_YYYY_MM`；
- `cmd/migrate` 每次启动都会幂等创建「当前月 + 下月」两个分区，避免跨月首写失败。

归档建议（按月执行）：

```sql
-- 1) 从主表摘分区（不再接收新写入）
ALTER TABLE audit_logs DETACH PARTITION audit_logs_2025_12;

-- 2) 可选：先备份该分区
-- pg_dump -t audit_logs_2025_12 <db_name> > audit_logs_2025_12.sql

-- 3) 归档完成后删除老分区
DROP TABLE IF EXISTS audit_logs_2025_12;
```

> 生产建议保留最近 N 个月热分区（例如 6~12 个月），更久数据走对象存储或冷库。

#### 6.1.2 存量迁移说明（生产）

`00028` 迁移可以把旧 `audit_logs` 表改造成分区主表，但对已有大量写入流量的生产环境，
仍建议按下面步骤执行，避免长事务期间写入抖动：

1. 申请短暂停写窗口（至少覆盖 DDL + 数据搬迁时间）。
2. 迁移前先做 `audit_logs` 全量备份（`pg_dump -t audit_logs ...`）。
3. 在维护窗口执行 migrate，让 `00028` 完成结构改造与历史数据导入。
4. 验证分区路由是否正确（插入几条不同月份 `ts`，检查落表）。
5. 恢复写流量，观察 `audit_queue_depth`、`audit_events_dropped_total`、`audit_degraded` 指标。

对于超大表（亿级行），推荐离线 `dump/restore` 到新分区结构后切换，避免在线迁移时间过长。

### 6.2 `telemetry_logs`

| 列              | 类型           | 说明                            |
|-----------------|----------------|---------------------------------|
| `id`            | BIGSERIAL PK   | 自增                            |
| `ts`            | TIMESTAMPTZ    | 前端事件时间                    |
| `request_id`    | varchar(64)    | 最近一次 X-Request-Id           |
| `session_id`    | varchar(64)    | 前端 tab session                |
| `tenant_id`     | varchar(64)    | 默认 `default`                  |
| `actor_id`      | varchar(64)    | user_id / 空                    |
| `app_key`       | varchar(64)    | 当前 app_key                    |
| `level`         | varchar(16)    | debug/info/warn/error           |
| `event`         | varchar(128)   | dot-case                        |
| `message`       | text           | 前端 error.message（截断）      |
| `url`           | text           | 前端路由                        |
| `user_agent/ip` | text/varchar   | 上报侧识别                      |
| `release`       | varchar(64)    | 构建版本 / git sha              |
| `payload`       | jsonb          | {context, session, viewport, ua, error} |
| `created_at`    | TIMESTAMPTZ    | 写入时间                        |

索引：`(tenant_id, ts)`、`(tenant_id, level, ts)`、`session_id`（partial）、`request_id`（partial）。

---

## 7. 端到端排查路径

### 7.1 前端报错 → 后端源头

1. 前端控制台 / telemetry_logs 里拿到 `request_id`
2. 后端 zap 日志按 `request_id=<value>` grep —— 同一链路的 trace 都带
3. `audit_logs WHERE request_id = '...'` 看有没有业务事件写入
4. 如果是权限拒绝，`outcome='denied'` 会附带 `error_code`，直接对到 apperr 码表

### 7.2 某 event 异常飙升

- 前端日志：`telemetry_logs WHERE event = 'http.error' AND ts > now() - interval '1h'`
- 按 `level / session_id / ip` 聚合定位异常租户/会话
- 业务事件：`audit_logs WHERE action = 'system.app.create' AND outcome = 'error'`

### 7.3 观测信号缺失

- `request_id` 为空 → 检查是否绕过 `middleware.RequestID`（路由没用 SetupRouter？）
- `actor_id` 全空 → AuthMiddleware 没写 `logger.WithActor`
- 前端 logger 不上报 → DEV 环境 `minReportLevel`、`disableRemote` 配置；Network 面板看 `/telemetry/logs`

### 7.4 后台查询 UI（审计 / 遥测日志页）

`audit_logs` / `telemetry_logs` 是 append-only 存证表；为避免开发同学开 DB 客户端
直接 `SELECT`（容易漏 tenant、忘 LIMIT、查生产触发全表扫），配套上线了两个只读
后台页面 + 一个 trace 聚合端点。这里把接线方式、权限、索引命中条件都写清，便于
PR 审查和性能排障。

#### 7.4.1 端点一览

| OperationId                  | 路径                                       | 权限键                        | 类型              |
|------------------------------|--------------------------------------------|-------------------------------|-------------------|
| `listAuditLogs`              | `GET /observability/audit-logs`            | `observability.audit.read`    | 分页列表 + 多字段过滤 |
| `getAuditLog`                | `GET /observability/audit-logs/{id}`       | `observability.audit.read`    | 单行详情（带 before/after/metadata） |
| `listTelemetryLogs`          | `GET /observability/telemetry-logs`        | `observability.telemetry.read` | 分页列表 + 多字段过滤 |
| `getTelemetryLog`            | `GET /observability/telemetry-logs/{id}`   | `observability.telemetry.read` | 单行详情（带 payload） |
| `getObservabilityTrace`      | `GET /observability/trace/{request_id}`    | `observability.audit.read`    | 聚合：同 request_id 下的 audit + telemetry，按 `ts asc` |

所有端点 `x-tenant-scoped: true`、`x-app-scope: optional`；handler 都硬加
`WHERE tenant_id = ?`（`logger.TenantFromContext`）。写入链路不受此影响——recorder /
ingester 的异步通道仍是唯一的写入者。

> 注意：`getObservabilityTrace` 故意只要 `audit.read`。telemetry 行作为审计现场的
> 从属信息一起返回，不要求调用方二次授权；这是产品决策，不是遗漏。

#### 7.4.2 前端页面

| 页面                       | 路由 name            | 文件                                        | 备注                         |
|----------------------------|----------------------|---------------------------------------------|------------------------------|
| 审计日志                   | `SystemAuditLog`     | `frontend/src/views/system/audit-log/index.vue`     | 列表 + 详情抽屉              |
| 前端遥测日志               | `SystemTelemetryLog` | `frontend/src/views/system/telemetry-log/index.vue` | 列表 + 详情抽屉              |
| Trace 抽屉（共享）         | 组件，无独立路由     | `frontend/src/components/Observability/TraceDrawer.vue` | 两页的 request_id 点击后弹出 |
| JSON 查看器（共享）        | 组件，无独立路由     | `frontend/src/components/Observability/JsonViewer.vue`  | before/after/metadata/payload |
| 日期 shortcut（共享）      | 工具文件             | `frontend/src/views/system/_shared/observability-shortcuts.ts` | 最近 1h/24h/7d/30d/今天/昨天 |

菜单挂载走 `router/menu` 里的 `system` 分组；菜单项在目录树里渲染时按其各自路由的
`meta.roles` / `meta.permissions` 过滤。没有 `observability.audit.read` 的账号根本
看不到入口。

#### 7.4.3 典型查询场景

1. **按 Actor 拉一个人的操作痕迹**
   - 审计页 `Actor` 框填 `<user_id>`，时间区间选「最近 7 天」，`Outcome` 置空
   - 对应请求：`GET /observability/audit-logs?actor_id=...&from=...&to=...`
   - 命中索引：`idx_audit_logs_tenant_actor_ts`（tenant_id, actor_id, ts DESC）

2. **定位权限拒绝原因**
   - `Outcome = denied`，`Action` 填业务事件（如 `system.user.update`）
   - 结果里 `error_code` 对应 apperr 码；拿 `request_id` 点击跳 trace 看前端侧
     同一请求的 telemetry 行（是否 toast 过、URL 是什么）

3. **排查前端报错爆发**
   - 遥测页 `Level = error`，`Event` 填 `http.error`，时间区间选「最近 1 小时」
   - 请求：`GET /observability/telemetry-logs?level=error&event=http.error&from=...`
   - 命中索引：`idx_telemetry_logs_tenant_level_ts`（tenant_id, level, ts DESC）
   - 结果里 `request_id` 点链接 → trace 抽屉 → 如果后端有对应 audit 行，
     说明请求真的落后端了；反之可能前端侧自爆（如 render 阶段）

4. **从某条日志跳整条链路**
   - 任一页详情抽屉里，`Request ID` 行已改成可点击蓝字链接
   - 同列表 `Request ID` 列也是 `ElButton link`，点击后 `ev.stopPropagation()`
     阻止行点击
   - 抽屉打开 → 调 `/observability/trace/{request_id}` → `ElTimeline` 合并两源
     按 `ts asc` 渲染（`TraceDrawer.vue`）

#### 7.4.4 过滤字段与索引匹配

`audit_logs`：

| 过滤字段组合                         | 命中索引                                | 典型用例           |
|--------------------------------------|-----------------------------------------|--------------------|
| `tenant_id + ts DESC`（默认排序）    | `idx_audit_logs_tenant_ts`              | 最近一屏           |
| `tenant_id + actor_id + ts DESC`     | `idx_audit_logs_tenant_actor_ts`        | 按人拉痕迹         |
| `tenant_id + action + ts DESC`       | `idx_audit_logs_tenant_action_ts`       | 某业务事件复核     |
| `tenant_id + request_id`             | `idx_audit_logs_tenant_request_id`      | trace 聚合         |
| `tenant_id + resource_type + resource_id` | `idx_audit_logs_tenant_resource`    | 某资源历史         |

`telemetry_logs`：

| 过滤字段组合                         | 命中索引                                   |
|--------------------------------------|--------------------------------------------|
| `tenant_id + ts DESC`                | `idx_telemetry_logs_tenant_ts`             |
| `tenant_id + level + ts DESC`        | `idx_telemetry_logs_tenant_level_ts`       |
| `tenant_id + session_id + ts`        | `idx_telemetry_logs_tenant_session_ts`     |
| `tenant_id + request_id`             | `idx_telemetry_logs_tenant_request_id`     |
| `tenant_id + event + ts DESC`        | 00027 迁移后：`idx_telemetry_logs_tenant_event_ts` |
| `tenant_id + actor_id + ts DESC`     | 00027 迁移后：`idx_telemetry_logs_tenant_actor_ts` |

> 约定：handler 侧永远 `tenant_id` 打头，页面上即便没勾选过滤字段也会带时间段默认值
> 兜底（前端 shortcuts 给到了「最近 24 小时」一键）。不要写没有 `tenant_id` 的
> 查询——会走顺序扫，生产表几千万行直接把 DB 打爆。

#### 7.4.5 只读语义与数据边界

- 这些 handler **不**写 audit（否则查询流量会反向炸 audit_logs）；也不经过 recorder
- 单条 size 上限：`observabilityListMaxSize = 200`。`trace` 端点也复用这个上限，
  防止一个病态 request_id 挂几千行拖爆内存
- 未命中的 id / request_id：
  - `get*` 返回 `404`
  - `getObservabilityTrace` 返回 `200` + 空数组（聚合视图不 404，前端渲染稳定）
- `audit_logs.before/after/metadata` 与 `telemetry_logs.payload` 都是 `jsonb`，
  handler 用 `jsonBytesToRawMap` 转换成 ogen schema 要求的 `map[string, jx.Raw]`；
  空/非对象值落回 `null`，前端 `JsonViewer` 对 `null` 走空态渲染

#### 7.4.6 新增查询能力时的动作清单

1. 改 `backend/api/openapi/domains/observability/paths.yaml` + `schemas.yaml`
   - 加 `x-permission-key`、`x-tenant-scoped: true`、`x-app-scope: optional`、
     `x-access-mode: permission`
2. 根 spec `openapi.root.yaml` 注册 path ref
3. 跑 `make api`（或 `update-openapi.bat`）重生 ogen + permission seed
   —— 这一步同时把新 operation 写入 `openapi_seed.json`，后端启动时
   `mountOpenAPIBridgeRoutes` 会按 `x-access-mode` 自动把路由挂到 Gin，
   不需要再改 `router.go`
4. `internal/api/handlers/observability.go` 实现 handler：
   - 首行判 `userIDFromContext(ctx)`；未登录返 `*Unauthorized`
   - `tenant_id` 硬过滤；时间 / 分页参数走 `observabilityPagination`
   - 复用 `auditLogItemFromModel` / `telemetryLogRecordFromModel` 映射
5. 跑一下 `go test ./internal/api/router -count=1`：该对账测试验证 seed 里
   每条 op 都能在 Gin trie 里被找到、bridge 对齐、无 radix tree 冲突
6. 补 `integration_test.go` smoke（`go test -tags integration ./internal/api/handlers/...`）
7. 前端 `frontend/src/domains/governance/api/observability.ts` 封装 client 函数

### 7.5 观测指标 runbook（`/observability/metrics`）

运行时指标端点暴露 recorder / ingester 的**进程内**计数，不查 DB（响应 <1ms），
供健康检查、仪表盘、Prometheus agent 定期 scrape。权限复用
`observability.audit.read`；`x-app-scope: none`（基础设施指标，和租户无关）。

响应形状固定：`{ audit: ServiceStats, telemetry: ServiceStats, collected_at }`，
其中 `ServiceStats` 四字段含义：

| 字段              | 含义                                                     | 典型值（稳态）  |
|-------------------|----------------------------------------------------------|-----------------|
| `queue_depth`     | 当前 buffered channel 里待消费事件数；同步模式恒为 0       | 0 ~ 几十        |
| `queue_cap`       | channel 容量；`0` 表示同步模式或 Noop（未启用异步缓冲）    | audit/telemetry 按 config 配，默认 1024 |
| `accepted_total`  | 进程启动以来成功入队 / 同步写入的累计（**单调递增**）      | 随流量线性上涨 |
| `dropped_total`   | 进程启动以来因 channel 满 / rate-limit 被丢弃的累计         | 应长期贴近 0    |

> 字段单调递增；进程重启后归 0。多副本部署必须在 scrape 层按 `pod/replica` 打标
> 再求 `sum(rate(...))`，**不要**把不同副本的 `accepted_total` 直接加减当成一个
> 全局 counter——每个进程是独立 series。

#### 采集周期建议

| 场景                        | 建议周期   | 备注                                 |
|-----------------------------|-----------|--------------------------------------|
| Prometheus agent scrape     | `30s`     | counter 类型，grafana `rate(... [5m])` |
| 运维仪表盘手动刷新          | `60s`     | 肉眼读够用，避免压爆无价值查询         |
| CI / 冒烟测试               | 一次即可  | 断言形状 + Noop 零值                  |

#### 告警阈值建议

1. **drop 飙升**：`increase(audit_dropped_total[15m]) > 0`
   或 `increase(telemetry_dropped_total[15m]) > 0` 任一触发即告警。
   稳态下 `dropped_total` 不该增长，一旦出现 delta > 0 表示 channel 真的被灌满
   过，无论数量多少都值得看一眼。
2. **队列持续深水位**：`avg_over_time(audit_queue_depth[5m]) / audit_queue_cap > 0.8`
   持续 5 分钟告警。worker 追不上 producer，下一步就是 drop；提前扩 worker
   或降 action 频率。
3. （可选）**accepted 完全停止**：`rate(audit_accepted_total[10m]) == 0` 且租户流量
   未归零 → recorder 可能挂了或 AuthMiddleware 绕过了 Record 分支。

#### 排障 runbook（drop 飙升时）

1. **确认到底是谁在 drop**：先分 audit vs telemetry。telemetry drop 一般是前端
   刷屏/爬虫/rate-limit 起作用；audit drop 是 DB 真实卡顿的信号，优先级更高。
2. **看 zap warn 日志**：service 每次 drop 都会打 `audit.queue_full_drop` /
   `telemetry.queue_full_drop` 带 `dropped_total`。按 pod grep，定位是单副本热
   点还是全集群共性；日志里的 `action` / `event` 字段指向具体热源。
3. **看 `accepted_total` 斜率**：drop 期间 accepted 斜率是否骤增？骤增 → 业务侧
   有放量调用（看 audit 的 action 分布）；无变化 → DB 写入变慢（看
   `pg_stat_activity` 或慢查询日志，`audit_logs` 的写 latency 是否抬高）。
4. **临时止血**：
   - audit：调大 `AuditConfig.QueueSize` + `Workers`（重启生效，热更未支持）；
   - telemetry：调大 `TelemetryConfig.RateLimitCapacity`；或前端临时关某个高频
     `event`（修 `minReportLevel` / 打 `disableRemote`）。
5. **根因修复后验证**：等 15min，`increase(dropped_total[15m])` 归零即收敛。

> 注意：`/metrics` 本身是 Observer，不会反向写 audit（否则会和被观测系统耦合）。
> 不要在 recorder 内部调 `/metrics`——走进程变量直接拿 `Stats()` 就够。

#### 7.5.1 Prometheus scrape 接入（`/observability/metrics/prometheus`）

`GET /observability/metrics` 返回 JSON，面向前端 dashboard 与 `/healthz` 增强；
`GET /observability/metrics/prometheus` 返回 openmetrics-text v1.0.0，面向
Prometheus / Alertmanager 等通用监控系统 **直接 scrape**。两者底层共享
`audit.Recorder.Stats()` + `telemetry.Ingester.Stats()`，差异仅在呈现格式。
text 端点现在同时导出 audit 与 telemetry 两组指标，便于外部监控统一对比
后端审计写入链路与前端日志摄取链路。

字段映射（严格按 openmetrics `# HELP` / `# TYPE` 输出）：

| 指标名                         | 类型      | 含义                                               |
|-------------------------------|-----------|----------------------------------------------------|
| `audit_queue_depth`           | `gauge`   | 瞬时队列深度 `len(chan)`                            |
| `audit_queue_capacity`        | `gauge`   | 队列容量 `cap(chan)`，`0` 表示同步 / Noop             |
| `audit_events_accepted_total` | `counter` | 成功落库累计（进程重启归零，单调递增）              |
| `audit_events_dropped_total`  | `counter` | 丢弃累计（drop-newest；稳态应贴近 0）                |
| `telemetry_queue_depth`       | `gauge`   | telemetry ingest 队列瞬时深度 `len(chan)`            |
| `telemetry_queue_capacity`    | `gauge`   | telemetry ingest 队列容量 `cap(chan)`，`0` 表示同步 / Noop |
| `telemetry_events_accepted_total` | `counter` | telemetry 接收累计（进程重启归零，单调递增）      |
| `telemetry_events_dropped_total`  | `counter` | telemetry 丢弃累计（限流或队列满）                 |

示例响应体（Noop 或空闲进程）：

```text
# HELP audit_queue_depth audit recorder queue length (len(chan))
# TYPE audit_queue_depth gauge
audit_queue_depth 0
# HELP audit_queue_capacity audit recorder queue capacity (cap(chan))
# TYPE audit_queue_capacity gauge
audit_queue_capacity 0
# HELP audit_events_accepted_total cumulative audit events persisted since process start
# TYPE audit_events_accepted_total counter
audit_events_accepted_total 0
# HELP audit_events_dropped_total cumulative audit events dropped (queue full, drop-newest)
# TYPE audit_events_dropped_total counter
audit_events_dropped_total 0
# HELP telemetry_queue_depth telemetry ingester queue length (len(chan))
# TYPE telemetry_queue_depth gauge
telemetry_queue_depth 0
# HELP telemetry_queue_capacity telemetry ingester queue capacity (cap(chan))
# TYPE telemetry_queue_capacity gauge
telemetry_queue_capacity 0
# HELP telemetry_events_accepted_total cumulative telemetry events accepted since process start
# TYPE telemetry_events_accepted_total counter
telemetry_events_accepted_total 0
# HELP telemetry_events_dropped_total cumulative telemetry events dropped (rate limit or queue full)
# TYPE telemetry_events_dropped_total counter
telemetry_events_dropped_total 0
```

**`prometheus.yml` 配置样例**（与 `observability.audit.read` 一致的 bearer 授权）：

```yaml
scrape_configs:
  - job_name: gge-backend-observability
    scheme: https                 # 生产走 https;dev 直接 http 也可以
    metrics_path: /api/v1/observability/metrics/prometheus
    scrape_interval: 30s          # 与 runbook 建议的采集周期保持一致
    scrape_timeout: 10s
    authorization:
      type: Bearer
      # credentials_file 指向 K8s secret 挂载的 token 文件(也可走 credentials: <literal>)。
      # token 必须绑定 observability.audit.read 权限;建议单独发一个只读服务账号。
      credentials_file: /etc/prometheus/secrets/gge-metrics-token
    static_configs:
      - targets:
          - backend-0.gge.internal:443
          - backend-1.gge.internal:443
        labels:
          service: gge-backend
          tier: observability
    # 多副本时按 __address__ 或 pod relabel,避免把不同进程的 counter 合并。
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
```

配套告警（Prometheus rule 片段）：

```yaml
groups:
  - name: gge-observability
    rules:
      - alert: MaBenAuditDroppingEvents
        expr: increase(audit_events_dropped_total[15m]) > 0
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "audit recorder dropping events ({{ $labels.instance }})"
          runbook: "docs/guides/logging-spec.md#7-5"
      - alert: MaBenAuditQueueDeepWater
        # 乘法而不是除法避免 queue_capacity=0 时 DIV/0
        expr: avg_over_time(audit_queue_depth[5m]) > 0.8 * audit_queue_capacity
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "audit queue stays > 80% for 5m ({{ $labels.instance }})"
```

> 注意：本端点与 JSON `/observability/metrics` 共享 `observability.audit.read`
> 权限，不单独设计 prom 专属 key。Noop 模式下抓取返回 200 + 全零样本（不是
> 空响应或 404），便于在 Alertmanager 侧用同一套规则区分「暂未启用」与
> 「启用后出异常」两种状态。

---

## 8. 自查清单（新 handler / 整改 PR 提交前）

| # | 检查项                                                                                       | 证据             |
|---|----------------------------------------------------------------------------------------------|------------------|
| 1 | 新 handler 在成功/失败路径都调用了 `audit.Recorder.Record`（只读查询除外）                    | 代码审查          |
| 2 | zap 日志事件名用 `dot-case`，没有拼接字符串 message                                            | grep PR diff     |
| 3 | 所有 context 相关字段通过 `logger.With(ctx)` 自动带入，未在 Event / zap.Field 里手动冗余重填 | grep             |
| 4 | 前端错误全部走 `logger.error`，未残留 `console.error`                                         | grep `console\.` |
| 5 | 新敏感字段同时加到后端 `DefaultRedactFields` 与前端 `REDACT_KEYS`                             | diff             |
| 6 | middleware 顺序未调整（RequestID → Logger → Recovery → AppContext → DynamicAppSecurity）     | router diff      |
| 7 | OpenAPI spec 变更已通过 `make api`，audit/telemetry 相关 contract 无漂移                       | `git status`     |

---

## 9. 相关文档

- `docs/guides/frontend-observability-spec.md` — 前端可观测性（data-testid / form error / error code）
- `docs/API_OPENAPI_FIXED_FLOW.md` — 新接口 spec 变更的固定步骤
- `backend/api/openapi/README.md` — OpenAPI 多文件结构
- `backend/internal/api/apperr/codes.go` — 错误码唯一真源
- `frontend/src/utils/logger/index.ts` — 前端 logger 单例与脱敏
- `frontend/src/utils/http/error.ts` — HTTP 错误统一上报与 toast 策略

