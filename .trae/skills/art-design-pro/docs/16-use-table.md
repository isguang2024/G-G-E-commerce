# useTable Hook 文档

## 概述

`useTable` 是 Art Design Pro 的核心 Hook，提供了表格数据管理、分页、搜索、刷新等完整功能。

## 基础用法

```typescript
import { useTable } from '@/hooks/core/useTable'

const {
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
} = useTable({
  core: {
    apiFn: fetchGetUserList,
    apiParams: {
      current: 1,
      size: 20
    },
    columnsFactory: () => [...]
  }
})
```

## 配置参数

### core 配置

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| apiFn | Function | ✅ | API 请求函数 |
| apiParams | Object | ✅ | API 请求参数 |
| columnsFactory | Function | ✅ | 列配置工厂函数 |
| paginationKey | Object | ❌ | 分页字段映射 |

### transform 配置（数据转换）

| 参数 | 类型 | 说明 |
|------|------|------|
| dataTransformer | Function | 数据转换器 |
| responseTransformer | Function | 响应转换器 |

## API 函数要求

`apiFn` 必须返回一个 Promise，接收分页参数：

```typescript
// API 函数示例
export async function fetchGetUserList(params: {
  current: number
  size: number
  [key: string]: any
}) {
  return request.get<ApiResponseType<UserListItem[]>('/api/users', {
    params
  })
}

// 响应格式要求
interface ApiResponseType<T> {
  current: number      // 当前页
  size: number         // 每页数量
  total: number        // 总记录数
  records: T[]         // 数据列表
}
```

## 列配置

### 基础列配置

```typescript
columnsFactory: () => [
  {
    prop: 'id',        // 数据字段名
    label: 'ID',       // 列标题
    width: 100,        // 列宽度
    sortable: true     // 是否可排序
  }
]
```

### 自定义列渲染

使用 `formatter` 函数自定义列内容：

```typescript
{
  prop: 'status',
  label: '状态',
  formatter: (row: UserItem, column: any, value: any, index: number) => {
    return h(ElTag, { type: 'success' }, () => '启用')
  }
}
```

### 复杂列渲染

组合多个组件：

```typescript
{
  prop: 'userInfo',
  label: '用户信息',
  width: 280,
  formatter: (row) => {
    return h('div', { class: 'user flex-c' }, [
      h(ElImage, {
        class: 'size-9.5 rounded-md',
        src: row.avatar,
        previewSrcList: [row.avatar],
        previewTeleported: true
      }),
      h('div', { class: 'ml-2' }, [
        h('p', { class: 'user-name' }, row.userName),
        h('p', { class: 'email' }, row.userEmail)
      ])
    ])
  }
}
```

## 特殊列类型

### 选择列

```typescript
{ type: 'selection' }
```

### 序号列

```typescript
{
  type: 'index',
  width: 60,
  label: '序号'
}
```

### 操作列

```typescript
{
  prop: 'operation',
  label: '操作',
  width: 120,
  fixed: 'right',
  formatter: (row) =>
    h('div', [
      h(ArtButtonTable, {
        type: 'edit',
        onClick: () => handleEdit(row)
      }),
      h(ArtButtonTable, {
        type: 'delete',
        onClick: () => handleDelete(row)
      })
    ])
}
```

## 分页配置

### 默认分页字段

```typescript
// 默认分页参数名
apiParams: {
  current: 1,    // 当前页
  size: 20       // 每页数量
}

// 默认响应字段名
{
  current: number  // 当前页
  size: number     // 每页数量
  total: number    // 总记录数
}
```

### 自定义分页字段映射

如果 API 字段名不同，可以自定义映射：

```typescript
useTable({
  core: {
    apiFn: fetchList,
    apiParams: {
      current: 1,
      size: 20
    },
    paginationKey: {
      current: 'pageNum',      // 请求参数中的当前页字段
      size: 'pageSize'         // 请求参数中的每页数量字段
    }
  }
})
```

## 数据转换

### 数据转换器

转换 API 返回的数据：

```typescript
useTable({
  core: {...},
  transform: {
    dataTransformer: (records) => {
      return records.map(item => ({
        ...item,
        avatar: item.avatar || '/default-avatar.png'
      }))
    }
  }
})
```

### 响应转换器

转换整个响应：

```typescript
useTable({
  core: {...},
  transform: {
    responseTransformer: (response) => {
      return {
        current: response.data.page,
        size: response.data.pageSize,
        total: response.data.total,
        records: response.data.list
      }
    }
  }
})
```

