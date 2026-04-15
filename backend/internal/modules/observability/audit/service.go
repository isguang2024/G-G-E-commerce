// Package audit 提供业务操作审计日志的写入服务。
//
// 设计要点：
//  1. 对外只暴露一个接口 Recorder + 一个事件结构体 Event，所有 handler / service
//     通过 Recorder.Record(ctx, Event{...}) 写入。ctx 里自动携带 request_id / actor /
//     tenant / app / workspace，无需在 Event 里重复填。
//  2. 持久化走异步 channel + worker 池，不阻塞请求；channel 满时降级为
//     "drop-newest + Warn 日志"，保证 DB 抖动不会撑爆内存。
//  3. 脱敏在入队前同步完成，避免敏感字段以明文进入内存 channel。
//  4. 对单元测试 / CI 友好：当 cfg.Enabled=false 或 DB 为 nil 时，退化为 Noop，
//     业务代码不需要用 if recorder != nil 判空。
package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/observability/logpolicy"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

// Event 是一次待写入的审计事件。调用方只需要填业务关心的字段：
// Action（必填）、ResourceType、ResourceID、Outcome、ErrorCode、Before/After。
// request_id / actor / tenant / app / workspace / IP / UA 自动从 ctx 读取。
type Event struct {
	Action       string // 必填：领域.实体.动作，例如 "system.app.create"
	ResourceType string // 资源类型，例如 "app"、"role"
	ResourceID   string // 资源主键，字符串化
	Outcome      string // success | denied | error；为空时按 ErrorCode 自动判定
	ErrorCode    string // 对齐 apperr code；仅 Outcome != success 时有意义
	HTTPStatus   int    // 对应 HTTP 响应码，方便按状态聚合

	// Before / After 是操作前/后的结构化快照。会被 JSON 序列化并脱敏写入。
	// 建议只保留业务字段，避免把整条 GORM model 灌进去（含太多噪音列）。
	Before any
	After  any

	// Metadata 用于承载 Action 特有的扩展字段，例如 dispatch_id / send_count。
	Metadata map[string]any
}

// Recorder 是审计服务对外接口。
//
// Record 约定：永远不返回业务级错误（不阻塞 handler 主流程）；
// 内部失败（JSON 序列化、channel 满、DB 异常）只记 Warn/Error 日志。
// 调用方可以安全地忽略返回值。
//
// Stats 返回当前运行时观测指标（队列深度、累计 accepted / dropped 计数），
// 供 /metrics、健康检查或运维仪表盘读取。Noop 实现返回零值结构。
type Recorder interface {
	Record(ctx context.Context, e Event)
	Stats() Stats
	Shutdown(ctx context.Context) error
}

// Stats 是 Recorder 在当前进程内累计的运行时指标。
//
// 语义约定：
//   - QueueDepth: 当前 buffered channel 里待消费的事件数（同步模式恒为 0）；
//   - QueueCap:   channel 容量（同步模式恒为 0），0 可用于判断「是否启用了异步缓冲」；
//   - AcceptedTotal: 进程启动以来成功写入 DB 的事件累计；
//   - DroppedTotal:  进程启动以来因 channel 满被丢弃的事件累计（其他失败路径归类到日志而非 dropped）。
//
// 这些字段单调递增，不会因 Shutdown 或重启自动清零；进程重启后计数归 0。
type Stats struct {
	QueueDepth            int    `json:"queue_depth"`
	QueueCap              int    `json:"queue_cap"`
	AcceptedTotal         uint64 `json:"accepted_total"`
	DroppedTotal          uint64 `json:"dropped_total"`
	PolicyDroppedTotal    uint64 `json:"policy_dropped_total"`
	Degraded              bool   `json:"degraded"`
	DegradedAppendedTotal uint64 `json:"degraded_appended_total"`
}

// Config 控制 audit service 的运行参数。通过 config.AuditConfig 加载。
type Config struct {
	Enabled       bool
	RedactFields  []string
	QueueSize     int           // 异步 channel 缓冲，默认 1024
	Workers       int           // 消费 goroutine 数量，默认 2
	BatchSize     int           // 批量落库阈值，默认 100
	FlushInterval time.Duration // 批量最大等待时间，默认 1s
	AsyncMode     bool          // false = 同步写入（测试友好）；true = channel+worker
	DegradedFile  string        // 断路器打开时降级写入的 JSONL 文件路径
	PolicyEngine  logpolicy.Engine
}

