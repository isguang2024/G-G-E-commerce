# 搜索栏使用示例

## 概述

`ArtSearchBar` 是 Art Design Pro 的搜索栏组件，支持动态表单项、折叠展开、表单校验等功能。

## 基础用法

```vue
<template>
  <div class="search-page">
    <ArtSearchBar
      v-model="formData"
      :items="formItems"
      @search="handleSearch"
      @reset="handleReset"
    />
  </div>
</template>

<script setup lang="ts">
  import type { SearchFormItem } from '@/components/core/forms/art-search-bar/index.vue'

  const formData = ref({
    username: '',
    email: '',
    status: ''
  })

  // 表单项配置
  const formItems: SearchFormItem[] = [
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
      prop: 'email',
      label: '邮箱',
      component: 'ElInput',
      componentProps: {
        placeholder: '请输入邮箱',
        clearable: true
      }
    },
    {
      prop: 'status',
      label: '状态',
      component: 'ElSelect',
      componentProps: {
        placeholder: '请选择状态',
        clearable: true,
        options: [
          { label: '启用', value: '1' },
          { label: '禁用', value: '0' }
        ]
      }
    }
  ]

  const handleSearch = (params: Record<string, any>) => {
    console.log('搜索参数:', params)
    // 执行搜索逻辑
  }

  const handleReset = () => {
    console.log('重置表单')
  }
</script>
```

## 高级配置

### 默认展开

```vue
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :defaultExpanded="true"
  @search="handleSearch"
/>
```

### 自定义标签宽度

```vue
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :labelWidth="120"
  @search="handleSearch"
/>
```

### 设置一行显示的组件数

```vue
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :span="8"
  @search="handleSearch"
/>
```

### 表单校验

```vue
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :rules="rules"
  @search="handleSearch"
/>

<script setup lang="ts">
  const rules = {
    username: [
      { required: true, message: '请输入用户名', trigger: 'blur' }
    ],
    email: [
      { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
    ]
  }
</script>
```

## 表单项配置类型

```typescript
interface SearchFormItem {
  prop: string                    // 字段名
  label: string                   // 标签文本
  component: string               // 组件名称（如 'ElInput'）
  componentProps?: Record<string, any>  // 组件 props
  span?: number                   // 栅格占位格数
  itemProps?: Record<string, any> // FormItem props
}
```

## 支持的组件

- `ElInput` - 输入框
- `ElSelect` - 下拉选择
- `ElDatePicker` - 日期选择
- `ElTimePicker` - 时间选择
- `ElCascader` - 级联选择
- `ElSwitch` - 开关

## 动态操作

### 添加表单项

```typescript
const addFormItem = () => {
  formItems.value.push({
    prop: 'newField',
    label: '新字段',
    component: 'ElInput',
    componentProps: {
      placeholder: '请输入'
    }
  })
}
```

### 修改表单项

```typescript
const updateFormItem = () => {
  const index = formItems.value.findIndex(item => item.prop === 'username')
  if (index !== -1) {
    formItems.value[index].label = '用户名（已修改）'
  }
}
```

### 删除表单项

```typescript
const deleteFormItem = () => {
  const index = formItems.value.findIndex(item => item.prop === 'email')
  if (index !== -1) {
    formItems.value.splice(index, 1)
  }
}
```

## 插槽使用

```vue
<ArtSearchBar
  v-model="formData"
  :items="formItems"
>
  <template #slots>
    <ElInput v-model="formData.custom" placeholder="自定义组件" />
  </template>
</ArtSearchBar>
```

## 实际应用示例

### 结合 useTable 使用

```vue
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 表格 -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTable
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
    status: ''
  })

  const searchItems = [
    {
      prop: 'username',
      label: '用户名',
      component: 'ElInput'
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
  ]

  const {
    data,
    columns,
    pagination,
    getData,
    searchParams
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

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  const handleReset = () => {
    searchForm.value = {
      username: '',
      status: ''
    }
  }
</script>
```

## 相关组件

- [ArtSearchBar 组件文档](../../components/art-search-bar.md)
- [ArtTable 组件文档](../../components/art-table.md)
- [表单集合示例](./form-collection.md)

## 完整示例位置

`frontend/src/views/examples/forms/search-bar.vue`
