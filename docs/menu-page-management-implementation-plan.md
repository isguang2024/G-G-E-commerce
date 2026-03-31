# 菜单与页面管理实施现状

> 基线日期：2026-03-31。本文记录当前已经落地的实现与剩余收口项。

## 已落地后端能力

- `menus.kind` 已显式区分 `directory / entry / external`
- `page_space_bindings` 承载少量独立页空间暴露
- `GET /api/v1/runtime/navigation` 作为运行时统一入口
- 页面运行时过滤菜单直达重复页
- 空间解析与页面空间继承统一收口到后端编译链

## 已落地前端能力

- 路由守卫已切到用户信息、`runtime/navigation` 和动态注册
- 守卫里不再重复做菜单/页面权限裁剪
- 菜单管理弹窗按 `directory / entry / external` 编辑
- 页面管理标题已明确为“受管页面”

## 仍保留的兼容接口

- `/api/v1/menus/tree`
- `/api/v1/pages/runtime`
- `/api/v1/pages/runtime/public`

要求：
- 兼容接口结果必须与新 manifest 保持一致
- 不再新增新的双轨语义

## 仍需关注的边界

- 菜单管理中的“受管页面关系”只允许兼容只读展示
- 页面表单中的 `space_key` 仍是兼容输入口
- 菜单空间页操作仍偏多，后续可以继续收敛

## 最低验证矩阵

- `go test ./...`
- `pnpm exec vue-tsc --noEmit`
- 菜单管理、页面管理、菜单空间各走一轮
- 登录后入口菜单刷新、深链访问、默认首页回退各走一轮
