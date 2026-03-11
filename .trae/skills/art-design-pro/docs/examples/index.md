# Art Design Pro 使用示例

本目录包含从 HCG 项目中提取的真实使用案例，帮助你快速上手 Art Design Pro 组件库。

## 📚 示例目录

### 表格示例 ([tables/](./tables/))

| 示例 | 说明 | 难度 |
|------|------|------|
| [基础表格](./tables/basic-table.md) | 最简单的表格使用示例 | ⭐ |
| [高级表格](./tables/advanced-table.md) | 包含自定义列、排序、筛选等功能 | ⭐⭐⭐ |
| [树形表格](./tables/tree-table.md) | 树形结构数据展示 | ⭐⭐ |

### 表单示例 ([forms/](./forms/))

| 示例 | 说明 | 难度 |
|------|------|------|
| [搜索栏](./forms/search-bar.md) | ArtSearchBar 组件完整使用指南 | ⭐⭐ |
| [表单集合](./forms/form-collection.md) | 各种表单组件的组合使用 | ⭐⭐⭐ |

### 权限控制示例 ([permission/](./permission/))

| 示例 | 说明 | 难度 |
|------|------|------|
| [按钮权限](./permission/button-permission.md) | v-roles 和 v-auth 指令使用 | ⭐⭐ |
| [角色控制](./permission/role-control.md) | 基于角色的权限管理 | ⭐⭐⭐ |
| [页面可见性](./permission/page-visibility.md) | 页面级别的权限控制 | ⭐⭐ |

### 页面模板 ([templates/](./templates/))

| 模板 | 说明 | 难度 |
|------|------|------|
| [CRUD 页面](./templates/crud-page.md) | 完整的增删改查页面示例 | ⭐⭐⭐ |
| [仪表板](./templates/dashboard.md) | 数据统计和图表展示页面 | ⭐⭐⭐ |
| [列表页](./templates/list-page.md) | 简单的数据列表页面 | ⭐⭐ |

## 🚀 快速开始

### 1. 查看基础示例

如果你是第一次使用 Art Design Pro，建议按以下顺序阅读：

1. [基础表格](./tables/basic-table.md) - 了解最简单的表格使用
2. [搜索栏](./forms/search-bar.md) - 学习搜索功能实现
3. [CRUD 页面](./templates/crud-page.md) - 查看完整功能组合

### 2. 使用代码生成器

使用代码生成器快速创建页面模板：

```bash
# 生成 CRUD 页面
python3 .claude/skills/art-design-pro/scripts/generate.py \
  crud \
  --name "User" \
  --path "system/user" \
  --fields "username,email,phone,status"
```

详见：[代码生成器使用指南](../generator-guide.md)

### 3. 参考实际项目代码

所有示例都提取自实际项目，你可以在以下位置找到完整代码：

- 表格示例：`frontend/src/views/examples/tables/`
- 表单示例：`frontend/src/views/examples/forms/`
- 权限示例：`frontend/src/views/examples/permission/`
- 系统页面：`frontend/src/views/system/`

## 📖 常见模式

### 模式 1：搜索 + 表格

```vue
<template>
  <div class="page art-full-height">
    <!-- 搜索栏 -->
    <ArtSearchBar v-model="searchForm" @search="handleSearch" />

    <!-- 表格 -->
    <ElCard class="art-table-card">
      <ArtTable :data="data" :columns="columns" />
    </ElCard>
  </div>
</template>
```

### 模式 2：搜索 + 表格 + 弹窗

```vue
<template>
  <div class="page art-full-height">
    <ArtSearchBar v-model="searchForm" @search="handleSearch" />

    <ElCard class="art-table-card">
      <ArtTableHeader>
        <template #left>
          <ElButton @click="showDialog">新增</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable :data="data" :columns="columns" />

      <ElDialog v-model="dialogVisible">
        <ElForm>...</ElForm>
      </ElDialog>
    </ElCard>
  </div>
</template>
```

### 模式 3：权限控制

```vue
<template>
  <ElButton v-auth="'user:create'">新增</ElButton>
  <ElButton v-roles="['admin']">管理</ElButton>
</template>
```

## 🎯 学习路径

### 初级（1-2 天）

- [ ] 阅读基础表格示例
- [ ] 阅读搜索栏示例
- [ ] 创建一个简单的列表页面

### 中级（3-5 天）

- [ ] 学习完整的 CRUD 页面
- [ ] 掌握自定义列渲染
- [ ] 实现权限控制

### 高级（5-7 天）

- [ ] 学习高级表格功能
- [ ] 掌握表单校验
- [ ] 实现复杂业务逻辑

## 💡 最佳实践

### 1. 使用 useTable Hook

`useTable` Hook 提供了表格的完整功能，推荐所有表格页面使用：

```typescript
const { data, columns, loading, pagination } = useTable({...})
```

### 2. 使用 ArtSearchBar

`ArtSearchBar` 组件封装了搜索栏的通用逻辑：

```vue
<ArtSearchBar v-model="formData" :items="formItems" />
```

### 3. 权限控制

使用 `v-auth` 和 `v-roles` 指令控制按钮权限：

```vue
<ElButton v-auth="'user:create'">新增</ElButton>
```

### 4. 样式类

使用 Art Design Pro 提供的样式类：

- `art-full-height`: 自动计算剩余高度
- `art-table-card`: 表格卡片样式
- `flex-c`: Flex 垂直居中

## 📚 相关文档

- [组件文档](../components/)
- [Hook 文档](../hooks/)
- [官方文档](https://www.artd.pro/docs/zh/guide/)
- [代码生成器](../generator-guide.md)

## 🤝 贡献示例

如果你有好的示例想要分享，请：

1. 将代码提交到 `frontend/src/views/examples/`
2. 在本目录添加对应的文档
3. 更新索引文件

---

**最后更新**：2025-03-03
**维护者**：Art Design Pro Skill Team
