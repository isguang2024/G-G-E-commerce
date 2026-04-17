# frontend/truth.md

> 前端开发真相摘要。详细条目见 [truth_index.md](./truth_index.md) 和 [Truth/](./Truth/)。

## 核心铁律

- **已接真实接口**，不再使用 mock。
- **统一走 `v5Client`**：生成 schema 在 `src/api/v5/`（`client.ts` / `types.ts` / `schema.d.ts` / `error-codes.ts`）；业务封装写在 `src/api/*.ts` 与 `src/api/system-manage/*.ts`。**禁止**新增第二套请求体系、手写 axios DTO、或直接改生成文件。
- **OpenAPI 变更同步**：后端契约变更后必须执行 `pnpm run gen:api` 并修正调用层。
- **权限来自后端接口**：路由守卫与权限判断消费 `/permissions/explain` 等真实接口，不在前端手写权限规则。
- **前端不感知 `tenant_id`**：由后端中间件解析。
- **基座组件不 import**：`src/components/core/` 下 `Art` 前缀组件通过 `registerGlobalComponent.ts` 自动注册；业务组件在 `src/components/business/` 需要 import。
- **枚举走 DictSelect**：禁止硬编码 `<ElOption>` 列表。

## 视觉原则

- 克制、清晰、低噪。先用留白/层级/排版组织信息，不堆卡片皮肤。
- 后台场景优先稳定感与可读性。

## 关键目录

- `src/api/v5/` — OpenAPI 生成产物（只读）
- `src/api/*.ts`、`src/api/system-manage/*.ts` — 业务 API 封装
- `src/components/core/` — 基座全局组件（免 import）
- `src/components/business/` — 业务组件（需 import）
- `src/views/` — 正式页面
- `src/domains/upload/` — 上传 SDK

## 入口

- AI 协作约束：[../AGENTS.md](../AGENTS.md)
- 详细真相索引：[truth_index.md](./truth_index.md)
- 全部真相文档：[Truth/](./Truth/)
