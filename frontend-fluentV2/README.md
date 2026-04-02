# frontend-fluentV2

`frontend-fluentV2` 是当前仓库的 React + Fluent 2 Web 迁移主线。  
当前已进入第 8 版：以 `frontend/src/views/**/*.vue` 全量页面清单和 docker PostgreSQL 中的 `menus` / `ui_pages` 为双重来源，把所有已知页面与子页面统一收口到 React，并按 `Fluent UI React v9 + Fluent 2 Web + Teams` 重构页面工作面。

## 第 8 版目标

- 不再只做路由页外壳，而是把 Vue 侧 `modules/*.vue` 对应的弹窗、抽屉、搜索、预览、配置流全部迁到 React。
- `pages/*` 只保留路由装配和少量页面级状态，交互能力下沉到 `features/<domain>/components|dialogs|drawers|panels`。
- 所有已知页面统一到四类范式：
  - 系统治理页
  - Teams 协作页
  - 设置与资料页
  - 总览页
- 所有主链动作都走真实 API、真实 Query、真实 mutation，不再保留 mock 页逻辑。

## 当前已完成能力

### 应用基础层

- Fluent 2 React 基础壳层
- 真实登录、会话恢复、当前用户、refresh token 单飞刷新
- 真实运行时导航、真实菜单空间切换、`route registry + placeholder fallback`
- 域级懒加载与按 `auth / dashboard / workspace / message / system / team` 拆包
- 菜单无限层级导航、收缩态级联浮层、移动端全屏抽屉树
- 菜单按空间持久化展开记忆、深层路径自动展开、长标题截断与延迟 tooltip

### 全量页面承接

- 公共页：
  - `login / register / forgot-password`
  - `403 / 404 / 500`
  - `result/success / result/fail`
  - `outside/Iframe`
- 总览与工作区：
  - `dashboard/console`
  - `workspace/home`
  - `workspace/inbox`
  - `welcome`
  - `user-center`
- 系统治理页：
  - `system/menu`
  - `system/page`
  - `system/role`
  - `system/user`
  - `system/action-permission`
  - `system/api-endpoint`
  - `system/feature-package`
  - `system/fast-enter`
  - `system/menu-space`
  - `system/access-trace`
  - `system/more`
  - `system/message*`
  - `system/team-roles-permissions`
- 团队页：
  - `team/team`
  - `team/team-members`
  - `team/message*`
  - `team/more`

### 已迁入 React 子模块的 Vue `modules/*.vue`

- `dashboard/console/modules/*`
  - 已由 [`ConsoleModules.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/dashboard/components/ConsoleModules.tsx) 承接
- `workspace/inbox/index.vue`
  - 左中右三栏已由 [`InboxPanels.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/workspace/components/InboxPanels.tsx) 承接
- `message/modules/*`
  - 目录式治理页已由 [`MessageCatalogWorkspaces.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/message/components/MessageCatalogWorkspaces.tsx) 承接
  - 调度工作区由 [`MessageWorkspacePages.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/pages/message/MessageWorkspacePages.tsx) 承接
- `system/menu/modules/*`
  - 菜单树节点与只读字段由 [`MenuTreeNode.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/system/components/MenuTreeNode.tsx) 承接
  - 摘要、编辑、关联页、删除确认由 [`SystemMenuPanels.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/system/components/SystemMenuPanels.tsx) 承接
- `team/team/modules/*` 与 `team/team-members/modules/*`
  - 已由 [`TeamWorkspaces.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/team/components/TeamWorkspaces.tsx) 承接

## 当前页面范式

### 系统治理页

- 适用：`system/menu`、`system/page`、`system/role`、`system/user`、`system/action-permission`、`system/api-endpoint`、`system/feature-package`、`system/access-trace`
- 结构：标题区、筛选区、主工作区、右侧详情/编辑、危险区

### Teams 协作页

- 适用：`workspace/inbox`、`system/message*`、`team/message*`、`team/team`、`team/team-members`、`system/team-roles-permissions`
- 结构：左侧列表/分类，中部主工作区，右侧上下文/来源/状态时间线

### 设置与资料页

- 适用：`system/menu-space`、`system/fast-enter`、`user-center`
- 结构：分组 section 表单 + 局部保存反馈

### 总览页

- 适用：`dashboard/console`
- 结构：摘要卡、趋势、待办、入口卡片

## 当前分层

- `app/`
  - 应用入口、Provider、路由、错误边界
- `shared/`
  - 稳定类型、请求客户端、Query keys、通用 UI
- `features/`
  - 各域服务、组件、对话框、抽屉、工作区
- `pages/`
  - 路由装配层，只负责组装页面骨架与域模块

## 工程命令

### 安装

```bash
pnpm --dir frontend-fluentV2 install
```

### 开发

```bash
pnpm --dir frontend-fluentV2 dev
```

### 类型检查

```bash
pnpm --dir frontend-fluentV2 exec tsc --noEmit
```

### 构建

```bash
pnpm --dir frontend-fluentV2 build
```

## 仍未完成的收尾项

- `system/page`、`system/user`、`system/feature-package`、`system/action-permission`、`system/api-endpoint` 仍有部分内部模块未完全从页面文件中抽离到域组件。
- 部分低频治理动作仍直接内嵌在页面里，还没有全部沉到 `dialogs|drawers|panels`。
- `auth` 与 `vendor-fluent` chunk 仍偏大，需要继续拆包。
- 第 8 版结束前还要继续更新 `docs/page-inventory.md` 与 `docs/architecture.md`，确保它们与当前真实结构一致。
