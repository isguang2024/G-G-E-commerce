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

常见技能包括：

- `using-superpowers`
- `brainstorming`
- `writing-plans`
- `systematic-debugging`
- `verification-before-completion`
- `requesting-code-review`

## 注意点

- `superpowers` 会显著提高流程强度，尤其是设计、计划、TDD、review 和收尾步骤。
- 当前 `frontend/package.json` 没有默认 `test` 脚本，因此涉及 `test-driven-development` 时，需要结合当前仓库实际能力执行，不能机械套用。
- 当前全局 Codex 配置已启用 `multi_agent = true`，因此需要子代理的技能具备运行前提。

## 后续更新

若要升级 `superpowers`：

1. 从 `https://github.com/obra/superpowers` 拉取新的上游内容。
2. 用新的 `skills/` 快照覆盖 `.agents/skills/superpowers/`。
3. 同步更新 `.agents/skills/superpowers/UPSTREAM.md` 的 commit hash。
4. 重新打开 Codex 会话，确保技能重新发现。
