# 菜单与页面管理当前架构

> 现状基线：2026-03-27。本文描述已经采用的正式模型，不再保留“旧内页菜单继续作为主链”的阶段性方案。

## 1. 当前结论

当前菜单与页面体系已经明确分层：

- 菜单：导航入口
- 页面：访问单元
- 权限：访问或动作边界
- 功能包：开通范围

一句话总结：
- 菜单管入口，页面管访问，权限键管动作，功能包管开通。

## 2. 菜单正式模型

### 2.1 菜单职责

菜单当前只负责：
- 导航树结构
- 排序与显隐
- 图标与展示元数据
- 访问模式 `meta.accessMode`
- 与功能包的入口绑定

当前允许对系统管理等大型模块追加“分组目录”作为中间层，形成三级菜单；
这类分组目录只用于导航归类，不改变页面组件归属，也不要求业务页 URL 必须随层级加深。

菜单不再负责：
- 详情页、编辑页、流程页等访问单元管理
- “内页类型”页面模型

### 2.2 访问模式

菜单当前正式访问模式：
- `permission`
- `jwt`
- `public`

含义：
- `permission`：按功能包展开结果和菜单权限链控制可见性
- `jwt`：登录即可显示
- `public`：公开显示

当前运行时菜单树会对 `jwt/public` 菜单做默认放行，不依赖功能包分配。

### 2.3 菜单管理分组

菜单管理分组已独立为 `menu_manage_groups`。

它只影响：
- 菜单管理页的分组展示
- 菜单备份/恢复时的管理组织结构

它不影响：
- 运行时菜单树
- 权限计算
- 路由访问路径

## 3. 页面正式模型

### 3.1 页面主表

页面主表为 `ui_pages`，当前页面类型包括：
- `inner`
- `global`
- `group`
- `display_group`

### 3.2 各页面类型语义

- `inner`：内页，通常挂到菜单或上级页面下
- `global`：全局页，独立访问，不要求挂菜单
- `group`：逻辑分组，参与父子页面链路，不生成实际运行时页面
- `display_group`：普通分组，仅用于页面管理列表归类，不参与运行时注册

### 3.3 页面核心字段

页面正式语义主要由以下字段承载：
- `page_key`
- `route_name`
- `route_path`
- `component`
- `parent_menu_id`
- `parent_page_key`
- `display_group_key`
- `breadcrumb_mode`
- `access_mode`
- `permission_key`
- `active_menu_path`

## 4. 菜单与页面的边界

### 4.1 挂载关系

页面可以：
- 直接挂到菜单：`parent_menu_id`
- 挂到上级页面：`parent_page_key`
- 仅放在普通分组下管理：`display_group_key`
- 完全不挂载，作为独立访问页存在

### 4.2 路径解析规则

当前正式路径解析规则：
- 多段绝对路径直接按完整路径注册
- 单段路径会按上级菜单或上级页面路径自动拼接
- 外链页面在内嵌模式下使用 `/outside/Iframe`

### 4.3 面包屑与高亮

页面当前通过以下信息决定面包屑与菜单高亮：
- `active_menu_path`
- `parent_menu_id`
- `parent_page_key`
- `breadcrumb_mode`

支持模式：
- `inherit_menu`
- `inherit_page`
- `custom`

## 5. 页面访问模型

页面正式访问模式：
- `inherit`
- `public`
- `jwt`
- `permission`

解释顺序：

1. `public`：公开访问
2. `jwt`：登录即可
3. `permission`：按 `permission_key` 校验
4. `inherit`：优先继承上级页面，再继承上级菜单；若都不存在，则退回登录访问

## 6. 运行时路由主链

### 6.1 后端输出

后端正式提供：
- `/api/v1/pages/runtime`
- `/api/v1/pages/runtime/public`

它们输出当前运行时页面注册表，不再要求前端自行拼另一套页面配置。

### 6.2 前端消费

前端当前路由链路是：
- 静态路由
- 后端菜单路由
- 运行时页面路由

运行时页面由 `ManagedPageProcessor` 基于菜单树和 `ui_pages` 注册结果动态生成。

## 7. 页面同步与未注册页面

页面管理当前支持：
- 扫描 `frontend/src/views/**` 生成未注册页面候选
- 同步写入 `ui_pages`
- 从页面管理页直接将候选页面转成正式页面记录

当前扫描逻辑默认排除：
- `auth`
- `exception`
- `result`
- `views/**/modules`
- `/outside/Iframe`

## 8. 当前已经退出主链的旧语义

- 菜单“内页类型”已退出正式主链
- 页面管理不再依赖“旧 inner 菜单”做主访问模型
- 普通分组和菜单管理分组都不再承担运行时权限语义

## 9. 当前实施判断标准

- 需要导航入口：建菜单
- 需要详情页、编辑页、流程页、独立访问页：建页面
- 需要管理页归类：用 `display_group`
- 需要父子访问链：用 `parent_page_key`
- 需要控制菜单高亮：优先用 `parent_menu_id`，必要时补 `active_menu_path`
