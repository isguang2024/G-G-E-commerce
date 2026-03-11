# Art Design Pro 代码生成器使用指南

## 📖 简介

代码生成器是 Art Design Pro Skill 的核心工具之一，可以自动生成标准的 Vue3 页面代码，大幅提升开发效率。

## 🎯 支持的页面类型

### 1. CRUD 列表页（`crud`）

生成完整的增删改查页面，包含：
- ✅ 主页面（index.vue）
- ✅ 搜索栏组件（modules/xxx-search.vue）
- ✅ 弹窗组件（modules/xxx-dialog.vue）
- ✅ 集成 useTable Hook
- ✅ 集成 ArtSearchBar 和 ArtTable 组件
- ✅ 完整的类型定义

**使用示例**：

```bash
# 生成用户管理页面
python3 .claude/skills/art-design-pro/scripts/generate.py crud \
  --name "User" \
  --path "system/user" \
  --fields "username,email,phone,status"

# 生成产品管理页面
python3 .claude/skills/art-design-pro/scripts/generate.py crud \
  --name "Product" \
  --path "product/list" \
  --fields "name,price,stock,category,status"
```

**生成的文件结构**：

```
frontend/src/views/system/user/
├── index.vue              # 主页面
└── modules/
    ├── user-search.vue    # 搜索栏组件
    └── user-dialog.vue    # 弹窗组件
```

**特性**：
- 自动生成表格列配置
- 自动生成搜索表单
- 自动生成弹窗表单
- 包含完整的类型定义
- 集成 Art Design Pro 组件
- 符合项目规范

---

### 2. 基础表格页（`table`）

生成简单的数据展示表格页面，适用于：
- 只读数据列表
- 不需要复杂操作的页面
- 数据展示和导出

**使用示例**：

```bash
# 生成订单列表页面
python3 .claude/skills/art-design-pro/scripts/generate.py table \
  --name "Order" \
  --path "order/list" \
  --fields "order_no,customer,total,status,created_at"

# 生成日志列表页面
python3 .claude/skills/art-design-pro/scripts/generate.py table \
  --name "Log" \
  --path "system/logs" \
  --fields "id,level,message,created_at"
```

**特性**：
- 简洁的页面结构
- 只包含表格展示
- 支持分页
- 集成 useTable Hook

---

### 3. 仪表板页面（`dashboard`）

生成数据统计和图表展示页面，包含：
- 统计卡片（ArtStatsCard）
- 图表组件（ArtLineChart、ArtBarChart 等）
- 响应式布局

**使用示例**：

```bash
# 生成数据分析仪表板
python3 .claude/skills/art-design-pro/scripts/generate.py dashboard \
  --name "Analytics" \
  --path "dashboard/analytics" \
  --charts "line,bar,pie"

# 生成销售统计仪表板
python3 .claude/skills/art-design-pro/scripts/generate.py dashboard \
  --name "Sales" \
  --path "dashboard/sales" \
  --charts "line,bar"
```

**支持的图表类型**：
- `line` - 折线图（ArtLineChart）
- `bar` - 柱状图（ArtBarChart）
- `pie` - 饼图（ArtRingChart）
- `radar` - 雷达图（ArtRadarChart）

---

## 🔧 参数说明

### 通用参数

| 参数 | 必填 | 说明 | 示例 |
|------|------|------|------|
| `--name` | ✅ | 实体名称（英文） | `User`, `Product`, `Order` |
| `--path` | ✅ | 页面路径（相对于 views） | `system/user`, `product/list` |
| `--fields` | ✅* | 字段列表（逗号分隔） | `username,email,status` |
| `--charts` | ❌ | 图表类型（仅 dashboard） | `line,bar,pie` |
| `--output` | ❌ | 输出目录 | `/path/to/output` |

* `--fields` 对 `crud` 和 `table` 类型是必填的，对 `dashboard` 类型不需要。

### 字段类型定义

支持在字段名后指定类型（使用 `:` 分隔）：

```bash
# 默认为 string 类型
python3 generate.py crud --name "User" --fields "username,email"

# 指定类型
python3 generate.py crud --name "User" --fields "username:string,age:number,active:boolean"
```

支持的类型：
- `string` - 字符串（默认）
- `number` - 数字
- `boolean` - 布尔值
- `date` - 日期

---

## 📋 工作流程

### 标准开发流程

1. **生成代码**
   ```bash
   python3 .claude/skills/art-design-pro/scripts/generate.py crud \
     --name "User" \
     --path "system/user" \
     --fields "username,email,status"
   ```

