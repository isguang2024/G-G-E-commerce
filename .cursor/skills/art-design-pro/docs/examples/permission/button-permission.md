# 权限控制使用示例

## 概述

Art Design Pro 提供了完善的权限控制系统，支持基于角色（`v-roles`）和基于权限码（`v-auth`）的按钮级别权限控制。

## 权限指令

### v-roles - 基于角色的权限控制

根据用户角色控制元素显示：

```vue
<template>
  <div>
    <!-- 只有 admin 角色可见 -->
    <ElButton v-roles="['admin']">管理员操作</ElButton>

    <!-- admin 或 editor 角色可见 -->
    <ElButton v-roles="['admin', 'editor']">编辑操作</ElButton>

    <!-- 多个角色都具备时才可见（AND 逻辑） -->
    <ElButton v-roles="['admin', 'editor']" :mode="'and'">
      高级操作
    </ElButton>
  </div>
</template>
```

### v-auth - 基于权限码的控制

根据用户权限码控制元素显示：

```vue
<template>
  <div>
    <!-- 具备特定权限码可见 -->
    <ElButton v-auth="'user:create'">新增用户</ElButton>
    <ElButton v-auth="'user:edit'">编辑用户</ElButton>
    <ElButton v-auth="'user:delete'">删除用户</ElButton>

    <!-- 多个权限码都具备时才可见（AND 逻辑） -->
    <ElButton v-auth="['user:create', 'user:edit']" :mode="'and'">
      完整操作
    </ElButton>
  </div>
</template>
```

## 实际应用示例

### 表格操作列权限控制

```vue
<template>
  <ArtTable
    :data="data"
    :columns="columns"
  >
    <template #operation="{ row }">
      <ElSpace>
        <!-- 查看权限 - 所有用户可见 -->
        <ElButton
          type="primary"
          size="small"
          @click="handleView(row)"
        >
          查看
        </ElButton>

        <!-- 编辑权限 - admin 和 editor 可见 -->
        <ElButton
          v-roles="['admin', 'editor']"
          type="warning"
          size="small"
          @click="handleEdit(row)"
        >
          编辑
        </ElButton>

        <!-- 删除权限 - 只有 admin 可见 -->
        <ElButton
          v-roles="['admin']"
          type="danger"
          size="small"
          @click="handleDelete(row)"
        >
          删除
        </ElButton>
      </ElSpace>
    </template>
  </ArtTable>
</template>
```

### 页面头部按钮权限控制

```vue
<template>
  <div class="user-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading">
        <template #left>
          <ElSpace wrap>
            <!-- 新增权限 -->
            <ElButton
              v-auth="'user:create'"
              type="primary"
              @click="handleAdd"
            >
              新增用户
            </ElButton>

            <!-- 批量删除权限 -->
            <ElButton
              v-auth="'user:delete:batch'"
              type="danger"
              :disabled="selectedRows.length === 0"
              @click="handleBatchDelete"
            >
              批量删除
            </ElButton>

            <!-- 导出权限 -->
            <ElButton
              v-auth="'user:export'"
              @click="handleExport"
            >
              导出数据
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :data="data"
        :columns="columns"
        @selection-change="handleSelectionChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'

  const selectedRows = ref([])

  const { data, columns } = useTable({...})

  const handleSelectionChange = (selection: any[]) => {
    selectedRows.value = selection
  }

  // 各种操作方法...
</script>
```

### 自定义权限检查

使用 `useAuth` Hook 进行编程式权限检查：

```vue
<script setup lang="ts">
  import { useAuth } from '@/hooks/core/useAuth'

  const { hasRole, hasAuth, hasRoles, hasAuths } = useAuth()

  // 检查单个角色
  const isAdmin = hasRole('admin')

  // 检查多个角色（OR 逻辑）
  const canEdit = hasRoles(['admin', 'editor'])

  // 检查单个权限码
  const canCreate = hasAuth('user:create')

  // 检查多个权限码（OR 逻辑）
  const canManage = hasAuths(['user:create', 'user:edit'])

  // 使用示例
  const handleOperation = () => {
    if (!canManage) {
      ElMessage.error('权限不足')
      return
    }
    // 执行操作
  }
</script>
```

## 权限配置

### 角色定义

在用户数据结构中定义角色：

```typescript
interface User {
  id: string
  username: string
  roles: string[]      // 用户角色列表
  permissions: string[] // 用户权限码列表
}
```

### 权限码配置

```typescript
// 常见权限码命名规范
const permissionCodes = {
  // 用户管理
  'user:view': '查看用户',
  'user:create': '创建用户',
  'user:edit': '编辑用户',
  'user:delete': '删除用户',
  'user:export': '导出用户',

  // 角色管理
  'role:view': '查看角色',
  'role:create': '创建角色',
  'role:edit': '编辑角色',
  'role:delete': '删除角色',

  // 系统设置
  'system:view': '查看系统设置',
  'system:edit': '编辑系统设置'
}
```

## 指令参数说明

### v-roles 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| value | string\|string[] | - | 角色代码 |
| mode | 'or'\|'and' | 'or' | 多角色判断逻辑 |

### v-auth 参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| value | string\|string[] | - | 权限码 |
| mode | 'or'\|'and' | 'or' | 多权限码判断逻辑 |

## 最佳实践

### 1. 权限粒度设计

```typescript
// ✅ 推荐：细粒度权限控制
<ElButton v-auth="'user:create'">新增</ElButton>
<ElButton v-auth="'user:edit'">编辑</ElButton>
<ElButton v-auth="'user:delete'">删除</ElButton>

// ❌ 不推荐：粗粒度权限控制
<ElButton v-auth="'user:manage'">管理</ElButton>
```

### 2. 前后端权限验证

前端权限控制只是为了用户体验，真正的权限验证必须在后端进行：

```typescript
// 前端：控制 UI 显示
<ElButton v-auth="'user:delete'" @click="deleteUser">删除</ElButton>

// 后端：验证权限
async function deleteUser(userId: string) {
  // 检查用户是否有删除权限
  if (!hasPermission('user:delete')) {
    throw new Error('权限不足')
  }

  // 执行删除操作
  await api.deleteUser(userId)
}
```

### 3. 权限缓存

使用缓存提升权限检查性能：

```typescript
const { hasAuth } = useAuth()

// 缓存权限检查结果
const canDelete = computed(() => hasAuth('user:delete'))
```

## 相关文档

- [权限系统文档](../../14-permission.md)
- [useAuth Hook 文档](https://www.artd.pro/docs/zh/guide/hooks/use-auth.html)
- [角色管理示例](./role-control.md)

## 完整示例位置

- `frontend/src/views/examples/permission/button-auth/index.vue`
- `frontend/src/views/examples/permission/page-visibility/index.vue`
- `frontend/src/views/examples/permission/switch-role/index.vue`
