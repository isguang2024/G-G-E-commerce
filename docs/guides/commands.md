# 常用命令速查

> Windows 环境，命令在 `backend/` 目录下执行（除非另有标注）。
> `make` 不可用，所有命令直接用 `go run`。

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

### 1. 重新生成 OpenAPI 服务端代码（ogen）

每次修改 `api/openapi/openapi.yaml` 后执行：

```bash
go run github.com/ogen-go/ogen/cmd/ogen@latest \
  --target api/gen --package gen --clean api/openapi/openapi.yaml
```

生成产物：`api/gen/`（需提交）。

### 2. 刷新权限 seed

```bash
go run ./cmd/gen-permissions
```

生成产物：`internal/pkg/permissionseed/openapi_seed.json`（需提交）。

每次改了 openapi.yaml 都要跑，migrate 和 server 启动时读取此 JSON。

### 3. 重新生成前端 TypeScript 客户端

在 `frontend/` 目录执行：

```bash
npx openapi-typescript@7 ../backend/api/openapi/openapi.yaml -o src/api/v5/schema.d.ts
```

> 不要用 `pnpm run gen:api`，Windows worktree 里没有 node_modules。

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
