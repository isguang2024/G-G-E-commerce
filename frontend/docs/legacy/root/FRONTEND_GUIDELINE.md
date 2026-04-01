# 前端管理页定板规范

> 生效日期：2026-04-01  
> 适用范围：`frontend/src/views/system/**`、`frontend/src/views/team/**`、`frontend/src/views/message/**`、`frontend/src/views/workspace/**`、`frontend/src/views/dashboard/**`

## 1. 规则级别

- `Must`：默认必须遵守，除非页面存在明确历史约束且已在代码中说明原因。
- `Should`：优先遵守；若有业务特殊性，需在同模块内保持一致。

## 2. 前端架构共识

- `Must`：菜单入口与受管页面是两种显式对象，不能继续混写成一套表单心智。
- `Must`：菜单管理只维护导航树、入口路由、外链与展示元数据。
- `Must`：页面管理只维护非菜单直达页、父子页面链和管理分组。
- `Must`：页面访问、菜单可见性、动作权限统一消费后端正式接口，不在前端自行补第二套判定。
- `Must`：前端治理页默认先复用共享组件和共享壳层，再考虑页面级定制。


## 3. 页面壳层标准

### 3.1 主面板定义

- `Must`：主面板只指页面壳层上的大块区域：
  - 顶部标签
  - 搜索区
  - 概览面板
  - 告警条
  - 主表格卡
  - 主内容板
- `Must`：组件内部、卡片内部、表单内部的细粒度布局不属于主面板节奏，不纳入本轮 `10px` 统一口径。

### 3.2 主面板标准节奏

- `Must`：主面板纵向标准节奏固定为 `10px`。
- `Must`：页面不得再混用 `10px / 12px / 16px` 作为主面板间距。
- `Must`：主面板间距统一由页面壳层控制，禁止依赖单组件默认 `margin` 叠加。

### 3.3 仅允许的页面结构

- `Must`：治理页只允许以下三种合法结构：
  - `tabs -> page-top-stack -> art-table-card`
  - `page-top-stack -> art-table-card`
  - `AdminWorkspaceHero -> art-table-card`
  - `AdminWorkspaceHero -> alert -> main section`
- `Must`：若页面同时存在搜索区和概览面板，必须使用 `page-top-stack` 或等效顶部栈容器。
- `Must`：若页面没有搜索区，只保留 `AdminWorkspaceHero -> main` 结构，不得额外补顶边距。

### 3.4 顶部栈容器

- `Must`：`page-top-stack` 是治理页顶部主面板唯一标准容器。
- `Must`：`page-top-stack` 内部 `gap` 固定为 `10px`。
- `Must`：`page-top-stack` 与后续 `art-table-card` / `main section` 之间总间距固定为 `10px`。
- `Must`：使用 `page-top-stack` 后，后续紧邻的 `art-table-card` 不得再额外补 `margin-top`。
- `Must`：搜索区显示或隐藏时，概览面板与主内容区间距必须保持稳定，不得因 `v-show` 造成双倍空隙或清零。

### 3.5 概览面板

- `Must`：概览面板统一使用 `AdminWorkspaceHero`，不再手写统计小卡拼接区。
- `Must`：`AdminWorkspaceHero` 采用固定结构：
  - 标题和说明在上
  - 分割线在中
  - 指标区在下
- `Must`：指标区默认左对齐、紧凑包裹展示，不得为了占满宽度均分整行。
- `Must`：`AdminWorkspaceHero` 本身不再承担页面间距职责；与上下主面板的关系统一由页面壳层处理。

### 3.6 告警条

- `Must`：`ElAlert` 作为主面板时，不得再手写额外 `margin-top`。
- `Must`：告警条与前后主面板的间距统一由页面壳层 `10px` 节奏处理。

## 4. 搜索区标准

- `Must`：所有列表治理页搜索区统一使用 `ArtSearchBar`，禁止手写固定宽度栅格。
- `Must`：`ArtSearchBar` 默认参数统一为：
  - `label-position="top"`
  - `span=8`
  - `gutter=16`
  - `showExpand=true`
- `Must`：按钮顺序固定为：`查询` 在前、`重置` 在后、`展开/收起` 在最右。
- `Must`：搜索区默认收起；页面不得再将 `showSearchBar` 默认设为 `true`。
- `Must`：筛选条件变化、重置、切换页签后，分页状态统一回第一页。
- `Must`：搜索区存在时，必须与概览面板一起放入 `page-top-stack`，不得单独依赖页面级 `marginTop` 调节间距。

## 5. 表格卡与工具栏标准

- `Must`：列表页主内容优先使用 `ArtTable + ArtTableHeader + ElCard.art-table-card` 结构。
- `Must`：`ArtTable` 主列表必须启用并保持分页条可见，不允许父容器裁切导致分页不可见。
- `Must`：`art-table-card` 顶部间距统一由页面壳层处理，组件自身不再负责页面级 `margin-top`。
- `Must`：表格工具栏与表格主体属于同一主卡内部结构，使用内部节奏，不再单独外补页面级间距。
- `Must`：分页条必须贴在主卡底部，分页区顶部保留稳定留白，不得紧贴数据区。
- `Must`：列表行操作统一使用 `ArtButtonMore` 三点菜单。
- `Must`：不在表格操作列长期并排放置超过 2 个文本按钮。

