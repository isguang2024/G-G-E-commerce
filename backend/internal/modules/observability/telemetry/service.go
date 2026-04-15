// Package telemetry 负责前端批量上报的日志摄取。
//
// 设计动机：
//  1. /telemetry/logs 的流量完全由前端决定，恶意刷请求或 SPA bug 都可能导致
//     流量放大。服务端需要一个兜底的限流 + 安全截断层，保证写 DB 不会被打爆。
//  2. 服务端也要承担二次脱敏：虽然前端 logger 已经过了一遍 REDACT_KEYS，
//     但上游代码可能被绕过，所以入库前我们再用 audit.redactor 兜一遍。
//  3. 对外暴露的接口是 Ingester，handler 不直接操作 DB；方便测试注入 Noop。
package telemetry

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

	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
	"github.com/gg-ecommerce/backend/internal/modules/observability/logpolicy"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

// Entry 是一条待摄取的前端日志。字段对齐 openapi TelemetryLogEntry。
// 调用方（handler）负责把 ogen 生成的结构体 decode 成本结构体，隔离 gen 包。
type Entry struct {
	Level     string         // debug|info|warn|error；越界归并成 info
	Event     string         // 稳定事件名
	Timestamp time.Time      // 前端发生时间
	RequestID string         // 前端从响应头 X-Request-Id 取回的值
	Context   map[string]any // 结构化字段
	Error     *ErrorSnapshot // Error 快照
	Route     string         // 前端路由
	UserID    string         // 前端用户身份，空表示匿名
	SessionID string         // 前端 tab 级别 session 标识
	UserAgent string         // 前端 UA
	Viewport  Viewport       // 视口
}

// ErrorSnapshot 对应 TelemetryErrorSnapshot。
type ErrorSnapshot struct {
	Name    string
	Message string
	Stack   string
}

// Viewport 对应 TelemetryViewport。
type Viewport struct {
	W int
	H int
}

// Ingester 对外暴露的摄取接口。
//
// Ingest 约定：
//   - 不返回业务错误（批量上报绝对不能让前端看到 5xx / 4xx）；
//   - 返回 accepted / dropped 数量，调用方回传给前端做心理预期对齐；
//   - 被限流时把整个 batch 的 dropped 全部扣住，accepted=0，不再调用 DB。
//
// Stats 返回当前进程累计的观测指标（queue depth/cap、accepted/dropped 累计值），
// 供 /metrics 端点或运维仪表盘读取。Noop 返零值。
type Ingester interface {
	Ingest(ctx context.Context, entries []Entry, sessionKey, ipKey string) Result
	Stats() Stats
	Shutdown(ctx context.Context) error
}

// Result 描述一次 ingest 的结果。
type Result struct {
	Accepted int32
	Dropped  int32
}

// Stats 是 Ingester 在当前进程内累计的运行时指标。
//
// 语义约定：
//   - QueueDepth:    当前 buffered channel 里待消费的 entry 数（同步模式恒为 0）；
//   - QueueCap:      channel 容量（同步模式恒为 0）；
//   - AcceptedTotal: 进程启动以来成功入队 / 同步写 DB 的 entry 累计；
//   - DroppedTotal:  进程启动以来因限流（session / IP token bucket）或队列满被丢弃的 entry 累计。
//
// 单调递增，进程重启后归 0。
type Stats struct {
	QueueDepth         int    `json:"queue_depth"`
	QueueCap           int    `json:"queue_cap"`
	AcceptedTotal      uint64 `json:"accepted_total"`
	DroppedTotal       uint64 `json:"dropped_total"`
	PolicyDroppedTotal uint64 `json:"policy_dropped_total"`
}

