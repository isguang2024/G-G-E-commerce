package audit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/gg-ecommerce/backend/internal/modules/observability/logpolicy"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

func newAsyncTestService(cfg Config) *service {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 128
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 100 * time.Millisecond
	}
	s := &service{
		log:          zap.NewNop(),
		cfg:          cfg,
		redactor:     newRedactor(DefaultRedactFields),
		stopCh:       make(chan struct{}),
		cbState:      circuitClosed,
		policyEngine: cfg.PolicyEngine,
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

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal(msg)
}

type memoryDegradedSink struct {
	mu       sync.Mutex
	rows     []*AuditLog
	closed   bool
	replayed int
}

func (m *memoryDegradedSink) Append(row *AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("sink closed")
	}
	cp := *row
	m.rows = append(m.rows, &cp)
	return nil
}

func (m *memoryDegradedSink) Replay(apply func(*AuditLog) error) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errors.New("sink closed")
	}
	rows := make([]*AuditLog, len(m.rows))
	copy(rows, m.rows)
	m.rows = nil
	m.mu.Unlock()
	for _, row := range rows {
		if err := apply(row); err != nil {
			return err
		}
		m.mu.Lock()
		m.replayed++
		m.mu.Unlock()
	}
	return nil
}

func (m *memoryDegradedSink) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *memoryDegradedSink) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rows)
}

func (m *memoryDegradedSink) replayedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.replayed
}

type staticPolicyEngine struct {
	decision logpolicy.Decision
}

func (s staticPolicyEngine) Decide(_ string, _ map[string]string) logpolicy.Decision { return s.decision }
func (s staticPolicyEngine) Refresh(_ context.Context) error                          { return nil }
func (s staticPolicyEngine) Start(_ context.Context)                                   {}

type policyRepoForAuditTest struct {
	items []logpolicy.LogPolicy
}

