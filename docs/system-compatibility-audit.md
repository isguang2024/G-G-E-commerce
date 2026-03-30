# 系统兼容与收紧审计

> 更新时间：2026-03-30。目标是把系统侧现有兼容分支分成“正式保留”“已收紧”“待确认清理”三类，避免继续无边界膨胀。

## 1. 正式保留的兼容

### 1.1 菜单空间默认回退

- 位置：
  - [frontend/src/store/modules/menu-space.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/store/modules/menu-space.ts)
  - [backend/internal/modules/system/space/service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/space/service.go)
- 原因：
  - 当前默认运行模式仍是单域单菜单。
  - 未命中 Host、无额外空间配置或无权进入目标空间时，必须静默回退到 `default`。
- 保留结论：正式兼容，继续保留。

### 1.2 运行时页面注册表回退

- 位置：
  - [frontend/src/router/guards/beforeEach.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/guards/beforeEach.ts)
  - [frontend/src/router/core/ManagedPageProcessor.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/router/core/ManagedPageProcessor.ts)
  - [backend/internal/modules/system/page/runtime_cache.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/page/runtime_cache.go)
- 原因：
  - 页面中心与菜单主页面仍有并存模型。
  - 当前已经改成“服务端先按当前用户、团队、空间、访问模式裁剪 runtime pages，前端再做防御性过滤”。
  - 动态页面注册失败时，仍需保留一次收敛与重建机会，避免直接掉 404。
- 保留结论：正式兼容，继续保留。

### 1.3 面包屑父链兜底

- 位置：
  - [frontend/src/components/core/layouts/art-breadcrumb/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-breadcrumb/index.vue)
- 原因：
  - 内页可能只携带当前页元数据，缺失运行时面包屑链。
  - 当前仍需按 `activePath / customParent` 从菜单树补全父链。
- 保留结论：正式兼容，继续保留。

## 2. 已收紧为静默降级或开发期日志

### 2.1 菜单分组接口失败

- 位置：
  - [frontend/src/views/system/menu/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/menu/index.vue)
- 当前行为：
  - 不再弹业务式 warning。
  - 页面内提示“菜单分组暂不可用，当前按普通菜单树显示”。
  - 控制台仅保留开发期 `warnDev` 日志。
- 结论：已收紧。

### 2.2 菜单空间候选路径加载失败

- 位置：
  - [frontend/src/views/system/menu-space/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/menu-space/index.vue)
- 当前行为：
  - 默认首页候选加载失败时，回退为空列表。
  - 控制台仅在开发期输出 `warnDev`。
- 结论：已收紧。

### 2.3 菜单空间、菜单、页面、消息工作台首屏加载失败

- 位置：
  - [frontend/src/views/system/menu/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/menu/index.vue)
  - [frontend/src/views/system/page/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/page/index.vue)
  - [frontend/src/views/system/menu-space/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/menu-space/index.vue)
  - [frontend/src/views/message/modules/message-dispatch-console.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-dispatch-console.vue)
  - [frontend/src/views/message/modules/message-record-console.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-record-console.vue)
  - [frontend/src/views/message/modules/message-template-console.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-template-console.vue)
  - [frontend/src/views/message/modules/message-sender-console.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-sender-console.vue)
  - [frontend/src/views/message/modules/message-recipient-group-console.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/message/modules/message-recipient-group-console.vue)
  - [frontend/src/views/workspace/inbox/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/workspace/inbox/index.vue)
- 当前行为：
  - 首屏失败优先改成页内 `ElAlert` 或空态。
  - 用户主动动作失败仍保留 toast。
- 结论：已收紧。

### 2.4 快捷入口与工作标签噪音日志

- 位置：
  - [frontend/src/store/modules/worktab.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/store/modules/worktab.ts)
  - [frontend/src/store/modules/menu-space.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/store/modules/menu-space.ts)
  - [frontend/src/components/core/layouts/art-fast-enter/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/core/layouts/art-fast-enter/index.vue)
- 当前行为：
  - 非关键 fallback 只保留开发期日志。
  - 默认运行期不再持续刷 `console.warn`。
- 结论：已收紧。

## 3. 待后续确认清理

### 3.1 `frontend-copy`

- 路径：
  - [frontend-copy](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-copy)
- 当前判断：
  - 代码检索未发现正式引用。
  - 更像历史副本或人工备份目录。
- 当前处理：
  - 本轮不直接删除。
  - 先纳入收尾确认项，待确认无人工备份价值后单独清理。

### 3.2 部分运行时兼容层

- 可能涉及：
  - 动态路由与页面注册重叠兼容
  - 页面中心与菜单主页面并存解释层
- 当前处理：
  - 本轮先保留，不做猜测式重构。
  - 等真实业务模块进入后，再看是否有继续压缩价值。

## 4. 当前结论

- 系统侧兼容已经不是“看到问题就先加兜底”，而是开始分层管理。
- 后续新增兼容逻辑时，必须明确它属于：
  - 正式保留兼容
  - 临时静默降级
  - 待后续删除的历史兜底
- 若无法归类，不应直接引入新的兼容分支。