// Config 控制 telemetry 服务的运行参数。
type Config struct {
	Enabled         bool
	QueueSize       int // 异步写入 channel 缓冲；<= 0 则同步写
	Workers         int // 异步 worker 数
	RedactFields    []string
	Release         string  // 构建版本/git sha，写入 telemetry_logs.release
	PerSessionRate  float64 // 每秒 ingest 请求数上限（按 session_id）
	PerSessionBurst float64 // session token bucket 容量
	PerIPRate       float64 // 每秒 ingest 请求数上限（按 IP）
	PerIPBurst      float64 // IP token bucket 容量
	BucketIdleTTL   time.Duration
	MaxMessageBytes int // 单条 entry 序列化后最大字节数；超限截断
	PolicyEngine    logpolicy.Engine
}

// DefaultConfig 提供生产默认值。
func DefaultConfig() Config {
	return Config{
		Enabled:         true,
		QueueSize:       2048,
		Workers:         2,
		RedactFields:    audit.DefaultRedactFields,
		PerSessionRate:  10,
		PerSessionBurst: 30,
		PerIPRate:       100,
		PerIPBurst:      300,
		BucketIdleTTL:   5 * time.Minute,
		MaxMessageBytes: 16 * 1024,
	}
}

type service struct {
	db             *gorm.DB
	log            *zap.Logger
	cfg            Config
	redactor       redactorLike
	sessionLimiter *rateLimiter
	ipLimiter      *rateLimiter

	queue    chan *TelemetryLog
	stopCh   chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	shutdown bool
	dropped  uint64
	accepted uint64
	policyDropped uint64
	policyEngine  logpolicy.Engine
}

// redactorLike 复用 audit 的 redactor，避免重复实现。
type redactorLike interface {
	Redact(v any) any
}

type auditRedactorAdapter struct {
	r audit.RedactorPublic
}

func (a auditRedactorAdapter) Redact(v any) any { return a.r.Redact(v) }

