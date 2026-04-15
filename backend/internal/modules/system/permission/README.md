# permission

这个目录负责“权限定义管理”和“权限消费面分析”，不是运行时最终判权器。

## 当前文件

| 文件 | 说明 |
| --- | --- |
| `service.go` | 权限键、权限分组、消费者详情、批量更新、模板、影响面预览 |
| `audit.go` | 风险操作审计相关逻辑 |
| `audit_test.go` | 风险审计测试 |
| `service_test.go` | 权限服务测试 |

## 负责什么

- 权限键和权限分组管理
- 权限和 API / 页面 / 功能包 / 角色之间的消费关系分析
- 权限批量更新与模板保存
- 风险操作审计查询
- 权限种子清理和 API 端点绑定修正

## 不负责什么

- 不在这里做最终权限判定；运行时判权统一走 `internal/pkg/permission/evaluator`
- 不直接手改 OpenAPI 生成产物；权限种子来自 `backend/api/openapi/` 和生成链
- 不把菜单权限、功能权限、数据权限混成一个模型

## 相关目录

- `../role/`：角色与角色侧权限生效
- `../featurepackage/`：功能包与权限聚合
- `../apiendpoint/`：API 注册与权限绑定检查
- `../../../../internal/pkg/permissionseed/`：权限种子生成与 ensure
