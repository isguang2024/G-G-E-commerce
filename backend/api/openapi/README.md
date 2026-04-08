# backend/api/openapi — OpenAPI Spec 目录

GGE 5.0 后端 API 契约的唯一事实源。所有 Go server 代码、TS 客户端、权限种子均由此派生，禁止反向手改。

---

## 目录结构

```
api/openapi/
├── openapi.root.yaml          # 入口文件（info / servers / security / tags + $ref）
│                              # 不在这里写路径或 schema，只写引用
├── components/
│   ├── errors.yaml            # Error schema + 共享 responses（BadRequest/Unauthorized/…）
│   ├── common.yaml            # 全部共享 schema（跨 domain 引用的放这里）
│   └── security.yaml          # securitySchemes（BearerAuth）
├── domains/                   # 按业务域拆分的路径定义
│   ├── auth/
│   │   ├── paths.yaml         # /auth/* 路径
│   │   └── schemas.yaml       # 该域独有的 schema（逐步从 common 迁入）
│   ├── user/
│   ├── role/
│   ├── workspace/             # workspace + collaboration_workspace 两个 tag
│   ├── menu/                  # menu + navigation
│   ├── permission/
│   ├── featurepackage/
│   ├── system/
│   ├── message/
│   ├── page/
│   ├── api_endpoint/
│   └── media/
├── dist/
│   └── openapi.yaml           # ⚠ 机器生成，勿手改，由 make api-bundle 写入
├── redocly.yaml               # lint 规则
└── embed.go                   # Go embed：将 dist/openapi.yaml 嵌入二进制
```

---

## 编辑规则

| 要改什么 | 改哪里 |
|----------|--------|
| 新增/修改一条路径 | `domains/{domain}/paths.yaml` |
| 新增/修改 schema | `components/common.yaml`（多 domain 共用）或 `domains/{domain}/schemas.yaml`（单 domain 专用）|
| 修改错误结构 | `components/errors.yaml` |
| 修改认证方案 | `components/security.yaml` |
| 修改 info / servers / tags | `openapi.root.yaml` |
| 添加错误码 | **后端** `internal/api/apperr/codes.go`（真源），再更新 `components/errors.yaml` 注释 |

**禁止直接编辑 `dist/openapi.yaml`**，每次 bundle 都会覆盖。

---

## 生成链路

```
openapi.root.yaml
  + domains/*/paths.yaml
  + components/*.yaml
        │
        ▼ make api-bundle（redocly bundle）
dist/openapi.yaml
        │
        ├─▶ make api-gen（ogen）→ api/gen/     （Go server 接口）
        ├─▶ make api-perms（gen-permissions）
        │       ├─▶ internal/pkg/permissionseed/openapi_seed.json
        │       └─▶ frontend/src/api/v5/error-codes.ts
        └─▶ make api-front（openapi-typescript）
                └─▶ frontend/src/api/v5/schema.d.ts
```

### 常用命令

```bash
# 全量（改了任何 spec 时）
make api                   # Linux/Mac
update-openapi.bat         # Windows

# 仅重新生成权限种子 + 前端错误码（未改 schema/路径时）
go run ./cmd/gen-permissions

# 仅重新 bundle（调试 spec 合法性时）
make api-bundle
make api-lint
```

---

## 错误码体系

业务码真源：`internal/api/apperr/codes.go`

| 段 | 含义 |
|----|------|
| `1xxxx` | 参数 / 请求错误 |
| `2xxxx` | 认证 / 授权错误 |
| `3xxxx` | 业务 / 资源错误 |
| `5xxxx` | 服务端错误 |

前端对应文件 `frontend/src/api/v5/error-codes.ts` 由 `cmd/gen-permissions` 自动派生，不要手改。

---

## 添加新接口（完整流程）

详见 [`docs/guides/add-endpoint.md`](../../docs/guides/add-endpoint.md)。

快速参考：

1. 在对应 `domains/{domain}/paths.yaml` 添加 operation（必须含四个 `x-` 扩展）
2. 运行 `make api`（或 `update-openapi.bat`）
3. 在 `internal/api/handlers/{domain}.go` 实现生成的接口方法

---

## Lint 规则（redocly.yaml）

| 规则 | 当前等级 | 说明 |
|------|----------|------|
| `operation-operationId` | error | 每个 operation 必须有 operationId |
| `operation-operationId-unique` | error | operationId 不能重复 |
| `operation-tag-defined` | warn | tag 须在顶层 tags[] 声明（历史债，逐步修复） |
| `operation-4xx-response` | warn | 每个 operation 须有 4xx 响应（历史债，Step 5 迁移时改为 error）|
| `security-defined` | warn | security scheme 须在 components 声明 |
| `no-unused-components` | warn | 未使用的 component 定义 |