// New 构造 Ingester。DB 为 nil 或 Enabled=false 时退化为 Noop，保持调用方无需判空。
func New(db *gorm.DB, log *zap.Logger, cfg Config) Ingester {
	if !cfg.Enabled || db == nil {
		return Noop{}
	}
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.QueueSize < 0 {
		cfg.QueueSize = 0
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if cfg.PerSessionRate <= 0 {
		cfg.PerSessionRate = 10
	}
	if cfg.PerSessionBurst <= 0 {
		cfg.PerSessionBurst = cfg.PerSessionRate * 3
	}
	if cfg.PerIPRate <= 0 {
		cfg.PerIPRate = 100
	}
	if cfg.PerIPBurst <= 0 {
		cfg.PerIPBurst = cfg.PerIPRate * 3
	}
	if cfg.BucketIdleTTL <= 0 {
		cfg.BucketIdleTTL = 5 * time.Minute
	}
	if cfg.MaxMessageBytes <= 0 {
		cfg.MaxMessageBytes = 16 * 1024
	}

	fields := cfg.RedactFields
	if len(fields) == 0 {
		fields = audit.DefaultRedactFields
	}

	s := &service{
		db:             db,
		log:            log.Named("telemetry"),
		cfg:            cfg,
		redactor:       auditRedactorAdapter{r: audit.NewPublicRedactor(fields)},
		sessionLimiter: newRateLimiter(cfg.PerSessionBurst, cfg.PerSessionRate, cfg.BucketIdleTTL),
		ipLimiter:      newRateLimiter(cfg.PerIPBurst, cfg.PerIPRate, cfg.BucketIdleTTL),
		stopCh:         make(chan struct{}),
		policyEngine:   cfg.PolicyEngine,
	}
	if cfg.QueueSize > 0 {
		s.queue = make(chan *TelemetryLog, cfg.QueueSize)
		for i := 0; i < cfg.Workers; i++ {
			s.wg.Add(1)
			go s.runWorker()
		}
	}
	s.wg.Add(1)
	go s.runGC()
	return s
}

func (s *service) Ingest(ctx context.Context, entries []Entry, sessionKey, ipKey string) Result {
	if len(entries) == 0 {
		return Result{}
	}
	// 两层限流：session 维度通常先触发；IP 维度防止同 IP 多 tab 作弊。
	if !s.sessionLimiter.allow(sessionKey) || !s.ipLimiter.allow(ipKey) {
		s.mu.Lock()
		s.dropped += uint64(len(entries))
		dropped := s.dropped
		s.mu.Unlock()
		s.log.Warn("telemetry.rate_limited",
			zap.String("session", sessionKey),
			zap.String("ip", ipKey),
			zap.Int("dropped_batch", len(entries)),
			zap.Uint64("dropped_total", dropped),
		)
		return Result{Accepted: 0, Dropped: int32(len(entries))}
	}

	tenant := logger.TenantFromContext(ctx)
	if tenant == "" {
		tenant = "default"
	}
	appKey := logger.AppFromContext(ctx)
	actorID := logger.ActorIDFromContext(ctx)

	accepted := int32(0)
	dropped := int32(0)
	for i := range entries {
		if !s.shouldPersistByPolicy(&entries[i]) {
			dropped++
			s.mu.Lock()
			s.dropped++
			s.policyDropped++
			s.mu.Unlock()
			continue
		}
		row := s.build(tenant, appKey, actorID, ipKey, &entries[i])
		if row == nil {
			dropped++
			continue
		}
		if s.queue != nil {
			select {
			case s.queue <- row:
				accepted++
				s.mu.Lock()
				s.accepted++
				s.mu.Unlock()
			default:
				dropped++
				s.mu.Lock()
				s.dropped++
				s.mu.Unlock()
			}
		} else {
			// 同步模式（测试 / 小流量场景）。
			s.writeRow(ctx, row)
			accepted++
			s.mu.Lock()
			s.accepted++
			s.mu.Unlock()
		}
	}
	return Result{Accepted: accepted, Dropped: dropped}
}

// Stats 返回当前进程累计的遥测观测指标。供 /metrics 端点或 runtime 运维仪表盘读取。
func (s *service) Stats() Stats {
	s.mu.Lock()
	accepted := s.accepted
	dropped := s.dropped
	policyDropped := s.policyDropped
	s.mu.Unlock()
	depth, queueCap := 0, 0
	if s.queue != nil {
		depth = len(s.queue)
		queueCap = cap(s.queue)
	}
	return Stats{
		QueueDepth:         depth,
		QueueCap:           queueCap,
		AcceptedTotal:      accepted,
		DroppedTotal:       dropped,
		PolicyDroppedTotal: policyDropped,
	}
}

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
		return nil
	case <-ctx.Done():
		return errors.New("telemetry: shutdown timeout; some entries may be lost")
	}
}

// build 把 Entry 规整成 TelemetryLog；越界字段做截断，未知 level 归并 info。
// 返回 nil 表示这条 entry 没有业务价值（例如 event 为空），调用方按 dropped 计数。
func (s *service) build(tenant, appKey, actorID, ip string, e *Entry) *TelemetryLog {
	if e == nil || strings.TrimSpace(e.Event) == "" {
		return nil
	}
	level := strings.ToLower(strings.TrimSpace(e.Level))
	switch level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		// ok
	default:
		level = LevelInfo
	}

	ts := e.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	// 入库前再脱敏一次：前端 redact 是软约束，服务端必须兜底。
	var ctxJSON []byte
	if len(e.Context) > 0 {
		copyMap := make(map[string]any, len(e.Context))
		for k, v := range e.Context {
			copyMap[k] = v
		}
		redacted := s.redactor.Redact(copyMap)
		if raw, err := json.Marshal(redacted); err == nil {
			ctxJSON = truncate(raw, s.cfg.MaxMessageBytes)
		}
	}

	// error 快照直接作为 payload 的一部分写入（便于查询）。
	payload := map[string]any{
		"context":  json.RawMessage(nilBytesToNull(ctxJSON)),
		"route":    e.Route,
		"session":  e.SessionID,
		"viewport": map[string]int{"w": e.Viewport.W, "h": e.Viewport.H},
		"ua":       truncString(e.UserAgent, 512),
	}
	if e.Error != nil {
		payload["error"] = map[string]string{
			"name":    truncString(e.Error.Name, 128),
			"message": truncString(e.Error.Message, 2048),
			"stack":   truncString(e.Error.Stack, 8192),
		}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		payloadBytes = []byte(`{}`)
	}
	payloadBytes = truncate(payloadBytes, s.cfg.MaxMessageBytes)

	message := ""
	if e.Error != nil {
		message = truncString(e.Error.Message, 2048)
	}

	now := time.Now().UTC()
	// Level 列的主 ActorID 优先用前端上报值（未登录匿名会话也能追踪），
	// 后端 ctx 里的 actor_id 作为 fallback 保障登录态一致性。
	effectiveActor := strings.TrimSpace(e.UserID)
	if effectiveActor == "" {
		effectiveActor = actorID
	}

	return &TelemetryLog{
		Ts:        ts.UTC(),
		RequestID: truncString(e.RequestID, 64),
		SessionID: truncString(e.SessionID, 64),
		TenantID:  tenant,
		ActorID:   truncString(effectiveActor, 64),
		AppKey:    appKey,
		Level:     level,
		Event:     truncString(e.Event, 128),
		Message:   message,
		URL:       truncString(e.Route, 512),
		UserAgent: truncString(e.UserAgent, 512),
		IP:        truncString(ip, 64),
		Release:   truncString(s.cfg.Release, 64),
		Payload:   payloadBytes,
		CreatedAt: now,
	}
}

