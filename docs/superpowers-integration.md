# Superpowers 接入说明

## 接入方式

当前仓库采用项目级接入，而不是修改全局技能目录：

- 技能目录：`.agents/skills/superpowers/`
- 来源仓库：`obra/superpowers`
- 当前快照：`b7a8f76985f1e93e75dd2f2a3b424dc731bd9d37`

这样做的目的有两个：

1. 只让当前仓库启用这套工作流，不影响其他项目。
2. 可以在仓库内对约束做本地覆盖，避免与当前前端主线和协作文档冲突。

## 当前仓库里的落地约束

接入后仍然遵守根目录规范：

- 继续使用中文沟通。
- `frontend/` 仍是当前唯一有效前端主线。
- 当前阶段仍以 mock / adapter 骨架为主，不接真实登录、真实权限和无约束第二前端线。
- 如果 `superpowers` 的流程与根目录 `AGENTS.md`、`PROJECT_FRAMEWORK.md`、`FRONTEND_GUIDELINE.md` 冲突，以这些仓库文档为准。

## 已接入内容

- 完整 `skills/` 快照
- `agents/code-reviewer.md`
- 上游 `LICENSE`

## 使用说明

1. 打开当前仓库的新 Codex 会话，或重启 Codex。
2. 进入仓库后，Codex 会从 `.agents/skills/superpowers/` 发现技能。
3. 之后可以直接在当前仓库里使用 `superpowers` 的流程型技能。

当前仍启用的技能包括：

- `brainstorming`
- `dispatching-parallel-agents`
- `receiving-code-review`
- `requesting-code-review`
- `subagent-driven-development`
- `writing-plans`

当前已停用的技能包括：

- `using-superpowers`
- `using-git-worktrees`
- `executing-plans`
- `finishing-a-development-branch`
- `writing-skills`
- `systematic-debugging`

## 当前仓库的精简流程

### 保留为主流程

- `requesting-code-review`：大功能、跨前后端改动、权限 / 菜单 / App 上下文这类核心链路改动后再触发。
- `dispatching-parallel-agents` / `subagent-driven-development`：前者用于把多个互不相关的问题并行拆开，后者用于把一个已成形的计划按任务执行到底。

当前仓库对 `subagent-driven-development` 的本地化约束是：

- 默认按任务风险选择 review 深度，不再把“实现 + 双 review + 最终收尾”套到每个小任务上。
- 子代理默认不自动 commit；只有任务或用户明确要求时才进入提交或分支交付链。
- 任务完成后先停在“已验证的实现结果”，只有用户明确要 commit / PR / merge / cleanup 时才进入 `finishing-a-development-branch`。

### 条件触发

- `brainstorming`：当任务有明显设计分歧、边界变化或需要先选方案时触发；常见例子包括菜单 / 权限 / App 语义调整、跨前后端契约变化、复杂页面重组。
- `writing-plans`：只在多步、多文件、跨前后端或带迁移 / 接口契约变化时触发。
- `executing-plans`：只在已经有计划文件，且需要按顺序内联执行或批量执行时触发。
- `using-git-worktrees`：只在需要独立分支长期开发、避免和当前脏工作区冲突、明确要 PR / 隔离验证时触发。
- `finishing-a-development-branch`：只在用户明确要提交、开 PR、合并、清理分支时触发。
### 默认不做硬触发

- 不再把 `using-superpowers` 当成每次开口前都必须触发的总入口。
- 不再把 `brainstorming -> writing-plans -> executing-plans` 当成所有任务的默认链路。
- 不再把“每个任务都 code review”“所有任务都先 TDD”当成默认要求。
- 不再把 `finishing-a-development-branch` 当成任何任务完成后的默认尾闸。

### 当前停用策略

- 当前仓库直接停用了 `using-superpowers`、`using-git-worktrees`、`executing-plans`、`finishing-a-development-branch`、`test-driven-development`、`verification-before-completion`、`systematic-debugging`、`writing-skills` 的技能发现入口。
- 停用方式是保留目录内容，但移除 `SKILL.md` 作为发现入口；如需恢复，只要把对应的 `SKILL.disabled.md` 改回 `SKILL.md`。

## 推荐触发条件

- 触发 `brainstorming`：当你已经能看出有多个合理方案、或者改动会影响信息架构 / 边界语义时再用；不必为了所有跨模块任务都先 brainstorming。
- 触发 `writing-plans`：涉及 3 个以上文件且跨模块，或包含数据库 / 迁移 / 接口契约变化。
- 触发 `requesting-code-review`：核心域改动、跨前后端改动、准备提交 PR。

## 适合当前仓库的使用方式

- 小改动：直接读代码 -> 实现 -> 跑最小必要验证 -> 汇报结果。
- 中等改动：先写一个简短执行思路 -> 实现 -> 跑目标验证 -> 必要时补一次 review。
- 大改动或跨前后端：条件触发 `brainstorming -> writing-plans -> subagent-driven-development`。
- Bug / 异常：先做最小复现和定位，再修复并做最小复核。
- 合并 / 交付：当前不再依赖 `superpowers` 的 branch 收尾技能，按实际 git 工作流单独处理。

## 子代理默认档位

当前仓库对 `superpowers` 做了本地化收口，默认不要把子代理开到最高推理：

- 实现型子代理：`gpt-5.4-mini` + `low`
- 规格审查子代理：`gpt-5.4-mini` + `low`
- 代码审查子代理：`gpt-5.4` + `medium`
- 只有跨模块集成、复杂调试、架构判断或高风险 review 时，再升到 `gpt-5.4` + `high`

默认不建议对子代理使用 `xhigh`。

## 注意点

- `superpowers` 会显著提高流程强度，尤其是设计、计划、TDD、review 和收尾步骤。
- 当前 `frontend/package.json` 没有默认 `test` 脚本，因此测试相关验证需要结合当前仓库实际能力执行，不能机械套用。
- 当前全局 Codex 配置已启用 `multi_agent = true`，因此需要子代理的技能具备运行前提。

## 后续更新

若要升级 `superpowers`：

1. 从 `https://github.com/obra/superpowers` 拉取新的上游内容。
2. 用新的 `skills/` 快照覆盖 `.agents/skills/superpowers/`。
3. 同步更新 `.agents/skills/superpowers/UPSTREAM.md` 的 commit hash。
4. 重新打开 Codex 会话，确保技能重新发现。
