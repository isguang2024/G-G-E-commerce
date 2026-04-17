# 权限系统结构与调试

---

## 最终权限公式

```
用户实际可访问权限 = 空间已开通功能包权限键 ∩ 成员角色权限键
```

两个集合必须同时满足，缺一不可。有两个旁路：

- **owner / admin 角色**：绕过功能包检查，直接通过
- **super_admin 账号**：绕过所有检查

---

## 三层数据结构

```
功能包（FeaturePackage）
  └─ 包含多个权限键（permission_keys）
  └─ 绑定到协作空间（collaboration_workspace）

角色（Role）
  └─ 绑定多个功能包（role_feature_packages）

用户（User）
  └─ 在空间内有角色（collaboration_workspace_members）
  └─ 角色 → 功能包 → 权限键（运行时交集）
```

---

## 个人空间的实际叠加规则

协作空间的主模型是“空间功能包 ∩ 成员角色”。个人空间当前实现额外保留了“用户直配功能包”和“用户菜单裁剪”两条链路，所以最终结果不是只看角色。

### 个人空间能力来源

- 角色功能包：来自个人空间角色绑定
- 用户功能包：直接挂在用户个人空间上的功能包
- 角色菜单裁剪：从角色功能包展开菜单中做减法
- 用户菜单裁剪：对个人空间最终候选菜单再做一次减法

### 个人空间菜单计算顺序

```text
角色功能包展开菜单
- 角色隐藏菜单
+ 用户功能包展开菜单
+ 公共菜单
- 用户隐藏菜单
= 用户最终可见菜单
```

对应当前实现可以理解为：

- 角色功能包与用户功能包是并集，不是覆盖
- 角色里隐藏掉的菜单，如果用户又直配了包含该菜单的功能包，该菜单会重新进入候选集合
- 只有用户菜单裁剪再次隐藏，菜单才会在最终结果里消失

### 例子

假设：

- 角色功能包：`A + B + C`
- 用户功能包：`A + B`
- 角色隐藏菜单：`B.a`

当前结果：

- 功能包并集仍然是 `A + B + C`
- `B.a` 会先从角色菜单贡献中被减掉
- 但用户功能包里的 `B` 会再次把 `B.a` 带回候选菜单
- 如果用户菜单裁剪没有继续隐藏 `B.a`，那么 `B.a` 最终仍会显示

这也是为什么：

- 只要还保留“用户功能包”入口，就不能单独删除“用户菜单裁剪”
- 如果后续要收敛到“角色为主”，应该成对下线“用户功能包 + 用户菜单裁剪”，统一改为“个人空间角色 + 角色边界”

---

## 请求鉴权流程

```
HTTP 请求
  │
  ▼
ogen middleware（openapiperm.go）
  │  读取 operation 的 x-access-mode
  ├─ public       → 直接放行
  ├─ authenticated → 只验 JWT
  └─ permission   → 调用 evaluator.Can(userID, permissionKey, workspaceID)
                         │
                         ▼
                   evaluator.Resolve(ctx)
                         │
                    ┌────┴────┐
                    │         │
               功能包集合   角色集合
                    │         │
                    └────∩────┘
                         │
                    true / false
```

---

## 评估器调用方式

`evaluator.Can` 在中间件中自动调用，handler 内无需手动检查权限。

如需在 service 层做二次校验（如数据过滤），注入 `evaluator` 并调用：

```go
ok, err := h.evaluator.Can(ctx, userID, "widget.read", workspaceID)
```

---

## 调试：/permissions/explain 接口

调用以下接口可以看到当前用户对某个权限键的评估明细：

```
GET /api/v1/permissions/explain?key=widget.read
Authorization: Bearer <token>
X-Workspace-ID: <workspace_uuid>
```

响应会列出：
- 用户当前空间的功能包
- 角色携带的权限键
- 最终是否通过

---

## 启动时一致性检查

server 启动时会对比 `openapi_seed.json` 中的权限键与 DB 中 `permission_keys` 表的差异，不一致时打印 WARN 日志（不会阻止启动）。

如果看到类似日志：

```
WARN  permission key in spec but missing in DB  {"key": "widget.read"}
```

说明 `go run ./cmd/migrate` 尚未运行，执行一次即可。

---

## 权限键命名规范

格式：`{模块}.{动作}` 或 `{模块}.{子模块}.{动作}`

| 示例 | 说明 |
|------|------|
| `user.list` | 查看用户列表 |
| `user.create` | 创建用户 |
| `feature_package.assign_collaboration_workspace` | 给协作空间分配功能包 |
| `system.api_registry.view` | 查看 API 注册表 |

第一段为模块名，会自动映射到 `permission_keys.module_code`，同时作为接口分类推断的输入。

---

## 相关代码位置

| 功能 | 文件 |
|------|------|
| 评估器核心逻辑 | `backend/internal/pkg/permission/evaluator/evaluator.go` |
| ogen 中间件 | `backend/internal/api/middleware/openapiperm.go` |
| `/permissions/explain` handler | `backend/internal/api/handlers/permission.go` |
| 功能包 seed | `backend/internal/pkg/permissionseed/seeds.go` |
| DB 一致性检查 | `backend/internal/pkg/permissionseed/openapi_ensure.go` |
