# Art Design Pro — core/ 组件目录

共 **54 个**组件，全部通过 `utils/registerGlobalComponent.ts` **全局自动注册**，页面直接使用，无需 `import`。

props / events / slots 详情直接看对应 `index.vue` 源文件，IDE 有完整类型提示。
不要修改 `core/` 内组件的 API 形状；需扩展时在 `business/` 里包一层。

---

## banners（2 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtBasicBanner` | `core/banners/art-basic-banner` | 纯文字提示横幅 |
| `ArtCardBanner` | `core/banners/art-card-banner` | 带卡片样式的引导横幅 |

---

## base（3 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtBackToTop` | `core/base/art-back-to-top` | 回到顶部按钮 |
| `ArtLogo` | `core/base/art-logo` | 品牌 Logo |
| `ArtSvgIcon` | `core/base/art-svg-icon` | SVG 图标渲染 |

---

## cards（8 个）

仪表盘卡片，已内置骨架屏、响应式、空态。**优先用 cards 而非直接用 charts 原件。**

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtStatsCard` | `core/cards/art-stats-card` | 数字统计卡（标题 + 大数字 + 趋势） |
| `ArtBarChartCard` | `core/cards/art-bar-chart-card` | 柱状图卡片 |
| `ArtLineChartCard` | `core/cards/art-line-chart-card` | 折线图卡片 |
| `ArtDonutChartCard` | `core/cards/art-donut-chart-card` | 环形图卡片 |
| `ArtImageCard` | `core/cards/art-image-card` | 图片展示卡片 |
| `ArtProgressCard` | `core/cards/art-progress-card` | 进度条卡片 |
| `ArtDataListCard` | `core/cards/art-data-list-card` | 数据列表卡片 |
| `ArtTimelineListCard` | `core/cards/art-timeline-list-card` | 时间线列表卡片 |

---

## charts（8 个）

图表原件，被 `cards/` 层复用。**需要自定义容器时才直接使用。**

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtBarChart` | `core/charts/art-bar-chart` | 柱状图 |
| `ArtDualBarCompareChart` | `core/charts/art-dual-bar-compare-chart` | 双柱对比图 |
| `ArtHBarChart` | `core/charts/art-h-bar-chart` | 横向柱状图 |
| `ArtKLineChart` | `core/charts/art-k-line-chart` | K 线图 |
| `ArtLineChart` | `core/charts/art-line-chart` | 折线图 |
| `ArtRadarChart` | `core/charts/art-radar-chart` | 雷达图 |
| `ArtRingChart` | `core/charts/art-ring-chart` | 环形图 |
| `ArtScatterChart` | `core/charts/art-scatter-chart` | 散点图 |

---

## forms（8 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtButtonMore` | `core/forms/art-button-more` | "更多"下拉按钮 |
| `ArtButtonTable` | `core/forms/art-button-table` | 表格行操作按钮组 |
| `ArtDragVerify` | `core/forms/art-drag-verify` | 拖拽滑块验证 |
| `ArtExcelExport` | `core/forms/art-excel-export` | Excel 导出按钮 |
| `ArtExcelImport` | `core/forms/art-excel-import` | Excel 导入按钮 |
| `ArtForm` | `core/forms/art-form` | 通用表单容器 |
| `ArtSearchBar` | `core/forms/art-search-bar` | 搜索栏（字段 + 展开/收起） |
| `ArtWangEditor` | `core/forms/art-wang-editor` | 富文本编辑器（WangEditor 封装） |

---

## layouts（15 个）

