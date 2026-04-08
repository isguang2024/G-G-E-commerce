# 接口自动入库机制

> 本文说明 `api_endpoints`、`api_endpoint_permission_bindings`、`permission_keys` 三张表的自动补齐逻辑，以及新接口上线时开发者需要做的操作。

---

## 整体流程

```
openapi.yaml
    │
    ▼  go run ./cmd/gen-permissions
openapi_seed.json  ←── 编译时 embed 进二进制
    │
    ▼  go run ./cmd/migrate
┌─────────────────────────────────────────┐
│  EnsureDefaultAPIEndpointCategories     │ 分类表
│  EnsureOpenAPIPermissionKeys            │ 权限键表
│  EnsureOpenAPIEndpoints                 │ 接口表
│  EnsureOpenAPIPermissionBindings        │ 绑定表
└─────────────────────────────────────────┘
```

三张表**全部幂等**：已存在的行不会被覆盖，只回填空字段。

---

## 新接口上线只需两条命令

```bash
# 1. 重新生成 seed（修改 openapi.yaml 之后）
go run ./cmd/gen-permissions

# 2. 运行迁移（部署或本地重置时）
go run ./cmd/migrate
```

不需要写任何额外代码。

---

## 接口分类（api_endpoint_categories）自动推断

分类按以下优先级解析，越靠前优先级越高：

| 优先级 | 来源 | 示例 |
|--------|------|------|
| 1 | openapi.yaml 的 `x-api-category` | `x-api-category: user` |
| 2 | operation 的第一个 `tags` 值 | `tags: [user]` | 
| 3 | `x-permission-key` 的第一段 | `user.list` → `user` |
| 4 | URL 的第一个有意义路径段 | `/api/v1/users` → `user` |
| 5 | 兜底 | `uncategorized` |

**推荐做法**：只要 `tags` 写正确，分类自动命中，无需额外配置。

---

## 接口 ID 生成规则

每条接口的 UUID 由以下公式确定性生成（UUIDv5）：

```
StableID("openapi-api-endpoint", "GET /api/v1/users")
```

同一接口在所有环境和数据库重置后 ID 始终一致，便于关联表引用。

---

## 幂等规则（不会覆盖手工改过的值）

`EnsureOpenAPIEndpoints` 发现已有行时，**只回填空字段**：

| 字段 | 行为 |
|------|------|
| `code` | 为空时补稳定 ID |
| `summary` | 为空时从 openapi.yaml `summary` 补 |
| `category_id` | 为 NULL 时按推断规则补 |
| 其他字段 | 不修改，保留运维在后台手工改的值 |

---

## 各 Ensure 函数调用顺序

```
EnsureDefaultAPIEndpointCategories   ← 必须最先，后续依赖 category_id
EnsureDefaultPermissionGroups
EnsureDefaultPermissionKeys
EnsureOpenAPIPermissionKeys          ← 补 permission_keys 表
EnsureOpenAPIEndpoints               ← 补 api_endpoints 表
EnsureOpenAPIPermissionBindings      ← 补绑定表
```

均在 `cmd/migrate/main.go` 的启动函数链中按顺序调用。

---

## 权限绑定（api_endpoint_permission_bindings）

绑定表的 UUID 同样确定性生成：

```
StableID("openapi-endpoint-binding", endpointCode+":"+permissionKey)
```

每个接口与其 `x-permission-key` 形成 1:1 绑定，`MatchMode = ANY`。

---

## 为什么不能完全省掉 gen-permissions 步骤

`openapi_seed.json` 通过 `//go:embed` 在**编译时**固化进二进制。运行时读的是快照，不是磁盘上的 yaml。因此：

- 改 openapi.yaml → 不跑 gen-permissions → migrate 看不到新接口
- 跑了 gen-permissions → 不重新编译/go run → embed 里还是旧内容

这是 Go embed 的机制约束，不是设计缺陷。

---

## 相关代码位置

| 功能 | 文件 |
|------|------|
| seed 生成器 | `backend/cmd/gen-permissions/main.go` |
| seed 加载 | `backend/internal/pkg/permissionseed/openapi_loader.go` |
| 接口 Ensure | `backend/internal/pkg/permissionseed/openapi_endpoints_ensure.go` |
| 权限键 Ensure | `backend/internal/pkg/permissionseed/openapi_ensure.go` |
| 分类默认值 | `backend/internal/pkg/permissionseed/seeds.go` |
| migrate 调用入口 | `backend/cmd/migrate/main.go`（`initOpenAPIEndpoints`） |
