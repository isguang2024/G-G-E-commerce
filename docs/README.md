# docs 文档索引

> 现状基线：2026-03-27。`docs/` 只保留当前有效架构说明与近期里程碑，不再保留阶段性草案堆叠。

## 1. 推荐阅读顺序

1. [PROJECT_FRAMEWORK.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/PROJECT_FRAMEWORK.md)：项目主框架、边界与执行清单
2. [permission-overall-summary.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-overall-summary.md)：权限、功能包、快照主链
3. [menu-page-management-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/menu-page-management-design.md)：菜单与页面正式架构
4. [menu-page-management-implementation-plan.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/menu-page-management-implementation-plan.md)：菜单与页面当前实现现状
5. [FRONTEND_GUIDELINE.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/FRONTEND_GUIDELINE.md)：前端管理页统一规范
6. [change-log.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/change-log.md)：近期里程碑

## 2. 文档分工

- 根目录文档：全局长期规则
- `permission-overall-summary.md`：权限、功能包、API、快照的正式语义
- `menu-page-management-design.md`：菜单与页面的稳定模型
- `menu-page-management-implementation-plan.md`：当前已落地实现、剩余边界、实施注意点
- `change-log.md`：只记录近期有效里程碑，不记碎片化推进过程

## 3. 维护规则

- 文档与代码冲突时，以代码真实行为为准，然后更新文档。
- 新增核心链路变更时，必须同步修改对应专题文档。
- 重复、过期、阶段性说明要合并后删除，避免多版本并存。
- 删除或改名文档时，必须同步检查根文档、专题文档和变更日志中的旧链接。

## 4. 变更日志保留策略

- `change-log.md` 只保留近期里程碑与当前架构仍有参考价值的收口记录。
- 过细过程日志、同日重复推进日志、已经被专题文档吸收的旧记录，应合并后移除。
- 需要长期保留的结论，写回专题文档，不依赖日志存活。
