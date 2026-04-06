# 空间权限迁移说明

> 当前文档定义最终目标语义，并作为后续前后端、迁移、种子和 API 回写的唯一基线。

## 目标模型

- 当前系统只有一套权限模型：`workspace`
- `workspace_type` 只有两个有效值：
  - `personal`
  - `collaboration`
- 中文统一为：
  - `个人空间`
  - `协作空间`
- `tenant` 不再表示当前协作业务空间，预留给未来真正的多租户系统

## 核心规则

### 1. 权限主体

- 所有权限、角色、功能包、菜单、动作都绑定在 `workspace`
- 当前权限判断永远基于空间，不基于账号，也不基于 `tenant`
- 运行时权限上下文固定为：
  - `auth_workspace_id`
  - `auth_workspace_type`
- 跨空间操作时再显式携带：
  - `target_workspace_id`

### 2. 空间类型

- `personal workspace` 表示个人空间
- `collaboration workspace` 表示协作空间
- `personal` 与 `collaboration` 是同一套空间模型下的两种类型，不是两套权限系统
- `platform` 只是业务域 / app，不是权限上下文类型，也不能等价于 `personal workspace`

### 3. 业务域与导航

- `app` 表示业务域，例如平台后台或某个子产品域
- `menu-space` 表示 APP 内导航宿主，不表示权限空间
- 是否能访问某个业务域能力，由 `workspace` 上绑定的权限集合决定，而不是由“平台上下文 / 团队上下文”决定

## 当前收口目标

### 1. 后端

- 所有活跃 handler、service、runtime helper 都使用 `workspace` 语义
- 当前活跃接口只保留：
  - `/api/v1/workspaces/*`
  - `/api/v1/collaboration-workspaces/*`
  - `/api/v1/collaboration-workspaces/current/*`
- 当前活跃请求头只保留：
  - `X-Auth-Workspace-Id`
  - `X-Collaboration-Workspace-Id`

### 2. 前端

- 唯一授权上下文源是 `workspaceStore`
- 协作空间派生字段统一使用：
  - `currentCollaborationWorkspaceId`
  - `collaborationWorkspaceList`
- 页面、组件、类型、文案统一使用：
  - `空间权限`
  - `个人空间权限`
  - `协作空间权限`

### 3. 迁移与种子

- 当前项目为本地开发空项目，本轮按硬切处理
- 不保留旧 `tenant/team` 主输出，不维持长期双轨兼容
- schema、默认值、seed、权限键、菜单、页面、消息模板统一切到：
  - `workspace`
  - `personal`
  - `collaboration`

## 不再允许扩散的旧语义

- `tenant` 表示当前协作空间
- `workspace_type = team`
- `X-Tenant-ID`
- `X-Team-Workspace-Id`
- `current_tenant_id`
- `target_tenant_id`
- `source_tenant_id`
- `/api/v1/tenants/*`
- “平台上下文 / 团队上下文” 作为权限模型一级概念
- “平台权限 vs 团队权限” 作为顶层权限分类

## 实施顺序

1. 先收口本文档与术语表，定死最终语义
2. 再回写后端运行时、迁移、模型、种子和 API
3. 再回写前端 API、store、页面、文案和路由
4. 最后做 lint、build、编译检查与场景核对

## 完成标准

- 当前系统只存在一套 `workspace` 权限模型
- `workspace_type` 只有 `personal | collaboration`
- 活跃前后端契约不再输出 `tenant/team` 当前业务语义
- 活跃文案不再把 `personal` 写成“平台上下文”
- 活跃文案不再把“平台权限 / 团队权限”当成一级权限模型
