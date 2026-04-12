# PROJECT_STRUCTURE

## 1. 报告目的
- 记录阶段 7 收口时的项目结构快照。
- 给出本轮优化后的目录导航与变更统计。

## 2. 顶层结构（快照）
```text
G-G-E-commerce/
├─ .claude/
├─ .codex/
├─ .codex-tmp/
├─ backend/
├─ docs/
├─ frontend/
├─ output/
├─ AGENTS.md
├─ CHANGELOG.md
├─ FRONTEND_GUIDELINE.md
├─ PROJECT_FRAMEWORK.md
├─ README.md
├─ start-backend.bat
└─ start-frontend.bat
```

## 3. 核心子结构

### 3.1 前端 `frontend/src/`
- `api`
- `assets`
- `components`
- `composables`
- `config`
- `directives`
- `domains`
- `enums`
- `hooks`
- `locales`
- `mock`
- `plugins`
- `router`
- `store`
- `types`
- `utils`
- `views`

### 3.2 后端 `backend/internal/`
- `api`
- `config`
- `modules`
- `pkg`

### 3.3 文档 `docs/`
- `guides/`
- `reports/`
- `GUIDELINES.md`
- `INDEX.md`
- `README.md`
- `V5_REFACTOR_TASKS.md`

## 4. 指标快照（当前工作区）
- 文件总数（过滤 `.git/node_modules/dist/tmp/output`）：`5517`
- Markdown 文件数：`154`
- TypeScript/Vue 文件数（`.ts/.tsx/.vue`）：`2936`
- Go 文件数：`1450`

## 5. 本轮优化后的变更统计（基于 `git status --short`）
- 新增：`12`
- 修改：`8`
- 删除：`10`

## 6. 前后对比结论
- 文档体系从“混合存放”转为“导航中枢 + 指南 + 报告”三段式结构。
- 历史/重复文档已清理，迁移到 `.claude/Instructions/` 的功能与流程文档已纳入导航。
- 当前结构已具备可维护性：入口统一、目录职责清晰、阶段报告可追踪。
