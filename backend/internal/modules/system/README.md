# backend/internal/modules/system

`system/` 是当前后端的基础业务域，聚合“身份、权限、导航、页面、空间、注册”这几类通用能力。

## 子目录

| 目录 | 说明 |
| --- | --- |
| `apiendpoint/` | API 注册表、权限绑定检查、权限审计补偿 |
| `app/` | 应用实体与应用级配置 |
| `auth/` | 登录、会话、当前用户上下文、中间件 |
| `collaborationworkspace/` | 协作空间与成员管理、协作空间角色切换 |
| `dictionary/` | 字典查询与基础枚举读取 |
| `featurepackage/` | 功能包、功能包版本、分配与审计 |
| `menu/` | 菜单定义、菜单树、菜单序列化 |
| `models/` | system 域集中模型定义 |
| `navigation/` | 导航编译与运行时导航结果 |
| `page/` | 页面定义、页面运行时、缓存与同步 |
| `permission/` | 权限键、权限分组、消费者分析与风险审计 |
| `register/` | 注册入口解析、注册页配置与仓储读写 |
| `role/` | 角色、角色权限与角色生效结果 |
| `social/` | 社交登录 HTTP 入口、OAuth 流程与状态存储 |
| `space/` | 菜单空间、空间访问模式、Host 绑定和初始化 |
| `system/` | 系统聚合 facade、系统消息能力 |
| `user/` | 用户、用户仓储、用户侧权限视图 |
| `workspace/` | 个人空间与协作空间对应的统一工作区模型 |

## 关系说明

- `workspace/` 负责统一工作区模型；`collaborationworkspace/` 负责协作空间实体和成员关系；`space/` 负责菜单空间与访问入口，不是同一层概念。
- `permission/`、`role/`、`featurepackage/`、`apiendpoint/` 一起构成权限主链；权限判断仍以 `internal/pkg/permission/evaluator` 为准。
- `menu/`、`navigation/`、`page/` 一起负责运行时导航和页面暴露。
- `models/` 是 system 域共享模型层，新字段或新表优先先落模型和迁移，再回到 service 接入。

## 修改原则

- 先按目录找职责，不要在相邻目录重复造一层 service。
- 目录下没有 `handler.go` 并不代表模块无效；HTTP 入口已经逐步收束到 `internal/api/handlers/`。
- 需要新增 README 时，优先写局部边界、关键文件和上下游关系，不重复写全局规则。
