# 业务模块开发说明

## 定位

`backend/internal/modules` 是项目的业务核心层。

这套项目不是只服务单一后台，而是作为未来多个业务项目可复用的基础脚手架。后续新增商品、订单、内容、素材等模块时，都应直接接入当前这套模块化结构、菜单体系、功能权限体系和租户上下文体系。

## 当前模块结构

- `system/`
  系统基础模块。负责认证、用户、角色、菜单、作用域、功能权限、API 注册表、租户与团队上下文。

未来新增业务模块时，直接在 `modules/` 下增加目录，例如：

- `product/`
- `order/`
- `content/`

## 新业务模块接入最小清单

新增一个业务模块时，至少处理以下事项：

1. 新增模块目录和标准文件：
   - `module.go`
   - `handler.go`
   - `service.go`
   - `repository.go`
2. 如需新表，先接入 `database.AutoMigrate`
3. 给接口注册明确的 API 元数据：
   - `resource_code`
   - `action_code`
   - `scope`
   - `feature_kind`
   - `module`
4. 菜单入口按角色控制，按钮和接口按功能权限控制
5. 如功能同时支持平台和团队两种上下文：
   - 保留两套薄路由入口
   - 共用同一套 service
   - 权限按不同作用域分别注册
6. 如模块涉及团队数据，业务表必须带 `tenant_id`

## 推荐开发模式

### 1. 入口分层

- 平台入口：
  面向系统管理员或平台角色
- 团队入口：
  面向当前团队上下文

推荐保留两套路由入口，但不要复制两套业务实现。两个 handler 进入同一套 service，由 service 接收统一的访问上下文。

### 2. 权限分层

- 菜单权限：
  决定入口是否显示
- 功能权限：
  决定按钮与接口是否允许执行
- 数据权限：
  决定数据范围

### 3. 文档位置

以后所有系统基础能力相关说明，不再堆在根目录。

说明文档统一写在业务目录下：

- `backend/internal/modules/system/README.md`
- `backend/internal/modules/system/permission/README.md`
- `backend/internal/modules/system/tenant/README.md`

## 重要约束

- 不要把“前台/后台”当成权限边界，统一按“入口上下文 + 功能能力 + 数据范围”建模
- 不要让一个接口同时混跑多个作用域语义
- 同一业务动作如果要支持多个作用域，权限定义必须拆成多条，按 `resource + action + scope` 唯一
- 菜单权限保留在角色层，用户层默认只做功能权限覆盖
