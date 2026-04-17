# 自封装组件 & Hooks

本仓 `business/` 下沉淀的跨页面组件，以及配套 Hooks。所有组件均需**显式 `import`**。

新增组件后必须在本文件对应表格加一行，否则其他页面不知道它存在。

---

## business/common — 通用小件

| 组件 | 路径 | 主要 props | 用途 |
| --- | --- | --- | --- |
| `FieldLabel` | `business/common/FieldLabel.vue` | `label` / `help?` / `required?` / `rawContent?` | 表单字段标签 + 问号 tooltip。替代裸 `<span>`，把说明文字塞进 tooltip 让页面更干净。`help` 支持 HTML（配合 `rawContent=true`）。 |

---

## business/dictionary — 字典下拉

| 组件 | 路径 | 主要 props | 用途 |
| --- | --- | --- | --- |
| `DictSelect` | `business/dictionary/DictSelect.vue` | `code`（必填）/ `multiple` / `allowCreate` / `fallbackOptions` / `autoSelectDefault` / `clearable` | **所有枚举下拉统一用这个**，不要手写 `<ElOption>` 列表。内置 filterable、模块级缓存、默认项排序。后端新增字典项 → 下次打开自动生效，无需改前端代码。 |

---

## business/app

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `AppKeySelect` | `business/app/AppKeySelect.vue` | 按 app_key 选择应用的下拉 |

---

## business/collaboration-workspace

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `NoCollaborationWorkspaceState` | `business/collaboration-workspace/NoCollaborationWorkspaceState.vue` | 无协作工作空间时的引导占位态 |

---

## business/layout

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `AdminWorkspaceHero` | `business/layout/AdminWorkspaceHero.vue` | 管理端工作区顶部 Hero（标题 / 统计 / 操作区） |
| `AppContextBadge` | `business/layout/AppContextBadge.vue` | 当前 app 上下文徽标 |

---

## business/permission — 权限工作台

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `PermissionActionWorkbench` | `business/permission/PermissionActionWorkbench.vue` | 动作权限主工作台（角色/套餐授权主画布） |
| `PermissionActionCascaderPanel` | `business/permission/PermissionActionCascaderPanel.vue` | 模块 → 资源 → 动作级联选择 |
| `PermissionSourcePanels` | `business/permission/PermissionSourcePanels.vue` | 权限来源（角色 / 套餐 / 直授）并列展示 |
| `PermissionSummaryTags` | `business/permission/PermissionSummaryTags.vue` | 权限摘要标签云 |
| `ActionPermissionTreePanel` | `business/permission/action-permission-tree-panel.vue` | 动作权限树视图 |
| `FeaturePackageGrantPreview` | `business/permission/FeaturePackageGrantPreview.vue` | 套餐授权预览 |

---

## business/tables

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `WorkspacePagination` | `business/tables/WorkspacePagination.vue` | ⚠️ **遗留，停止新增使用**。分页改用 `ArtTable` 内置 `:pagination` prop。 |

---

## Observability（技术向）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `JsonViewer` | `Observability/JsonViewer.vue` | 结构化 JSON 查看器（日志、trace、事件详情） |
| `TraceDrawer` | `Observability/TraceDrawer.vue` | 追踪抽屉 |

---

## Hooks

### 基座提供（Art Design Pro）

| Hook | 说明 |
| --- | --- |
| `useTable` | 表格数据、分页、筛选、排序的标准封装，与 `ArtTable` 搭配使用 |

### 本仓沉淀

| Hook | 路径 | 说明 |
| --- | --- | --- |
| `useDictionary(code)` | `hooks/business/useDictionary.ts` | 单个字典响应式：`options` / `loading` / `map` / `defaultValue` |
| `useDictionaries(codes[])` | 同上 | 批量字典，0ms 微任务去重合并请求 + 模块级缓存 |
| `invalidateDict(code)` / `invalidateAllDicts()` | 同上 | 字典变更后强制失效缓存 |
| `useUpload` | `domains/upload/use-upload.ts` | 上传 prepare → upload → complete 全流程封装（支持 relay / direct、失败重试） |