2. **复制代码到项目**
   - 将输出的代码保存到对应文件
   - 创建目录结构：`frontend/src/views/system/user/`

3. **调整代码**
   - 替换 `yourApiFunction` 为实际的 API 函数
   - 更新类型定义（`Api.YourModule.UserItem`）
   - 调整字段映射和校验规则
   - 添加业务逻辑

4. **测试页面**
   - 访问页面验证功能
   - 测试增删改查操作
   - 检查样式和响应式

---

## 💡 最佳实践

### 1. 字段命名规范

使用英文命名，代码生成器会自动生成中文标签：

```bash
# 推荐命名
--fields "username,email,phone,status,created_at"

# 自动生成的标签
# username → 用户名
# email → 邮箱
# phone → 手机号
# status → 状态
# created_at → 创建时间
```

### 2. 字段数量控制

- **搜索字段**：建议 3-5 个（最多显示前 5 个）
- **表格列**：建议 5-8 个（最多显示前 8 个）
- **表单字段**：不限制，按需添加

### 3. 图表搭配

```bash
# 分析类仪表板（趋势分析）
--charts "line,bar"

# 概览类仪表板（数据分布）
--charts "pie,radar"

# 综合仪表板
--charts "line,bar,pie,radar"
```

---

## 🐛 常见问题

### Q1: 生成的代码需要手动调整哪些地方？

**A**：主要需要调整：
1. API 函数：`yourApiFunction` → 实际的 API 函数
2. 类型定义：`Api.YourModule.UserItem` → 实际的类型定义
3. 字段校验规则：在弹窗组件的 `rules` 对象中添加
4. 业务逻辑：删除、提交等操作的实际实现

### Q2: 如何添加自定义列？

**A**：在生成代码后，修改 `columnsFactory` 中的列配置：

```typescript
{ prop: 'status', label: '状态',
  formatter: (row) => {
    return h(ElTag, { type: row.status === '1' ? 'success' : 'danger' },
      () => row.status === '1' ? '启用' : '禁用')
  }
}
```

### Q3: 如何集成真实的 API？

**A**：修改 `useTable` 中的 `apiFn` 参数：

```typescript
// 修改前
apiFn: yourApiFunction

// 修改后
import { fetchGetUserList } from '@/api/system-manage'

apiFn: fetchGetUserList
```

### Q4: 生成的代码不符合项目规范？

**A**：代码生成器生成的代码符合 Art Design Pro 项目规范，如果需要调整：
1. 修改生成器模板（`scripts/generate.py`）
2. 或者生成代码后手动调整

---

## 🚀 高级用法

### 1. 批量生成页面

使用 shell 脚本批量生成多个页面：

```bash
#!/bin/bash
# batch-generate.sh

pages=(
  "User:system/user:username,email,phone,status"
  "Role:system/role:name,code,description"
  "Permission:system/permission:name,code,type"
)

for page in "${pages[@]}"; do
  IFS=':' read -r name path fields <<< "$page"
  python3 .claude/skills/art-design-pro/scripts/generate.py crud \
    --name "$name" \
    --path "$path" \
    --fields "$fields"
done
```

### 2. 自定义模板

修改 `scripts/generate.py` 中的模板字符串，自定义生成的代码风格。

### 3. 集成到项目脚本

在项目的 `package.json` 中添加快捷命令：

```json
{
  "scripts": {
    "generate:page": "python3 .claude/skills/art-design-pro/scripts/generate.py"
  }
}
```

使用：
```bash
npm run generate:page crud --name "User" --path "system/user" --fields "username,email"
```

---

## 📊 效率对比

| 操作 | 手动编写 | 使用生成器 | 效率提升 |
|------|----------|------------|----------|
| CRUD 页面（3 个文件） | 60-90 分钟 | 5-10 分钟 | **600%-900%** |
| 基础表格页 | 20-30 分钟 | 2-5 分钟 | **400%-500%** |
| 仪表板页 | 40-60 分钟 | 5-10 分钟 | **400%-500%** |

---

## 📚 相关文档

- [useTable Hook 文档](https://www.artd.pro/docs/zh/guide/hooks/use-table.html)
- [ArtTable 组件文档](https://www.artd.pro/docs/zh/guide/components/art-table.html)
- [ArtSearchBar 组件文档](https://www.artd.pro/docs/zh/guide/components/art-search-bar.html)
- [项目示例页面](frontend/src/views/)

---

**最后更新**：2025-03-03
**维护者**：Art Design Pro Skill Team
