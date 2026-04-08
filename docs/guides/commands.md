# 常用命令速查

> Windows 环境，命令在 `backend/` 目录下执行（除非另有标注）。

---

## 启动服务

```bash
# 直接双击根目录脚本（推荐）
start-backend.bat          # 后端，占用 8080
start-frontend.bat         # 前端，占用 5173

# 手动启动后端
go run ./cmd/server
```

---

## 代码生成

### 全量生成（改了任何 spec 时）

```bash
# Windows（推荐）
update-openapi.bat

# Linux / Mac
make api
```

依次执行：bundle → lint → ogen → 权限种子 → 前端错误码 TS。

生成产物（均需提交）：

| 产物 | 路径 |
|------|------|
| bundled spec | `api/openapi/dist/openapi.yaml` |
| Go server 接口 | `api/gen/` |
| 权限种子 | `internal/pkg/permissionseed/openapi_seed.json` |
| 前端错误码 | `frontend/src/api/v5/error-codes.ts` |

### 分步执行（调试或局部刷新时）

```bash
make api-bundle            # 仅 bundle：openapi.root.yaml + domains/* → dist/openapi.yaml
make api-lint              # 仅 lint（依赖 bundle）
make api-gen               # 仅 ogen（依赖 bundle + lint）
make api-perms             # 仅权限种子 + 错误码（依赖 bundle）
make api-front             # 仅前端 TS 类型（依赖 bundle）

# 只改了错误码或权限注解（未改 schema/路径）时，直接跑：
go run ./cmd/gen-permissions
```

### Spec 文件位置

```
backend/api/openapi/
├── openapi.root.yaml      # 入口（只写 $ref，不写路径）
├── components/            # 共享 schema / 错误 / 认证
└── domains/               # 按 domain 拆分的路径定义
    ├── auth/paths.yaml
    ├── user/paths.yaml
    └── ...
```

详见 [`backend/api/openapi/README.md`](../../backend/api/openapi/README.md)。

---

## 数据库

```bash
# 运行所有待执行迁移（goose up）
go run ./cmd/migrate

# 完整重置并重新 seed（清库 + 建表 + 默认数据）
go run ./cmd/dbreset

# 初始化超管账号
go run ./cmd/init-admin

# 初始化演示数据
go run ./cmd/init-demo
```

详见 [database.md](database.md)。

---

## 调试与诊断

```bash
# 诊断命令（检查配置、DB 连接等）
go run ./cmd/diagnose

# 打印所有已注册路由及其权限 key
go run ./cmd/routecodegen
```

---

## 编译检查

```bash
# 编译后端（不启动）
go build ./...

# 运行单元测试
go test ./...

# 运行集成测试（需要 Docker 中的 Postgres）
go test -tags integration ./internal/api/handlers/...
```

---

## Git 提交注意

```bash
# Windows 上避免大规模 CRLF 噪音
git -c core.autocrlf=false commit -m "..."
```
