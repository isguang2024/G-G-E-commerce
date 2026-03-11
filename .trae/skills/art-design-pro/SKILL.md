---
name: art-design-pro
description: "Art Design Pro 组件库智能助手。帮助你在编写 Vue3 页面时准确使用项目已有的 Art Design Pro 组件，避免重复造轮子。支持 65+ 组件，涵盖表格、表单、图表、布局等场景。"
---

# Art Design Pro 组件库智能助手

让 Claude Code 在编写 Vue3 页面时准确使用项目中已有的 Art Design Pro 组件，避免重复造轮子。

## 核心原则

**优先使用现有组件，而不是从零编写 Vue 代码。**

## 何时使用

在以下场景中，**必须**先使用本 skill 查找可用组件：

1. ✅ 创建新的页面或视图
2. ✅ 添加表格、表单、图表等 UI 元素
3. ✅ 实现搜索、筛选、导出等功能
4. ✅ 添加布局组件（面包屑、页头、标签页等）
5. ✅ 需要数据可视化（图表、统计卡片等）

## 组件分类速查

### 📊 **表格与表单**
| 组件 | 用途 | 导入路径 |
|------|------|---------|
| `ArtTable` | 数据表格（支持分页、排序、自定义列） | `@/components/core/tables/art-table` |
| `ArtTableHeader` | 表格头部工具栏 | `@/components/core/tables/art-table-header` |
| `ArtForm` | 表单（支持响应式、校验、各种表单项） | `@/components/core/forms/art-form` |
| `ArtSearchBar` | 搜索栏 | `@/components/core/forms/art-search-bar` |
| `ArtButtonTable` | 表格操作按钮组 | `@/components/core/forms/art-button-table` |
| `ArtButtonMore` | 更多操作下拉按钮 | `@/components/core/forms/art-button-more` |
| `ArtDragVerify` | 拖拽验证 | `@/components/core/forms/art-drag-verify` |

### 📈 **图表与数据展示**
| 组件 | 用途 | 导入路径 |
|------|------|---------|
| `ArtStatsCard` | 统计卡片 | `@/components/core/cards/art-stats-card` |
| `ArtBarChartCard` | 柱状图卡片 | `@/components/core/cards/art-bar-chart-card` |
| `ArtLineChartCard` | 折线图卡片 | `@/components/core/cards/art-line-chart-card` |
| `ArtDonutChartCard` | 环形图卡片 | `@/components/core/cards/art-donut-chart-card` |
| `ArtProgressCard` | 进度卡片 | `@/components/core/cards/art-progress-card` |
| `ArtDataListCard` | 数据列表卡片 | `@/components/core/cards/art-data-list-card` |
| `ArtTimelineListCard` | 时间轴列表卡片 | `@/components/core/cards/art-timeline-list-card` |
| `ArtImageCard` | 图片卡片 | `@/components/core/cards/art-image-card` |
| `ArtBarChart` | 柱状图 | `@/components/core/charts/art-bar-chart` |
| `ArtLineChart` | 折线图 | `@/components/core/charts/art-line-chart` |
| `ArtRingChart` | 环形图 | `@/components/core/charts/art-ring-chart` |
| `ArtRadarChart` | 雷达图 | `@/components/core/charts/art-radar-chart` |
| `ArtMapChart` | 地图图表 | `@/components/core/charts/art-map-chart` |

### 🎨 **布局与导航**
| 组件 | 用途 | 导入路径 |
|------|------|---------|
| `ArtPageContent` | 页面内容容器 | `@/components/core/layouts/art-page-content` |
| `ArtBreadcrumb` | 面包屑导航 | `@/components/core/layouts/art-breadcrumb` |
| `ArtHeaderBar` | 页头工具栏 | `@/components/core/layouts/art-header-bar` |
| `ArtWorkTab` | 多标签页 | `@/components/core/layouts/art-work-tab` |
| `ArtSidebarMenu` | 侧边栏菜单 | `@/components/core/layouts/art-menus/art-sidebar-menu` |
| `ArtHorizontalMenu` | 水平菜单 | `@/components/core/layouts/art-menus/art-horizontal-menu` |
| `ArtFastEnter` | 快捷入口 | `@/components/core/layouts/art-fast-enter` |

### 🔧 **工具类组件**
| 组件 | 用途 | 导入路径 |
|------|------|---------|
| `ArtExcelExport` | Excel 导出 | `@/components/core/forms/art-excel-export` |
| `ArtExcelImport` | Excel 导入 | `@/components/core/forms/art-excel-import` |
| `ArtWangEditor` | 富文本编辑器 | `@/components/core/forms/art-wang-editor` |
| `ArtVideoPlayer` | 视频播放器 | `@/components/core/media/art-video-player` |
| `ArtCutterImg` | 图片裁剪 | `@/components/core/media/art-cutter-img` |
| `ArtNotification` | 通知中心 | `@/components/core/layouts/art-notification` |

### 🎯 **基础组件**
| 组件 | 用途 | 导入路径 |
|------|------|---------|
| `ArtLogo` | 系统logo | `@/components/core/base/art-logo` |
| `ArtSvgIcon` | SVG 图标 | `@/components/core/base/art-svg-icon` |
| `ArtBackToTop` | 返回顶部 | `@/components/core/base/art-back-to-top` |
| `ArtIconButton` | 图标按钮 | `@/components/core/widget/art-icon-button` |

## 使用方法

### 方法 1：搜索组件（推荐）

当需要添加某个功能时，使用 Python 脚本搜索相关组件：

```bash
# 搜索关键词（支持中文和英文）
python3 .claude/skills/art-design-pro/scripts/search.py "表格"

# 搜索表单相关组件
python3 .claude/skills/art-design-pro/scripts/search.py "form"

# 搜索图表
python3 .claude/skills/art-design-pro/scripts/search.py "chart"

# 搜索特定分类
python3 .claude/skills/art-design-pro/scripts/search.py "table" --category tables
```