系统外壳组件，**业务页面不要直接引用**，由路由布局层自动加载。

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtBreadcrumb` | `core/layouts/art-breadcrumb` | 面包屑导航 |
| `ArtChatWindow` | `core/layouts/art-chat-window` | 聊天窗口浮层 |
| `ArtFastEnter` | `core/layouts/art-fast-enter` | 快捷入口面板 |
| `ArtFireworksEffect` | `core/layouts/art-fireworks-effect` | 烟花特效（节日装饰） |
| `ArtGlobalComponent` | `core/layouts/art-global-component` | 全局组件挂载容器 |
| `ArtGlobalSearch` | `core/layouts/art-global-search` | 全局搜索弹窗 |
| `ArtHeaderBar` | `core/layouts/art-header-bar` | 顶部导航栏 |
| `ArtHorizontalMenu` | `core/layouts/art-menus/art-horizontal-menu` | 顶部水平菜单 |
| `ArtMixedMenu` | `core/layouts/art-menus/art-mixed-menu` | 混合布局菜单 |
| `ArtSidebarMenu` | `core/layouts/art-menus/art-sidebar-menu` | 侧栏菜单 |
| `ArtNotification` | `core/layouts/art-notification` | 通知中心弹层 |
| `ArtPageContent` | `core/layouts/art-page-content` | 页面内容区容器 |
| `ArtScreenLock` | `core/layouts/art-screen-lock` | 锁屏遮罩 |
| `ArtSettingsPanel` | `core/layouts/art-settings-panel` | 主题/布局设置抽屉 |
| `ArtWorkTab` | `core/layouts/art-work-tab` | 多标签页工作区 |

---

## media（2 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtCutterImg` | `core/media/art-cutter-img` | 图片裁剪 |
| `ArtVideoPlayer` | `core/media/art-video-player` | 视频播放器（xgplayer 封装） |

---

## others（2 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtMenuRight` | `core/others/art-menu-right` | 右键菜单 |
| `ArtWatermark` | `core/others/art-watermark` | 页面水印 |

---

## tables（2 个）

**新表格页统一用这两个组件，不要自造 wrapper。**

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtTable` | `core/tables/art-table` | 统一表格壳：loading、列配置、操作列、空状态、**内置分页** |
| `ArtTableHeader` | `core/tables/art-table-header` | 表格头：标题 / 副标题 / 操作区 |

### ArtTable 分页用法

项目**所有列表页**的分页统一用 `ArtTable` 内置 `:pagination` prop，不引入 `ElPagination` 或 `WorkspacePagination`。

```vue
<ArtTable
  :data="data"
  :loading="loading"
  :pagination="pagination"
  @pagination:size-change="handleSizeChange"
  @pagination:current-change="handleCurrentChange"
>
  <ElTableColumn prop="name" label="名称" />
</ArtTable>
```

| prop / event | 类型 | 说明 |
| --- | --- | --- |
| `:pagination` | `{ current: number; size: number; total: number }` | 三字段必填，`total` 驱动分页器 |
| `@pagination:size-change` | `(val: number) => void` | 每页条数变更，通常重置 `current = 1` |
| `@pagination:current-change` | `(val: number) => void` | 页码变更 |
| `:show-table-header` | `boolean`（默认 `true`） | 无 `ArtTableHeader` 时设 `false` |
| `:pagination-options` | `PaginationOptions` | 可选扩展：`pageSizes / align / layout / background / size` 等 |

**服务端分页样板：**

```ts
const pagination = reactive({ current: 1, size: 20, total: 0 })

function handleSizeChange(val: number) {
  pagination.size = val
  pagination.current = 1
  loadData()
}
function handleCurrentChange(val: number) {
  pagination.current = val
  loadData()
}
```

**客户端分页样板（本地过滤后分页）：**

```ts
const page = reactive({ current: 1, size: 20 })
const filtered  = computed(() => rawList.value.filter(...))
const paged     = computed(() => filtered.value.slice((page.current - 1) * page.size, page.current * page.size))
const pagination = computed(() => ({ current: page.current, size: page.size, total: filtered.value.length }))

function handleSizeChange(val: number) { page.size = val; page.current = 1 }
function handleCurrentChange(val: number) { page.current = val }
```

> 若使用了 `useTable` hook，分页状态由 hook 内部维护，直接透传 `pagination` 对象即可。

---

## text-effect（3 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtCountTo` | `core/text-effect/art-count-to` | 数字滚动动效 |
| `ArtFestivalTextScroll` | `core/text-effect/art-festival-text-scroll` | 节日文字滚动 |
| `ArtTextScroll` | `core/text-effect/art-text-scroll` | 通用文字滚动跑马灯 |

---

## theme（1 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ThemeSvg` | `core/theme/theme-svg` | 主题色 SVG 渲染 |

---

## widget（1 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtIconButton` | `core/widget/art-icon-button` | 图标按钮 |

---

## runtime（1 个）

| 组件 | 路径 | 用途 |
| --- | --- | --- |
| `ArtAppErrorBoundary` | `core/runtime/ArtAppErrorBoundary` | 顶层应用错误边界 |
