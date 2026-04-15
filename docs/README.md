# 文档中枢

`docs/` 只做三件事：导航、长期手册、历史归档。

## 入口分工

- 总导航：[INDEX.md](INDEX.md)
- 结构说明：[project-structure.md](project-structure.md)
- API 流程：[API_OPENAPI_FIXED_FLOW.md](API_OPENAPI_FIXED_FLOW.md)
- 专题手册：[guides/README.md](guides/README.md)
- 历史归档：[archive/README.md](archive/README.md)

## 约束

- 入口文档只保留职责、阅读顺序和链接，不重复展开实现细节。
- 同一主题只保留一个主说明文档，其他位置改为跳转。
- 阶段验收、回滚预案、专项审计、故障复盘默认进入 `archive/`，不进入主手册导航。
