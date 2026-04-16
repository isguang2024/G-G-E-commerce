package handlers

import (
	"context"
	"encoding/json"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/observability/telemetry"
	"github.com/maben/backend/internal/pkg/logger"
)

// IngestTelemetryLogs 实现前端日志批量上报端点。
//
// 策略：
//   - 服务端即使遇到解析失败或限流也不返回 4xx：前端只要一次 4xx 就可能
//     把同一批条目重试，造成雪崩。服务端只在彻底无法写入时返回 accepted=0,
//     dropped=N，前端 sendBeacon/fetch keepalive 也不会再重试。
//   - 正文由 ogen 校验（maxItems=100、字段必填）；这里只做模型转换。
//   - session_id / IP 作为限流 key：前端同 tab 同会话只有一个 session_id；
//     IP 用 ctx 中 client_ip（代理剥离后的真实来源）。
func (h *telemetryAPIHandler) IngestTelemetryLogs(ctx context.Context, req *gen.TelemetryIngestRequest) (gen.IngestTelemetryLogsRes, error) {
	if req == nil || len(req.Entries) == 0 {
		return &gen.TelemetryIngestResponse{Accepted: 0, Dropped: 0}, nil
	}

	entries := make([]telemetry.Entry, 0, len(req.Entries))
	sessionKey := ""
	for i := range req.Entries {
		src := req.Entries[i]
		if sessionKey == "" {
			sessionKey = src.SessionID
		}
		entry := telemetry.Entry{
			Level:     string(src.Level),
			Event:     src.Event,
			Timestamp: src.Timestamp,
			RequestID: src.RequestID.Or(""),
			Route:     src.Route.Or(""),
			UserID:    src.UserID.Or(""),
			SessionID: src.SessionID,
			UserAgent: src.UserAgent,
			Viewport: telemetry.Viewport{
				W: src.Viewport.W,
				H: src.Viewport.H,
			},
		}
		if src.Context.Set {
			entry.Context = mapAnyToContext(src.Context.Value)
		}
		if src.Error.Set {
			entry.Error = &telemetry.ErrorSnapshot{
				Name:    src.Error.Value.Name.Or(""),
				Message: src.Error.Value.Message.Or(""),
				Stack:   src.Error.Value.Stack.Or(""),
			}
		}
		entries = append(entries, entry)
	}

	ipKey := logger.ClientIPFromContext(ctx)
	res := h.telemetry.Ingest(ctx, entries, sessionKey, ipKey)
	return &gen.TelemetryIngestResponse{
		Accepted: res.Accepted,
		Dropped:  res.Dropped,
	}, nil
}

// mapAnyToContext 把 ogen 生成的 jx.Raw map 解码成 map[string]any。
// TelemetryLogEntryContext 本身是 map[string]jx.Raw（additionalProperties:true 的表达），
// 每个 value 是未反序列化的 JSON bytes。我们把它解成 any，入口模块再统一脱敏。
// 解码失败的字段直接丢弃，不让一条脏数据废掉整条 entry。
func mapAnyToContext(v gen.TelemetryLogEntryContext) map[string]any {
	if len(v) == 0 {
		return nil
	}
	out := make(map[string]any, len(v))
	for k, raw := range v {
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			continue
		}
		out[k] = decoded
	}
	return out
}