func (s *service) runWorker() {
	defer s.wg.Done()
	for row := range s.queue {
		if row == nil {
			continue
		}
		writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.writeRow(writeCtx, row)
		cancel()
	}
}

func (s *service) writeRow(ctx context.Context, row *TelemetryLog) {
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		s.log.Error("telemetry.write_failed",
			zap.Error(err),
			zap.String("event", row.Event),
			zap.String("request_id", row.RequestID),
		)
	}
}

// runGC 定期清理长期空闲的 bucket，防止 session 漂移导致内存线性增长。
func (s *service) runGC() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.cfg.BucketIdleTTL)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case now := <-ticker.C:
			s.sessionLimiter.gc(now)
			s.ipLimiter.gc(now)
		}
	}
}

// Noop 是关闭 / 无 DB 场景下的 Ingester；Ingest 始终返回 accepted=len,dropped=0。
// 用 "全部 accepted" 而不是 "全部 dropped"，是为了不让测试环境里前端看到诡异的
// 大量 dropped 指标；业务上"没人要"也等价于"已经收下了"。
type Noop struct{}

func (Noop) Ingest(_ context.Context, entries []Entry, _, _ string) Result {
	return Result{Accepted: int32(len(entries)), Dropped: 0}
}

// Stats 返回零值 Stats（表示 Noop 不承载任何运行时指标）。
func (Noop) Stats() Stats                     { return Stats{} }
func (Noop) Shutdown(_ context.Context) error { return nil }

// --- helpers ---

func truncString(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

func truncate(b []byte, n int) []byte {
	if n <= 0 || len(b) <= n {
		return b
	}
	return b[:n]
}

func nilBytesToNull(b []byte) []byte {
	if len(b) == 0 {
		return []byte("null")
	}
	return b
}

func (s *service) shouldPersistByPolicy(entry *Entry) bool {
	if s.policyEngine == nil || entry == nil {
		return true
	}
	decision := s.policyEngine.Decide(logpolicy.PipelineTelemetry, map[string]string{
		logpolicy.MatchFieldLevel: entry.Level,
		logpolicy.MatchFieldEvent: entry.Event,
		logpolicy.MatchFieldRoute: entry.Route,
	})
	switch decision.Decision {
	case logpolicy.DecisionDeny:
		return false
	case logpolicy.DecisionSample:
		key := fmt.Sprintf("%s|%s|%s|%s", entry.SessionID, entry.Event, entry.Route, entry.Timestamp.UTC().Format(time.RFC3339Nano))
		return logpolicy.ShouldKeepBySample(decision.SampleRate, key)
	default:
		return true
	}
}
