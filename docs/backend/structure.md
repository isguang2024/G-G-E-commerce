# 后端代码结构

> 项目说明类文档：概述 `backend/` 下面有什么。不是开发真相，遇冲突以 `backend/truth.md` 和 `backend/Truth/` 为准。

## 启动与命令

- `backend/cmd/server/`：服务入口
- `backend/cmd/migrate/`：数据库迁移入口
- `backend/cmd/init-admin/`：初始化管理员
- `backend/cmd/init-demo/`：初始化示例数据
- `backend/cmd/diagnose/`：诊断命令
- `backend/cmd/dbreset/`：数据库重置命令
- `backend/cmd/routecodegen/`：路由代码生成

## 核心模块

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

## 数据模型

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

## 业务映射（后端入口）

| 业务主题 | 后端入口 |
| --- | --- |
| 权限 | `backend/internal/modules/system/permission/`、`role/`、`user/`、`featurepackage/`、`apiendpoint/` |
| 用户 | `backend/internal/modules/system/auth/`、`user/` |
| 空间 | `backend/internal/modules/system/workspace/`、`collaborationworkspace/`、`space/` |
| 菜单 | `backend/internal/modules/system/menu/`、`navigation/`、`page/` |
| APP | `backend/internal/modules/system/app/` |

## 关键边界

- 权限优先看 `permission`、`role`、`user`、`featurepackage`、`apiendpoint`
- 用户优先看 `auth` 与 `user`
- 空间优先看 `workspace`、`collaborationworkspace`、`space`
- 菜单优先看 `menu`、`navigation`、`page`
- APP 优先看 `app`
- 后端业务逻辑优先改 `backend/internal/modules/system/`
