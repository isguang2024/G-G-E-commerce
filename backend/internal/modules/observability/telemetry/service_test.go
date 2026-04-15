package telemetry

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
	"github.com/gg-ecommerce/backend/internal/modules/observability/logpolicy"
)

// newTestService 构造一个不依赖真实 DB / worker 的 *service，用于 Stats() 相关测试。
// 关键点：
//   - queueSize > 0 才创建 chan，但不启动 workers，这样 queue 可以被"塞满"；
//   - sessionRate / ipRate 由调用方控制，支持限流被动触发的测试场景。
func newTestService(queueSize int, sessionCap, sessionRate, ipCap, ipRate float64, policyEngine logpolicy.Engine) *service {
	s := &service{
		log: zap.NewNop(),
		cfg: Config{
			Enabled:         true,
			QueueSize:       queueSize,
			Workers:         0, // 刻意不启动
			PerSessionRate:  sessionRate,
			PerSessionBurst: sessionCap,
			PerIPRate:       ipRate,
			PerIPBurst:      ipCap,
			BucketIdleTTL:   5 * time.Minute,
			MaxMessageBytes: 16 * 1024,
			PolicyEngine:    policyEngine,
		},
		redactor:       auditRedactorAdapter{r: audit.NewPublicRedactor(audit.DefaultRedactFields)},
		sessionLimiter: newRateLimiter(sessionCap, sessionRate, 5*time.Minute),
		ipLimiter:      newRateLimiter(ipCap, ipRate, 5*time.Minute),
		stopCh:         make(chan struct{}),
		policyEngine:   policyEngine,
	}
	if queueSize > 0 {
		s.queue = make(chan *TelemetryLog, queueSize)
	}
	return s
}

type staticPolicyEngine struct {
	decision logpolicy.Decision
}

func (s staticPolicyEngine) Decide(_ string, _ map[string]string) logpolicy.Decision {
	return s.decision
}
func (s staticPolicyEngine) Refresh(_ context.Context) error { return nil }
func (s staticPolicyEngine) Start(_ context.Context)         {}

// TestIngesterStats_QueueFullDropsAndAccepted 验证 queue 满时的 dropped 计数路径，
// 以及成功入队的 accepted 计数。worker=0 保证 queue 不会被消费。
func TestIngesterStats_QueueFullDropsAndAccepted(t *testing.T) {
	s := newTestService(2, 100, 100, 100, 100, nil)
	entries := make([]Entry, 5)
	for i := range entries {
		entries[i] = Entry{Level: "info", Event: "page.view"}
	}
	res := s.Ingest(context.Background(), entries, "sess-A", "1.2.3.4")
	if res.Accepted != 2 || res.Dropped != 3 {
		t.Fatalf("per-batch Result=%+v, want Accepted=2 Dropped=3", res)
	}
	got := s.Stats()
	if got.QueueCap != 2 {
		t.Fatalf("QueueCap = %d, want 2", got.QueueCap)
	}
	if got.QueueDepth != 2 {
		t.Fatalf("QueueDepth = %d, want 2", got.QueueDepth)
	}
	if got.AcceptedTotal != 2 {
		t.Fatalf("AcceptedTotal = %d, want 2", got.AcceptedTotal)
	}
	if got.DroppedTotal != 3 {
		t.Fatalf("DroppedTotal = %d, want 3", got.DroppedTotal)
	}
}

