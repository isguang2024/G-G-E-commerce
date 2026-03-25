# 前端页面统一规范（系统管理）

本规范用于统一高频交互与样式，默认适用于 `frontend/src/views/system/**`。

## 0. 规则级别

- `Must`：默认必须遵守，除非页面已有明确历史约束。
- `Should`：优先遵守，若存在业务特殊性需要在代码中保持一致解释。

## 1. 搜索区规范

- `Must`：统一使用 `ArtSearchBar` 组件承载查询区。
- `Should`：搜索区默认可折叠，页面自行决定默认展开态。
- `Should`：字段顺序为主筛选（类型/状态）-> 路径/关键词 -> 辅助筛选。
- `Must`：按钮顺序固定为 `查询` 在前，`重置` 在后。
- `Must`：搜索区与表格卡片间距统一为 `12px`。

## 2. 列表操作列规范

- `Must`：单元格操作统一使用 `ArtButtonMore` 三点菜单。
- `Must`：不在表格内并排放置超过 2 个文本按钮。
- `Must`：危险操作（删除/停用）统一放在菜单末尾，并使用危险色。
- `Should`：常见操作顺序为查看/配置 -> 编辑 -> 状态切换 -> 删除。
- `Should`：操作列宽度控制在 `60 ~ 80`，居中展示。

## 3. 弹窗工具栏规范

- `Must`：承载大表单、权限配置、绑定关系、树选择、批量配置的主交互容器，统一使用右侧 `Drawer`，不再使用居中 `Dialog`。
- `Must`：主配置抽屉统一追加 `config-drawer` 类，复用全局 header/body/footer 排版基线，不在业务组件内重复定义通用抽屉间距。
- `Should`：以下场景继续保留 `Dialog`：轻量确认、小型单字段输入、预览、未注册列表、仅浏览型帮助弹层。
- `Must`：顶部工具栏布局统一为左侧筛选/选择，右侧主操作按钮。
- `Should`：常用选择器宽度控制在 `480 ~ 560px`。
- `Must`：弹窗内分页、筛选优先放在当前交互容器内部（例如下拉 footer），避免散落在弹窗外层。
- `Must`：主按钮文案使用动作词：`新增`、`保存`、`确认`。

## 4. 状态与辅助信息展示

- `Must`：状态统一用 `ElTag`，语义固定为：
  - 正常/启用：`success`
  - 停用/禁用：`danger` 或 `warning`
  - 中性信息：`info`
- `Must`：路径或主信息后的补充说明（summary/副标题）使用浅色文本：
  - `color: var(--el-text-color-secondary)`
  - `font-size: 12px`
- `Must`：空值统一显示 `-`。

## 5. 落地要求

- 新页面默认按本规范实现。
- 旧页面改造优先处理：搜索区、操作列、弹窗工具栏三类高频区域。
- 与历史页面冲突时：优先保证交互一致性，再做视觉微调。

## 6. 标准片段

搜索区标准：

```vue
<SearchBar v-show="showSearchBar" v-model="searchForm" @search="handleSearch" @reset="handleReset" />
<ElCard class="art-table-card" shadow="never" :style="{ marginTop: showSearchBar ? '12px' : '0' }">
```

操作列标准：

```ts
return h(ArtButtonMore, {
  list,
  onClick: (item: ButtonMoreItem) => handleAction(item, row)
})
```

## 7. 反例

- 禁止在表格操作列长期并排放置多个危险按钮。
- 禁止把下拉选择相关的分页或筛选散落到弹窗底部独立区域。
- 禁止同类页面一部分用三点菜单、一部分用纯文本链接操作列。
