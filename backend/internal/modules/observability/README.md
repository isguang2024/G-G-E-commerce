# backend/internal/modules/observability

`observability/` 负责后端可观测性能力，当前拆成业务审计日志和前端日志摄取两条链路。

## 子目录

| 目录 | 说明 |
| --- | --- |
| `audit/` | 业务审计事件写入、脱敏、异步入库、运行时统计 |
| `telemetry/` | 前端批量日志摄取、限流、二次脱敏、入库统计 |

## 当前边界

- `audit/` 面向后端业务操作，调用方通过 `Recorder.Record(ctx, Event)` 写入审计事件。
- `telemetry/` 面向前端 `/telemetry/logs` 上报，调用方通过 `Ingester.Ingest(...)` 批量接收日志。
- 两个子目录都负责“写入前兜底脱敏”和“非阻塞/限流保护”，不让观测链路反向拖垮主流程。

## 阅读入口

- `audit/service.go`：审计链路的主入口与队列模型
- `telemetry/service.go`：前端日志摄取入口与限流逻辑