// TestIngesterStats_RateLimitedBatchDropsAll 验证 session / IP 限流时整批 dropped
// 累加到 Stats.DroppedTotal，accepted 不变（符合 Ingest 注释"整批被扣住"的语义）。
func TestIngesterStats_RateLimitedBatchDropsAll(t *testing.T) {
	// cap=1, rate=0 → 第 2 次 allow 永远返回 false
	s := newTestService(16, 1, 0, 1000, 1000, nil)

	// 第 1 次 ingest 可通过，消耗 session token。
	first := []Entry{{Level: "info", Event: "first.event"}}
	if r := s.Ingest(context.Background(), first, "sess-B", "5.6.7.8"); r.Accepted != 1 {
		t.Fatalf("first batch expected accepted=1, got %+v", r)
	}

	// 第 2 次 ingest 因 session token 耗尽直接 dropped，整批 3 条被扣住。
	second := make([]Entry, 3)
	for i := range second {
		second[i] = Entry{Level: "warn", Event: "rate.limited"}
	}
	res := s.Ingest(context.Background(), second, "sess-B", "5.6.7.8")
	if res.Accepted != 0 || res.Dropped != 3 {
		t.Fatalf("rate-limited Result=%+v, want Accepted=0 Dropped=3", res)
	}

	got := s.Stats()
	if got.AcceptedTotal != 1 {
		t.Fatalf("AcceptedTotal = %d, want 1", got.AcceptedTotal)
	}
	if got.DroppedTotal != 3 {
		t.Fatalf("DroppedTotal = %d, want 3", got.DroppedTotal)
	}
}

// TestIngesterStats_NoopReturnsZero 验证 Noop 实现 Stats() 返回零值。
// 关闭 telemetry / 无 DB 的环境下不应 panic。
func TestIngesterStats_NoopReturnsZero(t *testing.T) {
	var in Ingester = Noop{}
	got := in.Stats()
	if got != (Stats{}) {
		t.Fatalf("Noop.Stats() = %+v, want zero value", got)
	}
}

func TestIngesterPolicy_DenyDropsAndCounts(t *testing.T) {
	s := newTestService(4, 100, 100, 100, 100, staticPolicyEngine{
		decision: logpolicy.Decision{Decision: logpolicy.DecisionDeny},
	})
	res := s.Ingest(context.Background(), []Entry{{Level: "info", Event: "frontend.click", Route: "/home"}}, "sess-deny", "1.1.1.1")
	if res.Accepted != 0 || res.Dropped != 1 {
		t.Fatalf("Result = %+v, want accepted=0 dropped=1", res)
	}
	stats := s.Stats()
	if stats.PolicyDroppedTotal != 1 || stats.DroppedTotal != 1 || stats.AcceptedTotal != 0 {
		t.Fatalf("Stats = %+v, want policy_dropped=1 dropped=1 accepted=0", stats)
	}
}

func TestIngesterPolicy_Sample50Distribution(t *testing.T) {
	s := newTestService(1200, 5000, 5000, 5000, 5000, staticPolicyEngine{
		decision: logpolicy.Decision{
			Decision:   logpolicy.DecisionSample,
			SampleRate: 50,
		},
	})
	entries := make([]Entry, 1000)
	base := time.Now().UTC()
	for i := range entries {
		entries[i] = Entry{
			Level:     "info",
			Event:     "frontend.sample",
			Route:     "/sample",
			SessionID: "sess-sample",
			Timestamp: base.Add(time.Duration(i) * time.Nanosecond),
		}
	}
	res := s.Ingest(context.Background(), entries, "sess-sample", "2.2.2.2")
	kept := float64(res.Accepted) / 1000.0
	if kept < 0.4 || kept > 0.6 {
		t.Fatalf("sample kept ratio %.4f out of range, result=%+v", kept, res)
	}
	stats := s.Stats()
	if stats.PolicyDroppedTotal == 0 {
		t.Fatalf("expected policy drops > 0, stats=%+v", stats)
	}
}

func TestIngesterPolicy_NilEngineAllowAll(t *testing.T) {
	s := newTestService(4, 100, 100, 100, 100, nil)
	res := s.Ingest(context.Background(), []Entry{
		{Level: "info", Event: "frontend.a", Route: "/a"},
		{Level: "warn", Event: "frontend.b", Route: "/b"},
	}, "sess-allow", "3.3.3.3")
	if res.Accepted != 2 || res.Dropped != 0 {
		t.Fatalf("Result = %+v, want accepted=2 dropped=0", res)
	}
	stats := s.Stats()
	if stats.PolicyDroppedTotal != 0 || stats.AcceptedTotal != 2 {
		t.Fatalf("Stats = %+v, want policy_dropped=0 accepted=2", stats)
	}
}
