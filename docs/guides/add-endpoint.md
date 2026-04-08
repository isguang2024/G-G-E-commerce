# 新增后端接口

> 规则：**OpenAPI spec 是唯一事实源**。所有步骤从改 yaml 开始，禁止在 Go 里先写 handler 再补文档。

---

## 核心流程（3 步）

### 步骤一：在 spec 中声明接口

找到对应 domain，编辑 `backend/api/openapi/domains/{domain}/paths.yaml`：

```yaml
paths:
  /widgets/{id}:
    get:
      operationId: getWidget          # 唯一，格式：{domain}{Action} 驼峰
      summary: 获取组件详情
      tags: [widget]                  # 第一个 tag 用于自动分配接口分类
      x-permission-key: widget.read   # 权限键；public/authenticated 模式时可省
      x-tenant-scoped: true           # v5 一律 true
      x-app-scope: optional           # required | optional | none
      x-access-mode: permission       # permission | authenticated | public
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '../../components/common.yaml#/components/schemas/Widget'
        '401':
          $ref: '../../components/errors.yaml#/components/responses/Unauthorized'
        '404':
          $ref: '../../components/errors.yaml#/components/responses/NotFound'
```

若需要新 schema，添加到 `components/common.yaml`（多 domain 共用）或 `domains/{domain}/schemas.yaml`（单 domain 专用）。

**`x-access-mode` 说明：**

| 值 | 含义 |
|----|------|
| `permission` | 需要 JWT + 权限键（默认） |
| `authenticated` | 需要 JWT，不校验权限键 |
| `public` | 完全公开，无需登录 |

---

### 步骤二：一键生成所有派生产物

```bash
# Windows
update-openapi.bat

# Linux / Mac
cd backend && make api
```

这一条命令依次完成：

| 子步骤 | 产物 | 是否需提交 |
|--------|------|-----------|
| `redocly bundle` | `dist/openapi.yaml` | ✓ |
| `redocly lint` | 校验，无产物 | — |
| `ogen` | `api/gen/` | ✓ |
| `gen-permissions` | `openapi_seed.json` + `frontend/src/api/v5/error-codes.ts` | ✓ |

> 如果只改了权限注解（未改 schema/路径），可以只跑：
> ```bash
> go run ./cmd/gen-permissions
> ```

---

### 步骤三：实现 handler

在 `internal/api/handlers/{domain}.go` 实现 ogen 生成的接口方法：

```go
func (h *APIHandler) GetWidget(ctx context.Context, params gen.GetWidgetParams) (gen.GetWidgetRes, error) {
    widget, err := h.widgetSvc.Get(ctx, params.ID)
    if err != nil {
        return nil, err   // 由 apperr.ErrorHandler 统一翻译成 JSON 错误响应
    }
    return &gen.Widget{ID: widget.ID, Name: widget.Name}, nil
}
```

**错误处理规则：**
- 直接 `return nil, err`，不在 handler 里构造 `gen.XxxUnauthorized{...}` 等
- `internal/api/apperr/mapper.go` 是唯一翻译表；新增错误映射只改 mapper，不改 handler

---

## 后续步骤

### 权限键自动入库

运行 `go run ./cmd/migrate`，`EnsureOpenAPIPermissionKeys` 自动将 `x-permission-key` 写入 `permission_keys` 表，`EnsureOpenAPIEndpoints` 写入 `api_endpoints` 表。

详见 [api-auto-registration.md](api-auto-registration.md)。

### 添加集成测试

```go
//go:build integration

func TestGetWidget(t *testing.T) {
    token := integToken(t, "admin@example.com")
    res := integDo(t, token, "GET", "/api/v1/widgets/"+testWidgetID, nil)
    assert.Equal(t, 200, res.StatusCode)
}
```

```bash
go test -tags integration ./internal/api/handlers/...
```

### 更新前端客户端

```bash
# 生成 TypeScript 类型（已由 make api 包含，单独执行时：）
cd frontend && npm run gen:api
```

---

## Spec 文件位置速查

```
backend/api/openapi/
├── openapi.root.yaml          # 入口（不直接写路径）
├── components/
│   ├── errors.yaml            # Error schema + 共享 responses
│   ├── common.yaml            # 跨 domain 的 schema
│   └── security.yaml          # BearerAuth
└── domains/
    ├── auth/paths.yaml
    ├── user/paths.yaml
    ├── role/paths.yaml
    ├── workspace/paths.yaml   # workspace + collaboration_workspace
    ├── menu/paths.yaml        # menu + navigation
    ├── permission/paths.yaml
    ├── featurepackage/paths.yaml
    ├── system/paths.yaml
    ├── message/paths.yaml
    ├── page/paths.yaml
    ├── api_endpoint/paths.yaml
    └── media/paths.yaml
```

详细说明：[`backend/api/openapi/README.md`](../../backend/api/openapi/README.md)

---

## 检查清单

| # | 完成标志 |
|---|----------|
| 1 | `domains/{domain}/paths.yaml` 有 operationId、tags、x-permission-key、x-access-mode |
| 2 | `update-openapi.bat` / `make api` 无报错 |
| 3 | `api/gen/` 和 `openapi_seed.json` 已提交 |
| 4 | handler 方法实现，`go build ./...` 通过 |
| 5 | `go run ./cmd/migrate` 成功，DB 里出现新行 |
| 6 | 集成测试通过 |
| 7 | 前端 `schema.d.ts` 已更新 |

---

## 禁止事项

- 不直接编辑 `dist/openapi.yaml`（机器生成，会被覆盖）
- 不重新引入 legacy Gin module shell（`internal/modules/system/*/module.go` 已删除，不得重建）
- 不在 handler 里手写 `&gen.LoginUnauthorized{Code:401, ...}`，一律 `return nil, err`
- 不手工编辑 `openapi_seed.json` 和 `frontend/src/api/v5/error-codes.ts`（均为生成产物）
- 不 import `internal/api/errcode`（已废弃，仅 legacy gin handler 使用）
- 不 import `internal/pkg/authorization`（已废弃，等待删除）
