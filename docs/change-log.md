# Change Log

用于记录大型代码改动的简要摘要，便于后续交接和继续开发。

## 2026-03-23 初始化收尾规则与日志机制

### 本次改动
- 在仓库根目录新增 `AGENTS.md`，明确中文回答、UTF-8 编码、大型改动收尾格式和日志写入要求。
- 新建 `change-wrapup` 自定义技能，统一代码咨询和大型改动后的收尾总结逻辑。
- 创建 `docs/change-log.md` 作为大型改动固定日志文件，并约定后续在任务完成后追加记录。
- 为技能补充 `agents/openai.yaml`，完善技能元数据与触发描述。

### 下次方向
- 在后续真实的大型代码改动任务中，按既定格式持续追加日志，验证规则是否稳定执行。
- 如有需要，可进一步细化日志模板，例如补充影响文件、测试状态、风险等级等字段。

## 2026-03-23 团队边界拆分为功能包展开与手工补充

### 本次改动

- 新增 `team_manual_action_permissions`，把团队手工补充权限从 `tenant_action_permissions` 中拆出来单独存储。
- 调整功能包同步、团队权限保存、权限删除、团队删除、边界来源查询等链路，最终仍汇总回 `tenant_action_permissions` 作为当前生效缓存。
- 更新权限设计文档 [permission-package-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-package-design.md)，删除已完成阶段，改为现状说明。
- 已完成验证：`go test ./...`、`pnpm exec vue-tsc --noEmit`、`pnpm build`。

### 下次方向

- 继续把团队角色/成员分配完全收口到“仅可分配已开通能力”。
- 评估是否将运行时鉴权从直接读取 `tenant_action_permissions` 进一步封装为统一的团队边界聚合服务。

## 2026-03-23 权限模型现状收口与团队成员分配语义整理

### 本次改动

- 调整团队成员角色分配弹窗，明确区分“基础角色”和“团队自定义角色”，并按角色来源排序展示。
- 清理成员功能权限弹窗中的旧分类文案，使页面语义与当前 `permission_key + context_type + 功能包` 模型一致。
- 重写 [permission-package-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-package-design.md)，删除已过时阶段描述，改为“当前已完成 / 正式规则 / 未完成事项”现状文档。
- 新增 [permission-overall-summary.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-overall-summary.md)，总结整套权限改造的完成面、当前正式模型和下一步优先级。
- 已完成验证：`pnpm exec vue-tsc --noEmit`、`pnpm build`。

### 下次方向

- 继续把 `tenant_action_permissions` 从当前生效边界表收口为真正的团队边界缓存。
- 在团队角色和团队成员分配页面进一步显式展示“功能包来源 / 团队补充来源”，把前端交互完全收口到“仅可分配已开通能力”。