## 搜索功能

### 搜索参数

```typescript
const searchForm = ref({
  username: '',
  status: '1'
})

const { searchParams, getData } = useTable({
  core: {
    apiFn: fetchGetUserList,
    apiParams: {
      current: 1,
      size: 20,
      ...searchForm.value
    },
    columnsFactory: () => [...]
  }
})

// 执行搜索
const handleSearch = (params: Record<string, any>) => {
  Object.assign(searchParams, params)
  getData()
}
```

### 重置搜索

```typescript
const resetSearchParams = () => {
  searchParams.value = {
    username: '',
    status: '1'
  }
  getData()
}
```

## 刷新数据

```typescript
const refreshData = () => {
  getData()
}
```

## 完整示例

### 基础表格

```typescript
const { data, columns, loading, pagination } = useTable({
  core: {
    apiFn: fetchGetUserList,
    apiParams: {
      current: 1,
      size: 20
    },
    columnsFactory: () => [
      { type: 'selection' },
      { type: 'index', width: 60, label: '序号' },
      { prop: 'id', label: 'ID' },
      { prop: 'username', label: '用户名' },
      { prop: 'email', label: '邮箱' }
    ]
  }
})
```

### 带搜索的表格

```vue
<template>
  <div class="user-page art-full-height">
    <ArtSearchBar v-model="searchForm" @search="handleSearch" />

    <ElCard class="art-table-card">
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'

  const searchForm = ref({
    username: '',
    status: '1'
  })

  const {
    data,
    columns,
    loading,
    pagination,
    searchParams,
    getData,
    handleCurrentChange
  } = useTable({
    core: {
      apiFn: fetchGetUserList,
      apiParams: {
        current: 1,
        size: 20,
        ...searchForm.value
      },
      columnsFactory: () => [...]
    }
  })

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }
</script>
```

## API 参考

### 返回值

| 属性 | 类型 | 说明 |
|------|------|------|
| data | Ref<T[]> | 表格数据 |
| columns | Ref<Column[]> | 列配置 |
| loading | Ref<boolean> | 加载状态 |
| pagination | Ref<Pagination> | 分页信息 |
| getData | Function | 获取数据方法 |
| searchParams | Ref<Object> | 搜索参数 |
| resetSearchParams | Function | 重置搜索参数 |
| handleSizeChange | Function | 每页数量变化处理 |
| handleCurrentChange | Function | 当前页变化处理 |
| refreshData | Function | 刷新数据 |

### Pagination 类型

```typescript
interface Pagination {
  current: number    // 当前页
  size: number       // 每页数量
  total: number      // 总记录数
}
```

## 最佳实践

### 1. 使用 columnsFactory

使用工厂函数确保每次获取最新的列配置：

```typescript
// ✅ 推荐
columnsFactory: () => [
  { prop: 'id', label: 'ID' }
]

// ❌ 不推荐
columns: [
  { prop: 'id', label: 'ID' }
]
```

### 2. 列配置提取

将复杂的列配置提取到单独的函数：

```typescript
const getUserColumns = () => [
  { type: 'selection' },
  { type: 'index', width: 60, label: '序号' },
  // ... 其他列
]

useTable({
  core: {
    columnsFactory: getUserColumns
  }
})
```

### 3. 类型定义

为表格数据定义类型：

```typescript
type UserListItem = Api.SystemManage.UserListItem

const { data } = useTable<UserListItem>({...})
```

## 常见问题

### Q: 如何实现自定义分页？

A: 通过 `paginationKey` 配置自定义字段映射：

```typescript
paginationKey: {
  current: 'page',
  size: 'pageSize'
}
```

### Q: 如何处理不规则的数据结构？

A: 使用 `dataTransformer` 或 `responseTransformer`：

```typescript
transform: {
  responseTransformer: (res) => {
    return {
      current: res.page,
      size: res.pageSize,
      total: res.total,
      records: res.data.list
    }
  }
}
```

### Q: 如何实现列的显示/隐藏？

A: 使用 `columnChecks` 控制列显示：

```vue
<ArtTableHeader v-model:columns="columnChecks" />
<ArtTable :columns="columns" />
```

## 相关文档

- [ArtTable 组件](./components/art-table.md)
- [基础表格示例](./examples/tables/basic-table.md)
- [CRUD 页面示例](./examples/templates/crud-page.md)

## 官方文档

- [useTable 官方文档](https://www.artd.pro/docs/zh/guide/hooks/use-table.html)

---

**最后更新**：2025-03-03
