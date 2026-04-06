# Workspace 权限迁移阶段记录

> 当前文档只保留最终阶段结论，不再继续维护早期过渡态细节。

## 当前状态

- 已完成：Phase 0、Phase 1、Phase 2、Phase 3、Phase 4、Phase 5、Phase 6、Phase 7、Phase 8、Phase 9、Phase 10
- 当前主线：`workspace / personal workspace / collaboration workspace`
- 当前结论：`tenant` 已退出当前业务主语义，保留给未来多租户系统

## 分阶段收口结果

| 阶段 | 状态 | 收口结果 |
| --- | --- | --- |
| Phase 0 | 已完成 | 建立迁移基线，明确 `workspace != menu-space`，明确平台权限属于 `personal workspace` |
| Phase 1 | 已完成 | 建立 workspace 领域模型、成员关系、角色绑定、功能包绑定和默认回填 |
| Phase 2 | 已完成 | 建立 `auth_workspace_id` 运行时上下文和 `/api/v1/workspaces/*` 主线接口 |
| Phase 3 | 已完成 | 角色与功能包主链切到 `workspace_role_bindings / workspace_feature_packages` 优先 |
| Phase 4 | 已完成 | 平台代管请求统一走 `target_workspace_id` 精确校验 |
| Phase 5 | 已完成 | 当前协作主接口切到 `/api/v1/collaboration-workspaces/*` |
| Phase 6 | 已完成 | 前端统一以 `workspaceStore` 为授权上下文源，公共层与页面文案切到个人空间 / 协作空间 |
| Phase 7 | 已完成 | 导航、菜单、消息域与运行时上下文统一按 workspace / collaboration workspace 产出 |
| Phase 8 | 已完成 | 迁移、预重命名、种子和默认数据切到 collaboration 命名 |
| Phase 9 | 已完成 | 后端编译检查、前端 lint、前端 build 与关键场景核对通过 |
| Phase 10 | 已完成 | 文档、术语表、变更日志与兼容边界完成最终收口 |

## 最终保留项

### 允许继续存在

- 旧表、旧索引和旧值的迁移处理逻辑
- 少量 legacy package / module path 中仍叫 `tenant` 的代码目录
- 为历史数据升级保留的迁移期 rename / value backfill 逻辑

### 不再允许作为当前主语义扩散

- `tenant` 表示协作空间
- `workspace_type = team`
- `X-Tenant-ID`
- `X-Team-Workspace-Id`
- `current_tenant_id`
- `source_tenant_id`
- `/api/v1/tenants/*`
- 页面级 `currentTenantId`

## 最终兼容边界

当前兼容边界只允许出现在以下类型中：

1. 升级迁移链：识别旧 tenant/team 数据并改写到 collaboration 命名。
2. 历史映射层：把旧库中的 team 语义转换为 collaboration workspace。
3. 少量 legacy 目录或模块名：作为代码组织遗留，不表示当前对外契约。

## 完成标准核对

- 当前主输出字段已经使用 collaboration 命名。
- 当前主请求头已经使用 `X-Auth-Workspace-Id` 与 `X-Collaboration-Workspace-Id`。
- 当前主 API 已使用 `/api/v1/workspaces/*` 与 `/api/v1/collaboration-workspaces/*`。
- 当前平台角色与平台功能包通过 `personal workspace` 生效。
- 当前协作空间角色与协作空间功能包通过 `collaboration workspace` 生效。
- 当前文档已把 `tenant` 定义为未来多租户保留名词。

## 后续建议

1. 如果继续演进，优先清理代码目录和内部变量里残留的 `team/tenant` 历史命名。
2. 未来真正引入多租户时，应新建 `tenant` 领域与专属 schema，不与当前协作空间复用。
3. 新增模块、字段、路由、header、种子和文档必须直接使用 collaboration 语义，禁止回退到 tenant/team 双轨。
