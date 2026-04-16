## batch_create_stages 参数约束不直观
**问题**: `task_tree_batch_create_stages` 的 `acceptance_criteria` 需要 `string[]`，但技能文档在阶段示例里没有明确展示这一点，首次调用时容易按字符串传参导致失败。
**期望**: 在 `task-tree` 技能文档的阶段创建示例里明确给出 `acceptance_criteria: ["..."]` 的格式。
**建议**: 在“阶段管理”或“常见陷阱”中补一条：`acceptance_criteria` 为数组，不是单字符串。
