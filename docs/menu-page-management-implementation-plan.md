# 菜单与页面管理实施草案

## 1. 文档定位

本文档承接：
- [菜单与页面管理设计方案](./menu-page-management-design.md)

本文档只回答“怎么落地”，不重复讲为什么设计成这样。

范围限定：
- 保留现有菜单系统和旧 `inner` 菜单。
- 不改现有菜单主链语义。
- 第一阶段仅新增：
  - 全局菜单访问模式
  - 页面管理主表
  - 页面管理接口
  - 前端路由守卫接入

## 2. 第一阶段目标

第一阶段交付完成后，应具备以下能力：

- 现有菜单中可以配置“全局菜单”，登录后默认可访问，不必绑定功能权限。
- 后端存在正式页面主表，可维护 `inner` 与 `global` 页面。
- 前端路由可以根据 `pageKey` 进入页面管理访问判定链路。
- 数据内页可以绑定上级菜单，继承面包屑和菜单高亮。
- 旧 `inner` 菜单继续能访问，作为兼容兜底。

## 3. 数据库设计

### 3.1 菜单访问模式

不新增菜单表字段，第一阶段继续写入 `menus.meta`：

```json
{
  "accessMode": "jwt"
}
```

允许值：
- `public`
- `jwt`
- `permission`

兼容规则：
- 未配置 `accessMode` 视为旧菜单，不强制回填。
- 旧菜单继续按当前链路解释。

### 3.2 页面主表

建议主表：`ui_pages`

参考现有 `models` 风格，字段建议：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `uuid` | 主键，`gen_random_uuid()` |
| `page_key` | `varchar(150)` | 页面唯一标识 |
| `name` | `varchar(150)` | 页面名称 |
| `route_name` | `varchar(150)` | 路由名称 |
| `route_path` | `varchar(255)` | 路由路径 |
| `component` | `varchar(255)` | 组件路径 |
| `page_type` | `varchar(20)` | `inner` / `global` |
| `source` | `varchar(20)` | `seed` / `sync` / `manual` |
| `module_key` | `varchar(100)` | 模块标识 |
| `parent_menu_id` | `uuid` | 上级菜单 ID |
| `parent_page_key` | `varchar(150)` | 上级页面 key |
| `active_menu_path` | `varchar(255)` | 高亮菜单路径 |
| `breadcrumb_mode` | `varchar(20)` | `inherit_menu` / `inherit_page` / `custom` |
| `access_mode` | `varchar(20)` | `inherit` / `public` / `jwt` / `permission` |
| `permission_key` | `varchar(150)` | 页面独立权限键 |
| `inherit_permission` | `bool` | 是否继承上级权限 |
| `keep_alive` | `bool` | 是否缓存 |
| `is_full_page` | `bool` | 是否全屏 |
| `status` | `varchar(20)` | `normal` / `suspended` |
| `meta` | `jsonb` | 扩展信息 |
| `created_at` | `timestamp` | 创建时间 |
| `updated_at` | `timestamp` | 更新时间 |
| `deleted_at` | `timestamp` | 软删时间 |

索引建议：
- `unique index idx_ui_pages_page_key`
- `unique index idx_ui_pages_route_name`
- `index idx_ui_pages_parent_menu_id`
- `index idx_ui_pages_module_key`
- `index idx_ui_pages_page_type_status`
- `index idx_ui_pages_access_mode`

约束建议：
- `page_key` 非空
- `route_name` 非空
- `route_path` 非空
- `page_type` 非空
- `source` 非空
- `status` 非空

### 3.3 推荐 GORM 模型

建议放在：
- `backend/internal/modules/system/models/model.go`

结构草案：