func (r policyRepoForAuditTest) List(_ context.Context, _, pipeline string, enabled *bool) ([]logpolicy.LogPolicy, error) {
	result := make([]logpolicy.LogPolicy, 0)
	for _, item := range r.items {
		if pipeline != "" && item.Pipeline != pipeline {
			continue
		}
		if enabled != nil && item.Enabled != *enabled {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}
func (r policyRepoForAuditTest) Get(_ context.Context, _ string, _ uuid.UUID) (*logpolicy.LogPolicy, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r policyRepoForAuditTest) Create(_ context.Context, _ *logpolicy.LogPolicy) error { return nil }
func (r policyRepoForAuditTest) Update(_ context.Context, _ *logpolicy.LogPolicy) error { return nil }
func (r policyRepoForAuditTest) Delete(_ context.Context, _ string, _ uuid.UUID) error  { return nil }
func (r policyRepoForAuditTest) ListEnabled(ctx context.Context, tenantID, pipeline string) ([]logpolicy.LogPolicy, error) {
	enabled := true
	return r.List(ctx, tenantID, pipeline, &enabled)
}

// TestRecorderStats_AsyncDropsWhenQueueFull 验证异步 channel 满时 dropped 正确累加，
// 且 accepted 只在成功写入 DB 后累加（此测试故意不启动 worker，因此 accepted 保持 0）。
func TestRecorderStats_AsyncDropsWhenQueueFull(t *testing.T) {
	s := &service{
		log:      zap.NewNop(),
		cfg:      Config{Enabled: true, AsyncMode: true, QueueSize: 1, Workers: 0},
		redactor: newRedactor(DefaultRedactFields),
		queue:    make(chan *AuditLog, 1),
		stopCh:   make(chan struct{}),
	}
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		s.Record(ctx, Event{Action: "test.action"})
	}
	got := s.Stats()
	if got.QueueCap != 1 {
		t.Fatalf("QueueCap = %d, want 1", got.QueueCap)
	}
	if got.QueueDepth != 1 {
		t.Fatalf("QueueDepth = %d, want 1", got.QueueDepth)
	}
	if got.AcceptedTotal != 0 {
		t.Fatalf("AcceptedTotal = %d, want 0 (worker 未写库)", got.AcceptedTotal)
	}
	if got.DroppedTotal != 4 {
		t.Fatalf("DroppedTotal = %d, want 4", got.DroppedTotal)
	}
}

// TestRecorderStats_BuildFailureDoesNotCount 验证 build 失败（action 为空）既不算 accepted 也不算 dropped。
// 这保证指标不会被「坏入参」污染，对齐注释里"dropped 只记 channel 满"的语义。
func TestRecorderStats_BuildFailureDoesNotCount(t *testing.T) {
	s := &service{
		log:      zap.NewNop(),
		cfg:      Config{Enabled: true, AsyncMode: true, QueueSize: 8, Workers: 0},
		redactor: newRedactor(DefaultRedactFields),
		queue:    make(chan *AuditLog, 8),
		stopCh:   make(chan struct{}),
	}
	s.Record(context.Background(), Event{}) // action 缺失 → build 失败

	got := s.Stats()
	if got.AcceptedTotal != 0 || got.DroppedTotal != 0 {
		t.Fatalf("build-failure Record should not move counters, got %+v", got)
	}
}

// TestRecorderBatchFlush_TimerFlush99 验证 99 条事件在未达 batch_size 时由定时器触发一次 flush。
func TestRecorderBatchFlush_TimerFlush99(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     true,
		QueueSize:     256,
		Workers:       1,
		BatchSize:     100,
		FlushInterval: 120 * time.Millisecond,
	})
	var batchFlushes atomic.Int64
	var persisted atomic.Int64
	s.writeBatchFn = func(_ context.Context, rows []*AuditLog) error {
		batchFlushes.Add(1)
		persisted.Add(int64(len(rows)))
		return nil
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	for i := 0; i < 99; i++ {
		s.Record(context.Background(), Event{Action: "audit.timer.flush"})
	}

	waitUntil(t, time.Second, func() bool {
		return persisted.Load() == 99
	}, "timer flush did not persist 99 rows in time")

	// 额外等两个 interval，确认不会重复刷空 batch。
	time.Sleep(260 * time.Millisecond)
	if got := batchFlushes.Load(); got != 1 {
		t.Fatalf("batch flush count = %d, want 1", got)
	}
	stats := s.Stats()
	if stats.AcceptedTotal != 99 || stats.DroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=99 dropped=0", stats)
	}
}

// TestRecorderBatchFlush_ImmediateWhenFull 验证达到 100 条时立即 flush，不依赖定时器。
func TestRecorderBatchFlush_ImmediateWhenFull(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     true,
		QueueSize:     256,
		Workers:       1,
		BatchSize:     100,
		FlushInterval: 5 * time.Second,
	})
	var batchFlushes atomic.Int64
	var persisted atomic.Int64
	s.writeBatchFn = func(_ context.Context, rows []*AuditLog) error {
		batchFlushes.Add(1)
		persisted.Add(int64(len(rows)))
		return nil
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	for i := 0; i < 100; i++ {
		s.Record(context.Background(), Event{Action: "audit.size.flush"})
	}

	waitUntil(t, 800*time.Millisecond, func() bool {
		return persisted.Load() == 100
	}, "size-based flush did not persist 100 rows in time")
	time.Sleep(150 * time.Millisecond)
	if got := batchFlushes.Load(); got != 1 {
		t.Fatalf("batch flush count = %d, want 1", got)
	}
	stats := s.Stats()
	if stats.AcceptedTotal != 100 || stats.DroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=100 dropped=0", stats)
	}
}