// DefaultConfig 提供生产默认值。对应 config.example.yaml 里的 audit 默认。
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		RedactFields:  DefaultRedactFields,
		QueueSize:     1024,
		Workers:       2,
		BatchSize:     100,
		FlushInterval: time.Second,
		AsyncMode:     true,
		DegradedFile:  "./data/audit_degraded.jsonl",
	}
}

type circuitState string

const (
	circuitClosed   circuitState = "closed"
	circuitOpen     circuitState = "open"
	circuitHalfOpen circuitState = "half_open"
)

const (
	circuitFailureThreshold = 5
	circuitFailureWindow    = 10 * time.Second
	circuitHalfOpenAfter    = 30 * time.Second
)

// service 是 Recorder 的默认实现。
type service struct {
	db       *gorm.DB
	log      *zap.Logger
	cfg      Config
	redactor *redactor

	mu       sync.Mutex
	queue    chan *AuditLog
	wg       sync.WaitGroup
	stopCh   chan struct{}
	shutdown bool
	dropped  uint64 // 累计丢弃数量（channel 满时）
	accepted uint64 // 累计 accepted 数量（成功写 DB）
	policyDropped uint64
	degradedTotal uint64

	// 测试替身：为空时走真实 DB，非空时由测试接管写入行为。
	writeBatchFn func(context.Context, []*AuditLog) error
	writeRowFn   func(context.Context, *AuditLog) error

	cbMu            sync.Mutex
	cbState         circuitState
	cbOpenedAt      time.Time
	cbProbeInFlight bool
	cbFailures      []time.Time
	sink            degradedSink
	policyEngine    logpolicy.Engine
}

// New 构建一个 Recorder。
//   - db 为 nil 或 cfg.Enabled=false 时，返回 Noop，业务侧继续调 Record 不报错；
//   - AsyncMode=true 时启动 workers 个后台消费 goroutine；
//   - 调用方 main 应在 shutdown 阶段 defer recorder.Shutdown(ctx)，保证 channel 里的
//     事件能被 drain 干净再退出。
func New(db *gorm.DB, log *zap.Logger, cfg Config) Recorder {
	if !cfg.Enabled || db == nil {
		return Noop{}
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = time.Second
	}
	if cfg.DegradedFile == "" {
		cfg.DegradedFile = "./data/audit_degraded.jsonl"
	}
	if log == nil {
		log = zap.NewNop()
	}

	fields := cfg.RedactFields
	if len(fields) == 0 {
		fields = DefaultRedactFields
	}

	s := &service{
		db:       db,
		log:      log.Named("audit"),
		cfg:      cfg,
		redactor: newRedactor(fields),
		stopCh:   make(chan struct{}),
		cbState:  circuitClosed,
		policyEngine: cfg.PolicyEngine,
	}
	sink, err := newJSONLDegradedSink(cfg.DegradedFile)
	if err != nil {
		s.log.Warn("audit.degraded_sink_disabled", zap.Error(err))
	} else {
		s.sink = sink
	}
	if cfg.AsyncMode {
		s.queue = make(chan *AuditLog, cfg.QueueSize)
		for i := 0; i < cfg.Workers; i++ {
			s.wg.Add(1)
			go s.runWorker()
		}
	}
	return s
}

// Record 实现 Recorder 接口。ctx / event 组合成 *AuditLog 并入队 / 同步写入。
func (s *service) Record(ctx context.Context, e Event) {
	if !s.shouldPersistByPolicy(ctx, e) {
		s.addPolicyDropped(1)
		return
	}
	row, err := s.build(ctx, e)
	if err != nil {
		s.log.Warn("audit.build_failed",
			zap.Error(err),
			zap.String("action", e.Action),
			zap.String("request_id", logger.RequestIDFromContext(ctx)),
		)
		return
	}
	if !s.cfg.AsyncMode {
		if s.persistRow(ctx, row) {
			s.addAccepted(1)
		}
		return
	}
	select {
	case s.queue <- row:
	default:
		dropped := s.addDropped(1)
		s.log.Warn("audit.queue_full_drop",
			zap.String("action", row.Action),
			zap.Uint64("dropped_total", dropped),
		)
	}
}