```go
type UIPage struct {
    ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    PageKey           string         `gorm:"type:varchar(150);not null;uniqueIndex" json:"page_key"`
    Name              string         `gorm:"type:varchar(150);not null" json:"name"`
    RouteName         string         `gorm:"type:varchar(150);not null;uniqueIndex" json:"route_name"`
    RoutePath         string         `gorm:"type:varchar(255);not null" json:"route_path"`
    Component         string         `gorm:"type:varchar(255);not null" json:"component"`
    PageType          string         `gorm:"type:varchar(20);not null;default:'inner'" json:"page_type"`
    Source            string         `gorm:"type:varchar(20);not null;default:'manual'" json:"source"`
    ModuleKey         string         `gorm:"type:varchar(100);not null;default:''" json:"module_key"`
    ParentMenuID      *uuid.UUID     `gorm:"type:uuid;index" json:"parent_menu_id"`
    ParentPageKey     string         `gorm:"type:varchar(150);default:''" json:"parent_page_key"`
    ActiveMenuPath    string         `gorm:"type:varchar(255);default:''" json:"active_menu_path"`
    BreadcrumbMode    string         `gorm:"type:varchar(20);not null;default:'inherit_menu'" json:"breadcrumb_mode"`
    AccessMode        string         `gorm:"type:varchar(20);not null;default:'inherit'" json:"access_mode"`
    PermissionKey     string         `gorm:"type:varchar(150);default:''" json:"permission_key"`
    InheritPermission bool           `gorm:"not null;default:true" json:"inherit_permission"`
    KeepAlive         bool           `gorm:"not null;default:false" json:"keep_alive"`
    IsFullPage        bool           `gorm:"not null;default:false" json:"is_full_page"`
    Status            string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
    Meta              MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
    CreatedAt         time.Time      `json:"created_at"`
    UpdatedAt         time.Time      `json:"updated_at"`
    DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

## 4. 后端模块设计

### 4.1 模块命名

建议新增模块：
- `backend/internal/modules/system/page`

目录建议：
- `handler.go`
- `service.go`
- `module.go`

保持和 `apiendpoint`、`featurepackage`、`menu` 一致。

### 4.2 DTO 结构建议

列表请求：

```go
type ListRequest struct {
    Current         int    `form:"current"`
    Size            int    `form:"size"`
    Keyword         string `form:"keyword"`
    PageType        string `form:"page_type"`
    ModuleKey       string `form:"module_key"`
    ParentMenuID    string `form:"parent_menu_id"`
    AccessMode      string `form:"access_mode"`
    Source          string `form:"source"`
    Status          string `form:"status"`
}
```

创建/更新请求：

```go
type SaveRequest struct {
    PageKey           string                 `json:"page_key"`
    Name              string                 `json:"name"`
    RouteName         string                 `json:"route_name"`
    RoutePath         string                 `json:"route_path"`
    Component         string                 `json:"component"`
    PageType          string                 `json:"page_type"`
    Source            string                 `json:"source"`
    ModuleKey         string                 `json:"module_key"`
    ParentMenuID      string                 `json:"parent_menu_id"`
    ParentPageKey     string                 `json:"parent_page_key"`
    ActiveMenuPath    string                 `json:"active_menu_path"`
    BreadcrumbMode    string                 `json:"breadcrumb_mode"`
    AccessMode        string                 `json:"access_mode"`
    PermissionKey     string                 `json:"permission_key"`
    InheritPermission bool                   `json:"inherit_permission"`
    KeepAlive         bool                   `json:"keep_alive"`
    IsFullPage        bool                   `json:"is_full_page"`
    Status            string                 `json:"status"`
    Meta              map[string]interface{} `json:"meta"`
}
```

### 4.3 接口草案

建议接口：

- `GET /api/v1/pages`
- `GET /api/v1/pages/:id`
- `POST /api/v1/pages`
- `PUT /api/v1/pages/:id`
- `DELETE /api/v1/pages/:id`
- `POST /api/v1/pages/sync`
- `GET /api/v1/pages/unregistered`
- `POST /api/v1/pages/:id/preview-breadcrumb`

推荐权限建议：
- 页面管理页整体能力：`system.page.manage`
- 页面同步：`system.page.sync`

### 4.4 校验规则

创建/更新时应校验：
- `page_key` 唯一
- `route_name` 唯一
- `route_path` 非空
- `page_type` 合法
- `access_mode` 合法
- `breadcrumb_mode` 合法
- `inner` 页面必须至少满足一项：
  - `parent_menu_id` 非空
  - `parent_page_key` 非空
- 若 `access_mode = permission`，则 `permission_key` 必须非空

## 5. 前端路由接入方案

### 5.1 当前现状

当前前端主访问控制基于：
- `MenuProcessor`
- `RouteRegistry`
- `RoutePermissionValidator`
- `beforeEach.ts`

当前问题：
- `RoutePermissionValidator` 只按菜单路径集合校验
- 不识别“页面管理中的 inner/global 页面”

### 5.2 第一阶段改造原则

不推翻当前菜单路由主链，只在守卫中新增一层页面判定：

建议顺序：
1. `public` 静态路由直接放行
2. 登录校验
3. 菜单动态路由初始化
4. 先判定页面管理
5. 再判定旧菜单路径权限
6. 再判定旧 `inner` 菜单兜底

### 5.3 页面管理前端清单

建议新增：
- `frontend/src/router/pageManifest.ts`

导出结构建议：

```ts
export interface ManagedPageManifestItem {
  pageKey: string
  name: string
  routeName: string
  routePath: string
  component: string
  moduleKey?: string
  pageType: 'inner' | 'global'
  defaultAccessMode?: 'inherit' | 'public' | 'jwt' | 'permission'
}
```

用途：
- 页面同步
- 页面未注册检查
- 路由守卫辅助映射

### 5.4 路由 meta 建议

受管理页面统一补充：

```ts
meta: {
  managedPage: true,
  pageKey: 'order.detail',
  moduleKey: 'order'
}
```

### 5.5 守卫判定流程

建议新增一个 `PageAccessResolver`，放在：
- `frontend/src/router/core/PageAccessResolver.ts`

职责：
- 根据 `pageKey` 读取页面管理配置
- 解析 `accessMode`
- 生成：
  - 是否允许访问
  - 上级菜单高亮
  - 面包屑来源
  - fallback 路径

建议结果结构：

```ts
type ResolveResult = {
  allowed: boolean
  accessMode: 'public' | 'jwt' | 'permission' | 'inherit'
  activeMenuPath?: string
  breadcrumbSource?: 'menu' | 'page' | 'custom'
  fallbackPath?: string
}
```

### 5.6 RoutePermissionValidator 改造方向

当前它只消费菜单路径集合。

第一阶段不直接重写，建议新增上层组合器：
- `ManagedRoutePermissionValidator`

判定顺序：
1. 若不是 `managedPage`，走当前 `RoutePermissionValidator`
2. 若命中页面管理：
  - `public` 直接放行
  - `jwt` 验 token 后放行
  - `permission` 验 `permissionKey`
  - `inherit` 向上解析
3. 若页面管理未命中，则回退旧菜单链

## 6. 面包屑与高亮实现草案

### 6.1 前端 store 建议

建议页面管理接入以下 store：
- `menuStore`
- `worktabStore`

新增可选 store：
- `pageStore`

`pageStore` 可缓存：
- 页面配置
- pageKey -> parentMenuPath 映射
- pageKey -> breadcrumb 配置映射

### 6.2 页面进入后的表现

当进入 `inner` 页面：
- 左侧菜单高亮：`active_menu_path` 或 `parent_menu_id` 对应菜单路径
- 面包屑：
  - 上级菜单链
  - 当前页面名
- 标签页标题：
  - 页面名称优先
  - 缺失时回退路由标题

### 6.3 第一阶段只做最小闭环

第一阶段只要求：
- 左侧菜单高亮正确
- 面包屑继承菜单正确
- 访问判定正确

不要求：
- 自定义 breadcrumb 编辑器
- 多级页面继承链编排器

## 7. 迁移策略

### 7.1 菜单部分

不做菜单数据迁移。

只新增运行时解释规则：
- `meta.accessMode = jwt` 时视为全局菜单

### 7.2 页面部分

新增空表迁移即可：
- `create_ui_pages_table`

第一阶段不自动迁移旧 `inner` 菜单数据。

原因：
- 先验证页面管理链路
- 避免旧 `inner` 数据语义不整齐时一次性迁移出错

### 7.3 可选 seed

可考虑增加少量 seed 页面：
- 个人中心
- 消息中心
- 全局搜索页

但第一阶段不是必须。

## 8. 后台页面设计草案

建议新增前端页面：
- `frontend/src/views/system/page/index.vue`

模块建议：
- `page-search.vue`
- `page-dialog.vue`
- `page-breadcrumb-preview-dialog.vue`

表单字段建议：
- 页面名称
- 页面标识
- 页面类型
- 路由名称
- 路由路径
- 组件路径
- 所属模块
- 上级菜单
- 上级页面
- 面包屑模式
- 访问模式
- 权限键
- 是否继承权限
- 是否缓存
- 是否全屏
- 状态

## 9. 推荐拆任务顺序

### Task 1
- 后端新增 `UIPage` 模型
- 迁移创建 `ui_pages`

### Task 2
- 后端新增页面管理模块：
  - 列表
  - 新增
  - 编辑
  - 删除
  - 详情

### Task 3
- 菜单管理补 `meta.accessMode`
- 前端菜单编辑页支持全局菜单访问模式

### Task 4
- 前端新增页面管理页

### Task 5
- 前端路由守卫新增页面管理判定链

### Task 6
- 页面同步与未注册页面检查

## 10. 验证清单

第一阶段至少验证：

- 菜单页：
  - 普通权限菜单访问正常
  - 全局菜单登录后可访问
- 页面管理：
  - `inner` 页面可绑定上级菜单
  - 页面列表和编辑正常
- 路由守卫：
  - `jwt` 页面无需权限即可访问
  - `permission` 页面按权限校验
  - `inherit` 页面能继承上级菜单或模块权限
- 兼容：
  - 旧 `inner` 菜单不受影响

## 11. 当前不建议做的事

- 不要第一阶段就做旧 `inner` 菜单批量迁移
- 不要第一阶段就让页面管理接管全部菜单页
- 不要把页面同步设计成扫描运行中前端包产物
- 不要把“全局菜单”再做成第二套路由菜单系统

## 12. 实施结论

第一阶段最稳的路线是：
- 菜单保留不动，只补 `accessMode`
- 页面管理新增但不抢现有菜单职责
- 前端守卫用“页面管理优先、旧 inner 兼容回退”的方式平滑接入

这样做可以在风险可控的前提下，把未来“一个模块大量数据内页”的管理复杂度真正降下来。
