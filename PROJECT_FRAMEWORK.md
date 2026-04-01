# PROJECT_FRAMEWORK.md

## 当前项目主框架

仓库当前存在两条前端线：

1. `frontend/`
   - Vue 3 + TypeScript + Vite
   - 作为现有业务主线继续保留
2. `frontend-fluentV2/`
   - React + TypeScript + Vite
   - Fluent UI React v9
   - 作为后续迁移基座

## React 迁移线固定约束

- 路由：`HashRouter`
- 主题：`FluentProvider`
- 状态：`Zustand` 只管理壳层 UI 与上下文
- 数据：`TanStack Query` + mock / adapter
- 请求层：`Axios` 基础 client + 拦截器骨架

## 实施顺序

1. 先阅读现有 Vue 主线，提取信息架构和页面壳层规律。
2. 在 `frontend-fluentV2/` 落地稳定的应用壳、Provider、路由注册和 mock 数据边界。
3. 逐页迁移内容区实现，不推翻壳层、路由和页面容器。
4. 在壳层稳定后再逐步接入真实 API 和上下文同步。

## 当前非目标

- 不在本期接入真实后端 API
- 不在本期做真实登录
- 不在本期做真实权限校验
- 不在本期做旧页面 1:1 React 翻译
