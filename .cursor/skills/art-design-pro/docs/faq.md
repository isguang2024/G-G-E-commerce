# Art Design Pro 常见问题解答 (FAQ)

本文档收集了 Art Design Pro 使用过程中的常见问题和解决方案。

## 📋 目录

- [组件使用问题](#组件使用问题)
- [表格相关问题](#表格相关问题)
- [表单相关问题](#表单相关问题)
- [路由配置问题](#路由配置问题)
- [样式主题问题](#样式主题问题)
- [权限控制问题](#权限控制问题)
- [构建部署问题](#构建部署问题)
- [性能优化问题](#性能优化问题)

---

## 组件使用问题

### Q1: 如何自定义表格列的内容？

**A**: 使用 `formatter` 函数自定义列内容：

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

### Q2: 如何实现表格列的显示/隐藏？

**A**: 使用 `ArtTableHeader` 组件的 `v-model:columns` 功能：

```vue
<template>
  <ArtTableHeader v-model:columns="columnChecks" />
  <ArtTable :columns="columns" />
</template>

<script setup lang="ts">
  const { columns, columnChecks } = useTable({...})
</script>
```

### Q3: 如何给表格添加操作按钮？

**A**: 在列配置中添加操作列：

```typescript
{
  prop: 'operation',
  label: '操作',
  width: 120,
  fixed: 'right',
  formatter: (row) =>
    h('div', [
      h(ElButton, {
        type: 'primary',
        size: 'small',
        onClick: () => handleEdit(row)
      }, () => '编辑'),
      h(ElButton, {
        type: 'danger',
        size: 'small',
        onClick: () => handleDelete(row)
      }, () => '删除')
    ])
}
```

### Q4: ArtSearchBar 如何自定义表单项？

**A**: 使用 `items` 配置数组：

```vue
<script setup lang="ts">
  const formItems = [
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
  ]
</script>
```

### Q5: 如何实现表单校验？

**A**: 在 `ArtSearchBar` 中添加 `rules` 配置：

```vue
<ArtSearchBar
  v-model="formData"
  :items="formItems"
  :rules="rules"
/>

<script setup lang="ts">
  const rules = {
    username: [
      { required: true, message: '请输入用户名', trigger: 'blur' },
      { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
    ],
    email: [
      { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
    ]
  }
</script>
```

---

## 表格相关问题

### Q6: useTable Hook 如何处理不规则的数据结构？

**A**: 使用 `transform` 配置进行数据转换：

```typescript
const { data } = useTable({
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

### Q7: 如何实现表格排序？

**A**: 在列配置中添加 `sortable` 属性：

```typescript
{
  prop: 'createTime',
  label: '创建时间',
  sortable: true
}
```

### Q8: 如何实现表格分页？

**A**: `useTable` 自动处理分页，只需在模板中使用：

```vue
<ArtTable
  :data="data"
  :columns="columns"
  :pagination="pagination"
  @pagination:size-change="handleSizeChange"
  @pagination:current-change="handleCurrentChange"
/>
```

### Q9: 如何获取选中的行数据？

**A**: 使用 `@selection-change` 事件：

```vue
<ArtTable
  :columns="columns"
  @selection-change="handleSelectionChange"
/>

<script setup lang="ts">
  const selectedRows = ref([])

  const handleSelectionChange = (selection: any[]) => {
    selectedRows.value = selection
  }
</script>
```

### Q10: 如何实现表格数据导出？

**A**: 使用 `ArtExcelExport` 组件：

```vue
<ArtExcelExport
  :data="exportData"
  :columns="exportColumns"
  filename="用户列表"
/>
```

---

## 表单相关问题

### Q11: 如何实现动态表单项？

**A**: 动态修改 `items` 数组：

```typescript
const formItems = ref([])

const addFormItem = () => {
  formItems.value.push({
    prop: 'newField',
    label: '新字段',
    component: 'ElInput'
  })
}
```

### Q12: 如何实现表单重置？

**A**: 监听 `@reset` 事件：

```vue
<ArtSearchBar
  v-model="formData"
  @reset="handleReset"
/>

<script setup lang="ts">
  const handleReset = () => {
    formData.value = {
      username: '',
      status: '1'
    }
  }
</script>
```

### Q13: 如何实现表单联动？

**A**: 使用 `watch` 监听表单变化：

```typescript
watch(() => formData.value.type, (newType) => {
  if (newType === 'special') {
    // 显示特殊字段
  }
})
```

---

## 路由配置问题

### Q14: 如何添加新页面路由？

**A**: 在 `src/router/modules/` 下创建路由模块：

```typescript
// src/router/modules/user.ts
export default {
  path: '/user',
  name: 'User',
  component: () => import('@/views/system/user/index.vue'),
  meta: {
    title: '用户管理',
    icon: 'user',
    roles: ['admin']
  }
}
```

### Q15: 静态路由和动态路由有什么区别？

**A**:
- **静态路由**：不需要权限即可访问的路由，如登录页、404 页
- **动态路由**：需要根据用户权限动态添加的路由，如管理页面

### Q16: 如何配置路由权限？

**A**: 在路由的 `meta` 中配置 `roles` 或 `auth`：

```typescript
{
  path: '/admin',
  meta: {
    roles: ['admin'],  // 需要管理员角色
    auth: ['user:view'] // 或使用权限码
  }
}
```

### Q17: 如何实现面包屑导航？

**A**: 使用 `ArtBreadcrumb` 组件，会自动根据路由生成：

```vue
<ArtBreadcrumb />
```

---

## 样式主题问题

### Q18: 如何修改主题色？

**A**: 在设置面板中修改主题色，或在 `src/config/setting.ts` 中配置默认主题：

```typescript
export default {
  theme: {
    primaryColor: '#409EFF',
    // ...
  }
}
```

### Q19: 如何自定义 CSS 变量？

**A**: 在 `src/assets/styles/core/` 下的样式文件中定义：

```scss
:root {
  --my-custom-color: #custom;
}
```

### Q20: Dark/Light 模式如何切换？

**A**: 使用设置面板或 `useTheme` Hook：

```typescript
import { useTheme } from '@/hooks/core/useTheme'

const { setTheme } = useTheme()

// 切换到暗色模式
setTheme('dark')

// 切换到亮色模式
setTheme('light')
```

### Q21: 如何自定义组件样式？

**A**: 使用深度选择器或全局样式：

```vue
<style scoped lang="scss">
// 方式 1: 使用 ::v-deep
:deep(.el-input__inner) {
  // 自定义样式
}

// 方式 2: 使用全局样式类
.my-custom-class {
  // 自定义样式
}
</style>
```

---

## 权限控制问题

### Q22: 如何使用 v-auth 指令？

**A**: 在需要权限控制的元素上添加 `v-auth` 指令：

```vue
<ElButton v-auth="'user:create'">新增用户</ElButton>
<ElButton v-auth="['user:edit', 'user:delete']">编辑/删除</ElButton>
```

### Q23: 如何使用 v-roles 指令？

**A**: 根据用户角色控制元素显示：

```vue
<ElButton v-roles="['admin']">管理员操作</ElButton>
<ElButton v-roles="['admin', 'editor']">编辑操作</ElButton>
```

### Q24: 如何在代码中检查权限？

**A**: 使用 `useAuth` Hook：

```typescript
import { useAuth } from '@/hooks/core/useAuth'

const { hasAuth, hasRole } = useAuth()

// 检查权限码
if (hasAuth('user:create')) {
  // 有创建用户权限
}

// 检查角色
if (hasRole('admin')) {
  // 是管理员
}
```

---

## 构建部署问题

### Q25: 打包后资源路径错误怎么办？

**A**: 在 `vite.config.ts` 中配置 `base`：

```typescript
export default defineConfig({
  base: '/your-sub-path/', // 如果部署在子路径下
  // ...
})
```

### Q26: 如何优化打包体积？

**A**:
1. 使用路由懒加载
2. 配置代码分割
3. 压缩图片资源
4. 移除未使用的依赖

```typescript
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus']
        }
      }
    }
  }
})
```

### Q27: 如何配置 Nginx？

**A**: 基本的 Nginx 配置：

```nginx
server {
  listen 80;
  server_name your-domain.com;
  root /path/to/dist;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }

  location /api {
    proxy_pass http://backend-server;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
  }
}
```

### Q28: 如何配置环境变量？

**A**:
1. 在项目根目录创建 `.env.development` 和 `.env.production`
2. 在代码中使用 `import.meta.env.VITE_XXX` 访问

```bash
# .env.development
VITE_API_BASE_URL=http://localhost:3000/api

# .env.production
VITE_API_BASE_URL=https://api.example.com
```

```typescript
const apiUrl = import.meta.env.VITE_API_BASE_URL
```

---

## 性能优化问题

### Q29: 表格数据量大时如何优化？

**A**:
1. 使用虚拟滚动
2. 分页加载
3. 懒加载图片
4. 优化列渲染

```typescript
// 优化前
formatter: (row) => h('div', { /* 复杂渲染 */ })

// 优化后：缓存渲染结果
const renderCell = (row) => {
  return useMemo(() => h('div', { /* 复杂渲染 */ }), [row.id])
}
```

### Q30: 如何减少首屏加载时间？

**A**:
1. 路由懒加载
2. 组件懒加载
3. 预加载关键资源
4. 使用 CDN

```typescript
// 路由懒加载
const User = () => import('@/views/system/user/index.vue')

// 组件懒加载
const HeavyComponent = defineAsyncComponent(() =>
  import('./HeavyComponent.vue')
)
```

### Q31: 如何优化图表性能？

**A**:
1. 限制数据点数量
2. 使用数据抽样
3. 防抖更新
4. 懒加载图表组件

```typescript
// 防抖更新
const debouncedUpdateChart = debounce(() => {
  chart.value?.setOption(options)
}, 300)
```

---

## 其他问题

### Q32: 如何使用 SVG 图标？

**A**: 使用 `ArtSvgIcon` 组件：

```vue
<ArtSvgIcon name="user" />
<ArtSvgIcon name="settings" />
```

### Q33: 如何使用国际化？

**A**: 使用 `$t()` 函数：

```vue
<template>
  <h1>{{ $t('menus.home') }}</h1>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const title = t('menus.home')
</script>
```

### Q34: 如何使用 WebSocket？

**A**: 使用项目中封装的 Socket 工具：

```typescript
import { socket } from '@/utils/socket'

// 连接
socket.connect()

// 监听消息
socket.on('message', (data) => {
  console.log('收到消息:', data)
})

// 发送消息
socket.emit('message', { data: 'hello' })
```

### Q35: 如何上传文件？

**A**: 使用 Element Plus 的 `ElUpload` 组件：

```vue
<ElUpload
  action="/api/upload"
  :on-success="handleSuccess"
  :before-upload="beforeUpload"
>
  <ElButton type="primary">点击上传</ElButton>
</ElUpload>
```

---

## 🔍 搜索问题

如果以上没有解答你的问题，可以：

1. 查看 [官方文档](https://www.artd.pro/docs/zh/guide/)
2. 查看 [示例代码](../examples/)
3. 使用组件搜索工具：`python3 scripts/search.py <关键词>`
4. 查看项目实际代码：`frontend/src/views/`

---

**最后更新**：2025-03-03
**维护者**：Art Design Pro Skill Team
