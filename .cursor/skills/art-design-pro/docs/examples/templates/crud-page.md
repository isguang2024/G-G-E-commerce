# 完整 CRUD 页面示例

## 概述

这是一个完整的用户管理 CRUD 页面示例，展示了 Art Design Pro 的核心功能组合使用。

## 页面结构

```
system/user/
├── index.vue              # 主页面
├── modules/
│   ├── user-search.vue    # 搜索栏组件
│   └── user-dialog.vue    # 弹窗组件
```

## 主页面代码

```vue
<!-- 用户管理页面 -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch
      v-model="searchForm"
      @search="handleSearch"
      @reset="resetSearchParams"
    />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton
              v-auth="'user:create'"
              type="primary"
              @click="showDialog('add')"
              v-ripple
            >
              新增用户
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <!-- 用户弹窗 -->
      <UserDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :user-data="currentUserData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetUserList } from '@/api/system-manage'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import { ElTag, ElMessageBox, ElImage } from 'element-plus'
  import { DialogType } from '@/types'

  defineOptions({ name: 'User' })

  type UserListItem = Api.SystemManage.UserListItem

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

  // 选中行
  const selectedRows = ref<UserListItem[]>([])

  // 搜索表单
  const searchForm = ref({
    userName: undefined,
    userGender: undefined,
    userPhone: undefined,
    userEmail: undefined,
    status: '1'
  })

  // 用户状态配置
  const USER_STATUS_CONFIG = {
    '1': { type: 'success' as const, text: '在线' },
    '2': { type: 'info' as const, text: '离线' },
    '3': { type: 'warning' as const, text: '异常' },
    '4': { type: 'danger' as const, text: '注销' }
  } as const

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetUserList,
      apiParams: {
        current: 1,
        size: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { type: 'selection' },
        { type: 'index', width: 60, label: '序号' },
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
        },
        {
          prop: 'userGender',
          label: '性别',
          sortable: true,
          formatter: (row) => row.userGender
        },
        { prop: 'userPhone', label: '手机号' },
        {
          prop: 'status',
          label: '状态',
          formatter: (row) => {
            const statusConfig = USER_STATUS_CONFIG[row.status as keyof typeof USER_STATUS_CONFIG]
            return h(ElTag, { type: statusConfig.type }, () => statusConfig.text)
          }
        },
        {
          prop: 'createTime',
          label: '创建日期',
          sortable: true
        },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row) =>
            h('div', [
              h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }),
              h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteUser(row)
              })
            ])
        }
      ]
    }
  })

  /**
   * 搜索处理
   */
  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  /**
   * 显示用户弹窗
   */
  const showDialog = (type: DialogType, row?: UserListItem): void => {
    dialogType.value = type
    currentUserData.value = row || {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /**
   * 删除用户
   */
  const deleteUser = (row: UserListItem): void => {
    ElMessageBox.confirm('确定要删除该用户吗？', '删除用户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    }).then(async () => {
      // 调用删除 API
      // await deleteUserApi(row.id)
      ElMessage.success('删除成功')
      refreshData()
    })
  }

  /**
   * 处理弹窗提交事件
   */
  const handleDialogSubmit = async () => {
    dialogVisible.value = false
    currentUserData.value = {}
    refreshData()
  }

  /**
   * 处理表格行选择变化
   */
  const handleSelectionChange = (selection: UserListItem[]): void => {
    selectedRows.value = selection
  }
</script>
```

## 关键功能点

### 1. useTable Hook 集成

```typescript
const {
  columns,           // 列配置
  columnChecks,      // 列显示控制
  data,              // 表格数据
  loading,           // 加载状态
  pagination,        // 分页信息
  getData,           // 获取数据
  searchParams,      // 搜索参数
  resetSearchParams, // 重置搜索
  handleSizeChange,  // 每页数量变化
  handleCurrentChange, // 当前页变化
  refreshData        // 刷新数据
} = useTable({...})
```

### 2. 自定义列渲染

使用 `formatter` 函数实现复杂的列渲染：

```typescript
{
  prop: 'userInfo',
  label: '用户信息',
  formatter: (row) => {
    return h('div', { class: 'user flex-c' }, [
      h(ElImage, { src: row.avatar }),
      h('div', { class: 'ml-2' }, [
        h('p', row.userName),
        h('p', row.userEmail)
      ])
    ])
  }
}
```

### 3. 权限控制

```vue
<ElButton
  v-auth="'user:create'"
  type="primary"
  @click="showDialog('add')"
>
  新增用户
</ElButton>
```

### 4. 搜索栏集成

```vue
<UserSearch
  v-model="searchForm"
  @search="handleSearch"
  @reset="resetSearchParams"
/>
```

### 5. 弹窗表单

```vue
<UserDialog
  v-model:visible="dialogVisible"
  :type="dialogType"
  :user-data="currentUserData"
  @submit="handleDialogSubmit"
/>
```

## 样式说明

- `art-full-height`: 自动计算页面剩余高度
- `art-table-card`: 符合系统样式的表格卡片，自动撑满剩余高度
- `flex-c`: Flex 布局，垂直居中
- `size-9.5`: 固定尺寸（约 38px）

## 相关组件

- [ArtTable 组件](../../components/art-table.md)
- [ArtTableHeader 组件](../../components/art-table-header.md)
- [ArtSearchBar 组件](../../components/art-search-bar.md)
- [useTable Hook](https://www.artd.pro/docs/zh/guide/hooks/use-table.html)

## 完整示例位置

`frontend/src/views/system/user/index.vue`

## 使用代码生成器

可以使用代码生成器快速创建类似页面：

```bash
python3 .claude/skills/art-design-pro/scripts/generate.py \
  crud \
  --name "User" \
  --path "system/user" \
  --fields "username,email,phone,status"
```

详见 [代码生成器使用指南](../../generator-guide.md)。
