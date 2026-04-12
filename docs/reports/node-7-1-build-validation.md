# Node 7.1 构建验证报告

## 执行时间
- 2026-04-13

## 执行命令
- `pnpm run build`（cwd: `frontend/`）
- `go test ./...`（cwd: `backend/`）
- `go build ./cmd/server`（cwd: `backend/`）
- `go test ./internal/api/handlers -count=1`（cwd: `backend/`）

## 结果摘要
- 前端构建通过（含 `vue-tsc --noEmit`）。
- 后端服务可构建（`go build ./cmd/server` 通过）。
- 后端关键 handler 测试通过。
- 后端全量测试未全绿，存在历史失败：
  - `internal/modules/system/navigation` 测试编译失败（`stubAppService` 缺失 `GetAppPreflight`）。
  - `internal/modules/system/permission` 3 个测试断言失败。

## 结论
- 构建链路可执行，前后端核心构建能力正常。
- 当前仓库存在既有后端全量测试失败，需在后续专项中修复后再达成“全量测试通过”。
