# docs 文档索引

> 现状基线：2026-03-31。`docs/` 只保留当前有效专题、固定清单与少量关键里程碑。

## 1. 核心入口

1. [PROJECT_FRAMEWORK.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/PROJECT_FRAMEWORK.md)：项目主框架、边界与执行清单
2. [permission-overall-summary.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/permission-overall-summary.md)：权限、功能包、API、快照主链
3. [menu-page-management-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/menu-page-management-design.md)：菜单与页面正式架构
4. [space-host-architecture-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/space-host-architecture-design.md)：空间与 Host 架构
5. [message-system-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/message-system-design.md)：消息系统设计
6. [system-regression-checklist.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/system-regression-checklist.md)：系统固定回归项
7. [system-demo-data.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/system-demo-data.md)：系统演示数据
8. [change-log.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/change-log.md)：关键里程碑

## 2. 保留原则

- 保留正式设计文档：定义稳定模型与边界
- 保留当前实现文档：记录已落地状态、兼容边界与最小验证
- 保留固定清单文档：回归、演示数据、长期可复用检查项
- 保留少量里程碑日志：仅承载近期关键收口

## 3. 删除原则

- 阶段性收尾计划：结论已稳定后，合并进主框架或专题文档，再删除
- 兼容审计摘要：如果没有新增长期约束，应并入实现文档，不单独保留
- 重构基线摘要：如果已经被正式设计和实施现状覆盖，应优先合并后删除
- 重复索引和重复总结：只保留一个主入口，避免同义文档并存

## 4. 维护规则

- 文档与代码冲突时，以代码真实行为为准，然后更新文档
- 新增核心链路变更时，必须同步修改对应专题文档
- 删除或改名文档时，必须同步清理根文档、专题文档和 `change-log.md` 中的旧链接

