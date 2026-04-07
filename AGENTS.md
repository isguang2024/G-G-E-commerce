# AGENTS.md

## 当前协作约束

- 使用中文沟通。
- 仓库已接入项目级 `superpowers` 技能集，路径为 `.agents/skills/superpowers/`；若技能流程与本仓库约束冲突，以本文件及根目录规范为准。
- 通过 Shell 读写文本时必须显式使用 UTF-8 编码，避免乱码。
- 前端壳层与页面调整默认遵循根目录 `PROJECT_FRAMEWORK.md` 与 `FRONTEND_GUIDELINE.md`。
- 大型改动收尾时需要同步更新 `docs/change-log.md`。

## 实施原则
- 代码修改不使用rg
- 当代码修改需要调整数据库结构或初始化数据时，允许先补一个临时迁移并执行验证；若迁移已完成且当前数据库已达成目标状态，后续应删除该临时迁移，避免长期保留一次性修复脚本。
- 默认初始化数据优先通过 seed / ensure 逻辑完善，迁移只负责一次性结构变更或历史数据修正，不把长期默认状态反复写进迁移链。

## 文档位置

- 根目录当前生效的协作文档只有三份：
  - `AGENTS.md`
  - `PROJECT_FRAMEWORK.md`
  - `FRONTEND_GUIDELINE.md`
- `docs/archive/frontend/` 仅作为 `frontend` 历史文档归档，不视为当前生效规范。