// TestRecorderBatchFlush_ShutdownDrain 验证 Shutdown 会 drain 当前残留 batch。
func TestRecorderBatchFlush_ShutdownDrain(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     true,
		QueueSize:     256,
		Workers:       1,
		BatchSize:     100,
		FlushInterval: 5 * time.Second,
	})
	var batchFlushes atomic.Int64
	var persisted atomic.Int64
	s.writeBatchFn = func(_ context.Context, rows []*AuditLog) error {
		batchFlushes.Add(1)
		persisted.Add(int64(len(rows)))
		return nil
	}

	for i := 0; i < 30; i++ {
		s.Record(context.Background(), Event{Action: "audit.shutdown.drain"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if got := batchFlushes.Load(); got != 1 {
		t.Fatalf("batch flush count = %d, want 1", got)
	}
	if got := persisted.Load(); got != 30 {
		t.Fatalf("persisted rows = %d, want 30", got)
	}
	stats := s.Stats()
	if stats.AcceptedTotal != 30 || stats.DroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=30 dropped=0", stats)
	}
}

// TestRecorderBatchFlush_FallbackToSingleRetry 验证批量写失败时降级逐条写入并统计成功条数。
func TestRecorderBatchFlush_FallbackToSingleRetry(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     true,
		QueueSize:     64,
		Workers:       1,
		BatchSize:     5,
		FlushInterval: 5 * time.Second,
	})
	var batchAttempts atomic.Int64
	var singleAttempts atomic.Int64
	var singleSucceeded atomic.Int64
	s.writeBatchFn = func(_ context.Context, _ []*AuditLog) error {
		batchAttempts.Add(1)
		return errors.New("mock batch error")
	}
	s.writeRowFn = func(_ context.Context, row *AuditLog) error {
		singleAttempts.Add(1)
		if row.Action == "audit.bad.row" {
			return errors.New("mock row error")
		}
		singleSucceeded.Add(1)
		return nil
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})

	for i := 0; i < 4; i++ {
		s.Record(context.Background(), Event{Action: "audit.good.row"})
	}
	s.Record(context.Background(), Event{Action: "audit.bad.row"})

	waitUntil(t, 800*time.Millisecond, func() bool {
		return singleAttempts.Load() == 5
	}, "fallback single-row retries did not run in time")

	if got := batchAttempts.Load(); got != 1 {
		t.Fatalf("batch attempts = %d, want 1", got)
	}
	if got := singleSucceeded.Load(); got != 4 {
		t.Fatalf("single succeeded = %d, want 4", got)
	}
	stats := s.Stats()
	if stats.AcceptedTotal != 4 || stats.DroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=4 dropped=0", stats)
	}
}

// TestRecorderStats_NoopReturnsZero 验证 Noop 实现 Stats() 返回零值结构。
// 线上 audit.Enabled=false / db=nil 时应用会持有 Noop，暴露 /metrics 的 handler
// 必须能安全调用 Stats() 而不 panic。
func TestRecorderStats_NoopReturnsZero(t *testing.T) {
	var r Recorder = Noop{}
	got := r.Stats()
	if got != (Stats{}) {
		t.Fatalf("Noop.Stats() = %+v, want zero value", got)
	}
}

func TestRecorderCircuitBreaker_OpensAfterConsecutiveFailures(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     false,
		BatchSize:     1,
		FlushInterval: time.Second,
	})
	sink := &memoryDegradedSink{}
	s.sink = sink
	s.writeBatchFn = func(_ context.Context, _ []*AuditLog) error {
		return errors.New("db down")
	}
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error {
		return errors.New("db down")
	}
	for i := 0; i < 6; i++ {
		_ = s.flushBatch([]*AuditLog{{Action: "audit.cb.open", Ts: time.Now().UTC()}})
	}

	if !s.Stats().Degraded {
		t.Fatal("expected circuit breaker opened")
	}
	if sink.count() == 0 {
		t.Fatalf("expected degraded sink appended rows, got 0")
	}
}

