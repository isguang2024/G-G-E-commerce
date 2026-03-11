# 基础表格使用示例

## 概述

这是一个最简单的表格使用示例，展示了如何使用 `useTable` Hook 快速创建一个数据表格。

## 完整代码

```vue
<!-- 基础表格 -->
<template>
  <div class="user-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetUserList } from '@/api/system-manage'

  defineOptions({ name: 'BasicTable' })

  const {
    data,
    columns,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange
  } = useTable({
    core: {
      apiFn: fetchGetUserList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        {
          prop: 'id',
          label: 'ID'
        },
        {
          prop: 'nickName',
          label: '昵称'
        },
        {
          prop: 'userGender',
          label: '性别',
          sortable: true,
          formatter: (row) => row.userGender || '未知'
        },
        {
          prop: 'userPhone',
          label: '手机号'
        },
        {
          prop: 'userEmail',
          label: '邮箱'
        }
      ]
    }
  })
</script>
```

## 关键点说明

### 1. useTable Hook

`useTable` 是 Art Design Pro 的核心 Hook，提供了表格的完整功能：

```typescript
const {
  data,              // 表格数据
  columns,           // 列配置
  loading,           // 加载状态
  pagination,        // 分页配置
  handleSizeChange,  // 每页数量变化处理
  handleCurrentChange // 当前页变化处理
} = useTable({...})
```

### 2. 列配置

通过 `columnsFactory` 定义表格列：

```typescript
columnsFactory: () => [
  {
    prop: 'id',           // 数据字段名
    label: 'ID'           // 列标题
  },
  {
    prop: 'userGender',
    label: '性别',
    sortable: true,       // 启用排序
    formatter: (row) =>   // 自定义格式化
      row.userGender || '未知'
  }
]
```

### 3. 样式类

- `art-full-height`: 自动计算页面剩余高度
- `art-table-card`: 符合系统样式的表格卡片，自动撑满剩余高度

## 常用配置

### 添加序号列

```typescript
columnsFactory: () => [
  { type: 'index', width: 60, label: '序号' },
  // ... 其他列
]
```

### 添加复选框列

```typescript
columnsFactory: () => [
  { type: 'selection' },
  // ... 其他列
]
```

### 自定义列渲染

使用 `formatter` 函数自定义列内容：

```typescript
{
  prop: 'status',
  label: '状态',
  formatter: (row) => {
    return h(ElTag, { type: 'success' }, () => '启用')
  }
}
```

## 分页配置

`useTable` 自动处理分页，默认配置：

```typescript
// 默认分页参数
apiParams: {
  current: 1,    // 当前页
  size: 20       // 每页数量
}
```

## 相关组件

- [ArtTable 组件文档](../../components/art-table.md)
- [useTable Hook 文档](https://www.artd.pro/docs/zh/guide/hooks/use-table.html)
- [高级表格示例](./advanced-table.md)
- [树形表格示例](./tree-table.md)

## 完整示例位置

`frontend/src/views/examples/tables/basic.vue`