// Stats 返回当前进程累计的审计事件观测指标。调用方需要周期读取并暴露到 /metrics。
// 读取仅拿一次锁+两次 chan 内建查询，代价恒定；无需担心高并发 Record 时被拖慢。
func (s *service) Stats() Stats {
	s.mu.Lock()
	accepted := s.accepted
	dropped := s.dropped
	policyDropped := s.policyDropped
	degradedTotal := s.degradedTotal
	s.mu.Unlock()
	degraded := s.isDegradedMode()
	depth, queueCap := 0, 0
	if s.queue != nil {
		depth = len(s.queue)
		queueCap = cap(s.queue)
	}
	return Stats{
		QueueDepth:            depth,
		QueueCap:              queueCap,
		AcceptedTotal:         accepted,
		DroppedTotal:          dropped,
		PolicyDroppedTotal:    policyDropped,
		Degraded:              degraded,
		DegradedAppendedTotal: degradedTotal,
	}
}

// Shutdown 在服务退出时调用，关闭 channel 并等待 workers 把剩余事件写完。
// 带超时 ctx，防止 DB 卡住导致进程不退出。
func (s *service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return nil
	}
	s.shutdown = true
	s.mu.Unlock()

	close(s.stopCh)
	if s.queue != nil {
		close(s.queue)
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return s.closeSink()
	case <-ctx.Done():
		_ = s.closeSink()
		return errors.New("audit: shutdown timeout; some events may be lost")
	}
}

func (s *service) runWorker() {
	defer s.wg.Done()
	batchSize := s.cfg.BatchSize
	flushInterval := s.cfg.FlushInterval
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	batch := make([]*AuditLog, 0, batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		flushed := s.flushBatch(batch)
		s.addAccepted(uint64(flushed))
		batch = batch[:0]
	}

	for {
		select {
		case row, ok := <-s.queue:
			if !ok {
				flush()
				return
			}
			if row == nil {
				continue
			}
			batch = append(batch, row)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (s *service) flushBatch(rows []*AuditLog) int {
	if len(rows) == 0 {
		return 0
	}
	now := time.Now().UTC()
	writeToDegraded, halfOpenProbe := s.beforeBatchWrite(now)
	if writeToDegraded {
		return s.appendToDegraded(rows, "circuit_open")
	}

	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := s.writeBatch(writeCtx, rows)
	cancel()
	if err == nil {
		s.onBatchWriteSuccess(now, halfOpenProbe)
		if halfOpenProbe {
			if replayErr := s.replayDegraded(); replayErr != nil {
				s.log.Warn("audit.degraded_replay_failed_reopen", zap.Error(replayErr))
				return len(rows)
			}
		}
		return len(rows)
	}

	s.log.Warn("audit.batch_write_failed_fallback_single",
		zap.Error(err),
		zap.Int("batch_size", len(rows)),
	)
	if halfOpenProbe {
		s.onHalfOpenFailure(now, err)
		return s.appendToDegraded(rows, "half_open_failed")
	}
	if s.onClosedBatchFailure(now) {
		return s.appendToDegraded(rows, "circuit_opened")
	}

	return s.persistRowsOneByOne(rows)
}

func (s *service) persistRow(ctx context.Context, row *AuditLog) bool {
	if err := s.writeRow(ctx, row); err != nil {
		s.log.Error("audit.write_failed",
			zap.Error(err),
			zap.String("action", row.Action),
			zap.String("request_id", row.RequestID),
		)
		return false
	}
	return true
}

func (s *service) writeBatch(ctx context.Context, rows []*AuditLog) error {
	if s.writeBatchFn != nil {
		return s.writeBatchFn(ctx, rows)
	}
	return s.db.WithContext(ctx).CreateInBatches(rows, s.cfg.BatchSize).Error
}

func (s *service) writeRow(ctx context.Context, row *AuditLog) error {
	if s.writeRowFn != nil {
		return s.writeRowFn(ctx, row)
	}
	return s.db.WithContext(ctx).Create(row).Error
}

func (s *service) persistRowsOneByOne(rows []*AuditLog) int {
	okCount := 0
	for _, row := range rows {
		if row == nil {
			continue
		}
		writeRowCtx, rowCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if s.persistRow(writeRowCtx, row) {
			okCount++
		}
		rowCancel()
	}
	return okCount
}

func (s *service) addDropped(n uint64) uint64 {
	s.mu.Lock()
	s.dropped += n
	total := s.dropped
	s.mu.Unlock()
	return total
}

func (s *service) addAccepted(n uint64) {
	if n == 0 {
		return
	}
	s.mu.Lock()
	s.accepted += n
	s.mu.Unlock()
}

func (s *service) addPolicyDropped(n uint64) {
	if n == 0 {
		return
	}
	s.mu.Lock()
	s.policyDropped += n
	s.mu.Unlock()
}

func (s *service) addDegradedAppended(n uint64) uint64 {
	if n == 0 {
		return 0
	}
	s.mu.Lock()
	s.degradedTotal += n
	total := s.degradedTotal
	s.mu.Unlock()
	return total
}

func (s *service) appendToDegraded(rows []*AuditLog, reason string) int {
	if len(rows) == 0 {
		return 0
	}
	if s.sink == nil {
		s.log.Error("audit.degraded_sink_unavailable",
			zap.String("reason", reason),
			zap.Int("batch_size", len(rows)),
		)
		return s.persistRowsOneByOne(rows)
	}

	appended := 0
	for idx, row := range rows {
		if row == nil {
			continue
		}
		if err := s.sink.Append(row); err != nil {
			s.log.Error("audit.degraded_append_failed",
				zap.Error(err),
				zap.String("reason", reason),
				zap.String("action", row.Action),
				zap.String("request_id", row.RequestID),
			)
			return s.persistRowsOneByOne(rows[idx:])
		}
		appended++
	}
	total := s.addDegradedAppended(uint64(appended))
	s.log.Warn("audit.degraded_appended",
		zap.String("reason", reason),
		zap.Int("batch_size", len(rows)),
		zap.Int("appended", appended),
		zap.Uint64("degraded_appended_total", total),
	)
	return 0
}

func (s *service) replayDegraded() error {
	if s.sink == nil {
		return nil
	}
	replayed := 0
	err := s.sink.Replay(func(row *AuditLog) error {
		if row == nil {
			return nil
		}
		writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.writeRow(writeCtx, row); err != nil {
			return fmt.Errorf("replay write row failed: %w", err)
		}
		replayed++
		return nil
	})
	if err != nil {
		s.forceOpen(time.Now().UTC(), err)
		return err
	}
	if replayed > 0 {
		s.addAccepted(uint64(replayed))
		s.log.Info("audit.degraded_replayed", zap.Int("rows", replayed))
	}
	return nil
}

func (s *service) beforeBatchWrite(now time.Time) (writeToDegraded bool, halfOpenProbe bool) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()

	switch s.cbState {
	case circuitOpen:
		if now.Sub(s.cbOpenedAt) >= circuitHalfOpenAfter && !s.cbProbeInFlight {
			s.cbState = circuitHalfOpen
			s.cbProbeInFlight = true
			s.log.Info("audit.circuit_half_open_probe")
			return false, true
		}
		return true, false
	case circuitHalfOpen:
		if s.cbProbeInFlight {
			return true, false
		}
		s.cbProbeInFlight = true
		return false, true
	default:
		return false, false
	}
}

func (s *service) onBatchWriteSuccess(now time.Time, halfOpenProbe bool) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	if halfOpenProbe {
		s.cbState = circuitClosed
		s.cbOpenedAt = time.Time{}
		s.cbProbeInFlight = false
		s.cbFailures = s.cbFailures[:0]
		s.log.Info("audit.circuit_closed_after_half_open")
		return
	}
	if s.cbState == circuitClosed && len(s.cbFailures) > 0 {
		s.cbFailures = s.cbFailures[:0]
	}
}

func (s *service) onHalfOpenFailure(now time.Time, err error) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	s.cbState = circuitOpen
	s.cbOpenedAt = now
	s.cbProbeInFlight = false
	s.cbFailures = s.cbFailures[:0]
	s.log.Warn("audit.circuit_reopen_after_half_open_failure", zap.Error(err))
}

