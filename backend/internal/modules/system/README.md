# backend/internal/modules/system

系统主域，聚合身份、权限、导航、页面、空间、注册等基础能力。

## 子目录

| 目录 | 说明 |
| --- | --- |
| `apiendpoint/` | API 注册表、权限绑定 |
| `app/` | 应用实体与应用级配置 |
| `auth/` | 登录、会话、当前用户上下文、中间件 |
| `dictionary/` | 字典查询与基础枚举 |
| `featurepackage/` | 功能包与版本分配 |
| `menu/` | 菜单定义、菜单树、序列化 |
| `models/` | system 域集中模型定义 |
| `navigation/` | 导航编译与运行时导航 |
| `page/` | 页面定义、页面运行时 |
| `permission/` | 权限键管理与消费分析 |
| `register/` | 注册入口与注册页配置 |
| `role/` | 角色与角色权限 |
| `siteconfig/` | 站点级配置 |
| `social/` | 社交登录 HTTP 入口与 OAuth 流程 |
| `space/` | 菜单空间、访问模式、Host 绑定 |
| `system/` | 系统聚合 facade、系统消息能力 |
| `upload/` | 上传系统（配置中心 + 运行时） |
| `user/` | 用户与用户权限视图 |
| `workspace/` | 统一工作区模型 |

## 真相入口

权限模型、模块接入、上传系统等开发约束见 [../../../Truth/](../../../Truth/)。