## 6. 分页标准

- `Must`：非 `ArtTable` 列表统一接入 `WorkspacePagination`，不得各页面重复造分页条样式。
- `Must`：本地分页状态模型统一为 `{ current, size, total }`。
- `Must`：非 `ArtTable` 列表统一采用“筛选后数据 -> 本地切片 -> `WorkspacePagination`”模式。
- `Should`：本地数据超过 10 条默认分页；小样本诊断表可不分页，但必须保证容器可滚动且不溢出。
- `Must`：卡片集合页不得一次性渲染全量列表，必须分页或分段加载。
- `Must`：分页条必须始终可见，不得被父容器 `overflow` 截断。

## 7. Drawer / Dialog 标准

- `Must`：大表单、绑定关系、权限配置、树选择、批量配置统一使用右侧 `Drawer`。
- `Must`：主配置抽屉统一追加 `config-drawer` 类。
- `Should`：轻量确认、预览、小型单字段输入、只读帮助继续使用 `Dialog`。
- `Must`：含列表的 Drawer / Dialog 统一遵循 `header + body + footer` 结构。
- `Must`：抽屉和弹窗内部的列表数据超过 10 条默认接 `WorkspacePagination`。
- `Must`：主配置抽屉顶部工具栏统一为左侧说明或筛选，右侧主操作按钮。

## 8. 响应式标准

- `Must`：搜索区列数统一为：桌面 3 列、平板 2 列、移动端 1 列。
- `Must`：卡片网格在 `1920 / 1600 / 1440` 分辨率下保持稳定列数，不允许过挤或过散。
- `Must`：`1024` 以下逐级降列，`768` 以下统一单列。
- `Must`：主面板节奏统一为 `10px`；卡片内部默认使用 `12px / 14px / 16px` 等细粒度节奏，但必须在组件内部自洽。
- `Must`：页面壳层禁止依赖固定高度和裁切式 `overflow` 处理分页、工具栏或底部操作区。

## 9. 页面头部与操作区

- `Must`：头部优先展示页面标题、摘要统计和主操作，不堆整排低频按钮。
- `Must`：危险操作放在菜单末尾，并使用危险色。
- `Should`：操作顺序统一为查看/配置 -> 编辑 -> 状态切换 -> 删除。
- `Must`：页头主操作和统计信息统一通过 `AdminWorkspaceHero` 承载。

## 10. 状态与辅助信息

- `Must`：状态统一使用 `ElTag`。
- `Must`：补充说明使用浅色小字。
- `Must`：空值统一显示 `-`。
- `Must`：路径、页面标识、权限键、Host、组件路径等技术字段优先用等宽或接近等宽风格展示。

## 11. 菜单与页面表单规范

### 11.1 菜单表单

- `Must`：菜单只配置 `directory / entry / external` 三类。
- `Must`：`entry` 直接配置 `path / component / accessMode`。
- `Must`：`external` 只配置稳定 `path` 与 `link`，不继续要求页面记录。
- `Must`：兼容期里的受管页面关系只允许只读展示，不作为主操作。

### 11.2 页面表单

- `Must`：页面表单只围绕受管页面展开：
  - 基础信息
  - 路由与渲染
  - 挂载与归属
  - 访问与行为
- `Should`：表单保留最终路径预览。
- `Must`：当页面 `accessMode = permission` 时，必须显式填写 `permissionKey`。
- `Must`：`spaceKey` 在页面表单里只能作为“当前空间视角 / 独立页暴露兼容输入”，不能再暗示页面按空间复制。

## 12. 路由与权限消费规则

- `Must`：前端首屏优先消费 `GET /api/v1/runtime/navigation`。
- `Must`：动态路由来源固定为：
  - 菜单入口路由
  - 受管页面路由
- `Must`：前端守卫只保留登录态检查、动态路由补注册、深链未命中兜底。
- `Must`：不再对菜单树 / 页面列表做前端重复权限裁剪。
- `Must`：`requiredAction / requiredActions` 仅保留给页面内按钮显示或提示用途。

## 13. 禁止项

- `Must`：禁止页面级手写条件 `marginTop` 控制搜索区、概览面板、告警条和主表格卡关系。
- `Must`：禁止 `AdminWorkspaceHero` 与 `art-table-card` 双重间距叠加。
- `Must`：禁止搜索区与概览区同时依赖容器 `gap` 和组件默认 `margin` 叠加。
- `Must`：禁止固定宽度筛选栅格，如筛选区内联 `style="width: xxxpx"` 组织布局。
- `Must`：禁止无分页的长列表直接渲染。
- `Must`：禁止分页条不可见或被裁切。
- `Must`：禁止表格底部被父容器 `overflow` 截断。
- `Must`：禁止继续把详情页、编辑页塞进菜单树当作导航节点维护。
- `Must`：禁止继续把菜单入口页重复登记到页面管理。
- `Must`：禁止前端按角色名或动作字段再推导一遍后端已经裁剪过的导航权限。

## 14. 落地要求

- `Must`：新管理页默认按本规范实现。
- `Must`：旧页面改造优先处理：
  - 页面壳层结构
  - 搜索区
  - 概览面板
  - 表格卡
  - 分页可见性
- `Must`：后续新增治理页时，开发者不需要再决定“到底用 10px 还是 12px”；主面板节奏固定按本规范执行。