func TestRecorderCircuitBreaker_OpenWritesDegradedFile(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     false,
		BatchSize:     10,
		FlushInterval: time.Second,
	})
	sink := &memoryDegradedSink{}
	s.sink = sink
	s.cbState = circuitOpen
	s.cbOpenedAt = time.Now().UTC()
	s.writeBatchFn = func(_ context.Context, _ []*AuditLog) error {
		t.Fatal("writeBatch should not be called when circuit open")
		return nil
	}

	flushed := s.flushBatch([]*AuditLog{{Action: "audit.cb.degraded", Ts: time.Now().UTC()}})
	if flushed != 0 {
		t.Fatalf("flushBatch returned %d, want 0", flushed)
	}
	if sink.count() != 1 {
		t.Fatalf("degraded sink rows = %d, want 1", sink.count())
	}
}

func TestRecorderCircuitBreaker_HalfOpenSuccessReplaysAndCloses(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     false,
		BatchSize:     10,
		FlushInterval: time.Second,
	})
	sink := &memoryDegradedSink{
		rows: []*AuditLog{
			{ID: 2, Ts: time.Now().UTC().Add(-2 * time.Second), Action: "audit.replay.old.1"},
			{ID: 3, Ts: time.Now().UTC().Add(-time.Second), Action: "audit.replay.old.2"},
		},
	}
	s.sink = sink
	s.cbState = circuitOpen
	s.cbOpenedAt = time.Now().UTC().Add(-31 * time.Second)
	var replayWrites atomic.Int64
	s.writeBatchFn = func(_ context.Context, _ []*AuditLog) error { return nil }
	s.writeRowFn = func(_ context.Context, row *AuditLog) error {
		if row != nil && row.Action != "audit.cb.probe" {
			replayWrites.Add(1)
		}
		return nil
	}

	flushed := s.flushBatch([]*AuditLog{{ID: 1, Ts: time.Now().UTC(), Action: "audit.cb.probe"}})
	if flushed != 1 {
		t.Fatalf("flushBatch returned %d, want 1", flushed)
	}
	if s.isDegradedMode() {
		t.Fatal("expected circuit closed after half-open success")
	}
	if got := replayWrites.Load(); got != 2 {
		t.Fatalf("replay writes = %d, want 2", got)
	}
	if sink.replayedCount() != 2 {
		t.Fatalf("sink replayed = %d, want 2", sink.replayedCount())
	}
}

func TestRecorderCircuitBreaker_HalfOpenFailureReturnsOpen(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:       true,
		AsyncMode:     false,
		BatchSize:     10,
		FlushInterval: time.Second,
	})
	sink := &memoryDegradedSink{}
	s.sink = sink
	s.cbState = circuitOpen
	s.cbOpenedAt = time.Now().UTC().Add(-31 * time.Second)
	s.writeBatchFn = func(_ context.Context, _ []*AuditLog) error {
		return errors.New("still down")
	}
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error {
		return errors.New("still down")
	}

	flushed := s.flushBatch([]*AuditLog{{Action: "audit.cb.halfopen.fail", Ts: time.Now().UTC()}})
	if flushed != 0 {
		t.Fatalf("flushBatch returned %d, want 0", flushed)
	}
	if !s.isDegradedMode() {
		t.Fatal("expected circuit to remain open")
	}
	if sink.count() != 1 {
		t.Fatalf("degraded sink rows = %d, want 1", sink.count())
	}
}

func TestRecorderCircuitBreaker_ShutdownClosesDegradedSink(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:   true,
		AsyncMode: false,
	})
	dir := t.TempDir()
	filePath := filepath.Join(dir, "audit_degraded.jsonl")
	sink, err := newJSONLDegradedSink(filePath)
	if err != nil {
		t.Fatalf("newJSONLDegradedSink() error = %v", err)
	}
	s.sink = sink

	if err := s.sink.Append(&AuditLog{Action: "audit.before.shutdown", Ts: time.Now().UTC()}); err != nil {
		t.Fatalf("append before shutdown error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := s.sink.Append(&AuditLog{Action: "audit.after.shutdown", Ts: time.Now().UTC()}); err == nil {
		t.Fatal("expected append after shutdown to fail because sink closed")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("degraded file should remain on disk, stat error = %v", err)
	}
}

