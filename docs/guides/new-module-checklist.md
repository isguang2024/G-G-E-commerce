# 新增系统模块完整接入 Checklist

新增一个带管理页面的系统模块（如上传配置、字典管理）时，必须完成以下所有步骤。
跳步会导致权限不生效、菜单不显示、seed 数据不完整。

---

## 前置概念

| 概念 | 说明 |
|------|------|
| 权限键（PermissionKey） | 格式 `{模块}.{子模块}.{动作}`，运行时鉴权的最小单元 |
| 权限键映射（Mapping） | `permissionkey.go` 中 legacy key → 标准 key 的转换表，同时决定 `ResourceCode`（模块归属） |
| 模块组（ModuleGroup） | 权限键的分类容器，用于管理后台分组展示 |
| 功能包（FeaturePackage） | 菜单 + 权限键的打包单元，角色通过绑定功能包获得能力 |
| 菜单种子（MenuSeed） | 左侧菜单树的默认条目，`PermissionKey` 控制可见性 |

---

## Checklist

### 1. 数据库迁移

文件：`backend/internal/pkg/database/migrations/{序号}_{描述}.sql`

```sql
-- +goose Up
CREATE TABLE my_resources (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     VARCHAR(64) NOT NULL DEFAULT 'default',
    -- 业务字段 ...
    status        VARCHAR(20) NOT NULL DEFAULT 'ready',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);
CREATE UNIQUE INDEX uix_my_resources_tenant_key
    ON my_resources(tenant_id, resource_key) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS my_resources;
```

**约束**：
- 每张表必须有 `tenant_id`
- 唯一索引写 `(tenant_id, business_key) WHERE deleted_at IS NULL`
- 默认数据不写迁移，走 seed / ensure 幂等逻辑

### 2. Go Model

文件：`backend/internal/modules/system/{module}/models.go`（或 `models/` 下）

```go
type MyResource struct {
    ID       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    TenantID string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
    Status   string         `gorm:"type:varchar(20);not null;default:'ready'" json:"status"`
    // ...
}

func (MyResource) TableName() string    { return "my_resources" }
func (m MyResource) GetStatus() string  { return m.Status }  // 若需 status 过滤
```

### 3. OpenAPI Spec

文件：
- `backend/api/openapi/domains/{tag}/paths.yaml`
- `backend/api/openapi/domains/{tag}/schemas.yaml`
- `backend/api/openapi/openapi.root.yaml`（注册新 domain 路径引用）

每个 operation 必须声明四个扩展字段：

```yaml
x-permission-key: system.mymodule.manage
x-tenant-scoped: true
x-app-scope: none          # none | optional | required
x-access-mode: permission  # permission | authenticated | public
```

### 4. 权限键映射

文件：`backend/internal/pkg/permissionkey/permissionkey.go`

在 `mappings` 中添加条目，**ResourceCode 决定模块归属**：

```go
"system.mymodule:manage": {
    Key:          "system.mymodule.manage",
    ResourceCode: "mymodule",           // 必须匹配 ModuleGroup.Code
    ActionCode:   "manage",
    Name:         "我的模块管理",
    Description:  "允许管理我的模块配置",
    ContextType:  "personal",
},
```

映射规则：
- lookup key 格式是 `resourceCode:actionCode`（冒号分隔）
- `ResourceCode` 会被 `newPermissionKeySeed` 用作 `ModuleCode` 和 `ModuleGroupCode`
- 若不添加映射，`FromLegacy` 会 fallback 到 `resource + "." + action` 拼接，但 `ModuleCode` 会是完整的 resourceCode 字符串（如 `system.mymodule`），无法匹配到正确的模块组

### 5. 权限键种子

文件：`backend/internal/pkg/permissionseed/seeds.go` → `DefaultPermissionKeys()`

```go
newPermissionKeySeed("system.mymodule", "manage", "我的模块管理", "允许管理我的模块配置"),
newPermissionKeySeed("system.mymodule", "view",   "查看我的模块", "允许查看我的模块数据"),
```

`newPermissionKeySeed` 内部流程：
1. 调用 `permissionkey.FromLegacy(resourceCode, actionCode)` 查映射表
2. 取 `mapping.ResourceCode` 作为 `ModuleCode`（若映射存在）
3. `ModuleGroupCode = ModuleCode`，所以映射中的 `ResourceCode` 必须匹配模块组的 `Code`

### 6. 模块组（如需新增分组）

