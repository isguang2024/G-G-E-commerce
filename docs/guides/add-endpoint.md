# 新增后端接口：完整 7 步

> 规则：**OpenAPI spec 是唯一事实源**。所有步骤从改 yaml 开始，不允许在 Go 里先写 handler 再补文档。

---

## 步骤一：在 openapi.yaml 声明接口

文件路径：`backend/api/openapi/openapi.yaml`

每条 operation 必须带以下四个 `x-` 扩展：

```yaml
/widgets/{id}:
  get:
    operationId: getWidget          # 唯一，驼峰命名
    summary: 获取组件详情            # 会自动写入 api_endpoints.summary
    tags: [widget]                  # 第一个 tag 用于自动分配接口分类
    x-permission-key: widget.read   # 权限键，public/authenticated 模式时可省
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
              $ref: '#/components/schemas/Widget'
```

**`x-access-mode` 说明：**

| 值 | 含义 |
|----|------|
| `permission` | 需要 JWT + 权限键（默认） |
| `authenticated` | 需要 JWT，不校验权限键 |
| `public` | 完全公开，无需登录 |

---

## 步骤二：重新生成 ogen 服务端代码

```bash
cd backend
go run github.com/ogen-go/ogen/cmd/ogen@latest \
  --target api/gen --package gen --clean api/openapi/openapi.yaml
```

生成的 `api/gen/` 目录内容需要提交到 git。

---

## 步骤三：刷新权限 seed

```bash
go run ./cmd/gen-permissions
```

这会更新 `internal/pkg/permissionseed/openapi_seed.json`，内含新接口的：
- 权限键
- summary / description
- tags（用于自动分配分类）

此文件需提交到 git。

---

## 步骤四：实现 handler

在 `internal/api/handlers/{domain}.go` 里实现 ogen 生成的接口方法：

```go
// GetWidget implements gen.Handler.
func (h *APIHandler) GetWidget(ctx context.Context, params gen.GetWidgetParams) (gen.GetWidgetRes, error) {
    userID := bridge.MustUserID(ctx)
    widget, err := h.widgetSvc.Get(ctx, userID, params.ID)
    if err != nil {
        return &gen.ErrorStatusCode{StatusCode: 404, Response: gen.Error{Message: "not found"}}, nil
    }
    return &gen.Widget{ID: widget.ID, Name: widget.Name}, nil
}
```

**不要**重新引入 legacy Gin module shell（`internal/modules/system/*/module.go` 已全部删除），ogen bridge 已统一接管所有路由。

---

## 步骤五：权限键自动入库

运行 `go run ./cmd/migrate`，`EnsureOpenAPIPermissionKeys` 会自动将新的 `x-permission-key` 写入 `permission_keys` 表。

`EnsureOpenAPIEndpoints` 同时会将新接口写入 `api_endpoints` 表，并绑定权限键。

详见 [api-auto-registration.md](api-auto-registration.md)。

---

## 步骤六：添加集成测试

在 `internal/api/handlers/integration_test.go` 追加：

```go
//go:build integration

func TestGetWidget(t *testing.T) {
    token := integToken(t, "admin@example.com")
    res := integDo(t, token, "GET", "/api/v1/widgets/"+testWidgetID, nil)
    assert.Equal(t, 200, res.StatusCode)
}
```

运行：

```bash
go test -tags integration ./internal/api/handlers/...
```

---

## 步骤七：更新前端客户端

在 `frontend/` 目录重新生成 TypeScript 类型：

```bash
npx openapi-typescript@7 ../backend/api/openapi/openapi.yaml -o src/api/v5/schema.d.ts
```

然后在前端代码中通过 `v5Client` 调用新接口，替换旧的 axios 调用。

---

## 检查清单

| # | 完成标志 |
|---|----------|
| 1 | `openapi.yaml` 有 operationId、tags、x-permission-key、x-access-mode |
| 2 | `api/gen/` 已重新生成并提交 |
| 3 | `openapi_seed.json` 已重新生成并提交 |
| 4 | handler 方法实现，`go build ./...` 通过 |
| 5 | `go run ./cmd/migrate` 成功，DB 里出现新行 |
| 6 | 集成测试通过 |
| 7 | 前端 `schema.d.ts` 已更新，旧 axios 调用已替换 |

---

## 禁止事项

- 不重新引入 legacy Gin module shell（`internal/modules/system/*/module.go` 已删除，不得重建）
- 不手写 DTO 绕过 ogen 生成
- 不在 handler 内直接读 `feature_package_keys` / `role_feature_packages` 表
- 不 import `internal/pkg/authorization`（已废弃，等待删除）
- 不手工编辑 `openapi_seed.json`（由 gen-permissions 生成）
