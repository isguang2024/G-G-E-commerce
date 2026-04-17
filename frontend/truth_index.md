# frontend/truth_index.md

> 前端开发真相详细索引。

## 真相文档一览

| 文档 | 位置 | 触发场景 |
| --- | --- | --- |
| 前端实现规范 | [Truth/frontend-guideline.md](./Truth/frontend-guideline.md) | 数据接口、视觉方向、壳层、页面组织、Pinia/路由约束 |
| 可观测性规范 | [Truth/observability-spec.md](./Truth/observability-spec.md) | `data-testid` 命名、表单错误语义、结构化 error 扩展 |
| 上传 SDK 指南 | [Truth/upload-sdk.md](./Truth/upload-sdk.md) | 前端如何通过 UploadKey 调用上传接口（场景示例） |

## 按场景找文档

| 你要做什么 | 看哪份 |
| --- | --- |
| 新写一个页面 | `frontend-guideline.md` |
| 加 E2E 测试钩子 | `observability-spec.md` |
| 接一个上传组件 | `upload-sdk.md`（后端架构见 `backend/Truth/upload-system/overview.md`）|
| 处理结构化 error | `observability-spec.md` |
