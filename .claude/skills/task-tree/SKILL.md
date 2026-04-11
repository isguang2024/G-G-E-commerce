# task-tree

跨会话持久任务树。任何跨轮、跨会话、5 步以上、或需要续跑的工作必须使用；不要在聊天里维护 TODO。通过本地 HTTP MCP `http://127.0.0.1:8879/mcp` 调用。

## 触发规则

按顺序判断，只要命中一条就必须用任务树：

1. 用户给了 `tsk_...`。
2. 用户说“继续上次那个”“接着之前的”。
3. 你正准备在聊天里写 TODO。
4. 工作明显不少于 5 步，或需要跨轮对话。
5. 要跨会话续跑。

单轮能做完的简单问答、一次性 edit、单个小 bug fix，不强制使用。

如果 MCP 不可用，要明确告诉用户“任务树服务未启动”，不能静默退回聊天 TODO。

## 标准执行流

拿到 `task_id` 后按这个顺序：

1. `task_tree_resume(task_id, view_mode="summary", filter_mode="focus")`
2. 看 `next_node`
3. 看 `next_node.siblings`
4. 看 `recent_events`
5. `task_tree_claim(node_id)` 后再开始执行
6. 中途用 `task_tree_progress`
7. 完成用 `task_tree_complete`
8. 再次 `resume` 取下一个节点

默认先读摘要，不要把全量树和全量事件当常规入口。

## 领取即标记（新增硬规则）

1. 开始任何节点前，必须先 `task_tree_claim(node_id)`，让节点进入 `running`。
2. 不允许先做实现、最后才补 claim 或状态更新。
3. 如果发现节点已经在执行但未 claim，必须立刻停下补 claim，再继续执行。
4. 若 claim 失败（被占用/过期），先处理状态一致性（重试 claim 或切换节点），再开始代码改动。
5. 完成时再 `task_tree_complete`；claim/complete 之间可按需用 `task_tree_progress`。

## 节点拆分

- 默认尽量多拆子节点，不要把多个并列动作塞进一条 instruction。
- 复杂任务至少先拆出 3 到 6 个节点，必要时继续向下拆到可直接执行。
- 节点一旦更像“容器”而不是“执行项”，就继续补子节点。
- 优先用树表达工作分解，不用聊天消息维护计划。

## progress / complete message

`progress` 和 `complete` 的 message 必须写四段：

```text
做了什么:
- ...

证据:
- ...

偏差:
- ...

遗留:
- ...
```

禁止只写 `done`、`ok`、`完成` 这类空消息。

## 遗留回灌规则

这是硬规则：

1. 每次读取 `complete` 或 `progress` 事件时，都要检查 `遗留:` 段。
2. 仍然有效的遗留，不能只停留在事件消息里，必须回灌到任务树。
3. 优先把遗留映射到现有的直接承接节点：
   - 更新该直接节点的 `instruction`
   - 必要时补 `acceptance_criteria`
4. 如果树上没有直接承接节点，必须在对应父节点下新增直接叶子节点。
5. 不要把一个遗留笼统挂到更高层 group；要尽量落到可直接执行的叶子节点。
6. 如果遗留已经被后续节点覆盖，要在本轮说明“已映射到哪个节点”，不要重复建树。

简化判断：

- 有直接节点可承接：更新节点。
- 没有直接节点可承接：新增直接叶子节点。
- 已被现有节点覆盖：记录映射关系，不重复创建。

## 仓库约束

- 本仓库后端在 `backend/`，前端在 `frontend/`。
- 涉及 API 契约变更，必须遵守 `AGENTS.md` 的 OpenAPI-first 链路。
- 涉及生成物时，不手改 `backend/api/gen/` 和 `frontend/src/api/v5/` 生成文件本体。
- Shell 读写文本必须显式 UTF-8。

## 推荐做法

- 任务详情默认用 `task_tree_resume(..., view_mode="summary", filter_mode="focus")`
- 批量浏览节点优先用 `task_tree_list_nodes_summary`
- 查遗留、查 warnings、查完成记录优先用 `task_tree_list_events`
- 要改 instruction / 验收标准时，再用 `task_tree_get_node` 或 `task_tree_update_node`

## 完成标准

一次合格的任务树维护，至少满足：

- 当前可执行节点清晰
- 遗留没有悬空在历史消息里
- instruction 和 acceptance_criteria 能直接指导下一步执行
- 最近一次 complete message 有足够证据，不是空话
