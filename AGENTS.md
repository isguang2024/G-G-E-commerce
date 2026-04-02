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
- 若存在 `frontend-fluentV2/docs/下次方向记录.md`，则未完成的后续方向应持续维护在该文件中，按条目增删，不按每轮“下次方向”整体覆盖；全部完成后直接删除该文件。

## 实施原则

- 不复制 Vue 页面代码到 React 工程。
- 先保留信息架构和职责边界，再替换页面实现。
- React 新工程优先复用 Fluent UI React v9 组件能力，不引入其他 UI 库。
- 当前阶段禁止接入真实 API、真实登录和真实权限校验；如需数据，先走 mock 或 adapter 骨架。
- 当代码修改需要调整数据库结构或初始化数据时，允许先补一个临时迁移并执行验证；若迁移已完成且当前数据库已达成目标状态，后续应删除该临时迁移，避免长期保留一次性修复脚本。
- 默认初始化数据优先通过 seed / ensure 逻辑完善，迁移只负责一次性结构变更或历史数据修正，不把长期默认状态反复写进迁移链。

## 文档位置

- 根目录三份文档只保留当前有效约束：
  - `AGENTS.md`
  - `PROJECT_FRAMEWORK.md`
  - `FRONTEND_GUIDELINE.md`
- React 新迁移线的正式专题文档放在 `frontend-fluentV2/docs/`。
- `frontend/docs/legacy/` 仅作为归档，不视为当前生效规范。
