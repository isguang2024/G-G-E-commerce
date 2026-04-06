# 协作空间全量改名落地 Implementation Plan

> **Execution note:** Use superpowers:subagent-driven-development when tasks are independent, or superpowers:executing-plans when this plan should be executed inline in a controlled sequence.

**Goal:** 将前端主线中的 `team/tenant` 协作语义统一切换为 `collaboration workspace`，并收敛成单一 canonical API、单一命名出口与兼容壳。

**Architecture:** 保留 `frontend/src/store/modules/workspace.ts` 作为总上下文，新增/收口 `currentCollaborationWorkspaceId`、`collaborationWorkspaceList` 等派生状态；将 `team.ts` 降级为兼容 re-export 壳，把 `collaboration-workspace.ts` 作为唯一 active API 入口。页面层通过目录迁移与路由/菜单/文案重命名完成协作空间主语义切换，消息、收件箱、成员和系统管理页面同步改为协作空间命名。HTTP 层统一主请求头为 `X-Auth-Workspace-Id` + `X-Collaboration-Workspace-Id`，旧 `tenant` 透传只作为兼容别名。

**Tech Stack:** Vue 3, TypeScript, Pinia, Vue Router, Axios, Element Plus, Vite

---

### Task 1: 收敛 API 契约与请求头

**Files:**
- Modify: `frontend/src/api/collaboration-workspace.ts`
- Modify: `frontend/src/api/team.ts`
- Modify: `frontend/src/utils/http/index.ts`
- Modify: `frontend/src/api/message.ts`
- Modify: `frontend/src/api/workspace.ts`
- Modify: `frontend/src/types/api/api.d.ts`

- [ ] **Implement**

把 `collaboration-workspace.ts` 改成唯一 canonical 实现，`team.ts` 仅保留兼容 re-export，避免 `collaboration-workspace -> team -> collaboration-workspace` 的循环依赖。
将 HTTP 请求头主线统一为 `X-Auth-Workspace-Id` + `X-Collaboration-Workspace-Id`，并把 `skipTenantHeader` 逐步收敛为兼容别名。
同步把 workspace / auth / message 相关响应字段补齐为 snake_case 主字段，保留必要 camelCase 兼容读法。

- [ ] **Verify**

Run: `pnpm --dir frontend lint`
Expected: 无循环依赖告警、无 type/lint 错误

- [ ] **Notes**

- 这一步先保证契约和请求层稳定，再动页面目录和路由。

### Task 2: 迁移协作空间页面目录与路由入口

**Files:**
- Create/Move: `frontend/src/views/collaboration-workspace/**`
- Modify: `frontend/src/views/team/**`
- Modify: `frontend/src/router/**`
- Modify: `frontend/src/config/modules/fastEnter.ts`
- Modify: `frontend/src/views/dashboard/console/index.vue`

- [ ] **Implement**

将 `frontend/src/views/team/` 迁移为 `frontend/src/views/collaboration-workspace/`，把页面 key、路由名、菜单 key 与对外文案统一改为协作空间。
保留旧路径作为兼容壳时，确保不再作为活跃导入源。
同步更新系统管理页、消息中心、成员管理、顶部切换器与徽标文案。

- [ ] **Verify**

Run: `pnpm --dir frontend build`
Expected: 路由与组件解析通过

- [ ] **Notes**

- 动态路由如仍依赖旧路径，兼容壳只保留过渡期使用。

### Task 3: 收口 store 与公共壳层命名

**Files:**
- Modify: `frontend/src/store/modules/workspace.ts`
- Modify: `frontend/src/store/modules/collaboration-workspace.ts`
- Modify: `frontend/src/store/modules/tenant.ts`
- Modify: `frontend/src/components/core/layouts/art-header-bar/widget/ArtCollaborationWorkspaceSwitcher.vue`
- Modify: `frontend/src/components/business/layout/AppContextBadge.vue`
- Modify: `frontend/src/views/workspace/inbox/index.vue`
- Modify: `frontend/src/views/message/modules/useMessageWorkspace.ts`

- [ ] **Implement**

保持 `workspaceStore` 作为总上下文，明确 `currentCollaborationWorkspaceId` 与 `collaborationWorkspaceList` 的派生关系。
让 `tenantStore` 只作为兼容壳，不参与新活跃链路。
公共 header、badge、消息中心、收件箱统一使用“个人工作空间 / 协作空间”措辞。

- [ ] **Verify**

Run: `pnpm --dir frontend lint`
Expected: 通过

- [ ] **Notes**

- 若需要，补充少量类型兼容映射，但不重建整套状态树。

### Task 4: 收尾验证与变更记录

**Files:**
- Modify: `docs/change-log.md`

- [ ] **Implement**

补充本次协作空间全量改名的收尾记录、兼容保留项和验证命令。

- [ ] **Verify**

Run: `pnpm --dir frontend lint && pnpm --dir frontend build`
Expected: 两个命令均成功

- [ ] **Notes**

- 若某些旧路径/旧字段仍必须保留，仅作为兼容壳，不再作为默认导入源。
