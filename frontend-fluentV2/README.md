# frontend-fluentV2

`frontend-fluentV2` 是当前仓库中的 React + Fluent 2 并行前端实验线，用来作为后续迁移 `frontend/` Vue 前端的基础壳。

## 项目目标

- 不修改现有 `frontend/` Vue 工程
- 不复制旧页面实现
- 先固化 React 应用壳、路由、导航、主题和 mock 数据边界
- 为后续逐页迁移建立稳定基座

## 技术栈

- React
- TypeScript
- Vite
- React Router（HashRouter）
- Fluent UI React v9
- Axios
- TanStack Query
- Zustand

## 目录说明

```text
src/
  app/                 应用入口、Provider、路由与错误边界
  shared/              api/config/lib/types/ui/mocks
  features/navigation  路由注册、导航查询与 metadata 查询
  features/session     本地假会话查询
  features/shell       App Shell、Header、SideNav、Breadcrumbs、PageContainer、壳层状态
  pages/               welcome/workspace/system/placeholder/not-found
docs/
  architecture.md
  navigation-shell.md
  migration-strategy.md
```

## 启动方式

```bash
pnpm --dir frontend-fluentV2 install
pnpm --dir frontend-fluentV2 dev
```

默认开发端口为 `9030`。

## 构建方式

```bash
pnpm --dir frontend-fluentV2 build
pnpm --dir frontend-fluentV2 exec tsc --noEmit
```

## 当前实现范围

- Fluent 2 主题 Provider
- Query / Router / Error Boundary 组合
- HashRouter 路由壳
- 顶部栏、侧边导航、菜单空间切换器、面包屑、页面标题区
- 基于 mock 的导航树、页面 metadata、当前用户与空间数据
- 首页、工作台首页、系统首页、系统菜单占位页
- 统一迁移占位页与 404 页面
- Axios client 与错误处理骨架

## 当前未实现范围

- 真实登录
- 真实权限校验
- 真实菜单裁剪
- 真实 API 对接
- 业务页 1:1 迁移
- 复杂表格、抽屉编辑和运行时菜单管理能力

## 后续迁移建议

1. 先迁移系统治理链路中的 `菜单 / 页面 / API / 角色 / 用户` 页面。
2. 再迁移工作台与消息域的列表、详情和三栏工作区页面。
3. 最后接入真实 API adapter 与上下文同步，保持壳层与路由模型不变。
