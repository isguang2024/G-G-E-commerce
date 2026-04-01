# AGENTS.md

## 当前协作约束

- 使用中文沟通。
- 仓库当前处于双前端并行阶段：
  - `frontend/` 为现有 Vue 主线
  - `frontend-fluentV2/` 为 React + Fluent 2 新迁移线
- 当前阶段所有 React 迁移工作默认优先落在 `frontend-fluentV2/`，不得修改现有 `frontend/` 的业务实现，除非明确要求。
- 通过 Shell 读写文本时必须显式使用 UTF-8 编码，避免乱码。
- 前端壳层与页面调整默认遵循根目录 `PROJECT_FRAMEWORK.md` 与 `FRONTEND_GUIDELINE.md`。
- 大型改动收尾时需要同步更新 `docs/change-log.md`。

## 实施原则

- 不复制 Vue 页面代码到 React 工程。
- 先保留信息架构和职责边界，再替换页面实现。
- React 新工程优先复用 Fluent UI React v9 组件能力，不引入其他 UI 库。
- 当前阶段禁止接入真实 API、真实登录和真实权限校验；如需数据，先走 mock 或 adapter 骨架。

## 文档位置

- 根目录三份文档只保留当前有效约束：
  - `AGENTS.md`
  - `PROJECT_FRAMEWORK.md`
  - `FRONTEND_GUIDELINE.md`
- React 新迁移线的正式专题文档放在 `frontend-fluentV2/docs/`。
- `frontend/docs/legacy/` 仅作为归档，不视为当前生效规范。
