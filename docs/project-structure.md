# 项目结构

> 本文只描述当前有效代码结构，不记录迁移历史。

## 总览

| 层级 | 路径 | 说明 |
| --- | --- | --- |
| 后端 | `backend/` | Go 后端 API、领域模块、模型与命令 |
| 前端 | `frontend/` | Vue 3 + TypeScript 管理端主工程 |
| 文档 | `docs/` | 当前结构说明、接入说明、变更记录 |

## 后端结构

### 启动与命令

- `backend/cmd/server/`：服务入口
- `backend/cmd/migrate/`：数据库迁移入口
- `backend/cmd/init-admin/`：初始化管理员
- `backend/cmd/init-demo/`：初始化示例数据
- `backend/cmd/diagnose/`：诊断命令
- `backend/cmd/dbreset/`：数据库重置命令
- `backend/cmd/routecodegen/`：路由代码生成

### 核心模块

| 模块 | 主要职责 | 典型文件 |
| --- | --- | --- |
| `auth` | 登录、JWT、中间件、当前用户信息 | `handler.go`、`middleware.go`、`service.go` |
| `app` | App 管理、Host 绑定、当前 App 解析 | `handler.go`、`service.go` |
| `workspace` | 个人空间与协作空间的统一主体 | `handler.go`、`service.go` |
| `collaborationworkspace` | 协作空间成员、边界、授权快照 | `handler.go`、`service.go` |
| `space` | 菜单空间、空间模式、Host 绑定、初始化 | `handler.go`、`service.go`、`util.go` |
| `menu` | 菜单树、菜单组、菜单权限、备份 | `handler.go`、`service.go`、`serializer.go` |
| `page` | 页面定义、页面空间绑定、运行时缓存 | `handler.go`、`service.go`、`runtime_cache.go` |
| `role` | 角色、生效 App 范围、角色权限 | `handler.go`、`service.go` |
| `user` | 用户、用户角色、个人覆盖、用户功能包 | `handler.go`、`service.go`、`repository.go` |
| `permission` | 功能权限、动作权限、权限审计 | `handler.go`、`service.go`、`audit.go` |
| `featurepackage` | 功能包、功能包菜单/动作绑定、组合包 | `handler.go`、`service.go` |
| `apiendpoint` | API 注册、分类、权限绑定 | `handler.go`、`service.go`、`permission_audit.go` |
| `navigation` | 导航编译、路由生成、动态菜单辅助 | `handler.go`、`service.go` |
| `system` | 系统级消息与基础能力 | `handler.go`、`service.go`、`message_service.go` |

### 数据模型

`backend/internal/modules/system/models/` 是领域模型集中地，当前主要对象包括：

- `App`、`AppHostBinding`
- `MenuSpace`、`MenuSpaceHostBinding`
- `MenuDefinition`、`SpaceMenuPlacement`
- `UIPage`、`PageSpaceBinding`
- `Workspace`、`WorkspaceMember`、`WorkspaceRoleBinding`、`WorkspaceFeaturePackage`
- `CollaborationWorkspace`、`CollaborationWorkspaceMember`
- `PermissionKey`、`PermissionGroup`
- `FeaturePackage`、`FeaturePackageBundle`、`FeaturePackageMenu`、`FeaturePackageKey`
- `RoleFeaturePackage`、`RoleDataPermission`、`RoleDisabledAction`
- `UserRole`、`UserFeaturePackage`、`UserActionPermission`
- `Message`、`MessageDelivery`、`MessageTemplate`、`MessageSender`、`MessageRecipientGroup`
- `APIEndpoint`、`APIEndpointPermissionBinding`、`APIEndpointCategory`

## 前端结构

### 入口与路由

- `frontend/src/main.ts`：前端入口
- `frontend/src/App.vue`：应用壳层
- `frontend/src/router/`：路由、守卫、动态菜单和页面编译
- `frontend/src/store/`：状态管理
- `frontend/src/api/`：请求封装
- `frontend/src/hooks/`：组合式逻辑
- `frontend/src/utils/`：工具函数

`frontend/src/router/routes/staticRoutes.ts` 只保留静态路由，当前包括登录、注册、忘记密码、异常页和外部 iframe 容器。其余管理端页面由菜单、页面和 App 上下文驱动。

### 正式页面

