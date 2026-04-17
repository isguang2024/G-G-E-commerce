# 前端代码结构（frontend）

> 项目说明类文档：概述 `frontend/` 下面有什么。不是开发真相，遇冲突以 `frontend/truth.md` 和 `frontend/Truth/` 为准。

## 入口与路由

- `frontend/src/main.ts`：前端入口
- `frontend/src/App.vue`：应用壳层
- `frontend/src/router/`：路由、守卫、动态菜单和页面编译
- `frontend/src/store/`：状态管理
- `frontend/src/api/`：请求封装
- `frontend/src/hooks/`：组合式逻辑
- `frontend/src/utils/`：工具函数

`frontend/src/router/routes/staticRoutes.ts` 只保留静态路由，当前包括登录、注册、忘记密码、异常页和外部 iframe 容器。其余管理端页面由菜单、页面和 App 上下文驱动。

## 正式页面

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
| 协作角色权限 | `frontend/src/views/system/` | 协作角色、菜单、动作、功能包权限 |
| 访问追踪 | `frontend/src/views/system/access-trace/` | 访问记录与追踪 |
| 协作空间 | `frontend/src/views/` | 协作工作台、成员、消息、空间主页 |
| 消息工作台 | `frontend/src/views/message/` | 消息发送、接收组、记录、模板、导航 |
| 工作区 | `frontend/src/views/workspace/` | 工作区收件箱 |
| 异常 | `frontend/src/views/exception/` | 403、404、500 |
| 结果页 | `frontend/src/views/result/` | 成功、失败结果页 |
| 外部页 | `frontend/src/views/outside/` | iframe 外挂容器 |

## 公共组件

- `frontend/src/components/business/layout/`：`AdminWorkspaceHero`、`AppContextBadge`
- `frontend/src/components/business/permission/`：权限树、权限摘要、权限工作台
- `frontend/src/components/business/tables/`：工作区分页等业务表格组件
- `frontend/src/components/core/layouts/`：头部栏、菜单、通知、页面容器、设置面板、工作台标签
- `frontend/src/components/core/views/`：登录页、异常页、结果页等基础视图

## 资源与工具

- `frontend/src/assets/`：图片、样式、图标
- `frontend/src/locales/`：国际化文案
- `frontend/src/types/`：接口和组件类型
- `frontend/src/mock/`：本地 mock 数据
- `frontend/src/plugins/`：插件注册

## 业务映射（前端入口）

| 业务主题 | 前端入口 |
| --- | --- |
| 权限 | `frontend/src/views/system/action-permission/`、`role/`、`user/`、`feature-package/`、`api-endpoint/`、`permission-simulator/` |
| 用户 | `frontend/src/views/auth/`、`frontend/src/views/system/user/` |
| 空间 | `frontend/src/views/workspace/`、`frontend/src/views/system/menu-space/`、`frontend/src/views/` |
| 菜单 | `frontend/src/views/system/menu/`、`frontend/src/views/system/page/` |
| APP | `frontend/src/views/system/app/` |

## 关键边界

- 前端正式页面优先改 `frontend/src/views/`
