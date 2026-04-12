# Backend 结构说明（V5）

本文档用于快速说明 `backend/` 目录下各层职责边界，并明确 V5 的 OpenAPI-first 工作方式。

## 1. 目录总览

```text
backend/
├─ api/                 # API 契约与生成代码（OpenAPI-first）
├─ cmd/                 # 可执行入口（server/migrate/诊断/修复/生成）
├─ internal/            # 业务实现（handler、service、repo、基础能力）
├─ configs/             # 配置模板（示例配置）
├─ update-openapi.bat   # Windows 下一键刷新 OpenAPI 生成链
├─ Makefile             # 后端生成与构建任务入口
└─ config.yaml          # 本地运行配置（环境相关）
```

## 2. 责任边界

### `api/`

- `api/openapi/` 是 API 契约唯一真相源（OpenAPI Source of Truth）。
- `api/openapi/dist/openapi.yaml` 为 bundle 产物，禁止手改。
- `api/gen/` 为 `ogen` 生成代码，禁止手改。
- 允许手工修改的只有 `api/openapi/` 源文件（`openapi.root.yaml`、`components/*`、`domains/*`）。

### `cmd/`

- 只放程序入口，不承载业务规则。
- 当前主要入口：
  - `cmd/server`：启动 HTTP 服务。
  - `cmd/migrate`：执行迁移、结构对齐与默认数据 ensure。
  - `cmd/gen-permissions`：基于 OpenAPI 派生权限种子与前端错误码生成物。
  - `cmd/init-admin`：初始化管理员账号。
  - `cmd/diagnose`：环境与数据库诊断。
  - `cmd/repair-workspaces`：历史工作空间数据一次性回填修复。

### `internal/`

- 放后端实现细节，不对外暴露。
- `internal/api/`：handler、middleware、router、DTO、错误映射。
- `internal/modules/`：领域模块（当前以 `system` 域为主）。
- `internal/config/`：配置加载与校验。
- `internal/pkg/`：通用基础能力（数据库、权限评估、权限种子、日志、JWT 等）。
- 权限判断必须走 `internal/pkg/permission/evaluator`，禁止在 handler/service 散写权限交集逻辑。
- 涉及多租户数据时，查询必须显式落实 `tenant_id` 过滤策略。

### `configs/`

- 放示例配置（当前 `config.example.yaml`）。
- 运行时读取由 `internal/config` 统一负责；本地可使用 `backend/config.yaml`。
- 禁止在业务代码中硬编码环境参数。

## 3. OpenAPI-first 约定（强制）

任何新增或修改 API，必须按以下顺序推进，不可跳步：

1. 修改 `backend/api/openapi/` 下的 spec 源文件。
2. 生成 bundle：刷新 `backend/api/openapi/dist/openapi.yaml`。
3. 运行 `ogen`：刷新 `backend/api/gen/`。
4. 运行权限生成：刷新 `internal/pkg/permissionseed/openapi_seed.json`（及相关派生产物）。
5. 前端执行 `pnpm run gen:api` 刷新 `frontend/src/api/v5/` 生成物。
6. 基于最新生成物修正后端 handler/service 与前端调用。
7. 执行后端与前端联编/测试校验。

## 4. 生成命令参考

- 后端 OpenAPI 生成链（推荐）：`make api`
- Windows 快捷方式：`update-openapi.bat`
- 仅权限派生：`go run ./cmd/gen-permissions`

说明：若本次变更涉及新表、字段或基线默认数据，除生成链外还需执行 `cmd/migrate` 完成结构与数据对齐。
