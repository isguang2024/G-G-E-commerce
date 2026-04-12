# .claude 目录说明

该目录用于存放协作运行时配置与项目级辅助内容。

## 当前内容

- `launch.json`：本地协作启动配置
- `settings.local.json`：本地会话设置
- `skills/`：项目内保留的技能文件（如 `change-wrapup`）
- `Instructions/`：功能说明与流程说明文档（面向执行与迁移）
- `worktrees/`：多 agent 协作产生的工作副本（按需清理）

## 维护约定

- `worktrees/` 仅作为协作临时副本，不作为业务源码真相源。
- 变更 `.claude` 配置前，需确认不会影响当前活跃会话。
