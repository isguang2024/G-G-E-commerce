// observability.go: /observability/* handler 实现。
//
// audit_logs / telemetry_logs 只支持只读查询，写入链路由异步 Recorder / Ingester
// 负责。本 handler 直接用 h.db 查询，不新增 service 层：
//   - 查询逻辑简单（where + order by ts + paginate），没有跨表 join；
//   - 读路径不写审计（否则查询流量会反向炸 audit_logs）；
//   - 所有端点都需要登录 + `observability.{audit|telemetry}.read` 权限；
//     权限由 ogen 中间件按 x-permission-key 校验，handler 侧仅兜底 401。
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-faster/jx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/telemetry"
	"github.com/maben/backend/internal/pkg/logger"
)

const (
	observabilityListDefaultCurrent = 1
	observabilityListDefaultSize    = 20
	observabilityListMaxSize        = 200
)

// ─── audit_logs ─────────────────────────────────────────────────────────────

func (h *observabilityAPIHandler) ListAuditLogs(ctx context.Context, params gen.ListAuditLogsParams) (gen.ListAuditLogsRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.ListAuditLogsUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	current, size := observabilityPagination(params.Current, params.Size)
	q := h.db.WithContext(ctx).
		Model(&audit.AuditLog{}).
		Where("tenant_id = ?", logger.TenantFromContext(ctx))
	if v := optString(params.Action); v != "" {
		q = q.Where("action = ?", v)
	}
	if v := optString(params.ActorID); v != "" {
		q = q.Where("actor_id = ?", v)
	}
	if v := optString(params.Outcome); v != "" {
		q = q.Where("outcome = ?", v)
	}
	if v := optString(params.ResourceType); v != "" {
		q = q.Where("resource_type = ?", v)
	}
	if v := optString(params.ResourceID); v != "" {
		q = q.Where("resource_id = ?", v)
	}
	if v := optString(params.RequestID); v != "" {
		q = q.Where("request_id = ?", v)
	}
	if params.From.Set {
		q = q.Where("ts >= ?", params.From.Value)
	}
	if params.To.Set {
		q = q.Where("ts <= ?", params.To.Value)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		h.logger.Error("list audit logs count failed", zap.Error(err))
		return nil, err
	}
	var rows []audit.AuditLog
	if err := q.Order("ts DESC, id DESC").
		Offset((current - 1) * size).
		Limit(size).
		Find(&rows).Error; err != nil {
		h.logger.Error("list audit logs failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.AuditLogItem, 0, len(rows))
	for i := range rows {
		records = append(records, auditLogItemFromModel(&rows[i]))
	}
	return &gen.AuditLogList{
		Records: records,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (h *observabilityAPIHandler) GetAuditLog(ctx context.Context, params gen.GetAuditLogParams) (gen.GetAuditLogRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetAuditLogUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	var row audit.AuditLog
	if err := h.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", params.ID, logger.TenantFromContext(ctx)).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.GetAuditLogNotFound{Code: 404, Message: "审计日志不存在"}, nil
		}
		h.logger.Error("get audit log failed", zap.Error(err))
		return nil, err
	}
	detail := auditLogDetailFromModel(&row)
	return &detail, nil
}

// ─── telemetry_logs ─────────────────────────────────────────────────────────

func (h *observabilityAPIHandler) ListTelemetryLogs(ctx context.Context, params gen.ListTelemetryLogsParams) (gen.ListTelemetryLogsRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.ListTelemetryLogsUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	current, size := observabilityPagination(params.Current, params.Size)
	q := h.db.WithContext(ctx).
		Model(&telemetry.TelemetryLog{}).
		Where("tenant_id = ?", logger.TenantFromContext(ctx))
	if v := optString(params.Level); v != "" {
		q = q.Where("level = ?", v)
	}
	if v := optString(params.Event); v != "" {
		q = q.Where("event = ?", v)
	}
	if v := optString(params.SessionID); v != "" {
		q = q.Where("session_id = ?", v)
	}
	if v := optString(params.ActorID); v != "" {
		q = q.Where("actor_id = ?", v)
	}
	if v := optString(params.RequestID); v != "" {
		q = q.Where("request_id = ?", v)
	}
	if params.From.Set {
		q = q.Where("ts >= ?", params.From.Value)
	}
	if params.To.Set {
		q = q.Where("ts <= ?", params.To.Value)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		h.logger.Error("list telemetry logs count failed", zap.Error(err))
		return nil, err
	}
	var rows []telemetry.TelemetryLog
	if err := q.Order("ts DESC, id DESC").
		Offset((current - 1) * size).
		Limit(size).
		Find(&rows).Error; err != nil {
		h.logger.Error("list telemetry logs failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.TelemetryLogRecord, 0, len(rows))
	for i := range rows {
		records = append(records, telemetryLogRecordFromModel(&rows[i]))
	}
	return &gen.TelemetryLogList{
		Records: records,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

// GetAuditLogStats 按 group_by 对 audit_logs 做聚合统计。dashboard / 运维小图调用。
//   - group_by=hour   : 按 date_trunc('hour', ts) 聚合，桶按时间升序；
//   - group_by=action : 按 action 聚合，桶按 count 降序；
//   - group_by=outcome: 按 outcome 聚合，桶按 count 降序。
//
// 非时间维度 limit 默认 20 / 最大 100；hour 维度忽略 limit（时间段由 from/to 自然限制）。
// 空区间 buckets 返回 []（非 null）。tenant_id 硬过滤。
func (h *observabilityAPIHandler) GetAuditLogStats(ctx context.Context, params gen.GetAuditLogStatsParams) (gen.GetAuditLogStatsRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetAuditLogStatsUnauthorized{Code: 401, Message: "未认证"}, nil
	}

	tenant := logger.TenantFromContext(ctx)
	q := h.db.WithContext(ctx).
		Model(&audit.AuditLog{}).
		Where("tenant_id = ?", tenant)

	// 默认时间窗兜底：缺省 from → now()-30d；缺省 to → now()。
	// 规避 group_by=hour 缺 from/to 时对 audit_logs 做全表扫（生产几千万行会打爆 DB）。
	// Spec 里参数仍是可选，只在 handler 侧兜底；调用方想要更大窗口请显式传 from/to。
	// 有了始终存在的 ts 上下界后，查询稳定命中 idx_audit_logs_tenant_ts
	// (tenant_id, ts DESC)，EXPLAIN 走 Index Scan 而非 Seq Scan。
	defaultStatsWindow := 30 * 24 * time.Hour
	now := time.Now().UTC()
	fromTS := now.Add(-defaultStatsWindow)
	if params.From.Set {
		fromTS = params.From.Value
	}
	toTS := now
	if params.To.Set {
		toTS = params.To.Value
	}
	q = q.Where("ts >= ?", fromTS).Where("ts <= ?", toTS)

	type rawRow struct {
		Bucket string
		Count  int64
	}
	var rows []rawRow

	switch params.GroupBy {
	case gen.GetAuditLogStatsGroupByAction:
		limit := statsBucketLimit(params.Limit)
		if err := q.
			Select("COALESCE(action, '') AS bucket, COUNT(*) AS count").
			Group("action").
			Order("count DESC, bucket ASC").
			Limit(limit).
			Scan(&rows).Error; err != nil {
			h.logger.Error("audit stats: group by action failed", zap.Error(err))
			return nil, err
		}
	case gen.GetAuditLogStatsGroupByOutcome:
		limit := statsBucketLimit(params.Limit)
		if err := q.
			Select("COALESCE(outcome, '') AS bucket, COUNT(*) AS count").
			Group("outcome").
			Order("count DESC, bucket ASC").
			Limit(limit).
			Scan(&rows).Error; err != nil {
			h.logger.Error("audit stats: group by outcome failed", zap.Error(err))
			return nil, err
		}
	case gen.GetAuditLogStatsGroupByHour:
		// 用 to_char 把 date_trunc 结果格式化成 ISO8601 字符串，让三个维度共用
		// 同一个 bucket:string + count:int64 输出形状。时区统一按 UTC 呈现。
		type timeRow struct {
			Bucket time.Time
			Count  int64
		}
		var tRows []timeRow
		if err := q.
			Select("date_trunc('hour', ts AT TIME ZONE 'UTC') AS bucket, COUNT(*) AS count").
			Group("date_trunc('hour', ts AT TIME ZONE 'UTC')").
			Order("bucket ASC").
			Scan(&tRows).Error; err != nil {
			h.logger.Error("audit stats: group by hour failed", zap.Error(err))
			return nil, err
		}
		rows = make([]rawRow, 0, len(tRows))
		for _, r := range tRows {
			rows = append(rows, rawRow{
				Bucket: r.Bucket.UTC().Format(time.RFC3339),
				Count:  r.Count,
			})
		}
	default:
		return &gen.GetAuditLogStatsBadRequest{Code: 400, Message: "不支持的 group_by"}, nil
	}

	buckets := make([]gen.AuditLogStatsBucket, 0, len(rows))
	for _, r := range rows {
		buckets = append(buckets, gen.AuditLogStatsBucket{
			Bucket: r.Bucket,
			Count:  r.Count,
		})
	}
	return &gen.AuditLogStats{
		GroupBy: gen.AuditLogStatsGroupBy(params.GroupBy),
		Buckets: buckets,
	}, nil
}

func statsBucketLimit(in gen.OptInt) int {
	v := optInt(in, 20)
	if v <= 0 {
		v = 20
	}
	if v > 100 {
		v = 100
	}
	return v
}

// GetObservabilityTrace 按 request_id 把同一次请求下的 audit + telemetry 两张
// 表折叠返回。权限按 audit.read 控制（见 paths.yaml 注释），即使调用方没有
// telemetry.read 权限，同一 request_id 的 telemetry 行也会返回，因为它们属于审计
// 现场的从属信息。两侧都按 ts ASC 排序，前端按时间线渲染。
func (h *observabilityAPIHandler) GetObservabilityTrace(ctx context.Context, params gen.GetObservabilityTraceParams) (gen.GetObservabilityTraceRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetObservabilityTraceUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	requestID := params.RequestID
	tenant := logger.TenantFromContext(ctx)

	var auditRows []audit.AuditLog
	if err := h.db.WithContext(ctx).
		Where("tenant_id = ? AND request_id = ?", tenant, requestID).
		Order("ts ASC, id ASC").
		Limit(observabilityListMaxSize).
		Find(&auditRows).Error; err != nil {
		h.logger.Error("get observability trace: audit query failed",
			zap.String("request_id", requestID), zap.Error(err))
		return nil, err
	}

	var telemetryRows []telemetry.TelemetryLog
	if err := h.db.WithContext(ctx).
		Where("tenant_id = ? AND request_id = ?", tenant, requestID).
		Order("ts ASC, id ASC").
		Limit(observabilityListMaxSize).
		Find(&telemetryRows).Error; err != nil {
		h.logger.Error("get observability trace: telemetry query failed",
			zap.String("request_id", requestID), zap.Error(err))
		return nil, err
	}

	auditItems := make([]gen.AuditLogItem, 0, len(auditRows))
	for i := range auditRows {
		auditItems = append(auditItems, auditLogItemFromModel(&auditRows[i]))
	}
	telemetryItems := make([]gen.TelemetryLogRecord, 0, len(telemetryRows))
	for i := range telemetryRows {
		telemetryItems = append(telemetryItems, telemetryLogRecordFromModel(&telemetryRows[i]))
	}
	return &gen.ObservabilityTraceBundle{
		RequestID:     requestID,
		AuditLogs:     auditItems,
		TelemetryLogs: telemetryItems,
	}, nil
}

// GetObservabilityMetrics 暴露 audit.Recorder / telemetry.Ingester 的运行时指标。
//
// 设计说明：
//  1. 数据源是进程内 Recorder / Ingester 的 Stats()，不读 DB——延迟 < 1ms；
//  2. 不按 tenant 分桶：这是基础设施级指标，进程内没有隔离，调用方看到的
//     queue_depth / dropped_total 是整机维度；需要多副本聚合时由 scraper 负责；
//  3. 权限复用 observability.audit.read：能看审计的人能看队列健康状况，不再
//     单独维护一个 metrics permission key；
//  4. 认证层走 ogen 中间件，handler 侧只兜底 401；Stats() 本身是 O(1) + 只加
//     一次锁，可以安全高频 scrape（Grafana / Prometheus agent 30s 一次）。
func (h *observabilityAPIHandler) GetObservabilityMetrics(ctx context.Context) (gen.GetObservabilityMetricsRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetObservabilityMetricsUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	auditStats := h.audit.Stats()
	telemetryStats := h.telemetry.Stats()
	return &gen.ObservabilityMetrics{
		Audit: gen.ObservabilityServiceStats{
			QueueDepth:           auditStats.QueueDepth,
			QueueCap:             auditStats.QueueCap,
			AcceptedTotal:        int64(auditStats.AcceptedTotal),
			DroppedTotal:         int64(auditStats.DroppedTotal),
			PolicyDroppedTotal:   int64(auditStats.PolicyDroppedTotal),
			Degraded:             auditStats.Degraded,
			DegradedAppendedTotal: int64(auditStats.DegradedAppendedTotal),
		},
		Telemetry: gen.ObservabilityServiceStats{
			QueueDepth:         telemetryStats.QueueDepth,
			QueueCap:           telemetryStats.QueueCap,
			AcceptedTotal:      int64(telemetryStats.AcceptedTotal),
			DroppedTotal:       int64(telemetryStats.DroppedTotal),
			PolicyDroppedTotal: int64(telemetryStats.PolicyDroppedTotal),
		},
		CollectedAt: time.Now().UTC(),
	}, nil
}

// GetObservabilityMetricsPrometheus 以 openmetrics-text v1.0.0 格式导出
// audit.Recorder + telemetry.Ingester 的核心指标，供 Prometheus / Alertmanager
// 等通用监控系统直接 scrape。
//
// 设计说明：
//  1. 数据内容与 GetObservabilityMetrics 一致，差异仅在呈现格式——Prometheus 系列
//     工具链读 text，前端 dashboard 读 JSON，各取所需；
//  2. 导出 audit + telemetry 两组指标：便于外部监控统一观测后端审计链路与前端日志
//     摄取链路；
//  3. 所有样本用单一硬编码 label（job="gge-backend"）保持 cardinality=1，符合
//     openmetrics 「metric name + labels 唯一确定时间序列」约束；
//  4. Noop 模式下 Stats() 返回全零，抓取仍然是 200 + 完整样本（不是空响应），
//     便于告警区分「暂未启用」与「启用后异常」。
//
// Content-Type 受 ogen 生成码约束为 text/plain; charset=utf-8——Prometheus 抓取
// 端默认兼容 text/plain 0.0.4 / openmetrics 1.0.0，内容格式本身符合 openmetrics
// 即可，不依赖 Content-Type 的 version 参数。
func (h *observabilityAPIHandler) GetObservabilityMetricsPrometheus(ctx context.Context) (gen.GetObservabilityMetricsPrometheusRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetObservabilityMetricsPrometheusUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	auditStats := h.audit.Stats()
	telemetryStats := h.telemetry.Stats()

	var b strings.Builder
	// gauge: 队列瞬时深度
	fmt.Fprintln(&b, "# HELP audit_queue_depth audit recorder queue length (len(chan))")
	fmt.Fprintln(&b, "# TYPE audit_queue_depth gauge")
	fmt.Fprintf(&b, "audit_queue_depth %d\n", auditStats.QueueDepth)
	// gauge: 队列容量
	fmt.Fprintln(&b, "# HELP audit_queue_capacity audit recorder queue capacity (cap(chan))")
	fmt.Fprintln(&b, "# TYPE audit_queue_capacity gauge")
	fmt.Fprintf(&b, "audit_queue_capacity %d\n", auditStats.QueueCap)
	// counter: 接收累计
	fmt.Fprintln(&b, "# HELP audit_events_accepted_total cumulative audit events persisted since process start")
	fmt.Fprintln(&b, "# TYPE audit_events_accepted_total counter")
	fmt.Fprintf(&b, "audit_events_accepted_total %d\n", auditStats.AcceptedTotal)
	// counter: 丢弃累计
	fmt.Fprintln(&b, "# HELP audit_events_dropped_total cumulative audit events dropped (queue full, drop-newest)")
	fmt.Fprintln(&b, "# TYPE audit_events_dropped_total counter")
	fmt.Fprintf(&b, "audit_events_dropped_total %d\n", auditStats.DroppedTotal)
	// counter: 策略丢弃累计
	fmt.Fprintln(&b, "# HELP audit_policy_dropped_total cumulative audit events dropped by log policy")
	fmt.Fprintln(&b, "# TYPE audit_policy_dropped_total counter")
	fmt.Fprintf(&b, "audit_policy_dropped_total %d\n", auditStats.PolicyDroppedTotal)
	// gauge: 是否降级模式（断路器非 closed）
	fmt.Fprintln(&b, "# HELP audit_degraded_mode audit recorder circuit breaker degraded state (1=open/half_open, 0=closed)")
	fmt.Fprintln(&b, "# TYPE audit_degraded_mode gauge")
	if auditStats.Degraded {
		fmt.Fprintln(&b, "audit_degraded_mode 1")
	} else {
		fmt.Fprintln(&b, "audit_degraded_mode 0")
	}
	// counter: 降级文件累计追加条数
	fmt.Fprintln(&b, "# HELP audit_degraded_appended_total cumulative audit events appended to degraded sink")
	fmt.Fprintln(&b, "# TYPE audit_degraded_appended_total counter")
	fmt.Fprintf(&b, "audit_degraded_appended_total %d\n", auditStats.DegradedAppendedTotal)

	// gauge: telemetry 队列瞬时深度
	fmt.Fprintln(&b, "# HELP telemetry_queue_depth telemetry ingester queue length (len(chan))")
	fmt.Fprintln(&b, "# TYPE telemetry_queue_depth gauge")
	fmt.Fprintf(&b, "telemetry_queue_depth %d\n", telemetryStats.QueueDepth)
	// gauge: telemetry 队列容量
	fmt.Fprintln(&b, "# HELP telemetry_queue_capacity telemetry ingester queue capacity (cap(chan))")
	fmt.Fprintln(&b, "# TYPE telemetry_queue_capacity gauge")
	fmt.Fprintf(&b, "telemetry_queue_capacity %d\n", telemetryStats.QueueCap)
	// counter: telemetry 接收累计
	fmt.Fprintln(&b, "# HELP telemetry_events_accepted_total cumulative telemetry events accepted since process start")
	fmt.Fprintln(&b, "# TYPE telemetry_events_accepted_total counter")
	fmt.Fprintf(&b, "telemetry_events_accepted_total %d\n", telemetryStats.AcceptedTotal)
	// counter: telemetry 丢弃累计
	fmt.Fprintln(&b, "# HELP telemetry_events_dropped_total cumulative telemetry events dropped (rate limit or queue full)")
	fmt.Fprintln(&b, "# TYPE telemetry_events_dropped_total counter")
	fmt.Fprintf(&b, "telemetry_events_dropped_total %d\n", telemetryStats.DroppedTotal)
	// counter: telemetry 策略丢弃累计
	fmt.Fprintln(&b, "# HELP telemetry_policy_dropped_total cumulative telemetry events dropped by log policy")
	fmt.Fprintln(&b, "# TYPE telemetry_policy_dropped_total counter")
	fmt.Fprintf(&b, "telemetry_policy_dropped_total %d\n", telemetryStats.PolicyDroppedTotal)

	return &gen.GetObservabilityMetricsPrometheusOK{Data: strings.NewReader(b.String())}, nil
}

func (h *observabilityAPIHandler) GetTelemetryLog(ctx context.Context, params gen.GetTelemetryLogParams) (gen.GetTelemetryLogRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetTelemetryLogUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	var row telemetry.TelemetryLog
	if err := h.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", params.ID, logger.TenantFromContext(ctx)).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.GetTelemetryLogNotFound{Code: 404, Message: "遥测日志不存在"}, nil
		}
		h.logger.Error("get telemetry log failed", zap.Error(err))
		return nil, err
	}
	detail := telemetryLogDetailFromModel(&row)
	return &detail, nil
}

// ─── helpers ────────────────────────────────────────────────────────────────

func observabilityPagination(current, size gen.OptInt) (int, int) {
	c := optInt(current, observabilityListDefaultCurrent)
	if c < 1 {
		c = observabilityListDefaultCurrent
	}
	s := optInt(size, observabilityListDefaultSize)
	if s <= 0 {
		s = observabilityListDefaultSize
	}
	if s > observabilityListMaxSize {
		s = observabilityListMaxSize
	}
	return c, s
}

func auditLogItemFromModel(row *audit.AuditLog) gen.AuditLogItem {
	out := gen.AuditLogItem{
		ID:        int64(row.ID),
		Ts:        row.Ts,
		TenantID:  row.TenantID,
		ActorID:   row.ActorID,
		ActorType: row.ActorType,
		Action:    row.Action,
		Outcome:   row.Outcome,
	}
	setOptStringIf(&out.RequestID, row.RequestID)
	setOptStringIf(&out.AppKey, row.AppKey)
	setOptStringIf(&out.WorkspaceID, row.WorkspaceID)
	setOptStringIf(&out.ResourceType, row.ResourceType)
	setOptStringIf(&out.ResourceID, row.ResourceID)
	setOptStringIf(&out.ErrorCode, row.ErrorCode)
	setOptStringIf(&out.IP, row.IP)
	setOptStringIf(&out.UserAgent, row.UserAgent)
	if row.HTTPStatus != 0 {
		out.HTTPStatus = gen.NewOptInt(row.HTTPStatus)
	}
	return out
}

func auditLogDetailFromModel(row *audit.AuditLog) gen.AuditLogDetail {
	out := gen.AuditLogDetail{
		ID:        int64(row.ID),
		Ts:        row.Ts,
		TenantID:  row.TenantID,
		ActorID:   row.ActorID,
		ActorType: row.ActorType,
		Action:    row.Action,
		Outcome:   row.Outcome,
	}
	setOptStringIf(&out.RequestID, row.RequestID)
	setOptStringIf(&out.AppKey, row.AppKey)
	setOptStringIf(&out.WorkspaceID, row.WorkspaceID)
	setOptStringIf(&out.ResourceType, row.ResourceType)
	setOptStringIf(&out.ResourceID, row.ResourceID)
	setOptStringIf(&out.ErrorCode, row.ErrorCode)
	setOptStringIf(&out.IP, row.IP)
	setOptStringIf(&out.UserAgent, row.UserAgent)
	if row.HTTPStatus != 0 {
		out.HTTPStatus = gen.NewOptInt(row.HTTPStatus)
	}
	if m, ok := jsonBytesToRawMap(row.BeforeJSON); ok {
		out.Before = gen.OptNilAuditLogDetailBefore{Value: gen.AuditLogDetailBefore(m), Set: true}
	} else if row.BeforeJSON != nil {
		out.Before = gen.OptNilAuditLogDetailBefore{Set: true, Null: true}
	}
	if m, ok := jsonBytesToRawMap(row.AfterJSON); ok {
		out.After = gen.OptNilAuditLogDetailAfter{Value: gen.AuditLogDetailAfter(m), Set: true}
	} else if row.AfterJSON != nil {
		out.After = gen.OptNilAuditLogDetailAfter{Set: true, Null: true}
	}
	if m, ok := jsonBytesToRawMap(row.Metadata); ok {
		out.Metadata = gen.OptNilAuditLogDetailMetadata{Value: gen.AuditLogDetailMetadata(m), Set: true}
	} else if row.Metadata != nil {
		out.Metadata = gen.OptNilAuditLogDetailMetadata{Set: true, Null: true}
	}
	return out
}

func telemetryLogRecordFromModel(row *telemetry.TelemetryLog) gen.TelemetryLogRecord {
	out := gen.TelemetryLogRecord{
		ID:       int64(row.ID),
		Ts:       row.Ts,
		TenantID: row.TenantID,
		Level:    row.Level,
		Event:    row.Event,
	}
	setOptStringIf(&out.RequestID, row.RequestID)
	setOptStringIf(&out.SessionID, row.SessionID)
	setOptStringIf(&out.ActorID, row.ActorID)
	setOptStringIf(&out.AppKey, row.AppKey)
	setOptStringIf(&out.Message, row.Message)
	setOptStringIf(&out.URL, row.URL)
	setOptStringIf(&out.UserAgent, row.UserAgent)
	setOptStringIf(&out.IP, row.IP)
	setOptStringIf(&out.Release, row.Release)
	return out
}

func telemetryLogDetailFromModel(row *telemetry.TelemetryLog) gen.TelemetryLogDetail {
	out := gen.TelemetryLogDetail{
		ID:       int64(row.ID),
		Ts:       row.Ts,
		TenantID: row.TenantID,
		Level:    row.Level,
		Event:    row.Event,
	}
	setOptStringIf(&out.RequestID, row.RequestID)
	setOptStringIf(&out.SessionID, row.SessionID)
	setOptStringIf(&out.ActorID, row.ActorID)
	setOptStringIf(&out.AppKey, row.AppKey)
	setOptStringIf(&out.Message, row.Message)
	setOptStringIf(&out.URL, row.URL)
	setOptStringIf(&out.UserAgent, row.UserAgent)
	setOptStringIf(&out.IP, row.IP)
	setOptStringIf(&out.Release, row.Release)
	if m, ok := jsonBytesToRawMap(row.Payload); ok {
		out.Payload = gen.OptNilTelemetryLogDetailPayload{Value: gen.TelemetryLogDetailPayload(m), Set: true}
	} else if row.Payload != nil {
		out.Payload = gen.OptNilTelemetryLogDetailPayload{Set: true, Null: true}
	}
	return out
}

func setOptStringIf(target *gen.OptString, value string) {
	if value == "" {
		return
	}
	*target = gen.NewOptString(value)
}

// jsonBytesToRawMap 将 jsonb 列的 []byte 反序列化成 gen 侧要求的
// map[string]jx.Raw；非对象 / 空 / 损坏的值统一返回 false 让调用方决定
// 是写 null 还是保持 unset。
func jsonBytesToRawMap(buf []byte) (map[string]jx.Raw, bool) {
	if len(buf) == 0 {
		return nil, false
	}
	var tmp map[string]json.RawMessage
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return nil, false
	}
	if len(tmp) == 0 {
		return nil, false
	}
	out := make(map[string]jx.Raw, len(tmp))
	for k, v := range tmp {
		out[k] = jx.Raw(v)
	}
	return out, true
}

// ensure time package is kept on the import even when ogen renames internal fields.
var _ = time.Time{}

