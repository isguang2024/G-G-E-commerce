# Art Design Pro 最佳实践指南

本文档提供了 Art Design Pro 组件库的最佳实践建议，帮助你编写更高质量、更易维护的代码。

## 📋 目录

- [表格最佳实践](#表格最佳实践)
- [表单最佳实践](#表单最佳实践)
- [图表最佳实践](#图表最佳实践)
- [权限控制最佳实践](#权限控制最佳实践)
- [代码组织最佳实践](#代码组织最佳实践)
- [性能优化最佳实践](#性能优化最佳实践)

---

## 表格最佳实践

### 1. 始终使用 useTable Hook

**✅ 推荐**：使用 `useTable` Hook

```typescript
const { data, columns, loading, pagination } = useTable({
  core: {
    apiFn: fetchGetUserList,
    apiParams: { current: 1, size: 20 },
    columnsFactory: () => [...]
  }
})
```

**❌ 不推荐**：手动管理表格状态

```typescript
const data = ref([])
const loading = ref(false)
const pagination = ref({ current: 1, size: 20, total: 0 })

// 需要手动处理加载、分页、刷新等逻辑
```

### 2. 使用 columnsFactory 而不是静态 columns

**✅ 推荐**：使用工厂函数

```typescript
columnsFactory: () => [
  { prop: 'id', label: 'ID' },
  { prop: 'name', label: '名称' }
]
```

**❌ 不推荐**：使用静态数组

```typescript
columns: [
  { prop: 'id', label: 'ID' },
  { prop: 'name', label: '名称' }
]
```

**原因**：工厂函数可以确保每次获取最新的配置，支持动态列。

### 3. 合理设置列宽

```typescript
// 固定重要列的宽度
{
  prop: 'id',
  label: 'ID',
  width: 100  // 固定宽度
}

// 让长内容自适应
{
  prop: 'description',
  label: '描述',
  minWidth: 200  // 最小宽度
}

// 操作列固定在右侧
{
  prop: 'operation',
  label: '操作',
  width: 120,
  fixed: 'right'
}
```

### 4. 使用 formatter 而不是 template

**✅ 推荐**：使用 `formatter`

```typescript
{
  prop: 'status',
  label: '状态',
  formatter: (row) => {
    return h(ElTag, { type: row.status === '1' ? 'success' : 'danger' }, () =>
      row.status === '1' ? '启用' : '禁用'
    )
  }
}
```

**❌ 不推荐**：使用 template 插槽

```vue
<template #status="{ row }">
  <ElTag :type="row.status === '1' ? 'success' : 'danger'">
    {{ row.status === '1' ? '启用' : '禁用' }}
  </ElTag>
</template>
```

**原因**：`formatter` 性能更好，代码更集中。

### 5. 大数据量优化

```typescript
// 1. 限制每页数量
const { data } = useTable({
  core: {
    apiParams: {
      size: 50  // 最多 50 条/页
    }
  }
})

// 2. 虚拟滚动（如果数据量特别大）
import { useVirtualList } from '@vueuse/core'

const { list, containerProps, wrapperProps } = useVirtualList(
  largeDataSource,
  { itemHeight: 50 }
)
```

### 6. 合理使用表格样式

```vue
<template>
  <div class="user-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTable
        :data="data"
        :columns="columns"
        :pagination="pagination"
      />
    </ElCard>
  </div>
</template>

<style scoped lang="scss">
// ✅ 使用提供的样式类
.art-full-height {
  // 自动计算剩余高度
}

.art-table-card {
  // 表格卡片样式
}
</style>
```

---

## 表单最佳实践

### 1. 使用 ArtSearchBar 统一搜索栏

**✅ 推荐**：使用 `ArtSearchBar`

```vue
<ArtSearchBar
  v-model="searchForm"
  :items="searchItems"
  @search="handleSearch"
  @reset="handleReset"
/>
```

**❌ 不推荐**：手动编写搜索栏

```vue
<ElForm :model="searchForm">
  <ElFormItem label="用户名">
    <ElInput v-model="searchForm.username" />
  </ElFormItem>
  <ElFormItem>
    <ElButton @click="handleSearch">搜索</ElButton>
    <ElButton @click="handleReset">重置</ElButton>
  </ElFormItem>
</ElForm>
```

### 2. 提取表单配置

```typescript
// ✅ 推荐：将表单配置提取为常量
const SEARCH_ITEMS = [
  {
    prop: 'username',
    label: '用户名',
    component: 'ElInput',
    componentProps: {
      placeholder: '请输入用户名',
      clearable: true
    }
  },
  {
    prop: 'status',
    label: '状态',
    component: 'ElSelect',
    componentProps: {
      options: [
        { label: '启用', value: '1' },
        { label: '禁用', value: '0' }
      ]
    }
  }
] as const

// ❌ 不推荐：直接写在模板中
```

### 3. 合理设置表单校验

```typescript
const rules = {
  // 必填校验
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],

  // 长度校验
  username: [
    { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
  ],

  // 格式校验
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],

  // 自定义校验
  age: [
    {
      validator: (rule, value, callback) => {
        if (value < 18) {
          callback(new Error('年龄必须大于 18 岁'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}
```

### 4. 表单布局建议

```typescript
// 默认收起（最多显示 4 个字段）
<ArtSearchBar
  v-model="formData"
  :items="formItems"
/>

// 默认展开
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :defaultExpanded="true"
/>

// 一行显示 3 个字段
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :span="8"
/>

// 自定义标签宽度
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :labelWidth="120"
/>
```

### 5. 响应式表单设计

```typescript
// 使用 Grid 系统实现响应式
const formItems = [
  {
    prop: 'username',
    label: '用户名',
    component: 'ElInput',
    span: {
      xs: 24,  // 手机端占满一行
      sm: 12,  // 小屏幕占半行
      md: 8,   // 中等屏幕占 1/3
      lg: 6    // 大屏幕占 1/4
    }
  }
]
```

---

## 图表最佳实践

### 1. 选择合适的图表类型

| 数据类型 | 推荐图表 | 组件 |
|---------|---------|------|
| 趋势数据（时间序列） | 折线图 | `ArtLineChart` |
| 分类对比 | 柱状图 | `ArtBarChart` |
| 占比分布 | 饼图/环形图 | `ArtRingChart` |
| 多维数据 | 雷达图 | `ArtRadarChart` |
| 地理分布 | 地图 | `ArtMapChart` |
| 股票/金融 | K线图 | `ArtKLineChart` |

### 2. 图表数据格式

```typescript
// ✅ 推荐：使用标准数据格式
const chartData = {
  xAxis: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
  series: [
    {
      name: '访问量',
      data: [120, 200, 150, 80, 70, 110, 130]
    }
  ]
}

// ❌ 不推荐：使用不规则格式
const chartData = {
  dates: ['2023-01-01', '2023-01-02'],
  values: [100, 200]
}
```

### 3. 图表配置建议

```typescript
// 1. 设置合适的高度
<ArtLineChart :data="chartData" height="400px" />

// 2. 使用主题色
const chartOptions = {
  color: [
    '#409EFF', '#67C23A', '#E6A23C', '#F56C6C'
  ]
}

// 3. 添加提示框
const chartOptions = {
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'shadow'
    }
  }
}

// 4. 添加图例
const chartOptions = {
  legend: {
    data: ['访问量', '下载量']
  }
}
```

### 4. 性能优化

```typescript
// 1. 限制数据点数量
const maxDataPoints = 1000
const chartData = rawData.slice(-maxDataPoints)

// 2. 使用数据抽样
function sampleData(data: number[], targetSize: number) {
  const step = Math.ceil(data.length / targetSize)
  return data.filter((_, index) => index % step === 0)
}

// 3. 防抖更新
import { debounce } from 'lodash-es'

const debouncedUpdateChart = debounce(() => {
  chart.value?.setOption(options)
}, 300)
```

### 5. 响应式图表

```vue
<template>
  <ElCard shadow="never">
    <ArtLineChart
      :data="chartData"
      :height="chartHeight"
    />
  </ElCard>
</template>

<script setup lang="ts">
import { useTableHeight } from '@/hooks/core/useTableHeight'

const chartHeight = computed(() => {
  // 根据屏幕高度动态调整
  return window.innerHeight < 800 ? '300px' : '400px'
})
</script>
```

---

## 权限控制最佳实践

### 1. 前后端双重验证

```typescript
// ✅ 推荐：前端 + 后端双重验证

// 前端：控制 UI 显示
<ElButton v-auth="'user:create'" @click="createUser">
  新增用户
</ElButton>

// 后端：API 验证
async function createUser(userData: User) {
  // 后端再次验证权限
  if (!hasPermission('user:create')) {
    throw new Error('权限不足')
  }

  // 执行创建操作
  await api.createUser(userData)
}
```

### 2. 细粒度权限设计

```typescript
// ✅ 推荐：细粒度权限
<ElButton v-auth="'user:create'">新增</ElButton>
<ElButton v-auth="'user:edit'">编辑</ElButton>
<ElButton v-auth="'user:delete'">删除</ElButton>

// ❌ 不推荐：粗粒度权限
<ElButton v-auth="'user:manage'">管理</ElButton>
```

### 3. 权限缓存

```typescript
// ✅ 推荐：缓存权限检查结果
const canDelete = computed(() => hasAuth('user:delete'))

// 在模板中使用
<ElButton v-if="canDelete">删除</ElButton>

// ❌ 不推荐：每次都检查
<ElButton v-if="hasAuth('user:delete')">删除</ElButton>
```

### 4. 权限码命名规范

```typescript
// ✅ 推荐：使用 <资源>:<操作> 格式
const permissionCodes = {
  'user:view': '查看用户',
  'user:create': '创建用户',
  'user:edit': '编辑用户',
  'user:delete': '删除用户',
  'user:export': '导出用户',
  'user:import': '导入用户'
}

// 批量操作使用批量后缀
'user:delete:batch'
'user:export:batch'
```

---

## 代码组织最佳实践

### 1. 目录结构

```
views/
├── system/
│   ├── user/
│   │   ├── index.vue              # 主页面
│   │   └── modules/
│   │       ├── user-search.vue    # 搜索栏组件
│   │       └── user-dialog.vue    # 弹窗组件
```

### 2. 组件拆分原则

```typescript
// ✅ 推荐：按功能拆分组件

// 主页面：只负责布局和数据流
// index.vue
<template>
  <div class="user-page art-full-height">
    <UserSearch v-model="searchForm" @search="handleSearch" />
    <UserTable :data="data" :columns="columns" />
    <UserDialog v-model:visible="dialogVisible" />
  </div>
</template>

// 搜索栏组件：负责搜索表单
// user-search.vue
// 表格组件：负责数据展示
// user-table.vue
// 弹窗组件：负责表单编辑
// user-dialog.vue

// ❌ 不推荐：所有逻辑写在一个组件
```

### 3. Composables 提取

```typescript
// ✅ 推荐：提取可复用逻辑
// composables/useUserTable.ts
export function useUserTable() {
  const { data, columns, loading, pagination } = useTable({...})

  const handleEdit = (row: User) => {
    // 编辑逻辑
  }

  const handleDelete = (row: User) => {
    // 删除逻辑
  }

  return {
    data,
    columns,
    loading,
    pagination,
    handleEdit,
    handleDelete
  }
}

// 在组件中使用
const { data, handleEdit, handleDelete } = useUserTable()
```

### 4. 类型定义

```typescript
// ✅ 推荐：明确定义类型
type UserListItem = Api.SystemManage.UserListItem

interface UserSearchParams {
  username?: string
  email?: string
  status?: string
}

interface UserDialogProps {
  visible: boolean
  type: DialogType
  data?: Partial<UserListItem>
}

// ❌ 不推荐：使用 any
const data: any = ref([])
```

---

## 性能优化最佳实践

### 1. 懒加载

```typescript
// ✅ 推荐：路由懒加载
const User = () => import('@/views/system/user/index.vue')

// 组件懒加载
const HeavyComponent = defineAsyncComponent(() =>
  import('./HeavyComponent.vue')
)

// 图片懒加载
<ElImage lazy :src="imageUrl" />
```

### 2. 计算属性缓存

```typescript
// ✅ 推荐：使用 computed
const filteredData = computed(() => {
  return rawData.value.filter(item => item.status === '1')
})

// ❌ 不推荐：使用 method（每次都重新计算）
const filteredData = () => {
  return rawData.value.filter(item => item.status === '1')
}
```

### 3. 防抖和节流

```typescript
import { debounce, throttle } from 'lodash-es'

// 搜索输入防抖
const handleSearchInput = debounce((value: string) => {
  searchParams.value.keyword = value
  getData()
}, 300)

// 滚动事件节流
const handleScroll = throttle(() => {
  // 处理滚动
}, 100)
```

### 4. 列表虚拟化

```vue
<!-- 长列表使用虚拟滚动 -->
<template>
  <div ref="containerRef" style="height: 400px; overflow: auto">
    <div
      v-for="item in virtualList"
      :key="item.id"
      :style="{ height: '50px' }"
    >
      {{ item.name }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useVirtualList } from '@vueuse/core'

const { list, containerProps, wrapperProps } = useVirtualList(
  largeDataSource,
  { itemHeight: 50 }
)
</script>
```

### 5. 按需引入

```typescript
// ✅ 推荐：按需引入组件
import { ElButton, ElInput, ElTable } from 'element-plus'

// ❌ 不推荐：全量引入
import ElementPlus from 'element-plus'
```

---

## 总结

### 核心原则

1. **使用 Hook 和组件**：充分利用 `useTable`、`ArtSearchBar` 等封装好的工具
2. **代码组织**：合理拆分组件，提取可复用逻辑
3. **性能优化**：懒加载、缓存、防抖节流
4. **权限控制**：细粒度设计，前后端双重验证
5. **类型安全**：使用 TypeScript 定义明确的类型

### 学习路径

1. **初级**：掌握基础组件使用
   - [ ] 会使用 `useTable`
   - [ ] 会使用 `ArtSearchBar`
   - [ ] 会配置表格列

2. **中级**：掌握组合使用
   - [ ] 能实现完整的 CRUD 页面
   - [ ] 能实现权限控制
   - [ ] 能优化代码结构

3. **高级**：掌握性能优化
   - [ ] 能优化大数据量表格
   - [ ] 能优化图表性能
   - [ ] 能优化首屏加载

---

**最后更新**：2025-03-03
**维护者**：Art Design Pro Skill Team
