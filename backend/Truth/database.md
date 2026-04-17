# 数据库操作指南

---

## 常用操作速查

```bash
# 运行所有待执行迁移（goose up）
go run ./cmd/migrate

# 完整重置（清库 + 建表 + 默认数据）
go run ./cmd/dbreset

# 初始化超管账号
go run ./cmd/init-admin

# 初始化演示数据
go run ./cmd/init-demo
```

---

## migrate 做了什么

`go run ./cmd/migrate` 按顺序执行：

1. **goose up** — 运行 `internal/pkg/database/migrations/` 下所有待执行的 .sql 文件
2. **schema 修正** — `finalizeAPIEndpointSchema` 等兼容性修正
3. **分类 seed** — `EnsureDefaultAPIEndpointCategories`
4. **权限组 seed** — `EnsureDefaultPermissionGroups`
5. **权限键 seed** — `EnsureDefaultPermissionKeys` + `EnsureOpenAPIPermissionKeys`
6. **功能包 seed** — `EnsureDefaultFeaturePackages` + `EnsureDefaultFeaturePackageBundles`
7. **角色绑定 seed** — `EnsureDefaultRoleFeaturePackages`
8. **接口自动注册** — `EnsureOpenAPIEndpoints` + `EnsureOpenAPIPermissionBindings`
9. **菜单 seed** — 同步默认菜单树
10. **消息模板 seed** — 初始化默认消息模板

所有 seed 步骤均**幂等**，重复运行安全。

---

## 直连 Docker 数据库（授权操作）

```bash
# 进入 postgres 容器
docker exec -it gge-postgres psql -U postgres -d gge

# 常用查询
\dt                          -- 列出所有表
SELECT * FROM api_endpoints LIMIT 10;
SELECT * FROM permission_keys WHERE permission_key LIKE 'user.%';
SELECT * FROM api_endpoint_categories ORDER BY sort_order;
```

---

## 迁移文件规范

迁移文件位于：`backend/internal/pkg/database/migrations/`

命名格式：`{序号}_{描述}.sql`，例如 `00003_add_widget_table.sql`

```sql
-- +goose Up
CREATE TABLE widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(100) NOT NULL DEFAULT 'default',
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX uix_widgets_tenant_name ON widgets(tenant_id, name) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS widgets;
```

**约束：**
- 每张业务表必须有 `tenant_id`，唯一性约束写成 `(tenant_id, business_key)`
- 默认数据**不写进迁移文件**，走 seed / ensure 幂等逻辑
- 临时修复型迁移在目标状态达成后需删除，不长期保留

---

## seed 与迁移的分工

| 类型 | 写在哪里 | 特点 |
|------|----------|------|
| 表结构 | goose 迁移文件（.sql） | 一次性，有 Down |
| 默认数据（分类、权限键等） | `permissionseed/seeds.go` + Ensure 函数 | 幂等，可重复跑 |
| OpenAPI 派生数据（接口、权限键） | `openapi_seed.json` + Ensure 函数 | 幂等，由 gen-permissions 生成 |

---

## 排查数据库问题

```bash
# 查看当前 goose 版本
docker exec -it gge-postgres psql -U postgres -d gge \
  -c "SELECT * FROM goose_db_version ORDER BY id DESC LIMIT 5;"

# 检查接口是否已入库
docker exec -it gge-postgres psql -U postgres -d gge \
  -c "SELECT method, path, summary, source FROM api_endpoints ORDER BY path;"

# 检查权限键是否存在
docker exec -it gge-postgres psql -U postgres -d gge \
  -c "SELECT permission_key, name FROM permission_keys WHERE permission_key = 'user.list';"
```