func (s *service) onClosedBatchFailure(now time.Time) (opened bool) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	if s.cbState != circuitClosed {
		return false
	}
	pruned := s.cbFailures[:0]
	for _, ts := range s.cbFailures {
		if now.Sub(ts) <= circuitFailureWindow {
			pruned = append(pruned, ts)
		}
	}
	s.cbFailures = append(pruned, now)
	if len(s.cbFailures) >= circuitFailureThreshold {
		s.cbState = circuitOpen
		s.cbOpenedAt = now
		s.cbProbeInFlight = false
		s.cbFailures = s.cbFailures[:0]
		s.log.Warn("audit.circuit_opened",
			zap.Int("failure_threshold", circuitFailureThreshold),
			zap.Duration("failure_window", circuitFailureWindow),
		)
		return true
	}
	return false
}

func (s *service) forceOpen(now time.Time, cause error) {
	s.cbMu.Lock()
	s.cbState = circuitOpen
	s.cbOpenedAt = now
	s.cbProbeInFlight = false
	s.cbFailures = s.cbFailures[:0]
	s.cbMu.Unlock()
	if cause != nil {
		s.log.Warn("audit.circuit_forced_open", zap.Error(cause))
	}
}

func (s *service) isDegradedMode() bool {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	return s.cbState != circuitClosed
}

