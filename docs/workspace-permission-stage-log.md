# 空间权限迁移阶段记录

> 当前阶段先锁术语，再按术语回写代码、迁移、种子和 API。

## 当前状态

- 已完成：术语基线重置
- 当前主线：`workspace / personal workspace / collaboration workspace`
- 当前目标：把前后端、迁移、种子、API 全部回写到这套语义

## 阶段结论

| 阶段 | 状态 | 结论 |
| --- | --- | --- |
| Phase 0-8 | 历史基础已形成 | 历史工作已把权限主线推进到 `workspace`，但文档和实现里仍残留“平台/团队/tenant”混用 |
| Phase 9 | 当前进行中 | 先收口文档，明确统一空间上下文模型：`auth_workspace_id + auth_workspace_type` |
| Phase 10 | 待回写代码 | 按新术语回写后端、迁移、种子、前端与 API，并删除旧 `tenant/team` 主输出 |

## 本轮固定结论

1. 当前系统只有一套 `workspace` 权限模型。
2. `workspace_type` 只有 `personal | collaboration`。
3. `personal` 表示个人空间，不等于“平台上下文”。
4. `collaboration` 表示协作空间，不等于“团队上下文”。
5. `platform` 只是业务域 / app，不是空间类型。
6. `tenant` 预留给未来多租户系统，不再表示当前协作空间。

## 本轮待完成项

1. 后端运行时、DTO、错误文案、诊断文案改成“空间权限”语义。
2. 迁移、模型和种子重建到 `personal | collaboration` 最终态。
3. 前端 API、store、页面、路由、文案移除 `tenant/team` 当前语义。
4. 删除旧主路径、旧 header、旧 DTO 输出。

## 不再允许继续扩散

- `workspace_type = team`
- `/api/v1/tenants/*`
- `X-Tenant-ID`
- `X-Team-Workspace-Id`
- `current_tenant_id`
- “平台上下文 / 团队上下文”
- “平台权限 / 团队权限”