文件：`backend/internal/pkg/permissionseed/seeds.go` → `DefaultPermissionModuleGroups()`

```go
{
    GroupType:   "module",
    Code:        "mymodule",            // 权限键映射的 ResourceCode 必须匹配此值
    Name:        "我的模块",
    NameEn:      "My Module",
    Description: "我的模块相关权限",
    Status:      "normal",
    SortOrder:   240,                   // 递增，参考现有最大值
    IsBuiltin:   true,
},
```

### 7. 菜单种子

文件：`backend/internal/pkg/permissionseed/seeds.go` → `DefaultMenus()`

```go
{
    Name:          "MyModule",
    ParentName:    "SystemIntegration",   // 挂载到哪个父菜单
    Path:          "/system/my-module",
    Component:     "/system/my-module",
    Title:         "我的模块",
    SortOrder:     8,
    PermissionKey: "system.mymodule.manage",
    Meta: usermodel.MetaJSON{
        "roles":       []interface{}{"R_SUPER"},
        "permissions": []interface{}{"system.mymodule.manage"},
        "keepAlive":   true,
    },
},
```

**约束**：
- 菜单名在整棵树中唯一
- `PermissionKey` 控制菜单对角色的可见性
- 菜单直达的页面不需要额外 `PageSeed`（seeds.go 注释："页面种子只保留非菜单直达页"）

### 8. 功能包绑定

文件：`backend/internal/pkg/permissionseed/seeds.go` → `DefaultFeaturePackages()`

在目标功能包（通常是 `platform_admin.system_manage`）中添加：

```go
MenuNames:      []string{..., "MyModule"},
PermissionKeys: []string{..., "system.mymodule.manage", "system.mymodule.view"},
```

**约束**：
- 菜单名必须在 `MenuNames` 中才会对绑定该功能包的角色可见
- 权限键必须在 `PermissionKeys` 中才会通过运行时鉴权

### 9. 生成刷新

```bash
# 后端 OpenAPI 生成链
cd backend && ./update-openapi.bat
# 或 make api

# 前端类型生成
cd frontend && pnpm run gen:api
```

生成物需一并提交：
- `backend/api/gen/*`
- `backend/internal/pkg/permissionseed/openapi_seed.json`
- `frontend/src/api/v5/schema.d.ts`

### 10. Handler + Service 实现

文件：`backend/internal/api/handlers/{domain}.go`

实现 `gen.Handler` 接口中新 operation 对应的方法，业务逻辑放在 `internal/modules/system/{module}/` 下。

### 11. 前端页面

文件：`frontend/src/views/system/{module}/index.vue`

- 使用 ArtTable（全局注册）+ Element Plus
- API 调用使用生成的 `v5Client` + `unwrap()` helper
- Vue 3 Composition API + `defineOptions({ name: 'MyModule' })`

### 12. 验证

```bash
# 后端编译
cd backend && go build ./...

# 后端测试
cd backend && go test ./internal/modules/system/{module}/...

# 前端类型检查
cd frontend && pnpm exec vue-tsc --noEmit

# 运行 migrate 落库
cd backend && go run ./cmd/migrate
```

---

## 速查：关键文件清单

| 步骤 | 文件 |
|------|------|
| 迁移 | `backend/internal/pkg/database/migrations/{序号}_{描述}.sql` |
| Model | `backend/internal/modules/system/{module}/` |
| OpenAPI spec | `backend/api/openapi/domains/{tag}/paths.yaml` + `schemas.yaml` |
| 权限键映射 | `backend/internal/pkg/permissionkey/permissionkey.go` |
| 种子总入口 | `backend/internal/pkg/permissionseed/seeds.go` |
| Handler | `backend/internal/api/handlers/{domain}.go` |
| 前端页面 | `frontend/src/views/system/{module}/index.vue` |

---

## 常见错误

| 现象 | 原因 |
|------|------|
| 菜单不显示 | 菜单名未加入功能包 `MenuNames` |
| 接口返回 403 | 权限键未加入功能包 `PermissionKeys`，或用户角色未绑定该功能包 |
| 权限键在管理后台归类错误 | `permissionkey.go` 映射中 `ResourceCode` 不匹配 `ModuleGroup.Code` |
| 权限键在管理后台不显示 | 未在 `DefaultPermissionKeys()` 中添加种子，或未运行 `cmd/migrate` |
| 菜单位置不对 | `ParentName` 写错或 `SortOrder` 冲突 |