func TestRecorderPolicy_DenyDropsAndCounts(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:      true,
		AsyncMode:    false,
		PolicyEngine: staticPolicyEngine{decision: logpolicy.Decision{Decision: logpolicy.DecisionDeny}},
	})
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error { return nil }
	s.Record(context.Background(), Event{Action: "system.user.update"})

	stats := s.Stats()
	if stats.AcceptedTotal != 0 || stats.PolicyDroppedTotal != 1 {
		t.Fatalf("stats = %+v, want accepted=0 policy_dropped=1", stats)
	}
}

func TestRecorderPolicy_Sample50Distribution(t *testing.T) {
	rate := 50
	s := newAsyncTestService(Config{
		Enabled:   true,
		AsyncMode: false,
		PolicyEngine: staticPolicyEngine{decision: logpolicy.Decision{
			Decision:   logpolicy.DecisionSample,
			SampleRate: rate,
		}},
	})
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error { return nil }
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		ctx = logger.WithRequestID(ctx, "req-policy-sample-"+time.Now().Add(time.Duration(i)*time.Nanosecond).Format(time.RFC3339Nano))
		s.Record(ctx, Event{Action: "system.user.update"})
	}
	stats := s.Stats()
	kept := float64(stats.AcceptedTotal) / 1000.0
	if kept < 0.4 || kept > 0.6 {
		t.Fatalf("sample kept ratio %.4f out of range, stats=%+v", kept, stats)
	}
	if stats.PolicyDroppedTotal == 0 {
		t.Fatalf("expected policy drops > 0, stats=%+v", stats)
	}
}

func TestRecorderPolicy_DefaultAllowWhenNoRule(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:   true,
		AsyncMode: false,
	})
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error { return nil }
	for i := 0; i < 10; i++ {
		s.Record(context.Background(), Event{Action: "system.user.update"})
	}
	stats := s.Stats()
	if stats.AcceptedTotal != 10 || stats.PolicyDroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=10 policy_dropped=0", stats)
	}
}

func TestRecorderPolicy_ComplianceLockAllows(t *testing.T) {
	repo := policyRepoForAuditTest{
		items: []logpolicy.LogPolicy{
			{
				ID:         uuid.New(),
				TenantID:   logpolicy.DefaultTenantID,
				Pipeline:   logpolicy.PipelineAudit,
				MatchField: logpolicy.MatchFieldAction,
				Pattern:    "observability.policy.*",
				Decision:   logpolicy.DecisionDeny,
				Priority:   999,
				Enabled:    true,
			},
		},
	}
	engine := logpolicy.NewEngine(repo, zap.NewNop())
	_ = engine.Refresh(context.Background())

	s := newAsyncTestService(Config{
		Enabled:      true,
		AsyncMode:    false,
		PolicyEngine: engine,
	})
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error { return nil }
	s.Record(context.Background(), Event{Action: "observability.policy.delete"})

	stats := s.Stats()
	if stats.AcceptedTotal != 1 || stats.PolicyDroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=1 policy_dropped=0", stats)
	}
}

func TestRecorderPolicy_NilEngineAllowAll(t *testing.T) {
	s := newAsyncTestService(Config{
		Enabled:      true,
		AsyncMode:    false,
		PolicyEngine: nil,
	})
	s.writeRowFn = func(_ context.Context, _ *AuditLog) error { return nil }
	s.Record(context.Background(), Event{Action: "system.user.update"})
	stats := s.Stats()
	if stats.AcceptedTotal != 1 || stats.PolicyDroppedTotal != 0 {
		t.Fatalf("stats = %+v, want accepted=1 policy_dropped=0", stats)
	}
}