| 页面域 | 路径 | 主要内容 |
| --- | --- | --- |
| 认证 | `frontend/src/views/auth/` | 登录、注册、忘记密码 |
| 控制台 | `frontend/src/views/dashboard/console/` | 首页概览、活跃用户、统计、待办 |
| App | `frontend/src/views/system/app/` | App 管理、Host 绑定、默认空间、空间模式 |
| 菜单 | `frontend/src/views/system/menu/` | 菜单树、菜单编辑、备份、权限绑定 |
| 菜单空间 | `frontend/src/views/system/menu-space/` | App 下的空间列表、空间模式、Host 绑定、初始化 |
| 页面 | `frontend/src/views/system/page/` | 页面定义、页面分组、页面空间暴露 |
| 用户 | `frontend/src/views/system/user/` | 用户管理、用户菜单、用户功能包、权限测试 |
| 角色 | `frontend/src/views/system/role/` | 角色管理、角色权限、角色功能包 |
| 功能包 | `frontend/src/views/system/feature-package/` | 功能包目录、菜单/动作/空间绑定 |
| 功能权限 | `frontend/src/views/system/action-permission/` | 动作权限、权限分组、空间范围 |
| API 注册 | `frontend/src/views/system/api-endpoint/` | API 注册、权限绑定、分类检索 |
| 权限模拟 | `frontend/src/views/system/permission-simulator/` | 权限结果模拟与验证 |
| 作用域 | `frontend/src/views/system/scope/` | 作用域相关管理页 |
| 协作角色权限 | `frontend/src/views/system/collaboration-workspace-roles-permissions/` | 协作空间角色、菜单、动作、功能包权限 |
| 访问追踪 | `frontend/src/views/system/access-trace/` | 访问记录与追踪 |
| 协作空间 | `frontend/src/views/collaboration-workspace/` | 协作空间工作台、成员、消息、空间主页 |
| 消息工作台 | `frontend/src/views/message/` | 消息发送、接收组、记录、模板、导航 |
| 工作区 | `frontend/src/views/workspace/` | 工作区收件箱 |
| 异常 | `frontend/src/views/exception/` | 403、404、500 |
| 结果页 | `frontend/src/views/result/` | 成功、失败结果页 |
| 外部页 | `frontend/src/views/outside/` | iframe 外挂容器 |

### 公共组件

- `frontend/src/components/business/layout/`：`AdminWorkspaceHero`、`AppContextBadge`
- `frontend/src/components/business/permission/`：权限树、权限摘要、权限工作台
- `frontend/src/components/business/tables/`：工作区分页等业务表格组件
- `frontend/src/components/core/layouts/`：头部栏、菜单、通知、页面容器、设置面板、工作台标签
- `frontend/src/components/core/views/`：登录页、异常页、结果页等基础视图

### 资源与工具

- `frontend/src/assets/`：图片、样式、图标
- `frontend/src/locales/`：国际化文案
- `frontend/src/types/`：接口和组件类型
- `frontend/src/mock/`：本地 mock 数据
- `frontend/src/plugins/`：插件注册

## 业务映射

| 业务主题 | 后端入口 | 前端入口 |
| --- | --- | --- |
| 权限 | `backend/internal/modules/system/permission/`、`role/`、`user/`、`featurepackage/`、`apiendpoint/` | `frontend/src/views/system/action-permission/`、`role/`、`user/`、`feature-package/`、`api-endpoint/`、`permission-simulator/` |
| 用户 | `backend/internal/modules/system/auth/`、`user/` | `frontend/src/views/auth/`、`frontend/src/views/system/user/` |
| 空间 | `backend/internal/modules/system/workspace/`、`collaborationworkspace/`、`space/` | `frontend/src/views/collaboration-workspace/`、`frontend/src/views/workspace/`、`frontend/src/views/system/menu-space/` |
| 菜单 | `backend/internal/modules/system/menu/`、`navigation/`、`page/` | `frontend/src/views/system/menu/`、`frontend/src/views/system/page/` |
| APP | `backend/internal/modules/system/app/` | `frontend/src/views/system/app/` |

## 关键边界

- 权限优先看 `permission`、`role`、`user`、`featurepackage`、`apiendpoint`
- 用户优先看 `auth` 与 `user`
- 空间优先看 `workspace`、`collaborationworkspace`、`space`
- 菜单优先看 `menu`、`navigation`、`page`
- APP 优先看 `app`
- 前端正式页面优先改 `frontend/src/views/`
- 后端业务逻辑优先改 `backend/internal/modules/system/`

## 当前有效文档

- `AGENTS.md` — 协作约束
- `docs/project-framework.md` — 项目架构与边界
- `docs/frontend-guideline.md` — 前端实现规范
- `backend/CLAUDE.md` — 后端开发指引
- `docs/project-structure.md` — 代码结构与模块分工
