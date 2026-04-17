# 组件选用指南

开发新页面前必读。说明什么场景用什么组件，避免重复造轮子。

## 三层体系

```
① Art Design Pro core/    全局自动注册，无需 import，优先使用
② Element Plus            按需自动导入，无需 import，Art 未覆盖时直接用
③ business/               本仓自封装，需显式 import，跨两个以上页面才沉淀
```

详细组件列表：
- Art core 全部 54 个组件 → [docs/art-core-components.md](docs/art-core-components.md)
- business/ 自封装组件 + Hooks → [docs/business-components.md](docs/business-components.md)
- Element Plus 原件 → 参考 `node_modules/element-plus` 或本地 IDE 类型提示

---

## 选型规则

### 列表 / 表格页

```
表格主体     → <ArtTable>（内置 loading、空态、列配置、分页）
表格头部     → <ArtTableHeader>（标题 / 副标题 / 操作区）
行操作按钮   → <ArtButtonTable>
"更多"下拉   → <ArtButtonMore>
搜索栏       → <ArtSearchBar>
分页         → ArtTable 内置 :pagination prop（⚠️ 不要单独引 ElPagination / WorkspacePagination）
```

### 表单页

```
表单容器     → <ArtForm> 或 <ElForm>
字段标签     → <FieldLabel>（替代裸 <span>，支持 tooltip 说明）
枚举下拉     → <DictSelect code="xxx">（不要手写 <ElOption> 列表）
应用选择     → <AppKeySelect>
富文本       → <ArtWangEditor>
图片裁剪     → <ArtCutterImg>
拖拽验证     → <ArtDragVerify>
```

### 数据展示

```
数字统计     → <ArtStatsCard>
柱状图卡片   → <ArtBarChartCard>
折线图卡片   → <ArtLineChartCard>
环形图卡片   → <ArtDonutChartCard>
进度卡片     → <ArtProgressCard>
时间线卡片   → <ArtTimelineListCard>
图表原件     → ArtBarChart / ArtLineChart / ArtRingChart 等（需自定义容器时）
```

### 页面布局

```
提示横幅     → <ArtBasicBanner> / <ArtCardBanner>
工作区 Hero  → <AdminWorkspaceHero>（管理页顶部标题 + 统计 + 操作）
app 徽标     → <AppContextBadge>
回到顶部     → <ArtBackToTop>
SVG 图标     → <ArtSvgIcon>
```

### 权限相关

```
权限工作台   → <PermissionActionWorkbench>
权限来源面板 → <PermissionSourcePanels>
权限摘要标签 → <PermissionSummaryTags>
级联选择     → <PermissionActionCascaderPanel>
套餐预览     → <FeaturePackageGrantPreview>
```

### 可观测性（技术向）

```
JSON 查看器  → <JsonViewer>
追踪抽屉     → <TraceDrawer>
```

---

## 新增组件的边界

- 只服务**一个页面** → 留在 `views/<page>/modules/`，不要放到这里
- 跨**两个及以上页面** → 沉淀到 `business/<domain>/`，并在 [docs/business-components.md](docs/business-components.md) 登记一行

---

## 不要绕开的规则

| 场景 | 禁止 | 应该用 |
| --- | --- | --- |
| 枚举下拉 | 手写 `<ElOption>` 列表 | `<DictSelect code="...">` |
| 表单字段标签 | `<span class="form-tip">` | `<FieldLabel>` |
| 列表分页 | `<ElPagination>` / `<WorkspacePagination>` | `ArtTable` 内置 `:pagination` |
| 改 core/ 组件 | 就地 fork | 在 `business/` 包一层或升级基座 |
| 新 business/ 组件 | 不登记文档 | 同 commit 内在文档表格加一行 |