### 方法 2：直接查阅组件文档

查看完整组件列表：

```bash
python3 .claude/skills/art-design-pro/scripts/list.py
```

## 常见场景组件选择

### 场景 1：创建 CRUD 列表页

**必选组件：**
- `ArtTable` - 数据表格
- `ArtSearchBar` - 搜索栏
- `ArtTableHeader` - 表格工具栏（新增、批量删除等）
- `ArtPageContent` - 页面容器

**可选组件：**
- `ArtExcelExport` - 数据导出
- `ArtButtonTable` - 表格行操作按钮

### 场景 2：创建表单页

**必选组件：**
- `ArtForm` - 表单组件

**可选组件：**
- `ArtDragVerify` - 滑块验证
- `ArtWangEditor` - 富文本编辑

### 场景 3：创建仪表板/统计页

**必选组件：**
- `ArtStatsCard` - 统计卡片
- `ArtLineChartCard` - 趋势图
- `ArtBarChart` - 对比图

**可选组件：**
- `ArtProgressCard` - 进度展示
- `ArtDataListCard` - 数据列表

### 场景 4：创建详情页

**必选组件：**
- `ArtPageContent` - 页面容器
- `ArtBreadcrumb` - 面包屑

**可选组件：**
- `ArtTimelineListCard` - 时间轴
- `ArtImageCard` - 图片展示

## 组件使用示例

### ArtTable - 数据表格

```vue
<template>
  <art-table
    :data="tableData"
    :columns="columns"
    :pagination="pagination"
    :loading="loading"
    @pagination:current-change="handlePageChange"
  >
    <!-- 自定义列 -->
    <template #status="{ row }">
      <el-tag :type="row.status === 1 ? 'success' : 'danger'">
        {{ row.status === 1 ? '启用' : '禁用' }}
      </el-tag>
    </template>

    <!-- 操作列 -->
    <template #action="{ row }">
      <el-button link @click="handleEdit(row)">编辑</el-button>
      <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
    </template>
  </art-table>
</template>

<script setup lang="ts">
import ArtTable from '@/components/core/tables/art-table/index.vue'

const columns = [
  { prop: 'name', label: '名称' },
  { prop: 'status', label: '状态', useSlot: true },
  { prop: 'action', label: '操作', useSlot: true, width: 200 }
]

const tableData = ref([])
const pagination = reactive({ current: 1, size: 10, total: 0 })
</script>
```

### ArtForm - 表单

```vue
<template>
  <art-form
    v-model="formData"
    :items="formItems"
    @submit="handleSubmit"
    @reset="handleReset"
  />
</template>

<script setup lang="ts">
import ArtForm from '@/components/core/forms/art-form/index.vue'

const formData = ref({})

const formItems = [
  { key: 'name', label: '名称', type: 'input', span: 12 },
  { key: 'email', label: '邮箱', type: 'input', span: 12 },
  { key: 'status', label: '状态', type: 'switch', span: 12 },
  { key: 'role', label: '角色', type: 'select', options: [
    { label: '管理员', value: 'admin' },
    { label: '用户', value: 'user' }
  ], span: 12 }
]
</script>
```

### ArtStatsCard - 统计卡片

```vue
<template>
  <art-stats-card
    title="总用户数"
    :value="userCount"
    icon="user"
    color="#409EFF"
    :trend="{ value: 12.5, isUp: true }"
  />
</template>

<script setup lang="ts">
import ArtStatsCard from '@/components/core/cards/art-stats-card/index.vue'

const userCount = ref(1234)
</script>
```

## 工作流程

当用户请求编写页面时：

1. **分析需求** - 确定页面类型（列表、表单、仪表板等）
2. **搜索组件** - 使用 `search.py` 查找相关组件
3. **选择组件** - 根据功能需求选择最合适的组件
4. **查看文档** - 阅读组件的 props、slots、events
5. **编写代码** - 使用组件而不是自己编写 Vue 代码

## 禁止行为

❌ **不要**自己编写表格组件 → 使用 `ArtTable`
❌ **不要**自己编写表单组件 → 使用 `ArtForm`
❌ **不要**自己编写图表组件 → 使用 `Art*Chart`
❌ **不要**自己编写统计卡片 → 使用 `ArtStatsCard`
❌ **不要**自己编写搜索栏 → 使用 `ArtSearchBar`

## 组件位置

所有组件位于：`frontend/src/components/core/`

## 相关资源

### 官方文档
- Art Design Pro 官方文档：https://www.artd.pro/docs/
- Element Plus 文档：https://element-plus.org/

### 本地文档（docs/ 目录）

**✅ 已保存 13/19 个官方文档到本地**，避免重复读取和幻觉问题。

**核心文档**：
- `docs/01-introduce.md` - 框架介绍和特色
- `docs/02-quick-start.md` - 快速开始指南
- `docs/03-must-read.md` - 必读：接口对接、网络请求、菜单配置
- `docs/05-project-introduce.md` - 项目结构
- `docs/07-route.md` - 路由和菜单配置
- `docs/08-settings.md` - 系统配置
- `docs/09-theme.md` - 主题配置和 CSS 变量
- `docs/10-icon.md` - 图标使用
- `docs/11-env-variables.md` - 环境变量配置
- `docs/12-build.md` - 构建和部署
- `docs/13-locale.md` - 国际化配置
- `docs/14-permission.md` - 权限管理
- `docs/18-standard.md` - 代码规范

查看完整文档列表：`docs/00-index.md`

**使用方法**：
所有文档都是本地 Markdown 文件，在编写代码前先查阅相关文档，确保遵循官方规范。