func (s *service) shouldPersistByPolicy(ctx context.Context, e Event) bool {
	if s.policyEngine == nil {
		return true
	}
	decision := s.policyEngine.Decide(logpolicy.PipelineAudit, map[string]string{
		logpolicy.MatchFieldAction:       strings.TrimSpace(e.Action),
		logpolicy.MatchFieldOutcome:      strings.TrimSpace(e.Outcome),
		logpolicy.MatchFieldResourceType: strings.TrimSpace(e.ResourceType),
	})
	switch decision.Decision {
	case logpolicy.DecisionDeny:
		return false
	case logpolicy.DecisionSample:
		key := strings.Join([]string{
			logger.RequestIDFromContext(ctx),
			e.Action,
			e.ResourceType,
			e.ResourceID,
		}, "|")
		return logpolicy.ShouldKeepBySample(decision.SampleRate, key)
	default:
		return true
	}
}

func (s *service) closeSink() error {
	if s.sink == nil {
		return nil
	}
	if err := s.sink.Close(); err != nil {
		return fmt.Errorf("audit degraded sink close failed: %w", err)
	}
	return nil
}

// build 将 Event + ctx 合成一条 AuditLog，完成：
//  1. 取 ctx 中的 request_id / actor / tenant / app / workspace / client info；
//  2. 对 Before/After/Metadata 做 JSON 序列化 + 敏感字段脱敏；
//  3. Outcome 缺省按 ErrorCode 推断。
func (s *service) build(ctx context.Context, e Event) (*AuditLog, error) {
	if e.Action == "" {
		return nil, errors.New("audit: action is required")
	}
	outcome := e.Outcome
	if outcome == "" {
		if e.ErrorCode != "" {
			outcome = OutcomeError
		} else {
			outcome = OutcomeSuccess
		}
	}

	before, err := s.marshalRedacted(e.Before)
	if err != nil {
		return nil, err
	}
	after, err := s.marshalRedacted(e.After)
	if err != nil {
		return nil, err
	}
	var metadata []byte
	if len(e.Metadata) > 0 {
		metadata, err = s.marshalRedactedMap(e.Metadata)
		if err != nil {
			return nil, err
		}
	} else {
		metadata = []byte(`{}`)
	}

	now := time.Now().UTC()
	row := &AuditLog{
		Ts:           now,
		RequestID:    logger.RequestIDFromContext(ctx),
		TenantID:     logger.TenantFromContext(ctx),
		ActorID:      logger.ActorIDFromContext(ctx),
		ActorType:    logger.ActorTypeFromContext(ctx),
		AppKey:       logger.AppFromContext(ctx),
		WorkspaceID:  logger.WorkspaceFromContext(ctx),
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		Outcome:      outcome,
		ErrorCode:    e.ErrorCode,
		HTTPStatus:   e.HTTPStatus,
		IP:           logger.ClientIPFromContext(ctx),
		UserAgent:    logger.UserAgentFromContext(ctx),
		BeforeJSON:   before,
		AfterJSON:    after,
		Metadata:     metadata,
		CreatedAt:    now,
	}
	return row, nil
}

// marshalRedacted 把任意结构 -> JSON bytes，然后读回 map 做脱敏，再序列化回去。
// nil 时写入字面量 null，DB 列默认 'null'::jsonb 对齐。
func (s *service) marshalRedacted(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	decoded = s.redactor.redact(decoded)
	return json.Marshal(decoded)
}

// marshalRedactedMap 是 marshalRedacted 针对顶层 map 的快捷版本。
func (s *service) marshalRedactedMap(m map[string]any) ([]byte, error) {
	copyMap := make(map[string]any, len(m))
	for k, v := range m {
		copyMap[k] = v
	}
	redacted := s.redactor.redact(copyMap)
	return json.Marshal(redacted)
}

// Noop 是一个显式的"什么都不做"的 Recorder，用于关闭审计 / 无 DB 环境。
// 特意放到包级可导出，方便测试 stub：audit.Noop{}。
type Noop struct{}

// Record ignores everything.
func (Noop) Record(_ context.Context, _ Event) {}

// Stats 返回零值 Stats，表示没有运行时数据。
func (Noop) Stats() Stats { return Stats{} }

// Shutdown 总是立即返回 nil。
func (Noop) Shutdown(_ context.Context) error { return nil }
