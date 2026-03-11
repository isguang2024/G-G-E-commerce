# Art Design Pro 组件速查表

快速查找 Art Design Pro 组件的导入路径、常用属性和使用示例。

## 📋 目录

- [表格组件](#表格组件)
- [表单组件](#表单组件)
- [图表组件](#图表组件)
- [布局组件](#布局组件)
- [卡片组件](#卡片组件)
- [其他组件](#其他组件)
- [Hook 列表](#hook-列表)
- [样式类](#样式类)
- [指令](#指令)

---

## 表格组件

### ArtTable

**导入路径**：`@/components/core/tables/art-table/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | `Array` | - | 表格数据 |
| `columns` | `Array` | - | 列配置 |
| `loading` | `Boolean` | `false` | 加载状态 |
| `pagination` | `Object` | - | 分页配置 |
| `rowKey` | `String` | `'id'` | 行数据的 Key |

**常用事件**：

| 事件 | 说明 | 参数 |
|------|------|------|
| `selection-change` | 选择变化 | 选中的行数据 |
| `pagination:size-change` | 每页数量变化 | 新的每页数量 |
| `pagination:current-change` | 当前页变化 | 新的当前页 |

**快速使用**：

```vue
<ArtTable
  :data="data"
  :columns="columns"
  :loading="loading"
  :pagination="pagination"
  @selection-change="handleSelectionChange"
/>
```

---

### ArtTableHeader

**导入路径**：`@/components/core/tables/art-table-header/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `columns` | `Array` | - | 列配置（支持 v-model） |
| `loading` | `Boolean` | `false` | 加载状态 |

**插槽**：

| 插槽名 | 说明 |
|--------|------|
| `left` | 左侧内容（如操作按钮） |
| `right` | 右侧内容 |

**快速使用**：

```vue
<ArtTableHeader
  v-model:columns="columnChecks"
  :loading="loading"
  @refresh="refreshData"
>
  <template #left>
    <ElButton type="primary">新增</ElButton>
  </template>
</ArtTableHeader>
```

---

## 表单组件

### ArtSearchBar

**导入路径**：`@/components/core/forms/art-search-bar/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `modelValue` | `Object` | - | 表单数据（支持 v-model） |
| `items` | `Array` | - | 表单项配置 |
| `rules` | `Object` | - | 校验规则 |
| `defaultExpanded` | `Boolean` | `false` | 默认展开 |
| `labelWidth` | `String/Number` | `'auto'` | 标签宽度 |

**常用事件**：

| 事件 | 说明 | 参数 |
|------|------|------|
| `search` | 搜索 | 表单数据 |
| `reset` | 重置 | - |

**表单项配置**：

```typescript
interface SearchFormItem {
  prop: string                    // 字段名
  label: string                   // 标签文本
  component: string               // 组件名称
  componentProps?: Record<string, any>  // 组件 props
  span?: number                   // 栅格占位格数
}
```

**快速使用**：

```vue
<ArtSearchBar
  v-model="searchForm"
  :items="searchItems"
  @search="handleSearch"
  @reset="handleReset"
/>
```

---

### ArtForm

**导入路径**：`@/components/core/forms/art-form/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `modelValue` | `Object` | - | 表单数据（支持 v-model） |
| `items` | `Array` | - | 表单项配置 |
| `rules` | `Object` | - | 校验规则 |
| `labelWidth` | `String/Number` | `'100px'` | 标签宽度 |
| `labelPosition` | `String` | `'right'` | 标签位置 |

**常用方法**：

| 方法名 | 说明 | 参数 |
|--------|------|------|
| `validate` | 校验表单 | - |
| `resetFields` | 重置表单 | - |
| `clearValidate` | 清除校验 | - |

---

### ArtButtonTable

**导入路径**：`@/components/core/forms/art-button-table/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `type` | `String` | - | 按钮类型（edit/delete） |

**快速使用**：

```vue
<h(ArtButtonTable, {
  type: 'edit',
  onClick: () => handleEdit(row)
})
```

---

## 图表组件

### ArtLineChart

**导入路径**：`@/components/core/charts/art-line-chart/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `data` | `Object` | - | 图表数据 |
| `height` | `String` | `'400px'` | 图表高度 |

**数据格式**：

```typescript
{
  xAxis: string[]          // X 轴数据
  series: Array<{          // 系列数据
    name: string
    data: number[]
  }>
}
```

---

### ArtBarChart

**导入路径**：`@/components/core/charts/art-bar-chart/index.vue`

**常用属性**：同 `ArtLineChart`

---

### ArtRingChart

**导入路径**：`@/components/core/charts/art-ring-chart/index.vue`

**数据格式**：

```typescript
{
  legend: string[]          // 图例
  series: Array<{          // 系列数据
    name: string
    value: number
  }>
}
```

---

### ArtRadarChart

**导入路径**：`@/components/core/charts/art-radar-chart/index.vue`

**数据格式**：

```typescript
{
  indicator: Array<{       // 指标
    name: string
    max: number
  }>
  series: Array<{
    name: string
    data: number[]
  }>
}
```

---

## 布局组件

### ArtBreadcrumb

**导入路径**：`@/components/core/layouts/art-breadcrumb/index.vue`

**快速使用**：

```vue
<ArtBreadcrumb />
```

---

### ArtHeaderBar

**导入路径**：`@/components/core/layouts/art-header-bar/index.vue`

**插槽**：

| 插槽名 | 说明 |
|--------|------|
| `left` | 左侧内容（面包屑） |
| `panel` | 面板按钮 |
| `notification` | 通知按钮 |
| `fullscreen` | 全屏按钮 |
| `user` | 用户菜单 |

---

### ArtMenus

**导入路径**：
- 侧边栏菜单：`@/components/core/layouts/art-menus/art-sidebar-menu/index.vue`
- 顶部菜单：`@/components/core/layouts/art-menus/art-horizontal-menu/index.vue`
- 混合菜单：`@/components/core/layouts/art-menus/art-mixed-menu/index.vue`

---

### ArtPageContent

**导入路径**：`@/components/core/layouts/art-page-content/index.vue`

**快速使用**：

```vue
<ArtPageContent>
  <template #content>
    <!-- 页面内容 -->
  </template>
</ArtPageContent>
```

---

### ArtWorkTab

**导入路径**：`@/components/core/layouts/art-work-tab/index.vue`

**快速使用**：

```vue
<ArtWorkTab />
```

---

## 卡片组件

### ArtStatsCard

**导入路径**：`@/components/core/cards/art-stats-card/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `title` | `String` | - | 卡片标题 |
| `value` | `String/Number` | - | 统计值 |
| `icon` | `String` | - | 图标名称 |
| `trend` | `String` | - | 趋势（up/down） |
| `trendValue` | `String` | - | 趋势值 |

**快速使用**：

```vue
<ArtStatsCard
  title="总用户数"
  :value="1234"
  icon="user"
  trend="up"
  trendValue="12%"
/>
```

---

## 其他组件

### ArtBackToTop

**导入路径**：`@/components/core/base/art-back-to-top/index.vue`

**快速使用**：

```vue
<ArtBackToTop />
```

---

### ArtLogo

**导入路径**：`@/components/core/base/art-logo/index.vue`

**快速使用**：

```vue
<ArtLogo />
```

---

### ArtSvgIcon

**导入路径**：`@/components/core/base/art-svg-icon/index.vue`

**常用属性**：

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `name` | `String` | - | 图标名称 |

**快速使用**：

```vue
<ArtSvgIcon name="user" />
```

---

### ArtNotification

**导入路径**：`@/components/core/layouts/art-notification/index.vue`

**快速使用**：

```vue
<ArtNotification />
```

---

### ArtSettingsPanel

**导入路径**：`@/components/core/layouts/art-settings-panel/index.vue`

**快速使用**：

```vue
<ArtSettingsPanel />
```

---

### ArtGlobalSearch

**导入路径**：`@/components/core/layouts/art-global-search/index.vue`

**快捷键**：`Ctrl + K` 或 `Cmd + K`

---

## Hook 列表

### useTable

**导入路径**：`@/hooks/core/useTable`

**返回值**：

```typescript
{
  data,              // 表格数据
  columns,           // 列配置
  loading,           // 加载状态
  pagination,        // 分页信息
  getData,           // 获取数据
  searchParams,      // 搜索参数
  resetSearchParams, // 重置搜索参数
  handleSizeChange,  // 每页数量变化
  handleCurrentChange, // 当前页变化
  refreshData        // 刷新数据
}
```

**快速使用**：

```typescript
const { data, columns, loading, pagination } = useTable({
  core: {
    apiFn: fetchGetUserList,
    apiParams: { current: 1, size: 20 },
    columnsFactory: () => [...]
  }
})
```

---

### useAuth

**导入路径**：`@/hooks/core/useAuth`

**返回值**：

```typescript
{
  hasRole,         // 检查角色
  hasRoles,        // 检查多个角色
  hasAuth,         // 检查权限码
  hasAuths,        // 检查多个权限码
}
```

**快速使用**：

```typescript
const { hasAuth, hasRole } = useAuth()

if (hasAuth('user:create')) {
  // 有创建权限
}

if (hasRole('admin')) {
  // 是管理员
}
```

---

### useTheme

**导入路径**：`@/hooks/core/useTheme`

**返回值**：

```typescript
{
  theme,           // 当前主题
  setTheme,        // 设置主题
}
```

**快速使用**：

```typescript
const { theme, setTheme } = useTheme()

// 切换到暗色模式
setTheme('dark')
```

---

### useTableHeight

**导入路径**：`@/hooks/core/useTableHeight`

**返回值**：

```typescript
{
  tableHeight,     // 表格高度
  calcTableHeight  // 计算表格高度
}
```

---

### useTableColumns

**导入路径**：`@/hooks/core/useTableColumns`

---

## 样式类

### 页面布局类

| 类名 | 说明 |
|------|------|
| `art-full-height` | 自动计算页面剩余高度 |
| `art-table-card` | 表格卡片样式，自动撑满剩余高度 |

### Flex 工具类

| 类名 | 说明 |
|------|------|
| `flex-c` | Flex 布局，垂直居中 |
| `flex-cb` | Flex 布局，两端对齐 |

### 尺寸类

| 类名 | 说明 |
|------|------|
| `size-9.5` | 固定尺寸（约 38px） |

---

## 指令

### v-auth

**说明**：基于权限码控制元素显示

**用法**：

```vue
<ElButton v-auth="'user:create'">新增</ElButton>
<ElButton v-auth="['user:edit', 'user:delete']">操作</ElButton>
```

---

### v-roles

**说明**：基于角色控制元素显示

**用法**：

```vue
<ElButton v-roles="['admin']">管理员操作</ElButton>
<ElButton v-roles="['admin', 'editor']">编辑操作</ElButton>
```

**参数**：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `mode` | `'or'\|'and'` | `'or'` | 多角色判断逻辑 |

---

## 组件搜索

使用搜索工具快速查找组件：

```bash
# 搜索组件
python3 .claude/skills/art-design-pro/scripts/search.py <关键词>

# 列出所有组件
python3 .claude/skills/art-design-pro/scripts/list.py

# 生成页面代码
python3 .claude/skills/art-design-pro/scripts/generate.py --help
```

---

## 完整组件列表

详见：`data/components.csv` 或使用 `python3 scripts/list.py` 查看。

---

**最后更新**：2025-03-03
**维护者**：Art Design Pro Skill Team
